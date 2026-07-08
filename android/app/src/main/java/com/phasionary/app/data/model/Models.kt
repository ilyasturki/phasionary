package com.phasionary.app.data.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/**
 * Wire models mirroring the Go domain JSON (see internal/domain/types.go). Field
 * names use @SerialName to match the server's snake_case tags. Every optional
 * field carries a default so a payload that omits it (the server uses omitempty)
 * still decodes.
 */
@Serializable
data class Project(
    val id: String,
    val name: String = "",
    @SerialName("created_at") val createdAt: String = "",
    @SerialName("updated_at") val updatedAt: String = "",
    val categories: List<Category> = emptyList(),
)

@Serializable
data class Category(
    val id: String,
    val name: String = "",
    @SerialName("created_at") val createdAt: String = "",
    @SerialName("updated_at") val updatedAt: String = "",
    @SerialName("estimate_minutes") val estimateMinutes: Int = 0,
    val tasks: List<Task> = emptyList(),
)

@Serializable
data class Task(
    val id: String,
    val title: String = "",
    val status: String = Status.TODO,
    @SerialName("created_at") val createdAt: String = "",
    @SerialName("updated_at") val updatedAt: String = "",
    val priority: String = Priority.NONE,
    @SerialName("completion_date") val completionDate: String = "",
    @SerialName("estimate_minutes") val estimateMinutes: Int = 0,
    val description: String = "",
    val kind: String = "",
) {
    /** True when this row is a separator/divider rather than a real task. */
    val isSeparator: Boolean get() = kind == Kind.SEPARATOR
}

/** Task kind wire values (must match domain.Kind* constants). "" = ordinary task. */
object Kind {
    const val SEPARATOR = "separator"
}

/** Status wire values (must match domain.Status* constants). */
object Status {
    const val TODO = "todo"
    const val IN_PROGRESS = "in_progress"
    const val COMPLETED = "completed"
    const val CANCELLED = "cancelled"

    val ALL = listOf(TODO, IN_PROGRESS, COMPLETED, CANCELLED)

    fun isValid(value: String): Boolean = value in ALL

    /** Forward cycle, matching domain.Task.CycleStatus: a chip tap advances to
     *  the next status (todo -> in_progress -> completed -> cancelled -> todo). */
    fun next(status: String): String = when (status) {
        TODO -> IN_PROGRESS
        IN_PROGRESS -> COMPLETED
        COMPLETED -> CANCELLED
        CANCELLED -> TODO
        else -> TODO
    }
}

/** Priority wire values (must match domain.Priority* constants). "" = none. */
object Priority {
    const val HIGH = "high"
    const val MEDIUM = "medium"
    const val LOW = "low"
    const val NONE = ""

    val ALL = listOf(HIGH, MEDIUM, LOW, NONE)
}

/** Per-category status tally, mirroring domain.StatusCounts. */
data class StatusCounts(
    val todo: Int = 0,
    val inProgress: Int = 0,
    val completed: Int = 0,
    val cancelled: Int = 0,
) {
    val total: Int get() = todo + inProgress + completed + cancelled
    val open: Int get() = todo + inProgress
    val done: Int get() = completed + cancelled
}

fun Category.statusCounts(): StatusCounts {
    var todo = 0
    var inProgress = 0
    var completed = 0
    var cancelled = 0
    for (task in tasks) {
        when (task.status) {
            Status.TODO -> todo++
            Status.IN_PROGRESS -> inProgress++
            Status.COMPLETED -> completed++
            Status.CANCELLED -> cancelled++
        }
    }
    return StatusCounts(todo, inProgress, completed, cancelled)
}

/** Project-wide tally across every category. */
fun Project.statusCounts(): StatusCounts {
    var todo = 0
    var inProgress = 0
    var completed = 0
    var cancelled = 0
    for (category in categories) {
        val c = category.statusCounts()
        todo += c.todo
        inProgress += c.inProgress
        completed += c.completed
        cancelled += c.cancelled
    }
    return StatusCounts(todo, inProgress, completed, cancelled)
}

/**
 * Returns a copy of the project with [transform] applied to the one matching
 * task (by category + task id). Used for optimistic status updates before the
 * server confirms.
 */
fun Project.updateTask(
    categoryId: String,
    taskId: String,
    transform: (Task) -> Task,
): Project = copy(
    categories = categories.map { category ->
        if (category.id != categoryId) {
            category
        } else {
            category.copy(
                tasks = category.tasks.map { task ->
                    if (task.id != taskId) task else transform(task)
                },
            )
        }
    },
)
