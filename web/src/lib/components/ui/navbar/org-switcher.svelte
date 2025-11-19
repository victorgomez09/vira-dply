<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { organizationsApi } from '$lib/api';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { Building2, Check, ChevronDown } from 'lucide-svelte';
	import type { Snippet } from 'svelte';

	interface Props {
		currentOrgId?: string;
		children?: Snippet;
	}

	let { currentOrgId, children }: Props = $props();

	const orgsQuery = createQuery(() => ({
		queryKey: ['organizations'],
		queryFn: () => organizationsApi.list()
	}));

	const currentOrg = $derived(
		orgsQuery.data?.find((org) => org.id === currentOrgId) || orgsQuery.data?.[0]
	);
</script>

<DropdownMenu.Root>
	<DropdownMenu.Trigger class="flex items-center gap-1.5 hover:text-foreground transition-colors">
		{#if children}
			{@render children()}
		{:else if currentOrg}
			<Building2 class="w-4 h-4" />
			{currentOrg.name}
			<ChevronDown class="w-3 h-3" />
		{:else}
			<Building2 class="w-4 h-4" />
			Organization
			<ChevronDown class="w-3 h-3" />
		{/if}
	</DropdownMenu.Trigger>
	<DropdownMenu.Content align="start">
		<DropdownMenu.Label>Switch Organization</DropdownMenu.Label>
		<DropdownMenu.Separator />
		{#if orgsQuery.isPending}
			<DropdownMenu.Item disabled>Loading...</DropdownMenu.Item>
		{:else if orgsQuery.isError}
			<DropdownMenu.Item disabled>Failed to load organizations</DropdownMenu.Item>
		{:else if orgsQuery.data}
			{#each orgsQuery.data as org (org.id)}
				<DropdownMenu.Item
					class="flex items-center justify-between"
					onclick={() => {
						window.location.href = `/dashboard?org=${org.id}`;
					}}
				>
					<span class="flex items-center gap-2">
						<Building2 class="w-4 h-4" />
						{org.name}
					</span>
					{#if org.id === currentOrg?.id}
						<Check class="w-4 h-4" />
					{/if}
				</DropdownMenu.Item>
			{/each}
		{/if}
	</DropdownMenu.Content>
</DropdownMenu.Root>
