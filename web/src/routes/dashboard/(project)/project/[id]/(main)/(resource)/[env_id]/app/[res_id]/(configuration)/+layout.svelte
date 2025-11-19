<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { Settings, HardDrive, Gauge, BarChart3, Scale, LayoutGrid } from 'lucide-svelte';

	const projectId = $derived(page.params.id);
	const envId = $derived(page.params.env_id);
	const resId = $derived(page.params.res_id);
	let { children } = $props();

	const navItems = [
		{ path: 'general', label: 'General', icon: LayoutGrid },
		{ path: 'networking', label: 'Networking', icon: LayoutGrid },
		{ path: 'settings', label: 'Settings', icon: Settings },
		{ path: 'storage', label: 'Persistent Storage', icon: HardDrive },
		{ path: 'limits', label: 'Resource Limits', icon: Gauge },
		{ path: 'metrics', label: 'Metrics', icon: BarChart3 },
		{ path: 'scaling', label: 'Scaling', icon: Scale }
	];

	const isActive = (path: string) => {
		return page.url.pathname.endsWith(`/${path}`);
	};
</script>

<div class="container max-w-7xl py-8">
	<div class="flex gap-6">
		<nav class="w-56 flex-shrink-0">
			<div class="space-y-1">
				{#each navItems as item}
					<button
						onclick={() =>
							goto(`/dashboard/project/${projectId}/${envId}/app/${resId}/${item.path}`)}
						class="w-full flex items-center gap-3 px-4 py-2 text-sm rounded-lg transition-colors {isActive(
							item.path
						)
							? 'bg-accent text-accent-foreground font-medium'
							: 'text-muted-foreground hover:bg-accent/50'}"
					>
						<item.icon class="h-4 w-4" />
						{item.label}
					</button>
				{/each}
			</div>
		</nav>

		<div class="flex-1">
			{@render children()}
		</div>
	</div>
</div>
