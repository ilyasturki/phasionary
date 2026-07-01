package com.phasionary.app.data

/**
 * Typed failures surfaced by the repository. The ViewModel maps these to
 * user-facing messages; keeping them typed lets the UI treat, e.g., an auth
 * failure differently from a transient network error.
 */
sealed class ApiException(message: String, cause: Throwable? = null) : Exception(message, cause) {

    /** Base URL or token not set yet — send the user to settings. */
    data object NotConfigured : ApiException("Server address or token not set")

    /** 401 — token missing or wrong. */
    data object Unauthorized : ApiException("Unauthorized — check your token")

    /** 400 — validation failure; [detail] is the server's message. */
    class BadRequest(val detail: String) : ApiException(detail)

    /** 404 — project/category/task not found. */
    class NotFound(val detail: String) : ApiException(detail)

    /** Any 5xx (or otherwise unexpected status). */
    class Server(val detail: String, val code: Int) : ApiException("Server error ($code): $detail")

    /** Connection failed / timed out / host unreachable. */
    class Network(cause: Throwable) : ApiException(cause.message ?: "Network error", cause)

    /** A 2xx response whose body could not be parsed as expected. */
    class Malformed(cause: Throwable) : ApiException("Unexpected response from server", cause)
}
