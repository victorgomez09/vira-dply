package dto

type CreateProjectRequest struct {
	Name        string `json:"name"`
	GitURL      string `json:"git_url"`
	Branch      string `json:"branch"`
	SourcePath  string `json:"source_path"`
	Environment string `json:"environment"`
}
