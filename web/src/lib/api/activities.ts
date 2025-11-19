import { apiClient } from './client';

export interface Activity {
	id: string;
	organization_id: string;
	activity_type: string;
	description: string;
	initiator_id?: string;
	initiator_name: string;
	resource_type?: string;
	resource_id?: string;
	resource_name?: string;
	metadata?: Record<string, any>;
	created_at: string;
}

export interface ActivitiesResponse {
	activities: Activity[];
	total: number;
}

export const activitiesApi = {
	async getRecent(orgId: string, limit = 20, offset = 0): Promise<ActivitiesResponse> {
		return apiClient.get<ActivitiesResponse>(`/activities/${orgId}?limit=${limit}&offset=${offset}`);
	},

	async getForResource(resourceType: string, resourceId: string, limit = 20, offset = 0): Promise<ActivitiesResponse> {
		return apiClient.get<ActivitiesResponse>(`/activities/${resourceType}/${resourceId}?limit=${limit}&offset=${offset}`);
	},
};
