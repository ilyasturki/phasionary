package com.phasionary.app.ui.components

import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.RectangleShape
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import com.phasionary.app.data.model.Status
import com.phasionary.app.ui.theme.PhasTheme
import com.phasionary.app.ui.theme.PhasionaryTheme
import com.phasionary.app.ui.theme.statusColor
import com.phasionary.app.ui.theme.statusLabel

/**
 * A bordered, always-labeled status token (no color-only states). Tapping it is
 * the touch equivalent of the TUI's `s` — the caller decides what "next" means.
 */
@Composable
fun StatusChip(
    status: String,
    modifier: Modifier = Modifier,
    onClick: (() -> Unit)? = null,
) {
    val color = statusColor(status, PhasTheme.colors)
    val chip = modifier
        .then(if (onClick != null) Modifier.clickable(onClick = onClick) else Modifier)
        .border(1.dp, color, RectangleShape)
        .padding(horizontal = 6.dp, vertical = 2.dp)
    Text(
        text = statusLabel(status),
        color = color,
        maxLines = 1,
        style = MaterialTheme.typography.labelSmall,
        modifier = chip,
    )
}

@Preview
@Composable
private fun StatusChipPreview() {
    PhasionaryTheme {
        StatusChip(status = Status.IN_PROGRESS)
    }
}
