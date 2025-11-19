<script lang="ts">
	import { page } from '$app/state';
	import { applicationsApi } from '$lib/api/applications';
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
	import { Textarea } from '$lib/components/ui/textarea';
	import { Save, Loader2 } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';

	const projectId = $derived(page.params.id);
	const applicationId = $derived(page.params.res_id);

	let name = $state('');
	let description = $state('');
	let loading = $state(false);
	let saving = $state(false);

	$effect(() => {
		loadApplication();
	});

	async function loadApplication() {
		loading = true;
		try {
			const app = await applicationsApi.get(projectId, applicationId);
			name = app.name;
			description = app.description || '';
		} catch (error) {
			toast.error('Failed to load application');
		} finally {
			loading = false;
		}
	}

	async function saveChanges() {
		if (!name.trim()) {
			toast.error('Application name is required');
			return;
		}

		saving = true;
		try {
			await applicationsApi.updateGeneral(projectId, applicationId, {
				name: name.trim(),
				description: description.trim() || undefined
			});
			toast.success('Application updated successfully');
		} catch (error) {
			toast.error('Failed to update application');
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>General Settings</title>
</svelte:head>

<div class="flex-1 p-6">
	<div class="max-w-2xl">
		<div class="mb-6">
			<h1 class="text-2xl font-semibold text-gray-900">General Settings</h1>
			<p class="text-sm text-gray-500 mt-1">Update your application's basic information.</p>
		</div>

		{#if loading}
			<div class="flex items-center justify-center py-12">
				<Loader2 class="w-8 h-8 animate-spin text-gray-400" />
			</div>
		{:else}
			<Card>
				<CardHeader>
					<CardTitle>Application Details</CardTitle>
					<CardDescription>Basic information about your application</CardDescription>
				</CardHeader>
				<CardContent class="space-y-4">
					<div class="space-y-2">
						<Label for="name">Application Name</Label>
						<Input id="name" bind:value={name} placeholder="My Application" disabled={saving} />
					</div>

					<div class="space-y-2">
						<Label for="description">Description</Label>
						<Textarea
							id="description"
							bind:value={description}
							placeholder="A brief description of your application"
							rows={4}
							disabled={saving}
						/>
					</div>

					<div class="flex justify-end pt-4">
						<Button onclick={saveChanges} disabled={saving}>
							{#if saving}
								<Loader2 class="w-4 h-4 mr-2 animate-spin" />
								Saving...
							{:else}
								<Save class="w-4 h-4 mr-2" />
								Save Changes
							{/if}
						</Button>
					</div>
				</CardContent>
			</Card>
		{/if}
	</div>
</div>
