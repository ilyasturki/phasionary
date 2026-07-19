package com.phasionary.app.ui.components

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import com.phasionary.app.data.model.Status
import com.phasionary.app.ui.theme.PhasTheme
import com.phasionary.app.ui.theme.PhasionaryTheme
import com.phasionary.app.ui.theme.statusColor
import com.phasionary.app.ui.theme.statusGlyph
import com.phasionary.app.ui.theme.statusLabel

/**
 * The TUI's status marker — `[ ]` `[/]` `[x]` `[-]` — colored, borderless, and
 * a fixed three characters wide in the app's monospace face, so a column of
 * tasks aligns. Tapping it is the touch equivalent of the TUI's `s`; the caller
 * decides what "next" means.
 *
 * The glyph carries the meaning visually, so the spelled-out status rides along
 * as the accessibility label rather than being dropped.
 */
@Composable
fun StatusChip(
    status: String,
    modifier: Modifier = Modifier,
    onClick: (() -> Unit)? = null,
) {
    Text(
        text = statusGlyph(status),
        color = statusColor(status, PhasTheme.colors),
        maxLines = 1,
        style = MaterialTheme.typography.bodyMedium,
        modifier = modifier
            .then(if (onClick != null) Modifier.clickable(onClick = onClick) else Modifier)
            // Padding, not size: it widens the tap target without pushing the
            // glyph off the row's baseline grid.
            .padding(horizontal = 6.dp, vertical = 10.dp)
            .semantics { contentDescription = statusLabel(status) },
    )
}

@Preview
@Composable
private fun StatusChipPreview() {
    PhasionaryTheme {
        StatusChip(status = Status.IN_PROGRESS)
    }
}
