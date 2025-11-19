<script lang="ts">
	import { page } from '$app/state';
	import { createQuery } from '@tanstack/svelte-query';
	import {
		projectsApi,
		environmentsApi,
		databasesApi,
		applicationsApi,
		organizationsApi
	} from '$lib/api';
	import { LoaderCircle } from 'lucide-svelte';
	import * as Breadcrumb from '$lib/components/ui/breadcrumb';
	import OrgSwitcher from './org-switcher.svelte';
	import ProjectSwitcher from './project-switcher.svelte';
	import ResourceSwitcher from './resource-switcher.svelte';

	const projectId = $derived(page.params.id);
	const envId = $derived(page.params.env_id);
	const resId = $derived(page.params.res_id);

	const resourceType = $derived(
		page.url.pathname.includes('/db/')
			? 'database'
			: page.url.pathname.includes('/app/')
				? 'application'
				: null
	);

	const orgQuery = createQuery(() => ({
		queryKey: ['organizations'],
		queryFn: () => {
			console.log('Organizations query executing...');
			return organizationsApi.list();
		},
		staleTime: 5 * 60 * 1000,
		refetchOnMount: true,
		enabled: true
	}));

	const projectQuery = createQuery(() => ({
		queryKey: ['project', projectId],
		queryFn: () => projectsApi.get(projectId!),
		enabled: !!projectId
	}));

	const environmentQuery = createQuery(() => ({
		queryKey: ['environment', projectId, envId],
		queryFn: () => environmentsApi.get(projectId!, envId!),
		enabled: !!projectId && !!envId
	}));

	const databaseQuery = createQuery(() => ({
		queryKey: ['database', projectId, resId],
		queryFn: () => databasesApi.get(projectId!, resId!),
		enabled: !!projectId && !!resId && resourceType === 'database'
	}));

	const applicationQuery = createQuery(() => ({
		queryKey: ['application', projectId, resId],
		queryFn: () => applicationsApi.get(projectId!, resId!),
		enabled: !!projectId && !!resId && resourceType === 'application'
	}));

	const currentOrg = $derived(orgQuery.data?.[0]);
	const shouldShowBreadcrumbs = $derived(true);
</script>

{#if shouldShowBreadcrumbs}
	<Breadcrumb.Root>
		<Breadcrumb.List>
			<Breadcrumb.Item>
				{#if currentOrg}
					<OrgSwitcher currentOrgId={currentOrg.id}>
						{currentOrg.name}
					</OrgSwitcher>
				{:else}
					<span class="flex items-center gap-1.5">
						<LoaderCircle class="w-3 h-3 animate-spin" />
						Loading...
					</span>
				{/if}
			</Breadcrumb.Item>

			{#if projectId}
				<Breadcrumb.Separator>/</Breadcrumb.Separator>
				<Breadcrumb.Item>
					{#if projectQuery.data}
						<ProjectSwitcher currentProjectId={projectId}>
							{projectQuery.data.name}
						</ProjectSwitcher>
					{:else}
						<span class="flex items-center gap-1.5">
							<LoaderCircle class="w-3 h-3 animate-spin" />
							...
						</span>
					{/if}
				</Breadcrumb.Item>
			{/if}

			{#if envId && environmentQuery.data}
				<Breadcrumb.Separator>/</Breadcrumb.Separator>
				<Breadcrumb.Item>
					<Breadcrumb.Page>{environmentQuery.data.name}</Breadcrumb.Page>
				</Breadcrumb.Item>
			{/if}

			{#if resId && resourceType && projectId}
				<Breadcrumb.Separator>/</Breadcrumb.Separator>
				<Breadcrumb.Item>
					{#if resourceType === 'database' && databaseQuery.data}
						<ResourceSwitcher {projectId} currentResourceId={resId} currentResourceType="database">
							{databaseQuery.data.name}
						</ResourceSwitcher>
					{:else if resourceType === 'application' && applicationQuery.data}
						<ResourceSwitcher
							{projectId}
							currentResourceId={resId}
							currentResourceType="application"
						>
							{applicationQuery.data.name}
						</ResourceSwitcher>
					{:else}
						<span class="flex items-center gap-1.5">
							<LoaderCircle class="w-3 h-3 animate-spin" />
							...
						</span>
					{/if}
				</Breadcrumb.Item>
			{/if}
		</Breadcrumb.List>
	</Breadcrumb.Root>
{/if}
