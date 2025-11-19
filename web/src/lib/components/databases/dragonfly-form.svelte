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
	import { generatePassword } from '$lib/utils/generators';
	import type { DragonflyConfig } from '$lib/api/databases';

	interface Props {
		config: DragonflyConfig;
		onConfigChange: (config: DragonflyConfig) => void;
	}

	let { config = $bindable(), onConfigChange }: Props = $props();

	const versions = ['1.23', '1.22', '1.21'];

	function updateConfig(updates: Partial<DragonflyConfig>) {
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
		label="Password"
		bind:value={config.password}
		onValueChange={(v) => updateConfig({ password: v })}
		type="password"
		placeholder="Enter password"
		required={false}
		generateFunction={() => generatePassword(24)}
	/>

	<div class="space-y-2">
		<Label for="port">Port *</Label>
		<Input
			id="port"
			type="number"
			value={(config.port || 6379).toString()}
			required
			onchange={(e) => updateConfig({ port: parseInt(e.currentTarget.value) || 6379 })}
		/>
	</div>
</div>
