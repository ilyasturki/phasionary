package com.phasionary.app.data.repo

import com.phasionary.app.data.ApiException
import com.phasionary.app.data.model.CreateTaskBody
import com.phasionary.app.data.model.UpdateTaskBody
import com.phasionary.app.data.net.phasionaryDefaults
import com.phasionary.app.data.settings.ServerConfig
import io.ktor.client.HttpClient
import io.ktor.client.engine.mock.MockEngine
import io.ktor.client.engine.mock.respond
import io.ktor.http.ContentType
import io.ktor.http.HttpHeaders
import io.ktor.http.HttpStatusCode
import io.ktor.http.content.TextContent
import io.ktor.http.headersOf
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

/**
 * Drives the repository through Ktor's MockEngine — a plain JVM test, no
 * emulator. Exercises URL construction, the bearer header, request-body
 * omission, success parsing, and status-code -> typed-error mapping.
 */
class RemotePhasionaryRepositoryTest {

    private val config = ServerConfig("http://test.local:7777", "secret")

    private val jsonHeader = headersOf(HttpHeaders.ContentType, ContentType.Application.Json.toString())

    private val projectsJson = """
        [{"id":"p1","name":"Phasionary","categories":[
          {"id":"c1","name":"Feature","tasks":[
            {"id":"t1","title":"A","status":"todo"},
            {"id":"t2","title":"B","status":"in_progress"}
          ]}
        ]}]
    """.trimIndent()

    private val createdTaskJson = """
        {"id":"tNew","title":"Quick capture","status":"todo"}
    """.trimIndent()

    @Test
    fun listProjects_parsesBodyAndSendsBearerToCorrectUrl() = runTest {
        var seenUrl = ""
        var seenAuth: String? = null
        var seenMethod = ""

        val client = HttpClient(MockEngine) {
            phasionaryDefaults()
            engine {
                addHandler { request ->
                    seenUrl = request.url.toString()
                    seenAuth = request.headers[HttpHeaders.Authorization]
                    seenMethod = request.method.value
                    respond(projectsJson, HttpStatusCode.OK, jsonHeader)
                }
            }
        }
        val repo = RemotePhasionaryRepository(client) { config }

        val projects = repo.listProjects()

        assertEquals(1, projects.size)
        assertEquals("Phasionary", projects[0].name)
        assertEquals(2, projects[0].categories[0].tasks.size)
        assertEquals("http://test.local:7777/api/v1/projects", seenUrl)
        assertEquals("Bearer secret", seenAuth)
        assertEquals("GET", seenMethod)
    }

    @Test
    fun createTask_postsMinimalBodyAndReturnsCreatedTask() = runTest {
        var seenMethod = ""
        var seenBody = ""
        var seenUrl = ""

        val client = HttpClient(MockEngine) {
            phasionaryDefaults()
            engine {
                addHandler { request ->
                    seenMethod = request.method.value
                    seenUrl = request.url.toString()
                    seenBody = (request.body as? TextContent)?.text.orEmpty()
                    respond(createdTaskJson, HttpStatusCode.Created, jsonHeader)
                }
            }
        }
        val repo = RemotePhasionaryRepository(client) { config }

        val task = repo.createTask("p1", "c1", CreateTaskBody(title = "Quick capture"))

        assertEquals("tNew", task.id)
        assertEquals("Quick capture", task.title)
        assertEquals("POST", seenMethod)
        assertEquals("http://test.local:7777/api/v1/projects/p1/categories/c1/tasks", seenUrl)
        assertTrue(seenBody.contains("\"title\":\"Quick capture\""))
        // Unset optionals are omitted -> the server applies defaults.
        assertFalse(seenBody.contains("priority"))
        assertFalse(seenBody.contains("status"))
    }

    @Test
    fun notFound_mapsToTypedError() = runTest {
        val client = HttpClient(MockEngine) {
            phasionaryDefaults()
            engine {
                addHandler { respond("""{"error":"project not found"}""", HttpStatusCode.NotFound, jsonHeader) }
            }
        }
        val repo = RemotePhasionaryRepository(client) { config }

        val thrown = runCatching { repo.getProject("nope") }.exceptionOrNull()

        assertTrue(thrown is ApiException.NotFound)
        assertEquals("project not found", (thrown as ApiException.NotFound).detail)
    }

    @Test
    fun unauthorized_mapsToTypedError() = runTest {
        val client = HttpClient(MockEngine) {
            phasionaryDefaults()
            engine {
                addHandler { respond("""{"error":"unauthorized"}""", HttpStatusCode.Unauthorized, jsonHeader) }
            }
        }
        val repo = RemotePhasionaryRepository(client) { config }

        val thrown = runCatching { repo.listProjects() }.exceptionOrNull()

        assertTrue(thrown is ApiException.Unauthorized)
    }

    @Test
    fun updateTask_patchesOnlyTheFieldsThatAreSet() = runTest {
        var seenMethod = ""
        var seenUrl = ""
        var seenBody = ""

        val client = HttpClient(MockEngine) {
            phasionaryDefaults()
            engine {
                addHandler { request ->
                    seenMethod = request.method.value
                    seenUrl = request.url.toString()
                    seenBody = (request.body as? TextContent)?.text.orEmpty()
                    respond("""{"id":"t1","title":"A","status":"completed"}""", HttpStatusCode.OK, jsonHeader)
                }
            }
        }
        val repo = RemotePhasionaryRepository(client) { config }

        val task = repo.updateTask("p1", "c1", "t1", UpdateTaskBody(status = "completed"))

        assertEquals("completed", task.status)
        assertEquals("PATCH", seenMethod)
        assertEquals("http://test.local:7777/api/v1/projects/p1/categories/c1/tasks/t1", seenUrl)
        assertTrue(seenBody.contains("\"status\":\"completed\""))
        // Null fields are omitted, which is what the server reads as "unchanged".
        assertFalse(seenBody.contains("title"))
        assertFalse(seenBody.contains("description"))
    }

    @Test
    fun updateTask_sendsExplicitEmptiesSoTheServerClearsThem() = runTest {
        var seenBody = ""
        val client = HttpClient(MockEngine) {
            phasionaryDefaults()
            engine {
                addHandler { request ->
                    seenBody = (request.body as? TextContent)?.text.orEmpty()
                    respond("""{"id":"t1","title":"A"}""", HttpStatusCode.OK, jsonHeader)
                }
            }
        }
        val repo = RemotePhasionaryRepository(client) { config }

        repo.updateTask("p1", "c1", "t1", UpdateTaskBody(description = "", estimateMinutes = 0))

        assertTrue(seenBody.contains("\"description\":\"\""))
        assertTrue(seenBody.contains("\"estimate_minutes\":0"))
    }

    @Test
    fun createSeparator_postsKindAndAnchor() = runTest {
        var seenUrl = ""
        var seenBody = ""

        val client = HttpClient(MockEngine) {
            phasionaryDefaults()
            engine {
                addHandler { request ->
                    seenUrl = request.url.toString()
                    seenBody = (request.body as? TextContent)?.text.orEmpty()
                    respond(
                        """{"id":"sNew","title":"","status":"","kind":"separator"}""",
                        HttpStatusCode.Created,
                        jsonHeader,
                    )
                }
            }
        }
        val repo = RemotePhasionaryRepository(client) { config }

        val separator = repo.createSeparator("p1", "c1", "t1")

        assertTrue(separator.isSeparator)
        assertEquals("http://test.local:7777/api/v1/projects/p1/categories/c1/tasks", seenUrl)
        assertTrue(seenBody.contains("\"kind\":\"separator\""))
        assertTrue(seenBody.contains("\"insert_after\":\"t1\""))
        // A separator must not carry task fields, or the server 400s it.
        assertFalse(seenBody.contains("status"))
        assertFalse(seenBody.contains("priority"))
        assertFalse(seenBody.contains("estimate_minutes"))
    }

    @Test
    fun deleteTask_sendsDeleteAndToleratesEmptyBody() = runTest {
        var seenMethod = ""
        var seenUrl = ""

        val client = HttpClient(MockEngine) {
            phasionaryDefaults()
            engine {
                addHandler { request ->
                    seenMethod = request.method.value
                    seenUrl = request.url.toString()
                    // 204: the repository must not try to parse a body here.
                    respond("", HttpStatusCode.NoContent)
                }
            }
        }
        val repo = RemotePhasionaryRepository(client) { config }

        repo.deleteTask("p1", "c1", "t1")

        assertEquals("DELETE", seenMethod)
        assertEquals("http://test.local:7777/api/v1/projects/p1/categories/c1/tasks/t1", seenUrl)
    }

    @Test
    fun createCategory_postsNameAndParsesCategory() = runTest {
        var seenUrl = ""
        var seenBody = ""

        val client = HttpClient(MockEngine) {
            phasionaryDefaults()
            engine {
                addHandler { request ->
                    seenUrl = request.url.toString()
                    seenBody = (request.body as? TextContent)?.text.orEmpty()
                    respond("""{"id":"cNew","name":"Research"}""", HttpStatusCode.Created, jsonHeader)
                }
            }
        }
        val repo = RemotePhasionaryRepository(client) { config }

        val category = repo.createCategory("p1", "Research")

        assertEquals("cNew", category.id)
        assertEquals("Research", category.name)
        assertEquals("http://test.local:7777/api/v1/projects/p1/categories", seenUrl)
        assertTrue(seenBody.contains("\"name\":\"Research\""))
    }

    @Test
    fun folds_areReadAndWrittenAsAWholeList() = runTest {
        var seenMethod = ""
        var seenUrl = ""
        var seenBody = ""

        val client = HttpClient(MockEngine) {
            phasionaryDefaults()
            engine {
                addHandler { request ->
                    seenMethod = request.method.value
                    seenUrl = request.url.toString()
                    seenBody = (request.body as? TextContent)?.text.orEmpty()
                    respond("""{"folded_categories":["c1","c2"]}""", HttpStatusCode.OK, jsonHeader)
                }
            }
        }
        val repo = RemotePhasionaryRepository(client) { config }

        assertEquals(listOf("c1", "c2"), repo.getFolds("p1"))
        assertEquals("GET", seenMethod)
        assertEquals("http://test.local:7777/api/v1/projects/p1/folds", seenUrl)

        repo.setFolds("p1", listOf("c1"))
        assertEquals("PUT", seenMethod)
        assertTrue(seenBody.contains("\"folded_categories\":[\"c1\"]"))
    }

    @Test
    fun setFolds_sendsEmptyListRatherThanDroppingTheField() = runTest {
        var seenBody = ""
        val client = HttpClient(MockEngine) {
            phasionaryDefaults()
            engine {
                addHandler { request ->
                    seenBody = (request.body as? TextContent)?.text.orEmpty()
                    respond("""{"folded_categories":[]}""", HttpStatusCode.OK, jsonHeader)
                }
            }
        }
        val repo = RemotePhasionaryRepository(client) { config }

        // encodeDefaults=false would drop a defaulted empty list, and the server
        // would then never learn that everything was unfolded.
        repo.setFolds("p1", emptyList())

        assertTrue(seenBody.contains("\"folded_categories\":[]"))
    }

    @Test
    fun blankConfig_failsFastWithoutHittingNetwork() = runTest {
        var handlerCalled = false
        val client = HttpClient(MockEngine) {
            phasionaryDefaults()
            engine {
                addHandler {
                    handlerCalled = true
                    respond("[]", HttpStatusCode.OK, jsonHeader)
                }
            }
        }
        val repo = RemotePhasionaryRepository(client) { ServerConfig("", "") }

        val thrown = runCatching { repo.listProjects() }.exceptionOrNull()

        assertTrue(thrown is ApiException.NotConfigured)
        assertFalse(handlerCalled)
    }
}
