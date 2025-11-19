<script>
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import { Globe, Plus, Trash2, Shuffle, Loader2, AlertTriangle } from 'lucide-svelte';
	import { applicationsApi } from '$lib/api/applications';
	import { toast } from 'svelte-sonner';

	const projectId = page.params.id;
	const envId = page.params.env_id;
	const appId = page.params.res_id;

	let app = $state(null);
	let loading = $state(true);
	let savingDomain = $state(false);
	let generatingDomain = $state(false);
	let savingPorts = $state(false);

	let domainInput = $state('');
	let exposedPorts = $state([]);
	let portMappings = $state([]);

	$effect(() => {
		loadApp();
	});

	async function loadApp() {
		try {
			loading = true;
			const data = await applicationsApi.get(projectId, appId);
			app = data;
			domainInput = data.custom_domain || '';
			exposedPorts = data.exposed_ports ? [...data.exposed_ports] : [];
			portMappings = data.port_mappings ? [...data.port_mappings] : [];
		} catch (error) {
			toast.error('Failed to load application');
			console.error(error);
		} finally {
			loading = false;
		}
	}

	async function saveDomain() {
		try {
			savingDomain = true;
			await applicationsApi.assignDomain(projectId, appId, domainInput || null);
			toast.success('Domain updated');
			await loadApp();
		} catch (error) {
			toast.error('Failed to update domain');
			console.error(error);
		} finally {
			savingDomain = false;
		}
	}

	async function generateDomain() {
		try {
			generatingDomain = true;
			const result = await applicationsApi.generateDomain(projectId, appId);
			domainInput = result.domain;
			toast.success('Domain generated');
			await loadApp();
		} catch (error) {
			toast.error('Failed to generate domain');
			console.error(error);
		} finally {
			generatingDomain = false;
		}
	}

	function addExposedPort() {
		exposedPorts = [...exposedPorts, 8080];
	}

	function removeExposedPort(index) {
		exposedPorts = exposedPorts.filter((_, i) => i !== index);
	}

	function addPortMapping() {
		portMappings = [...portMappings, { host: 8080, container: 8080 }];
	}

	function removePortMapping(index) {
		portMappings = portMappings.filter((_, i) => i !== index);
	}

	async function savePorts() {
		try {
			savingPorts = true;
			const portMappingsFormatted = portMappings.map((m) => ({
				host_port: m.host,
				container_port: m.container
			}));
			await applicationsApi.updatePorts(projectId, appId, {
				exposed_ports: exposedPorts,
				port_mappings: portMappingsFormatted
			});
			toast.success('Ports updated');
			await loadApp();
		} catch (error) {
			toast.error('Failed to update ports');
			console.error(error);
		} finally {
			savingPorts = false;
		}
	}
</script>

<svelte:head>
	<title>Networking - {app?.name || 'Application'}</title>
</svelte:head>

<div class="flex-1 p-6">
	{#if loading}
		<div class="flex items-center justify-center h-64">
			<Loader2 class="w-8 h-8 animate-spin text-gray-400" />
		</div>
	{:else if app}
		<div class="flex items-center justify-between mb-6">
			<div>
				<h1 class="text-2xl font-semibold text-gray-900">Networking</h1>
				<p class="text-sm text-gray-500 mt-1">
					Configure domain and port settings for {app.name}.
				</p>
			</div>
		</div>

		<Alert variant="default" class="mb-6 border-amber-200 bg-amber-50">
			<AlertTriangle class="text-amber-600" />
			<AlertDescription class="text-amber-800">
				Changes to domain or port settings require a redeployment to take effect. The application must be recreated for Traefik routing to update.
			</AlertDescription>
		</Alert>

		<div class="space-y-6">
			<Card>
				<CardHeader>
					<CardTitle>Domain Configuration</CardTitle>
				</CardHeader>
				<CardContent class="space-y-4">
					<div>
						<Label for="domain">Custom Domain</Label>
						<div class="flex space-x-2 mt-1">
							<Input
								id="domain"
								bind:value={domainInput}
								placeholder="app.example.com"
								class="flex-1"
							/>
							<Button onclick={saveDomain} disabled={savingDomain}>
								{#if savingDomain}
									<Loader2 class="w-4 h-4 mr-2 animate-spin" />
								{/if}
								Save
							</Button>
						</div>
						<p class="text-sm text-gray-500 mt-1">
							Assign a custom domain for this application. Leave empty to use generated domain.
						</p>
					</div>

					<div>
						<Label>Generated Domain</Label>
						<div class="flex space-x-2 mt-1">
							<Input
								value={app.generated_domain || 'Not generated'}
								disabled
								class="flex-1 bg-gray-50"
							/>
							<Button onclick={generateDomain} disabled={generatingDomain} variant="outline">
								{#if generatingDomain}
									<Loader2 class="w-4 h-4 mr-2 animate-spin" />
								{:else}
									<Shuffle class="w-4 h-4 mr-2" />
								{/if}
								Generate
							</Button>
						</div>
						<p class="text-sm text-gray-500 mt-1">
							Automatically generate a random sslip.io domain.
						</p>
					</div>

					<div class="pt-2 border-t">
						<div class="flex items-center space-x-2 text-sm">
							<Globe class="w-4 h-4 text-blue-500" />
							<span class="font-medium">Active Domain:</span>
							<span class="text-gray-600">
								{app.custom_domain || app.generated_domain || 'No domain configured'}
							</span>
						</div>
					</div>
				</CardContent>
			</Card>

			<Card>
				<CardHeader>
					<div class="flex items-center justify-between">
						<CardTitle>Exposed Ports</CardTitle>
						<Button size="sm" onclick={addExposedPort}>
							<Plus class="w-4 h-4 mr-2" />
							Add Port
						</Button>
					</div>
				</CardHeader>
				<CardContent class="space-y-4">
					{#if exposedPorts.length === 0}
						<p class="text-sm text-gray-500">
							No exposed ports configured. Using default port 8080.
						</p>
					{:else}
						<div class="space-y-2">
							{#each exposedPorts as port, index (index)}
								<div class="flex items-center space-x-2">
									<Input
										type="number"
										bind:value={exposedPorts[index]}
										placeholder="8080"
										class="flex-1"
									/>
									<Button size="sm" variant="outline" onclick={() => removeExposedPort(index)}>
										<Trash2 class="w-4 h-4" />
									</Button>
								</div>
							{/each}
						</div>
					{/if}
					<div class="pt-4 border-t">
						<Button onclick={savePorts} disabled={savingPorts}>
							{#if savingPorts}
								<Loader2 class="w-4 h-4 mr-2 animate-spin" />
							{/if}
							Save Ports
						</Button>
					</div>
					<p class="text-sm text-gray-500">
						Ports that the container exposes for external access via Traefik.
					</p>
				</CardContent>
			</Card>

			<Card>
				<CardHeader>
					<div class="flex items-center justify-between">
						<CardTitle>Port Mappings</CardTitle>
						<Button size="sm" onclick={addPortMapping}>
							<Plus class="w-4 h-4 mr-2" />
							Add Mapping
						</Button>
					</div>
				</CardHeader>
				<CardContent class="space-y-4">
					{#if portMappings.length === 0}
						<p class="text-sm text-gray-500">No port mappings configured. Container will not expose ports to host.</p>
					{:else}
						<div class="space-y-2">
							{#each portMappings as mapping, index (index)}
								<div class="flex items-center space-x-2">
									<Input
										type="number"
										bind:value={portMappings[index].host}
										placeholder="Host port"
										class="flex-1"
									/>
									<span class="text-gray-500">:</span>
									<Input
										type="number"
										bind:value={portMappings[index].container}
										placeholder="Container port"
										class="flex-1"
									/>
									<Button size="sm" variant="outline" onclick={() => removePortMapping(index)}>
										<Trash2 class="w-4 h-4" />
									</Button>
								</div>
							{/each}
						</div>
					{/if}
					<div class="pt-4 border-t">
						<Button onclick={savePorts} disabled={savingPorts}>
							{#if savingPorts}
								<Loader2 class="w-4 h-4 mr-2 animate-spin" />
							{/if}
							Save Ports
						</Button>
					</div>
					<p class="text-sm text-gray-500">
						Map host ports to container ports (format: host:container).
					</p>
				</CardContent>
			</Card>
		</div>
	{/if}
</div>
