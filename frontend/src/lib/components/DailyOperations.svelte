<script lang="ts">
    import { Recycle, Trash2 } from 'lucide-svelte';
    import {
        completeTickTickTask,
        createTickTickTask,
        loadDailyOperations,
        updateTickTickTask,
    } from '$lib/api/dashboard';
    import type { DailyOperationsModel, DailyTask } from '$lib/types/dashboard';

    export let data: DailyOperationsModel;

    const EMPTY_DAILY_OPERATIONS: DailyOperationsModel = {
        tasks: [],
        garbage: {
            nextPickupDate: '',
            nextTrashPickupDate: '',
            nextRecyclingPickupDate: '',
            isRecyclingWeek: false,
            showIndicator: false,
            showTrashTakeOutReminder: false,
            showRecyclingTakeOutReminder: false,
        },
    };

    let viewData: DailyOperationsModel = EMPTY_DAILY_OPERATIONS;
    let completingTaskId = '';
    let completionError = '';
    let mutatingTaskId = '';
    let mutationError = '';
    let newTaskTitle = '';
    let newTaskDueAt = defaultDueAtInputValue();
    let editingTaskId = '';
    let editTaskTitle = '';
    let editTaskDueAt = '';

    function cloneDailyOperations(
        source: DailyOperationsModel,
    ): DailyOperationsModel {
        return {
            tasks: [...source.tasks],
            garbage: { ...source.garbage },
        };
    }

    $: if (data) {
        viewData = cloneDailyOperations(data);
    }

    function formatDate(dateIso: string): string {
        return new Date(dateIso).toLocaleDateString(undefined, {
            weekday: 'short',
            month: 'short',
            day: 'numeric',
        });
    }

    function formatOptionalDate(dateIso: string): string {
        if (!dateIso) {
            return 'Not available';
        }

        return formatDate(dateIso);
    }

    function defaultDueAtInputValue(): string {
        const now = new Date();
        now.setMinutes(0, 0, 0);
        now.setHours(now.getHours() + 1);
        return toLocalDateTimeInputValue(now.toISOString());
    }

    function toLocalDateTimeInputValue(dateIso: string): string {
        if (!dateIso) {
            return '';
        }

        const date = new Date(dateIso);
        if (Number.isNaN(date.getTime())) {
            return '';
        }

        const localTime = new Date(date.getTime() - date.getTimezoneOffset() * 60_000);
        return localTime.toISOString().slice(0, 16);
    }

    function toISOFromDateTimeInput(value: string): string {
        const trimmed = value.trim();
        if (trimmed === '') {
            return '';
        }

        const localDate = new Date(trimmed);
        if (Number.isNaN(localDate.getTime())) {
            return '';
        }

        return localDate.toISOString();
    }

    async function handleTaskCompletion(
        taskId: string,
        source: 'ticktick' | 'system',
    ): Promise<void> {
        if (source !== 'ticktick') {
            return;
        }

        if (completingTaskId !== '') {
            return;
        }

        if (mutatingTaskId !== '') {
            return;
        }

        const currentSnapshot = cloneDailyOperations(viewData);
        completionError = '';
        completingTaskId = taskId;

        viewData = {
            ...viewData,
            tasks: viewData.tasks.map((task: DailyTask) =>
                task.id === taskId ? { ...task, completed: true } : task,
            ),
        };

        try {
            await completeTickTickTask(taskId);
            viewData = await loadDailyOperations();
        } catch {
            viewData = currentSnapshot;
            completionError = 'Unable to complete task right now.';
        } finally {
            completingTaskId = '';
        }
    }

    async function handleCreateTask(): Promise<void> {
        if (mutatingTaskId !== '') {
            return;
        }

        mutationError = '';
        const dueAtISO = toISOFromDateTimeInput(newTaskDueAt);
        if (newTaskTitle.trim() === '' || dueAtISO === '') {
            mutationError = 'Provide title and due date.';
            return;
        }

        mutatingTaskId = 'create';
        try {
            await createTickTickTask(newTaskTitle, dueAtISO);
            viewData = await loadDailyOperations();
            newTaskTitle = '';
            newTaskDueAt = defaultDueAtInputValue();
        } catch {
            mutationError = 'Unable to create task right now.';
        } finally {
            mutatingTaskId = '';
        }
    }

    function startEditingTask(task: DailyTask): void {
        editingTaskId = task.id;
        editTaskTitle = task.title;
        editTaskDueAt = toLocalDateTimeInputValue(task.dueAt);
        mutationError = '';
    }

    function cancelEditingTask(): void {
        editingTaskId = '';
        editTaskTitle = '';
        editTaskDueAt = '';
    }

    async function saveTaskEdit(taskID: string): Promise<void> {
        if (mutatingTaskId !== '') {
            return;
        }

        mutationError = '';
        const dueAtISO = toISOFromDateTimeInput(editTaskDueAt);
        if (editTaskTitle.trim() === '' || dueAtISO === '') {
            mutationError = 'Provide title and due date before saving.';
            return;
        }

        mutatingTaskId = taskID;
        try {
            await updateTickTickTask(taskID, editTaskTitle, dueAtISO);
            viewData = await loadDailyOperations();
            cancelEditingTask();
        } catch {
            mutationError = 'Unable to update task right now.';
        } finally {
            mutatingTaskId = '';
        }
    }
</script>

<section class="panel">
    <header class="flex flex-wrap items-start justify-between gap-3">
        <div>
            <h2 class="text-lg font-semibold">Daily Operations</h2>
            <p class="mt-1 text-sm text-[var(--color-text)]/70">
                Priority tasks for the next 24 hours
            </p>
        </div>
        <div class="flex flex-wrap gap-2">
            {#if viewData.garbage.showTrashTakeOutReminder}
                <span
                    class="inline-flex items-center gap-1 rounded-full border border-[var(--color-primary)]/30 bg-[var(--color-primary)]/10 px-3 py-1 text-xs font-semibold text-[var(--color-primary)]"
                >
                    <Trash2 size={14} strokeWidth={2} aria-hidden="true" />
                    Take out trash tonight
                </span>
            {/if}
            {#if viewData.garbage.showRecyclingTakeOutReminder}
                <span
                    class="inline-flex items-center gap-1 rounded-full border border-[var(--color-primary)]/30 bg-[var(--color-primary)]/10 px-3 py-1 text-xs font-semibold text-[var(--color-primary)]"
                >
                    <Recycle size={14} strokeWidth={2} aria-hidden="true" />
                    Take out recycling tonight
                </span>
            {/if}
            {#if !viewData.garbage.showTrashTakeOutReminder && !viewData.garbage.showRecyclingTakeOutReminder}
                <span
                    class="rounded-full border border-[var(--color-secondary)]/30 bg-[var(--color-background)]/35 px-3 py-1 text-xs font-medium text-[var(--color-text)]/70"
                >
                    No pickup scheduled in the next 24 hours
                </span>
            {/if}
        </div>
    </header>

    <div class="mt-5 grid gap-2 sm:grid-cols-2">
        <div
            class="rounded-xl border border-[var(--color-secondary)]/25 bg-[var(--color-background)]/35 p-3"
        >
            <p
                class="inline-flex items-center gap-1 text-xs font-medium uppercase tracking-wide text-[var(--color-text)]/70"
            >
                <Trash2 size={14} strokeWidth={2} aria-hidden="true" />
                Next Trash
            </p>
            <p class="mt-1 text-sm font-semibold">
                {formatOptionalDate(viewData.garbage.nextTrashPickupDate)}
            </p>
        </div>
        <div
            class="rounded-xl border border-[var(--color-secondary)]/25 bg-[var(--color-background)]/35 p-3"
        >
            <p
                class="inline-flex items-center gap-1 text-xs font-medium uppercase tracking-wide text-[var(--color-text)]/70"
            >
                <Recycle size={14} strokeWidth={2} aria-hidden="true" />
                Next Recycling
            </p>
            <p class="mt-1 text-sm font-semibold">
                {formatOptionalDate(viewData.garbage.nextRecyclingPickupDate)}
            </p>
        </div>
    </div>

    <div
        class="mt-5 rounded-xl border border-[var(--color-secondary)]/25 bg-[var(--color-background)]/35 p-3"
    >
        <p class="text-xs font-medium uppercase tracking-wide text-[var(--color-text)]/70">
            Add TickTick Task
        </p>
        <div class="mt-2 grid gap-2 sm:grid-cols-[1fr_auto_auto]">
            <input
                type="text"
                class="rounded-lg border border-[var(--color-secondary)]/30 bg-transparent px-3 py-2 text-sm"
                placeholder="Task title"
                bind:value={newTaskTitle}
                disabled={mutatingTaskId !== ''}
            />
            <input
                type="datetime-local"
                class="rounded-lg border border-[var(--color-secondary)]/30 bg-transparent px-3 py-2 text-sm"
                bind:value={newTaskDueAt}
                disabled={mutatingTaskId !== ''}
            />
            <button
                type="button"
                class="rounded-lg border border-[var(--color-primary)]/40 bg-[var(--color-primary)]/10 px-3 py-2 text-sm font-medium text-[var(--color-primary)]"
                onclick={handleCreateTask}
                disabled={mutatingTaskId !== ''}
            >
                Add
            </button>
        </div>
    </div>

    {#if viewData.tasks.length > 0}
        <ul class="mt-5 space-y-2.5">
            {#each viewData.tasks as task}
                {@const typedTask = task as DailyTask}
                <li
                    class="flex items-center gap-3 rounded-xl border border-[var(--color-secondary)]/25 bg-[var(--color-background)]/35 px-3 py-2.5"
                >
                    <input
                        type="checkbox"
                        checked={typedTask.completed}
                        class="h-4 w-4 rounded border-[var(--color-secondary)] text-[var(--color-primary)] accent-[var(--color-primary)]"
                        disabled={typedTask.source !== 'ticktick' ||
                            completingTaskId !== ''}
                        onchange={() =>
                            handleTaskCompletion(
                                typedTask.id,
                                typedTask.source,
                            )}
                    />
                    <div class="min-w-0">
                        {#if editingTaskId === typedTask.id}
                            <div class="grid gap-2 sm:grid-cols-[1fr_auto_auto_auto]">
                                <input
                                    type="text"
                                    class="rounded-lg border border-[var(--color-secondary)]/30 bg-transparent px-2.5 py-1.5 text-sm"
                                    bind:value={editTaskTitle}
                                    disabled={mutatingTaskId !== ''}
                                />
                                <input
                                    type="datetime-local"
                                    class="rounded-lg border border-[var(--color-secondary)]/30 bg-transparent px-2.5 py-1.5 text-xs"
                                    bind:value={editTaskDueAt}
                                    disabled={mutatingTaskId !== ''}
                                />
                                <button
                                    type="button"
                                    class="rounded-lg border border-[var(--color-primary)]/40 bg-[var(--color-primary)]/10 px-2.5 py-1.5 text-xs font-medium text-[var(--color-primary)]"
                                    onclick={() => saveTaskEdit(typedTask.id)}
                                    disabled={mutatingTaskId !== ''}
                                >
                                    Save
                                </button>
                                <button
                                    type="button"
                                    class="rounded-lg border border-[var(--color-secondary)]/40 px-2.5 py-1.5 text-xs"
                                    onclick={cancelEditingTask}
                                    disabled={mutatingTaskId !== ''}
                                >
                                    Cancel
                                </button>
                            </div>
                        {:else}
                            <p class="truncate text-sm font-medium">
                                {typedTask.title}
                            </p>
                            <p class="mt-0.5 text-xs text-[var(--color-text)]/70">
                                Due {formatDate(typedTask.dueAt)}
                            </p>
                        {/if}
                    </div>

                    {#if typedTask.source === 'ticktick' && editingTaskId !== typedTask.id}
                        <button
                            type="button"
                            class="rounded-lg border border-[var(--color-secondary)]/40 px-2 py-1 text-xs"
                            onclick={() => startEditingTask(typedTask)}
                            disabled={mutatingTaskId !== '' || completingTaskId !== ''}
                        >
                            Edit
                        </button>
                    {/if}
                </li>
            {/each}
        </ul>
        {#if completionError}
            <p class="mt-3 text-sm text-[var(--color-error)]">
                {completionError}
            </p>
        {/if}
        {#if mutationError}
            <p class="mt-3 text-sm text-[var(--color-error)]">
                {mutationError}
            </p>
        {/if}
    {:else}
        <p
            class="mt-5 rounded-xl border border-[var(--color-secondary)]/30 bg-[var(--color-background)]/35 p-3 text-sm text-[var(--color-text)]/70"
        >
            No daily tasks available yet.
        </p>
        {#if mutationError}
            <p class="mt-3 text-sm text-[var(--color-error)]">
                {mutationError}
            </p>
        {/if}
    {/if}
</section>
