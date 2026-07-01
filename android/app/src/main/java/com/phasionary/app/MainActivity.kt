package com.phasionary.app

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import com.phasionary.app.ui.navigation.PhasApp
import com.phasionary.app.ui.theme.PhasionaryTheme

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            PhasionaryTheme {
                PhasApp()
            }
        }
    }
}
