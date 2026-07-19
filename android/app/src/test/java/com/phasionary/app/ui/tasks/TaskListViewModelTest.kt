package com.phasionary.app.ui.tasks

import androidx.lifecycle.SavedStateHandle
import com.phasionary.app.data.ApiException
import com.phasionary.app.data.model.Category
import com.phasionary.app.data.model.Kind
import com.phasionary.app.data.model.Project
import com.phasionary.app.data.model.Status
import com.phasionary.app.data.model.Task
import com.phasionary.app.data.model.statusCounts
import com.phasionary.app.data.repo.PhasionaryRepository
import com.phasionary.app.testutil.FakePhasionaryRepository
import com.phasionary.app.testutil.MainDispatcherRule
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test

class TaskListViewModelTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private fun sampleProject() = Project(
        id = "p1",
        name = "Project",
        categories = listOf(
            Category(
                id = "c1",
                name = "Feature",
                tasks = listOf(
                    Task(id = "t1", title = "A", status = Status.IN_PROGRESS),
                    Task(id = "t2", title = "B", status = Status.TODO),
                ),
            ),
            Category(id = "c2", name = "Fix", tasks = emptyList()),
        ),
    )

    private fun viewModel(repo: PhasionaryRepository) =
        TaskListViewModel(SavedStateHandle(mapOf(PROJECT_ID_ARG to "p1")), repo)

    private fun TaskListViewModel.task(id: String): Task =
        state.value.project!!.categories[0].tasks.first { it.id == id }

    @Test
    fun load_populatesProjectFromRepository() {
        val vm = viewModel(FakePhasionaryRepository(sampleProject()))

        val s = vm.state.value
        assertFalse(s.loading)
        assertEquals("p1", s.project?.id)
        assertNull(s.error)
    }

    @Test
    fun cycleStatus_optimisticallyAdvancesThenConfirms() {
        val repo = FakePhasionaryRepository(sampleProject())
        val vm = viewModel(repo)

        vm.cycleStatus("c1", vm.task("t2")) // todo -> in_progress

        assertEquals(Status.IN_PROGRESS, vm.task("t2").status)
        assertEquals(1, repo.setStatusCalls.size)
        assertEquals(Status.IN_PROGRESS, repo.setStatusCalls[0].status)
        assertNull(vm.state.value.message)
    }

    @Test
    fun cycleStatus_revertsAndMessagesOnFailure() {
        val repo = FakePhasionaryRepository(sampleProject(), failSetStatus = true)
        val vm = viewModel(repo)

        vm.cycleStatus("c1", vm.task("t2"))

        assertEquals(Status.TODO, vm.task("t2").status) // reverted
        assertNotNull(vm.state.value.message)
    }

    @Test
    fun toggleCategory_collapsesAndExpands() {
        val vm = viewModel(FakePhasionaryRepository(sampleProject()))

        vm.toggleCategory("c1")
        assertTrue("c1" in vm.state.value.collapsed)

        vm.toggleCategory("c1")
        assertFalse("c1" in vm.state.value.collapsed)
    }

    @Test
    fun load_readsFoldsFromServer() {
        val repo = FakePhasionaryRepository(sampleProject(), folds = listOf("c2"))
        val vm = viewModel(repo)

        assertEquals(setOf("c2"), vm.state.value.collapsed)
    }

    @Test
    fun toggleCategory_writesWholeFoldSetToServer() {
        val repo = FakePhasionaryRepository(sampleProject(), folds = listOf("c2"))
        val vm = viewModel(repo)

        vm.toggleCategory("c1")

        // The API replaces the list wholesale, so the write carries both.
        assertEquals(1, repo.setFoldsCalls.size)
        assertEquals(setOf("c1", "c2"), repo.setFoldsCalls[0].toSet())
    }

    @Test
    fun toggleCategory_revertsAndMessagesWhenFoldWriteFails() {
        val repo = FakePhasionaryRepository(sampleProject(), failSetFolds = true)
        val vm = viewModel(repo)

        vm.toggleCategory("c1")

        assertFalse("c1" in vm.state.value.collapsed)
        assertNotNull(vm.state.value.message)
    }

    @Test
    fun load_survivesFoldFetchFailure() {
        val repo = object : PhasionaryRepository by FakePhasionaryRepository(sampleProject()) {
            override suspend fun getFolds(projectId: String): List<String> =
                throw ApiException.Network(RuntimeException("boom"))
        }
        val vm = viewModel(repo)

        // Folds are cosmetic; losing them must not fail the whole screen.
        assertNotNull(vm.state.value.project)
        assertNull(vm.state.value.error)
        assertTrue(vm.state.value.collapsed.isEmpty())
    }

    @Test
    fun addTask_appendsCreatedTaskToCategory() {
        val repo = FakePhasionaryRepository(sampleProject())
        val vm = viewModel(repo)

        vm.addTask("c1", "  New thing  ")

        assertEquals(1, repo.createTaskCalls.size)
        assertEquals("New thing", repo.createTaskCalls[0].second.title)
        val tasks = vm.state.value.project!!.categories[0].tasks
        assertEquals("New thing", tasks.last().title)
    }

    @Test
    fun addTask_ignoresBlankTitle() {
        val repo = FakePhasionaryRepository(sampleProject())
        val vm = viewModel(repo)

        vm.addTask("c1", "   ")

        assertTrue(repo.createTaskCalls.isEmpty())
    }

    @Test
    fun addTask_messagesOnFailure() {
        val repo = FakePhasionaryRepository(sampleProject(), failCreate = true)
        val vm = viewModel(repo)

        vm.addTask("c1", "New thing")

        assertNotNull(vm.state.value.message)
        assertEquals(2, vm.state.value.project!!.categories[0].tasks.size)
    }

    @Test
    fun addCategory_appendsCreatedCategory() {
        val repo = FakePhasionaryRepository(sampleProject())
        val vm = viewModel(repo)

        vm.addCategory("Research")

        assertEquals(listOf("Research"), repo.createCategoryCalls)
        assertEquals("Research", vm.state.value.project!!.categories.last().name)
    }

    private fun projectWithSeparator() = Project(
        id = "p1",
        name = "Project",
        categories = listOf(
            Category(
                id = "c1",
                name = "Feature",
                tasks = listOf(
                    Task(id = "t1", title = "A", status = Status.TODO),
                    Task(id = "s1", title = "Later", status = "", kind = Kind.SEPARATOR),
                    Task(id = "t2", title = "B", status = Status.TODO),
                ),
            ),
        ),
    )

    @Test
    fun renameSeparator_updatesLabelOptimisticallyAndPatchesTitleOnly() {
        val repo = FakePhasionaryRepository(projectWithSeparator())
        val vm = viewModel(repo)

        vm.renameSeparator("c1", "s1", "  Much later  ")

        val sep = vm.state.value.project!!.categories[0].tasks.first { it.id == "s1" }
        assertEquals("Much later", sep.title)
        assertEquals(1, repo.updateCalls.size)
        val body = repo.updateCalls[0].body
        assertEquals("Much later", body.title)
        // A separator has nothing else to set; sending anything else 400s.
        assertNull(body.status)
        assertNull(body.priority)
        assertNull(body.estimateMinutes)
        assertNull(body.description)
    }

    @Test
    fun renameSeparator_allowsBlankLabel() {
        val repo = FakePhasionaryRepository(projectWithSeparator())
        val vm = viewModel(repo)

        // Blank is meaningful for a separator: it goes back to a plain rule.
        vm.renameSeparator("c1", "s1", "   ")

        assertEquals(1, repo.updateCalls.size)
        assertEquals("", repo.updateCalls[0].body.title)
    }

    @Test
    fun deleteRow_removesSeparatorFromTheList() {
        val repo = FakePhasionaryRepository(projectWithSeparator())
        val vm = viewModel(repo)

        vm.deleteRow("c1", "s1")

        val tasks = vm.state.value.project!!.categories[0].tasks
        assertEquals(listOf("t1", "t2"), tasks.map { it.id })
        assertEquals(listOf("s1"), repo.deleteCalls)
    }

    @Test
    fun deleteRow_messagesAndKeepsRowOnFailure() {
        val repo = FakePhasionaryRepository(projectWithSeparator(), failDelete = true)
        val vm = viewModel(repo)

        vm.deleteRow("c1", "s1")

        assertNotNull(vm.state.value.message)
        assertEquals(3, vm.state.value.project!!.categories[0].tasks.size)
    }

    @Test
    fun statusCounts_excludeSeparators() {
        val vm = viewModel(FakePhasionaryRepository(projectWithSeparator()))

        val counts = vm.state.value.project!!.categories[0].statusCounts()

        // Two real tasks; the divider between them is not work.
        assertEquals(2, counts.total)
        assertEquals(2, counts.todo)
    }

    @Test
    fun refresh_keepsShowingDataWhileReloading() {
        val repo = FakePhasionaryRepository(sampleProject())
        val vm = viewModel(repo)

        vm.refresh()

        // No spinner flash on the way back from the editor.
        assertFalse(vm.state.value.loading)
        assertNotNull(vm.state.value.project)
    }
}
