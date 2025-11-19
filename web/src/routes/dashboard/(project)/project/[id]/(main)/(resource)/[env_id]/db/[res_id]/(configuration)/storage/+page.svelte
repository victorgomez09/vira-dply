<script lang="ts">
	import { page } from '$app/state';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Badge } from '$lib/components/ui/badge';
	import { HardDrive, Plus, Trash2 } from 'lucide-svelte';
	import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
	import { databasesApi } from '$lib/api/databases';
	import { disksApi, type CreateDiskRequest } from '$lib/api/disks';

	const projectId = $derived(page.params.id);
	const resId = $derived(page.params.res_id);
	const queryClient = useQueryClient();

	const databaseQuery = createQuery(() => ({
		queryKey: ['database', projectId, resId],
		queryFn: () => databasesApi.get(projectId, resId),
		enabled: !!projectId && !!resId
	}));

	const disksQuery = createQuery(() => ({
		queryKey: ['disks', projectId],
		queryFn: () => disksApi.list(projectId),
		enabled: !!projectId
	}));

	const deleteMutation = createMutation(() => ({
		mutationFn: (diskId: string) => disksApi.delete(projectId, diskId),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['disks', projectId] });
		}
	}));

	const detachMutation = createMutation(() => ({
		mutationFn: (diskId: string) => disksApi.detach(projectId, diskId),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['disks', projectId] });
		}
	}));

	const createMutation_ = createMutation(() => ({
		mutationFn: async (data: CreateDiskRequest) => {
			const disk = await disksApi.create(projectId, data);
			await disksApi.attach(projectId, disk.id, { service_id: resId });
			return disk;
		},
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['disks', projectId] });
			mountPath = '';
			volumeSize = '';
		}
	}));

	const database = $derived(databaseQuery.data);
	const allDisks = $derived(disksQuery.data ?? []);
	const mountedVolumes = $derived(allDisks.filter((disk) => disk.service_id === resId));

	let mountPath = $state('');
	let volumeSize = $state('');

	function handleSubmit(e: Event) {
		e.preventDefault();
		if (!mountPath || !volumeSize) return;

		const data: CreateDiskRequest = {
			name: `${database?.name || 'db'}-storage-${Date.now()}`,
			size_gb: parseInt(volumeSize),
			mount_path: mountPath,
			filesystem: 'ext4',
			persistent: true
		};

		createMutation_.mutate(data);
	}

	function handleDetach(diskId: string) {
		if (confirm('Are you sure you want to detach this volume?')) {
			detachMutation.mutate(diskId);
		}
	}
</script>

<div class="space-y-6">
	<div>
		<h2 class="text-2xl font-bold tracking-tight">Persistent Storage</h2>
		<p class="text-muted-foreground">
			Manage persistent storage volumes for {database?.name || 'database'}
		</p>
	</div>

	<Card>
		<CardHeader>
			<CardTitle>Mounted Volumes</CardTitle>
			<CardDescription>Persistent volumes attached to this database container</CardDescription>
		</CardHeader>
		<CardContent>
			<div class="space-y-4">
				{#if disksQuery.isLoading}
					<div class="text-center py-8 text-muted-foreground">
						<p>Loading volumes...</p>
					</div>
				{:else if mountedVolumes.length === 0}
					<div class="text-center py-8 text-muted-foreground">
						<HardDrive class="mx-auto h-12 w-12 mb-2 opacity-50" />
						<p>No persistent volumes mounted</p>
					</div>
				{:else}
					<div class="space-y-3">
						{#each mountedVolumes as volume}
							<div class="flex items-center justify-between p-4 border rounded-lg">
								<div class="flex items-center gap-3">
									<HardDrive class="h-5 w-5 text-muted-foreground" />
									<div>
										<p class="font-medium">{volume.mount_path}</p>
										<p class="text-sm text-muted-foreground">
											{volume.size_gb} GB • {volume.filesystem}
										</p>
									</div>
								</div>
								<div class="flex items-center gap-2">
									<Badge variant="secondary">{volume.status}</Badge>
									<Button
										variant="ghost"
										size="icon"
										onclick={() => handleDetach(volume.id)}
										disabled={detachMutation.isPending}
									>
										<Trash2 class="h-4 w-4" />
									</Button>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</CardContent>
	</Card>

	<Card>
		<CardHeader>
			<CardTitle>Add New Volume</CardTitle>
			<CardDescription>Mount a new persistent volume to this database</CardDescription>
		</CardHeader>
		<CardContent>
			<form class="space-y-4" onsubmit={handleSubmit}>
				<div class="grid gap-4 sm:grid-cols-2">
					<div class="space-y-2">
						<Label for="mount-path">Mount Path</Label>
						<Input id="mount-path" placeholder="/data" type="text" bind:value={mountPath} />
					</div>
					<div class="space-y-2">
						<Label for="volume-size">Size (GB)</Label>
						<Input
							id="volume-size"
							placeholder="10"
							type="number"
							min="1"
							bind:value={volumeSize}
						/>
					</div>
				</div>
				<Button
					type="submit"
					class="w-full sm:w-auto"
					disabled={createMutation_.isPending || !mountPath || !volumeSize}
				>
					<Plus class="h-4 w-4 mr-2" />
					{createMutation_.isPending ? 'Adding...' : 'Add Volume'}
				</Button>
			</form>
		</CardContent>
	</Card>
</div>
