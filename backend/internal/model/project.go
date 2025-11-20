package model

import "gorm.io/gorm"

type Project struct {
	gorm.Model

	ID           uint `gorm:"primaryKey;autoIncrement"`
	Name         string
	Description  string
	GitUrl       string
	GitBranch    string
	GitSourcePath string
	Status       string
	K8sNamespace string
	PublicUrl    string
}

func (Project) TableName() string {
	return "projects"
}
