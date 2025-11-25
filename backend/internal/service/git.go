package service

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/victorgomez09/vira-dply/internal/model"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func (s *DeployerService) cloneRepository(ctx context.Context, p *model.Project, targetDir string) error {
	log.Printf("   > Clonando GitURL: %s en %s", p.GitUrl, targetDir)

	log.Printf("git branch %s", p.GitBranch)
	_, err := git.PlainCloneContext(ctx, targetDir, false, &git.CloneOptions{
		URL:               p.GitUrl,
		ReferenceName:     plumbing.NewBranchReferenceName(p.GitBranch),
		SingleBranch:      true,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Tags:              git.NoTags,
		Progress:          os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("error clonando repositorio: %w", err)
	}
	return nil

	// log.Printf("   > Clonando usando la CLI de Git...")

	// // Comando: git clone --depth 1 --branch <rama> <url> <directorio>
	// cmd := exec.CommandContext(ctx, "git",
	// 	"clone",
	// 	"--depth", "1",
	// 	"--branch", p.GitBranch,
	// 	p.GitUrl,
	// 	targetDir,
	// )

	// // Opcional: Configurar variables de entorno si se requiere autenticación (ej. GIT_ASKPASS)

	// output, err := cmd.CombinedOutput()
	// if err != nil {
	// 	log.Printf("Output de la CLI de Git:\n%s", string(output))
	// 	return fmt.Errorf("error al clonar con la CLI de Git: %w", err)
	// }
	// log.Println("   > Clonación con CLI exitosa.")
	// return nil
}
