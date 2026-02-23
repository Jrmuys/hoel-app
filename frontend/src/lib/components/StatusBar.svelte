<script lang="ts">
    import { clearStatusAlerts, loadStatusBar } from '$lib/api/dashboard';
    import type { StatusBarModel } from '$lib/types/dashboard';

    const EMPTY_STATUS: StatusBarModel = {
        systemHealth: 'ok',
        alerts: [],
        integrations: [],
    };

    let { data }: { data: StatusBarModel } = $props();
    let viewData = $state<StatusBarModel>(EMPTY_STATUS);
    let isClearingAlerts = $state(false);
    let clearAlertsError = $state('');

    $effect(() => {
        viewData = data;
    });

    function formatDate(dateIso: string | null): string {
        if (!dateIso) {
            return 'No successful sync yet';
        }

        return new Date(dateIso).toLocaleTimeString(undefined, {
            hour: 'numeric',
            minute: '2-digit',
        });
    }

    async function handleClearAlerts() {
        if (isClearingAlerts || viewData.alerts.length === 0) {
            return;
        }

        isClearingAlerts = true;
        clearAlertsError = '';

        try {
            await clearStatusAlerts();
            viewData = await loadStatusBar();
        } catch (error) {
            clearAlertsError =
                error instanceof Error
                    ? error.message
                    : 'Unable to clear status alerts right now.';
        } finally {
            isClearingAlerts = false;
        }
    }
</script>

<section class="panel">
    <header class="flex flex-wrap items-start justify-between gap-3">
        <div>
            <h2 class="text-lg font-semibold">Status</h2>
            <p class="mt-1 text-sm text-[var(--color-text)]/70">
                Integration health and operational alerts
            </p>
        </div>
        <span
            class="rounded-full border border-[var(--color-primary)]/30 bg-[var(--color-primary)]/10 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-[var(--color-primary)]"
        >
            {viewData.systemHealth}
        </span>
    </header>

    {#if viewData.integrations.length > 0}
        <div class="mt-5 grid gap-3 sm:grid-cols-2">
            {#each viewData.integrations as integration}
                <div
                    class="rounded-xl border border-[var(--color-secondary)]/18 bg-[var(--color-background)]/35 p-3 transition-colors duration-150 hover:bg-[var(--color-background)]/55"
                >
                    <div class="flex items-center justify-between gap-3">
                        <p class="text-sm font-semibold capitalize">
                            {integration.service}
                        </p>
                        <span
                            class="text-xs font-medium uppercase tracking-wide text-[var(--color-text)]/70"
                        >
                            {integration.state}
                        </span>
                    </div>
                    <p class="mt-2 text-xs text-[var(--color-text)]/70">
                        Last success: {formatDate(integration.lastSuccessAt)}
                    </p>
                    <p class="mt-1 text-xs text-[var(--color-text)]/70">
                        Consecutive failures: {integration.consecutiveFailures}
                    </p>
                </div>
            {/each}
        </div>
    {:else}
        <p
            class="mt-5 rounded-xl border border-[var(--color-secondary)]/18 bg-[var(--color-background)]/35 p-3 text-sm text-[var(--color-text)]/70"
        >
            No integration telemetry yet. Start the backend integrations to
            populate status.
        </p>
    {/if}

    {#if viewData.alerts.length > 0}
        <div class="mt-5 flex items-center justify-between gap-2">
            <p class="text-xs text-[var(--color-text)]/70">
                {viewData.alerts.length} active alert{viewData.alerts.length === 1
                    ? ''
                    : 's'}
            </p>
            <button
                type="button"
                class="inline-flex h-8 items-center rounded-lg border border-[var(--color-secondary)]/40 px-3 text-xs"
                onclick={handleClearAlerts}
                disabled={isClearingAlerts}
            >
                {isClearingAlerts ? 'Clearing…' : 'Clear alerts'}
            </button>
        </div>
        <ul class="mt-2 space-y-2">
            {#each viewData.alerts as alert}
                <li
                    class="rounded-lg border border-[var(--color-secondary)]/18 bg-[var(--color-background)]/35 px-3 py-2"
                    title={`${alert.source}: ${alert.message}`}
                >
                    <p
                        class="truncate font-mono text-xs text-[var(--color-text)]/65"
                    >
                        {alert.source}: {alert.message}
                    </p>
                </li>
            {/each}
        </ul>

        {#if clearAlertsError}
            <p class="mt-3 text-sm text-[var(--color-error)]">
                {clearAlertsError}
            </p>
        {/if}
    {:else}
        <p class="mt-5 text-sm text-[var(--color-text)]/70">
            No active alerts.
        </p>
    {/if}
</section>
