import type { ColumnDef } from '@tanstack/table-core';
import { renderComponent } from '$lib/components/ui/data-table/index.js';
import DataTableActions from './data-table-actions.svelte';

export type TableRow = Record<string, any>;

export function createColumns(
	schema: { name: string; type: string; nullable: boolean }[],
	onEdit: (row: TableRow) => void,
	onDelete: (row: TableRow) => void
): ColumnDef<TableRow>[] {
	const dataColumns: ColumnDef<TableRow>[] = schema.map((col) => ({
		accessorKey: col.name,
		header: col.name,
		cell: ({ row }) => {
			const value = row.getValue(col.name);
			if (value === null) {
				return null;
			}
			if (typeof value === 'boolean') {
				return value;
			}
			if (typeof value === 'number') {
				return value;
			}
			return String(value);
		},
		meta: {
			type: col.type,
			nullable: col.nullable
		}
	}));

	dataColumns.push({
		id: 'actions',
		cell: ({ row }) => {
			return renderComponent(DataTableActions, {
				row: row.original,
				onEdit,
				onDelete
			});
		},
		enableSorting: false,
		enableHiding: false
	});

	return dataColumns;
}
