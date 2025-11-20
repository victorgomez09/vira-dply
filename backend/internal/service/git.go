package service

import (
	"context"
	"fmt"
	"log"

	"github.com/victorgomez09/vira-dply/internal/model"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func (s *DeployerService) cloneRepository(ctx context.Context, p *model.Project, targetDir string) error {
	log.Printf("   > Clonando GitURL: %s en %s", p.GitUrl, targetDir)

	branchRef := plumbing.ReferenceName("refs/heads/" + p.GitBranch)
	_, err := git.PlainCloneContext(ctx, targetDir, false, &git.CloneOptions{
		URL:           p.GitUrl,
		ReferenceName: branchRef,
		SingleBranch:  true,
		Depth:         1,
		Tags:          git.NoTags,
		// Si es un repo privado, necesitarás Auth
		// Auth:          ssh.NewPublicKeysFromFile("git", "/path/to/key", ""),
	})
	if err != nil {
		return fmt.Errorf("error clonando repositorio: %w", err)
	}
	return nil
}
