<script lang="ts">
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Separator } from '$lib/components/ui/separator';
	import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { applicationsApi } from '$lib/api/applications';
	import { deploymentsApi } from '$lib/api/deployments';
	import { toast } from 'svelte-sonner';
	import {
		Play,
		Square,
		RefreshCw,
		ExternalLink,
		GitBranch,
		GitCommit,
		User,
		Calendar,
		FileText,
		Eye,
		RotateCcw
	} from 'lucide-svelte';

	const projectId = $derived(page.params.id);
	const resId = $derived(page.params.res_id);
	const envId = $derived(page.params.env_id);

	const queryClient = useQueryClient();

	const applicationQuery = createQuery(() => ({
		queryKey: ['application', projectId, resId],
		queryFn: () => applicationsApi.get(projectId, resId),
		enabled: !!projectId && !!resId
	}));

	const deploymentsQuery = createQuery(() => ({
		queryKey: ['deployments', projectId, resId],
		queryFn: () => deploymentsApi.list(projectId, resId),
		enabled: !!projectId && !!resId,
		refetchInterval: 5000
	}));

	const application = $derived(applicationQuery.data);
	const latestDeployment = $derived(deploymentsQuery.data?.[0]);

	const startMutation = createMutation(() => ({
		mutationFn: () => applicationsApi.start(projectId, resId),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['application', projectId, resId] });
			toast.success('Application started successfully');
		},
		onError: (error: Error) => {
			toast.error(`Failed to start application: ${error.message}`);
		}
	}));

	const stopMutation = createMutation(() => ({
		mutationFn: () => applicationsApi.stop(projectId, resId),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['application', projectId, resId] });
			toast.success('Application stopped successfully');
		},
		onError: (error: Error) => {
			toast.error(`Failed to stop application: ${error.message}`);
		}
	}));

	const restartMutation = createMutation(() => ({
		mutationFn: () => applicationsApi.restart(projectId, resId),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['application', projectId, resId] });
			toast.success('Application restarted successfully');
		},
		onError: (error: Error) => {
			toast.error(`Failed to restart application: ${error.message}`);
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
			case 'success':
				return 'default';
			case 'stopped':
				return 'secondary';
			case 'failed':
				return 'destructive';
			case 'building':
			case 'deploying':
				return 'outline';
			default:
				return 'outline';
		}
	}

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function formatRelativeTime(timestamp: string): string {
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
</script>

{#if application}
	<div class="space-y-6">
		<Card>
			<CardHeader>
				<div class="flex items-center justify-between">
					<div class="flex items-center space-x-3">
						<CardTitle>Latest Deployment</CardTitle>
						{#if latestDeployment}
							<Badge variant={getStatusBadgeVariant(latestDeployment.status)}>
								{latestDeployment.status}
							</Badge>
						{/if}
					</div>
					<div class="flex gap-2">
						{#if application.status === 'stopped' || application.status === 'pending' || application.status === 'created'}
							<Button
								size="sm"
								disabled={isAnyActionPending}
								onclick={() => startMutation.mutate()}
							>
								<Play class="mr-2 h-4 w-4" />
								{startMutation.isPending ? 'Starting...' : 'Start'}
							</Button>
						{:else if application.status === 'running'}
							<Button
								size="sm"
								variant="outline"
								disabled={isAnyActionPending}
								onclick={() => stopMutation.mutate()}
							>
								<Square class="mr-2 h-4 w-4" />
								{stopMutation.isPending ? 'Stopping...' : 'Stop'}
							</Button>
						{/if}
						{#if application.status !== 'pending'}
							<Button
								size="sm"
								variant="outline"
								disabled={isAnyActionPending}
								onclick={() => restartMutation.mutate()}
							>
								<RefreshCw class="mr-2 h-4 w-4 {restartMutation.isPending ? 'animate-spin' : ''}" />
								{restartMutation.isPending ? 'Restarting...' : 'Restart'}
							</Button>
						{/if}
					</div>
				</div>
			</CardHeader>
			<CardContent>
				{#if latestDeployment}
					<div class="space-y-4">
						<div class="grid grid-cols-2 gap-6">
							<div>
								<div class="flex items-center gap-2 text-sm text-muted-foreground mb-1">
									<Calendar class="h-4 w-4" />
									<span>Deployed</span>
								</div>
								<p class="font-medium">
									{formatDate(latestDeployment.started_at)}
								</p>
								<p class="text-sm text-muted-foreground">
									{formatRelativeTime(latestDeployment.started_at)}
								</p>
							</div>
							{#if latestDeployment.commit_hash}
								<div>
									<div class="flex items-center gap-2 text-sm text-muted-foreground mb-1">
										<GitCommit class="h-4 w-4" />
										<span>Commit</span>
									</div>
									<p class="font-mono font-medium">{latestDeployment.commit_hash.slice(0, 7)}</p>
								</div>
							{/if}
						</div>

						<Separator />

						{#if latestDeployment.branch}
							<div>
								<div class="flex items-center gap-2 text-sm text-muted-foreground mb-1">
									<GitBranch class="h-4 w-4" />
									<span>Branch</span>
								</div>
								<p class="font-medium">{latestDeployment.branch}</p>
							</div>
						{/if}

						{#if latestDeployment.commit_message}
							<div>
								<div class="flex items-center gap-2 text-sm text-muted-foreground mb-1">
									<FileText class="h-4 w-4" />
									<span>Commit Message</span>
								</div>
								<p class="font-medium">{latestDeployment.commit_message}</p>
							</div>
						{/if}

						{#if latestDeployment.author}
							<div>
								<div class="flex items-center gap-2 text-sm text-muted-foreground mb-1">
									<User class="h-4 w-4" />
									<span>Author</span>
								</div>
								<p class="font-medium">{latestDeployment.author}</p>
							</div>
						{/if}

						<Separator />

						<div class="flex gap-2">
							<Button
								size="sm"
								variant="outline"
								href="/dashboard/project/{projectId}/{envId}/app/{resId}/deployments/{latestDeployment.id}"
							>
								<Eye class="mr-2 h-4 w-4" />
								View Logs
							</Button>
							{#if latestDeployment.status === 'success'}
								<Button size="sm" variant="outline">
									<RotateCcw class="mr-2 h-4 w-4" />
									Instant Rollback
								</Button>
							{/if}
						</div>
					</div>
				{:else}
					<div class="text-center py-8 text-muted-foreground">
						<p>No deployments yet</p>
					</div>
				{/if}
			</CardContent>
		</Card>

		{#if application.domain}
			<Card>
				<CardHeader>
					<CardTitle>Domain</CardTitle>
				</CardHeader>
				<CardContent>
					<div class="flex items-center justify-between">
						<div>
							<p class="font-medium mb-1">{application.domain}</p>
							<p class="text-sm text-muted-foreground">Application is accessible at this domain</p>
						</div>
						<Button size="sm" variant="outline" href="https://{application.domain}" target="_blank">
							<ExternalLink class="h-4 w-4" />
						</Button>
					</div>
				</CardContent>
			</Card>
		{/if}

		<Card>
			<CardHeader>
				<CardTitle>Source</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="space-y-4">
					{#if application.deployment_source}
						{@const source = application.deployment_source}
						{#if source.source_type === 'git'}
							<div>
								<div class="flex items-center gap-2 text-sm text-muted-foreground mb-1">
									<GitBranch class="h-4 w-4" />
									<span>Repository</span>
								</div>
								<p class="font-medium font-mono text-sm break-all">
									{source.git_url || 'Not configured'}
								</p>
							</div>
							{#if source.git_branch}
								<div>
									<div class="flex items-center gap-2 text-sm text-muted-foreground mb-1">
										<GitBranch class="h-4 w-4" />
										<span>Branch</span>
									</div>
									<p class="font-medium">{source.git_branch}</p>
								</div>
							{/if}
							{#if source.git_path}
								<div>
									<div class="flex items-center gap-2 text-sm text-muted-foreground mb-1">
										<FileText class="h-4 w-4" />
										<span>Path</span>
									</div>
									<p class="font-medium font-mono text-sm">{source.git_path}</p>
								</div>
							{/if}
						{:else if source.source_type === 'docker'}
							<div>
								<div class="flex items-center gap-2 text-sm text-muted-foreground mb-1">
									<span>Docker Image</span>
								</div>
								<p class="font-medium font-mono text-sm">
									{source.docker_image || 'Not configured'}
								</p>
							</div>
						{/if}
					{:else}
						<p class="text-muted-foreground">No source configured</p>
					{/if}
				</div>
			</CardContent>
		</Card>
	</div>
{/if}
