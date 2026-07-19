package com.phasionary.app.testutil

import com.phasionary.app.data.ApiException
import com.phasionary.app.data.model.Category
import com.phasionary.app.data.model.CreateTaskBody
import com.phasionary.app.data.model.Kind
import com.phasionary.app.data.model.Project
import com.phasionary.app.data.model.Status
import com.phasionary.app.data.model.Task
import com.phasionary.app.data.model.UpdateTaskBody
import com.phasionary.app.data.repo.PhasionaryRepository

/** In-memory repository double for ViewModel tests. */
class FakePhasionaryRepository(
    var project: Project,
    var failSetStatus: Boolean = false,
    var failSetFolds: Boolean = false,
    var failCreate: Boolean = false,
    var failUpdate: Boolean = false,
    var failDelete: Boolean = false,
    /** What getFolds reports; also what setFolds writes into. */
    var folds: List<String> = emptyList(),
) : PhasionaryRepository {

    data class SetStatusCall(
        val projectId: String,
        val categoryId: String,
        val taskId: String,
        val status: String,
    )

    data class UpdateCall(
        val projectId: String,
        val categoryId: String,
        val taskId: String,
        val body: UpdateTaskBody,
    )

    val setStatusCalls = mutableListOf<SetStatusCall>()
    val setFoldsCalls = mutableListOf<List<String>>()
    val createTaskCalls = mutableListOf<Pair<String, CreateTaskBody>>()

    /** (categoryId, afterTaskId) for each separator inserted. */
    val createSeparatorCalls = mutableListOf<Pair<String, String>>()
    val createCategoryCalls = mutableListOf<String>()
    val updateCalls = mutableListOf<UpdateCall>()
    val deleteCalls = mutableListOf<String>()

    override suspend fun listProjects(): List<Project> = listOf(project)

    override suspend fun getProject(projectId: String): Project = project

    override suspend fun createTask(
        projectId: String,
        categoryId: String,
        body: CreateTaskBody,
    ): Task {
        if (failCreate) throw ApiException.Network(RuntimeException("boom"))
        createTaskCalls += categoryId to body
        return Task(
            id = "generated",
            title = body.title,
            status = body.status ?: Status.TODO,
            priority = body.priority ?: "",
            estimateMinutes = body.estimateMinutes ?: 0,
            description = body.description ?: "",
        )
    }

    override suspend fun createSeparator(
        projectId: String,
        categoryId: String,
        afterTaskId: String,
    ): Task {
        if (failCreate) throw ApiException.Network(RuntimeException("boom"))
        createSeparatorCalls += categoryId to afterTaskId
        return Task(id = "sep-generated", title = "", status = "", kind = Kind.SEPARATOR)
    }

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

    override suspend fun updateTask(
        projectId: String,
        categoryId: String,
        taskId: String,
        body: UpdateTaskBody,
    ): Task {
        if (failUpdate) throw ApiException.Network(RuntimeException("boom"))
        updateCalls += UpdateCall(projectId, categoryId, taskId, body)
        val existing = project.categories
            .firstOrNull { it.id == categoryId }
            ?.tasks
            ?.firstOrNull { it.id == taskId }
            ?: Task(id = taskId, title = "")
        return existing.copy(
            title = body.title ?: existing.title,
            status = body.status ?: existing.status,
            priority = body.priority ?: existing.priority,
            estimateMinutes = body.estimateMinutes ?: existing.estimateMinutes,
            description = body.description ?: existing.description,
        )
    }

    override suspend fun deleteTask(projectId: String, categoryId: String, taskId: String) {
        if (failDelete) throw ApiException.Network(RuntimeException("boom"))
        deleteCalls += taskId
    }

    override suspend fun createCategory(projectId: String, name: String): Category {
        if (failCreate) throw ApiException.Network(RuntimeException("boom"))
        createCategoryCalls += name
        return Category(id = "cat-generated", name = name)
    }

    override suspend fun getFolds(projectId: String): List<String> = folds

    override suspend fun setFolds(projectId: String, categoryIds: List<String>) {
        if (failSetFolds) throw ApiException.Network(RuntimeException("boom"))
        setFoldsCalls += categoryIds
        folds = categoryIds
    }
}
