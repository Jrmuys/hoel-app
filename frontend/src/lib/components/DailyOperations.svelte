<script lang="ts">
    import type { DailyOperationsModel } from '$lib/types/dashboard';

    let { data }: { data: DailyOperationsModel } = $props();

    function formatDate(dateIso: string): string {
        return new Date(dateIso).toLocaleDateString(undefined, {
            weekday: 'short',
            month: 'short',
            day: 'numeric',
        });
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
        {#if data.garbage.showIndicator}
            <span
                class="rounded-full border border-[var(--color-primary)]/30 bg-[var(--color-primary)]/10 px-3 py-1 text-xs font-semibold text-[var(--color-primary)]"
            >
                {data.garbage.isRecyclingWeek ? 'Recycling' : 'Trash'} pickup {formatDate(
                    data.garbage.nextPickupDate,
                )}
            </span>
        {/if}
    </header>

    <ul class="mt-5 space-y-2.5">
        {#each data.tasks as task}
            <li
                class="flex items-center gap-3 rounded-xl border border-[var(--color-secondary)]/25 bg-[var(--color-background)]/35 px-3 py-2.5"
            >
                <input
                    type="checkbox"
                    checked={task.completed}
                    class="h-4 w-4 rounded border-[var(--color-secondary)] text-[var(--color-primary)] accent-[var(--color-primary)]"
                    disabled
                />
                <div class="min-w-0">
                    <p class="truncate text-sm font-medium">{task.title}</p>
                    <p class="mt-0.5 text-xs text-[var(--color-text)]/70">
                        Due {formatDate(task.dueAt)}
                    </p>
                </div>
            </li>
        {/each}
    </ul>
</section>
