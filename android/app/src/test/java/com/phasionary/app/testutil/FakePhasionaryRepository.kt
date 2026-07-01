package com.phasionary.app.testutil

import com.phasionary.app.data.ApiException
import com.phasionary.app.data.model.CreateTaskBody
import com.phasionary.app.data.model.Project
import com.phasionary.app.data.model.Status
import com.phasionary.app.data.model.Task
import com.phasionary.app.data.repo.PhasionaryRepository

/** In-memory repository double for ViewModel tests. */
class FakePhasionaryRepository(
    var project: Project,
    var failSetStatus: Boolean = false,
) : PhasionaryRepository {

    data class SetStatusCall(
        val projectId: String,
        val categoryId: String,
        val taskId: String,
        val status: String,
    )

    val setStatusCalls = mutableListOf<SetStatusCall>()

    override suspend fun listProjects(): List<Project> = listOf(project)

    override suspend fun getProject(projectId: String): Project = project

    override suspend fun createTask(
        projectId: String,
        categoryId: String,
        body: CreateTaskBody,
    ): Task = Task(
        id = "generated",
        title = body.title,
        status = body.status ?: Status.TODO,
        priority = body.priority ?: "",
        estimateMinutes = body.estimateMinutes ?: 0,
        description = body.description ?: "",
    )

    override suspend fun setTaskStatus(
        projectId: String,
        categoryId: String,
        taskId: String,
        status: String,
    ): Task {
        if (failSetStatus) throw ApiException.Network(RuntimeException("boom"))
        setStatusCalls += SetStatusCall(projectId, categoryId, taskId, status)
        val existing = project.categories
            .firstOrNull { it.id == categoryId }
            ?.tasks
            ?.firstOrNull { it.id == taskId }
        return (existing ?: Task(id = taskId, title = "")).copy(status = status)
    }
}
