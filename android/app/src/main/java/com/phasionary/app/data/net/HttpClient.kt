package com.phasionary.app.data.net

import io.ktor.client.HttpClient
import io.ktor.client.HttpClientConfig
import io.ktor.client.engine.okhttp.OkHttp
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.plugins.logging.LogLevel
import io.ktor.client.plugins.logging.Logging
import io.ktor.serialization.kotlinx.json.json
import kotlinx.serialization.json.Json

/**
 * Shared JSON config. ignoreUnknownKeys keeps the client forward-compatible with
 * new server fields; explicitNulls=false + encodeDefaults=false keep request
 * bodies minimal so the server applies its own defaults for omitted fields.
 */
val PhasionaryJson: Json = Json {
    ignoreUnknownKeys = true
    explicitNulls = false
    encodeDefaults = false
    isLenient = true
}

/**
 * Applied to both the real (OkHttp) client and the MockEngine client used in
 * tests, so tests exercise the same (de)serialization path as production.
 * expectSuccess=false because the repository inspects status codes itself.
 */
fun HttpClientConfig<*>.phasionaryDefaults() {
    install(ContentNegotiation) {
        json(PhasionaryJson)
    }
    install(Logging) {
        level = LogLevel.NONE
    }
    expectSuccess = false
}

/** Production client, OkHttp engine. */
fun defaultHttpClient(): HttpClient = HttpClient(OkHttp) {
    phasionaryDefaults()
}
