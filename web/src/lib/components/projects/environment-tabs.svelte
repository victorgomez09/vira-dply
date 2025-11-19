<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Plus } from 'lucide-svelte';
	import type { Environment } from '$lib/api';

	interface Props {
		environments: Environment[];
		selectedEnvironmentId?: string;
		onSelect: (environmentId: string | undefined) => void;
		onAdd: () => void;
		counts?: Record<string, number>;
	}

	let {
		environments,
		selectedEnvironmentId = $bindable(),
		onSelect,
		onAdd,
		counts = {}
	}: Props = $props();

	function handleTabClick(envId: string | undefined) {
		selectedEnvironmentId = envId;
		onSelect(envId);
	}
</script>

<div class="flex items-center gap-2">
	<div class="flex items-center gap-1 overflow-x-auto">
		{#each environments as env (env.id)}
			<Button
				variant="secondary"
				class=" bg-secondary-new px-4 py-2 text-sm font-medium transition-colors hover:text-foreground {selectedEnvironmentId ===
				env.id
					? 'border-b-2 border-secondary-foreground text-foreground'
					: 'text-muted-foreground'}"
				onclick={() => handleTabClick(env.id)}
			>
				{env.name}
				{#if counts[env.id] !== undefined}
					<Badge variant="secondary" class="ml-2 bg-secondary-foreground text-muted-foreground"
						>{counts[env.id]}</Badge
					>
				{/if}
			</Button>
		{/each}
	</div>

	<Button size="sm" variant="ghost" onclick={onAdd} class="ml-auto">
		<Plus class="h-4 w-4" />
		Add
	</Button>
</div>
