package com.phasionary.app.ui.tasks

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.createSavedStateHandle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.initializer
import androidx.lifecycle.viewmodel.viewModelFactory
import com.phasionary.app.data.ApiException
import com.phasionary.app.data.model.Priority
import com.phasionary.app.data.model.Status
import com.phasionary.app.data.model.Task
import com.phasionary.app.data.model.UpdateTaskBody
import com.phasionary.app.data.repo.PhasionaryRepository
import com.phasionary.app.ui.appContainer
import com.phasionary.app.ui.userMessage
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

const val CATEGORY_ID_ARG = "categoryId"
const val TASK_ID_ARG = "taskId"

/**
 * The editable form, held separately from the loaded task so [TaskEditUiState]
 * can compare the two and know whether anything is dirty. Estimate is text
 * because a half-typed number isn't an Int yet.
 */
data class TaskForm(
    val title: String = "",
    val description: String = "",
    val status: String = Status.TODO,
    val priority: String = Priority.NONE,
    val estimateText: String = "",
) {
    /** Null when the field isn't a usable number — the UI blocks saving on it. */
    val estimateMinutes: Int?
        get() = if (estimateText.isBlank()) 0 else estimateText.trim().toIntOrNull()?.takeIf { it >= 0 }

    companion object {
        fun of(task: Task) = TaskForm(
            title = task.title,
            description = task.description,
            status = task.status,
            priority = task.priority,
            estimateText = if (task.estimateMinutes > 0) task.estimateMinutes.toString() else "",
        )
    }
}

data class TaskEditUiState(
    val loading: Boolean = true,
    val saving: Boolean = false,
    val error: String? = null,
    val notConfigured: Boolean = false,
    val message: String? = null,
    /** The task as the server last gave it to us; null until loaded. */
    val original: Task? = null,
    val form: TaskForm = TaskForm(),
    /** Set once the write lands, so the screen knows to navigate back. */
    val done: Boolean = false,
) {
    val dirty: Boolean
        get() = original != null && form != TaskForm.of(original)

    /** Blank titles and unparseable estimates are rejected before any request. */
    val canSave: Boolean
        get() = !saving && dirty && form.title.isNotBlank() && form.estimateMinutes != null
}

class TaskEditViewModel(
    savedStateHandle: SavedStateHandle,
    private val repository: PhasionaryRepository,
) : ViewModel() {

    private val projectId: String = savedStateHandle.get<String>(PROJECT_ID_ARG).orEmpty()
    private val categoryId: String = savedStateHandle.get<String>(CATEGORY_ID_ARG).orEmpty()
    private val taskId: String = savedStateHandle.get<String>(TASK_ID_ARG).orEmpty()

    private val _state = MutableStateFlow(TaskEditUiState())
    val state: StateFlow<TaskEditUiState> = _state.asStateFlow()

    init {
        load()
    }

    /** There is no single-task GET, so the task comes out of the project fetch. */
    fun load() {
        viewModelScope.launch {
            _state.update { it.copy(loading = true, error = null, notConfigured = false) }
            try {
                val project = repository.getProject(projectId)
                val task = project.categories
                    .firstOrNull { it.id == categoryId }
                    ?.tasks
                    ?.firstOrNull { it.id == taskId }

                if (task == null) {
                    _state.update {
                        it.copy(loading = false, error = "This task no longer exists.")
                    }
                    return@launch
                }
                _state.update {
                    it.copy(
                        loading = false,
                        original = task,
                        form = TaskForm.of(task),
                        error = null,
                    )
                }
            } catch (e: ApiException) {
                _state.update {
                    it.copy(
                        loading = false,
                        error = userMessage(e),
                        notConfigured = e is ApiException.NotConfigured,
                    )
                }
            }
        }
    }

    fun updateForm(transform: (TaskForm) -> TaskForm) {
        _state.update { it.copy(form = transform(it.form)) }
    }

    /**
     * Writes the edits in one PATCH carrying only the changed fields, so a
     * concurrent TUI edit to a field the user didn't touch survives.
     */
    fun save() {
        val current = _state.value
        val original = current.original ?: return
        if (!current.canSave) return

        val body = diff(original, current.form)
        if (body.isEmpty) {
            _state.update { it.copy(done = true) }
            return
        }

        _state.update { it.copy(saving = true) }
        viewModelScope.launch {
            try {
                repository.updateTask(projectId, categoryId, taskId, body)
                _state.update { it.copy(saving = false, done = true) }
            } catch (e: ApiException) {
                _state.update { it.copy(saving = false, message = userMessage(e)) }
            }
        }
    }

    /**
     * Inserts a divider directly below this task — the touch equivalent of the
     * TUI's `-`. It deliberately stays on this page rather than navigating
     * back: leaving would either discard unsaved edits or need a second prompt,
     * and a snackbar is enough confirmation. The new separator shows up in the
     * list, ready to be labeled, when the screen resumes.
     */
    fun insertSeparatorBelow() {
        val current = _state.value
        if (current.original == null || current.saving) return

        _state.update { it.copy(saving = true) }
        viewModelScope.launch {
            try {
                repository.createSeparator(projectId, categoryId, taskId)
                _state.update { it.copy(saving = false, message = "Separator inserted below.") }
            } catch (e: ApiException) {
                _state.update { it.copy(saving = false, message = userMessage(e)) }
            }
        }
    }

    fun delete() {
        _state.update { it.copy(saving = true) }
        viewModelScope.launch {
            try {
                repository.deleteTask(projectId, categoryId, taskId)
                _state.update { it.copy(saving = false, done = true) }
            } catch (e: ApiException) {
                _state.update { it.copy(saving = false, message = userMessage(e)) }
            }
        }
    }

    fun consumeMessage() {
        _state.update { it.copy(message = null) }
    }

    companion object {
        /**
         * Builds the partial body: a field equal to the loaded task stays null
         * and is omitted from the JSON, which the server reads as "leave alone".
         */
        internal fun diff(original: Task, form: TaskForm): UpdateTaskBody {
            val estimate = form.estimateMinutes
            return UpdateTaskBody(
                title = form.title.trim().takeIf { it != original.title },
                status = form.status.takeIf { it != original.status },
                priority = form.priority.takeIf { it != original.priority },
                estimateMinutes = estimate?.takeIf { it != original.estimateMinutes },
                description = form.description.takeIf { it != original.description },
            )
        }

        val Factory = viewModelFactory {
            initializer {
                TaskEditViewModel(
                    savedStateHandle = createSavedStateHandle(),
                    repository = appContainer().repository,
                )
            }
        }
    }
}
