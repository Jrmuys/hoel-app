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

    function formatOptionalDate(dateIso: string): string {
        if (!dateIso) {
            return 'Not available';
        }

        return formatDate(dateIso);
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
            {#if data.garbage.showTrashTakeOutReminder}
                <span
                    class="inline-flex items-center gap-1 rounded-full border border-[var(--color-primary)]/30 bg-[var(--color-primary)]/10 px-3 py-1 text-xs font-semibold text-[var(--color-primary)]"
                >
                    <svg
                        aria-hidden="true"
                        viewBox="0 0 24 24"
                        class="h-3.5 w-3.5"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                    >
                        <path d="M3 6h18" />
                        <path d="M8 6V4h8v2" />
                        <path d="M19 6l-1 14H6L5 6" />
                        <path d="M10 11v6" />
                        <path d="M14 11v6" />
                    </svg>
                    Take out trash tonight
                </span>
            {/if}
            {#if data.garbage.showRecyclingTakeOutReminder}
                <span
                    class="inline-flex items-center gap-1 rounded-full border border-[var(--color-primary)]/30 bg-[var(--color-primary)]/10 px-3 py-1 text-xs font-semibold text-[var(--color-primary)]"
                >
                    <svg
                        aria-hidden="true"
                        viewBox="0 0 24 24"
                        class="h-3.5 w-3.5"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                    >
                        <path d="M7 7h4l2-3" />
                        <path d="M17 7l-2-3h-1" />
                        <path d="M17 7l2 3-2 3" />
                        <path d="M7 7l-2 3" />
                        <path d="M5 10h4" />
                        <path d="M9 10l2 3-2 3" />
                        <path d="M11 16h4" />
                        <path d="M15 16l2-3" />
                    </svg>
                    Take out recycling tonight
                </span>
            {/if}
            {#if !data.garbage.showTrashTakeOutReminder && !data.garbage.showRecyclingTakeOutReminder}
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
                <svg
                    aria-hidden="true"
                    viewBox="0 0 24 24"
                    class="h-3.5 w-3.5"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                >
                    <path d="M3 6h18" />
                    <path d="M8 6V4h8v2" />
                    <path d="M19 6l-1 14H6L5 6" />
                    <path d="M10 11v6" />
                    <path d="M14 11v6" />
                </svg>
                Next Trash
            </p>
            <p class="mt-1 text-sm font-semibold">
                {formatOptionalDate(data.garbage.nextTrashPickupDate)}
            </p>
        </div>
        <div
            class="rounded-xl border border-[var(--color-secondary)]/25 bg-[var(--color-background)]/35 p-3"
        >
            <p
                class="inline-flex items-center gap-1 text-xs font-medium uppercase tracking-wide text-[var(--color-text)]/70"
            >
                <svg
                    aria-hidden="true"
                    viewBox="0 0 24 24"
                    class="h-3.5 w-3.5"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                >
                    <path d="M7 7h4l2-3" />
                    <path d="M17 7l-2-3h-1" />
                    <path d="M17 7l2 3-2 3" />
                    <path d="M7 7l-2 3" />
                    <path d="M5 10h4" />
                    <path d="M9 10l2 3-2 3" />
                    <path d="M11 16h4" />
                    <path d="M15 16l2-3" />
                </svg>
                Next Recycling
            </p>
            <p class="mt-1 text-sm font-semibold">
                {formatOptionalDate(data.garbage.nextRecyclingPickupDate)}
            </p>
        </div>
    </div>

    {#if data.tasks.length > 0}
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
    {:else}
        <p
            class="mt-5 rounded-xl border border-[var(--color-secondary)]/30 bg-[var(--color-background)]/35 p-3 text-sm text-[var(--color-text)]/70"
        >
            No daily tasks available yet.
        </p>
    {/if}
</section>
