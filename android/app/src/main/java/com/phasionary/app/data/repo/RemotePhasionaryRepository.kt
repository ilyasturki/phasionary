package com.phasionary.app.data.repo

import com.phasionary.app.data.ApiException
import com.phasionary.app.data.model.CreateTaskBody
import com.phasionary.app.data.model.ErrorEnvelope
import com.phasionary.app.data.model.Project
import com.phasionary.app.data.model.StatusBody
import com.phasionary.app.data.model.Task
import com.phasionary.app.data.net.PhasionaryJson
import com.phasionary.app.data.settings.ServerConfig
import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.request.accept
import io.ktor.client.request.header
import io.ktor.client.request.request
import io.ktor.client.request.setBody
import io.ktor.client.statement.HttpResponse
import io.ktor.http.ContentType
import io.ktor.http.HttpHeaders
import io.ktor.http.HttpMethod
import io.ktor.http.content.TextContent
import io.ktor.http.isSuccess
import kotlinx.coroutines.CancellationException
import kotlinx.serialization.encodeToString

/**
 * Talks to the /api/v1 JSON API over Ktor. [configProvider] is read on every
 * call so a settings change (new base URL / token) takes effect immediately,
 * without rebuilding the client.
 *
 * Request bodies are pre-serialized to a [TextContent] so ContentNegotiation
 * (which still deserializes responses) passes them through untouched — the
 * generic request core carries no compile-time body type to hand it otherwise.
 */
class RemotePhasionaryRepository(
    private val client: HttpClient,
    private val configProvider: suspend () -> ServerConfig,
) : PhasionaryRepository {

    override suspend fun listProjects(): List<Project> =
        request(HttpMethod.Get, "projects", jsonBody = null)

    override suspend fun getProject(projectId: String): Project =
        request(HttpMethod.Get, "projects/$projectId", jsonBody = null)

    override suspend fun createTask(
        projectId: String,
        categoryId: String,
        body: CreateTaskBody,
    ): Task = request(
        HttpMethod.Post,
        "projects/$projectId/categories/$categoryId/tasks",
        jsonBody = PhasionaryJson.encodeToString(body),
    )

    override suspend fun setTaskStatus(
        projectId: String,
        categoryId: String,
        taskId: String,
        status: String,
    ): Task = request(
        HttpMethod.Post,
        "projects/$projectId/categories/$categoryId/tasks/$taskId/status",
        jsonBody = PhasionaryJson.encodeToString(StatusBody(status)),
    )

    private suspend inline fun <reified T> request(
        method: HttpMethod,
        path: String,
        jsonBody: String?,
    ): T {
        val cfg = configProvider()
        if (!cfg.isConfigured) throw ApiException.NotConfigured

        val response: HttpResponse = try {
            client.request(buildUrl(cfg.baseUrl, path)) {
                this.method = method
                header(HttpHeaders.Authorization, "Bearer ${cfg.token}")
                accept(ContentType.Application.Json)
                if (jsonBody != null) {
                    setBody(TextContent(jsonBody, ContentType.Application.Json))
                }
            }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Throwable) {
            throw ApiException.Network(e)
        }

        if (response.status.isSuccess()) {
            return try {
                response.body()
            } catch (e: CancellationException) {
                throw e
            } catch (e: Throwable) {
                throw ApiException.Malformed(e)
            }
        }
        throw errorFor(response)
    }

    /** Reads the {"error": …} envelope (falling back to the status text) and maps
     *  the status code to a typed exception. */
    private suspend fun errorFor(response: HttpResponse): ApiException {
        val detail = runCatching { response.body<ErrorEnvelope>().error }
            .getOrNull()
            ?.takeIf { it.isNotBlank() }
            ?: response.status.description

        return when (response.status.value) {
            401 -> ApiException.Unauthorized
            400 -> ApiException.BadRequest(detail)
            404 -> ApiException.NotFound(detail)
            else -> ApiException.Server(detail, response.status.value)
        }
    }
}

/** Joins the configured server root with the API path. Base URL is the server
 *  root (scheme + host + port), e.g. http://100.x.y.z:7777. */
internal fun buildUrl(baseUrl: String, path: String): String =
    "${baseUrl.trimEnd('/')}/api/v1/$path"
