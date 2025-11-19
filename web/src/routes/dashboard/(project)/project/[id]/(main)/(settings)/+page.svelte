<script lang="ts">
	import { page } from '$app/state';
	import { createQuery } from '@tanstack/svelte-query';
	import { goto } from '$app/navigation';
	import {
		getProject,
		listEnvironments,
		listApplications,
		listDatabases,
		type Application,
		type Database
	} from '$lib/api';
	import { Input } from '$lib/components/ui/input';
	import { Search } from 'lucide-svelte';
	import EnvironmentTabs from '$lib/components/projects/environment-tabs.svelte';
	import AddResourceMenu from '$lib/components/projects/add-resource-menu.svelte';
	import ApplicationCard from '$lib/components/projects/application-card.svelte';
	import DatabaseCard from '$lib/components/projects/database-card.svelte';

	const projectId = $derived(page.params.id);

	let selectedEnvironmentId = $state<string | undefined>(undefined);
	let searchQuery = $state('');

	const projectQuery = createQuery(() => ({
		queryKey: ['project', projectId],
		queryFn: () => getProject(projectId)
	}));

	const environmentsQuery = createQuery(() => ({
		queryKey: ['environments', projectId],
		queryFn: () => listEnvironments(projectId)
	}));

	const applicationsQuery = createQuery(() => ({
		queryKey: ['applications', projectId],
		queryFn: () => listApplications(projectId)
	}));

	const databasesQuery = createQuery(() => ({
		queryKey: ['databases', projectId],
		queryFn: () => listDatabases(projectId)
	}));

	$effect(() => {
		if (environmentsQuery.data && !selectedEnvironmentId) {
			const productionEnv = environmentsQuery.data.find((env) => env.name === 'production');
			if (productionEnv) {
				selectedEnvironmentId = productionEnv.id;
			}
		}
	});

	const filteredResources = $derived.by(() => {
		const apps: Application[] = applicationsQuery.data || [];
		const dbs: Database[] = databasesQuery.data || [];

		let filtered: Array<{ type: 'application' | 'database'; data: Application | Database }> = [
			...apps.map((app) => ({ type: 'application' as const, data: app })),
			...dbs.map((db) => ({ type: 'database' as const, data: db }))
		];

		if (selectedEnvironmentId) {
			filtered = filtered.filter(
				(item) =>
					('environment_id' in item.data && item.data.environment_id === selectedEnvironmentId) ||
					('environment' in item.data && item.data.environment === selectedEnvironmentId)
			);
		}

		if (searchQuery.trim()) {
			const query = searchQuery.toLowerCase();
			filtered = filtered.filter((item) => item.data.name.toLowerCase().includes(query));
		}

		return filtered;
	});

	const resourceCounts = $derived.by(() => {
		const apps = applicationsQuery.data || [];
		const dbs = databasesQuery.data || [];
		const envs = environmentsQuery.data || [];

		const counts: Record<string, number> = {
			all: apps.length + dbs.length
		};

		envs.forEach((env) => {
			const envApps = apps.filter((app) => app.environment_id === env.id);
			const envDbs = dbs.filter((db) => 'environment' in db && db.environment === env.id);
			counts[env.id] = envApps.length + envDbs.length;
		});

		return counts;
	});

	function handleAddApplication() {
		if (!selectedEnvironmentId) {
			return;
		}
		goto(`/dashboard/project/${projectId}/${selectedEnvironmentId}/create-app`);
	}

	function handleAddDatabase() {
		if (!selectedEnvironmentId) {
			return;
		}
		goto(`/dashboard/project/${projectId}/${selectedEnvironmentId}/create-db`);
	}

	function handleAddTemplate() {
		if (!selectedEnvironmentId) {
			return;
		}
		goto(`/dashboard/project/${projectId}/${selectedEnvironmentId}/create-service`);
	}

	function handleAddEnvironment() {
		console.log('Add environment clicked');
	}
</script>

<div class="flex flex-col gap-6 p-6">
	{#if projectQuery.isLoading}
		<div class="text-muted-foreground">Loading project...</div>
	{:else if projectQuery.error}
		<div class="text-destructive">Error loading project: {projectQuery.error.message}</div>
	{:else if projectQuery.data}
		<div class="space-y-6">
			<div>
				<h1 class="text-3xl font-bold">{projectQuery.data.name}</h1>
				{#if projectQuery.data.description}
					<p class="text-muted-foreground">{projectQuery.data.description}</p>
				{/if}
			</div>

			<EnvironmentTabs
				environments={environmentsQuery.data || []}
				bind:selectedEnvironmentId
				onSelect={(envId) => (selectedEnvironmentId = envId)}
				onAdd={handleAddEnvironment}
				counts={resourceCounts}
			/>

			<div class="flex items-center justify-between gap-4">
				<div class="relative flex-1 max-w-md">
					<Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
					<Input
						type="text"
						placeholder="Search resources..."
						bind:value={searchQuery}
						class="pl-9"
					/>
				</div>

				<AddResourceMenu
					onAddApplication={handleAddApplication}
					onAddDatabase={handleAddDatabase}
					onAddTemplate={handleAddTemplate}
				/>
			</div>

			{#if applicationsQuery.isLoading || databasesQuery.isLoading}
				<div class="text-muted-foreground">Loading resources...</div>
			{:else if filteredResources.length === 0}
				<div class="flex flex-col items-center justify-center py-12 text-center">
					<p class="text-lg font-medium text-muted-foreground">No resources found</p>
					<p class="text-sm text-muted-foreground">
						{searchQuery.trim()
							? 'Try adjusting your search query'
							: 'Get started by adding an application or database'}
					</p>
				</div>
			{:else}
				<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
					{#each filteredResources as resource (resource.type === 'application' ? `app-${resource.data.id}` : `db-${resource.data.id}`)}
						{#if resource.type === 'application'}
							<ApplicationCard
								application={resource.data}
								onclick={() => {
									const envId = resource.data.environment_id;
									console.log(envId);
									goto(`/dashboard/project/${projectId}/${envId}/app/${resource.data.id}`);
								}}
							/>
						{:else}
							<DatabaseCard
								database={resource.data}
								onclick={() => {
									const envId = resource.data.environment_id;
									goto(`/dashboard/project/${projectId}/${envId}/db/${resource.data.id}`);
								}}
							/>
						{/if}
					{/each}
				</div>
			{/if}
		</div>
	{/if}
</div>
