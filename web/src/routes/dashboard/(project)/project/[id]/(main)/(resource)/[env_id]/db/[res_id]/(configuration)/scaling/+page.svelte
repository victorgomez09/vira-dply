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
	import { Switch } from '$lib/components/ui/switch';
	import { Badge } from '$lib/components/ui/badge';
	import { Scale, Plus, Minus } from 'lucide-svelte';
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

	let autoScaling = $state(false);
	let minInstances = $state(1);
	let maxInstances = $state(5);
	let targetCPU = $state(70);
	let targetMemory = $state(80);
</script>

{#if database}
	<div class="space-y-6">
		<div>
			<h2 class="text-2xl font-bold tracking-tight">Scaling Configuration</h2>
			<p class="text-muted-foreground">
				Configure horizontal and vertical scaling for {database.name}
			</p>
		</div>

		<Card>
			<CardHeader>
				<CardTitle>Current Scale</CardTitle>
				<CardDescription>Active instances and resource allocation</CardDescription>
			</CardHeader>
			<CardContent>
				<div class="space-y-4">
					<div class="flex items-center justify-between">
						<div>
							<p class="font-medium">Running Instances</p>
							<p class="text-sm text-muted-foreground">Current number of replicas</p>
						</div>
						<Badge variant="secondary" class="text-lg px-4 py-1">1</Badge>
					</div>
					<div class="flex items-center gap-2">
						<Button variant="outline" size="icon">
							<Minus class="h-4 w-4" />
						</Button>
						<span class="text-sm text-muted-foreground">Manual scaling controls</span>
						<Button variant="outline" size="icon">
							<Plus class="h-4 w-4" />
						</Button>
					</div>
				</div>
			</CardContent>
		</Card>

		<Card>
			<CardHeader>
				<CardTitle>Horizontal Auto-Scaling</CardTitle>
				<CardDescription>
					Automatically adjust the number of instances based on resource usage
				</CardDescription>
			</CardHeader>
			<CardContent>
				<form class="space-y-6">
					<div class="flex items-center justify-between">
						<div class="space-y-0.5">
							<Label for="auto-scaling">Enable Auto-Scaling</Label>
							<p class="text-sm text-muted-foreground">
								Scale instances automatically based on CPU and memory thresholds
							</p>
						</div>
						<Switch id="auto-scaling" bind:checked={autoScaling} />
					</div>

					{#if autoScaling}
						<div class="space-y-4 pt-4 border-t">
							<div class="grid gap-4 sm:grid-cols-2">
								<div class="space-y-2">
									<Label for="min-instances">Minimum Instances</Label>
									<Input
										id="min-instances"
										type="number"
										bind:value={minInstances}
										min="1"
										max={maxInstances}
									/>
									<p class="text-xs text-muted-foreground">
										Minimum number of instances to keep running
									</p>
								</div>
								<div class="space-y-2">
									<Label for="max-instances">Maximum Instances</Label>
									<Input
										id="max-instances"
										type="number"
										bind:value={maxInstances}
										min={minInstances}
										max="10"
									/>
									<p class="text-xs text-muted-foreground">
										Maximum number of instances to scale up to
									</p>
								</div>
							</div>

							<div class="space-y-2">
								<Label for="target-cpu">Target CPU Utilization (%)</Label>
								<Input id="target-cpu" type="number" bind:value={targetCPU} min="1" max="100" />
								<p class="text-xs text-muted-foreground">
									Scale up when average CPU usage exceeds this threshold
								</p>
							</div>

							<div class="space-y-2">
								<Label for="target-memory">Target Memory Utilization (%)</Label>
								<Input
									id="target-memory"
									type="number"
									bind:value={targetMemory}
									min="1"
									max="100"
								/>
								<p class="text-xs text-muted-foreground">
									Scale up when average memory usage exceeds this threshold
								</p>
							</div>
						</div>
					{/if}
				</form>
			</CardContent>
		</Card>

		<Card>
			<CardHeader>
				<CardTitle>Vertical Scaling</CardTitle>
				<CardDescription>
					Increase or decrease CPU and memory resources per instance
				</CardDescription>
			</CardHeader>
			<CardContent>
				<div class="space-y-4">
					<div class="p-4 border rounded-lg">
						<div class="flex items-center justify-between mb-2">
							<span class="font-medium">Current Configuration</span>
							<Badge variant="outline">1 vCPU / 1024 MB</Badge>
						</div>
						<p class="text-sm text-muted-foreground">
							Vertical scaling requires restarting the database instance
						</p>
					</div>
					<Button variant="outline" class="w-full">
						<Scale class="h-4 w-4 mr-2" />
						Upgrade Instance Size
					</Button>
				</div>
			</CardContent>
		</Card>

		<div class="flex justify-end gap-2">
			<Button variant="outline">Cancel</Button>
			<Button type="submit">Save Scaling Configuration</Button>
		</div>
	</div>
{/if}
