package com.phasionary.app.ui.tasks

import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.AlertDialog
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
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.RectangleShape
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.selected
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.phasionary.app.data.model.Priority
import com.phasionary.app.data.model.Status
import com.phasionary.app.ui.components.ErrorView
import com.phasionary.app.ui.components.LoadingView
import com.phasionary.app.ui.theme.PhasTheme
import com.phasionary.app.ui.theme.priorityLabel
import com.phasionary.app.ui.theme.statusColor
import com.phasionary.app.ui.theme.statusGlyph
import com.phasionary.app.ui.theme.statusLabel

/**
 * The full-page task editor a task row opens. Edits are local until Save sends
 * one PATCH; leaving with unsaved changes asks first. Deleting is here too,
 * behind a confirmation, since this is the only place a task is "open".
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TaskEditScreen(
    onClose: () -> Unit,
    viewModel: TaskEditViewModel = viewModel(factory = TaskEditViewModel.Factory),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val colors = PhasTheme.colors
    val snackbarHostState = remember { SnackbarHostState() }

    var confirmDiscard by rememberSaveable { mutableStateOf(false) }
    var confirmDelete by rememberSaveable { mutableStateOf(false) }

    // A saved or deleted task has nothing left to edit — leave immediately.
    LaunchedEffect(state.done) {
        if (state.done) onClose()
    }

    LaunchedEffect(state.message) {
        val msg = state.message
        if (msg != null) {
            snackbarHostState.showSnackbar(msg)
            viewModel.consumeMessage()
        }
    }

    val attemptClose = {
        if (state.dirty) confirmDiscard = true else onClose()
    }

    BackHandler(enabled = state.dirty) { confirmDiscard = true }

    Scaffold(
        containerColor = colors.background,
        snackbarHost = { SnackbarHost(snackbarHostState) },
        topBar = {
            TopAppBar(
                title = { Text("edit task") },
                navigationIcon = {
                    TextButton(onClick = attemptClose) {
                        Text("‹", style = MaterialTheme.typography.titleLarge)
                    }
                },
                actions = {
                    TextButton(onClick = viewModel::save, enabled = state.canSave) {
                        Text(
                            text = "save",
                            color = if (state.canSave) colors.accent else colors.textMuted,
                            style = MaterialTheme.typography.labelLarge,
                        )
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
            when {
                state.loading -> LoadingView()

                state.original == null -> ErrorView(
                    message = state.error ?: "Nothing here.",
                    actionLabel = "Retry",
                    onAction = viewModel::load,
                )

                else -> TaskEditForm(
                    form = state.form,
                    onFormChange = viewModel::updateForm,
                    onInsertSeparator = viewModel::insertSeparatorBelow,
                    onDelete = { confirmDelete = true },
                )
            }
        }
    }

    if (confirmDiscard) {
        ConfirmDialog(
            title = "Discard changes?",
            body = "The edits on this task haven't been saved.",
            confirmLabel = "Discard",
            confirmColor = colors.danger,
            onConfirm = {
                confirmDiscard = false
                onClose()
            },
            onDismiss = { confirmDiscard = false },
        )
    }

    if (confirmDelete) {
        ConfirmDialog(
            title = "Delete this task?",
            body = "It will be removed from the project. This can't be undone from the phone.",
            confirmLabel = "Delete",
            confirmColor = colors.danger,
            onConfirm = {
                confirmDelete = false
                viewModel.delete()
            },
            onDismiss = { confirmDelete = false },
        )
    }
}

@Composable
private fun TaskEditForm(
    form: TaskForm,
    onFormChange: ((TaskForm) -> TaskForm) -> Unit,
    onInsertSeparator: () -> Unit,
    onDelete: () -> Unit,
) {
    val colors = PhasTheme.colors

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(horizontal = 12.dp, vertical = 8.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp),
    ) {
        Field(label = "title") {
            FlatTextField(
                value = form.title,
                onValueChange = { value -> onFormChange { it.copy(title = value) } },
                singleLine = true,
                placeholder = "task title",
            )
        }

        Field(label = "status") {
            Row(horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                Status.ALL.forEach { status ->
                    OptionToken(
                        text = statusGlyph(status),
                        description = statusLabel(status),
                        selected = form.status == status,
                        color = statusColor(status, colors),
                        onClick = { onFormChange { it.copy(status = status) } },
                    )
                }
            }
        }

        Field(label = "priority") {
            Row(
                horizontalArrangement = Arrangement.spacedBy(4.dp),
                modifier = Modifier.fillMaxWidth(),
            ) {
                Priority.ALL.forEach { priority ->
                    OptionToken(
                        text = priorityLabel(priority),
                        description = priorityLabel(priority),
                        selected = form.priority == priority,
                        color = colors.textPrimary,
                        onClick = { onFormChange { it.copy(priority = priority) } },
                        modifier = Modifier.weight(1f),
                    )
                }
            }
        }

        Field(
            label = "estimate (minutes)",
            // Whole minutes only: the API takes an int, and a bad value would
            // otherwise fail late, at save time.
            hint = if (form.estimateMinutes == null) "whole minutes, 0 or more" else null,
        ) {
            FlatTextField(
                value = form.estimateText,
                onValueChange = { value ->
                    onFormChange { it.copy(estimateText = value.filter(Char::isDigit)) }
                },
                singleLine = true,
                placeholder = "0",
                keyboardType = KeyboardType.Number,
            )
        }

        Field(label = "description") {
            FlatTextField(
                value = form.description,
                onValueChange = { value -> onFormChange { it.copy(description = value) } },
                singleLine = false,
                placeholder = "notes",
                minHeight = 120.dp,
            )
        }

        // The TUI inserts a separator below the selection with `-`; on touch,
        // "the selection" is whichever task you have open.
        Text(
            text = "insert separator below",
            color = colors.textMuted,
            style = MaterialTheme.typography.labelLarge,
            modifier = Modifier
                .fillMaxWidth()
                .border(1.dp, colors.divider, RectangleShape)
                .clickable(onClick = onInsertSeparator)
                .padding(vertical = 12.dp),
            textAlign = TextAlign.Center,
        )

        Text(
            text = "delete task",
            color = colors.danger,
            style = MaterialTheme.typography.labelLarge,
            modifier = Modifier
                .fillMaxWidth()
                .border(1.dp, colors.danger, RectangleShape)
                .clickable(onClick = onDelete)
                .padding(vertical = 12.dp),
            textAlign = TextAlign.Center,
        )
    }
}

@Composable
private fun Field(
    label: String,
    hint: String? = null,
    content: @Composable () -> Unit,
) {
    val colors = PhasTheme.colors
    Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
        Text(
            text = label,
            color = colors.textMuted,
            style = MaterialTheme.typography.labelMedium,
        )
        content()
        if (hint != null) {
            Text(
                text = hint,
                color = colors.danger,
                style = MaterialTheme.typography.labelSmall,
            )
        }
    }
}

@Composable
private fun FlatTextField(
    value: String,
    onValueChange: (String) -> Unit,
    singleLine: Boolean,
    placeholder: String,
    keyboardType: KeyboardType = KeyboardType.Text,
    minHeight: Dp = 44.dp,
) {
    val colors = PhasTheme.colors
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .border(1.dp, colors.divider, RectangleShape)
            .background(colors.surface)
            .heightIn(min = minHeight)
            .padding(horizontal = 8.dp, vertical = 10.dp),
    ) {
        if (value.isEmpty()) {
            Text(
                text = placeholder,
                color = colors.textMuted,
                style = MaterialTheme.typography.bodyMedium,
            )
        }
        BasicTextField(
            value = value,
            onValueChange = onValueChange,
            singleLine = singleLine,
            textStyle = MaterialTheme.typography.bodyMedium.copy(color = colors.textPrimary),
            cursorBrush = SolidColor(colors.accent),
            keyboardOptions = KeyboardOptions(
                keyboardType = keyboardType,
                imeAction = if (singleLine) ImeAction.Done else ImeAction.Default,
            ),
            modifier = Modifier.fillMaxWidth(),
        )
    }
}

/** A flat, square selectable token — the picker idiom for status and priority. */
@Composable
private fun OptionToken(
    text: String,
    description: String,
    selected: Boolean,
    color: Color,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val colors = PhasTheme.colors
    Text(
        text = text,
        color = if (selected) colors.textPrimary else colors.textMuted,
        style = MaterialTheme.typography.bodyMedium,
        maxLines = 1,
        textAlign = TextAlign.Center,
        modifier = modifier
            .border(1.dp, if (selected) color else colors.divider, RectangleShape)
            .background(if (selected) colors.selection else colors.background)
            .clickable(onClick = onClick)
            .heightIn(min = 44.dp)
            .padding(horizontal = 8.dp, vertical = 12.dp)
            // The status tokens are glyphs, so the spoken name has to come from
            // semantics rather than from the text itself.
            .semantics {
                contentDescription = description
                this.selected = selected
            },
    )
}

/**
 * A flat two-button confirmation. Shared with the task list, which needs the
 * same "this can't be undone from the phone" prompt for deleting a separator.
 */
@Composable
fun ConfirmDialog(
    title: String,
    body: String,
    confirmLabel: String,
    confirmColor: Color,
    onConfirm: () -> Unit,
    onDismiss: () -> Unit,
) {
    val colors = PhasTheme.colors
    AlertDialog(
        onDismissRequest = onDismiss,
        containerColor = colors.surface,
        titleContentColor = colors.textPrimary,
        textContentColor = colors.textMuted,
        title = { Text(title, style = MaterialTheme.typography.titleMedium) },
        text = { Text(body, style = MaterialTheme.typography.bodyMedium) },
        confirmButton = {
            TextButton(onClick = onConfirm) {
                Text(confirmLabel, color = confirmColor)
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("Cancel", color = colors.textMuted)
            }
        },
    )
}
