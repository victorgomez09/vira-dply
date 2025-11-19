import { apiClient } from './client';

export type ColumnType =
	| 'string'
	| 'integer'
	| 'float'
	| 'boolean'
	| 'date'
	| 'datetime'
	| 'timestamp'
	| 'json'
	| 'binary'
	| 'text'
	| 'uuid'
	| 'array';

export interface Column {
	name: string;
	type: ColumnType;
	nullable: boolean;
	default_value: string | null;
	primary_key: boolean;
	auto_increment: boolean;
	unique: boolean;
	comment: string;
}

export interface ForeignKey {
	column: string;
	referenced_table: string;
	referenced_column: string;
	on_delete: string;
	on_update: string;
}

export interface Index {
	name: string;
	columns: string[];
	unique: boolean;
	primary: boolean;
}

export interface TableSchema {
	name: string;
	columns: Column[];
	foreign_keys: ForeignKey[];
	indexes: Index[];
	row_count: number;
	comment: string;
}

export interface QueryResult {
	columns: string[];
	rows: Record<string, any>[];
	rows_affected: number;
	execution_time_ms: number;
}

export interface TableDataResult {
	columns: string[];
	rows: Record<string, any>[];
	total_rows: number;
	total_pages: number;
	page: number;
	page_size: number;
	limit: number;
	offset: number;
}

export type FilterOperator =
	| 'eq'
	| 'neq'
	| 'gt'
	| 'gte'
	| 'lt'
	| 'lte'
	| 'like'
	| 'not_like'
	| 'in'
	| 'not_in'
	| 'is_null'
	| 'is_not_null';

export interface Filter {
	column: string;
	operator: FilterOperator;
	value: any;
}

export interface Sort {
	column: string;
	direction: 'asc' | 'desc';
}

export interface TableDataOptions {
	page?: number;
	page_size?: number;
	limit?: number;
	offset?: number;
	filters?: Filter[];
	sorts?: Sort[];
}

export interface DatabaseInfo {
	version: string;
	size: number;
}

export interface ExecuteQueryRequest {
	query: string;
	limit?: number;
	offset?: number;
}

export interface InsertRowRequest {
	data: Record<string, any>;
}

export interface UpdateRowRequest {
	primary_key: Record<string, any>;
	data: Record<string, any>;
}

export interface DeleteRowRequest {
	primary_key: Record<string, any>;
}

export const studioApi = {
	async getDatabaseInfo(projectId: string, databaseId: string): Promise<DatabaseInfo> {
		return apiClient.get<DatabaseInfo>(
			`/projects/${projectId}/databases/${databaseId}/studio/info`
		);
	},

	async listSchemas(projectId: string, databaseId: string): Promise<string[]> {
		const response = await apiClient.get<{ schemas: string[] }>(
			`/projects/${projectId}/databases/${databaseId}/studio/schemas`
		);
		return response.schemas;
	},

	async listTables(projectId: string, databaseId: string, options?: { schema?: string }): Promise<string[]> {
		const params = new URLSearchParams();
		if (options?.schema) params.set('schema', options.schema);
		
		const url = params.toString()
			? `/projects/${projectId}/databases/${databaseId}/studio/tables?${params.toString()}`
			: `/projects/${projectId}/databases/${databaseId}/studio/tables`;
		const response = await apiClient.get<{ tables: string[] }>(url);
		return response.tables;
	},

	async getTableSchema(
		projectId: string,
		databaseId: string,
		tableName: string,
		options?: { schema?: string }
	): Promise<TableSchema> {
		const params = new URLSearchParams();
		if (options?.schema) params.set('schema', options.schema);
		
		const url = params.toString()
			? `/projects/${projectId}/databases/${databaseId}/studio/tables/${tableName}/schema?${params.toString()}`
			: `/projects/${projectId}/databases/${databaseId}/studio/tables/${tableName}/schema`;
		return apiClient.get<TableSchema>(url);
	},

	async getTableData(
		projectId: string,
		databaseId: string,
		tableName: string,
		options?: { page?: number; page_size?: number; filters?: Filter[]; sorts?: Sort[]; schema?: string }
	): Promise<TableDataResult> {
		const params = new URLSearchParams();
		if (options?.schema) params.set('schema', options.schema);
		
		const baseUrl = params.toString()
			? `/projects/${projectId}/databases/${databaseId}/studio/tables/${tableName}/data?${params.toString()}`
			: `/projects/${projectId}/databases/${databaseId}/studio/tables/${tableName}/data`;

		if (options && (options.filters || options.sorts)) {
			return apiClient.post<TableDataResult>(baseUrl, options);
		}

		const queryParams = new URLSearchParams();
		if (options?.page) queryParams.set('page', options.page.toString());
		if (options?.page_size) queryParams.set('page_size', options.page_size.toString());

		const finalUrl = queryParams.toString() ? `${baseUrl}&${queryParams.toString()}` : baseUrl;
		return apiClient.get<TableDataResult>(finalUrl);
	},

	async executeQuery(
		projectId: string,
		databaseId: string,
		request: ExecuteQueryRequest
	): Promise<QueryResult> {
		return apiClient.post<QueryResult>(
			`/projects/${projectId}/databases/${databaseId}/studio/query`,
			request
		);
	},

	async insertRow(
		projectId: string,
		databaseId: string,
		tableName: string,
		options: { data: Record<string, any>; schema?: string }
	): Promise<void> {
		const params = new URLSearchParams();
		if (options.schema) params.set('schema', options.schema);
		
		const url = params.toString()
			? `/projects/${projectId}/databases/${databaseId}/studio/tables/${tableName}/rows?${params.toString()}`
			: `/projects/${projectId}/databases/${databaseId}/studio/tables/${tableName}/rows`;
		await apiClient.post<void>(url, { data: options.data });
	},

	async updateRow(
		projectId: string,
		databaseId: string,
		tableName: string,
		options: { primary_key: Record<string, any>; data: Record<string, any>; schema?: string }
	): Promise<void> {
		const params = new URLSearchParams();
		if (options.schema) params.set('schema', options.schema);
		
		const url = params.toString()
			? `/projects/${projectId}/databases/${databaseId}/studio/tables/${tableName}/rows?${params.toString()}`
			: `/projects/${projectId}/databases/${databaseId}/studio/tables/${tableName}/rows`;
		await apiClient.put<void>(url, { primary_key: options.primary_key, data: options.data });
	},

	async deleteRow(
		projectId: string,
		databaseId: string,
		tableName: string,
		options: { primary_key: Record<string, any>; schema?: string }
	): Promise<void> {
		const params = new URLSearchParams();
		if (options.schema) params.set('schema', options.schema);
		
		const url = params.toString()
			? `/projects/${projectId}/databases/${databaseId}/studio/tables/${tableName}/rows?${params.toString()}`
			: `/projects/${projectId}/databases/${databaseId}/studio/tables/${tableName}/rows`;
		await apiClient.delete<void>(url, { primary_key: options.primary_key });
	}
};
