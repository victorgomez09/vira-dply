package git

import (
	"time"
)

type GitProvider string

const (
	GitProviderGitHub    GitProvider = "github"
	GitProviderGitLab    GitProvider = "gitlab"
	GitProviderBitbucket GitProvider = "bitbucket"
	GitProviderCustom    GitProvider = "custom"
)

type GitSource struct {
	ID             string      `json:"id"`
	OrgID          string      `json:"org_id"`
	UserID         string      `json:"user_id"`
	Provider       GitProvider `json:"provider"`
	Name           string      `json:"name"`
	AccessToken    string      `json:"-"`
	RefreshToken   string      `json:"-"`
	TokenExpiresAt *time.Time  `json:"token_expires_at,omitempty"`
	CustomURL      *string     `json:"custom_url,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

type Repository struct {
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	URL           string `json:"url"`
	CloneURL      string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
	IsPrivate     bool   `json:"is_private"`
}

type Branch struct {
	Name      string `json:"name"`
	Commit    string `json:"commit"`
	Protected bool   `json:"protected"`
}

type BuildMethod string

const (
	BuildMethodBuildpack  BuildMethod = "buildpack"
	BuildMethodDockerfile BuildMethod = "dockerfile"
	BuildMethodCompose    BuildMethod = "compose"
)

type DetectedBuildConfig struct {
	HasDockerfile     bool        `json:"has_dockerfile"`
	DockerfilePath    string      `json:"dockerfile_path,omitempty"`
	HasDockerCompose  bool        `json:"has_docker_compose"`
	DockerComposePath string      `json:"docker_compose_path,omitempty"`
	SuggestedMethod   BuildMethod `json:"suggested_method"`
}

type ValidateRepositoryRequest struct {
	Provider   GitProvider `json:"provider"`
	Repository string      `json:"repository"`
	CustomURL  *string     `json:"custom_url,omitempty"`
	SourceID   *string     `json:"source_id,omitempty"`
}

type ValidateRepositoryResponse struct {
	Valid      bool        `json:"valid"`
	Message    string      `json:"message"`
	Repository *Repository `json:"repository,omitempty"`
}

type ListBranchesRequest struct {
	Provider   GitProvider `json:"provider"`
	Repository string      `json:"repository"`
	CustomURL  *string     `json:"custom_url,omitempty"`
	SourceID   *string     `json:"source_id,omitempty"`
}

type ListBranchesResponse struct {
	Branches []Branch `json:"branches"`
}

type DetectBuildMethodRequest struct {
	Provider   GitProvider `json:"provider"`
	Repository string      `json:"repository"`
	Branch     string      `json:"branch"`
	CustomURL  *string     `json:"custom_url,omitempty"`
	SourceID   *string     `json:"source_id,omitempty"`
}

type DetectBuildMethodResponse struct {
	Config DetectedBuildConfig `json:"config"`
}

type CreateGitSourceRequest struct {
	Provider    GitProvider `json:"provider"`
	Name        string      `json:"name"`
	AccessToken string      `json:"access_token"`
	CustomURL   *string     `json:"custom_url,omitempty"`
}

type UpdateGitSourceRequest struct {
	Name        *string `json:"name,omitempty"`
	AccessToken *string `json:"access_token,omitempty"`
}
