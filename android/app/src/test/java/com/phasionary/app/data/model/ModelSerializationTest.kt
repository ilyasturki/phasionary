package com.phasionary.app.data.model

import com.phasionary.app.data.net.PhasionaryJson
import kotlinx.serialization.decodeFromString
import kotlinx.serialization.encodeToString
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

/**
 * Verifies the wire models decode the server's snake_case JSON, tolerate omitted
 * optionals + unknown fields, and that request bodies omit unset fields (so the
 * server applies its own defaults). The JSON literals are hand-written from the
 * documented schema — an independent source, not produced by these classes.
 */
class ModelSerializationTest {

    private val projectsJson = """
        [
          {
            "id": "p1",
            "name": "Phasionary",
            "created_at": "2026-01-01T00:00:00Z",
            "updated_at": "2026-01-02T00:00:00Z",
            "future_field": "ignored",
            "categories": [
              {
                "id": "c1",
                "name": "Feature",
                "created_at": "2026-01-01T00:00:00Z",
                "tasks": [
                  {
                    "id": "t1",
                    "title": "Wire the list",
                    "status": "in_progress",
                    "created_at": "2026-01-01T00:00:00Z",
                    "updated_at": "2026-01-01T00:00:00Z",
                    "priority": "high",
                    "estimate_minutes": 90,
                    "description": "note"
                  },
                  {
                    "id": "t2",
                    "title": "Bare task",
                    "status": "todo",
                    "created_at": "2026-01-01T00:00:00Z",
                    "updated_at": "2026-01-01T00:00:00Z"
                  }
                ]
              }
            ]
          }
        ]
    """.trimIndent()

    @Test
    fun decodesProjectsWithNestedCategoriesAndTasks() {
        val projects = PhasionaryJson.decodeFromString<List<Project>>(projectsJson)

        assertEquals(1, projects.size)
        val project = projects[0]
        assertEquals("p1", project.id)
        assertEquals("Phasionary", project.name)
        assertEquals(1, project.categories.size)

        val tasks = project.categories[0].tasks
        assertEquals(2, tasks.size)

        val t1 = tasks[0]
        assertEquals("in_progress", t1.status)
        assertEquals("high", t1.priority)
        assertEquals(90, t1.estimateMinutes)
        assertEquals("note", t1.description)
    }

    @Test
    fun omittedOptionalsFallBackToDefaults() {
        val projects = PhasionaryJson.decodeFromString<List<Project>>(projectsJson)
        val t2 = projects[0].categories[0].tasks[1]

        assertEquals(Priority.NONE, t2.priority)
        assertEquals(0, t2.estimateMinutes)
        assertEquals("", t2.completionDate)
        assertEquals("", t2.description)
    }

    @Test
    fun createTaskBodyOmitsUnsetFields() {
        val body = CreateTaskBody(title = "Quick capture")
        val json = PhasionaryJson.encodeToString(body)

        assertTrue(json.contains("\"title\":\"Quick capture\""))
        // Unset optionals must not be serialized, so the server defaults apply.
        assertFalse(json.contains("status"))
        assertFalse(json.contains("priority"))
        assertFalse(json.contains("estimate_minutes"))
        assertFalse(json.contains("description"))
    }

    @Test
    fun createTaskBodyIncludesSetFields() {
        val body = CreateTaskBody(
            title = "With priority",
            priority = Priority.HIGH,
            estimateMinutes = 60,
        )
        val json = PhasionaryJson.encodeToString(body)

        assertTrue(json.contains("\"priority\":\"high\""))
        assertTrue(json.contains("\"estimate_minutes\":60"))
        assertFalse(json.contains("status"))
    }

    @Test
    fun categoryStatusCountsTally() {
        val category = Category(
            id = "c1",
            name = "Feature",
            tasks = listOf(
                Task(id = "1", title = "a", status = Status.TODO),
                Task(id = "2", title = "b", status = Status.TODO),
                Task(id = "3", title = "c", status = Status.IN_PROGRESS),
                Task(id = "4", title = "d", status = Status.COMPLETED),
                Task(id = "5", title = "e", status = Status.CANCELLED),
            ),
        )

        val counts = category.statusCounts()
        assertEquals(2, counts.todo)
        assertEquals(1, counts.inProgress)
        assertEquals(1, counts.completed)
        assertEquals(1, counts.cancelled)
        assertEquals(5, counts.total)
        assertEquals(3, counts.open)
        assertEquals(2, counts.done)
    }
}
