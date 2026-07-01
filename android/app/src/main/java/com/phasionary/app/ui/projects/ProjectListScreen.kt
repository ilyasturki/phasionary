package com.phasionary.app.ui.projects

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.phasionary.app.data.model.Project
import com.phasionary.app.data.model.statusCounts
import com.phasionary.app.ui.components.EmptyView
import com.phasionary.app.ui.components.ErrorView
import com.phasionary.app.ui.components.LoadingView
import com.phasionary.app.ui.theme.PhasTheme

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ProjectListScreen(
    onOpenProject: (String) -> Unit,
    onOpenSettings: () -> Unit,
    viewModel: ProjectListViewModel = viewModel(factory = ProjectListViewModel.Factory),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val colors = PhasTheme.colors

    Scaffold(
        containerColor = colors.background,
        topBar = {
            TopAppBar(
                title = { Text("Phasionary") },
                actions = {
                    TextButton(onClick = onOpenSettings) {
                        Text("⚙", style = MaterialTheme.typography.titleMedium)
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = colors.surface,
                    titleContentColor = colors.textPrimary,
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
            when (val s = state) {
                is ProjectListUiState.Loading -> LoadingView()

                is ProjectListUiState.Error -> ErrorView(
                    message = s.message,
                    actionLabel = if (s.notConfigured) "Open Settings" else "Retry",
                    onAction = if (s.notConfigured) onOpenSettings else viewModel::load,
                )

                is ProjectListUiState.Success -> {
                    if (s.projects.isEmpty()) {
                        EmptyView("No projects yet.")
                    } else {
                        LazyColumn(modifier = Modifier.fillMaxSize()) {
                            items(items = s.projects, key = { it.id }) { project ->
                                ProjectRow(project = project, onClick = { onOpenProject(project.id) })
                                HorizontalDivider(color = colors.divider)
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun ProjectRow(project: Project, onClick: () -> Unit) {
    val colors = PhasTheme.colors
    val counts = project.statusCounts()
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .padding(horizontal = 12.dp, vertical = 10.dp),
    ) {
        Text(
            text = project.name.ifBlank { "(untitled)" },
            color = colors.textPrimary,
            style = MaterialTheme.typography.titleMedium,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )
        Text(
            text = "${project.categories.size} categories · ${counts.open} open · ${counts.total} tasks",
            color = colors.textMuted,
            style = MaterialTheme.typography.labelMedium,
        )
    }
}
