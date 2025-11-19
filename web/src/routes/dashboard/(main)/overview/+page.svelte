<script lang="ts">
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import { Ellipsis, Globe, Plus } from 'lucide-svelte';
	import {
		activitiesApi,
		serversApi,
		projectsApi,
		type Activity,
		type Server,
		type Project
	} from '$lib/api';
	import { onMount } from 'svelte';
	import Skeleton from '$lib/components/ui/skeleton/skeleton.svelte';

	let projects = $state<Project[]>([]);
	let activities = $state<Activity[]>([]);
	let servers = $state<Server[]>([]);
	let loadingProjects = $state(true);
	let loadingActivities = $state(true);
	let loadingServers = $state(true);

	onMount(async () => {
		try {
			projects = await projectsApi.list();
		} catch (error) {
			console.error('Failed to load projects:', error);
		} finally {
			loadingProjects = false;
		}

		try {
			const orgId = '00000000-0000-0000-0000-000000000000';
			const response = await activitiesApi.getRecent(orgId, 10);
			activities = response.activities;
		} catch (error) {
			console.error('Failed to load activities:', error);
		} finally {
			loadingActivities = false;
		}

		try {
			servers = await serversApi.list();
		} catch (error) {
			console.error('Failed to load servers:', error);
		} finally {
			loadingServers = false;
		}
	});

	function goToProject(projectId: string) {
		goto(`/dashboard/project/${projectId}`);
	}

	function getActivityIcon(action: string) {
		if (action.includes('create')) return '✓';
		if (action.includes('delete')) return '✕';
		if (action.includes('update')) return '↻';
		if (action.includes('deploy')) return '🚀';
		return '•';
	}

	function getActivityColor(action: string) {
		if (action.includes('create')) return 'bg-green-500';
		if (action.includes('delete')) return 'bg-red-500';
		if (action.includes('update')) return 'bg-blue-500';
		if (action.includes('deploy')) return 'bg-purple-500';
		return 'bg-gray-500';
	}

	function formatTimeAgo(dateStr: string) {
		const date = new Date(dateStr);
		const now = new Date();
		const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

		if (seconds < 60) return `${seconds}s ago`;
		if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
		if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
		return `${Math.floor(seconds / 86400)}d ago`;
	}

	function getServerStatusColor(status: string) {
		switch (status) {
			case 'online':
				return 'bg-green-500';
			case 'offline':
				return 'bg-red-500';
			case 'maintenance':
				return 'bg-yellow-500';
			case 'error':
				return 'bg-red-600';
			default:
				return 'bg-gray-500';
		}
	}
</script>

<svelte:head>
	<title>Dashboard - mikrocloud</title>
</svelte:head>

<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
	<div class="lg:col-span-2 space-y-6">
		<div>
			<h2 class="text-lg font-semibold mb-4">Projects</h2>
			{#if loadingProjects}
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					{#each Array(4) as _, i}
						<div class="skeleton h-32 w-full rounded-lg" />
					{/each}
				</div>
			{:else if projects.length === 0}
				<div class="bg-white/5 border border-white/10 rounded-lg p-8 text-center">
					<p class="text-gray-400 text-sm mb-4">No projects yet</p>
					<button onclick={() => goto('/dashboard/projects/new')} class="btn">
						<Plus class="w-4 h-4 mr-2" />
						Create Project
					</button>
				</div>
			{:else}
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					{#each projects as project (project.id)}
						<div
							class="card card-border bg-base-100 cursor-pointer"
							onclick={() => goToProject(project.id)}
							role="button"
							tabindex="0"
							onkeydown={(e) => e.key === 'Enter' && goToProject(project.id)}
						>
							<div class="card-body">
								<div class="card-title">
									<div class="flex items-center space-x-2">
										<Globe class="w-4 h-4 text-gray-400" />
										<span class="text-sm">{project.name}</span>
									</div>
									<button class="text-gray-400 hover:text-white" onclick={(e) => e.stopPropagation()}>
										<Ellipsis class="w-5 h-5" />
									</button>
								</div>
								<div class="card-actions flex items-center justify-between">
									<span class="text-xs text-gray-500">
										Created {formatTimeAgo(project.created_at)}
									</span>
									{#if project.description}
										<span class="text-xs text-gray-400 truncate max-w-[150px]">
											{project.description}
										</span>
									{/if}
								</div>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<div>
			<h2 class="text-lg font-semibold mb-4">Servers</h2>
			<div class="bg-white/5 border border-white/10 rounded-lg p-4">
				{#if loadingServers}
					<div class="space-y-4">
						<div class="skeleton h-24 w-full" />
						<div class="skeleton h-24 w-full" />
					</div>
				{:else if servers.length === 0}
					<div class="text-center text-gray-400 text-sm py-4">No servers configured</div>
				{:else}
					{#each servers as server (server.id)}
						<div class="pb-4 mb-4 last:pb-0 last:mb-0 border-b border-white/10 last:border-0">
							<div class="flex items-center justify-between mb-2">
								<div class="flex items-center space-x-2">
									<span class="text-sm">🖥️</span>
									<span class="text-sm font-medium">{server.name}</span>
								</div>
								<div class="flex items-center space-x-2">
									<div class="w-2 h-2 rounded-full {getServerStatusColor(server.status)}"></div>
									<span class="text-xs text-gray-400">{server.status}</span>
								</div>
							</div>
							<div class="text-xs text-gray-500 mb-2">{server.hostname}</div>
							{#if server.tags && server.tags.length > 0}
								<div class="flex flex-wrap gap-1">
									{#each server.tags as tag}
										<span class="text-xs bg-white/5 border border-white/10 rounded px-2 py-0.5">
											{tag}
										</span>
									{/each}
								</div>
							{/if}
						</div>
					{/each}
				{/if}
			</div>
		</div>
	</div>

	<div>
		<h2 class="text-lg font-semibold mb-4">Activity</h2>
		<div class="bg-white/5 border border-white/10 rounded-lg p-4">
			{#if loadingActivities}
				<div class="space-y-6">
					{#each Array(5) as _, i}
						<div class="flex items-start space-x-3">
							<div class="skeleton h-8 w-8 rounded-full" />
							<div class="flex-1 space-y-2">
								<div class="skeleton h-4 w-32" />
								<div class="skeleton h-3 w-48" />
								<div class="skeleton h-5 w-24" />
							</div>
						</div>
					{/each}
				</div>
			{:else if activities.length === 0}
				<div class="text-center text-gray-400 text-sm py-8">No recent activity</div>
			{:else}
				<div class="space-y-6">
					{#each activities as activity, index (activity.id)}
						<div class="flex items-start space-x-3">
							<div class="relative">
								<div
									class="w-8 h-8 rounded-full {getActivityColor(
										activity.activity_type
									)} flex items-center justify-center text-xs flex-shrink-0"
								>
									{getActivityIcon(activity.activity_type)}
								</div>
								{#if index !== activities.length - 1}
									<div class="absolute top-8 left-4 w-px h-8 bg-white/10"></div>
								{/if}
							</div>
							<div class="flex-1 min-w-0">
								<div class="flex items-center justify-between mb-1">
									<span class="text-sm font-medium">{activity.initiator_name || 'System'}</span>
									<span class="text-xs text-gray-500">{formatTimeAgo(activity.created_at)}</span>
								</div>
								<div class="text-xs text-gray-400 mb-1">{activity.activity_type}</div>
								{#if activity.resource_type && activity.resource_id}
									<div
										class="text-xs bg-white/5 border border-white/10 rounded px-2 py-1 inline-block"
									>
										{activity.resource_type}: {activity.resource_id.substring(0, 8)}
									</div>
								{/if}
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	</div>
</div>
