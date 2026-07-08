package com.phasionary.app.ui.tasks

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontStyle
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.phasionary.app.data.model.Project
import com.phasionary.app.data.model.Task
import com.phasionary.app.data.model.statusCounts
import com.phasionary.app.ui.components.CategoryHeader
import com.phasionary.app.ui.components.EmptyView
import com.phasionary.app.ui.components.ErrorView
import com.phasionary.app.ui.components.LoadingView
import com.phasionary.app.ui.components.SeparatorRow
import com.phasionary.app.ui.components.TaskRow
import com.phasionary.app.ui.theme.PhasTheme

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TaskListScreen(
    onBack: () -> Unit,
    onOpenSettings: () -> Unit,
    viewModel: TaskListViewModel = viewModel(factory = TaskListViewModel.Factory),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val colors = PhasTheme.colors
    val snackbarHostState = remember { SnackbarHostState() }

    LaunchedEffect(state.message) {
        val msg = state.message
        if (msg != null) {
            snackbarHostState.showSnackbar(msg)
            viewModel.consumeMessage()
        }
    }

    Scaffold(
        containerColor = colors.background,
        snackbarHost = { SnackbarHost(snackbarHostState) },
        topBar = {
            TopAppBar(
                title = {
                    Text(
                        text = state.project?.name?.ifBlank { "Tasks" } ?: "Tasks",
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                    )
                },
                navigationIcon = {
                    TextButton(onClick = onBack) {
                        Text("‹", style = MaterialTheme.typography.titleLarge)
                    }
                },
                actions = {
                    TextButton(onClick = onOpenSettings) {
                        Text("⚙", style = MaterialTheme.typography.titleMedium)
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = colors.surface,
                    titleContentColor = colors.textPrimary,
                    navigationIconContentColor = colors.textPrimary,
                    actionIconContentColor = colors.textPrimary,
                ),
            )
        },
    ) { padding ->
        Box(
            modifier = Modifier
                .padding(padding)
                .fillMaxSize(),
        ) {
            val project = state.project
            when {
                state.loading && project == null -> LoadingView()

                state.error != null && project == null -> ErrorView(
                    message = state.error!!,
                    actionLabel = if (state.notConfigured) "Open Settings" else "Retry",
                    onAction = if (state.notConfigured) onOpenSettings else viewModel::load,
                )

                project == null -> EmptyView("Nothing here.")

                else -> TaskListContent(
                    project = project,
                    collapsed = state.collapsed,
                    onToggleCategory = viewModel::toggleCategory,
                    onCycleStatus = viewModel::cycleStatus,
                )
            }
        }
    }
}

@Composable
private fun TaskListContent(
    project: Project,
    collapsed: Set<String>,
    onToggleCategory: (String) -> Unit,
    onCycleStatus: (String, Task) -> Unit,
) {
    if (project.categories.isEmpty()) {
        EmptyView("No categories yet.")
        return
    }

    // Tap a row to reveal its description inline (kept local — not worth
    // persisting across process death for v1).
    val expanded = remember { mutableStateListOf<String>() }

    LazyColumn(modifier = Modifier.fillMaxSize()) {
        project.categories.forEach { category ->
            val isExpanded = category.id !in collapsed

            item(key = "cat-${category.id}") {
                CategoryHeader(
                    name = category.name,
                    counts = category.statusCounts(),
                    expanded = isExpanded,
                    onClick = { onToggleCategory(category.id) },
                )
            }

            if (isExpanded) {
                if (category.tasks.isEmpty()) {
                    item(key = "empty-${category.id}") { EmptyCategoryHint() }
                } else {
                    items(items = category.tasks, key = { "task-${it.id}" }) { task ->
                        if (task.isSeparator) {
                            SeparatorRow(label = task.title)
                        } else {
                            Column {
                                TaskRow(
                                    task = task,
                                    onClick = {
                                        if (!expanded.remove(task.id)) expanded.add(task.id)
                                    },
                                    onStatusClick = { onCycleStatus(category.id, task) },
                                )
                                if (task.id in expanded && task.description.isNotBlank()) {
                                    DescriptionBlock(task.description)
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun EmptyCategoryHint() {
    Text(
        text = "(no tasks)",
        color = PhasTheme.colors.textMuted,
        style = MaterialTheme.typography.labelMedium,
        modifier = Modifier.padding(horizontal = 24.dp, vertical = 6.dp),
    )
}

@Composable
private fun DescriptionBlock(text: String) {
    val colors = PhasTheme.colors
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(start = 12.dp, end = 12.dp, bottom = 8.dp),
    ) {
        Text("▎", color = colors.textMuted, style = MaterialTheme.typography.bodySmall)
        Spacer(Modifier.width(6.dp))
        Text(
            text = text,
            color = colors.textMuted,
            style = MaterialTheme.typography.bodySmall.copy(fontStyle = FontStyle.Italic),
        )
    }
}
