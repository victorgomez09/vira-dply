package servers

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Server struct {
	id             ServerID
	name           ServerName
	description    string
	hostname       string
	ipAddress      string
	port           int
	sshKey         string
	serverType     ServerType
	status         ServerStatus
	cpuCores       *int
	memoryMB       *int
	diskGB         *int
	os             *string
	osVersion      *string
	metadata       string
	tags           []string
	organizationID uuid.UUID
	createdAt      time.Time
	updatedAt      time.Time
}

type ServerID struct {
	value string
}

func NewServerID() ServerID {
	return ServerID{value: uuid.New().String()}
}

func ServerIDFromString(s string) (ServerID, error) {
	if s == "" {
		return ServerID{}, fmt.Errorf("server ID cannot be empty")
	}
	return ServerID{value: s}, nil
}

func (id ServerID) String() string {
	return id.value
}

type ServerName struct {
	value string
}

func NewServerName(name string) (ServerName, error) {
	if name == "" {
		return ServerName{}, fmt.Errorf("server name cannot be empty")
	}
	if len(name) > 100 {
		return ServerName{}, fmt.Errorf("server name cannot exceed 100 characters")
	}
	return ServerName{value: name}, nil
}

func (n ServerName) String() string {
	return n.value
}

type ServerType string

const (
	ServerTypeControlPlane ServerType = "control_plane"
	ServerTypeWorker       ServerType = "worker"
	ServerTypeDatabase     ServerType = "database"
	ServerTypeProxy        ServerType = "proxy"
)

type ServerStatus string

const (
	ServerStatusOnline      ServerStatus = "online"
	ServerStatusOffline     ServerStatus = "offline"
	ServerStatusMaintenance ServerStatus = "maintenance"
	ServerStatusError       ServerStatus = "error"
	ServerStatusUnknown     ServerStatus = "unknown"
)

func NewServer(name ServerName, hostname, ipAddress string, port int, serverType ServerType, organizationID uuid.UUID) *Server {
	now := time.Now()
	return &Server{
		id:             NewServerID(),
		name:           name,
		hostname:       hostname,
		ipAddress:      ipAddress,
		port:           port,
		serverType:     serverType,
		status:         ServerStatusOnline,
		metadata:       "{}",
		tags:           make([]string, 0),
		organizationID: organizationID,
		createdAt:      now,
		updatedAt:      now,
	}
}

func (s *Server) ID() ServerID {
	return s.id
}

func (s *Server) Name() ServerName {
	return s.name
}

func (s *Server) Description() string {
	return s.description
}

func (s *Server) Hostname() string {
	return s.hostname
}

func (s *Server) IPAddress() string {
	return s.ipAddress
}

func (s *Server) Port() int {
	return s.port
}

func (s *Server) SSHKey() string {
	return s.sshKey
}

func (s *Server) ServerType() ServerType {
	return s.serverType
}

func (s *Server) Status() ServerStatus {
	return s.status
}

func (s *Server) CPUCores() *int {
	return s.cpuCores
}

func (s *Server) MemoryMB() *int {
	return s.memoryMB
}

func (s *Server) DiskGB() *int {
	return s.diskGB
}

func (s *Server) OS() *string {
	return s.os
}

func (s *Server) OSVersion() *string {
	return s.osVersion
}

func (s *Server) Metadata() string {
	return s.metadata
}

func (s *Server) Tags() []string {
	// Return a copy to maintain encapsulation
	tags := make([]string, len(s.tags))
	copy(tags, s.tags)
	return tags
}

func (s *Server) OrganizationID() uuid.UUID {
	return s.organizationID
}

func (s *Server) CreatedAt() time.Time {
	return s.createdAt
}

func (s *Server) UpdatedAt() time.Time {
	return s.updatedAt
}

func (s *Server) UpdateDescription(description string) {
	s.description = description
	s.updatedAt = time.Now()
}

func (s *Server) UpdateHostname(hostname string) {
	s.hostname = hostname
	s.updatedAt = time.Now()
}

func (s *Server) UpdateIPAddress(ipAddress string) {
	s.ipAddress = ipAddress
	s.updatedAt = time.Now()
}

func (s *Server) UpdatePort(port int) {
	s.port = port
	s.updatedAt = time.Now()
}

func (s *Server) SetSSHKey(sshKey string) {
	s.sshKey = sshKey
	s.updatedAt = time.Now()
}

func (s *Server) ChangeStatus(status ServerStatus) {
	s.status = status
	s.updatedAt = time.Now()
}

func (s *Server) AddTag(tag string) {
	for _, existingTag := range s.tags {
		if existingTag == tag {
			return // Tag already exists
		}
	}
	s.tags = append(s.tags, tag)
	s.updatedAt = time.Now()
}

func (s *Server) RemoveTag(tag string) {
	for i, existingTag := range s.tags {
		if existingTag == tag {
			s.tags = append(s.tags[:i], s.tags[i+1:]...)
			s.updatedAt = time.Now()
			return
		}
	}
}

func (s *Server) SetTags(tags []string) {
	s.tags = make([]string, len(tags))
	copy(s.tags, tags)
	s.updatedAt = time.Now()
}

func (s *Server) UpdateSpecs(cpuCores, memoryMB, diskGB *int, os, osVersion *string) {
	s.cpuCores = cpuCores
	s.memoryMB = memoryMB
	s.diskGB = diskGB
	s.os = os
	s.osVersion = osVersion
	s.updatedAt = time.Now()
}

func (s *Server) UpdateMetadata(metadata string) {
	s.metadata = metadata
	s.updatedAt = time.Now()
}

func (s *Server) CanConnect() bool {
	return s.status == ServerStatusOnline
}

func (s *Server) IsHealthy() bool {
	return s.status == ServerStatusOnline
}

type Domain struct {
	id       DomainID
	serverID ServerID
	domain   string
	isActive bool
}

type DomainID struct {
	value string
}

func NewDomainID() DomainID {
	return DomainID{value: uuid.New().String()}
}

func (id DomainID) String() string {
	return id.value
}

type Destination struct {
	id           DestinationID
	serverID     ServerID
	name         string
	destType     DestinationType
	dockerEngine *DockerEngine
	podmanEngine *PodmanEngine
}

type DestinationID struct {
	value string
}

func NewDestinationID() DestinationID {
	return DestinationID{value: uuid.New().String()}
}

func (id DestinationID) String() string {
	return id.value
}

type DestinationType string

const (
	DestinationTypeDocker DestinationType = "docker"
	DestinationTypePodman DestinationType = "podman"
)

type DockerEngine struct {
	socketPath string
	version    string
}

type PodmanEngine struct {
	socketPath string
	version    string
}

type ProxySettings struct {
	id       ProxyID
	serverID ServerID
	enabled  bool
	port     int
	sslPort  int
}

type ProxyID struct {
	value string
}

func NewProxyID() ProxyID {
	return ProxyID{value: uuid.New().String()}
}

func (id ProxyID) String() string {
	return id.value
}

type ServerResources struct {
	cpuUsage    float64
	memoryUsage float64
	diskUsage   float64
	networkIn   int64
	networkOut  int64
	timestamp   time.Time
}

func ReconstructServer(
	id ServerID,
	name ServerName,
	description, hostname, ipAddress string,
	port int,
	sshKey string,
	serverType ServerType,
	status ServerStatus,
	cpuCores, memoryMB, diskGB *int,
	os, osVersion *string,
	metadata string,
	tags []string,
	organizationID uuid.UUID,
	createdAt, updatedAt time.Time,
) *Server {
	return &Server{
		id:             id,
		name:           name,
		description:    description,
		hostname:       hostname,
		ipAddress:      ipAddress,
		port:           port,
		sshKey:         sshKey,
		serverType:     serverType,
		status:         status,
		cpuCores:       cpuCores,
		memoryMB:       memoryMB,
		diskGB:         diskGB,
		os:             os,
		osVersion:      osVersion,
		metadata:       metadata,
		tags:           tags,
		organizationID: organizationID,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}
