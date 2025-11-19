<script>
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import {
		ArrowLeft,
		Plus,
		Globe,
		Shield,
		ExternalLink,
		Copy,
		Settings,
		Trash2,
		CheckCircle,
		AlertCircle,
		Clock
	} from 'lucide-svelte';

	const projectId = page.params.id;

	let project = $state({
		name: 'Focalpoint Dashboard',
		workspace: 'Focalpoint',
		category: 'Applications'
	});

	let showAddDomainModal = $state(false);
	let showSSLModal = $state(false);
	let selectedDomain = $state(null);

	let newDomain = $state({
		domain: '',
		subdomain: '',
		redirect: false,
		redirectUrl: ''
	});

	// Mock domains data
	let domains = $state([
		{
			id: 1,
			domain: 'app.focalpoint.com',
			type: 'primary',
			status: 'active',
			ssl: 'valid',
			sslExpiry: '2024-12-15',
			redirect: false,
			redirectUrl: null,
			verified: true,
			createdAt: '2024-01-15'
		},
		{
			id: 2,
			domain: 'focalpoint.app',
			type: 'custom',
			status: 'active',
			ssl: 'valid',
			sslExpiry: '2024-11-20',
			redirect: true,
			redirectUrl: 'https://app.focalpoint.com',
			verified: true,
			createdAt: '2024-01-20'
		},
		{
			id: 3,
			domain: 'staging.focalpoint.com',
			type: 'subdomain',
			status: 'pending',
			ssl: 'pending',
			sslExpiry: null,
			redirect: false,
			redirectUrl: null,
			verified: false,
			createdAt: '2024-01-30'
		}
	]);

	function getStatusColor(status) {
		switch (status) {
			case 'active':
				return 'text-green-600 bg-green-50';
			case 'pending':
				return 'text-yellow-600 bg-yellow-50';
			case 'error':
				return 'text-red-600 bg-red-50';
			default:
				return 'text-gray-600 bg-gray-50';
		}
	}

	function getSSLIcon(ssl) {
		switch (ssl) {
			case 'valid':
				return CheckCircle;
			case 'expiring':
				return AlertCircle;
			case 'expired':
				return AlertCircle;
			case 'pending':
				return Clock;
			default:
				return AlertCircle;
		}
	}

	function getSSLColor(ssl) {
		switch (ssl) {
			case 'valid':
				return 'text-green-500';
			case 'expiring':
				return 'text-yellow-500';
			case 'expired':
				return 'text-red-500';
			case 'pending':
				return 'text-blue-500';
			default:
				return 'text-gray-500';
		}
	}

	function addDomain() {
		if (newDomain.domain) {
			const newId = Math.max(...domains.map((d) => d.id)) + 1;
			domains = [
				...domains,
				{
					id: newId,
					domain: newDomain.subdomain
						? `${newDomain.subdomain}.${newDomain.domain}`
						: newDomain.domain,
					type: newDomain.subdomain ? 'subdomain' : 'custom',
					status: 'pending',
					ssl: 'pending',
					sslExpiry: null,
					redirect: newDomain.redirect,
					redirectUrl: newDomain.redirectUrl || null,
					verified: false,
					createdAt: new Date().toISOString().split('T')[0]
				}
			];
			newDomain = { domain: '', subdomain: '', redirect: false, redirectUrl: '' };
			showAddDomainModal = false;
		}
	}

	function deleteDomain(domainId) {
		domains = domains.filter((d) => d.id !== domainId);
	}

	function copyDomain(domain) {
		navigator.clipboard.writeText(domain);
	}

	function renewSSL(domainId) {
		const domain = domains.find((d) => d.id === domainId);
		if (domain) {
			domain.ssl = 'pending';
			setTimeout(() => {
				domain.ssl = 'valid';
				domain.sslExpiry = '2025-01-31';
			}, 2000);
		}
	}

	function openSSLModal(domain) {
		selectedDomain = domain;
		showSSLModal = true;
	}
</script>

<svelte:head>
	<title>Domains - {project.name}</title>
</svelte:head>

<div class="flex-1 p-6">
	<!-- Header -->
	<div class="flex items-center justify-between mb-6">
		<div class="flex items-center space-x-4">
			<div>
				<h1 class="text-2xl font-semibold text-gray-900">Domains</h1>
				<p class="text-sm text-gray-500 mt-1">
					Manage custom domains and SSL certificates for {project.name}.
				</p>
			</div>
		</div>
		<Button onclick={() => (showAddDomainModal = true)}>
			<Plus class="w-4 h-4 mr-2" />
			Add Domain
		</Button>
	</div>

	<!-- Domains List -->
	<div class="space-y-4">
		{#each domains as domain (domain.id)}
			<Card>
				<CardContent class="p-6">
					<div class="flex items-center justify-between">
						<div class="flex-1">
							<div class="flex items-center space-x-3 mb-2">
								<Globe class="w-5 h-5 text-gray-600" />
								<h3 class="font-medium text-gray-900">{domain.domain}</h3>
								<Badge variant="outline" class="text-xs {getStatusColor(domain.status)}">
									{domain.status}
								</Badge>
								{#if domain.type === 'primary'}
									<Badge variant="default" class="text-xs">Primary</Badge>
								{/if}
							</div>

							<div class="flex items-center space-x-6 text-sm text-gray-600">
								<div class="flex items-center space-x-2">
									{#if domain.ssl === 'valid'}
										<CheckCircle class="w-4 h-4 text-green-500" />
									{:else if domain.ssl === 'pending'}
										<Clock class="w-4 h-4 text-blue-500" />
									{:else}
										<AlertCircle class="w-4 h-4 text-red-500" />
									{/if}
									<span>SSL: {domain.ssl}</span>
									{#if domain.sslExpiry}
										<span>• Expires {domain.sslExpiry}</span>
									{/if}
								</div>
								{#if domain.redirect}
									<div class="flex items-center space-x-1">
										<ExternalLink class="w-4 h-4" />
										<span>Redirects to {domain.redirectUrl}</span>
									</div>
								{/if}
							</div>
						</div>

						<div class="flex items-center space-x-2">
							<Button size="sm" variant="outline" onclick={() => copyDomain(domain.domain)}>
								<Copy class="w-4 h-4" />
							</Button>
							<Button
								size="sm"
								variant="outline"
								onclick={() => window.open(`https://${domain.domain}`, '_blank')}
							>
								<ExternalLink class="w-4 h-4" />
							</Button>
							<Button size="sm" variant="outline" onclick={() => openSSLModal(domain)}>
								<Shield class="w-4 h-4" />
							</Button>
							<Button size="sm" variant="outline">
								<Settings class="w-4 h-4" />
							</Button>
							{#if domain.type !== 'primary'}
								<Button size="sm" variant="outline" onclick={() => deleteDomain(domain.id)}>
									<Trash2 class="w-4 h-4" />
								</Button>
							{/if}
						</div>
					</div>
				</CardContent>
			</Card>
		{/each}
	</div>
</div>

<!-- Add Domain Modal -->
{#if showAddDomainModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
		<Card class="w-full max-w-md">
			<CardHeader>
				<CardTitle>Add Custom Domain</CardTitle>
			</CardHeader>
			<CardContent class="space-y-4">
				<div>
					<Label for="domain">Domain</Label>
					<Input id="domain" bind:value={newDomain.domain} placeholder="example.com" />
				</div>
				<div>
					<Label for="subdomain">Subdomain (optional)</Label>
					<Input id="subdomain" bind:value={newDomain.subdomain} placeholder="www" />
				</div>
				<div class="flex items-center space-x-2">
					<input type="checkbox" id="redirect" bind:checked={newDomain.redirect} class="rounded" />
					<Label for="redirect">Redirect to another URL</Label>
				</div>
				{#if newDomain.redirect}
					<div>
						<Label for="redirectUrl">Redirect URL</Label>
						<Input
							id="redirectUrl"
							bind:value={newDomain.redirectUrl}
							placeholder="https://example.com"
						/>
					</div>
				{/if}
				<div class="flex space-x-2 pt-4">
					<Button onclick={addDomain} class="flex-1">Add Domain</Button>
					<Button variant="outline" onclick={() => (showAddDomainModal = false)} class="flex-1"
						>Cancel</Button
					>
				</div>
			</CardContent>
		</Card>
	</div>
{/if}

<!-- SSL Modal -->
{#if showSSLModal && selectedDomain}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
		<Card class="w-full max-w-md">
			<CardHeader>
				<CardTitle>SSL Certificate - {selectedDomain.domain}</CardTitle>
			</CardHeader>
			<CardContent class="space-y-4">
				<div class="flex items-center space-x-3">
					{#if selectedDomain.ssl === 'valid'}
						<CheckCircle class="w-6 h-6 text-green-500" />
						<div>
							<p class="font-medium text-green-600">Certificate Valid</p>
							<p class="text-sm text-gray-600">Expires on {selectedDomain.sslExpiry}</p>
						</div>
					{:else if selectedDomain.ssl === 'pending'}
						<Clock class="w-6 h-6 text-blue-500" />
						<div>
							<p class="font-medium text-blue-600">Certificate Pending</p>
							<p class="text-sm text-gray-600">SSL certificate is being issued</p>
						</div>
					{:else}
						<AlertCircle class="w-6 h-6 text-red-500" />
						<div>
							<p class="font-medium text-red-600">Certificate Issue</p>
							<p class="text-sm text-gray-600">SSL certificate needs attention</p>
						</div>
					{/if}
				</div>

				<div class="bg-gray-50 rounded-lg p-4">
					<h4 class="font-medium text-gray-900 mb-2">Certificate Details</h4>
					<div class="space-y-1 text-sm text-gray-600">
						<p>Issuer: Let's Encrypt</p>
						<p>Type: Domain Validated (DV)</p>
						<p>Encryption: RSA 2048-bit</p>
					</div>
				</div>

				<div class="flex space-x-2 pt-4">
					{#if selectedDomain.ssl !== 'pending'}
						<Button onclick={() => renewSSL(selectedDomain.id)} class="flex-1"
							>Renew Certificate</Button
						>
					{/if}
					<Button variant="outline" onclick={() => (showSSLModal = false)} class="flex-1"
						>Close</Button
					>
				</div>
			</CardContent>
		</Card>
	</div>
{/if}
