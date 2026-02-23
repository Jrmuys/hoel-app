<script lang="ts">
	import type { DailyOperationsModel } from '$lib/types/dashboard';

	let { data }: { data: DailyOperationsModel } = $props();

	function formatDate(dateIso: string): string {
		return new Date(dateIso).toLocaleDateString(undefined, {
			weekday: 'short',
			month: 'short',
			day: 'numeric'
		});
	}
</script>

<section class="rounded-xl border border-[var(--color-secondary)]/30 bg-[var(--color-surface)] p-4 shadow-sm">
	<header class="flex items-center justify-between gap-2">
		<h2 class="text-lg font-semibold">Daily Operations</h2>
		{#if data.garbage.showIndicator}
			<span class="rounded-full bg-[var(--color-primary)] px-3 py-1 text-xs font-semibold text-white">
				{data.garbage.isRecyclingWeek ? 'Recycling' : 'Trash'} pickup {formatDate(data.garbage.nextPickupDate)}
			</span>
		{/if}
	</header>

	<ul class="mt-4 space-y-2">
		{#each data.tasks as task}
			<li class="flex items-center gap-3 rounded-lg border border-[var(--color-secondary)]/20 px-3 py-2">
				<input
					type="checkbox"
					checked={task.completed}
					class="h-4 w-4 rounded border-[var(--color-secondary)] text-[var(--color-primary)]"
					disabled
				/>
				<div>
					<p class="text-sm font-medium">{task.title}</p>
					<p class="text-xs text-[var(--color-text)]/70">Due {formatDate(task.dueAt)}</p>
				</div>
			</li>
		{/each}
	</ul>
</section>
