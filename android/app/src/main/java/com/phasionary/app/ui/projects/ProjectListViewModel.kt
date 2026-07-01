package com.phasionary.app.ui.projects

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.initializer
import androidx.lifecycle.viewmodel.viewModelFactory
import com.phasionary.app.data.ApiException
import com.phasionary.app.data.model.Project
import com.phasionary.app.data.repo.PhasionaryRepository
import com.phasionary.app.ui.appContainer
import com.phasionary.app.ui.userMessage
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

sealed interface ProjectListUiState {
    data object Loading : ProjectListUiState
    data class Error(val message: String, val notConfigured: Boolean) : ProjectListUiState
    data class Success(val projects: List<Project>) : ProjectListUiState
}

class ProjectListViewModel(
    private val repository: PhasionaryRepository,
) : ViewModel() {

    private val _state = MutableStateFlow<ProjectListUiState>(ProjectListUiState.Loading)
    val state: StateFlow<ProjectListUiState> = _state.asStateFlow()

    init {
        load()
    }

    fun load() {
        viewModelScope.launch {
            _state.value = ProjectListUiState.Loading
            _state.value = try {
                ProjectListUiState.Success(repository.listProjects())
            } catch (e: ApiException) {
                ProjectListUiState.Error(
                    message = userMessage(e),
                    notConfigured = e is ApiException.NotConfigured,
                )
            }
        }
    }

    companion object {
        val Factory = viewModelFactory {
            initializer { ProjectListViewModel(appContainer().repository) }
        }
    }
}
