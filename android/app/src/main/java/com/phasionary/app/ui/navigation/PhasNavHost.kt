package com.phasionary.app.ui.navigation

import androidx.compose.material3.Surface
import androidx.compose.runtime.Composable
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.phasionary.app.ui.projects.ProjectListScreen
import com.phasionary.app.ui.settings.SettingsScreen
import com.phasionary.app.ui.tasks.PROJECT_ID_ARG
import com.phasionary.app.ui.tasks.TaskListScreen
import com.phasionary.app.ui.theme.PhasTheme

object Routes {
    const val PROJECTS = "projects"
    const val SETTINGS = "settings"
    const val TASKS_PATTERN = "tasks/{$PROJECT_ID_ARG}"
    fun tasks(projectId: String): String = "tasks/$projectId"
}

@Composable
fun PhasApp() {
    val navController = rememberNavController()

    Surface(color = PhasTheme.colors.background) {
        NavHost(navController = navController, startDestination = Routes.PROJECTS) {
            composable(Routes.PROJECTS) {
                ProjectListScreen(
                    onOpenProject = { projectId -> navController.navigate(Routes.tasks(projectId)) },
                    onOpenSettings = { navController.navigate(Routes.SETTINGS) },
                )
            }
            composable(
                route = Routes.TASKS_PATTERN,
                arguments = listOf(navArgument(PROJECT_ID_ARG) { type = NavType.StringType }),
            ) {
                TaskListScreen(
                    onBack = { navController.popBackStack() },
                    onOpenSettings = { navController.navigate(Routes.SETTINGS) },
                )
            }
            composable(Routes.SETTINGS) {
                SettingsScreen(onDone = { navController.popBackStack() })
            }
        }
    }
}
