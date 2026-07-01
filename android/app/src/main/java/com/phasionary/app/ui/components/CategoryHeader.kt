package com.phasionary.app.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import com.phasionary.app.data.model.StatusCounts
import com.phasionary.app.ui.theme.PhasTheme
import com.phasionary.app.ui.theme.PhasionaryTheme

/**
 * A collapsible category group header: a caret, the name, and a done/total
 * tally. Tapping toggles collapse (the TUI's h/l on a category group).
 */
@Composable
fun CategoryHeader(
    name: String,
    counts: StatusCounts,
    expanded: Boolean,
    modifier: Modifier = Modifier,
    onClick: () -> Unit,
) {
    val colors = PhasTheme.colors
    Row(
        modifier = modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .background(colors.surface)
            .heightIn(min = 40.dp)
            .padding(horizontal = 12.dp, vertical = 6.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        Text(
            text = if (expanded) "▾" else "▸",
            color = colors.textMuted,
            style = MaterialTheme.typography.titleMedium,
        )
        Text(
            text = name,
            color = colors.textPrimary,
            style = MaterialTheme.typography.titleMedium,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier.weight(1f),
        )
        Text(
            text = "${counts.done}/${counts.total}",
            color = colors.textMuted,
            style = MaterialTheme.typography.labelMedium,
        )
    }
}

@Preview(showBackground = true)
@Composable
private fun CategoryHeaderPreview() {
    PhasionaryTheme {
        CategoryHeader(
            name = "Feature",
            counts = StatusCounts(todo = 3, inProgress = 1, completed = 2),
            expanded = true,
            onClick = {},
        )
    }
}
