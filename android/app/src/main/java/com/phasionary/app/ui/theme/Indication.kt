package com.phasionary.app.ui.theme

import androidx.compose.foundation.IndicationNodeFactory
import androidx.compose.foundation.interaction.InteractionSource
import androidx.compose.foundation.interaction.PressInteraction
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.drawscope.ContentDrawScope
import androidx.compose.ui.node.DelegatableNode
import androidx.compose.ui.node.DrawModifierNode
import androidx.compose.ui.node.invalidateDraw
import androidx.compose.ui.Modifier
import kotlinx.coroutines.launch

/**
 * Press feedback with no animation: the row fills with [color] the instant a
 * finger lands and clears the instant it lifts — the touch equivalent of the
 * TUI's reversed selection bar. Material's ripple is a spreading, fading circle;
 * this app wants terminal-fast, so [PhasionaryTheme] provides this as
 * LocalIndication and turns Material's ripple off.
 */
class FlatIndication(private val color: Color) : IndicationNodeFactory {

    override fun create(interactionSource: InteractionSource): DelegatableNode =
        FlatIndicationNode(interactionSource, color)

    override fun equals(other: Any?): Boolean = other is FlatIndication && other.color == color

    override fun hashCode(): Int = color.hashCode()
}

private class FlatIndicationNode(
    private val interactionSource: InteractionSource,
    private val color: Color,
) : Modifier.Node(), DrawModifierNode {

    private var pressCount = 0
    private var pressed = false

    override fun onAttach() {
        coroutineScope.launch {
            interactionSource.interactions.collect { interaction ->
                when (interaction) {
                    is PressInteraction.Press -> pressCount++
                    // A press can end as a release or a cancel (finger dragged
                    // off); both must undo their Press or the row stays lit.
                    is PressInteraction.Release, is PressInteraction.Cancel -> pressCount--
                }
                val nowPressed = pressCount > 0
                if (nowPressed != pressed) {
                    pressed = nowPressed
                    invalidateDraw()
                }
            }
        }
    }

    override fun ContentDrawScope.draw() {
        if (pressed) drawRect(color = color)
        drawContent()
    }
}
