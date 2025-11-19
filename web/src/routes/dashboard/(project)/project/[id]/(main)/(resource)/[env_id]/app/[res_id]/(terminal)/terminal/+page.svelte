<script>
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import Terminal from '$lib/components/Terminal.svelte';
	import { Terminal as TerminalIcon, Maximize2, Minimize2 } from 'lucide-svelte';

	let projectId = $derived(page.params.id);
	let envId = $derived(page.params.env_id);
	let resId = $derived(page.params.res_id);

	let isConnected = $state(false);
	let isFullscreen = $state(false);
	let connectionError = $state(null);

	const endpoint = $derived(`/api/projects/${projectId}/applications/${resId}/terminal`);

	function handleConnect() {
		isConnected = true;
		connectionError = null;
	}

	function handleDisconnect() {
		isConnected = false;
	}

	function handleError(error) {
		isConnected = false;
		connectionError = error.message;
	}

	function toggleFullscreen() {
		isFullscreen = !isFullscreen;
	}
</script>

<svelte:head>
	<title>Application Terminal</title>
</svelte:head>

<div class="flex-1 p-6">
	<div class="flex items-center justify-between mb-6">
		<div class="flex items-center space-x-4">
			<div>
				<h1 class="text-2xl font-semibold text-gray-900">Application Terminal</h1>
				<p class="text-sm text-gray-500 mt-1">
					Access your application container directly through the web terminal.
				</p>
			</div>
		</div>
		<div class="flex items-center space-x-3">
			<Button variant="outline" onclick={toggleFullscreen}>
				{#if isFullscreen}
					<Minimize2 class="w-4 h-4" />
				{:else}
					<Maximize2 class="w-4 h-4" />
				{/if}
			</Button>
		</div>
	</div>

	{#if connectionError}
		<div class="mb-4 p-4 bg-red-50 border border-red-200 rounded-md">
			<p class="text-sm text-red-600">Connection error: {connectionError}</p>
		</div>
	{/if}

	<Card class={isFullscreen ? 'fixed inset-4 z-50' : 'h-[calc(100vh-300px)]'}>
		<CardHeader class="pb-2">
			<div class="flex items-center justify-between">
				<div class="flex items-center space-x-2">
					<TerminalIcon class="w-5 h-5 text-gray-600" />
					<CardTitle class="text-lg">Application Terminal</CardTitle>
					<div class="flex items-center space-x-1">
						<div class="w-2 h-2 rounded-full {isConnected ? 'bg-green-500' : 'bg-red-500'}"></div>
						<span class="text-sm text-gray-600">{isConnected ? 'Connected' : 'Connecting...'}</span>
					</div>
				</div>
			</div>
		</CardHeader>
		<CardContent class="p-0 h-[calc(100%-4rem)]">
			<Terminal
				containerID={resId}
				projectID={projectId}
				{endpoint}
				onConnect={handleConnect}
				onDisconnect={handleDisconnect}
				onError={handleError}
			/>
		</CardContent>
	</Card>
</div>
