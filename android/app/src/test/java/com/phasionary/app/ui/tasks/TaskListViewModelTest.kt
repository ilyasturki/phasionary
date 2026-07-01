package com.phasionary.app.ui.tasks

import androidx.lifecycle.SavedStateHandle
import com.phasionary.app.data.model.Category
import com.phasionary.app.data.model.Project
import com.phasionary.app.data.model.Status
import com.phasionary.app.data.model.Task
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
}
