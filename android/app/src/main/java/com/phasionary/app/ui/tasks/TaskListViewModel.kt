package com.phasionary.app.ui.tasks

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.createSavedStateHandle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.initializer
import androidx.lifecycle.viewmodel.viewModelFactory
import com.phasionary.app.data.ApiException
import com.phasionary.app.data.model.CreateTaskBody
import com.phasionary.app.data.model.Project
import com.phasionary.app.data.model.Status
import com.phasionary.app.data.model.Task
import com.phasionary.app.data.model.UpdateTaskBody
import com.phasionary.app.data.model.updateTask
import com.phasionary.app.data.repo.PhasionaryRepository
import com.phasionary.app.ui.appContainer
import com.phasionary.app.ui.userMessage
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class TaskListUiState(
    val loading: Boolean = true,
    val project: Project? = null,
    val error: String? = null,
    val notConfigured: Boolean = false,
    val collapsed: Set<String> = emptySet(),
    /** One-shot message for a snackbar (e.g. a failed status change). */
    val message: String? = null,
)

const val PROJECT_ID_ARG = "projectId"

class TaskListViewModel(
    savedStateHandle: SavedStateHandle,
    private val repository: PhasionaryRepository,
) : ViewModel() {

    private val projectId: String = savedStateHandle.get<String>(PROJECT_ID_ARG).orEmpty()

    private val _state = MutableStateFlow(TaskListUiState())
    val state: StateFlow<TaskListUiState> = _state.asStateFlow()

    init {
        load()
    }

    fun load() = fetch(showSpinner = true)

    /**
     * Re-reads the project without flashing the spinner. The screen calls this on
     * resume, so returning from the editor — or from the TUI having changed
     * something on the other end — shows current data instead of a stale list.
     */
    fun refresh() {
        if (_state.value.project == null) load() else fetch(showSpinner = false)
    }

    private fun fetch(showSpinner: Boolean) {
        viewModelScope.launch {
            if (showSpinner) {
                _state.update { it.copy(loading = true, error = null, notConfigured = false) }
            }
            try {
                val project = repository.getProject(projectId)
                // Folds live on the server so the TUI and the phone agree on
                // what's collapsed. They're a display nicety, so a failure here
                // must not fail the load — fall back to what we already show.
                val collapsed = try {
                    repository.getFolds(projectId).toSet()
                } catch (e: ApiException) {
                    _state.value.collapsed
                }
                _state.update {
                    it.copy(
                        loading = false,
                        project = project,
                        collapsed = collapsed,
                        error = null,
                        notConfigured = false,
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

    /**
     * Collapses or expands a category, optimistically, then writes the whole
     * collapsed set back to the server (the API replaces the list wholesale).
     * A failure reverts, so the caret never lies about what was saved.
     */
    fun toggleCategory(categoryId: String) {
        val previous = _state.value.collapsed
        val next = previous.toMutableSet().apply {
            if (!add(categoryId)) remove(categoryId)
        }
        _state.update { it.copy(collapsed = next) }

        viewModelScope.launch {
            try {
                repository.setFolds(projectId, next.toList())
            } catch (e: ApiException) {
                _state.update { it.copy(collapsed = previous, message = userMessage(e)) }
            }
        }
    }

    /**
     * Advances a task's status (todo -> in_progress -> completed -> cancelled ->
     * todo). Updates the UI optimistically, then reconciles with the server's
     * returned task; on failure it reverts and surfaces a message.
     */
    fun cycleStatus(categoryId: String, task: Task) {
        val previous = task.status
        val next = Status.next(previous)

        _state.update { s ->
            val project = s.project?.updateTask(categoryId, task.id) { it.copy(status = next) }
            s.copy(project = project)
        }

        viewModelScope.launch {
            try {
                val updated = repository.setTaskStatus(projectId, categoryId, task.id, next)
                _state.update { s ->
                    val project = s.project?.updateTask(categoryId, task.id) { updated }
                    s.copy(project = project)
                }
            } catch (e: ApiException) {
                _state.update { s ->
                    val project = s.project?.updateTask(categoryId, task.id) { it.copy(status = previous) }
                    s.copy(project = project, message = userMessage(e))
                }
            }
        }
    }

    /**
     * Quick-captures a task with only a title (everything else takes the
     * server's defaults; the editor fills in the rest). The created task is
     * appended locally so it appears without a full reload.
     */
    fun addTask(categoryId: String, title: String) {
        val trimmed = title.trim()
        if (trimmed.isEmpty()) return

        viewModelScope.launch {
            try {
                val created = repository.createTask(
                    projectId,
                    categoryId,
                    CreateTaskBody(title = trimmed),
                )
                _state.update { s ->
                    val project = s.project?.let { p ->
                        p.copy(
                            categories = p.categories.map { category ->
                                if (category.id != categoryId) {
                                    category
                                } else {
                                    category.copy(tasks = category.tasks + created)
                                }
                            },
                        )
                    }
                    s.copy(project = project)
                }
            } catch (e: ApiException) {
                _state.update { it.copy(message = userMessage(e)) }
            }
        }
    }

    /**
     * Sets a separator's label. A blank label is meaningful here — it turns the
     * separator back into a plain rule — so it isn't filtered out the way a
     * blank task title is.
     */
    fun renameSeparator(categoryId: String, taskId: String, label: String) {
        val trimmed = label.trim()
        _state.update { s ->
            val project = s.project?.updateTask(categoryId, taskId) { it.copy(title = trimmed) }
            s.copy(project = project)
        }

        viewModelScope.launch {
            try {
                val updated = repository.updateTask(
                    projectId,
                    categoryId,
                    taskId,
                    UpdateTaskBody(title = trimmed),
                )
                _state.update { s ->
                    val project = s.project?.updateTask(categoryId, taskId) { updated }
                    s.copy(project = project)
                }
            } catch (e: ApiException) {
                // No revert value to restore optimistically here, so re-read
                // rather than leave the row showing a label the server rejected.
                _state.update { it.copy(message = userMessage(e)) }
                refresh()
            }
        }
    }

    /** Removes a row (used for separators, which have no editor page). */
    fun deleteRow(categoryId: String, taskId: String) {
        viewModelScope.launch {
            try {
                repository.deleteTask(projectId, categoryId, taskId)
                _state.update { s ->
                    val project = s.project?.let { p ->
                        p.copy(
                            categories = p.categories.map { category ->
                                if (category.id != categoryId) {
                                    category
                                } else {
                                    category.copy(tasks = category.tasks.filterNot { it.id == taskId })
                                }
                            },
                        )
                    }
                    s.copy(project = project)
                }
            } catch (e: ApiException) {
                _state.update { it.copy(message = userMessage(e)) }
            }
        }
    }

    /** Adds an empty category at the end of the project. */
    fun addCategory(name: String) {
        val trimmed = name.trim()
        if (trimmed.isEmpty()) return

        viewModelScope.launch {
            try {
                val created = repository.createCategory(projectId, trimmed)
                _state.update { s ->
                    val project = s.project?.let { p ->
                        p.copy(categories = p.categories + created)
                    }
                    s.copy(project = project)
                }
            } catch (e: ApiException) {
                _state.update { it.copy(message = userMessage(e)) }
            }
        }
    }

    fun consumeMessage() {
        _state.update { it.copy(message = null) }
    }

    companion object {
        val Factory = viewModelFactory {
            initializer {
                TaskListViewModel(
                    savedStateHandle = createSavedStateHandle(),
                    repository = appContainer().repository,
                )
            }
        }
    }
}
