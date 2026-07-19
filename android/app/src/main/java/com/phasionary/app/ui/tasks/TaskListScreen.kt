package com.phasionary.app.ui.tasks

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
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
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LifecycleEventEffect
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.phasionary.app.data.model.Project
import com.phasionary.app.data.model.Task
import com.phasionary.app.data.model.statusCounts
import com.phasionary.app.ui.components.CategoryHeader
import com.phasionary.app.ui.components.EmptyView
import com.phasionary.app.ui.components.ErrorView
import com.phasionary.app.ui.components.InlineInputRow
import com.phasionary.app.ui.components.LoadingView
import com.phasionary.app.ui.components.SeparatorRow
import com.phasionary.app.ui.components.TaskRow
import com.phasionary.app.ui.theme.PhasTheme

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TaskListScreen(
    onBack: () -> Unit,
    onOpenSettings: () -> Unit,
    onOpenTask: (categoryId: String, taskId: String) -> Unit,
    viewModel: TaskListViewModel = viewModel(factory = TaskListViewModel.Factory),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val colors = PhasTheme.colors
    val snackbarHostState = remember { SnackbarHostState() }

    // Coming back from the editor (or from anywhere) re-reads the project, so an
    // edit made here — or in the TUI, on the other end — is never shown stale.
    // The first resume is skipped: the ViewModel's own init already loaded, and
    // refreshing on top of it would just double the request on every entry.
    var initialResume by remember { mutableStateOf(true) }
    LifecycleEventEffect(Lifecycle.Event.ON_RESUME) {
        if (initialResume) initialResume = false else viewModel.refresh()
    }

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
                    onOpenTask = onOpenTask,
                    onAddTask = viewModel::addTask,
                    onAddCategory = viewModel::addCategory,
                    onRenameSeparator = viewModel::renameSeparator,
                    onDeleteSeparator = viewModel::deleteRow,
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
    onOpenTask: (categoryId: String, taskId: String) -> Unit,
    onAddTask: (categoryId: String, title: String) -> Unit,
    onAddCategory: (String) -> Unit,
    onRenameSeparator: (categoryId: String, taskId: String, label: String) -> Unit,
    onDeleteSeparator: (categoryId: String, taskId: String) -> Unit,
) {
    // Which capture row is open: at most one, so the list never sprouts several
    // keyboards' worth of inputs. Null means none.
    var addingTaskIn by rememberSaveable { mutableStateOf<String?>(null) }
    var addingCategory by rememberSaveable { mutableStateOf(false) }
    // The separator being relabeled in place, and the one awaiting a delete
    // confirmation (the phone has no undo, unlike the TUI's instant delete).
    var renamingSeparator by rememberSaveable { mutableStateOf<String?>(null) }
    var confirmDeleteSeparator by rememberSaveable { mutableStateOf<Pair<String, String>?>(null) }

    LazyColumn(modifier = Modifier.fillMaxSize()) {
        project.categories.forEach { category ->
            val isExpanded = category.id !in collapsed

            item(key = "cat-${category.id}") {
                CategoryHeader(
                    name = category.name,
                    counts = category.statusCounts(),
                    expanded = isExpanded,
                    onAdd = {
                        addingCategory = false
                        addingTaskIn = category.id
                    },
                    onClick = { onToggleCategory(category.id) },
                )
            }

            // The capture row sits under its own header even when the category
            // is collapsed — tapping "+" should always give you somewhere to type.
            if (addingTaskIn == category.id) {
                item(key = "add-task-${category.id}") {
                    InlineInputRow(
                        placeholder = "new task",
                        onSubmit = { title -> onAddTask(category.id, title) },
                        onDismiss = { addingTaskIn = null },
                    )
                }
            }

            if (isExpanded) {
                if (category.tasks.isEmpty()) {
                    item(key = "empty-${category.id}") { EmptyCategoryHint() }
                } else {
                    items(items = category.tasks, key = { "task-${it.id}" }) { task ->
                        when {
                            // A separator has only a label, so it gets an inline
                            // rename row rather than the task editor page.
                            task.isSeparator && renamingSeparator == task.id -> InlineInputRow(
                                placeholder = "separator label",
                                initialValue = task.title,
                                commitBlank = true,
                                closeOnSubmit = true,
                                onSubmit = { label ->
                                    onRenameSeparator(category.id, task.id, label)
                                },
                                onDismiss = { renamingSeparator = null },
                                onDelete = {
                                    confirmDeleteSeparator = category.id to task.id
                                },
                            )

                            task.isSeparator -> SeparatorRow(
                                label = task.title,
                                onClick = {
                                    addingTaskIn = null
                                    addingCategory = false
                                    renamingSeparator = task.id
                                },
                            )

                            else -> TaskRow(
                                task = task,
                                onClick = { onOpenTask(category.id, task.id) },
                                onStatusClick = { onCycleStatus(category.id, task) },
                            )
                        }
                    }
                }
            }
        }

        item(key = "add-category") {
            if (addingCategory) {
                InlineInputRow(
                    placeholder = "new category",
                    onSubmit = onAddCategory,
                    onDismiss = { addingCategory = false },
                )
            } else {
                AddCategoryRow(
                    onClick = {
                        addingTaskIn = null
                        addingCategory = true
                    },
                )
            }
        }
    }

    // Outside the LazyColumn: an item that scrolls out of view is disposed, and
    // the dialog would vanish with it.
    val pending = confirmDeleteSeparator
    if (pending != null) {
        ConfirmDialog(
            title = "Delete this separator?",
            body = "The divider is removed; the tasks around it stay. " +
                "This can't be undone from the phone.",
            confirmLabel = "Delete",
            confirmColor = PhasTheme.colors.danger,
            onConfirm = {
                confirmDeleteSeparator = null
                renamingSeparator = null
                onDeleteSeparator(pending.first, pending.second)
            },
            onDismiss = { confirmDeleteSeparator = null },
        )
    }
}

@Composable
private fun AddCategoryRow(onClick: () -> Unit) {
    Text(
        text = "+ add category",
        color = PhasTheme.colors.textMuted,
        style = MaterialTheme.typography.labelMedium,
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .heightIn(min = 44.dp)
            .padding(horizontal = 12.dp, vertical = 14.dp),
    )
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
