<script lang="ts">
    import DailyOperations from '$lib/components/DailyOperations.svelte';
    import StatusBar from '$lib/components/StatusBar.svelte';
    import { loadDailyOperations, loadStatusBar } from '$lib/mock/dashboard';

    const statusPromise = loadStatusBar();
    const dailyPromise = loadDailyOperations();

    const todayLabel = new Intl.DateTimeFormat(undefined, {
        weekday: 'long',
        month: 'long',
        day: 'numeric',
    }).format(new Date());
</script>

<main
    class="app-shell mx-auto grid min-h-screen w-full max-w-6xl gap-6 p-4 sm:p-8"
>
    <section class="panel">
        <p
            class="text-xs font-semibold uppercase tracking-[0.18em] text-[var(--color-secondary)]"
        >
            Household Operations
        </p>
        <div class="mt-2 flex flex-wrap items-end justify-between gap-3">
            <h1 class="text-3xl font-semibold text-[var(--color-primary)]">
                Dashboard
            </h1>
            <p class="text-sm text-[var(--color-text)]/70">{todayLabel}</p>
        </div>
    </section>

    {#await statusPromise}
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

    {#await dailyPromise}
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
</main>
