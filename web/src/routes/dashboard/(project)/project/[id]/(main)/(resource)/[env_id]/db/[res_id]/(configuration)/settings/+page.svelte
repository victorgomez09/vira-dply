<script lang="ts">
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Button } from '$lib/components/ui/button';
	import { page } from '$app/state';
	import { createQuery } from '@tanstack/svelte-query';
	import { databasesApi } from '$lib/api/databases';
	import { Trash2 } from 'lucide-svelte';

	const projectId = $derived(page.params.id);
	const resId = $derived(page.params.res_id);

	const databaseQuery = createQuery(() => ({
		queryKey: ['database', projectId, resId],
		queryFn: () => databasesApi.get(projectId, resId),
		enabled: !!projectId && !!resId
	}));

	const database = $derived(databaseQuery.data);

	let name = $state('');
	let description = $state('');

	$effect(() => {
		if (database) {
			name = database.name;
			description = database.description || '';
		}
	});
</script>

{#if database}
	<div class="space-y-6">
		<Card>
			<CardHeader>
				<CardTitle>General Settings</CardTitle>
				<p class="text-sm text-muted-foreground">
					Update your database name and description
				</p>
			</CardHeader>
			<CardContent class="space-y-4">
				<div class="space-y-2">
					<Label for="db-name">Database Name</Label>
					<Input id="db-name" bind:value={name} placeholder="my-database" />
				</div>

				<div class="space-y-2">
					<Label for="db-description">Description</Label>
					<Textarea
						id="db-description"
						bind:value={description}
						placeholder="Describe your database..."
						rows={3}
					/>
				</div>

				<div class="pt-2">
					<Button>Save Changes</Button>
				</div>
			</CardContent>
		</Card>

		<Card>
			<CardHeader>
				<CardTitle>Environment Variables</CardTitle>
				<p class="text-sm text-muted-foreground">
					Manage environment variables for your database
				</p>
			</CardHeader>
			<CardContent>
				<div class="text-center py-8 text-muted-foreground">
					<p>No environment variables configured</p>
					<Button variant="outline" class="mt-4">Add Variable</Button>
				</div>
			</CardContent>
		</Card>

		<Card class="border-destructive">
			<CardHeader>
				<CardTitle class="text-destructive">Danger Zone</CardTitle>
				<p class="text-sm text-muted-foreground">
					Irreversible and destructive actions
				</p>
			</CardHeader>
			<CardContent class="space-y-4">
				<div class="flex items-center justify-between p-4 border border-destructive rounded-lg">
					<div>
						<h4 class="font-medium">Delete Database</h4>
						<p class="text-sm text-muted-foreground">
							Permanently delete this database and all of its data
						</p>
					</div>
					<Button variant="destructive">
						<Trash2 class="mr-2 h-4 w-4" />
						Delete
					</Button>
				</div>
			</CardContent>
		</Card>
	</div>
{/if}
