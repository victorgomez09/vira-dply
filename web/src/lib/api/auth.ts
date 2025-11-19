import { apiClient } from './client';

export interface User {
	id: string;
	name: string;
	email: string;
	username?: string;
	role?: string;
}

export interface LoginRequest {
	email: string;
	password: string;
}

export interface RegisterRequest {
	name: string;
	email: string;
	password: string;
}

export interface AuthResponse {
	token: string;
	user: User;
}

export const authApi = {
	async login(credentials: LoginRequest): Promise<AuthResponse> {
		const response = await apiClient.post<AuthResponse>('/auth/login', credentials);
		localStorage.setItem('auth_token', response.token);
		return response;
	},

	async register(data: RegisterRequest): Promise<AuthResponse> {
		const response = await apiClient.post<AuthResponse>('/auth/register', data);
		localStorage.setItem('auth_token', response.token);
		return response;
	},

	async getProfile(): Promise<User> {
		return apiClient.get<User>('/auth/profile');
	},

	logout(): void {
		localStorage.removeItem('auth_token');
	},

	getToken(): string | null {
		return localStorage.getItem('auth_token');
	},

	isAuthenticated(): boolean {
		return !!this.getToken();
	},
};
