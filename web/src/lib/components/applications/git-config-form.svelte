<script lang="ts">
	import { Label } from '$lib/components/ui/label';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Tabs, TabsList, TabsTrigger } from '$lib/components/ui/tabs';
	import { RefreshCw, Check, AlertCircle } from 'lucide-svelte';
	import { Select, SelectContent, SelectItem, SelectTrigger } from '$lib/components/ui/select';
	import { gitApi, type GitProvider } from '$lib/api/git';

	interface Props {
		provider: 'github' | 'gitlab' | 'bitbucket' | 'custom';
		onProviderChange: (provider: 'github' | 'gitlab' | 'bitbucket' | 'custom') => void;
		repository: string;
		onRepositoryChange: (repository: string) => void;
		branch: string;
		onBranchChange: (branch: string) => void;
		autoDeploy: boolean;
		onAutoDeployChange: (autoDeploy: boolean) => void;
		isPrivate: boolean;
		onIsPrivateChange: (isPrivate: boolean) => void;
		customGitUrl?: string;
		onCustomGitUrlChange?: (url: string) => void;
		basePath?: string;
		onBasePathChange?: (basePath: string) => void;
	}

	let {
		provider,
		onProviderChange,
		repository,
		onRepositoryChange,
		branch,
		onBranchChange,
		autoDeploy,
		onAutoDeployChange,
		isPrivate,
		onIsPrivateChange,
		customGitUrl = '',
		onCustomGitUrlChange,
		basePath = '/',
		onBasePathChange
	}: Props = $props();

	let branches = $state<string[]>(['main', 'master', 'develop', 'staging', 'production']);
	let isValidating = $state(false);
	let validationStatus = $state<'idle' | 'valid' | 'invalid'>('idle');
	let validationMessage = $state('');
	let isFetchingBranches = $state(false);

	async function validateRepository() {
		if (!repository) return;

		isValidating = true;
		validationStatus = 'idle';

		try {
			const result = await gitApi.validateRepository({
				provider: provider as GitProvider,
				repository: repository,
				branch: branch || undefined,
				custom_url: provider === 'custom' ? customGitUrl : undefined
			});

			if (result.valid) {
				validationStatus = 'valid';
				validationMessage = result.message || 'Repository is accessible';
			} else {
				validationStatus = 'invalid';
				validationMessage = result.message || 'Unable to access repository';
			}
		} catch (error) {
			validationStatus = 'invalid';
			validationMessage = error instanceof Error ? error.message : 'Unable to access repository';
		} finally {
			isValidating = false;
		}
	}

	async function fetchBranches() {
		if (!repository) return;

		isFetchingBranches = true;
		try {
			const result = await gitApi.listBranches({
				provider: provider as GitProvider,
				repository: repository,
				custom_url: provider === 'custom' ? customGitUrl : undefined
			});

			branches = result.branches.map((b) => b.name);
			if (branches.length > 0 && !branch) {
				onBranchChange(branches[0]);
			}
		} catch (error) {
			console.error('Failed to fetch branches:', error);
			branches = ['main'];
		} finally {
			isFetchingBranches = false;
		}
	}
</script>

<div class="space-y-6">
	<div class="flex gap-4">
		<Button
			variant={!isPrivate ? 'default' : 'outline'}
			onclick={() => onIsPrivateChange(false)}
			class="flex-1"
		>
			Public repository
		</Button>
		<Button
			variant={isPrivate ? 'default' : 'outline'}
			onclick={() => onIsPrivateChange(true)}
			class="flex-1"
		>
			Private repository
		</Button>
	</div>

	<Tabs
		value={provider}
		onValueChange={(v: string | undefined) =>
			v && onProviderChange(v as 'github' | 'gitlab' | 'bitbucket' | 'custom')}
	>
		<TabsList class="grid w-full grid-cols-4">
			<TabsTrigger value="github">GitHub</TabsTrigger>
			<TabsTrigger value="gitlab">GitLab</TabsTrigger>
			<TabsTrigger value="bitbucket">Bitbucket</TabsTrigger>
			<TabsTrigger value="custom">Custom</TabsTrigger>
		</TabsList>
	</Tabs>

	{#if provider === 'custom' && onCustomGitUrlChange}
		<div class="space-y-2">
			<Label for="custom-git-url">Git URL</Label>
			<Input
				id="custom-git-url"
				placeholder="https://git.example.com or git@example.com:repo.git"
				value={customGitUrl}
				oninput={(e) => onCustomGitUrlChange(e.currentTarget.value)}
			/>
			<p class="text-xs text-muted-foreground">Enter the full Git URL (HTTPS or SSH)</p>
		</div>
	{/if}

	<div class="space-y-2">
		<Label for="repository">Repository</Label>
		<div class="flex gap-2">
			<div class="flex-1 relative">
				<Input
					id="repository"
					placeholder={provider === 'custom'
						? 'Leave empty if URL contains path'
						: 'username/repo-name'}
					value={repository}
					oninput={(e) => {
						onRepositoryChange(e.currentTarget.value);
						validationStatus = 'idle';
					}}
					class="pr-8"
				/>
				{#if validationStatus === 'valid'}
					<Check class="h-4 w-4 text-green-600 absolute right-2 top-1/2 -translate-y-1/2" />
				{:else if validationStatus === 'invalid'}
					<AlertCircle class="h-4 w-4 text-destructive absolute right-2 top-1/2 -translate-y-1/2" />
				{/if}
			</div>
			<Button
				variant="outline"
				size="icon"
				onclick={validateRepository}
				disabled={isValidating || !repository}
			>
				<RefreshCw class="h-4 w-4 {isValidating ? 'animate-spin' : ''}" />
			</Button>
		</div>
		{#if validationMessage}
			<p
				class="text-xs"
				class:text-green-600={validationStatus === 'valid'}
				class:text-destructive={validationStatus === 'invalid'}
			>
				{validationMessage}
			</p>
		{:else}
			<p class="text-xs text-muted-foreground">
				{provider === 'custom'
					? 'Optional: specify repository path if not in URL'
					: 'Enter the repository path (e.g., username/repository-name)'}
			</p>
		{/if}
	</div>

	<div class="space-y-2">
		<div class="flex items-center justify-between">
			<Label for="branch">Default branch</Label>
			<Button
				variant="ghost"
				size="sm"
				onclick={fetchBranches}
				disabled={isFetchingBranches || !repository}
				class="h-7 text-xs"
			>
				<RefreshCw class="h-3 w-3 mr-1 {isFetchingBranches ? 'animate-spin' : ''}" />
				Fetch branches
			</Button>
		</div>
		<Select value={branch} onValueChange={(v) => v && onBranchChange(v)}>
			<SelectTrigger id="branch">
				{branch || 'Select a branch'}
			</SelectTrigger>
			<SelectContent>
				{#each branches as branchOption}
					<SelectItem value={branchOption}>{branchOption}</SelectItem>
				{/each}
			</SelectContent>
		</Select>
	</div>

	{#if onBasePathChange}
		<div class="space-y-2">
			<Label for="base-path">Base path</Label>
			<Input
				id="base-path"
				placeholder="/"
				value={basePath}
				oninput={(e) => onBasePathChange(e.currentTarget.value)}
			/>
			<p class="text-xs text-muted-foreground">
				The directory containing your Dockerfile or docker-compose.yml (default: /)
			</p>
		</div>
	{/if}

	<div class="flex items-center space-x-2">
		<input
			type="checkbox"
			id="auto-deploy"
			checked={autoDeploy}
			onchange={(e) => onAutoDeployChange(e.currentTarget.checked)}
			class="h-4 w-4 rounded border-input"
		/>
		<Label for="auto-deploy" class="font-normal cursor-pointer">
			Automatic deployment on commit
		</Label>
	</div>
</div>
