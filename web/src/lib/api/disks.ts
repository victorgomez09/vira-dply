import { apiClient } from './client';

export interface Disk {
	id: string;
	name: string;
	project_id: string;
	service_id?: string;
	size: number;
	size_gb: number;
	mount_path: string;
	filesystem: string;
	status: string;
	persistent: boolean;
	backup_enabled: boolean;
	created_at: string;
	updated_at: string;
}

export interface CreateDiskRequest {
	name: string;
	size_gb: number;
	mount_path: string;
	filesystem: 'ext4' | 'xfs' | 'btrfs' | 'zfs';
	persistent: boolean;
}

export interface ResizeDiskRequest {
	size_gb: number;
}

export interface AttachDiskRequest {
	service_id: string;
}

export interface ListDisksResponse {
	disks: Disk[];
}

export const disksApi = {
	async list(projectId: string): Promise<Disk[]> {
		const response = await apiClient.get<ListDisksResponse>(
			`/projects/${projectId}/disks`
		);
		return response.disks;
	},

	async get(projectId: string, diskId: string): Promise<Disk> {
		return apiClient.get<Disk>(`/projects/${projectId}/disks/${diskId}`);
	},

	async create(projectId: string, data: CreateDiskRequest): Promise<Disk> {
		return apiClient.post<Disk>(`/projects/${projectId}/disks`, data);
	},

	async resize(projectId: string, diskId: string, data: ResizeDiskRequest): Promise<Disk> {
		return apiClient.put<Disk>(`/projects/${projectId}/disks/${diskId}/resize`, data);
	},

	async delete(projectId: string, diskId: string): Promise<void> {
		return apiClient.delete<void>(`/projects/${projectId}/disks/${diskId}`);
	},

	async attach(projectId: string, diskId: string, data: AttachDiskRequest): Promise<Disk> {
		return apiClient.post<Disk>(`/projects/${projectId}/disks/${diskId}/attach`, data);
	},

	async detach(projectId: string, diskId: string): Promise<Disk> {
		return apiClient.post<Disk>(`/projects/${projectId}/disks/${diskId}/detach`, {});
	},
};
