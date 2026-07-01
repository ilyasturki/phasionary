package com.phasionary.app.data.repo

import com.phasionary.app.data.ApiException
import com.phasionary.app.data.model.CreateTaskBody
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
