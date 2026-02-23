<script lang="ts">
    import DailyOperations from '$lib/components/DailyOperations.svelte';
    import LogisticsPlanning from '$lib/components/LogisticsPlanning.svelte';
    import StatusBar from '$lib/components/StatusBar.svelte';
    import type { PageData } from './$types';

    let { data }: { data: PageData } = $props();
</script>

<main
    class="app-shell mx-auto grid min-h-screen w-full max-w-6xl gap-6 p-4 sm:p-8"
>
    <section class="panel">
        <div class="flex flex-wrap items-end justify-between gap-3">
            <h1 class="text-3xl font-semibold text-[var(--color-primary)]">
                Dashboard
            </h1>
            <p class="text-sm text-[var(--color-text)]/70">{data.todayLabel}</p>
        </div>
    </section>

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
