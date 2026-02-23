<script lang="ts">
    import { Recycle, Trash2 } from 'lucide-svelte';
    import {
        completeTickTickTask,
        loadDailyOperations,
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
                        <p class="truncate text-sm font-medium">
                            {typedTask.title}
                        </p>
                        <p class="mt-0.5 text-xs text-[var(--color-text)]/70">
                            Due {formatDate(typedTask.dueAt)}
                        </p>
                    </div>
                </li>
            {/each}
        </ul>
        {#if completionError}
            <p class="mt-3 text-sm text-[var(--color-error)]">
                {completionError}
            </p>
        {/if}
    {:else}
        <p
            class="mt-5 rounded-xl border border-[var(--color-secondary)]/30 bg-[var(--color-background)]/35 p-3 text-sm text-[var(--color-text)]/70"
        >
            No daily tasks available yet.
        </p>
    {/if}
</section>
