package com.phasionary.app.ui.theme

import androidx.compose.ui.graphics.Color
import com.phasionary.app.data.model.Priority
import com.phasionary.app.data.model.Status

/**
 * Display mappers shared by every component so status/priority always render the
 * same way. Labels always carry text (no color-only states) per the design doc.
 */

fun statusLabel(status: String): String = when (status) {
    Status.TODO -> "todo"
    Status.IN_PROGRESS -> "in progress"
    Status.COMPLETED -> "completed"
    Status.CANCELLED -> "cancelled"
    else -> status
}

fun statusColor(status: String, colors: PhasColors): Color = when (status) {
    Status.TODO -> colors.statusTodo
    Status.IN_PROGRESS -> colors.statusInProgress
    Status.COMPLETED -> colors.statusCompleted
    Status.CANCELLED -> colors.statusCancelled
    else -> colors.textMuted
}

/** The triangle marker the TUI uses; empty for medium/none. */
fun priorityGlyph(priority: String): String = when (priority) {
    Priority.HIGH -> "▲"
    Priority.LOW -> "▼"
    else -> ""
}

/** Color for the priority glyph / title tint; null when there is no tint. */
fun priorityColor(priority: String, colors: PhasColors): Color? = when (priority) {
    Priority.HIGH -> colors.priorityHigh
    Priority.LOW -> colors.priorityLow
    else -> null
}

/** Completed/cancelled tasks read "dimmed", matching the TUI's Faint treatment. */
fun isMuted(status: String): Boolean =
    status == Status.COMPLETED || status == Status.CANCELLED
