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
	import type { MySQLConfig } from '$lib/api/databases';

	interface Props {
		config: MySQLConfig;
		onConfigChange: (config: MySQLConfig) => void;
	}

	let { config = $bindable(), onConfigChange }: Props = $props();

	const versions = ['8.4', '8.0', '5.7'];

	function updateConfig(updates: Partial<MySQLConfig>) {
		const newConfig = { ...config, ...updates };
		config = newConfig;
		onConfigChange(newConfig);
	}
</script>

<div class="space-y-4">
	<div class="space-y-2">
		<Label for="version">Version *</Label>
		<Select type="single" bind:value={config.version}>
			<SelectTrigger>
				{config.version || 'Select version'}
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
		placeholder="mysql_user"
		required={true}
		generateFunction={generateUsername}
	/>

	<FieldWithGenerate
		label="Root Password"
		bind:value={config.root_password}
		onValueChange={(v) => updateConfig({ root_password: v })}
		type="password"
		placeholder="Enter root password"
		required={true}
		generateFunction={() => generatePassword(24)}
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
			value={config.port || 3306}
			required
			onchange={(e) => updateConfig({ port: parseInt(e.currentTarget.value) || 3306 })}
		/>
	</div>
</div>
