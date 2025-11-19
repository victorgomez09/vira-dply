package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/git"
	"github.com/mikrocloud/mikrocloud/internal/domain/git/repository"
)

type GitService struct {
	repo       repository.GitRepository
	httpClient *http.Client
}

func NewGitService(repo repository.GitRepository) *GitService {
	return &GitService{
		repo: repo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *GitService) ValidateRepository(ctx context.Context, req git.ValidateRepositoryRequest) (*git.ValidateRepositoryResponse, error) {
	var token string
	var baseURL string

	if req.SourceID != nil {
		source, err := s.repo.GetByID(ctx, *req.SourceID)
		if err != nil {
			return &git.ValidateRepositoryResponse{
				Valid:   false,
				Message: "Git source not found",
			}, nil
		}
		token = source.AccessToken
		if source.CustomURL != nil {
			baseURL = *source.CustomURL
		}
	}

	if req.CustomURL != nil {
		baseURL = *req.CustomURL
	}

	slog.Info("provider", "provider", req.Provider)

	switch req.Provider {
	case git.GitProviderGitHub:
		return s.validateGitHubRepository(ctx, req.Repository, token, baseURL)
	case git.GitProviderGitLab:
		return s.validateGitLabRepository(ctx, req.Repository, token, baseURL)
	case git.GitProviderBitbucket:
		return s.validateBitbucketRepository(ctx, req.Repository, token, baseURL)
	case git.GitProviderCustom:
		if baseURL == "" {
			return &git.ValidateRepositoryResponse{
				Valid:   false,
				Message: "Custom URL is required for custom provider",
			}, nil
		}
		return s.validateCustomRepository(ctx, req.Repository, token, baseURL)
	default:
		return &git.ValidateRepositoryResponse{
			Valid:   false,
			Message: fmt.Sprintf("Unsupported provider: %s", req.Provider),
		}, nil
	}
}

func (s *GitService) ListBranches(ctx context.Context, req git.ListBranchesRequest) (*git.ListBranchesResponse, error) {
	var token string
	var baseURL string

	if req.SourceID != nil {
		source, err := s.repo.GetByID(ctx, *req.SourceID)
		if err != nil {
			return nil, fmt.Errorf("git source not found: %w", err)
		}
		token = source.AccessToken
		if source.CustomURL != nil {
			baseURL = *source.CustomURL
		}
	}

	if req.CustomURL != nil {
		baseURL = *req.CustomURL
	}

	switch req.Provider {
	case git.GitProviderGitHub:
		return s.listGitHubBranches(ctx, req.Repository, token, baseURL)
	case git.GitProviderGitLab:
		return s.listGitLabBranches(ctx, req.Repository, token, baseURL)
	case git.GitProviderBitbucket:
		return s.listBitbucketBranches(ctx, req.Repository, token, baseURL)
	case git.GitProviderCustom:
		if baseURL == "" {
			return nil, fmt.Errorf("custom URL is required for custom provider")
		}
		return s.listCustomBranches(ctx, req.Repository, token, baseURL)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", req.Provider)
	}
}

func (s *GitService) DetectBuildMethod(ctx context.Context, req git.DetectBuildMethodRequest) (*git.DetectBuildMethodResponse, error) {
	var token string
	var baseURL string

	if req.SourceID != nil {
		source, err := s.repo.GetByID(ctx, *req.SourceID)
		if err != nil {
			return nil, fmt.Errorf("git source not found: %w", err)
		}
		token = source.AccessToken
		if source.CustomURL != nil {
			baseURL = *source.CustomURL
		}
	}

	if req.CustomURL != nil {
		baseURL = *req.CustomURL
	}

	switch req.Provider {
	case git.GitProviderGitHub:
		return s.detectGitHubBuildMethod(ctx, req.Repository, req.Branch, token, baseURL)
	case git.GitProviderGitLab:
		return s.detectGitLabBuildMethod(ctx, req.Repository, req.Branch, token, baseURL)
	case git.GitProviderBitbucket:
		return s.detectBitbucketBuildMethod(ctx, req.Repository, req.Branch, token, baseURL)
	case git.GitProviderCustom:
		if baseURL == "" {
			return nil, fmt.Errorf("custom URL is required for custom provider")
		}
		return s.detectCustomBuildMethod(ctx, req.Repository, req.Branch, token, baseURL)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", req.Provider)
	}
}

func (s *GitService) CreateGitSource(ctx context.Context, orgID, userID string, req git.CreateGitSourceRequest) (*git.GitSource, error) {
	source := &git.GitSource{
		ID:          uuid.New().String(),
		OrgID:       orgID,
		UserID:      userID,
		Provider:    req.Provider,
		Name:        req.Name,
		AccessToken: req.AccessToken,
		CustomURL:   req.CustomURL,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, source); err != nil {
		return nil, fmt.Errorf("failed to create git source: %w", err)
	}

	return source, nil
}

func (s *GitService) GetGitSource(ctx context.Context, id string) (*git.GitSource, error) {
	source, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get git source: %w", err)
	}
	return source, nil
}

func (s *GitService) ListGitSourcesByUser(ctx context.Context, userID string) ([]*git.GitSource, error) {
	sources, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list git sources: %w", err)
	}
	return sources, nil
}

func (s *GitService) ListGitSourcesByOrg(ctx context.Context, orgID string) ([]*git.GitSource, error) {
	sources, err := s.repo.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list git sources: %w", err)
	}
	return sources, nil
}

func (s *GitService) UpdateGitSource(ctx context.Context, id string, req git.UpdateGitSourceRequest) (*git.GitSource, error) {
	source, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("git source not found: %w", err)
	}

	if req.Name != nil {
		source.Name = *req.Name
	}
	if req.AccessToken != nil {
		source.AccessToken = *req.AccessToken
	}

	source.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, id, source); err != nil {
		return nil, fmt.Errorf("failed to update git source: %w", err)
	}

	return source, nil
}

func (s *GitService) DeleteGitSource(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete git source: %w", err)
	}
	return nil
}

func (s *GitService) validateGitHubRepository(ctx context.Context, repo, token, baseURL string) (*git.ValidateRepositoryResponse, error) {
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}

	url := fmt.Sprintf("%s/repos/%s", baseURL, repo)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	fmt.Println(url)

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return &git.ValidateRepositoryResponse{
			Valid:   false,
			Message: "Failed to connect to GitHub",
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return &git.ValidateRepositoryResponse{
			Valid:   false,
			Message: "Repository not found",
		}, nil
	}

	if resp.StatusCode != 200 {
		return &git.ValidateRepositoryResponse{
			Valid:   false,
			Message: fmt.Sprintf("GitHub API error: %d", resp.StatusCode),
		}, nil
	}

	var ghRepo struct {
		Name          string `json:"name"`
		FullName      string `json:"full_name"`
		HTMLURL       string `json:"html_url"`
		CloneURL      string `json:"clone_url"`
		DefaultBranch string `json:"default_branch"`
		Private       bool   `json:"private"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ghRepo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &git.ValidateRepositoryResponse{
		Valid:   true,
		Message: "Repository is valid",
		Repository: &git.Repository{
			Name:          ghRepo.Name,
			FullName:      ghRepo.FullName,
			URL:           ghRepo.HTMLURL,
			CloneURL:      ghRepo.CloneURL,
			DefaultBranch: ghRepo.DefaultBranch,
			IsPrivate:     ghRepo.Private,
		},
	}, nil
}

func (s *GitService) listGitHubBranches(ctx context.Context, repo, token, baseURL string) (*git.ListBranchesResponse, error) {
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}

	url := fmt.Sprintf("%s/repos/%s/branches", baseURL, repo)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch branches: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("github API error: %d", resp.StatusCode)
	}

	var ghBranches []struct {
		Name      string `json:"name"`
		Protected bool   `json:"protected"`
		Commit    struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ghBranches); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	branches := make([]git.Branch, len(ghBranches))
	for i, b := range ghBranches {
		branches[i] = git.Branch{
			Name:      b.Name,
			Commit:    b.Commit.SHA,
			Protected: b.Protected,
		}
	}

	return &git.ListBranchesResponse{
		Branches: branches,
	}, nil
}

func (s *GitService) detectGitHubBuildMethod(ctx context.Context, repo, branch, token, baseURL string) (*git.DetectBuildMethodResponse, error) {
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}

	config := git.DetectedBuildConfig{
		SuggestedMethod: git.BuildMethodBuildpack,
	}

	dockerfiles := []string{"Dockerfile", "dockerfile"}
	for _, filename := range dockerfiles {
		url := fmt.Sprintf("%s/repos/%s/contents/%s?ref=%s", baseURL, repo, filename, branch)
		exists, path := s.checkFileExists(ctx, url, token)
		if exists {
			config.HasDockerfile = true
			config.DockerfilePath = path
			config.SuggestedMethod = git.BuildMethodDockerfile
			break
		}
	}

	composeFiles := []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"}
	for _, filename := range composeFiles {
		url := fmt.Sprintf("%s/repos/%s/contents/%s?ref=%s", baseURL, repo, filename, branch)
		exists, path := s.checkFileExists(ctx, url, token)
		if exists {
			config.HasDockerCompose = true
			config.DockerComposePath = path
			config.SuggestedMethod = git.BuildMethodCompose
			break
		}
	}

	return &git.DetectBuildMethodResponse{
		Config: config,
	}, nil
}

func (s *GitService) checkFileExists(ctx context.Context, url, token string) (bool, string) {
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return false, ""
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false, ""
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		parts := strings.Split(url, "/contents/")
		if len(parts) == 2 {
			path := strings.Split(parts[1], "?")[0]
			return true, path
		}
	}

	return false, ""
}

func (s *GitService) validateGitLabRepository(ctx context.Context, repo, token, baseURL string) (*git.ValidateRepositoryResponse, error) {
	if baseURL == "" {
		baseURL = "https://gitlab.com/api/v4"
	}

	encodedRepo := strings.ReplaceAll(repo, "/", "%2F")
	url := fmt.Sprintf("%s/projects/%s", baseURL, encodedRepo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return &git.ValidateRepositoryResponse{
			Valid:   false,
			Message: "Failed to connect to GitLab",
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return &git.ValidateRepositoryResponse{
			Valid:   false,
			Message: "Repository not found",
		}, nil
	}

	if resp.StatusCode != 200 {
		return &git.ValidateRepositoryResponse{
			Valid:   false,
			Message: fmt.Sprintf("GitLab API error: %d", resp.StatusCode),
		}, nil
	}

	var glRepo struct {
		Name              string `json:"name"`
		PathWithNamespace string `json:"path_with_namespace"`
		WebURL            string `json:"web_url"`
		HTTPURLToRepo     string `json:"http_url_to_repo"`
		DefaultBranch     string `json:"default_branch"`
		Visibility        string `json:"visibility"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&glRepo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &git.ValidateRepositoryResponse{
		Valid:   true,
		Message: "Repository is valid",
		Repository: &git.Repository{
			Name:          glRepo.Name,
			FullName:      glRepo.PathWithNamespace,
			URL:           glRepo.WebURL,
			CloneURL:      glRepo.HTTPURLToRepo,
			DefaultBranch: glRepo.DefaultBranch,
			IsPrivate:     glRepo.Visibility != "public",
		},
	}, nil
}

func (s *GitService) listGitLabBranches(ctx context.Context, repo, token, baseURL string) (*git.ListBranchesResponse, error) {
	if baseURL == "" {
		baseURL = "https://gitlab.com/api/v4"
	}

	encodedRepo := strings.ReplaceAll(repo, "/", "%2F")
	url := fmt.Sprintf("%s/projects/%s/repository/branches", baseURL, encodedRepo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch branches: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("gitlab API error: %d", resp.StatusCode)
	}

	var glBranches []struct {
		Name      string `json:"name"`
		Protected bool   `json:"protected"`
		Commit    struct {
			ID string `json:"id"`
		} `json:"commit"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&glBranches); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	branches := make([]git.Branch, len(glBranches))
	for i, b := range glBranches {
		branches[i] = git.Branch{
			Name:      b.Name,
			Commit:    b.Commit.ID,
			Protected: b.Protected,
		}
	}

	return &git.ListBranchesResponse{
		Branches: branches,
	}, nil
}

func (s *GitService) detectGitLabBuildMethod(ctx context.Context, repo, branch, token, baseURL string) (*git.DetectBuildMethodResponse, error) {
	if baseURL == "" {
		baseURL = "https://gitlab.com/api/v4"
	}

	encodedRepo := strings.ReplaceAll(repo, "/", "%2F")
	config := git.DetectedBuildConfig{
		SuggestedMethod: git.BuildMethodBuildpack,
	}

	dockerfiles := []string{"Dockerfile", "dockerfile"}
	for _, filename := range dockerfiles {
		url := fmt.Sprintf("%s/projects/%s/repository/files/%s?ref=%s", baseURL, encodedRepo, filename, branch)
		exists, path := s.checkFileExists(ctx, url, token)
		if exists {
			config.HasDockerfile = true
			config.DockerfilePath = path
			config.SuggestedMethod = git.BuildMethodDockerfile
			break
		}
	}

	composeFiles := []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"}
	for _, filename := range composeFiles {
		url := fmt.Sprintf("%s/projects/%s/repository/files/%s?ref=%s", baseURL, encodedRepo, filename, branch)
		exists, path := s.checkFileExists(ctx, url, token)
		if exists {
			config.HasDockerCompose = true
			config.DockerComposePath = path
			config.SuggestedMethod = git.BuildMethodCompose
			break
		}
	}

	return &git.DetectBuildMethodResponse{
		Config: config,
	}, nil
}

func (s *GitService) validateBitbucketRepository(ctx context.Context, repo, token, baseURL string) (*git.ValidateRepositoryResponse, error) {
	if baseURL == "" {
		baseURL = "https://api.bitbucket.org/2.0"
	}

	url := fmt.Sprintf("%s/repositories/%s", baseURL, repo)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return &git.ValidateRepositoryResponse{
			Valid:   false,
			Message: "Failed to connect to Bitbucket",
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return &git.ValidateRepositoryResponse{
			Valid:   false,
			Message: "Repository not found",
		}, nil
	}

	if resp.StatusCode != 200 {
		return &git.ValidateRepositoryResponse{
			Valid:   false,
			Message: fmt.Sprintf("Bitbucket API error: %d", resp.StatusCode),
		}, nil
	}

	var bbRepo struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		Links    struct {
			HTML struct {
				Href string `json:"href"`
			} `json:"html"`
			Clone []struct {
				Name string `json:"name"`
				Href string `json:"href"`
			} `json:"clone"`
		} `json:"links"`
		Mainbranch struct {
			Name string `json:"name"`
		} `json:"mainbranch"`
		IsPrivate bool `json:"is_private"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&bbRepo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	cloneURL := ""
	for _, link := range bbRepo.Links.Clone {
		if link.Name == "https" {
			cloneURL = link.Href
			break
		}
	}

	return &git.ValidateRepositoryResponse{
		Valid:   true,
		Message: "Repository is valid",
		Repository: &git.Repository{
			Name:          bbRepo.Name,
			FullName:      bbRepo.FullName,
			URL:           bbRepo.Links.HTML.Href,
			CloneURL:      cloneURL,
			DefaultBranch: bbRepo.Mainbranch.Name,
			IsPrivate:     bbRepo.IsPrivate,
		},
	}, nil
}

func (s *GitService) listBitbucketBranches(ctx context.Context, repo, token, baseURL string) (*git.ListBranchesResponse, error) {
	if baseURL == "" {
		baseURL = "https://api.bitbucket.org/2.0"
	}

	url := fmt.Sprintf("%s/repositories/%s/refs/branches", baseURL, repo)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch branches: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bitbucket API error: %d", resp.StatusCode)
	}

	var bbResp struct {
		Values []struct {
			Name   string `json:"name"`
			Target struct {
				Hash string `json:"hash"`
			} `json:"target"`
		} `json:"values"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&bbResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	branches := make([]git.Branch, len(bbResp.Values))
	for i, b := range bbResp.Values {
		branches[i] = git.Branch{
			Name:      b.Name,
			Commit:    b.Target.Hash,
			Protected: false,
		}
	}

	return &git.ListBranchesResponse{
		Branches: branches,
	}, nil
}

func (s *GitService) detectBitbucketBuildMethod(ctx context.Context, repo, branch, token, baseURL string) (*git.DetectBuildMethodResponse, error) {
	if baseURL == "" {
		baseURL = "https://api.bitbucket.org/2.0"
	}

	config := git.DetectedBuildConfig{
		SuggestedMethod: git.BuildMethodBuildpack,
	}

	dockerfiles := []string{"Dockerfile", "dockerfile"}
	for _, filename := range dockerfiles {
		url := fmt.Sprintf("%s/repositories/%s/src/%s/%s", baseURL, repo, branch, filename)
		exists, path := s.checkFileExists(ctx, url, token)
		if exists {
			config.HasDockerfile = true
			config.DockerfilePath = path
			config.SuggestedMethod = git.BuildMethodDockerfile
			break
		}
	}

	composeFiles := []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"}
	for _, filename := range composeFiles {
		url := fmt.Sprintf("%s/repositories/%s/src/%s/%s", baseURL, repo, branch, filename)
		exists, path := s.checkFileExists(ctx, url, token)
		if exists {
			config.HasDockerCompose = true
			config.DockerComposePath = path
			config.SuggestedMethod = git.BuildMethodCompose
			break
		}
	}

	return &git.DetectBuildMethodResponse{
		Config: config,
	}, nil
}

func (s *GitService) validateCustomRepository(ctx context.Context, repo, token, baseURL string) (*git.ValidateRepositoryResponse, error) {
	return &git.ValidateRepositoryResponse{
		Valid:   true,
		Message: "Custom repository validation not implemented",
		Repository: &git.Repository{
			Name:          repo,
			FullName:      repo,
			URL:           baseURL + "/" + repo,
			CloneURL:      baseURL + "/" + repo + ".git",
			DefaultBranch: "main",
			IsPrivate:     false,
		},
	}, nil
}

func (s *GitService) listCustomBranches(ctx context.Context, repo, token, baseURL string) (*git.ListBranchesResponse, error) {
	return &git.ListBranchesResponse{
		Branches: []git.Branch{
			{Name: "main", Commit: "", Protected: false},
		},
	}, nil
}

func (s *GitService) detectCustomBuildMethod(ctx context.Context, repo, branch, token, baseURL string) (*git.DetectBuildMethodResponse, error) {
	return &git.DetectBuildMethodResponse{
		Config: git.DetectedBuildConfig{
			SuggestedMethod: git.BuildMethodBuildpack,
		},
	}, nil
}
