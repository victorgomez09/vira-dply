package repository

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/mikrocloud/mikrocloud/internal/domain/settings"
)

type SettingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) GetGeneralSettings() (*settings.GeneralSettings, error) {
	var jsonData string
	err := r.db.QueryRow("SELECT value FROM system_settings WHERE key = 'general'").Scan(&jsonData)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &settings.GeneralSettings{
				Timezone:           "UTC",
				AllowRegistrations: true,
				DoNotTrack:         false,
			}, nil
		}
		return nil, err
	}

	var generalSettings settings.GeneralSettings
	if err := json.Unmarshal([]byte(jsonData), &generalSettings); err != nil {
		return nil, err
	}

	return &generalSettings, nil
}

func (r *SettingsRepository) SaveGeneralSettings(s *settings.GeneralSettings) error {
	jsonData, err := json.Marshal(s)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`
		INSERT INTO system_settings (key, value) VALUES ('general', ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP
	`, string(jsonData))

	return err
}

func (r *SettingsRepository) GetAdvancedSettings() (*settings.AdvancedSettings, error) {
	var jsonData string
	err := r.db.QueryRow("SELECT value FROM system_settings WHERE key = 'advanced'").Scan(&jsonData)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &settings.AdvancedSettings{
				DNSValidation: true,
				DNSServers:    "8.8.8.8,1.1.1.1",
				APIAccess:     false,
				AllowedIPs:    "0.0.0.0",
			}, nil
		}
		return nil, err
	}

	var advancedSettings settings.AdvancedSettings
	if err := json.Unmarshal([]byte(jsonData), &advancedSettings); err != nil {
		return nil, err
	}

	return &advancedSettings, nil
}

func (r *SettingsRepository) SaveAdvancedSettings(s *settings.AdvancedSettings) error {
	jsonData, err := json.Marshal(s)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`
		INSERT INTO system_settings (key, value) VALUES ('advanced', ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP
	`, string(jsonData))

	return err
}

func (r *SettingsRepository) GetUpdateSettings() (*settings.UpdateSettings, error) {
	var jsonData string
	err := r.db.QueryRow("SELECT value FROM system_settings WHERE key = 'updates'").Scan(&jsonData)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &settings.UpdateSettings{
				UpdateCheckFrequency: "hourly",
				AutoUpdate:           true,
				AutoUpdateFrequency:  "daily",
				AutoUpdateTime:       "00:00",
			}, nil
		}
		return nil, err
	}

	var updateSettings settings.UpdateSettings
	if err := json.Unmarshal([]byte(jsonData), &updateSettings); err != nil {
		return nil, err
	}

	return &updateSettings, nil
}

func (r *SettingsRepository) SaveUpdateSettings(s *settings.UpdateSettings) error {
	jsonData, err := json.Marshal(s)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`
		INSERT INTO system_settings (key, value) VALUES ('updates', ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP
	`, string(jsonData))

	return err
}
