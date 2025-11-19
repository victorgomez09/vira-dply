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
	import type { PostgreSQLConfig } from '$lib/api/databases';

	interface Props {
		config: PostgreSQLConfig;
		onConfigChange: (config: PostgreSQLConfig) => void;
	}

	let { config = $bindable(), onConfigChange }: Props = $props();

	const versions = ['17', '16', '15', '14', '13'];

	function updateConfig(updates: Partial<PostgreSQLConfig>) {
		const newConfig = { ...config, ...updates };
		config = newConfig;
		onConfigChange(newConfig);
	}

	const version = versions.map((v) =>
		v === config.version ? config.version : 'Select version'
	)[0];
</script>

<div class="space-y-4">
	<div class="space-y-2">
		<Label for="version">Version *</Label>
		<Select type="single" bind:value={config.version}>
			<SelectTrigger>
				{version}
			</SelectTrigger>
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
		placeholder="postgres_user"
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
		<Label for="port">Port *</Label>
		<Input
			id="port"
			type="number"
			value={config.port || 5432}
			required
			onchange={(e) => updateConfig({ port: parseInt(e.currentTarget.value) || 5432 })}
		/>
	</div>
</div>
