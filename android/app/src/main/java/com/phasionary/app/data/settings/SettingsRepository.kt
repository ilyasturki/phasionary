package com.phasionary.app.data.settings

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map

/** Server connection settings. Blank until the user configures them. */
data class ServerConfig(
    val baseUrl: String = "",
    val token: String = "",
) {
    val isConfigured: Boolean get() = baseUrl.isNotBlank() && token.isNotBlank()
}

// Top-level delegate as required by the DataStore API (one instance per process).
private val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = "phasionary_settings")

/**
 * Persists the base URL + bearer token in a Preferences DataStore.
 *
 * v1 stores the token in plaintext DataStore, which is acceptable for a private,
 * single-user, Tailscale-only setup. A later pass can move it to an encrypted
 * store without changing this interface.
 */
class SettingsRepository(context: Context) {

    private val appContext = context.applicationContext

    private object Keys {
        val BASE_URL = stringPreferencesKey("base_url")
        val TOKEN = stringPreferencesKey("token")
    }

    val config: Flow<ServerConfig> = appContext.dataStore.data.map { prefs ->
        ServerConfig(
            baseUrl = prefs[Keys.BASE_URL].orEmpty(),
            token = prefs[Keys.TOKEN].orEmpty(),
        )
    }

    suspend fun current(): ServerConfig = config.first()

    suspend fun save(baseUrl: String, token: String) {
        appContext.dataStore.edit { prefs ->
            prefs[Keys.BASE_URL] = baseUrl.trim().trimEnd('/')
            prefs[Keys.TOKEN] = token.trim()
        }
    }
}
