package com.phasionary.app.ui.theme

import androidx.compose.material3.Typography
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.sp

// Monospace everywhere (the terminal aesthetic). Uses the system monospace font
// so no font asset ships with v1; a JetBrains Mono .ttf can be dropped into
// res/font and referenced here later without other changes.
private val Mono = FontFamily.Monospace

val PhasTypography = Typography(
    bodyLarge = TextStyle(fontFamily = Mono, fontWeight = FontWeight.Normal, fontSize = 15.sp, lineHeight = 22.sp),
    bodyMedium = TextStyle(fontFamily = Mono, fontWeight = FontWeight.Normal, fontSize = 14.sp, lineHeight = 20.sp),
    bodySmall = TextStyle(fontFamily = Mono, fontWeight = FontWeight.Normal, fontSize = 12.sp, lineHeight = 16.sp),
    titleLarge = TextStyle(fontFamily = Mono, fontWeight = FontWeight.Bold, fontSize = 18.sp, lineHeight = 24.sp),
    titleMedium = TextStyle(fontFamily = Mono, fontWeight = FontWeight.Bold, fontSize = 16.sp, lineHeight = 22.sp),
    labelLarge = TextStyle(fontFamily = Mono, fontWeight = FontWeight.Medium, fontSize = 13.sp, lineHeight = 16.sp),
    labelMedium = TextStyle(fontFamily = Mono, fontWeight = FontWeight.Normal, fontSize = 12.sp, lineHeight = 16.sp),
    labelSmall = TextStyle(fontFamily = Mono, fontWeight = FontWeight.Normal, fontSize = 11.sp, lineHeight = 14.sp),
)
