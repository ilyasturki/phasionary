package com.phasionary.app.util

import org.junit.Assert.assertEquals
import org.junit.Test

/**
 * Expected values come from the TUI's formatEstimateShort spec (a "day" is 8
 * hours; integer division at each step) — an independent source of truth, not a
 * re-derivation of this implementation.
 */
class EstimateFormatTest {

    @Test
    fun minutesUnderAnHour() {
        assertEquals("0m", formatEstimateShort(0))
        assertEquals("30m", formatEstimateShort(30))
        assertEquals("59m", formatEstimateShort(59))
    }

    @Test
    fun hoursUnderEight() {
        assertEquals("1h", formatEstimateShort(60))
        assertEquals("1h", formatEstimateShort(90))
        assertEquals("2h", formatEstimateShort(120))
        assertEquals("7h", formatEstimateShort(479))
    }

    @Test
    fun eightHoursOrMoreBecomeDays() {
        assertEquals("1d", formatEstimateShort(480))
        assertEquals("2d", formatEstimateShort(960))
        assertEquals("3d", formatEstimateShort(1440))
        assertEquals("5d", formatEstimateShort(2400))
    }
}
