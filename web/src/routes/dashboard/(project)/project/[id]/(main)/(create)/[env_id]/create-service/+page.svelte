<script lang="ts">
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import {
		Select,
		SelectContent,
		SelectItem,
		SelectTrigger
	} from '$lib/components/ui/select/index';
	import {
		Sheet,
		SheetClose,
		SheetContent,
		SheetDescription,
		SheetFooter,
		SheetHeader,
		SheetTitle,
		SheetTrigger
	} from '$lib/components/ui/sheet';
	import { Badge } from '$lib/components/ui/badge';
	import { createQuery, createMutation } from '@tanstack/svelte-query';
	import {
		templatesApi,
		type ServiceTemplate,
		type DeployTemplateRequest
	} from '$lib/api/templates';

	import { ArrowLeft, Search, Loader2, X } from 'lucide-svelte';
	import { goto } from '$app/navigation';

	const projectId = $derived(page.params.id);
	const envId = $derived(page.params.env_id);

	let searchQuery = $state('');
	let selectedCategory = $state<string>('all');
	let deploySheetOpen = $state(false);
	let selectedTemplate = $state<ServiceTemplate | null>(null);
	let deploymentName = $state('');
	let customEnvVars = $state<Record<string, string>>({});

	const templatesQuery = createQuery(() => ({
		queryKey: ['templates', selectedCategory],
		queryFn: () => templatesApi.list(selectedCategory === 'all' ? undefined : selectedCategory)
	}));

	const deployMutation = createMutation(() => ({
		mutationFn: (request: { templateId: string; data: DeployTemplateRequest }) =>
			templatesApi.deploy(request.templateId, request.data),
		onSuccess: () => {
			deploySheetOpen = false;
			goto(`/dashboard/project/${projectId}`);
		}
	}));

	const filteredTemplates = $derived(
		templatesQuery.data?.filter((template: ServiceTemplate) =>
			template.name.toLowerCase().includes(searchQuery.toLowerCase())
		) || []
	);

	const categories = [
		{ value: 'all', label: 'All Templates' },
		{ value: 'database', label: 'Database' },
		{ value: 'webapp', label: 'Web App' },
		{ value: 'api', label: 'API' },
		{ value: 'worker', label: 'Worker' },
		{ value: 'cache', label: 'Cache' },
		{ value: 'analytics', label: 'Analytics' },
		{ value: 'monitoring', label: 'Monitoring' },
		{ value: 'other', label: 'Other' }
	];

	function handleBack() {
		goto(`/dashboard/project/${projectId}`);
	}

	function handleDeploy(template: ServiceTemplate) {
		selectedTemplate = template;
		deploymentName = template.name.toLowerCase().replace(/\s+/g, '-');
		customEnvVars = { ...template.environment };
		deploySheetOpen = true;
	}

	function handleAddEnvVar() {
		customEnvVars[`NEW_VAR_${Object.keys(customEnvVars).length + 1}`] = '';
	}

	function handleRemoveEnvVar(key: string) {
		const { [key]: _, ...rest } = customEnvVars;
		customEnvVars = rest;
	}

	function handleSubmitDeployment() {
		if (!selectedTemplate || !deploymentName) {
			return;
		}

		deployMutation.mutate({
			templateId: selectedTemplate.id,
			data: {
				name: deploymentName,
				project_id: projectId,
				environment_id: envId,
				environment: customEnvVars
			}
		});
	}
</script>

<div class="container max-w-7xl py-8">
	<div class="mb-6">
		<Button variant="ghost" onclick={handleBack} class="mb-4">
			<ArrowLeft class="mr-2 h-4 w-4" />
			Back
		</Button>
		<h1 class="text-3xl font-bold">Service Templates</h1>
		<p class="text-muted-foreground">Deploy services from pre-configured templates</p>
	</div>

	<div class="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div class="relative flex-1 max-w-md">
			<Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
			<Input
				type="search"
				placeholder="Search templates..."
				bind:value={searchQuery}
				class="pl-10"
			/>
		</div>

		<Select bind:value={selectedCategory}>
			<SelectTrigger class="w-[200px]">
				{categories.find((c) => c.value === selectedCategory)?.label || 'Select category'}
			</SelectTrigger>
			<SelectContent>
				{#each categories as category}
					<SelectItem value={category.value}>{category.label}</SelectItem>
				{/each}
			</SelectContent>
		</Select>
	</div>

	{#if templatesQuery.isLoading}
		<div class="flex items-center justify-center py-12">
			<Loader2 class="h-8 w-8 animate-spin text-muted-foreground" />
		</div>
	{:else if templatesQuery.isError}
		<Card class="border-destructive">
			<CardContent class="py-6">
				<p class="text-destructive">Failed to load templates. Please try again.</p>
			</CardContent>
		</Card>
	{:else if filteredTemplates.length === 0}
		<Card>
			<CardContent class="py-12 text-center">
				<p class="text-muted-foreground">No templates found</p>
			</CardContent>
		</Card>
	{:else}
		<div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
			{#each filteredTemplates as template}
				<Card class="flex flex-col hover:border-primary/50 transition-colors">
					<CardHeader>
						<div class="flex items-start justify-between gap-2">
							<div class="flex-1">
								<CardTitle class="text-xl">{template.name}</CardTitle>
								{#if template.version}
									<Badge variant="outline" class="mt-2">v{template.version}</Badge>
								{/if}
							</div>
						</div>
						<CardDescription class="line-clamp-2">{template.description}</CardDescription>
					</CardHeader>
					<CardContent class="flex-1 flex flex-col">
						<div class="flex flex-wrap gap-2 mb-4">
							<Badge variant="secondary">{template.category}</Badge>
						</div>
						<div class="mt-auto">
							<Button onclick={() => handleDeploy(template)} class="w-full">Deploy</Button>
						</div>
					</CardContent>
				</Card>
			{/each}
		</div>
	{/if}
</div>

<Sheet bind:open={deploySheetOpen}>
	<SheetContent class="w-full sm:max-w-2xl overflow-y-auto">
		<SheetHeader>
			<SheetTitle>Deploy {selectedTemplate?.name}</SheetTitle>
			<SheetDescription>
				Configure your deployment settings and environment variables
			</SheetDescription>
		</SheetHeader>

		{#if selectedTemplate}
			<form
				onsubmit={(e) => {
					e.preventDefault();
					handleSubmitDeployment();
				}}
				class="space-y-6 py-6"
			>
				<div class="space-y-4">
					<div class="space-y-2">
						<Label for="deployment-name">Service Name *</Label>
						<Input
							id="deployment-name"
							placeholder="my-service"
							bind:value={deploymentName}
							required
						/>
					</div>
				</div>

				<div class="space-y-4">
					<div class="flex items-center justify-between">
						<Label>Environment Variables</Label>
						<Button type="button" variant="outline" size="sm" onclick={handleAddEnvVar}>
							Add Variable
						</Button>
					</div>

					<div class="space-y-3 max-h-96 overflow-y-auto">
						{#each Object.entries(customEnvVars) as [key, value]}
							<div class="flex gap-2">
								<Input placeholder="KEY" value={key} readonly class="flex-1" />
								<Input placeholder="value" bind:value={customEnvVars[key]} class="flex-1" />
								<Button
									type="button"
									variant="ghost"
									size="icon"
									onclick={() => handleRemoveEnvVar(key)}
								>
									<X class="h-4 w-4" />
								</Button>
							</div>
						{/each}
					</div>
				</div>

				<SheetFooter>
					<SheetClose>
						<Button type="button" variant="outline">Cancel</Button>
					</SheetClose>
					<Button type="submit" disabled={deployMutation.isPending}>
						{deployMutation.isPending ? 'Deploying...' : 'Deploy'}
					</Button>
				</SheetFooter>
			</form>
		{/if}
	</SheetContent>
</Sheet>
