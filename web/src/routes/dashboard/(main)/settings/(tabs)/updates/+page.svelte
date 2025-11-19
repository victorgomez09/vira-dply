<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Switch } from '$lib/components/ui/switch';
	import * as Card from '$lib/components/ui/card';

	let updateCheckFrequency = $state('hourly');
	let autoUpdate = $state(true);
	let autoUpdateFrequency = $state('daily');
	let autoUpdateTime = $state('00:00');

	const cronPresets = [
		{ value: 'every_minute', label: 'Every Minute' },
		{ value: 'hourly', label: 'Hourly' },
		{ value: 'daily', label: 'Daily' },
		{ value: 'weekly', label: 'Weekly' },
		{ value: 'monthly', label: 'Monthly' },
		{ value: 'yearly', label: 'Yearly' }
	];

	onMount(async () => {
		try {
			const response = await fetch('/api/settings/updates');
			if (response.ok) {
				const data = await response.json();
				updateCheckFrequency = data.update_check_frequency || 'hourly';
				autoUpdate = data.auto_update ?? true;
				autoUpdateFrequency = data.auto_update_frequency || 'daily';
				autoUpdateTime = data.auto_update_time || '00:00';
			}
		} catch (error) {
			console.error('Failed to load settings:', error);
		}
	});

	async function handleSave() {
		try {
			const response = await fetch('/api/settings/updates', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					update_check_frequency: updateCheckFrequency,
					auto_update: autoUpdate,
					auto_update_frequency: autoUpdateFrequency,
					auto_update_time: autoUpdateTime
				})
			});

			if (!response.ok) {
				throw new Error('Failed to save settings');
			}
		} catch (error) {
			console.error('Failed to save settings:', error);
		}
	}
</script>

<div class="space-y-6">
	<Card.Root>
		<Card.Header>
			<Card.Title>Version Checks</Card.Title>
		</Card.Header>
		<Card.Content class="space-y-6">
			<div class="space-y-2">
				<Label for="update-check">Update Check Frequency</Label>
				<select
					id="update-check"
					bind:value={updateCheckFrequency}
					class="border-input-new bg-secondary-new selection:bg-primary dark:bg-input/30 selection:text-primary-foreground ring-offset-background shadow-xs flex h-9 w-full min-w-0 rounded-md border px-3 py-1 text-base outline-none transition-[color,box-shadow] disabled:cursor-not-allowed disabled:opacity-50 md:text-sm focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]"
				>
					{#each cronPresets as preset}
						<option value={preset.value}>{preset.label}</option>
					{/each}
				</select>
				<p class="text-xs text-muted-foreground">
					Frequency (cron expression) to check for new versions and pull new Service Templates from
					CDN. Default is every hour.
				</p>
			</div>
		</Card.Content>
	</Card.Root>

	<Card.Root>
		<Card.Header>
			<Card.Title>Automatic Updates</Card.Title>
		</Card.Header>
		<Card.Content class="space-y-6">
			<div class="flex items-center justify-between">
				<div class="space-y-1">
					<Label for="auto-update">Enable Automatic Updates</Label>
					<p class="text-xs text-muted-foreground">
						Automatically install updates when they become available (coming soon).
					</p>
				</div>
				<Switch id="auto-update" bind:checked={autoUpdate} />
			</div>

			{#if autoUpdate}
				<div class="space-y-2">
					<Label for="auto-update-frequency">Auto Update Frequency</Label>
					<select
						id="auto-update-frequency"
						bind:value={autoUpdateFrequency}
						class="border-input-new bg-secondary-new selection:bg-primary dark:bg-input/30 selection:text-primary-foreground ring-offset-background shadow-xs flex h-9 w-full min-w-0 rounded-md border px-3 py-1 text-base outline-none transition-[color,box-shadow] disabled:cursor-not-allowed disabled:opacity-50 md:text-sm focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]"
					>
						{#each cronPresets as preset}
							<option value={preset.value}>{preset.label}</option>
						{/each}
					</select>
					<p class="text-xs text-muted-foreground">
						How often to check and install updates. Default is daily at midnight.
					</p>
				</div>

				{#if autoUpdateFrequency === 'daily'}
					<div class="space-y-2">
						<Label for="auto-update-time">Update Time</Label>
						<Input id="auto-update-time" type="time" bind:value={autoUpdateTime} />
						<p class="text-xs text-muted-foreground">Time of day to run automatic updates.</p>
					</div>
				{/if}
			{/if}
		</Card.Content>
	</Card.Root>

	<div class="flex justify-end">
		<Button onclick={handleSave}>Save Changes</Button>
	</div>
</div>
