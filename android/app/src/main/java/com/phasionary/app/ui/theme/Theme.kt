package com.phasionary.app.ui.theme

import androidx.compose.foundation.LocalIndication
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.LocalRippleConfiguration
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Shapes
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.ReadOnlyComposable
import androidx.compose.runtime.remember
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp

val LocalPhasColors = staticCompositionLocalOf { LightPhasColors }

/** Ergonomic accessor: PhasTheme.colors.statusTodo, etc. */
object PhasTheme {
    val colors: PhasColors
        @Composable
        @ReadOnlyComposable
        get() = LocalPhasColors.current
}

// Sharp edges everywhere — no rounded corners (design system rule). This also
// squares off Material components (FAB, bottom sheet) to match the TUI look.
// Material3's Shapes slots require a CornerBasedShape, so a zero-radius rounded
// corner is how we express "square" here (RectangleShape wouldn't type-check).
private val Square = RoundedCornerShape(0.dp)
private val FlatShapes = Shapes(
    extraSmall = Square,
    small = Square,
    medium = Square,
    large = Square,
    extraLarge = Square,
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PhasionaryTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    content: @Composable () -> Unit,
) {
    val phas = if (darkTheme) DarkPhasColors else LightPhasColors

    // Feed the tokens into a Material scheme so Material components (Scaffold,
    // FAB, ModalBottomSheet, TextField) inherit the same palette.
    val scheme = if (darkTheme) {
        darkColorScheme(
            primary = phas.accent,
            onPrimary = phas.background,
            secondary = phas.statusInProgress,
            background = phas.background,
            onBackground = phas.textPrimary,
            surface = phas.surface,
            onSurface = phas.textPrimary,
            surfaceVariant = phas.surface,
            onSurfaceVariant = phas.textMuted,
            error = phas.danger,
            onError = phas.background,
            outline = phas.divider,
        )
    } else {
        lightColorScheme(
            primary = phas.accent,
            onPrimary = Color.White,
            secondary = phas.statusInProgress,
            background = phas.background,
            onBackground = phas.textPrimary,
            surface = phas.surface,
            onSurface = phas.textPrimary,
            surfaceVariant = phas.surface,
            onSurfaceVariant = phas.textMuted,
            error = phas.danger,
            onError = Color.White,
            outline = phas.divider,
        )
    }

    // Kill every default animation the design system doesn't want:
    // LocalIndication swaps the press effect under Modifier.clickable, and a
    // null LocalRippleConfiguration disables ripple inside Material components
    // (which reach for it directly rather than through LocalIndication).
    val indication = remember(phas.selection) { FlatIndication(phas.selection) }

    CompositionLocalProvider(
        LocalPhasColors provides phas,
        LocalIndication provides indication,
        LocalRippleConfiguration provides null,
    ) {
        MaterialTheme(
            colorScheme = scheme,
            typography = PhasTypography,
            shapes = FlatShapes,
            content = content,
        )
    }
}
