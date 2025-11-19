package service

import (
	"github.com/mikrocloud/mikrocloud/internal/domain/settings"
	"github.com/mikrocloud/mikrocloud/internal/domain/settings/repository"
)

type SettingsService struct {
	repo *repository.SettingsRepository
}

func NewSettingsService(repo *repository.SettingsRepository) *SettingsService {
	return &SettingsService{repo: repo}
}

func (s *SettingsService) GetGeneralSettings() (*settings.GeneralSettings, error) {
	return s.repo.GetGeneralSettings()
}

func (s *SettingsService) SaveGeneralSettings(generalSettings *settings.GeneralSettings) error {
	return s.repo.SaveGeneralSettings(generalSettings)
}

func (s *SettingsService) GetAdvancedSettings() (*settings.AdvancedSettings, error) {
	return s.repo.GetAdvancedSettings()
}

func (s *SettingsService) SaveAdvancedSettings(advancedSettings *settings.AdvancedSettings) error {
	return s.repo.SaveAdvancedSettings(advancedSettings)
}

func (s *SettingsService) GetUpdateSettings() (*settings.UpdateSettings, error) {
	return s.repo.GetUpdateSettings()
}

func (s *SettingsService) SaveUpdateSettings(updateSettings *settings.UpdateSettings) error {
	return s.repo.SaveUpdateSettings(updateSettings)
}
