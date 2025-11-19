package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/servers"
)

type ServersRepository struct {
	db *sql.DB
}

func NewServersRepository(db *sql.DB) *ServersRepository {
	return &ServersRepository{db: db}
}

type ServerDTO struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Hostname       string    `json:"hostname"`
	IPAddress      string    `json:"ip_address"`
	Port           int       `json:"port"`
	SSHKey         string    `json:"ssh_key,omitempty"`
	ServerType     string    `json:"server_type"`
	Status         string    `json:"status"`
	CPUCores       *int      `json:"cpu_cores,omitempty"`
	MemoryMB       *int      `json:"memory_mb,omitempty"`
	DiskGB         *int      `json:"disk_gb,omitempty"`
	OS             *string   `json:"os,omitempty"`
	OSVersion      *string   `json:"os_version,omitempty"`
	Metadata       any       `json:"metadata,omitempty"`
	Tags           []string  `json:"tags"`
	OrganizationID string    `json:"organization_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (r *ServersRepository) Create(server *servers.Server) error {
	tagsJSON, err := json.Marshal(server.Tags())
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	_, err = r.db.Exec(`
		INSERT INTO servers (
			id, name, description, hostname, ip_address, port, ssh_key,
			server_type, status, cpu_cores, memory_mb, disk_gb, os, os_version,
			metadata, tags, organization_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		server.ID().String(),
		server.Name().String(),
		server.Description(),
		server.Hostname(),
		server.IPAddress(),
		server.Port(),
		server.SSHKey(),
		server.ServerType(),
		server.Status(),
		server.CPUCores(),
		server.MemoryMB(),
		server.DiskGB(),
		server.OS(),
		server.OSVersion(),
		server.Metadata(),
		string(tagsJSON),
		server.OrganizationID().String(),
		server.CreatedAt(),
		server.UpdatedAt(),
	)

	return err
}

func (r *ServersRepository) GetByID(id servers.ServerID) (*servers.Server, error) {
	var (
		idStr, name, description, hostname, ipAddress, sshKey, metadata, tagsJSON string
		port                                                                      int
		serverType, status                                                        string
		cpuCores, memoryMB, diskGB                                                *int
		os, osVersion                                                             *string
		orgIDStr                                                                  string
		createdAt, updatedAt                                                      time.Time
	)

	err := r.db.QueryRow(`
		SELECT id, name, description, hostname, ip_address, port, ssh_key,
			server_type, status, cpu_cores, memory_mb, disk_gb, os, os_version,
			metadata, tags, organization_id, created_at, updated_at
		FROM servers WHERE id = ?
	`, id.String()).Scan(
		&idStr, &name, &description, &hostname, &ipAddress, &port, &sshKey,
		&serverType, &status, &cpuCores, &memoryMB, &diskGB, &os, &osVersion,
		&metadata, &tagsJSON, &orgIDStr, &createdAt, &updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("server not found")
	}
	if err != nil {
		return nil, err
	}

	return r.scanServer(idStr, name, description, hostname, ipAddress, port, sshKey,
		serverType, status, cpuCores, memoryMB, diskGB, os, osVersion,
		metadata, tagsJSON, orgIDStr, createdAt, updatedAt)
}

func (r *ServersRepository) Update(server *servers.Server) error {
	tagsJSON, err := json.Marshal(server.Tags())
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	_, err = r.db.Exec(`
		UPDATE servers SET
			name = ?, description = ?, hostname = ?, ip_address = ?, port = ?,
			ssh_key = ?, status = ?, cpu_cores = ?, memory_mb = ?, disk_gb = ?,
			os = ?, os_version = ?, metadata = ?, tags = ?, updated_at = ?
		WHERE id = ?
	`,
		server.Name().String(),
		server.Description(),
		server.Hostname(),
		server.IPAddress(),
		server.Port(),
		server.SSHKey(),
		server.Status(),
		server.CPUCores(),
		server.MemoryMB(),
		server.DiskGB(),
		server.OS(),
		server.OSVersion(),
		server.Metadata(),
		string(tagsJSON),
		server.UpdatedAt(),
		server.ID().String(),
	)

	return err
}

func (r *ServersRepository) Delete(id servers.ServerID) error {
	_, err := r.db.Exec(`DELETE FROM servers WHERE id = ?`, id.String())
	return err
}

func (r *ServersRepository) ListByOrganization(organizationID uuid.UUID) ([]*servers.Server, error) {
	rows, err := r.db.Query(`
		SELECT id, name, description, hostname, ip_address, port, ssh_key,
			server_type, status, cpu_cores, memory_mb, disk_gb, os, os_version,
			metadata, tags, organization_id, created_at, updated_at
		FROM servers WHERE organization_id = ?
		ORDER BY created_at DESC
	`, organizationID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanServers(rows)
}

func (r *ServersRepository) ListByType(organizationID uuid.UUID, serverType servers.ServerType) ([]*servers.Server, error) {
	rows, err := r.db.Query(`
		SELECT id, name, description, hostname, ip_address, port, ssh_key,
			server_type, status, cpu_cores, memory_mb, disk_gb, os, os_version,
			metadata, tags, organization_id, created_at, updated_at
		FROM servers WHERE organization_id = ? AND server_type = ?
		ORDER BY created_at DESC
	`, organizationID.String(), serverType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanServers(rows)
}

func (r *ServersRepository) GetByHostname(hostname string) (*servers.Server, error) {
	var (
		idStr, name, description, hostnameStr, ipAddress, sshKey, metadata, tagsJSON string
		port                                                                         int
		serverType, status                                                           string
		cpuCores, memoryMB, diskGB                                                   *int
		os, osVersion                                                                *string
		orgIDStr                                                                     string
		createdAt, updatedAt                                                         time.Time
	)

	err := r.db.QueryRow(`
		SELECT id, name, description, hostname, ip_address, port, ssh_key,
			server_type, status, cpu_cores, memory_mb, disk_gb, os, os_version,
			metadata, tags, organization_id, created_at, updated_at
		FROM servers WHERE hostname = ?
	`, hostname).Scan(
		&idStr, &name, &description, &hostnameStr, &ipAddress, &port, &sshKey,
		&serverType, &status, &cpuCores, &memoryMB, &diskGB, &os, &osVersion,
		&metadata, &tagsJSON, &orgIDStr, &createdAt, &updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("server not found")
	}
	if err != nil {
		return nil, err
	}

	return r.scanServer(idStr, name, description, hostnameStr, ipAddress, port, sshKey,
		serverType, status, cpuCores, memoryMB, diskGB, os, osVersion,
		metadata, tagsJSON, orgIDStr, createdAt, updatedAt)
}

func (r *ServersRepository) scanServers(rows *sql.Rows) ([]*servers.Server, error) {
	var serverList []*servers.Server

	for rows.Next() {
		var (
			idStr, name, description, hostname, ipAddress, sshKey, metadata, tagsJSON string
			port                                                                      int
			serverType, status                                                        string
			cpuCores, memoryMB, diskGB                                                *int
			os, osVersion                                                             *string
			orgIDStr                                                                  string
			createdAt, updatedAt                                                      time.Time
		)

		err := rows.Scan(
			&idStr, &name, &description, &hostname, &ipAddress, &port, &sshKey,
			&serverType, &status, &cpuCores, &memoryMB, &diskGB, &os, &osVersion,
			&metadata, &tagsJSON, &orgIDStr, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}

		server, err := r.scanServer(idStr, name, description, hostname, ipAddress, port, sshKey,
			serverType, status, cpuCores, memoryMB, diskGB, os, osVersion,
			metadata, tagsJSON, orgIDStr, createdAt, updatedAt)
		if err != nil {
			continue
		}

		serverList = append(serverList, server)
	}

	return serverList, nil
}

func (r *ServersRepository) scanServer(
	idStr, name, description, hostname, ipAddress string, port int, sshKey,
	serverType, status string, cpuCores, memoryMB, diskGB *int, os, osVersion *string,
	metadata, tagsJSON, orgIDStr string, createdAt, updatedAt time.Time,
) (*servers.Server, error) {
	serverID, err := servers.ServerIDFromString(idStr)
	if err != nil {
		return nil, err
	}

	serverName, err := servers.NewServerName(name)
	if err != nil {
		return nil, err
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return nil, err
	}

	var tags []string
	if tagsJSON != "" {
		if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
			tags = []string{}
		}
	}

	return servers.ReconstructServer(
		serverID, serverName, description, hostname, ipAddress, port, sshKey,
		servers.ServerType(serverType), servers.ServerStatus(status),
		cpuCores, memoryMB, diskGB, os, osVersion, metadata, tags,
		orgID, createdAt, updatedAt,
	), nil
}

func ToDTO(server *servers.Server) (*ServerDTO, error) {
	var metadataMap map[string]any
	if server.Metadata() != "" && server.Metadata() != "{}" {
		if err := json.Unmarshal([]byte(server.Metadata()), &metadataMap); err != nil {
			metadataMap = make(map[string]any)
		}
	}

	return &ServerDTO{
		ID:             server.ID().String(),
		Name:           server.Name().String(),
		Description:    server.Description(),
		Hostname:       server.Hostname(),
		IPAddress:      server.IPAddress(),
		Port:           server.Port(),
		SSHKey:         strings.TrimSpace(server.SSHKey()),
		ServerType:     string(server.ServerType()),
		Status:         string(server.Status()),
		CPUCores:       server.CPUCores(),
		MemoryMB:       server.MemoryMB(),
		DiskGB:         server.DiskGB(),
		OS:             server.OS(),
		OSVersion:      server.OSVersion(),
		Metadata:       metadataMap,
		Tags:           server.Tags(),
		OrganizationID: server.OrganizationID().String(),
		CreatedAt:      server.CreatedAt(),
		UpdatedAt:      server.UpdatedAt(),
	}, nil
}
