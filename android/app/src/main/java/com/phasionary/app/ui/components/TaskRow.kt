package com.phasionary.app.ui.components

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import com.phasionary.app.data.model.Priority
import com.phasionary.app.data.model.Status
import com.phasionary.app.data.model.Task
import com.phasionary.app.ui.theme.PhasTheme
import com.phasionary.app.ui.theme.PhasionaryTheme
import com.phasionary.app.ui.theme.isMuted
import com.phasionary.app.ui.theme.priorityColor
import com.phasionary.app.ui.theme.priorityGlyph
import com.phasionary.app.util.formatEstimateShort

/**
 * The one-line dense task row, mirroring the TUI:
 *   [status] ▲ Title … ¶ ~30m
 * Title is tinted by priority (high=red, low=cyan) and dimmed when the task is
 * completed/cancelled. Trailing glyphs: ¶ if it has a description, ~est if it
 * has an estimate.
 */
@Composable
fun TaskRow(
    task: Task,
    modifier: Modifier = Modifier,
    onClick: () -> Unit = {},
    onStatusClick: () -> Unit = {},
) {
    val colors = PhasTheme.colors
    val muted = isMuted(task.status)
    val pColor = priorityColor(task.priority, colors)
    val titleColor = when {
        muted -> colors.textMuted
        else -> pColor ?: colors.textPrimary
    }
    val glyph = priorityGlyph(task.priority)

    Row(
        modifier = modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .heightIn(min = 44.dp)
            .padding(horizontal = 12.dp, vertical = 6.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        StatusChip(status = task.status, onClick = onStatusClick)

        if (glyph.isNotEmpty()) {
            Text(
                text = glyph,
                color = if (muted) colors.textMuted else (pColor ?: colors.textMuted),
                style = MaterialTheme.typography.bodyMedium,
            )
        }

        Text(
            text = task.title,
            color = titleColor,
            style = MaterialTheme.typography.bodyMedium,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier.weight(1f),
        )

        if (task.description.isNotBlank()) {
            Text(
                text = "¶",
                color = colors.textMuted,
                style = MaterialTheme.typography.labelSmall,
            )
        }
        if (task.estimateMinutes > 0) {
            Text(
                text = "~" + formatEstimateShort(task.estimateMinutes),
                color = colors.textMuted,
                style = MaterialTheme.typography.labelSmall,
            )
        }
    }
}

/**
 * A read-only in-category divider, mirroring the TUI separator: a rule spanning
 * the row, or "── label ──" when the separator carries a label. Not clickable.
 */
@Composable
fun SeparatorRow(
    label: String,
    modifier: Modifier = Modifier,
) {
    val colors = PhasTheme.colors
    val trimmed = label.trim()
    Row(
        modifier = modifier
            .fillMaxWidth()
            .heightIn(min = 32.dp)
            .padding(horizontal = 12.dp, vertical = 6.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        if (trimmed.isEmpty()) {
            HorizontalDivider(modifier = Modifier.weight(1f), color = colors.textMuted)
        } else {
            HorizontalDivider(modifier = Modifier.width(16.dp), color = colors.textMuted)
            Text(
                text = trimmed,
                color = colors.textMuted,
                style = MaterialTheme.typography.labelMedium,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            HorizontalDivider(modifier = Modifier.weight(1f), color = colors.textMuted)
        }
    }
}

@Preview(showBackground = true)
@Composable
private fun TaskRowPreview() {
    PhasionaryTheme {
        TaskRow(
            task = Task(
                id = "1",
                title = "Wire the task list to the API",
                status = Status.IN_PROGRESS,
                priority = Priority.HIGH,
                estimateMinutes = 90,
                description = "notes",
            ),
        )
    }
}
