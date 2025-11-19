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
	import type { MongoDBConfig } from '$lib/api/databases';

	interface Props {
		config: MongoDBConfig;
		onConfigChange: (config: MongoDBConfig) => void;
	}

	let { config = $bindable(), onConfigChange }: Props = $props();

	const versions = ['8.0', '7.0', '6.0', '5.0'];

	function updateConfig(updates: Partial<MongoDBConfig>) {
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
		placeholder="mongodb_user"
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
			value={config.port || 27017}
			required
			onchange={(e) => updateConfig({ port: parseInt(e.currentTarget.value) || 27017 })}
		/>
	</div>
</div>
