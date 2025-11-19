<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { projectsApi, type Project } from '$lib/api';
	import ProjectCard from '$lib/components/projects/project-card.svelte';
	import AddProjectModal from '$lib/components/projects/add-project-modal.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Plus, Search } from 'lucide-svelte';

	let projects = $state<Project[]>([]);
	let loading = $state(true);
	let error = $state('');
	let searchQuery = $state('');
	let showAddModal = $state(false);

	const loadProjects = async () => {
		try {
			loading = true;
			error = '';
			projects = await projectsApi.list();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load projects';
		} finally {
			loading = false;
		}
	};

	const handleCreateProject = async (data: { name: string; description?: string }) => {
		const newProject = await projectsApi.create(data);
		projects = [...projects, newProject];
	};

	const handleDeleteProject = async (id: string) => {
		if (!confirm('Are you sure you want to delete this project?')) return;

		try {
			await projectsApi.delete(id);
			projects = projects.filter((p) => p.id !== id);
		} catch (err) {
			alert(err instanceof Error ? err.message : 'Failed to delete project');
		}
	};

	const filteredProjects = $derived(
		projects.filter(
			(p) =>
				p.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
				p.description?.toLowerCase().includes(searchQuery.toLowerCase())
		)
	);

	onMount(() => {
		loadProjects();
	});
</script>

<div class="container mx-auto space-y-6 p-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="font-bold text-3xl">Projects</h1>
			<p class="text-muted-foreground">Manage your projects and deployments</p>
		</div>
		<Button onclick={() => (showAddModal = true)}>
			<Plus class="size-4" />
			Add Project
		</Button>
	</div>

	<div class="relative">
		<Search class="text-muted-foreground absolute left-3 top-1/2 size-4 -translate-y-1/2" />
		<Input
			bind:value={searchQuery}
			placeholder="Search projects..."
			class="pl-9"
			disabled={loading}
		/>
	</div>

	{#if loading}
		<div class="flex min-h-[400px] items-center justify-center">
			<p class="text-muted-foreground">Loading projects...</p>
		</div>
	{:else if error}
		<div class="flex min-h-[400px] items-center justify-center">
			<div class="text-center">
				<p class="text-destructive mb-2">{error}</p>
				<Button variant="outline" onclick={loadProjects}>Retry</Button>
			</div>
		</div>
	{:else if filteredProjects.length === 0}
		<div class="flex min-h-[400px] items-center justify-center">
			<div class="text-center">
				{#if searchQuery}
					<p class="text-muted-foreground">No projects found matching "{searchQuery}"</p>
				{:else}
					<p class="text-muted-foreground mb-4">No projects yet. Create your first project to get started.</p>
					<Button onclick={() => (showAddModal = true)}>
						<Plus class="size-4" />
						Create Project
					</Button>
				{/if}
			</div>
		</div>
	{:else}
		<div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
			{#each filteredProjects as project (project.id)}
				<ProjectCard
					{project}
					onDelete={handleDeleteProject}
					onclick={() => goto(`/dashboard/project/${project.id}`)}
				/>
			{/each}
		</div>
	{/if}
</div>

<AddProjectModal bind:open={showAddModal} onSubmit={handleCreateProject} />
