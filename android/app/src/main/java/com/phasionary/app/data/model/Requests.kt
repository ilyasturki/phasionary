package com.phasionary.app.data.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/**
 * Quick-capture body for POST .../tasks. Nullable optionals + the client's
 * explicitNulls=false Json mean unset fields are simply omitted, so the server
 * applies its own defaults (status -> todo, priority -> none, estimate -> 0).
 */
@Serializable
data class CreateTaskBody(
    val title: String,
    val status: String? = null,
    val priority: String? = null,
    @SerialName("estimate_minutes") val estimateMinutes: Int? = null,
    val description: String? = null,
)

/** Body for POST .../status — an explicit set (the API does not cycle). */
@Serializable
data class StatusBody(
    val status: String,
)

/** The server's consistent error shape: {"error": "..."}. */
@Serializable
data class ErrorEnvelope(
    val error: String = "",
)
