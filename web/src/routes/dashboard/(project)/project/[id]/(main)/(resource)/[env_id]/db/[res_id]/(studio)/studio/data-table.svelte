<script lang="ts" generics="TData, TValue">
	import { getCoreRowModel } from '@tanstack/table-core';
	import type { ColumnDef } from '@tanstack/table-core';
	import { createSvelteTable } from '$lib/components/ui/data-table/data-table.svelte';
	import FlexRender from '$lib/components/ui/data-table/flex-render.svelte';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { ChevronLeft, ChevronRight } from 'lucide-svelte';

	interface Props {
		data: TData[];
		columns: ColumnDef<TData, TValue>[];
		currentPage: number;
		pageSize: number;
		totalItems: number;
		onPageChange: (page: number) => void;
	}

	let { data, columns, currentPage, pageSize, totalItems, onPageChange }: Props = $props();

	const totalPages = $derived(Math.ceil(totalItems / pageSize));

	const table = createSvelteTable({
		data,
		columns,
		getCoreRowModel: getCoreRowModel(),
		manualPagination: true,
		pageCount: totalPages
	});
</script>

<div class="space-y-4">
	<div class="rounded-md border">
		<Table.Root>
			<Table.Header>
				{#each table.getHeaderGroups() as headerGroup}
					<Table.Row>
						{#each headerGroup.headers as header}
							<Table.Head>
								{#if !header.isPlaceholder}
									<FlexRender content={header.column.columnDef.header} context={header.getContext()} />
								{/if}
							</Table.Head>
						{/each}
					</Table.Row>
				{/each}
			</Table.Header>
			<Table.Body>
				{#if table.getRowModel().rows?.length}
					{#each table.getRowModel().rows as row}
						<Table.Row data-state={row.getIsSelected() && 'selected'}>
							{#each row.getVisibleCells() as cell}
								<Table.Cell>
									<FlexRender content={cell.column.columnDef.cell} context={cell.getContext()} />
								</Table.Cell>
							{/each}
						</Table.Row>
					{/each}
				{:else}
					<Table.Row>
						<Table.Cell colspan={columns.length} class="h-24 text-center">
							No results.
						</Table.Cell>
					</Table.Row>
				{/if}
			</Table.Body>
		</Table.Root>
	</div>

	<div class="flex items-center justify-between">
		<div class="text-sm text-muted-foreground">
			Showing {(currentPage - 1) * pageSize + 1} to {Math.min(currentPage * pageSize, totalItems)}
			of {totalItems} rows
		</div>
		<div class="flex items-center space-x-2">
			<Button
				variant="outline"
				size="sm"
				disabled={currentPage <= 1}
				onclick={() => onPageChange(currentPage - 1)}
			>
				<ChevronLeft class="h-4 w-4" />
				Previous
			</Button>
			<div class="text-sm">
				Page {currentPage} of {totalPages}
			</div>
			<Button
				variant="outline"
				size="sm"
				disabled={currentPage >= totalPages}
				onclick={() => onPageChange(currentPage + 1)}
			>
				Next
				<ChevronRight class="h-4 w-4" />
			</Button>
		</div>
	</div>
</div>
