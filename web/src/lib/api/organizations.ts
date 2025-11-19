import { apiClient } from './client';

export interface Organization {
	id: string;
	name: string;
	slug: string;
	description: string;
	owner_id: string;
	billing_email: string;
	plan: 'free' | 'pro' | 'enterprise';
	status: 'active' | 'suspended' | 'deleted';
	created_at: string;
	updated_at: string;
}

export const organizationsApi = {
	async list(): Promise<Organization[]> {
		return apiClient.get<Organization[]>('/organizations');
	},

	async get(orgId: string): Promise<Organization> {
		return apiClient.get<Organization>(`/organizations/${orgId}`);
	}
};
