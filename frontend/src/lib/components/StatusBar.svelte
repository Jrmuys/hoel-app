<script lang="ts">
    import type { StatusBarModel } from '$lib/types/dashboard';

    let { data }: { data: StatusBarModel } = $props();

    function formatDate(dateIso: string | null): string {
        if (!dateIso) {
            return 'No successful sync yet';
        }

        return new Date(dateIso).toLocaleTimeString(undefined, {
            hour: 'numeric',
            minute: '2-digit',
        });
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
            {data.systemHealth}
        </span>
    </header>

    {#if data.integrations.length > 0}
        <div class="mt-5 grid gap-3 sm:grid-cols-2">
            {#each data.integrations as integration}
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

    {#if data.alerts.length > 0}
        <ul class="mt-5 space-y-2">
            {#each data.alerts as alert}
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
    {:else}
        <p class="mt-5 text-sm text-[var(--color-text)]/70">
            No active alerts.
        </p>
    {/if}
</section>
