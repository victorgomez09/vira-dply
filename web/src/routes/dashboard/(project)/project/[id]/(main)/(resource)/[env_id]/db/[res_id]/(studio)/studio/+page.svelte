<script lang="ts">
	import { page } from '$app/state';
	import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { studioApi, type QueryResult } from '$lib/api';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Label } from '$lib/components/ui/label';
	import { Input } from '$lib/components/ui/input';
	import {
		Sheet,
		SheetContent,
		SheetDescription,
		SheetFooter,
		SheetHeader,
		SheetTitle
	} from '$lib/components/ui/sheet';
	import { Select, SelectContent, SelectItem, SelectTrigger } from '$lib/components/ui/select';
	import MonacoEditor from '$lib/components/MonacoEditor.svelte';
	import { Database, Table2, RefreshCw, Play, Plus } from 'lucide-svelte';
	import DataTable from './data-table.svelte';
	import { createColumns } from './columns';

	let projectId = $derived(page.params.id);
	let resId = $derived(page.params.res_id);

	let selectedSchema = $state('public');
	let selectedTable = $state<string | null>(null);
	let currentPage = $state(1);
	let pageSize = $state(50);
	let sqlQuery = $state('');
	let showSqlEditor = $state(false);
	let queryResult = $state<QueryResult | null>(null);
	let queryError = $state<string | null>(null);
	let showInsertSheet = $state(false);
	let showEditSheet = $state(false);
	let showDeleteSheet = $state(false);
	let insertFormData = $state<Record<string, any>>({});
	let editFormData = $state<Record<string, any>>({});
	let selectedRow = $state<Record<string, any> | null>(null);

	const queryClient = useQueryClient();

	const schemasQuery = createQuery(() => ({
		queryKey: ['studio', 'schemas', projectId, resId] as const,
		queryFn: async () => studioApi.listSchemas(projectId!, resId!),
		refetchOnWindowFocus: false
	}));

	const tablesQuery = createQuery(() => ({
		queryKey: ['studio', 'tables', projectId, resId, selectedSchema] as const,
		queryFn: async () => studioApi.listTables(projectId!, resId!, { schema: selectedSchema }),
		refetchOnWindowFocus: false
	}));

	const schemaQuery = createQuery(() => ({
		queryKey: ['studio', 'schema', projectId, resId, selectedSchema, selectedTable] as const,
		queryFn: async () =>
			studioApi.getTableSchema(projectId!, resId!, selectedTable!, { schema: selectedSchema }),
		enabled: !!selectedTable
	}));

	const dataQuery = createQuery(() => ({
		queryKey: [
			'studio',
			'data',
			projectId,
			resId,
			selectedSchema,
			selectedTable,
			currentPage,
			pageSize
		] as const,
		queryFn: async () =>
			studioApi.getTableData(projectId!, resId!, selectedTable!, {
				page: currentPage,
				page_size: pageSize,
				schema: selectedSchema
			}),
		enabled: !!selectedTable
	}));

	const infoQuery = createQuery(() => ({
		queryKey: ['studio', 'info', projectId, resId] as const,
		queryFn: async () => studioApi.getDatabaseInfo(projectId!, resId!)
	}));

	const executeQueryMutation = createMutation(() => ({
		mutationFn: async (query: string) => studioApi.executeQuery(projectId!, resId!, { query }),
		onSuccess: (data: QueryResult) => {
			queryResult = data;
			queryError = null;
		},
		onError: (error: Error) => {
			queryError = error.message || 'Failed to execute query';
			queryResult = null;
		}
	}));

	const insertRowMutation = createMutation(() => ({
		mutationFn: async (data: Record<string, any>) =>
			studioApi.insertRow(projectId!, resId!, selectedTable!, { data, schema: selectedSchema }),
		onSuccess: () => {
			queryClient.invalidateQueries({
				queryKey: ['studio', 'data', projectId, resId, selectedSchema, selectedTable] as const
			});
			showInsertSheet = false;
			insertFormData = {};
		}
	}));

	const updateRowMutation = createMutation(() => ({
		mutationFn: async (vars: { primary_key: Record<string, any>; data: Record<string, any> }) =>
			studioApi.updateRow(projectId!, resId!, selectedTable!, {
				primary_key: vars.primary_key,
				data: vars.data,
				schema: selectedSchema
			}),
		onSuccess: () => {
			queryClient.invalidateQueries({
				queryKey: ['studio', 'data', projectId, resId, selectedSchema, selectedTable] as const
			});
			showEditSheet = false;
			editFormData = {};
			selectedRow = null;
		}
	}));

	const deleteRowMutation = createMutation(() => ({
		mutationFn: async (primary_key: Record<string, any>) =>
			studioApi.deleteRow(projectId!, resId!, selectedTable!, {
				primary_key,
				schema: selectedSchema
			}),
		onSuccess: () => {
			queryClient.invalidateQueries({
				queryKey: ['studio', 'data', projectId, resId, selectedSchema, selectedTable] as const
			});
			showDeleteSheet = false;
			selectedRow = null;
		}
	}));

	let columns = $derived.by(() => {
		if (!schemaQuery.data) return [];
		return createColumns(schemaQuery.data.columns, openEditSheet, openDeleteSheet);
	});

	let tableData = $derived(dataQuery.data?.rows ?? []);

	function selectSchema(schema: string) {
		selectedSchema = schema;
		selectedTable = null;
		currentPage = 1;
	}

	function selectTable(tableName: string) {
		selectedTable = tableName;
		currentPage = 1;
		showSqlEditor = false;
		queryResult = null;
		queryError = null;
	}

	function handlePageChange(page: number) {
		currentPage = page;
	}

	async function executeQuery() {
		if (!sqlQuery.trim()) return;
		executeQueryMutation.mutate(sqlQuery);
	}

	function openInsertSheet() {
		insertFormData = {};
		if (schemaQuery.data) {
			for (const col of schemaQuery.data.columns) {
				insertFormData[col.name] = '';
			}
		}
		showInsertSheet = true;
	}

	function openEditSheet(row: Record<string, any>) {
		selectedRow = row;
		editFormData = { ...row };
		showEditSheet = true;
	}

	function openDeleteSheet(row: Record<string, any>) {
		selectedRow = row;
		showDeleteSheet = true;
	}

	function handleInsert() {
		const cleanedData: Record<string, any> = {};
		for (const [key, value] of Object.entries(insertFormData)) {
			if (value === '' || value === null) {
				cleanedData[key] = null;
			} else {
				cleanedData[key] = value;
			}
		}
		insertRowMutation.mutate(cleanedData);
	}

	function handleUpdate() {
		if (!selectedRow) return;
		const cleanedData: Record<string, any> = {};
		for (const [key, value] of Object.entries(editFormData)) {
			if (value === '' || value === null) {
				cleanedData[key] = null;
			} else {
				cleanedData[key] = value;
			}
		}
		updateRowMutation.mutate({ primary_key: selectedRow, data: cleanedData });
	}

	function handleDelete() {
		if (!selectedRow) return;
		deleteRowMutation.mutate(selectedRow);
	}
</script>

<svelte:head>
	<title>Database Studio</title>
</svelte:head>

<div class="flex h-screen bg-background">
	<div class="w-64 bg-card border-r border-border flex flex-col">
		<div class="p-4 border-b border-border">
			<div class="flex items-center space-x-2">
				<Database class="w-5 h-5 text-primary" />
				<h2 class="font-semibold text-foreground">Database</h2>
			</div>
			{#if infoQuery.data}
				<p class="text-xs text-muted-foreground mt-1">{infoQuery.data.version}</p>
			{/if}
		</div>

		<div class="p-3 border-b border-border">
			<Label class="text-xs text-muted-foreground mb-1.5 block">Schema</Label>
			{#if schemasQuery.isLoading}
				<div class="text-xs text-muted-foreground">Loading...</div>
			{:else if schemasQuery.data && schemasQuery.data.length > 0}
				<Select
					selected={{ value: selectedSchema, label: selectedSchema }}
					onSelectedChange={(v) => v && selectSchema(v.value)}
				>
					<SelectTrigger class="w-full h-8 text-sm">
						{selectedSchema}
					</SelectTrigger>
					<SelectContent>
						{#each schemasQuery.data as schema}
							<SelectItem value={schema} label={schema}>
								{schema}
							</SelectItem>
						{/each}
					</SelectContent>
				</Select>
			{:else}
				<div class="text-xs text-muted-foreground">No schemas</div>
			{/if}
		</div>

		<div class="px-4 py-2 border-b border-border">
			<h3 class="text-xs font-medium text-muted-foreground uppercase">Tables</h3>
		</div>

		<div class="flex-1 overflow-y-auto">
			{#if tablesQuery.isLoading}
				<div class="p-4 text-sm text-muted-foreground">Loading tables...</div>
			{:else if tablesQuery.error}
				<div class="p-4 text-sm text-destructive">Failed to load tables</div>
			{:else if tablesQuery.data}
				<div class="p-2">
					{#each tablesQuery.data as table}
						<button
							onclick={() => selectTable(table)}
							class="w-full text-left px-3 py-2 rounded text-sm hover:bg-accent transition-colors {selectedTable ===
							table
								? 'bg-accent text-accent-foreground font-medium'
								: 'text-foreground'}"
						>
							<div class="flex items-center space-x-2">
								<Table2 class="w-4 h-4" />
								<span>{table}</span>
							</div>
						</button>
					{/each}
				</div>
			{/if}
		</div>

		<div class="p-4 border-t border-border">
			<Button
				onclick={() => (showSqlEditor = !showSqlEditor)}
				class="w-full"
				variant={showSqlEditor ? 'default' : 'outline'}
			>
				<Play class="w-4 h-4 mr-2" />
				SQL Editor
			</Button>
		</div>
	</div>

	<div class="flex-1 flex flex-col">
		<div class="bg-card border-b border-border p-4">
			<div class="flex items-center justify-between">
				<div>
					{#if selectedTable}
						<h1 class="text-xl font-semibold text-foreground">{selectedTable}</h1>
						{#if schemaQuery.data}
							<p class="text-sm text-muted-foreground mt-1">
								{schemaQuery.data.columns.length} columns · {dataQuery.data?.total_rows ?? 0} rows
							</p>
						{/if}
					{:else if showSqlEditor}
						<h1 class="text-xl font-semibold text-foreground">SQL Editor</h1>
						<p class="text-sm text-muted-foreground mt-1">Execute custom SQL queries</p>
					{:else}
						<h1 class="text-xl font-semibold text-foreground">Database Studio</h1>
						<p class="text-sm text-muted-foreground mt-1">Select a table to view its data</p>
					{/if}
				</div>
				<div class="flex space-x-2">
					{#if selectedTable}
						<Button onclick={openInsertSheet} variant="default" size="sm">
							<Plus class="w-4 h-4 mr-2" />
							Insert Row
						</Button>
						<Button
							onclick={() => {
								dataQuery.refetch();
								schemaQuery.refetch();
							}}
							variant="outline"
							size="sm"
						>
							<RefreshCw class="w-4 h-4" />
						</Button>
					{/if}
				</div>
			</div>
		</div>

		<div class="flex-1 overflow-auto p-6">
			{#if showSqlEditor}
				<Card>
					<CardHeader>
						<CardTitle>SQL Query</CardTitle>
					</CardHeader>
					<CardContent class="space-y-4">
						<div>
							<Label for="sql-query">Query</Label>
							<div class="mt-2 border border-border rounded-md overflow-hidden">
								<MonacoEditor
									bind:value={sqlQuery}
									language="sql"
									height="300px"
									theme="vs-light"
								/>
							</div>
						</div>
						<div class="flex justify-end space-x-2">
							<Button onclick={executeQuery} disabled={executeQueryMutation.isPending}>
								{#if executeQueryMutation.isPending}
									<RefreshCw class="w-4 h-4 mr-2 animate-spin" />
								{:else}
									<Play class="w-4 h-4 mr-2" />
								{/if}
								Execute Query
							</Button>
						</div>

						{#if queryError}
							<div class="p-4 bg-destructive/10 border border-destructive/50 rounded-md">
								<p class="text-sm text-destructive">{queryError}</p>
							</div>
						{/if}

						{#if queryResult}
							<div class="mt-4">
								<div class="flex items-center justify-between mb-2">
									<h3 class="text-sm font-medium text-foreground">Query Result</h3>
									<span class="text-xs text-muted-foreground">
										{queryResult.rows.length} rows · {queryResult.execution_time_ms}ms
									</span>
								</div>
								<div class="border border-border rounded-md overflow-x-auto">
									<table class="w-full text-sm">
										<thead class="bg-muted/50 border-b border-border">
											<tr>
												{#each queryResult.columns as column}
													<th
														class="px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase"
													>
														{column}
													</th>
												{/each}
											</tr>
										</thead>
										<tbody class="divide-y divide-border">
											{#each queryResult.rows as row}
												<tr class="hover:bg-accent/50">
													{#each queryResult.columns as column}
														<td class="px-3 py-2 text-foreground">
															{#if row[column] === null}
																<span class="text-muted-foreground italic">null</span>
															{:else if typeof row[column] === 'boolean'}
																<span class="text-purple-500">{row[column]}</span>
															{:else if typeof row[column] === 'number'}
																<span class="text-blue-500">{row[column]}</span>
															{:else}
																<span>{row[column]}</span>
															{/if}
														</td>
													{/each}
												</tr>
											{/each}
										</tbody>
									</table>
								</div>
							</div>
						{/if}
					</CardContent>
				</Card>
			{:else if selectedTable}
				{#if dataQuery.isLoading}
					<div class="text-center py-12 text-muted-foreground">Loading data...</div>
				{:else if dataQuery.error}
					<div class="text-center py-12 text-destructive">Failed to load data</div>
				{:else if dataQuery.data && schemaQuery.data && columns.length > 0}
					<DataTable
						data={tableData}
						{columns}
						{currentPage}
						{pageSize}
						totalItems={dataQuery.data.total_rows}
						onPageChange={handlePageChange}
					/>
				{/if}
			{:else}
				<div class="text-center py-12">
					<Database class="w-16 h-16 mx-auto text-muted-foreground mb-4" />
					<h3 class="text-lg font-medium text-foreground mb-2">No Table Selected</h3>
					<p class="text-muted-foreground">Select a table from the sidebar to view its data</p>
				</div>
			{/if}
		</div>
	</div>
</div>

<Sheet bind:open={showInsertSheet}>
	<SheetContent side="right" class="w-[500px] sm:w-[600px] overflow-y-auto">
		<SheetHeader>
			<SheetTitle>Insert Row</SheetTitle>
			<SheetDescription>Add a new row to {selectedTable}</SheetDescription>
		</SheetHeader>
		<div class="py-6 space-y-4">
			{#if schemaQuery.data}
				{#each schemaQuery.data.columns as column}
					<div>
						<Label for={`insert-${column.name}`}>
							{column.name}
							<span class="text-xs text-muted-foreground ml-2">
								{column.type}
								{#if column.nullable}
									<span class="text-muted-foreground/80">(nullable)</span>
								{/if}
							</span>
						</Label>
						<Input
							id={`insert-${column.name}`}
							bind:value={insertFormData[column.name]}
							placeholder={column.nullable ? 'null' : ''}
							class="mt-1"
						/>
					</div>
				{/each}
			{/if}
		</div>
		<SheetFooter>
			<Button variant="outline" onclick={() => (showInsertSheet = false)}>Cancel</Button>
			<Button onclick={handleInsert} disabled={insertRowMutation.isPending}>
				{#if insertRowMutation.isPending}
					<RefreshCw class="w-4 h-4 mr-2 animate-spin" />
				{/if}
				Insert Row
			</Button>
		</SheetFooter>
	</SheetContent>
</Sheet>

<Sheet bind:open={showEditSheet}>
	<SheetContent side="right" class="w-[500px] sm:w-[600px] overflow-y-auto">
		<SheetHeader>
			<SheetTitle>Edit Row</SheetTitle>
			<SheetDescription>Update the selected row in {selectedTable}</SheetDescription>
		</SheetHeader>
		<div class="py-6 space-y-4">
			{#if schemaQuery.data && editFormData}
				{#each schemaQuery.data.columns as column}
					<div>
						<Label for={`edit-${column.name}`}>
							{column.name}
							<span class="text-xs text-muted-foreground ml-2">
								{column.type}
								{#if column.nullable}
									<span class="text-muted-foreground/80">(nullable)</span>
								{/if}
							</span>
						</Label>
						<Input
							id={`edit-${column.name}`}
							bind:value={editFormData[column.name]}
							placeholder={column.nullable ? 'null' : ''}
							class="mt-1"
						/>
					</div>
				{/each}
			{/if}
		</div>
		<SheetFooter>
			<Button variant="outline" onclick={() => (showEditSheet = false)}>Cancel</Button>
			<Button onclick={handleUpdate} disabled={updateRowMutation.isPending}>
				{#if updateRowMutation.isPending}
					<RefreshCw class="w-4 h-4 mr-2 animate-spin" />
				{/if}
				Update Row
			</Button>
		</SheetFooter>
	</SheetContent>
</Sheet>

<Sheet bind:open={showDeleteSheet}>
	<SheetContent side="right" class="w-[500px] sm:w-[600px] overflow-y-auto">
		<SheetHeader>
			<SheetTitle>Delete Row</SheetTitle>
			<SheetDescription
				>Are you sure you want to delete this row from {selectedTable}?</SheetDescription
			>
		</SheetHeader>
		<div class="py-6">
			{#if selectedRow && schemaQuery.data}
				<div class="space-y-2">
					{#each schemaQuery.data.columns as column}
						<div class="flex justify-between py-2 border-b border-border">
							<span class="font-medium text-foreground">{column.name}</span>
							<span class="text-foreground">
								{#if selectedRow[column.name] === null}
									<span class="text-muted-foreground italic">null</span>
								{:else if typeof selectedRow[column.name] === 'boolean'}
									<span class="text-purple-500">{selectedRow[column.name]}</span>
								{:else if typeof selectedRow[column.name] === 'number'}
									<span class="text-blue-500">{selectedRow[column.name]}</span>
								{:else}
									{selectedRow[column.name]}
								{/if}
							</span>
						</div>
					{/each}
				</div>
			{/if}
		</div>
		<SheetFooter>
			<Button variant="outline" onclick={() => (showDeleteSheet = false)}>Cancel</Button>
			<Button variant="destructive" onclick={handleDelete} disabled={deleteRowMutation.isPending}>
				{#if deleteRowMutation.isPending}
					<RefreshCw class="w-4 h-4 mr-2 animate-spin" />
				{/if}
				Delete Row
			</Button>
		</SheetFooter>
	</SheetContent>
</Sheet>
