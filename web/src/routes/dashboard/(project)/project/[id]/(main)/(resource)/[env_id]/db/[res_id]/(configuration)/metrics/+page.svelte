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
	import { Tabs, TabsContent, TabsList, TabsTrigger } from '$lib/components/ui/tabs';
	import { Activity, TrendingUp, Database, Clock } from 'lucide-svelte';
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

	const metricsData = $state({
		cpu: [
			{ time: '00:00', value: 25 },
			{ time: '04:00', value: 30 },
			{ time: '08:00', value: 45 },
			{ time: '12:00', value: 60 },
			{ time: '16:00', value: 55 },
			{ time: '20:00', value: 35 }
		],
		memory: [
			{ time: '00:00', value: 512 },
			{ time: '04:00', value: 580 },
			{ time: '08:00', value: 720 },
			{ time: '12:00', value: 890 },
			{ time: '16:00', value: 820 },
			{ time: '20:00', value: 640 }
		]
	});
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h2 class="text-2xl font-bold tracking-tight">Performance Metrics</h2>
			<p class="text-muted-foreground">Monitor resource usage for {database.name}</p>
		</div>
		<Button variant="outline">
			<Clock class="h-4 w-4 mr-2" />
			Last 24 hours
		</Button>
	</div>

	<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
		<Card>
			<CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
				<CardTitle class="text-sm font-medium">CPU Usage</CardTitle>
				<Activity class="h-4 w-4 text-muted-foreground" />
			</CardHeader>
			<CardContent>
				<div class="text-2xl font-bold">32%</div>
				<p class="text-xs text-muted-foreground">
					<span class="text-green-500">↓ 12%</span> from last hour
				</p>
			</CardContent>
		</Card>
		<Card>
			<CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
				<CardTitle class="text-sm font-medium">Memory Usage</CardTitle>
				<TrendingUp class="h-4 w-4 text-muted-foreground" />
			</CardHeader>
			<CardContent>
				<div class="text-2xl font-bold">648 MB</div>
				<p class="text-xs text-muted-foreground">
					<span class="text-red-500">↑ 8%</span> from last hour
				</p>
			</CardContent>
		</Card>
		<Card>
			<CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
				<CardTitle class="text-sm font-medium">Connections</CardTitle>
				<Database class="h-4 w-4 text-muted-foreground" />
			</CardHeader>
			<CardContent>
				<div class="text-2xl font-bold">24</div>
				<p class="text-xs text-muted-foreground">Active connections</p>
			</CardContent>
		</Card>
		<Card>
			<CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
				<CardTitle class="text-sm font-medium">Queries/sec</CardTitle>
				<Activity class="h-4 w-4 text-muted-foreground" />
			</CardHeader>
			<CardContent>
				<div class="text-2xl font-bold">1,247</div>
				<p class="text-xs text-muted-foreground">
					<span class="text-green-500">↑ 23%</span> from last hour
				</p>
			</CardContent>
		</Card>
	</div>

	<Tabs value="cpu" class="space-y-4">
		<TabsList>
			<TabsTrigger value="cpu">CPU</TabsTrigger>
			<TabsTrigger value="memory">Memory</TabsTrigger>
			<TabsTrigger value="disk">Disk I/O</TabsTrigger>
			<TabsTrigger value="network">Network</TabsTrigger>
		</TabsList>

		<TabsContent value="cpu" class="space-y-4">
			<Card>
				<CardHeader>
					<CardTitle>CPU Usage Over Time</CardTitle>
					<CardDescription>Percentage of CPU cores used</CardDescription>
				</CardHeader>
				<CardContent>
					<div class="h-[300px] flex items-end justify-between gap-2">
						{#each metricsData.cpu as point}
							<div class="flex-1 flex flex-col items-center gap-2">
								<div class="w-full bg-primary rounded-t" style="height: {point.value * 3}px"></div>
								<span class="text-xs text-muted-foreground">{point.time}</span>
							</div>
						{/each}
					</div>
					<div class="mt-4 text-center text-sm text-muted-foreground">Peak: 60% at 12:00</div>
				</CardContent>
			</Card>
		</TabsContent>

		<TabsContent value="memory" class="space-y-4">
			<Card>
				<CardHeader>
					<CardTitle>Memory Usage Over Time</CardTitle>
					<CardDescription>Memory consumption in MB</CardDescription>
				</CardHeader>
				<CardContent>
					<div class="h-[300px] flex items-end justify-between gap-2">
						{#each metricsData.memory as point}
							<div class="flex-1 flex flex-col items-center gap-2">
								<div class="w-full bg-primary rounded-t" style="height: {point.value / 3}px"></div>
								<span class="text-xs text-muted-foreground">{point.time}</span>
							</div>
						{/each}
					</div>
					<div class="mt-4 text-center text-sm text-muted-foreground">Peak: 890 MB at 12:00</div>
				</CardContent>
			</Card>
		</TabsContent>

		<TabsContent value="disk" class="space-y-4">
			<Card>
				<CardHeader>
					<CardTitle>Disk I/O</CardTitle>
					<CardDescription>Read and write operations per second</CardDescription>
				</CardHeader>
				<CardContent>
					<div class="text-center py-12 text-muted-foreground">
						<Database class="mx-auto h-12 w-12 mb-2 opacity-50" />
						<p>Disk I/O metrics will be displayed here</p>
					</div>
				</CardContent>
			</Card>
		</TabsContent>

		<TabsContent value="network" class="space-y-4">
			<Card>
				<CardHeader>
					<CardTitle>Network Traffic</CardTitle>
					<CardDescription>Inbound and outbound network traffic</CardDescription>
				</CardHeader>
				<CardContent>
					<div class="text-center py-12 text-muted-foreground">
						<Activity class="mx-auto h-12 w-12 mb-2 opacity-50" />
						<p>Network metrics will be displayed here</p>
					</div>
				</CardContent>
			</Card>
		</TabsContent>
	</Tabs>
</div>
