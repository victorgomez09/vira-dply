package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/proxy"
)

type SQLiteProxyRepository struct {
	db *sql.DB
}

func NewSQLiteProxyRepository(db *sql.DB) *SQLiteProxyRepository {
	return &SQLiteProxyRepository{db: db}
}

func (r *SQLiteProxyRepository) Create(ctx context.Context, config *proxy.ProxyConfig) error {
	hostnames, _ := json.Marshal(config.Hostnames())
	middlewares, _ := json.Marshal(config.Middlewares())

	query := `
		INSERT INTO proxy_configs (
			id, name, project_id, service_name, container_id, hostnames, target_url, 
			port, protocol, path_prefix, strip_prefix, middlewares, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		config.ID().String(),
		config.Name().String(),
		config.ProjectID().String(),
		config.ServiceName(),
		config.ContainerID(),
		string(hostnames),
		config.TargetURL(),
		config.Port(),
		string(config.Protocol()),
		config.PathPrefix(),
		config.StripPrefix(),
		string(middlewares),
		string(config.Status()),
		config.CreatedAt(),
		config.UpdatedAt(),
	)

	return err
}

func (r *SQLiteProxyRepository) GetByID(ctx context.Context, id proxy.ProxyConfigID) (*proxy.ProxyConfig, error) {
	query := `
		SELECT id, name, project_id, service_name, container_id, hostnames, target_url,
			   port, protocol, path_prefix, strip_prefix, middlewares, status, created_at, updated_at
		FROM proxy_configs WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id.String())
	return r.scanProxyConfig(row)
}

func (r *SQLiteProxyRepository) GetByContainerID(ctx context.Context, containerID string) (*proxy.ProxyConfig, error) {
	query := `
		SELECT id, name, project_id, service_name, container_id, hostnames, target_url,
			   port, protocol, path_prefix, strip_prefix, middlewares, status, created_at, updated_at
		FROM proxy_configs WHERE container_id = ?
	`

	row := r.db.QueryRowContext(ctx, query, containerID)
	return r.scanProxyConfig(row)
}

func (r *SQLiteProxyRepository) GetByServiceName(ctx context.Context, projectID uuid.UUID, serviceName string) (*proxy.ProxyConfig, error) {
	query := `
		SELECT id, name, project_id, service_name, container_id, hostnames, target_url,
			   port, protocol, path_prefix, strip_prefix, middlewares, status, created_at, updated_at
		FROM proxy_configs WHERE project_id = ? AND service_name = ?
	`

	row := r.db.QueryRowContext(ctx, query, projectID.String(), serviceName)
	return r.scanProxyConfig(row)
}

func (r *SQLiteProxyRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]*proxy.ProxyConfig, error) {
	query := `
		SELECT id, name, project_id, service_name, container_id, hostnames, target_url,
			   port, protocol, path_prefix, strip_prefix, middlewares, status, created_at, updated_at
		FROM proxy_configs WHERE project_id = ? ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanProxyConfigs(rows)
}

func (r *SQLiteProxyRepository) ListAll(ctx context.Context) ([]*proxy.ProxyConfig, error) {
	query := `
		SELECT id, name, project_id, service_name, container_id, hostnames, target_url,
			   port, protocol, path_prefix, strip_prefix, middlewares, status, created_at, updated_at
		FROM proxy_configs ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanProxyConfigs(rows)
}

func (r *SQLiteProxyRepository) ListByStatus(ctx context.Context, status proxy.ProxyStatus) ([]*proxy.ProxyConfig, error) {
	query := `
		SELECT id, name, project_id, service_name, container_id, hostnames, target_url,
			   port, protocol, path_prefix, strip_prefix, middlewares, status, created_at, updated_at
		FROM proxy_configs WHERE status = ? ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, string(status))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanProxyConfigs(rows)
}

func (r *SQLiteProxyRepository) Update(ctx context.Context, config *proxy.ProxyConfig) error {
	hostnames, _ := json.Marshal(config.Hostnames())
	middlewares, _ := json.Marshal(config.Middlewares())

	query := `
		UPDATE proxy_configs SET
			name = ?, service_name = ?, container_id = ?, hostnames = ?, target_url = ?,
			port = ?, protocol = ?, path_prefix = ?, strip_prefix = ?, middlewares = ?,
			status = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		config.Name().String(),
		config.ServiceName(),
		config.ContainerID(),
		string(hostnames),
		config.TargetURL(),
		config.Port(),
		string(config.Protocol()),
		config.PathPrefix(),
		config.StripPrefix(),
		string(middlewares),
		string(config.Status()),
		config.UpdatedAt(),
		config.ID().String(),
	)

	return err
}

func (r *SQLiteProxyRepository) Delete(ctx context.Context, id proxy.ProxyConfigID) error {
	query := `DELETE FROM proxy_configs WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id.String())
	return err
}

func (r *SQLiteProxyRepository) DeleteByContainerID(ctx context.Context, containerID string) error {
	query := `DELETE FROM proxy_configs WHERE container_id = ?`
	_, err := r.db.ExecContext(ctx, query, containerID)
	return err
}

func (r *SQLiteProxyRepository) Exists(ctx context.Context, id proxy.ProxyConfigID) (bool, error) {
	query := `SELECT 1 FROM proxy_configs WHERE id = ?`
	var exists int
	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (r *SQLiteProxyRepository) ExistsByHostname(ctx context.Context, hostname string) (bool, error) {
	query := `SELECT 1 FROM proxy_configs WHERE hostnames LIKE ?`
	var exists int
	searchPattern := fmt.Sprintf("%%\"%s\"%%", hostname)
	err := r.db.QueryRowContext(ctx, query, searchPattern).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (r *SQLiteProxyRepository) scanProxyConfig(row *sql.Row) (*proxy.ProxyConfig, error) {
	var (
		id, name, projectIDStr, serviceName, containerID, hostnamesJSON, targetURL string
		port                                                                       int
		protocolStr, pathPrefix, middlewaresJSON, statusStr                        string
		stripPrefix                                                                bool
		createdAt, updatedAt                                                       time.Time
	)

	err := row.Scan(
		&id, &name, &projectIDStr, &serviceName, &containerID, &hostnamesJSON, &targetURL,
		&port, &protocolStr, &pathPrefix, &stripPrefix, &middlewaresJSON, &statusStr,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	return r.buildProxyConfig(
		id, name, projectIDStr, serviceName, containerID, hostnamesJSON, targetURL,
		port, protocolStr, pathPrefix, stripPrefix, middlewaresJSON, statusStr,
		createdAt, updatedAt,
	)
}

func (r *SQLiteProxyRepository) scanProxyConfigs(rows *sql.Rows) ([]*proxy.ProxyConfig, error) {
	var configs []*proxy.ProxyConfig

	for rows.Next() {
		var (
			id, name, projectIDStr, serviceName, containerID, hostnamesJSON, targetURL string
			port                                                                       int
			protocolStr, pathPrefix, middlewaresJSON, statusStr                        string
			stripPrefix                                                                bool
			createdAt, updatedAt                                                       time.Time
		)

		err := rows.Scan(
			&id, &name, &projectIDStr, &serviceName, &containerID, &hostnamesJSON, &targetURL,
			&port, &protocolStr, &pathPrefix, &stripPrefix, &middlewaresJSON, &statusStr,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}

		config, err := r.buildProxyConfig(
			id, name, projectIDStr, serviceName, containerID, hostnamesJSON, targetURL,
			port, protocolStr, pathPrefix, stripPrefix, middlewaresJSON, statusStr,
			createdAt, updatedAt,
		)
		if err != nil {
			return nil, err
		}

		configs = append(configs, config)
	}

	return configs, rows.Err()
}

func (r *SQLiteProxyRepository) buildProxyConfig(
	id, name, projectIDStr, serviceName, containerID, hostnamesJSON, targetURL string,
	port int, protocolStr, pathPrefix string, stripPrefix bool, middlewaresJSON, statusStr string,
	createdAt, updatedAt time.Time,
) (*proxy.ProxyConfig, error) {
	configID, err := proxy.ProxyConfigIDFromString(id)
	if err != nil {
		return nil, err
	}

	configName, err := proxy.NewProxyConfigName(name)
	if err != nil {
		return nil, err
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return nil, err
	}

	var hostnames []string
	if err := json.Unmarshal([]byte(hostnamesJSON), &hostnames); err != nil {
		return nil, err
	}

	var middlewares []proxy.MiddlewareConfig
	if middlewaresJSON != "" {
		if err := json.Unmarshal([]byte(middlewaresJSON), &middlewares); err != nil {
			return nil, err
		}
	}

	return proxy.ReconstructProxyConfig(
		configID,
		configName,
		projectID,
		serviceName,
		containerID,
		hostnames,
		targetURL,
		port,
		proxy.ProxyProtocol(protocolStr),
		pathPrefix,
		stripPrefix,
		nil, // TLS config - TODO: implement if needed
		middlewares,
		nil, // Health check config - TODO: implement if needed
		nil, // Load balancing config - TODO: implement if needed
		proxy.ProxyStatus(statusStr),
		createdAt,
		updatedAt,
	), nil
}

type SQLiteTraefikConfigRepository struct {
	db *sql.DB
}

func NewSQLiteTraefikConfigRepository(db *sql.DB) *SQLiteTraefikConfigRepository {
	return &SQLiteTraefikConfigRepository{db: db}
}

func (r *SQLiteTraefikConfigRepository) Create(ctx context.Context, config *proxy.TraefikGlobalConfig) error {
	query := `
		INSERT INTO traefik_configs (id, version, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		config.ID().String(),
		"v3.0",
		config.CreatedAt(),
		config.UpdatedAt(),
	)

	return err
}

func (r *SQLiteTraefikConfigRepository) GetCurrent(ctx context.Context) (*proxy.TraefikGlobalConfig, error) {
	query := `
		SELECT id, version, created_at, updated_at
		FROM traefik_configs
		ORDER BY created_at DESC
		LIMIT 1
	`

	var id, version string
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query).Scan(&id, &version, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return default config if none exists
			return proxy.NewTraefikGlobalConfig(), nil
		}
		return nil, err
	}

	// For now, return a basic config. In a full implementation,
	// we'd store and retrieve the full configuration
	return proxy.NewTraefikGlobalConfig(), nil
}

func (r *SQLiteTraefikConfigRepository) Update(ctx context.Context, config *proxy.TraefikGlobalConfig) error {
	query := `
		UPDATE traefik_configs SET updated_at = ? WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, config.UpdatedAt(), config.ID().String())
	return err
}

func (r *SQLiteTraefikConfigRepository) Delete(ctx context.Context, id proxy.TraefikConfigID) error {
	query := `DELETE FROM traefik_configs WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id.String())
	return err
}

func (r *SQLiteTraefikConfigRepository) Exists(ctx context.Context) (bool, error) {
	query := `SELECT 1 FROM traefik_configs LIMIT 1`
	var exists int
	err := r.db.QueryRowContext(ctx, query).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}
