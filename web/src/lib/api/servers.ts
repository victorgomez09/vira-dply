import { apiClient } from './client';

export interface Server {
	id: string;
	name: string;
	description?: string;
	hostname: string;
	ip_address: string;
	port: number;
	ssh_key?: string;
	server_type: 'control_plane' | 'worker' | 'database' | 'proxy';
	status: 'online' | 'offline' | 'maintenance' | 'error' | 'unknown';
	cpu_cores?: number;
	memory_mb?: number;
	disk_gb?: number;
	os?: string;
	os_version?: string;
	metadata?: string;
	tags?: string[];
	organization_id: string;
	created_at: string;
	updated_at: string;
}

export interface CreateServerRequest {
	name: string;
	description?: string;
	hostname: string;
	ip_address: string;
	port: number;
	ssh_key?: string;
	server_type: 'control_plane' | 'worker' | 'database' | 'proxy';
	tags?: string[];
}

export interface UpdateServerRequest {
	name?: string;
	description?: string;
	hostname?: string;
	ip_address?: string;
	port?: number;
	ssh_key?: string;
	status?: 'online' | 'offline' | 'maintenance' | 'error' | 'unknown';
	cpu_cores?: number;
	memory_mb?: number;
	disk_gb?: number;
	os?: string;
	os_version?: string;
	tags?: string[];
}

export interface ServersResponse {
	servers: Server[];
}

export const serversApi = {
	async list(): Promise<Server[]> {
		const response = await apiClient.get<ServersResponse>('/servers');
		return response.servers;
	},

	async get(id: string): Promise<Server> {
		return apiClient.get<Server>(`/servers/${id}`);
	},

	async create(data: CreateServerRequest): Promise<Server> {
		return apiClient.post<Server>('/servers', data);
	},

	async update(id: string, data: UpdateServerRequest): Promise<Server> {
		return apiClient.put<Server>(`/servers/${id}`, data);
	},

	async delete(id: string): Promise<void> {
		return apiClient.delete<void>(`/servers/${id}`);
	},
};
