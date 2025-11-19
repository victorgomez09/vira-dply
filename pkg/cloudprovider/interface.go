package cloudprovider

import (
	"context"
	"time"

	"github.com/mikrocloud/mikrocloud/internal/domain"
)

// CloudProvider defines the interface that all cloud providers must implement
type CloudProvider interface {
	// Static Site Hosting
	DeployStaticSite(ctx context.Context, config StaticSiteConfig) (*StaticSiteDeployment, error)
	UpdateStaticSite(ctx context.Context, deploymentID string, config StaticSiteConfig) error
	DeleteStaticSite(ctx context.Context, deploymentID string) error
	
	// Serverless Functions
	DeployServerlessFunction(ctx context.Context, config FunctionConfig) (*FunctionDeployment, error)
	UpdateServerlessFunction(ctx context.Context, functionID string, config FunctionConfig) error
	DeleteServerlessFunction(ctx context.Context, functionID string) error
	InvokeFunction(ctx context.Context, functionID string, payload []byte) (*FunctionResponse, error)
	
	// Container Deployments
	DeployContainer(ctx context.Context, config ContainerConfig) (*ContainerDeployment, error)
	UpdateContainer(ctx context.Context, deploymentID string, config ContainerConfig) error
	DeleteContainer(ctx context.Context, deploymentID string) error
	ScaleContainer(ctx context.Context, deploymentID string, instances int) error
	
	// Database Management
	CreateDatabase(ctx context.Context, config DatabaseConfig) (*DatabaseDeployment, error)
	UpdateDatabase(ctx context.Context, databaseID string, config DatabaseConfig) error
	DeleteDatabase(ctx context.Context, databaseID string) error
	BackupDatabase(ctx context.Context, databaseID string) (*DatabaseBackup, error)
	RestoreDatabase(ctx context.Context, databaseID string, backupID string) error
	
	// Domain & SSL Management
	SetupDomain(ctx context.Context, domain string, target string) (*DNSRecord, error)
	SetupSSLCertificate(ctx context.Context, domain string) (*SSLCertificate, error)
	DeleteDomain(ctx context.Context, domain string) error
	
	// Cost Management
	GetCosts(ctx context.Context, projectID string, timeRange TimeRange) (*CostData, error)
	SetCostBudget(ctx context.Context, projectID string, budget CostBudget) error
	GetCostAlerts(ctx context.Context, projectID string) ([]CostAlert, error)
	
	// Networking
	SetupNetworking(ctx context.Context, config NetworkConfig) (*NetworkDeployment, error)
	DeleteNetworking(ctx context.Context, networkID string) error
	
	// Secret Management
	CreateSecret(ctx context.Context, key string, value string) (*Secret, error)
	GetSecret(ctx context.Context, key string) (*Secret, error)
	UpdateSecret(ctx context.Context, key string, value string) error
	DeleteSecret(ctx context.Context, key string) error
	
	// Monitoring & Logging
	GetMetrics(ctx context.Context, resourceID string, metricName string, timeRange TimeRange) (*MetricData, error)
	GetLogs(ctx context.Context, resourceID string, timeRange TimeRange) (*LogData, error)
	CreateAlert(ctx context.Context, config AlertConfig) (*Alert, error)
	
	// Provider Info
	GetProviderType() domain.CloudProvider
	GetRegions() ([]Region, error)
	ValidateCredentials(ctx context.Context) error
}

// Configuration structs
type StaticSiteConfig struct {
	ProjectID     string            `json:"project_id"`
	Name          string            `json:"name"`
	Domain        string            `json:"domain,omitempty"`
	BuildCommand  string            `json:"build_command,omitempty"`
	BuildDir      string            `json:"build_dir"`
	Environment   map[string]string `json:"environment,omitempty"`
	CustomHeaders map[string]string `json:"custom_headers,omitempty"`
}

type FunctionConfig struct {
	ProjectID     string            `json:"project_id"`
	Name          string            `json:"name"`
	Runtime       string            `json:"runtime"` // nodejs18, python39, go121, etc.
	Handler       string            `json:"handler"`
	Code          []byte            `json:"code"` // ZIP archive
	Environment   map[string]string `json:"environment,omitempty"`
	Timeout       int               `json:"timeout"`       // seconds
	MemorySize    int               `json:"memory_size"`   // MB
	VPCConfig     *VPCConfig        `json:"vpc_config,omitempty"`
}

type ContainerConfig struct {
	ProjectID     string            `json:"project_id"`
	Name          string            `json:"name"`
	Image         string            `json:"image"`
	Port          int               `json:"port"`
	Environment   map[string]string `json:"environment,omitempty"`
	CPU           float64           `json:"cpu"`    // CPU units
	Memory        int               `json:"memory"` // MB
	Instances     int               `json:"instances"`
	VPCConfig     *VPCConfig        `json:"vpc_config,omitempty"`
	HealthCheck   *HealthCheck      `json:"health_check,omitempty"`
}

type DatabaseConfig struct {
	ProjectID        string            `json:"project_id"`
	Name             string            `json:"name"`
	Engine           string            `json:"engine"` // postgresql, mysql, mongodb, redis
	Version          string            `json:"version"`
	InstanceClass    string            `json:"instance_class"`
	AllocatedStorage int               `json:"allocated_storage"` // GB
	MultiAZ          bool              `json:"multi_az"`
	BackupRetention  int               `json:"backup_retention"` // days
	VPCConfig        *VPCConfig        `json:"vpc_config,omitempty"`
	Encryption       bool              `json:"encryption"`
}

type NetworkConfig struct {
	ProjectID        string   `json:"project_id"`
	VPCName          string   `json:"vpc_name"`
	CIDRBlock        string   `json:"cidr_block"`
	PublicSubnets    []string `json:"public_subnets"`
	PrivateSubnets   []string `json:"private_subnets"`
	EnableNATGateway bool     `json:"enable_nat_gateway"`
	EnableVPNGateway bool     `json:"enable_vpn_gateway"`
}

type VPCConfig struct {
	VPCID           string   `json:"vpc_id"`
	SubnetIDs       []string `json:"subnet_ids"`
	SecurityGroupIDs []string `json:"security_group_ids"`
}

type HealthCheck struct {
	Path                string `json:"path"`
	IntervalSeconds     int    `json:"interval_seconds"`
	TimeoutSeconds      int    `json:"timeout_seconds"`
	HealthyThreshold    int    `json:"healthy_threshold"`
	UnhealthyThreshold  int    `json:"unhealthy_threshold"`
}

// Response structs
type StaticSiteDeployment struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	CDNEndpoint string    `json:"cdn_endpoint"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type FunctionDeployment struct {
	ID        string    `json:"id"`
	ARN       string    `json:"arn"`
	URL       string    `json:"url,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type ContainerDeployment struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Status    string    `json:"status"`
	Instances int       `json:"instances"`
	CreatedAt time.Time `json:"created_at"`
}

type DatabaseDeployment struct {
	ID               string    `json:"id"`
	Endpoint         string    `json:"endpoint"`
	Port             int       `json:"port"`
	ConnectionString string    `json:"connection_string"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
}

type DatabaseBackup struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

type DNSRecord struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Value    string `json:"value"`
	TTL      int    `json:"ttl"`
	ZoneID   string `json:"zone_id"`
}

type SSLCertificate struct {
	ID       string    `json:"id"`
	Domain   string    `json:"domain"`
	Status   string    `json:"status"`
	IssuedAt time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type NetworkDeployment struct {
	ID              string   `json:"id"`
	VPCID           string   `json:"vpc_id"`
	PublicSubnetIDs []string `json:"public_subnet_ids"`
	PrivateSubnetIDs []string `json:"private_subnet_ids"`
	Status          string   `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

type Secret struct {
	Key       string    `json:"key"`
	Value     string    `json:"value,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CostData struct {
	ProjectID   string      `json:"project_id"`
	TotalCost   float64     `json:"total_cost"`
	Currency    string      `json:"currency"`
	TimeRange   TimeRange   `json:"time_range"`
	Breakdown   []CostItem  `json:"breakdown"`
}

type CostItem struct {
	Service string  `json:"service"`
	Cost    float64 `json:"cost"`
	Usage   string  `json:"usage"`
}

type CostBudget struct {
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	Threshold  float64 `json:"threshold"` // percentage
}

type CostAlert struct {
	ID        string    `json:"id"`
	Message   string    `json:"message"`
	Severity  string    `json:"severity"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type MetricData struct {
	MetricName string        `json:"metric_name"`
	DataPoints []DataPoint   `json:"data_points"`
	TimeRange  TimeRange     `json:"time_range"`
}

type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
}

type LogData struct {
	Events    []LogEvent `json:"events"`
	NextToken string     `json:"next_token,omitempty"`
	TimeRange TimeRange  `json:"time_range"`
}

type LogEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	Level     string    `json:"level"`
	Source    string    `json:"source"`
}

type AlertConfig struct {
	Name        string                 `json:"name"`
	MetricName  string                 `json:"metric_name"`
	Threshold   float64                `json:"threshold"`
	Comparison  string                 `json:"comparison"` // GreaterThan, LessThan, etc.
	Actions     []AlertAction          `json:"actions"`
	Tags        map[string]string      `json:"tags,omitempty"`
}

type AlertAction struct {
	Type   string            `json:"type"` // email, webhook, sms
	Target string            `json:"target"`
	Config map[string]string `json:"config,omitempty"`
}

type Alert struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Region struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Available   bool   `json:"available"`
}

type FunctionResponse struct {
	StatusCode int               `json:"status_code"`
	Body       string            `json:"body"`
	Headers    map[string]string `json:"headers"`
	Logs       string            `json:"logs"`
}