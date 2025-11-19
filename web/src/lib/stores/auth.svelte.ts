import { authApi, type User } from '$lib/api';

interface AuthState {
	user: User | null;
	isAuthenticated: boolean;
	isLoading: boolean;
}

class AuthStore {
	private state = $state<AuthState>({
		user: null,
		isAuthenticated: false,
		isLoading: true,
	});

	get user() {
		return this.state.user;
	}

	get isAuthenticated() {
		return this.state.isAuthenticated;
	}

	get isLoading() {
		return this.state.isLoading;
	}

	setUser(user: User | null) {
		this.state.user = user;
		this.state.isAuthenticated = !!user;
	}

	setLoading(loading: boolean) {
		this.state.isLoading = loading;
	}

	logout() {
		authApi.logout();
		this.setUser(null);
	}

	initialize() {
		const token = authApi.getToken();
		if (token) {
			this.state.isAuthenticated = true;
		}
		this.state.isLoading = false;
	}
}

export const authStore = new AuthStore();
