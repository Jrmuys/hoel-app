<script lang="ts">
    import type { StatusBarModel } from '$lib/types/dashboard';

    let { data }: { data: StatusBarModel } = $props();

    function badgeClass(level: 'critical' | 'warning' | 'info'): string {
        if (level === 'critical') {
            return 'border-[var(--color-error)]/30 bg-[var(--color-error)]/10 text-[var(--color-error)]';
        }
        if (level === 'warning') {
            return 'border-[var(--color-secondary)]/40 bg-[var(--color-secondary)]/20 text-[var(--color-text)]';
        }
        return 'border-[var(--color-primary)]/30 bg-[var(--color-primary)]/10 text-[var(--color-primary)]';
    }

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

    <div class="mt-5 grid gap-3 sm:grid-cols-2">
        {#each data.integrations as integration}
            <div
                class="rounded-xl border border-[var(--color-secondary)]/30 bg-[var(--color-background)]/35 p-3"
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

    {#if data.alerts.length > 0}
        <ul class="mt-5 flex flex-wrap gap-2">
            {#each data.alerts as alert}
                <li
                    class={`rounded-full border px-3 py-1 text-xs font-medium ${badgeClass(alert.level)}`}
                >
                    {alert.source}: {alert.message}
                </li>
            {/each}
        </ul>
    {/if}
</section>
