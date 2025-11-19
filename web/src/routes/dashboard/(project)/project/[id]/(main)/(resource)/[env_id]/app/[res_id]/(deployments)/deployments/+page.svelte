<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Card, CardContent } from '$lib/components/ui/card';
	import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { deploymentsApi, type DeploymentStatus } from '$lib/api/deployments';
	import { toast } from 'svelte-sonner';
	import {
		CheckCircle,
		AlertCircle,
		XCircle,
		Clock,
		GitBranch,
		GitCommit,
		User,
		RotateCcw,
		Eye
	} from 'lucide-svelte';

	const projectId = $derived(page.params.id);
	const envId = $derived(page.params.env_id);
	const resId = $derived(page.params.res_id);

	const queryClient = useQueryClient();

	const deploymentsQuery = createQuery(() => ({
		queryKey: ['deployments', projectId, resId],
		queryFn: () => deploymentsApi.list(projectId, resId),
		enabled: !!projectId && !!resId,
		refetchInterval: 5000
	}));

	const deployments = $derived(deploymentsQuery.data || []);

	const redeployMutation = createMutation(() => ({
		mutationFn: (deploymentId: string) => deploymentsApi.redeploy(projectId, resId, deploymentId),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['deployments', projectId, resId] });
			toast.success('Redeployment started');
		},
		onError: (error: Error) => {
			toast.error(`Failed to redeploy: ${error.message}`);
		}
	}));

	function getStatusIcon(status: DeploymentStatus) {
		switch (status) {
			case 'success':
				return CheckCircle;
			case 'failed':
				return AlertCircle;
			case 'cancelled':
				return XCircle;
			case 'building':
			case 'deploying':
			case 'pending':
				return Clock;
			default:
				return Clock;
		}
	}

	function getStatusColor(status: DeploymentStatus) {
		switch (status) {
			case 'success':
				return 'text-green-500';
			case 'failed':
				return 'text-red-500';
			case 'cancelled':
				return 'text-gray-500';
			case 'building':
			case 'deploying':
			case 'pending':
				return 'text-blue-500';
			default:
				return 'text-gray-500';
		}
	}

	function getStatusBadgeVariant(
		status: DeploymentStatus
	): 'default' | 'secondary' | 'destructive' | 'outline' {
		switch (status) {
			case 'success':
				return 'default';
			case 'failed':
				return 'destructive';
			case 'cancelled':
				return 'secondary';
			case 'building':
			case 'deploying':
			case 'pending':
				return 'outline';
			default:
				return 'secondary';
		}
	}

	function formatDuration(seconds?: number): string {
		if (!seconds) return 'N/A';
		const mins = Math.floor(seconds / 60);
		const secs = seconds % 60;
		return `${mins}m ${secs}s`;
	}

	function formatTime(timestamp: string): string {
		const date = new Date(timestamp);
		const now = new Date();
		const diff = now.getTime() - date.getTime();
		const seconds = Math.floor(diff / 1000);
		const minutes = Math.floor(seconds / 60);
		const hours = Math.floor(minutes / 60);
		const days = Math.floor(hours / 24);

		if (days > 0) return `${days} day${days > 1 ? 's' : ''} ago`;
		if (hours > 0) return `${hours} hour${hours > 1 ? 's' : ''} ago`;
		if (minutes > 0) return `${minutes} minute${minutes > 1 ? 's' : ''} ago`;
		return `${seconds} second${seconds > 1 ? 's' : ''} ago`;
	}

	function viewDeployment(deploymentId: string) {
		goto(`/dashboard/project/${projectId}/${envId}/app/${resId}/deployments/${deploymentId}`);
	}

	function redeployCommit(deploymentId: string) {
		redeployMutation.mutate(deploymentId);
	}
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h2 class="text-2xl font-bold tracking-tight">Deployments</h2>
			<p class="text-muted-foreground">View and manage all deployments for this application</p>
		</div>
	</div>

	{#if deploymentsQuery.isLoading}
		<div class="flex items-center justify-center py-12">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
		</div>
	{:else if deploymentsQuery.isError}
		<Card>
			<CardContent class="p-6">
				<div class="text-center">
					<AlertCircle class="mx-auto h-12 w-12 text-destructive mb-4" />
					<h3 class="text-lg font-medium">Failed to load deployments</h3>
					<p class="text-muted-foreground mt-2">
						{deploymentsQuery.error?.message || 'An error occurred'}
					</p>
				</div>
			</CardContent>
		</Card>
	{:else if deployments.length === 0}
		<Card>
			<CardContent class="p-12">
				<div class="text-center">
					<GitBranch class="mx-auto h-12 w-12 text-muted-foreground mb-4" />
					<h3 class="text-lg font-medium">No deployments yet</h3>
					<p class="text-muted-foreground mt-2">
						Deployments will appear here once you trigger your first deployment.
					</p>
				</div>
			</CardContent>
		</Card>
	{:else}
		<div class="space-y-4">
			{#each deployments as deployment (deployment.id)}
				{@const StatusIcon = getStatusIcon(deployment.status)}
				<Card class="hover:shadow-sm transition-shadow">
					<CardContent class="p-6">
						<div class="flex items-center justify-between">
							<div class="flex items-center space-x-4 flex-1">
								<StatusIcon class="w-5 h-5 {getStatusColor(deployment.status)}" />

								<div class="flex-1">
									<div class="flex items-center space-x-3 mb-1">
										<h3 class="font-medium">
											{deployment.commit_message || 'Deployment'}
										</h3>
										<Badge variant={getStatusBadgeVariant(deployment.status)} class="text-xs">
											{deployment.status}
										</Badge>
									</div>

									<div
										class="flex items-center space-x-4 text-sm text-muted-foreground flex-wrap gap-2"
									>
										{#if deployment.commit_hash}
											<div class="flex items-center space-x-1">
												<GitCommit class="w-4 h-4" />
												<span class="font-mono">{deployment.commit_hash.slice(0, 7)}</span>
											</div>
										{/if}
										{#if deployment.branch}
											<div class="flex items-center space-x-1">
												<GitBranch class="w-4 h-4" />
												<span>{deployment.branch}</span>
											</div>
										{/if}
										{#if deployment.author}
											<div class="flex items-center space-x-1">
												<User class="w-4 h-4" />
												<span>{deployment.author}</span>
											</div>
										{/if}
										{#if deployment.duration}
											<div class="flex items-center space-x-1">
												<Clock class="w-4 h-4" />
												<span>{formatDuration(deployment.duration)}</span>
											</div>
										{/if}
									</div>
								</div>
							</div>

							<div class="flex items-center space-x-4 ml-4">
								<span class="text-sm text-muted-foreground whitespace-nowrap">
									{formatTime(deployment.started_at)}
								</span>
								<div class="flex items-center space-x-2">
									<Button size="sm" variant="outline" onclick={() => viewDeployment(deployment.id)}>
										<Eye class="w-4 h-4 mr-1" />
										View
									</Button>
									{#if deployment.status === 'success'}
										<Button
											size="sm"
											variant="outline"
											disabled={redeployMutation.isPending}
											onclick={() => redeployCommit(deployment.id)}
										>
											<RotateCcw
												class="w-4 h-4 mr-1 {redeployMutation.isPending ? 'animate-spin' : ''}"
											/>
											Redeploy
										</Button>
									{/if}
								</div>
							</div>
						</div>
					</CardContent>
				</Card>
			{/each}
		</div>
	{/if}
</div>
