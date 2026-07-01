package com.phasionary.app.ui

import com.phasionary.app.data.ApiException

/** Maps a thrown error to a short, user-facing message. */
fun userMessage(error: Throwable): String = when (error) {
    is ApiException.NotConfigured -> "Set your server address and token in Settings."
    is ApiException.Unauthorized -> "Unauthorized — check your token in Settings."
    is ApiException.NotFound -> error.detail
    is ApiException.BadRequest -> error.detail
    is ApiException.Server -> "Server error: ${error.detail}"
    is ApiException.Network -> "Can't reach the server. Check the address and your connection."
    is ApiException.Malformed -> "Unexpected response from the server."
    else -> error.message ?: "Something went wrong."
}
