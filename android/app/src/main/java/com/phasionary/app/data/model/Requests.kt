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
    /** "separator" for a divider; omitted for an ordinary task. */
    val kind: String? = null,
    /** Places the new row directly below this task; omitted appends. */
    @SerialName("insert_after") val insertAfter: String? = null,
)

/**
 * Partial-edit body for PATCH .../tasks/{tid}. The same explicitNulls=false rule
 * gives us exactly what the server's pointer fields expect: a null field is
 * omitted and left untouched, while an explicit "" or 0 clears it. The editor
 * sends only the fields the user actually changed.
 */
@Serializable
data class UpdateTaskBody(
    val title: String? = null,
    val status: String? = null,
    val priority: String? = null,
    @SerialName("estimate_minutes") val estimateMinutes: Int? = null,
    val description: String? = null,
) {
    /** True when the body would touch nothing — skip the request entirely. */
    val isEmpty: Boolean
        get() = title == null && status == null && priority == null &&
            estimateMinutes == null && description == null
}

/** Body for POST .../categories. */
@Serializable
data class CreateCategoryBody(
    val name: String,
)

/**
 * GET/PUT .../folds — the complete set of collapsed category IDs for a project,
 * shared with the TUI through the server's state.json. PUT replaces the whole
 * list. No default on the field: it must survive encodeDefaults=false so an
 * empty list still goes out as [] (which unfolds everything) rather than being
 * dropped from the body.
 */
@Serializable
data class FoldsBody(
    @SerialName("folded_categories") val foldedCategories: List<String>,
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
