package com.phasionary.app.ui.settings

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.initializer
import androidx.lifecycle.viewmodel.viewModelFactory
import com.phasionary.app.data.settings.SettingsRepository
import com.phasionary.app.ui.appContainer
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.io.IOException

data class SettingsUiState(
    val baseUrl: String = "",
    val token: String = "",
    val loaded: Boolean = false,
    /** One-shot: set after a successful save to trigger navigation back. */
    val saved: Boolean = false,
    /** One-shot snackbar message (e.g. a save failure). */
    val message: String? = null,
)

class SettingsViewModel(
    private val settings: SettingsRepository,
) : ViewModel() {

    private val _state = MutableStateFlow(SettingsUiState())
    val state: StateFlow<SettingsUiState> = _state.asStateFlow()

    init {
        viewModelScope.launch {
            val cfg = settings.current()
            _state.update { it.copy(baseUrl = cfg.baseUrl, token = cfg.token, loaded = true) }
        }
    }

    fun onBaseUrlChange(value: String) {
        _state.update { it.copy(baseUrl = value, saved = false) }
    }

    fun onTokenChange(value: String) {
        _state.update { it.copy(token = value, saved = false) }
    }

    fun save() {
        viewModelScope.launch {
            try {
                settings.save(_state.value.baseUrl, _state.value.token)
                _state.update { it.copy(saved = true) }
            } catch (e: IOException) {
                _state.update { it.copy(message = "Couldn't save settings: ${e.message ?: "unknown error"}") }
            }
        }
    }

    fun consumeMessage() {
        _state.update { it.copy(message = null) }
    }

    companion object {
        val Factory = viewModelFactory {
            initializer { SettingsViewModel(appContainer().settings) }
        }
    }
}
