<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import {
		Select,
		SelectContent,
		SelectItem,
		SelectTrigger
	} from '$lib/components/ui/select/index';
	import FieldWithGenerate from './field-with-generate.svelte';
	import { generatePassword, generateUsername, generateDatabaseName } from '$lib/utils/generators';
	import type { ClickHouseConfig } from '$lib/api/databases';

	interface Props {
		config: ClickHouseConfig;
		onConfigChange: (config: ClickHouseConfig) => void;
	}

	let { config = $bindable(), onConfigChange }: Props = $props();

	const versions = ['24.10', '24.9', '24.8', '23.8'];

	function updateConfig(updates: Partial<ClickHouseConfig>) {
		const newConfig = { ...config, ...updates };
		config = newConfig;
		onConfigChange(newConfig);
	}

	const triggerContent = $derived(versions.find((f) => f === config.version) ?? 'Select version');
</script>

<div class="space-y-4">
	<div class="space-y-2">
		<Label for="version">Version *</Label>
		<Select type="single" bind:value={config.version}>
			<SelectTrigger>{triggerContent}</SelectTrigger>
			<SelectContent>
				{#each versions as version}
					<SelectItem value={version}>{version}</SelectItem>
				{/each}
			</SelectContent>
		</Select>
	</div>

	<FieldWithGenerate
		label="Database Name"
		bind:value={config.database_name}
		onValueChange={(v) => updateConfig({ database_name: v })}
		placeholder="my_database"
		required={true}
		generateFunction={generateDatabaseName}
	/>

	<FieldWithGenerate
		label="Username"
		bind:value={config.username}
		onValueChange={(v) => updateConfig({ username: v })}
		placeholder="clickhouse_user"
		required={true}
		generateFunction={generateUsername}
	/>

	<FieldWithGenerate
		label="Password"
		bind:value={config.password}
		onValueChange={(v) => updateConfig({ password: v })}
		type="password"
		placeholder="Enter password"
		required={true}
		generateFunction={() => generatePassword(24)}
	/>

	<div class="space-y-2">
		<Label for="http_port">HTTP Port *</Label>
		<Input
			id="http_port"
			type="number"
			value={config.http_port || 8123}
			required
			onchange={(e) => updateConfig({ http_port: parseInt(e.currentTarget.value) || 8123 })}
		/>
	</div>

	<div class="space-y-2">
		<Label for="port">Native Port *</Label>
		<Input
			id="port"
			type="number"
			value={config.port || 9000}
			required
			onchange={(e) => updateConfig({ port: parseInt(e.currentTarget.value) || 9000 })}
		/>
	</div>
</div>
