<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Label } from '$lib/components/ui/label';
	import * as Card from '$lib/components/ui/card';
	import { Download, Upload, Database } from 'lucide-svelte';

	let lastBackup = $state<string | null>(null);

	async function handleCreateBackup() {
		try {
			const response = await fetch('/api/settings/backup', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' }
			});

			if (!response.ok) {
				throw new Error('Failed to create backup');
			}

			const blob = await response.blob();
			const url = window.URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = `mikrocloud-backup-${new Date().toISOString()}.tar.gz`;
			document.body.appendChild(a);
			a.click();
			window.URL.revokeObjectURL(url);
			document.body.removeChild(a);

			lastBackup = new Date().toLocaleString();
		} catch (error) {
			console.error('Failed to create backup:', error);
		}
	}

	async function handleRestoreBackup(event: Event) {
		const input = event.target as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;

		try {
			const formData = new FormData();
			formData.append('backup', file);

			const response = await fetch('/api/settings/restore', {
				method: 'POST',
				body: formData
			});

			if (!response.ok) {
				throw new Error('Failed to restore backup');
			}
		} catch (error) {
			console.error('Failed to restore backup:', error);
		}
	}
</script>

<div class="space-y-6">
	<Card.Root>
		<Card.Header>
			<Card.Title>Create Backup</Card.Title>
			<Card.Description>
				Download a complete backup of your MikroCloud instance including all databases,
				configurations, and settings.
			</Card.Description>
		</Card.Header>
		<Card.Content class="space-y-6">
			<div class="flex items-center justify-between p-4 border rounded-lg">
				<div class="flex items-center gap-3">
					<div class="p-2 bg-primary/10 rounded-lg">
						<Database class="h-5 w-5 text-primary" />
					</div>
					<div>
						<p class="text-sm font-medium">Full System Backup</p>
						<p class="text-xs text-muted-foreground">
							{#if lastBackup}
								Last backup: {lastBackup}
							{:else}
								No backups created yet
							{/if}
						</p>
					</div>
				</div>
				<Button onclick={handleCreateBackup}>
					<Download class="h-4 w-4 mr-2" />
					Create Backup
				</Button>
			</div>
		</Card.Content>
	</Card.Root>

	<Card.Root>
		<Card.Header>
			<Card.Title>Restore Backup</Card.Title>
			<Card.Description>
				Upload and restore a previous backup. Warning: This will overwrite all current data.
			</Card.Description>
		</Card.Header>
		<Card.Content class="space-y-6">
			<div class="flex items-center justify-between p-4 border border-dashed rounded-lg">
				<div class="flex items-center gap-3">
					<div class="p-2 bg-orange-500/10 rounded-lg">
						<Upload class="h-5 w-5 text-orange-500" />
					</div>
					<div>
						<p class="text-sm font-medium">Upload Backup File</p>
						<p class="text-xs text-muted-foreground">Select a .tar.gz backup file to restore</p>
					</div>
				</div>
				<label>
					<input type="file" accept=".tar.gz,.tgz" class="hidden" onchange={handleRestoreBackup} />
					<Button>Select File</Button>
				</label>
			</div>

			<div
				class="bg-orange-50 dark:bg-orange-950/20 border border-orange-200 dark:border-orange-900 rounded-lg p-4"
			>
				<p class="text-sm font-medium text-orange-900 dark:text-orange-100">⚠️ Warning</p>
				<p class="text-xs text-orange-800 dark:text-orange-200 mt-1">
					Restoring a backup will replace all current data with the data from the backup file. This
					action cannot be undone. Make sure to create a backup of your current data before
					proceeding.
				</p>
			</div>
		</Card.Content>
	</Card.Root>
</div>
