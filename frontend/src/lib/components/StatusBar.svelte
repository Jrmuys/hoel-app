<script lang="ts">
	import type { StatusBarModel } from '$lib/types/dashboard';

	let { data }: { data: StatusBarModel } = $props();

	function badgeClass(level: 'critical' | 'warning' | 'info'): string {
		if (level === 'critical') {
			return 'bg-[var(--color-error)] text-white';
		}
		if (level === 'warning') {
			return 'bg-amber-500 text-white';
		}
		return 'bg-[var(--color-secondary)] text-white';
	}
</script>

<section class="rounded-xl border border-[var(--color-secondary)]/30 bg-[var(--color-surface)] p-4 shadow-sm">
	<header class="flex flex-wrap items-center justify-between gap-2">
		<h2 class="text-lg font-semibold">Status</h2>
		<span class="rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wide bg-[var(--color-primary)] text-white">
			{data.systemHealth}
		</span>
	</header>

	<div class="mt-4 grid gap-2 sm:grid-cols-2">
		{#each data.integrations as integration}
			<div class="rounded-lg border border-[var(--color-secondary)]/30 p-3">
				<p class="text-sm font-medium capitalize">{integration.service}</p>
				<p class="text-sm text-[var(--color-text)]/80">
					{integration.state} · failures: {integration.consecutiveFailures}
				</p>
			</div>
		{/each}
	</div>

	{#if data.alerts.length > 0}
		<ul class="mt-4 flex flex-wrap gap-2">
			{#each data.alerts as alert}
				<li class={`rounded-full px-3 py-1 text-xs font-medium ${badgeClass(alert.level)}`}>
					{alert.source}: {alert.message}
				</li>
			{/each}
		</ul>
	{/if}
</section>
