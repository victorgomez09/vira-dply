<script lang="ts">
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { AlertCircle, Download } from 'lucide-svelte';

	let activeBackupType = $state<'daily' | 'manual'>('daily');
</script>

<div class="space-y-6">
	<div class="flex gap-4 border-b pb-0">
		<button
			onclick={() => (activeBackupType = 'daily')}
			class="px-4 py-2 -mb-px {activeBackupType === 'daily'
				? 'border-b-2 border-primary font-medium'
				: 'text-muted-foreground'}"
		>
			Daily
		</button>
		<button
			onclick={() => (activeBackupType = 'manual')}
			class="px-4 py-2 -mb-px {activeBackupType === 'manual'
				? 'border-b-2 border-primary font-medium'
				: 'text-muted-foreground'}"
		>
			Manual
		</button>
	</div>

	{#if activeBackupType === 'daily'}
		<Card>
			<CardHeader>
				<div class="flex items-center gap-2">
					<CardTitle>Daily backups</CardTitle>
					<AlertCircle class="h-4 w-4 text-muted-foreground" />
				</div>
				<p class="text-sm text-muted-foreground">
					We automatically back up your database every day. Each daily backup will be stored for 7
					days.
				</p>
			</CardHeader>
			<CardContent>
				<div class="space-y-4">
					<div class="grid grid-cols-[1fr_1fr_auto] gap-4 pb-2 border-b font-medium">
						<div>Created ↓</div>
						<div>Expiry</div>
						<div>Restore</div>
					</div>

					<div class="py-8 text-center text-muted-foreground">
						<p>No backups available yet</p>
						<p class="text-sm">Daily backups will appear here once created</p>
					</div>
				</div>
			</CardContent>
		</Card>
	{:else}
		<div class="space-y-4">
			<div class="flex justify-between items-center">
				<p class="text-sm text-muted-foreground">
					Create manual backups to preserve specific database states
				</p>
				<Button>
					<Download class="mr-2 h-4 w-4" />
					Create Manual Backup
				</Button>
			</div>

			<Card>
				<CardContent class="pt-6">
					<div class="space-y-4">
						<div class="grid grid-cols-[1fr_1fr_auto] gap-4 pb-2 border-b font-medium">
							<div>Created ↓</div>
							<div>Name</div>
							<div>Actions</div>
						</div>

						<div class="py-8 text-center text-muted-foreground">
							<p>No manual backups created</p>
							<p class="text-sm">Manual backups will appear here</p>
						</div>
					</div>
				</CardContent>
			</Card>
		</div>
	{/if}
</div>
