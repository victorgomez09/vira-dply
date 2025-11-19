<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Button } from '$lib/components/ui/button';
	import { RefreshCw, Copy, Eye, EyeOff } from 'lucide-svelte';

	interface Props {
		label: string;
		value: string | undefined;
		onValueChange: (value: string) => void;
		type?: 'text' | 'password';
		placeholder?: string;
		required?: boolean;
		generateFunction?: () => string;
	}

	let {
		label,
		value = $bindable(),
		onValueChange,
		type = 'text',
		placeholder = '',
		required = false,
		generateFunction
	}: Props = $props();

	let showPassword = $state(false);
	let copied = $state(false);

	function generateValue() {
		if (generateFunction) {
			const newValue = generateFunction();
			value = newValue;
			onValueChange(newValue);
		}
	}

	function copyToClipboard() {
		navigator.clipboard.writeText(value);
		copied = true;
		setTimeout(() => (copied = false), 2000);
	}

	function togglePasswordVisibility() {
		showPassword = !showPassword;
	}

	const inputType = $derived(type === 'password' && !showPassword ? 'password' : 'text');
</script>

<div class="space-y-2">
	<Label for={label}>{label}{required ? ' *' : ''}</Label>
	<div class="flex gap-2">
		<Input
			id={label}
			type={inputType}
			{placeholder}
			{required}
			bind:value
			onchange={(e) => onValueChange(e.currentTarget.value)}
			class="flex-1"
		/>
		<div class="flex gap-1">
			{#if type === 'password'}
				<Button type="button" variant="outline" size="icon" onclick={togglePasswordVisibility} title={showPassword ? 'Hide' : 'Show'}>
					{#if showPassword}
						<EyeOff class="h-4 w-4" />
					{:else}
						<Eye class="h-4 w-4" />
					{/if}
				</Button>
			{/if}
			{#if generateFunction}
				<Button type="button" variant="outline" size="icon" onclick={generateValue} title="Generate">
					<RefreshCw class="h-4 w-4" />
				</Button>
			{/if}
			<Button
				type="button"
				variant="outline"
				size="icon"
				onclick={copyToClipboard}
				title={copied ? 'Copied!' : 'Copy'}
				disabled={!value}
			>
				<Copy class="h-4 w-4" />
			</Button>
		</div>
	</div>
</div>
