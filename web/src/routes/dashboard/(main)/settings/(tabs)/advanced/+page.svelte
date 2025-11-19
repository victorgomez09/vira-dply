<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Switch } from '$lib/components/ui/switch';
	import { Textarea } from '$lib/components/ui/textarea';
	import * as Card from '$lib/components/ui/card';

	let dnsValidation = $state(true);
	let dnsServers = $state('8.8.8.8,1.1.1.1');
	let apiAccess = $state(false);
	let allowedIPs = $state('0.0.0.0');

	onMount(async () => {
		try {
			const response = await fetch('/api/settings/advanced');
			if (response.ok) {
				const data = await response.json();
				dnsValidation = data.dns_validation ?? true;
				dnsServers = data.dns_servers || '8.8.8.8,1.1.1.1';
				apiAccess = data.api_access ?? false;
				allowedIPs = data.allowed_ips || '0.0.0.0';
			}
		} catch (error) {
			console.error('Failed to load settings:', error);
		}
	});

	async function handleSave() {
		try {
			const response = await fetch('/api/settings/advanced', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					dns_validation: dnsValidation,
					dns_servers: dnsServers,
					api_access: apiAccess,
					allowed_ips: allowedIPs
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
			<Card.Title>DNS Configuration</Card.Title>
		</Card.Header>
		<Card.Content class="space-y-6">
			<div class="flex items-center justify-between">
				<div class="space-y-1">
					<Label for="dns-validation">DNS Validation</Label>
					<p class="text-xs text-muted-foreground">
						If you set a custom domain, it will be validated in your DNS provider.
					</p>
				</div>
				<Switch id="dns-validation" bind:checked={dnsValidation} />
			</div>

			<div class="space-y-2">
				<Label for="dns-servers">Custom DNS Servers</Label>
				<Input id="dns-servers" type="text" bind:value={dnsServers} placeholder="8.8.8.8,1.1.1.1" />
				<p class="text-xs text-muted-foreground">
					DNS servers to validate domains against. A comma separated list of DNS servers.
				</p>
			</div>
		</Card.Content>
	</Card.Root>

	<Card.Root>
		<Card.Header>
			<Card.Title>API Settings</Card.Title>
		</Card.Header>
		<Card.Content class="space-y-6">
			<div class="flex items-center justify-between">
				<div class="space-y-1">
					<Label for="api-access">API Access</Label>
					<p class="text-xs text-muted-foreground">
						If enabled, the API will be enabled. If disabled, the API will be disabled. This is the
						public facing API (not yet implemented).
					</p>
				</div>
				<Switch id="api-access" bind:checked={apiAccess} />
			</div>

			<div class="space-y-2">
				<Label for="allowed-ips">Allowed IP Addresses or Subnets</Label>
				<Textarea
					id="allowed-ips"
					bind:value={allowedIPs}
					placeholder="192.168.1.0/24,10.0.0.100"
					rows={4}
				/>
				<p class="text-xs text-muted-foreground">
					Supports single IPs (192.168.1.100) and CIDR notation (192.168.1.0/24). Use comma to
					separate multiple entries. Use 0.0.0.0 or leave empty to allow from anywhere.
				</p>
			</div>
		</Card.Content>
	</Card.Root>

	<div class="flex justify-end">
		<Button onclick={handleSave}>Save Changes</Button>
	</div>
</div>
