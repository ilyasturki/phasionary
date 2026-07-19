package com.phasionary.app.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.unit.dp
import com.phasionary.app.ui.theme.PhasTheme

/**
 * A single-line text row that opens in place — the touch stand-in for the TUI's
 * inline prompt. It focuses itself on appearance and commits on the keyboard's
 * Done action. Two shapes, set by the flags:
 *
 *  - **capture** (defaults): starts empty, blank input closes the row, and a
 *    commit clears the field but keeps it open so several entries can be typed
 *    in a row.
 *  - **rename** ([initialValue] set, [commitBlank], [closeOnSubmit]): starts
 *    from the current text, a commit closes the row, and blank commits rather
 *    than dismissing — which is how a separator's label is cleared back to a
 *    plain rule.
 */
@Composable
fun InlineInputRow(
    placeholder: String,
    onSubmit: (String) -> Unit,
    onDismiss: () -> Unit,
    modifier: Modifier = Modifier,
    initialValue: String = "",
    commitBlank: Boolean = false,
    closeOnSubmit: Boolean = false,
    onDelete: (() -> Unit)? = null,
) {
    val colors = PhasTheme.colors
    var text by remember { mutableStateOf(initialValue) }
    val focusRequester = remember { FocusRequester() }

    LaunchedEffect(Unit) { focusRequester.requestFocus() }

    val submit = {
        val value = text.trim()
        if (value.isEmpty() && !commitBlank) {
            // Pressing Done on an empty capture field means "I'm finished".
            onDismiss()
        } else {
            onSubmit(value)
            if (closeOnSubmit) onDismiss() else text = ""
        }
    }

    Row(
        modifier = modifier
            .fillMaxWidth()
            .background(colors.surface)
            .heightIn(min = 44.dp)
            .padding(start = 12.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        Text(
            text = "›",
            color = colors.accent,
            style = MaterialTheme.typography.bodyMedium,
        )
        Box(modifier = Modifier.weight(1f)) {
            if (text.isEmpty()) {
                Text(
                    text = placeholder,
                    color = colors.textMuted,
                    style = MaterialTheme.typography.bodyMedium,
                )
            }
            BasicTextField(
                value = text,
                onValueChange = { text = it.replace("\n", "") },
                singleLine = true,
                textStyle = MaterialTheme.typography.bodyMedium.copy(color = colors.textPrimary),
                cursorBrush = SolidColor(colors.accent),
                keyboardOptions = KeyboardOptions(imeAction = ImeAction.Done),
                keyboardActions = KeyboardActions(onDone = { submit() }),
                modifier = Modifier
                    .fillMaxWidth()
                    .focusRequester(focusRequester),
            )
        }
        if (onDelete != null) {
            Text(
                text = "del",
                color = colors.danger,
                style = MaterialTheme.typography.labelMedium,
                modifier = Modifier
                    .clickable(onClick = onDelete)
                    .padding(horizontal = 8.dp, vertical = 12.dp)
                    .semantics { contentDescription = "Delete" },
            )
        }
        Text(
            text = "×",
            color = colors.textMuted,
            style = MaterialTheme.typography.titleMedium,
            modifier = Modifier
                .clickable(onClick = onDismiss)
                .padding(horizontal = 12.dp, vertical = 10.dp)
                .semantics { contentDescription = "Cancel" },
        )
    }
}
