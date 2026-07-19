package com.phasionary.app.ui.navigation

import androidx.compose.animation.EnterTransition
import androidx.compose.animation.ExitTransition
import androidx.compose.material3.Surface
import androidx.compose.runtime.Composable
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.phasionary.app.ui.projects.ProjectListScreen
import com.phasionary.app.ui.settings.SettingsScreen
import com.phasionary.app.ui.tasks.CATEGORY_ID_ARG
import com.phasionary.app.ui.tasks.PROJECT_ID_ARG
import com.phasionary.app.ui.tasks.TASK_ID_ARG
import com.phasionary.app.ui.tasks.TaskEditScreen
import com.phasionary.app.ui.tasks.TaskListScreen
import com.phasionary.app.ui.theme.PhasTheme

object Routes {
    const val PROJECTS = "projects"
    const val SETTINGS = "settings"
    const val TASKS_PATTERN = "tasks/{$PROJECT_ID_ARG}"
    const val TASK_EDIT_PATTERN =
        "tasks/{$PROJECT_ID_ARG}/categories/{$CATEGORY_ID_ARG}/tasks/{$TASK_ID_ARG}"

    fun tasks(projectId: String): String = "tasks/$projectId"

    fun taskEdit(projectId: String, categoryId: String, taskId: String): String =
        "tasks/$projectId/categories/$categoryId/tasks/$taskId"
}

@Composable
fun PhasApp() {
    val navController = rememberNavController()

    Surface(color = PhasTheme.colors.background) {
        NavHost(
            navController = navController,
            startDestination = Routes.PROJECTS,
            // Screens cut over instantly, like a TUI redraw — no slide, no fade.
            // Set on the host so every destination inherits it.
            enterTransition = { EnterTransition.None },
            exitTransition = { ExitTransition.None },
            popEnterTransition = { EnterTransition.None },
            popExitTransition = { ExitTransition.None },
        ) {
            composable(Routes.PROJECTS) {
                ProjectListScreen(
                    onOpenProject = { projectId -> navController.navigate(Routes.tasks(projectId)) },
                    onOpenSettings = { navController.navigate(Routes.SETTINGS) },
                )
            }
            composable(
                route = Routes.TASKS_PATTERN,
                arguments = listOf(navArgument(PROJECT_ID_ARG) { type = NavType.StringType }),
            ) { entry ->
                val projectId = entry.arguments?.getString(PROJECT_ID_ARG).orEmpty()
                TaskListScreen(
                    onBack = { navController.popBackStack() },
                    onOpenSettings = { navController.navigate(Routes.SETTINGS) },
                    onOpenTask = { categoryId, taskId ->
                        navController.navigate(Routes.taskEdit(projectId, categoryId, taskId))
                    },
                )
            }
            composable(
                route = Routes.TASK_EDIT_PATTERN,
                arguments = listOf(
                    navArgument(PROJECT_ID_ARG) { type = NavType.StringType },
                    navArgument(CATEGORY_ID_ARG) { type = NavType.StringType },
                    navArgument(TASK_ID_ARG) { type = NavType.StringType },
                ),
            ) {
                TaskEditScreen(onClose = { navController.popBackStack() })
            }
            composable(Routes.SETTINGS) {
                SettingsScreen(onDone = { navController.popBackStack() })
            }
        }
    }
}
