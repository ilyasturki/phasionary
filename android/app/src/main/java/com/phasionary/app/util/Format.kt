package com.phasionary.app.util

/**
 * Short estimate label, matching the TUI's formatEstimateShort
 * (internal/app/components/taskline.go) exactly so both surfaces read the same:
 *   < 60 min      -> "Nm"
 *   < 8  hours    -> "Nh"   (integer hours)
 *   otherwise     -> "Nd"   (a "day" is 8 hours)
 */
fun formatEstimateShort(minutes: Int): String {
    if (minutes < 60) return "${minutes}m"
    val hours = minutes / 60
    if (hours < 8) return "${hours}h"
    val days = hours / 8
    return "${days}d"
}
