package com.phasionary.app.ui

import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewmodel.CreationExtras
import com.phasionary.app.PhasionaryApp
import com.phasionary.app.di.AppContainer

/**
 * Pulls the app's [AppContainer] out of the Application inside a ViewModel
 * factory initializer, so ViewModels get their dependencies without a DI
 * framework.
 */
fun CreationExtras.appContainer(): AppContainer {
    val app = this[ViewModelProvider.AndroidViewModelFactory.APPLICATION_KEY] as PhasionaryApp
    return app.container
}
