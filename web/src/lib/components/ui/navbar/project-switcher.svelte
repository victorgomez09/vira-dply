<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { projectsApi } from '$lib/api';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { FolderGit2, Check, ChevronDown } from 'lucide-svelte';
	import type { Snippet } from 'svelte';

	interface Props {
		currentProjectId?: string;
		children?: Snippet;
	}

	let { currentProjectId, children }: Props = $props();

	const projectsQuery = createQuery(() => ({
		queryKey: ['projects'],
		queryFn: () => projectsApi.list()
	}));

	const currentProject = $derived(
		projectsQuery.data?.find((proj) => proj.id === currentProjectId)
	);
</script>

<DropdownMenu.Root>
	<DropdownMenu.Trigger class="flex items-center gap-1.5 hover:text-foreground transition-colors">
		{#if children}
			{@render children()}
		{:else if currentProject}
			<FolderGit2 class="w-4 h-4" />
			{currentProject.name}
			<ChevronDown class="w-3 h-3" />
		{:else}
			<FolderGit2 class="w-4 h-4" />
			Project
			<ChevronDown class="w-3 h-3" />
		{/if}
	</DropdownMenu.Trigger>
	<DropdownMenu.Content align="start">
		<DropdownMenu.Label>Switch Project</DropdownMenu.Label>
		<DropdownMenu.Separator />
		{#if projectsQuery.isPending}
			<DropdownMenu.Item disabled>Loading...</DropdownMenu.Item>
		{:else if projectsQuery.isError}
			<DropdownMenu.Item disabled>Failed to load projects</DropdownMenu.Item>
		{:else if projectsQuery.data}
			{#each projectsQuery.data as project (project.id)}
				<DropdownMenu.Item
					class="flex items-center justify-between"
					onclick={() => {
						window.location.href = `/dashboard/project/${project.id}`;
					}}
				>
					<span class="flex items-center gap-2">
						<FolderGit2 class="w-4 h-4" />
						{project.name}
					</span>
					{#if project.id === currentProject?.id}
						<Check class="w-4 h-4" />
					{/if}
				</DropdownMenu.Item>
			{/each}
		{/if}
	</DropdownMenu.Content>
</DropdownMenu.Root>
