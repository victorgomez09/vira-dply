<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Plus, Ellipsis, Server, Cpu, HardDrive, Activity } from 'lucide-svelte';

	// Mock data for servers
	let servers = $state([
		{
			id: 1,
			name: 'production-web-01',
			region: 'us-east-1',
			status: 'running',
			cpu: 45,
			memory: 67,
			disk: 23,
			uptime: '15d 4h 23m',
			ip: '192.168.1.10',
			type: 't3.medium'
		},
		{
			id: 2,
			name: 'production-api-01',
			region: 'us-east-1',
			status: 'running',
			cpu: 23,
			memory: 45,
			disk: 67,
			uptime: '12d 8h 15m',
			ip: '192.168.1.11',
			type: 't3.large'
		},
		{
			id: 3,
			name: 'staging-web-01',
			region: 'us-west-2',
			status: 'stopped',
			cpu: 0,
			memory: 0,
			disk: 15,
			uptime: '0m',
			ip: '192.168.2.10',
			type: 't3.small'
		}
	]);

	function getStatusColor(status: string) {
		switch (status) {
			case 'running':
				return 'bg-green-500';
			case 'stopped':
				return 'bg-red-500';
			case 'starting':
				return 'bg-yellow-500 animate-pulse';
			default:
				return 'bg-gray-400';
		}
	}

	function getProgressColor(value: number) {
		if (value > 80) return 'bg-red-500';
		if (value > 60) return 'bg-yellow-500';
		return 'bg-green-500';
	}
</script>

<svelte:head>
	<title>Servers - Dashboard</title>
</svelte:head>

<div class="flex h-screen bg-gray-50">
	<!-- Main Content -->
	<div class="flex-1 flex flex-col overflow-hidden">
		<!-- Header -->
		<div class="bg-white border-b border-gray-200 px-6 py-4">
			<div class="flex items-center justify-between">
				<div>
					<h1 class="text-2xl font-semibold text-gray-900">Servers</h1>
					<p class="text-sm text-gray-500 mt-1">Monitor and manage your server infrastructure.</p>
				</div>
				<Button>
					<Plus class="w-4 h-4 mr-2" />
					New Server
				</Button>
			</div>
		</div>

		<!-- Servers Grid -->
		<div class="flex-1 overflow-auto p-6">
			<div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
				{#each servers as server (server.id)}
					<Card class="hover:shadow-md transition-shadow">
						<CardHeader class="pb-3">
							<div class="flex items-center justify-between">
								<div class="flex items-center space-x-2">
									<Server class="w-5 h-5 text-gray-600" />
									<CardTitle class="text-lg">{server.name}</CardTitle>
								</div>
								<div class="flex items-center space-x-2">
									<div class="w-2 h-2 rounded-full {getStatusColor(server.status)}"></div>
									<Badge variant="outline" class="text-xs">
										{server.status}
									</Badge>
								</div>
							</div>
							<div class="flex items-center space-x-4 text-sm text-gray-600">
								<span>{server.region}</span>
								<span>•</span>
								<span>{server.type}</span>
								<span>•</span>
								<span>{server.ip}</span>
							</div>
						</CardHeader>
						<CardContent>
							<div class="space-y-4">
								<!-- Resource Usage -->
								<div class="space-y-3">
									<div>
										<div class="flex items-center justify-between text-sm mb-1">
											<div class="flex items-center space-x-1">
												<Cpu class="w-4 h-4 text-gray-500" />
												<span class="text-gray-600">CPU</span>
											</div>
											<span class="font-medium">{server.cpu}%</span>
										</div>
										<div class="w-full bg-gray-200 rounded-full h-2">
											<div
												class="h-2 rounded-full {getProgressColor(server.cpu)}"
												style="width: {server.cpu}%"
											></div>
										</div>
									</div>

									<div>
										<div class="flex items-center justify-between text-sm mb-1">
											<div class="flex items-center space-x-1">
												<Activity class="w-4 h-4 text-gray-500" />
												<span class="text-gray-600">Memory</span>
											</div>
											<span class="font-medium">{server.memory}%</span>
										</div>
										<div class="w-full bg-gray-200 rounded-full h-2">
											<div
												class="h-2 rounded-full {getProgressColor(server.memory)}"
												style="width: {server.memory}%"
											></div>
										</div>
									</div>

									<div>
										<div class="flex items-center justify-between text-sm mb-1">
											<div class="flex items-center space-x-1">
												<HardDrive class="w-4 h-4 text-gray-500" />
												<span class="text-gray-600">Disk</span>
											</div>
											<span class="font-medium">{server.disk}%</span>
										</div>
										<div class="w-full bg-gray-200 rounded-full h-2">
											<div
												class="h-2 rounded-full {getProgressColor(server.disk)}"
												style="width: {server.disk}%"
											></div>
										</div>
									</div>
								</div>

								<!-- Uptime -->
								<div class="text-sm">
									<span class="text-gray-500">Uptime:</span>
									<span class="font-medium ml-1">{server.uptime}</span>
								</div>

								<!-- Actions -->
								<div class="flex space-x-2">
									<Button size="sm" variant="outline" class="flex-1">Connect</Button>
									<Button size="sm" variant="outline">
										<Ellipsis class="w-4 h-4" />
									</Button>
								</div>
							</div>
						</CardContent>
					</Card>
				{/each}
			</div>
		</div>
	</div>
</div>
