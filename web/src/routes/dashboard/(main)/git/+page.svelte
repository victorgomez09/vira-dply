<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import {
		Sheet,
		SheetContent,
		SheetDescription,
		SheetHeader,
		SheetTitle
	} from '$lib/components/ui/sheet';
	import { Select, SelectContent, SelectItem, SelectTrigger } from '$lib/components/ui/select';
	import { Plus, Trash2, Github, GitBranch, Server } from 'lucide-svelte';
	import { gitApi, type GitProvider, type GitSource } from '$lib/api/git';
	import { onMount } from 'svelte';

	let sources = $state<GitSource[]>([]);
	let isLoading = $state(true);
	let error = $state<string | null>(null);
	let isCreateSheetOpen = $state(false);
	let isDeleteModalOpen = $state(false);
	let selectedSource = $state<GitSource | null>(null);

	let formData = $state({
		name: '',
		provider: 'github' as GitProvider,
		access_token: '',
		refresh_token: '',
		custom_url: ''
	});

	onMount(async () => {
		await loadSources();
	});

	async function loadSources() {
		isLoading = true;
		error = null;
		try {
			sources = await gitApi.listGitSources();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load Git sources';
			console.error('Failed to load Git sources:', e);
		} finally {
			isLoading = false;
		}
	}

	async function createSource() {
		try {
			await gitApi.createGitSource({
				name: formData.name,
				provider: formData.provider,
				access_token: formData.access_token,
				refresh_token: formData.refresh_token || undefined,
				custom_url: formData.provider === 'custom' ? formData.custom_url : undefined
			});
			isCreateSheetOpen = false;
			resetForm();
			await loadSources();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to create Git source';
		}
	}

	async function deleteSource() {
		if (!selectedSource) return;
		try {
			await gitApi.deleteGitSource(selectedSource.id);
			isDeleteModalOpen = false;
			selectedSource = null;
			await loadSources();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to delete Git source';
		}
	}

	function resetForm() {
		formData = {
			name: '',
			provider: 'github',
			access_token: '',
			refresh_token: '',
			custom_url: ''
		};
	}

	function getProviderIcon(provider: GitProvider) {
		switch (provider) {
			case 'github':
				return Github;
			case 'gitlab':
			case 'bitbucket':
				return GitBranch;
			case 'custom':
				return Server;
			default:
				return GitBranch;
		}
	}

	function getProviderBadgeColor(provider: GitProvider) {
		switch (provider) {
			case 'github':
				return 'bg-gray-900 text-white';
			case 'gitlab':
				return 'bg-orange-600 text-white';
			case 'bitbucket':
				return 'bg-blue-600 text-white';
			case 'custom':
				return 'bg-purple-600 text-white';
			default:
				return 'bg-gray-500 text-white';
		}
	}

	function formatDate(dateStr: string) {
		return new Date(dateStr).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}
</script>

<svelte:head>
	<title>Git Sources - Dashboard</title>
</svelte:head>

<div class="flex-1 flex flex-col overflow-hidden">
	<div class="bg-white border-b border-gray-200 px-6 py-4">
		<div class="flex items-center justify-between">
			<div>
				<h1 class="text-2xl font-semibold text-gray-900">Git Sources</h1>
				<p class="text-sm text-gray-500 mt-1">Manage Git provider connections and access tokens</p>
			</div>
			<Button onclick={() => (isCreateSheetOpen = true)}>
				<Plus class="h-4 w-4 mr-2" />
				Add Source
			</Button>
		</div>
	</div>

	<Sheet bind:open={isCreateSheetOpen}>
		<SheetContent class="overflow-y-auto">
			<SheetHeader>
				<SheetTitle>Add Git Source</SheetTitle>
				<SheetDescription>Connect a Git provider by providing an access token</SheetDescription>
			</SheetHeader>
			<div class="space-y-4 py-4">
				<div class="space-y-2">
					<Label for="name">Name</Label>
					<Input id="name" placeholder="My GitHub Account" bind:value={formData.name} />
				</div>

				<div class="space-y-2">
					<Label for="provider">Provider</Label>
					<Select
						value={formData.provider}
						onValueChange={(v) => v && (formData.provider = v as GitProvider)}
					>
						<SelectTrigger id="provider">
							{formData.provider.charAt(0).toUpperCase() + formData.provider.slice(1)}
						</SelectTrigger>
						<SelectContent>
							<SelectItem value="github">GitHub</SelectItem>
							<SelectItem value="gitlab">GitLab</SelectItem>
							<SelectItem value="bitbucket">Bitbucket</SelectItem>
							<SelectItem value="custom">Custom/Self-hosted</SelectItem>
						</SelectContent>
					</Select>
				</div>

				{#if formData.provider === 'custom'}
					<div class="space-y-2">
						<Label for="custom_url">Custom Git URL</Label>
						<Input
							id="custom_url"
							placeholder="https://git.example.com"
							bind:value={formData.custom_url}
						/>
						<p class="text-xs text-muted-foreground">Base URL of your self-hosted Git instance</p>
					</div>
				{/if}

				<div class="space-y-2">
					<Label for="access_token">Access Token</Label>
					<Input
						id="access_token"
						type="password"
						placeholder="ghp_xxxxxxxxxxxx"
						bind:value={formData.access_token}
					/>
					<p class="text-xs text-muted-foreground">
						{#if formData.provider === 'github'}
							Create a personal access token at GitHub Settings → Developer settings
						{:else if formData.provider === 'gitlab'}
							Create an access token at GitLab Settings → Access Tokens
						{:else if formData.provider === 'bitbucket'}
							Create an app password at Bitbucket Settings → App passwords
						{:else}
							Create an access token in your Git provider's settings
						{/if}
					</p>
				</div>

				<div class="space-y-2">
					<Label for="refresh_token">Refresh Token (Optional)</Label>
					<Input
						id="refresh_token"
						type="password"
						placeholder="Optional refresh token"
						bind:value={formData.refresh_token}
					/>
					<p class="text-xs text-muted-foreground">If your provider supports token refresh</p>
				</div>
			</div>
			<div class="flex justify-end gap-2">
				<Button variant="outline" onclick={() => (isCreateSheetOpen = false)}>Cancel</Button>
				<Button onclick={createSource} disabled={!formData.name || !formData.access_token}>
					Add Source
				</Button>
			</div>
		</SheetContent>
	</Sheet>

	<div class="flex-1 overflow-y-auto p-6">
		{#if error}
			<Card class="border-red-200 bg-red-50">
				<CardContent class="pt-6">
					<p class="text-red-800">{error}</p>
				</CardContent>
			</Card>
		{/if}

		{#if isLoading}
			<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
				{#each [1, 2, 3] as i}
					<Card>
						<CardHeader>
							<div class="h-4 bg-gray-200 rounded animate-pulse w-1/2"></div>
							<div class="h-3 bg-gray-200 rounded animate-pulse w-1/3 mt-2"></div>
						</CardHeader>
						<CardContent>
							<div class="h-3 bg-gray-200 rounded animate-pulse w-full"></div>
						</CardContent>
					</Card>
				{/each}
			</div>
		{:else if sources.length === 0}
			<Card>
				<CardContent class="pt-6 text-center py-12">
					<GitBranch class="h-12 w-12 mx-auto text-gray-400 mb-4" />
					<h3 class="text-lg font-medium text-gray-900 mb-2">No Git sources yet</h3>
					<p class="text-gray-500 mb-4">
						Connect your Git providers to deploy applications from repositories
					</p>
					<Button onclick={() => (isCreateSheetOpen = true)}>
						<Plus class="h-4 w-4 mr-2" />
						Add Your First Source
					</Button>
				</CardContent>
			</Card>
		{:else}
			<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
				{#each sources as source (source.id)}
					{@const Icon = getProviderIcon(source.provider)}
					<Card>
						<CardHeader>
							<div class="flex items-start justify-between">
								<div class="flex items-center gap-3">
									<div class="p-2 rounded-lg {getProviderBadgeColor(source.provider)}">
										<Icon class="h-5 w-5" />
									</div>
									<div>
										<CardTitle class="text-lg">{source.name}</CardTitle>
										<Badge variant="secondary" class="mt-1">
											{source.provider}
										</Badge>
									</div>
								</div>
								<Button
									variant="ghost"
									size="sm"
									onclick={() => {
										selectedSource = source;
										isDeleteModalOpen = true;
									}}
								>
									<Trash2 class="h-4 w-4 text-red-600" />
								</Button>
							</div>
						</CardHeader>
						<CardContent>
							{#if source.custom_url}
								<p class="text-sm text-gray-600 mb-2">
									<span class="font-medium">URL:</span>
									{source.custom_url}
								</p>
							{/if}
							<p class="text-xs text-gray-500">
								Created {formatDate(source.created_at)}
							</p>
						</CardContent>
					</Card>
				{/each}
			</div>
		{/if}
	</div>
</div>

{#if isDeleteModalOpen && selectedSource}
	<div
		class="fixed inset-0 z-50 bg-black/80 flex items-center justify-center"
		onclick={() => (isDeleteModalOpen = false)}
	>
		<div
			class="bg-white rounded-lg shadow-lg p-6 max-w-md w-full mx-4"
			onclick={(e) => e.stopPropagation()}
		>
			<h2 class="text-xl font-semibold mb-2">Delete Git Source</h2>
			<p class="text-gray-600 mb-6">
				Are you sure you want to delete "{selectedSource.name}"? This action cannot be undone.
			</p>
			<div class="flex justify-end gap-2">
				<Button variant="outline" onclick={() => (isDeleteModalOpen = false)}>Cancel</Button>
				<Button variant="destructive" onclick={deleteSource}>Delete</Button>
			</div>
		</div>
	</div>
{/if}
