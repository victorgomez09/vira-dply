<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { databasesApi } from '$lib/api/databases';
	import { toast } from 'svelte-sonner';
	import {
		Database as DatabaseIcon,
		Play,
		Square,
		RefreshCw,
		Settings,
		HardDrive,
		Gauge,
		BarChart3,
		Scale,
		Clock,
		LayoutGrid
	} from 'lucide-svelte';

	let { children } = $props();
	const projectId = $derived(page.params.id);
	const envId = $derived(page.params.env_id);
	const resId = $derived(page.params.res_id);

	const queryClient = useQueryClient();

	const databaseQuery = createQuery(() => ({
		queryKey: ['database', projectId, resId],
		queryFn: () => databasesApi.get(projectId, resId),
		enabled: !!projectId && !!resId
	}));

	const database = $derived(databaseQuery.data);

	const startMutation = createMutation(() => ({
		mutationFn: () => databasesApi.start(projectId, resId),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['database', projectId, resId] });
			toast.success('Database started successfully');
		},
		onError: (error: Error) => {
			toast.error(`Failed to start database: ${error.message}`);
		}
	}));

	const stopMutation = createMutation(() => ({
		mutationFn: () => databasesApi.stop(projectId, resId),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['database', projectId, resId] });
			toast.success('Database stopped successfully');
		},
		onError: (error: Error) => {
			toast.error(`Failed to stop database: ${error.message}`);
		}
	}));

	const restartMutation = createMutation(() => ({
		mutationFn: () => databasesApi.restart(projectId, resId),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['database', projectId, resId] });
			toast.success('Database restarted successfully');
		},
		onError: (error: Error) => {
			toast.error(`Failed to restart database: ${error.message}`);
		}
	}));

	const isAnyActionPending = $derived(
		startMutation.isPending || stopMutation.isPending || restartMutation.isPending
	);

	function getStatusBadgeVariant(
		status: string
	): 'default' | 'secondary' | 'destructive' | 'outline' {
		switch (status) {
			case 'running':
				return 'default';
			case 'stopped':
				return 'secondary';
			case 'failed':
				return 'destructive';
			default:
				return 'outline';
		}
	}

	const navItems = [
		{ path: 'overview', label: 'Overview', icon: LayoutGrid },
		{ path: 'backups', label: 'Backups', icon: Clock },
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
	<div class="mb-6">
		{#if database}
			<div class="flex items-start justify-between">
				<div>
					<div class="flex items-center gap-3 mb-2">
						<DatabaseIcon class="h-8 w-8" />
						<h1 class="text-3xl font-bold">{database.name}</h1>
						<Badge variant={getStatusBadgeVariant(database.status)}>
							{database.status}
						</Badge>
					</div>
					{#if database.description}
						<p class="text-muted-foreground">{database.description}</p>
					{/if}
				</div>

				<div class="flex gap-2">
					{#if database.status === 'stopped' || database.status === 'created'}
						<Button
							variant="outline"
							size="sm"
							disabled={isAnyActionPending}
							onclick={() => startMutation.mutate()}
						>
							<Play class="mr-2 h-4 w-4" />
							{startMutation.isPending ? 'Starting...' : 'Start'}
						</Button>
					{:else if database.status === 'running'}
						<Button
							variant="outline"
							size="sm"
							disabled={isAnyActionPending}
							onclick={() => stopMutation.mutate()}
						>
							<Square class="mr-2 h-4 w-4" />
							{stopMutation.isPending ? 'Stopping...' : 'Stop'}
						</Button>
					{/if}
					{#if database.status !== 'created'}
						<Button
							variant="outline"
							size="sm"
							disabled={isAnyActionPending}
							onclick={() => restartMutation.mutate()}
						>
							<RefreshCw class="mr-2 h-4 w-4 {restartMutation.isPending ? 'animate-spin' : ''}" />
							{restartMutation.isPending ? 'Restarting...' : 'Restart'}
						</Button>
					{/if}
				</div>
			</div>
		{/if}
	</div>

	<div class="flex gap-6">
		<nav class="w-56 flex-shrink-0">
			<div class="space-y-1">
				{#each navItems as item}
					<button
						onclick={() =>
							goto(`/dashboard/project/${projectId}/${envId}/db/${resId}/${item.path}`)}
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
