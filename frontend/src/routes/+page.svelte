<script lang="ts">
    import DailyOperations from '$lib/components/DailyOperations.svelte';
    import LogisticsPlanning from '$lib/components/LogisticsPlanning.svelte';
    import StatusBar from '$lib/components/StatusBar.svelte';
    import type { PageData } from './$types';

    let { data }: { data: PageData } = $props();
    let showStatus = $state(false);
</script>

<main
    class="app-shell mx-auto grid min-h-dvh w-full max-w-6xl gap-6 p-4 sm:p-8"
>
    <section class="panel">
        <div class="flex flex-wrap items-end justify-between gap-3">
            <h1 class="text-3xl font-semibold text-[var(--color-primary)]">
                Dashboard
            </h1>
            <div class="flex flex-wrap items-center gap-2 sm:gap-3">
                <p class="text-sm text-[var(--color-text)]/70">{data.todayLabel}</p>
                {#await data.statusPromise}
                    <div
                        class="inline-flex h-9 items-center rounded-lg border border-[var(--color-secondary)]/30 bg-[var(--color-background)]/35 px-3 text-xs text-[var(--color-text)]/70"
                    >
                        Status loading…
                    </div>
                {:then status}
                    <div
                        class="inline-flex h-9 items-center gap-2 rounded-lg border border-[var(--color-secondary)]/30 bg-[var(--color-background)]/35 px-3"
                    >
                        <span
                            class="text-[10px] font-semibold uppercase tracking-wide text-[var(--color-text)]/65"
                        >
                            Status
                        </span>
                        <span class="text-xs font-medium text-[var(--color-text)]/80">
                            {status.systemHealth}
                        </span>
                        <span class="text-xs text-[var(--color-text)]/65">
                            • {status.alerts.length} alert{status.alerts.length ===
                            1
                                ? ''
                                : 's'}
                        </span>
                    </div>
                {:catch}
                    <div
                        class="inline-flex h-9 items-center rounded-lg border border-[var(--color-error)]/35 bg-[var(--color-error)]/8 px-3 text-xs text-[var(--color-error)]"
                    >
                        Status unavailable
                    </div>
                {/await}
                <button
                    type="button"
                    class="inline-flex h-9 items-center rounded-lg border border-[var(--color-secondary)]/40 px-3 text-sm"
                    onclick={() => (showStatus = !showStatus)}
                    aria-expanded={showStatus}
                >
                    {showStatus ? 'Hide Status' : 'Show Status'}
                </button>
            </div>
        </div>
    </section>

    {#if showStatus}
        {#await data.statusPromise}
            <section class="panel text-sm text-[var(--color-text)]/70">
                Loading status...
            </section>
        {:then status}
            <StatusBar data={status} />
        {:catch}
            <section
                class="panel border-[var(--color-error)]/40 text-sm text-[var(--color-error)]"
            >
                Unable to load status data.
            </section>
        {/await}
    {/if}

    {#await data.dailyPromise}
        <section class="panel text-sm text-[var(--color-text)]/70">
            Loading daily operations...
        </section>
    {:then daily}
        <DailyOperations data={daily} />
    {:catch}
        <section
            class="panel border-[var(--color-error)]/40 text-sm text-[var(--color-error)]"
        >
            Unable to load daily operations.
        </section>
    {/await}

    {#await data.logisticsPromise}
        <section class="panel text-sm text-[var(--color-text)]/70">
            Loading logistics &amp; planning...
        </section>
    {:then logistics}
        <LogisticsPlanning data={logistics} />
    {:catch}
        <section
            class="panel border-[var(--color-error)]/40 text-sm text-[var(--color-error)]"
        >
            Unable to load logistics &amp; planning.
        </section>
    {/await}
</main>
