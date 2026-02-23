<script lang="ts">
    import type {
        DailyTask,
        LogisticsPlanningModel,
    } from '$lib/types/dashboard';
    export let data: LogisticsPlanningModel;

    function formatDueDate(value: string, hasTime: boolean): string {
        const parsed = new Date(value);
        if (Number.isNaN(parsed.getTime())) {
            return hasTime ? value : `${value} (all day)`;
        }

        if (!hasTime) {
            return parsed.toLocaleDateString(undefined, {
                weekday: 'short',
                month: 'short',
                day: 'numeric',
            });
        }

        return parsed.toLocaleString(undefined, {
            weekday: 'short',
            month: 'short',
            day: 'numeric',
            hour: 'numeric',
            minute: '2-digit',
        });
    }
</script>

<section class="panel">
    <h2 class="text-lg font-semibold text-[var(--color-primary)]">
        Logistics &amp; Planning
    </h2>

    <div class="mt-4 grid gap-3 sm:grid-cols-2">
        <div
            class="rounded-xl border border-[var(--color-secondary)]/25 bg-[var(--color-background)]/35 p-3"
        >
            <p
                class="text-xs font-medium uppercase tracking-wide text-[var(--color-text)]/70"
            >
                Shopping
            </p>
            {#if data.shoppingTasks.length === 0}
                <p class="mt-2 text-sm text-[var(--color-text)]/70">
                    No shopping items right now.
                </p>
            {:else}
                <ul class="mt-2 space-y-1.5">
                    {#each data.shoppingTasks as task}
                        {@const typedTask = task as DailyTask}
                        <li class="text-sm">
                            <span class="font-medium">{typedTask.title}</span>
                            <span class="text-[var(--color-text)]/70">
                                • {formatDueDate(
                                    typedTask.dueAt,
                                    typedTask.hasTime,
                                )}
                            </span>
                        </li>
                    {/each}
                </ul>
            {/if}
        </div>

        <div
            class="rounded-xl border border-[var(--color-secondary)]/25 bg-[var(--color-background)]/35 p-3"
        >
            <p
                class="text-xs font-medium uppercase tracking-wide text-[var(--color-text)]/70"
            >
                Maintenance
            </p>
            {#if data.maintenanceTasks.length === 0}
                <p class="mt-2 text-sm text-[var(--color-text)]/70">
                    No maintenance tasks scheduled.
                </p>
            {:else}
                <ul class="mt-2 space-y-1.5">
                    {#each data.maintenanceTasks as task}
                        {@const typedTask = task as DailyTask}
                        <li class="text-sm">
                            <span class="font-medium">{typedTask.title}</span>
                            <span class="text-[var(--color-text)]/70">
                                • {formatDueDate(
                                    typedTask.dueAt,
                                    typedTask.hasTime,
                                )}
                            </span>
                        </li>
                    {/each}
                </ul>
            {/if}
        </div>
    </div>
</section>
