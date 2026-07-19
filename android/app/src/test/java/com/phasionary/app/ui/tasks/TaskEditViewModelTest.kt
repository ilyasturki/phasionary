package com.phasionary.app.ui.tasks

import androidx.lifecycle.SavedStateHandle
import com.phasionary.app.data.model.Category
import com.phasionary.app.data.model.Priority
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

class TaskEditViewModelTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private fun sampleTask() = Task(
        id = "t1",
        title = "Original",
        status = Status.TODO,
        priority = Priority.MEDIUM,
        estimateMinutes = 30,
        description = "notes",
    )

    private fun sampleProject(task: Task = sampleTask()) = Project(
        id = "p1",
        name = "Project",
        categories = listOf(Category(id = "c1", name = "Feature", tasks = listOf(task))),
    )

    private fun viewModel(repo: PhasionaryRepository, taskId: String = "t1") =
        TaskEditViewModel(
            SavedStateHandle(
                mapOf(
                    PROJECT_ID_ARG to "p1",
                    CATEGORY_ID_ARG to "c1",
                    TASK_ID_ARG to taskId,
                ),
            ),
            repo,
        )

    @Test
    fun load_fillsFormFromTask() {
        val vm = viewModel(FakePhasionaryRepository(sampleProject()))

        val s = vm.state.value
        assertFalse(s.loading)
        assertEquals("Original", s.form.title)
        assertEquals(Status.TODO, s.form.status)
        assertEquals(Priority.MEDIUM, s.form.priority)
        assertEquals("30", s.form.estimateText)
        assertEquals("notes", s.form.description)
        assertFalse(s.dirty)
    }

    @Test
    fun load_reportsMissingTask() {
        val vm = viewModel(FakePhasionaryRepository(sampleProject()), taskId = "gone")

        assertNotNull(vm.state.value.error)
        assertNull(vm.state.value.original)
    }

    @Test
    fun form_isNotDirtyUntilSomethingChanges() {
        val vm = viewModel(FakePhasionaryRepository(sampleProject()))

        assertFalse(vm.state.value.canSave)

        vm.updateForm { it.copy(title = "Renamed") }

        assertTrue(vm.state.value.dirty)
        assertTrue(vm.state.value.canSave)
    }

    @Test
    fun save_sendsOnlyChangedFields() {
        val repo = FakePhasionaryRepository(sampleProject())
        val vm = viewModel(repo)

        vm.updateForm { it.copy(status = Status.COMPLETED) }
        vm.save()

        assertEquals(1, repo.updateCalls.size)
        val body = repo.updateCalls[0].body
        assertEquals(Status.COMPLETED, body.status)
        // Untouched fields stay out of the payload so a concurrent TUI edit to
        // them survives.
        assertNull(body.title)
        assertNull(body.description)
        assertNull(body.priority)
        assertNull(body.estimateMinutes)
        assertTrue(vm.state.value.done)
    }

    @Test
    fun save_clearsDescriptionWithExplicitEmptyString() {
        val repo = FakePhasionaryRepository(sampleProject())
        val vm = viewModel(repo)

        vm.updateForm { it.copy(description = "") }
        vm.save()

        // "" (not null) is what tells the server to clear it.
        assertEquals("", repo.updateCalls[0].body.description)
    }

    @Test
    fun save_sendsZeroWhenEstimateCleared() {
        val repo = FakePhasionaryRepository(sampleProject())
        val vm = viewModel(repo)

        vm.updateForm { it.copy(estimateText = "") }
        vm.save()

        assertEquals(0, repo.updateCalls[0].body.estimateMinutes)
    }

    @Test
    fun save_trimsTitle() {
        val repo = FakePhasionaryRepository(sampleProject())
        val vm = viewModel(repo)

        vm.updateForm { it.copy(title = "  Renamed  ") }
        vm.save()

        assertEquals("Renamed", repo.updateCalls[0].body.title)
    }

    @Test
    fun save_isBlockedByBlankTitle() {
        val repo = FakePhasionaryRepository(sampleProject())
        val vm = viewModel(repo)

        vm.updateForm { it.copy(title = "   ") }

        assertFalse(vm.state.value.canSave)
        vm.save()
        assertTrue(repo.updateCalls.isEmpty())
    }

    @Test
    fun save_isBlockedByUnparseableEstimate() {
        val repo = FakePhasionaryRepository(sampleProject())
        val vm = viewModel(repo)

        vm.updateForm { it.copy(estimateText = "soon") }

        assertNull(vm.state.value.form.estimateMinutes)
        assertFalse(vm.state.value.canSave)
    }

    @Test
    fun save_messagesAndStaysOpenOnFailure() {
        val repo = FakePhasionaryRepository(sampleProject(), failUpdate = true)
        val vm = viewModel(repo)

        vm.updateForm { it.copy(title = "Renamed") }
        vm.save()

        assertNotNull(vm.state.value.message)
        assertFalse(vm.state.value.done)
        assertFalse(vm.state.value.saving)
    }

    @Test
    fun delete_closesTheScreen() {
        val repo = FakePhasionaryRepository(sampleProject())
        val vm = viewModel(repo)

        vm.delete()

        assertEquals(listOf("t1"), repo.deleteCalls)
        assertTrue(vm.state.value.done)
    }

    @Test
    fun delete_messagesOnFailure() {
        val repo = FakePhasionaryRepository(sampleProject(), failDelete = true)
        val vm = viewModel(repo)

        vm.delete()

        assertNotNull(vm.state.value.message)
        assertFalse(vm.state.value.done)
    }

    @Test
    fun insertSeparatorBelow_anchorsOnThisTaskAndStaysOpen() {
        val repo = FakePhasionaryRepository(sampleProject())
        val vm = viewModel(repo)

        vm.insertSeparatorBelow()

        assertEquals(listOf("c1" to "t1"), repo.createSeparatorCalls)
        assertNotNull(vm.state.value.message)
        // Staying put is the point: navigating away would silently drop any
        // unsaved edits on the form.
        assertFalse(vm.state.value.done)
        assertFalse(vm.state.value.saving)
    }

    @Test
    fun insertSeparatorBelow_keepsUnsavedEdits() {
        val repo = FakePhasionaryRepository(sampleProject())
        val vm = viewModel(repo)

        vm.updateForm { it.copy(title = "Renamed") }
        vm.insertSeparatorBelow()

        assertEquals("Renamed", vm.state.value.form.title)
        assertTrue(vm.state.value.dirty)
        assertTrue(repo.updateCalls.isEmpty())
    }

    @Test
    fun insertSeparatorBelow_messagesOnFailure() {
        val repo = FakePhasionaryRepository(sampleProject(), failCreate = true)
        val vm = viewModel(repo)

        vm.insertSeparatorBelow()

        assertNotNull(vm.state.value.message)
        assertFalse(vm.state.value.saving)
    }

    @Test
    fun diff_returnsEmptyBodyWhenNothingChanged() {
        val task = sampleTask()
        val body = TaskEditViewModel.diff(task, TaskForm.of(task))

        assertTrue(body.isEmpty)
    }
}
