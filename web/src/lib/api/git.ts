import { apiClient } from './client';

export type GitProvider = 'github' | 'gitlab' | 'bitbucket' | 'custom';

export interface ValidateRepositoryRequest {
  provider: GitProvider;
  repository: string;
  branch?: string;
  token?: string;
  custom_url?: string;
}

export interface ValidateRepositoryResponse {
  valid: boolean;
  message?: string;
}

export interface ListBranchesRequest {
  provider: GitProvider;
  repository: string;
  token?: string;
  custom_url?: string;
}

export interface Branch {
  name: string;
  protected: boolean;
}

export interface ListBranchesResponse {
  branches: Branch[];
}

export interface DetectBuildMethodRequest {
  provider: GitProvider;
  repository_url: string;
  branch: string;
  token?: string;
  custom_url?: string;
}

export interface DetectBuildMethodResponse {
  build_method: string;
  dockerfile_path?: string;
  compose_path?: string;
  message?: string;
}

export interface CreateGitSourceRequest {
  name: string;
  provider: GitProvider;
  access_token: string;
  refresh_token?: string;
  token_expires_at?: string;
  custom_url?: string;
}

export interface UpdateGitSourceRequest {
  name?: string;
  access_token?: string;
  refresh_token?: string;
  token_expires_at?: string;
  custom_url?: string;
}

export interface GitSource {
  id: string;
  org_id: string;
  user_id: string;
  provider: GitProvider;
  name: string;
  custom_url?: string;
  created_at: string;
  updated_at: string;
}

export interface GitSourcesResponse {
  sources: GitSource[];
}

export const gitApi = {
  async validateRepository(
    req: ValidateRepositoryRequest
  ): Promise<ValidateRepositoryResponse> {
    return apiClient.post<ValidateRepositoryResponse>('/git/validate', req);
  },

  async listBranches(req: ListBranchesRequest): Promise<ListBranchesResponse> {
    return apiClient.post<ListBranchesResponse>('/git/branches', req);
  },

  async detectBuildMethod(req: DetectBuildMethodRequest): Promise<DetectBuildMethodResponse> {
    return apiClient.post<DetectBuildMethodResponse>('/git/detect-build', req);
  },

  async listGitSources(): Promise<GitSource[]> {
    const response = await apiClient.get<GitSourcesResponse>('/git/sources');
    return response.sources;
  },

  async getGitSource(sourceId: string): Promise<GitSource> {
    return apiClient.get<GitSource>(`/git/sources/${sourceId}`);
  },

  async createGitSource(req: CreateGitSourceRequest): Promise<GitSource> {
    return apiClient.post<GitSource>('/git/sources', req);
  },

  async updateGitSource(sourceId: string, req: UpdateGitSourceRequest): Promise<GitSource> {
    return apiClient.put<GitSource>(`/git/sources/${sourceId}`, req);
  },

  async deleteGitSource(sourceId: string): Promise<void> {
    return apiClient.delete<void>(`/git/sources/${sourceId}`);
  }
};
