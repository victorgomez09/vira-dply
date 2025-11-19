export interface ApiError {
	message: string;
	status: number;
	error?: string;
}

interface ErrorResponse {
	error: string;
	message?: string;
}

export class ApiClient {
	get baseURL(): string {
		return '/api';
	}

	getHeaders(): HeadersInit {
		const token = localStorage.getItem('auth_token');
		const headers: HeadersInit = {
			'Content-Type': 'application/json',
		};

		if (token) {
			headers['Authorization'] = `Bearer ${token}`;
		}

		return headers;
	}

	private async request<T>(
		endpoint: string,
		options: RequestInit = {}
	): Promise<T> {
		const headers: HeadersInit = {
			...this.getHeaders(),
			...options.headers,
		};

		const response = await fetch(`${this.baseURL}${endpoint}`, {
			...options,
			headers,
			credentials: 'include',
		});

		if (!response.ok) {
			let errorMessage = 'An error occurred';
			try {
				const errorData: ErrorResponse = await response.json();
				errorMessage = errorData.message || errorData.error || errorMessage;
			} catch {
				errorMessage = await response.text() || errorMessage;
			}

			const error: ApiError = {
				message: errorMessage,
				status: response.status,
			};
			throw error;
		}

		return response.json();
	}

	async get<T>(endpoint: string): Promise<T> {
		return this.request<T>(endpoint, { method: 'GET' });
	}

	async post<T>(endpoint: string, data?: unknown): Promise<T> {
		return this.request<T>(endpoint, {
			method: 'POST',
			body: data ? JSON.stringify(data) : undefined,
		});
	}

	async put<T>(endpoint: string, data?: unknown): Promise<T> {
		return this.request<T>(endpoint, {
			method: 'PUT',
			body: data ? JSON.stringify(data) : undefined,
		});
	}

	async delete<T>(endpoint: string, data?: unknown): Promise<T> {
		return this.request<T>(endpoint, {
			method: 'DELETE',
			body: data ? JSON.stringify(data) : undefined,
		});
	}
}

export const apiClient = new ApiClient();
