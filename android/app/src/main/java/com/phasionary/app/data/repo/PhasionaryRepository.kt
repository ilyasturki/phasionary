package com.phasionary.app.data.repo

import com.phasionary.app.data.model.Category
import com.phasionary.app.data.model.CreateTaskBody
import com.phasionary.app.data.model.Project
import com.phasionary.app.data.model.Task
import com.phasionary.app.data.model.UpdateTaskBody

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

    /**
     * POST .../tasks with kind=separator — inserts a divider directly below
     * [afterTaskId]. Separators only make sense between rows, so the anchor is
     * required rather than optional. The label starts empty (a plain rule) and
     * is set afterwards by relabeling.
     */
    suspend fun createSeparator(
        projectId: String,
        categoryId: String,
        afterTaskId: String,
    ): Task

    /** POST .../tasks/{tid}/status — explicit set; returns the updated task. */
    suspend fun setTaskStatus(
        projectId: String,
        categoryId: String,
        taskId: String,
        status: String,
    ): Task

    /** PATCH .../tasks/{tid} — partial edit; returns the updated task. */
    suspend fun updateTask(
        projectId: String,
        categoryId: String,
        taskId: String,
        body: UpdateTaskBody,
    ): Task

    /** DELETE .../tasks/{tid}. */
    suspend fun deleteTask(projectId: String, categoryId: String, taskId: String)

    /** POST .../categories — returns the created (empty) category. */
    suspend fun createCategory(projectId: String, name: String): Category

    /** GET .../folds — the category IDs currently collapsed, shared with the TUI. */
    suspend fun getFolds(projectId: String): List<String>

    /** PUT .../folds — replaces the collapsed set with [categoryIds]. */
    suspend fun setFolds(projectId: String, categoryIds: List<String>)
}
