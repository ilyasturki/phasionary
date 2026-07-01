package com.phasionary.app.di

import android.content.Context
import com.phasionary.app.data.net.defaultHttpClient
import com.phasionary.app.data.repo.PhasionaryRepository
import com.phasionary.app.data.repo.RemotePhasionaryRepository
import com.phasionary.app.data.settings.SettingsRepository

/**
 * Manual dependency container held by the Application. v1 avoids a DI framework
 * (Hilt/KSP) to keep the build simple; if the graph grows, this is the single
 * place to swap in one.
 */
class AppContainer(context: Context) {

    val settings: SettingsRepository = SettingsRepository(context)

    private val httpClient = defaultHttpClient()

    val repository: PhasionaryRepository =
        RemotePhasionaryRepository(httpClient) { settings.current() }
}
