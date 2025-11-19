<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { Settings, Shield, RefreshCw, Database } from 'lucide-svelte';

	const navItems = [
		{ path: 'general', label: 'General', icon: Settings },
		{ path: 'advanced', label: 'Advanced', icon: Shield },
		{ path: 'updates', label: 'Updates', icon: RefreshCw },
		{ path: 'backup', label: 'Backup', icon: Database }
	];

	const isActive = (path: string) => {
		return (
			page.url.pathname.endsWith(`/${path}`) ||
			(page.url.pathname.endsWith('/settings') && path === 'general')
		);
	};
</script>

<div class="container max-w-7xl py-8">
	<div class="flex gap-6">
		<nav class="w-56 flex-shrink-0">
			<div class="space-y-1">
				{#each navItems as item}
					<button
						onclick={() => goto(`/dashboard/settings/${item.path}`)}
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
			<slot />
		</div>
	</div>
</div>
