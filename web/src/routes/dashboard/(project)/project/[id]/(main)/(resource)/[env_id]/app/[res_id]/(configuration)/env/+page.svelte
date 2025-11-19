<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import {
		ChevronDown,
		ArrowLeft,
		Plus,
		Eye,
		EyeOff,
		Copy,
		Edit,
		Trash2,
		Save,
		X
	} from 'lucide-svelte';

	const projectId = page.params.id;

	// Mock project data
	let project = $state({
		name: 'Focalpoint Dashboard',
		repo: 'focalpoint/dashboard',
		branch: 'main',
		workspace: 'Focalpoint',
		category: 'Applications'
	});

	let showAddForm = $state(false);
	let editingVar = $state(null);

	// New variable form
	let newVar = $state({
		key: '',
		value: '',
		description: ''
	});

	// Mock environment variables
	let envVars = $state([
		{
			id: 1,
			key: 'DATABASE_URL',
			value: 'postgresql://user:pass@localhost:5432/mydb',
			description: 'Primary database connection string',
			visible: false,
			lastModified: '2 days ago'
		},
		{
			id: 2,
			key: 'API_KEY',
			value: 'sk-1234567890abcdef',
			description: 'Third-party API authentication key',
			visible: false,
			lastModified: '1 week ago'
		},
		{
			id: 3,
			key: 'NODE_ENV',
			value: 'production',
			description: 'Application environment',
			visible: true,
			lastModified: '1 month ago'
		},
		{
			id: 4,
			key: 'REDIS_URL',
			value: 'redis://localhost:6379',
			description: 'Redis cache connection string',
			visible: false,
			lastModified: '3 days ago'
		}
	]);

	function toggleVisibility(id: number) {
		const envVar = envVars.find((v) => v.id === id);
		if (envVar) {
			envVar.visible = !envVar.visible;
		}
	}

	function copyToClipboard(value: string) {
		navigator.clipboard.writeText(value);
		// You could add a toast notification here
	}

	function startEditing(envVar) {
		editingVar = { ...envVar };
	}

	function cancelEditing() {
		editingVar = null;
	}

	function saveEdit() {
		const index = envVars.findIndex((v) => v.id === editingVar.id);
		if (index !== -1) {
			envVars[index] = { ...editingVar, lastModified: 'just now' };
		}
		editingVar = null;
	}

	function deleteVar(id: number) {
		envVars = envVars.filter((v) => v.id !== id);
	}

	function addNewVar() {
		if (newVar.key && newVar.value) {
			const newId = Math.max(...envVars.map((v) => v.id)) + 1;
			envVars = [
				...envVars,
				{
					id: newId,
					key: newVar.key,
					value: newVar.value,
					description: newVar.description,
					visible: false,
					lastModified: 'just now'
				}
			];
			newVar = { key: '', value: '', description: '' };
			showAddForm = false;
		}
	}
</script>

<svelte:head>
	<title>Environment Variables - {project.name}</title>
</svelte:head>

<div class="flex-1 p-6">
	<!-- Header -->
	<div class="flex items-center justify-between mb-6">
		<div class="flex items-center space-x-4">
			<div>
				<h1 class="text-2xl font-semibold text-gray-900">Environment Variables</h1>
				<p class="text-sm text-gray-500 mt-1">
					Manage environment variables for {project.name}.
				</p>
			</div>
		</div>
		<Button onclick={() => (showAddForm = true)}>
			<Plus class="w-4 h-4 mr-2" />
			Add Variable
		</Button>
	</div>

	<!-- Add New Variable Form -->
	{#if showAddForm}
		<Card class="mb-6">
			<CardHeader>
				<CardTitle>Add New Environment Variable</CardTitle>
			</CardHeader>
			<CardContent class="space-y-4">
				<div class="grid grid-cols-2 gap-4">
					<div>
						<Label for="key">Key</Label>
						<Input id="key" bind:value={newVar.key} placeholder="VARIABLE_NAME" class="font-mono" />
					</div>
					<div>
						<Label for="value">Value</Label>
						<Input
							id="value"
							bind:value={newVar.value}
							placeholder="variable_value"
							class="font-mono"
						/>
					</div>
				</div>
				<div>
					<Label for="description">Description (optional)</Label>
					<Input
						id="description"
						bind:value={newVar.description}
						placeholder="Brief description of this variable"
					/>
				</div>
				<div class="flex space-x-2">
					<Button onclick={addNewVar}>
						<Save class="w-4 h-4 mr-2" />
						Add Variable
					</Button>
					<Button variant="outline" onclick={() => (showAddForm = false)}>
						<X class="w-4 h-4 mr-2" />
						Cancel
					</Button>
				</div>
			</CardContent>
		</Card>
	{/if}

	<!-- Environment Variables List -->
	<div class="space-y-4">
		{#each envVars as envVar (envVar.id)}
			<Card>
				<CardContent class="p-6">
					{#if editingVar && editingVar.id === envVar.id}
						<!-- Edit Mode -->
						<div class="space-y-4">
							<div class="grid grid-cols-2 gap-4">
								<div>
									<Label for="edit-key">Key</Label>
									<Input id="edit-key" bind:value={editingVar.key} class="font-mono" />
								</div>
								<div>
									<Label for="edit-value">Value</Label>
									<Input id="edit-value" bind:value={editingVar.value} class="font-mono" />
								</div>
							</div>
							<div>
								<Label for="edit-description">Description</Label>
								<Input id="edit-description" bind:value={editingVar.description} />
							</div>
							<div class="flex space-x-2">
								<Button size="sm" onclick={saveEdit}>
									<Save class="w-4 h-4 mr-1" />
									Save
								</Button>
								<Button size="sm" variant="outline" onclick={cancelEditing}>
									<X class="w-4 h-4 mr-1" />
									Cancel
								</Button>
							</div>
						</div>
					{:else}
						<!-- View Mode -->
						<div class="flex items-center justify-between">
							<div class="flex-1 space-y-2">
								<div class="flex items-center space-x-3">
									<h3 class="font-mono font-medium text-gray-900">{envVar.key}</h3>
									<Badge variant="outline" class="text-xs">
										Modified {envVar.lastModified}
									</Badge>
								</div>
								<div class="flex items-center space-x-2">
									<Input
										type={envVar.visible ? 'text' : 'password'}
										value={envVar.visible ? envVar.value : '••••••••••••••••'}
										readonly
										class="font-mono text-sm bg-gray-50 max-w-md"
									/>
									<Button size="sm" variant="outline" onclick={() => toggleVisibility(envVar.id)}>
										{#if envVar.visible}
											<EyeOff class="w-4 h-4" />
										{:else}
											<Eye class="w-4 h-4" />
										{/if}
									</Button>
									<Button size="sm" variant="outline" onclick={() => copyToClipboard(envVar.value)}>
										<Copy class="w-4 h-4" />
									</Button>
								</div>
								{#if envVar.description}
									<p class="text-sm text-gray-600">{envVar.description}</p>
								{/if}
							</div>
							<div class="flex items-center space-x-2">
								<Button size="sm" variant="outline" onclick={() => startEditing(envVar)}>
									<Edit class="w-4 h-4" />
								</Button>
								<Button size="sm" variant="outline" onclick={() => deleteVar(envVar.id)}>
									<Trash2 class="w-4 h-4" />
								</Button>
							</div>
						</div>
					{/if}
				</CardContent>
			</Card>
		{/each}
	</div>
</div>
