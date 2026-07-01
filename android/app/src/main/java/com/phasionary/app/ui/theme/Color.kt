package com.phasionary.app.ui.theme

import androidx.compose.ui.graphics.Color

/**
 * The Phasionary token set. The TUI leans on the terminal's ANSI palette; here
 * we pin concrete hexes (GitHub Primer-derived) chosen to read well on BOTH
 * dark and light, keeping the terminal mapping:
 *   todo=yellow · in_progress=blue · completed=muted · cancelled=red
 *   priority high=red ▲ · low=cyan ▼ · medium=none · accent=green
 */
data class PhasColors(
    val background: Color,
    val surface: Color,
    val textPrimary: Color,
    val textMuted: Color,
    val divider: Color,
    val selection: Color,
    val accent: Color,
    val statusTodo: Color,
    val statusInProgress: Color,
    val statusCompleted: Color,
    val statusCancelled: Color,
    val priorityHigh: Color,
    val priorityLow: Color,
    val danger: Color,
    val searchHighlight: Color,
    val onSearchHighlight: Color,
    val isDark: Boolean,
)

val DarkPhasColors = PhasColors(
    background = Color(0xFF0D1117),
    surface = Color(0xFF161B22),
    textPrimary = Color(0xFFC9D1D9),
    textMuted = Color(0xFF8B949E),
    divider = Color(0xFF30363D),
    selection = Color(0xFF263041),
    accent = Color(0xFF3FB950),
    statusTodo = Color(0xFFD29922),
    statusInProgress = Color(0xFF58A6FF),
    statusCompleted = Color(0xFF6E7681),
    statusCancelled = Color(0xFFF85149),
    priorityHigh = Color(0xFFF85149),
    priorityLow = Color(0xFF39C5CF),
    danger = Color(0xFFF85149),
    searchHighlight = Color(0xFFD29922),
    onSearchHighlight = Color(0xFF0D1117),
    isDark = true,
)

val LightPhasColors = PhasColors(
    background = Color(0xFFFFFFFF),
    surface = Color(0xFFF6F8FA),
    textPrimary = Color(0xFF1F2328),
    textMuted = Color(0xFF656D76),
    divider = Color(0xFFD0D7DE),
    selection = Color(0xFFEAECEF),
    accent = Color(0xFF1A7F37),
    statusTodo = Color(0xFF9A6700),
    statusInProgress = Color(0xFF0969DA),
    statusCompleted = Color(0xFF8C959F),
    statusCancelled = Color(0xFFCF222E),
    priorityHigh = Color(0xFFCF222E),
    priorityLow = Color(0xFF1B7C83),
    danger = Color(0xFFCF222E),
    searchHighlight = Color(0xFFFFF8C5),
    onSearchHighlight = Color(0xFF1F2328),
    isDark = false,
)
