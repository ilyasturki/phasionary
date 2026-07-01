package com.phasionary.app.data.repo

import com.phasionary.app.data.model.CreateTaskBody
import com.phasionary.app.data.model.Project
import com.phasionary.app.data.model.Task

/**
 * The app's data seam. v1 has a single remote implementation, but keeping the
 * interface lets a local cache / offline path slot in later (per the build plan)
 * without touching the ViewModels.
 *
 * Implementations throw [com.phasionary.app.data.ApiException] on failure.
 */
interface PhasionaryRepository {

    /** GET /api/v1/projects — full projects (with categories + tasks). */
    suspend fun listProjects(): List<Project>

    /** GET /api/v1/projects/{pid}. */
    suspend fun getProject(projectId: String): Project

    /** POST .../tasks — quick capture; returns the created task. */
    suspend fun createTask(projectId: String, categoryId: String, body: CreateTaskBody): Task

    /** POST .../tasks/{tid}/status — explicit set; returns the updated task. */
    suspend fun setTaskStatus(
        projectId: String,
        categoryId: String,
        taskId: String,
        status: String,
    ): Task
}
