package com.phasionary.app

import android.app.Application
import com.phasionary.app.di.AppContainer

/** Holds the manual DI container for the process lifetime. */
class PhasionaryApp : Application() {

    lateinit var container: AppContainer
        private set

    override fun onCreate() {
        super.onCreate()
        container = AppContainer(this)
    }
}
