<script lang="ts">
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Separator } from '$lib/components/ui/separator';
	import { Switch } from '$lib/components/ui/switch';
	import { createQuery } from '@tanstack/svelte-query';
	import { databasesApi, type Database } from '$lib/api/databases';
	import { Copy, Eye, EyeOff, AlertCircle, Play } from 'lucide-svelte';

	const projectId = $derived(page.params.id);
	const envId = $derived(page.params.env_id);
	const resId = $derived(page.params.res_id);

	let showPassword = $state(false);
	let showConnectionString = $state(false);
	let enableSSL = $state(false);
	let publicAccess = $state(false);
	let proxyPort = $state('');

	const databaseQuery = createQuery(() => ({
		queryKey: ['database', projectId, resId],
		queryFn: () => databasesApi.get(projectId, resId),
		enabled: !!projectId && !!resId
	}));

	const database = $derived(databaseQuery.data);

	function copyToClipboard(text: string, label: string) {
		navigator.clipboard.writeText(text);
	}

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function getConfigForType(db: Database) {
		const type = db.config.type;
		return db.config[type];
	}

	function getInternalHostname(db: Database): string {
		return `${db.name}.${projectId}.${envId}.internal`;
	}

	function getDatabaseDisplayName(type: string): string {
		const names: Record<string, string> = {
			postgresql: 'PostgreSQL',
			mysql: 'MySQL',
			mariadb: 'MariaDB',
			redis: 'Redis',
			keydb: 'KeyDB',
			dragonfly: 'Dragonfly',
			mongodb: 'MongoDB',
			clickhouse: 'ClickHouse'
		};
		return names[type] || type;
	}
</script>

{#if database}
	<div class="space-y-6">
		<Card>
			<CardHeader>
				<div class="flex items-center justify-between">
					<CardTitle>Details</CardTitle>
					{#if database.status === 'stopped'}
						<Button size="sm">
							<Play class="mr-2 h-4 w-4" />
							Start Database
						</Button>
					{/if}
				</div>
			</CardHeader>
			<CardContent>
				<div class="grid grid-cols-4 gap-6">
					<div>
						<Label class="text-muted-foreground">Creation date</Label>
						<p class="font-medium">{formatDate(database.created_at)}</p>
					</div>
					<div>
						<Label class="text-muted-foreground">Version</Label>
						<p class="font-medium">
							{getDatabaseDisplayName(database.type)}
							{getConfigForType(database)?.version || 'N/A'}
						</p>
					</div>
					<div>
						<Label class="text-muted-foreground">Location</Label>
						<p class="font-medium">us-central1</p>
					</div>
					<div>
						<Label class="text-muted-foreground">Database size</Label>
						<p class="font-medium">-</p>
					</div>
				</div>

				<Separator class="my-4" />

				<div>
					<Label class="text-muted-foreground">Resources</Label>
					<p class="font-medium">-</p>
				</div>
			</CardContent>
		</Card>

		<Card>
			<CardHeader>
				<div class="flex items-center gap-2">
					<CardTitle>Internal connections</CardTitle>
					<AlertCircle class="h-4 w-4 text-muted-foreground" />
				</div>
				<p class="text-sm text-muted-foreground">
					Internal connections between your databases and apps are lightning-fast and secure.
				</p>
			</CardHeader>
			<CardContent class="space-y-6">
				<div class="grid grid-cols-2 gap-6">
					<div>
						<Label class="text-muted-foreground mb-2">Hostname</Label>
						<div class="flex gap-2">
							<Input value={getInternalHostname(database)} readonly class="font-mono text-sm" />
							<Button
								variant="outline"
								size="icon"
								onclick={() => copyToClipboard(getInternalHostname(database), 'Hostname')}
							>
								<Copy class="h-4 w-4" />
							</Button>
						</div>
					</div>
					<div>
						<Label class="text-muted-foreground mb-2">Port</Label>
						<div class="flex gap-2">
							<Input
								value={getConfigForType(database)?.port || '-'}
								readonly
								class="font-mono text-sm"
							/>
							<Button
								variant="outline"
								size="icon"
								onclick={() =>
									copyToClipboard(String(getConfigForType(database)?.port || ''), 'Port')}
							>
								<Copy class="h-4 w-4" />
							</Button>
						</div>
					</div>
				</div>

				<Separator />

				<div class="grid grid-cols-2 gap-6">
					<div>
						<Label class="text-muted-foreground mb-2">Username</Label>
						<div class="flex gap-2">
							<Input
								value={getConfigForType(database)?.username || '-'}
								readonly
								class="font-mono text-sm"
							/>
							<Button
								variant="outline"
								size="icon"
								onclick={() =>
									copyToClipboard(getConfigForType(database)?.username || '', 'Username')}
							>
								<Copy class="h-4 w-4" />
							</Button>
						</div>
					</div>
					<div>
						<Label class="text-muted-foreground mb-2">Password</Label>
						<div class="flex gap-2">
							<Input
								value={getConfigForType(database)?.password || '-'}
								type={showPassword ? 'text' : 'password'}
								readonly
								class="font-mono text-sm"
							/>
							<Button variant="outline" size="icon" onclick={() => (showPassword = !showPassword)}>
								{#if showPassword}
									<EyeOff class="h-4 w-4" />
								{:else}
									<Eye class="h-4 w-4" />
								{/if}
							</Button>
							<Button
								variant="outline"
								size="icon"
								onclick={() =>
									copyToClipboard(getConfigForType(database)?.password || '', 'Password')}
							>
								<Copy class="h-4 w-4" />
							</Button>
						</div>
					</div>
				</div>

				<Separator />

				<div>
					<Label class="text-muted-foreground mb-2">Database name</Label>
					<div class="flex gap-2">
						<Input
							value={getConfigForType(database)?.database_name || '-'}
							readonly
							class="font-mono text-sm"
						/>
						<Button
							variant="outline"
							size="icon"
							onclick={() =>
								copyToClipboard(getConfigForType(database)?.database_name || '', 'Database name')}
						>
							<Copy class="h-4 w-4" />
						</Button>
					</div>
				</div>

				<Separator />

				<div>
					<Label class="text-muted-foreground mb-2">Connection string</Label>
					<div class="flex gap-2">
						<Input
							value={database.connection_string || '-'}
							type={showConnectionString ? 'text' : 'password'}
							readonly
							class="font-mono text-sm"
						/>
						<Button
							variant="outline"
							size="icon"
							onclick={() => (showConnectionString = !showConnectionString)}
						>
							{#if showConnectionString}
								<EyeOff class="h-4 w-4" />
							{:else}
								<Eye class="h-4 w-4" />
							{/if}
						</Button>
						<Button
							variant="outline"
							size="icon"
							onclick={() => copyToClipboard(database.connection_string || '', 'Connection string')}
						>
							<Copy class="h-4 w-4" />
						</Button>
					</div>
				</div>
			</CardContent>
		</Card>

		<Card>
			<CardHeader>
				<CardTitle>Network Options</CardTitle>
				<p class="text-sm text-muted-foreground">
					Configure external access and network settings for your database.
				</p>
			</CardHeader>
			<CardContent class="space-y-6">
				<div class="flex items-center justify-between">
					<div class="space-y-0.5">
						<Label>Enable SSL</Label>
						<p class="text-sm text-muted-foreground">Use encrypted SSL connections</p>
					</div>
					<Switch bind:checked={enableSSL} />
				</div>

				<Separator />

				<div class="flex items-center justify-between">
					<div class="space-y-0.5">
						<Label>Public Access</Label>
						<p class="text-sm text-muted-foreground">
							Allow connections from outside your project network
						</p>
					</div>
					<Switch bind:checked={publicAccess} />
				</div>

				{#if publicAccess}
					<div>
						<Label for="proxy-port" class="text-muted-foreground mb-2">Proxy Port</Label>
						<Input
							id="proxy-port"
							type="number"
							bind:value={proxyPort}
							placeholder="e.g., 5432"
							class="max-w-xs"
						/>
						<p class="text-sm text-muted-foreground mt-2">
							External port for accessing your database through the proxy
						</p>
					</div>
				{/if}

				<Separator />

				<div>
					<Label class="text-muted-foreground mb-2">Port Mapping</Label>
					<div class="space-y-2">
						<div class="flex gap-2 items-center text-sm">
							<span class="font-mono bg-muted px-2 py-1 rounded">
								Internal: {getConfigForType(database)?.port || '-'}
							</span>
							{#if publicAccess && proxyPort}
								<span class="text-muted-foreground">→</span>
								<span class="font-mono bg-muted px-2 py-1 rounded">External: {proxyPort}</span>
							{:else}
								<span class="text-muted-foreground">(Internal only)</span>
							{/if}
						</div>
					</div>
				</div>

				<div class="pt-4">
					<Button>Save Network Settings</Button>
				</div>
			</CardContent>
		</Card>
	</div>
{/if}
