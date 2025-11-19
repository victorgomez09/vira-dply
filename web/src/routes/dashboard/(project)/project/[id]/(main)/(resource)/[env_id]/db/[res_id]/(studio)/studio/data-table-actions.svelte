<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { EllipsisVertical } from 'lucide-svelte';
	import type { TableRow } from './columns';

	interface Props {
		row: TableRow;
		onEdit: (row: TableRow) => void;
		onDelete: (row: TableRow) => void;
	}

	let { row, onEdit, onDelete }: Props = $props();
</script>

<DropdownMenu.Root>
	<DropdownMenu.Trigger>
		<Button variant="ghost" size="icon" class="h-8 w-8 p-0">
			{#snippet children()}
				<span class="sr-only">Open menu</span>
				<EllipsisVertical class="h-4 w-4" />
			{/snippet}
		</Button>
	</DropdownMenu.Trigger>
	<DropdownMenu.Content align="end">
		<DropdownMenu.Item onclick={() => onEdit(row)}>Edit</DropdownMenu.Item>
		<DropdownMenu.Item onclick={() => onDelete(row)} class="text-destructive">
			Delete
		</DropdownMenu.Item>
	</DropdownMenu.Content>
</DropdownMenu.Root>
