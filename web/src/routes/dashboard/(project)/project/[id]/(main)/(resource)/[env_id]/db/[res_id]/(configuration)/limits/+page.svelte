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
	import { Slider } from '$lib/components/ui/slider';
	import { Badge } from '$lib/components/ui/badge';
	import { Cpu, MemoryStick } from 'lucide-svelte';
	import { createQuery } from '@tanstack/svelte-query';
	import { databasesApi } from '$lib/api/databases';

	const projectId = $derived(page.params.id);
	const resId = $derived(page.params.res_id);

	const databaseQuery = createQuery(() => ({
		queryKey: ['database', projectId, resId],
		queryFn: () => databasesApi.get(projectId, resId),
		enabled: !!projectId && !!resId
	}));

	const database = $derived(databaseQuery.data);

	let cpuLimit = $state([1]);
	let cpuReservation = $state([0.5]);
	let memoryLimit = $state([1024]);
	let memoryReservation = $state([512]);
</script>

<div class="space-y-6">
	<div>
		<h2 class="text-2xl font-bold tracking-tight">Resource Limits</h2>
		<p class="text-muted-foreground">
			Configure CPU and memory limits for {database?.name || 'database'}
		</p>
	</div>

	<Card>
		<CardHeader>
			<CardTitle>Current Usage</CardTitle>
			<CardDescription>Real-time resource consumption</CardDescription>
		</CardHeader>
		<CardContent>
			<div class="grid gap-4 md:grid-cols-2">
				<div class="space-y-2">
					<div class="flex items-center gap-2">
						<Cpu class="h-4 w-4 text-muted-foreground" />
						<span class="font-medium">CPU</span>
					</div>
					<div class="space-y-1">
						<div class="flex items-center justify-between text-sm">
							<span class="text-muted-foreground">Current</span>
							<span class="font-medium">0.3 cores</span>
						</div>
						<div class="h-2 bg-secondary rounded-full overflow-hidden">
							<div class="h-full bg-primary" style="width: 30%"></div>
						</div>
					</div>
				</div>
				<div class="space-y-2">
					<div class="flex items-center gap-2">
						<MemoryStick class="h-4 w-4 text-muted-foreground" />
						<span class="font-medium">Memory</span>
					</div>
					<div class="space-y-1">
						<div class="flex items-center justify-between text-sm">
							<span class="text-muted-foreground">Current</span>
							<span class="font-medium">384 MB</span>
						</div>
						<div class="h-2 bg-secondary rounded-full overflow-hidden">
							<div class="h-full bg-primary" style="width: 37.5%"></div>
						</div>
					</div>
				</div>
			</div>
		</CardContent>
	</Card>

	<Card>
		<CardHeader>
			<CardTitle>CPU Limits</CardTitle>
			<CardDescription>Configure CPU allocation and limits for this database</CardDescription>
		</CardHeader>
		<CardContent>
			<form class="space-y-6">
				<div class="space-y-4">
					<div class="space-y-2">
						<div class="flex items-center justify-between">
							<Label for="cpu-limit">CPU Limit (cores)</Label>
							<Badge variant="secondary">{cpuLimit[0]}</Badge>
						</div>
						<Slider id="cpu-limit" bind:value={cpuLimit} min={0.1} max={8} step={0.1} />
						<p class="text-xs text-muted-foreground">Maximum CPU cores the database can use</p>
					</div>
					<div class="space-y-2">
						<div class="flex items-center justify-between">
							<Label for="cpu-reservation">CPU Reservation (cores)</Label>
							<Badge variant="secondary">{cpuReservation[0]}</Badge>
						</div>
						<Slider
							id="cpu-reservation"
							bind:value={cpuReservation}
							min={0.1}
							max={cpuLimit[0]}
							step={0.1}
						/>
						<p class="text-xs text-muted-foreground">
							Guaranteed CPU cores reserved for the database
						</p>
					</div>
				</div>
			</form>
		</CardContent>
	</Card>

	<Card>
		<CardHeader>
			<CardTitle>Memory Limits</CardTitle>
			<CardDescription>Configure memory allocation and limits for this database</CardDescription>
		</CardHeader>
		<CardContent>
			<form class="space-y-6">
				<div class="space-y-4">
					<div class="space-y-2">
						<div class="flex items-center justify-between">
							<Label for="memory-limit">Memory Limit (MB)</Label>
							<Input
								id="memory-limit"
								type="number"
								bind:value={memoryLimit[0]}
								min="128"
								step="128"
								class="w-24"
							/>
						</div>
						<Slider
							id="memory-limit-slider"
							bind:value={memoryLimit}
							min={128}
							max={16384}
							step={128}
						/>
						<p class="text-xs text-muted-foreground">Maximum memory the database can use</p>
					</div>
					<div class="space-y-2">
						<div class="flex items-center justify-between">
							<Label for="memory-reservation">Memory Reservation (MB)</Label>
							<Input
								id="memory-reservation"
								type="number"
								bind:value={memoryReservation[0]}
								min="128"
								step="128"
								class="w-24"
							/>
						</div>
						<Slider
							id="memory-reservation-slider"
							bind:value={memoryReservation}
							min={128}
							max={memoryLimit[0]}
							step={128}
						/>
						<p class="text-xs text-muted-foreground">Guaranteed memory reserved for the database</p>
					</div>
				</div>
			</form>
		</CardContent>
	</Card>

	<div class="flex justify-end">
		<Button type="submit">Save Resource Limits</Button>
	</div>
</div>
