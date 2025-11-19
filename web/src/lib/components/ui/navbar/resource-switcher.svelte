<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { applicationsApi, databasesApi } from '$lib/api';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { Box, Database, Check, ChevronDown } from 'lucide-svelte';
	import type { Snippet } from 'svelte';

	interface Props {
		projectId: string;
		currentResourceId?: string;
		currentResourceType?: 'application' | 'database';
		children?: Snippet;
	}

	let { projectId, currentResourceId, currentResourceType, children }: Props = $props();

	const appsQuery = createQuery(() => ({
		queryKey: ['applications', projectId],
		queryFn: () => applicationsApi.list(projectId),
		enabled: !!projectId
	}));

	const dbsQuery = createQuery(() => ({
		queryKey: ['databases', projectId],
		queryFn: () => databasesApi.list(projectId),
		enabled: !!projectId
	}));

	const currentResource = $derived(() => {
		if (currentResourceType === 'application') {
			return appsQuery.data?.find((app) => app.id === currentResourceId);
		} else if (currentResourceType === 'database') {
			return dbsQuery.data?.find((db) => db.id === currentResourceId);
		}
		return null;
	});
</script>

<DropdownMenu.Root>
	<DropdownMenu.Trigger class="flex items-center gap-1.5 hover:text-foreground transition-colors">
		{#if children}
			{@render children()}
		{:else if currentResource()}
			{#if currentResourceType === 'application'}
				<Box class="w-4 h-4" />
			{:else}
				<Database class="w-4 h-4" />
			{/if}
			{currentResource()?.name}
			<ChevronDown class="w-3 h-3" />
		{:else}
			<Box class="w-4 h-4" />
			Resource
			<ChevronDown class="w-3 h-3" />
		{/if}
	</DropdownMenu.Trigger>
	<DropdownMenu.Content align="start" class="w-56">
		<DropdownMenu.Label>Switch Resource</DropdownMenu.Label>
		<DropdownMenu.Separator />

		{#if appsQuery.data && appsQuery.data.length > 0}
			<DropdownMenu.GroupHeading>Applications</DropdownMenu.GroupHeading>
			{#each appsQuery.data as app (app.id)}
				<DropdownMenu.Item
					class="flex items-center justify-between"
					onclick={() => {
						window.location.href = `/dashboard/project/${projectId}/${app.environment_id}/app/${app.id}`;
					}}
				>
					<span class="flex items-center gap-2">
						<Box class="w-4 h-4" />
						{app.name}
					</span>
					{#if app.id === currentResourceId && currentResourceType === 'application'}
						<Check class="w-4 h-4" />
					{/if}
				</DropdownMenu.Item>
			{/each}
		{/if}

		{#if appsQuery.data && appsQuery.data.length > 0 && dbsQuery.data && dbsQuery.data.length > 0}
			<DropdownMenu.Separator />
		{/if}

		{#if dbsQuery.data && dbsQuery.data.length > 0}
			<DropdownMenu.GroupHeading>Databases</DropdownMenu.GroupHeading>
			{#each dbsQuery.data as db (db.id)}
				<DropdownMenu.Item
					class="flex items-center justify-between"
					onclick={() => {
						window.location.href = `/dashboard/project/${projectId}/${db.environment_id}/db/${db.id}`;
					}}
				>
					<span class="flex items-center gap-2">
						<Database class="w-4 h-4" />
						{db.name}
					</span>
					{#if db.id === currentResourceId && currentResourceType === 'database'}
						<Check class="w-4 h-4" />
					{/if}
				</DropdownMenu.Item>
			{/each}
		{/if}

		{#if (!appsQuery.data || appsQuery.data.length === 0) && (!dbsQuery.data || dbsQuery.data.length === 0)}
			{#if appsQuery.isPending || dbsQuery.isPending}
				<DropdownMenu.Item disabled>Loading...</DropdownMenu.Item>
			{:else}
				<DropdownMenu.Item disabled>No resources found</DropdownMenu.Item>
			{/if}
		{/if}
	</DropdownMenu.Content>
</DropdownMenu.Root>
