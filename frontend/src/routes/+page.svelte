<script lang="ts">
	import DailyOperations from '$lib/components/DailyOperations.svelte';
	import StatusBar from '$lib/components/StatusBar.svelte';
	import { loadDailyOperations, loadStatusBar } from '$lib/mock/dashboard';

	const statusPromise = loadStatusBar();
	const dailyPromise = loadDailyOperations();
</script>

<main class="mx-auto grid min-h-screen w-full max-w-5xl gap-4 p-4 sm:p-6">
	<h1 class="text-2xl font-bold text-[var(--color-primary)]">Household Operations Dashboard</h1>

	{#await statusPromise}
		<section class="rounded-xl border border-[var(--color-secondary)]/30 bg-[var(--color-surface)] p-4">Loading status...</section>
	{:then status}
		<StatusBar data={status} />
	{:catch}
		<section class="rounded-xl border border-[var(--color-error)]/40 bg-[var(--color-surface)] p-4 text-[var(--color-error)]">
			Unable to load status data.
		</section>
	{/await}

	{#await dailyPromise}
		<section class="rounded-xl border border-[var(--color-secondary)]/30 bg-[var(--color-surface)] p-4">Loading daily operations...</section>
	{:then daily}
		<DailyOperations data={daily} />
	{:catch}
		<section class="rounded-xl border border-[var(--color-error)]/40 bg-[var(--color-surface)] p-4 text-[var(--color-error)]">
			Unable to load daily operations.
		</section>
	{/await}
</main>
