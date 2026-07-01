package com.phasionary.app.ui.tasks

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.createSavedStateHandle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.initializer
import androidx.lifecycle.viewmodel.viewModelFactory
import com.phasionary.app.data.ApiException
import com.phasionary.app.data.model.Project
import com.phasionary.app.data.model.Status
import com.phasionary.app.data.model.Task
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

    fun load() {
        viewModelScope.launch {
            _state.update { it.copy(loading = true, error = null, notConfigured = false) }
            try {
                val project = repository.getProject(projectId)
                _state.update { it.copy(loading = false, project = project, error = null) }
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

    fun toggleCategory(categoryId: String) {
        _state.update { s ->
            val next = s.collapsed.toMutableSet().apply {
                if (!add(categoryId)) remove(categoryId)
            }
            s.copy(collapsed = next)
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
