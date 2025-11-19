package cloudprovider

import (
	"fmt"

	"github.com/mikrocloud/mikrocloud/internal/domain"
)

// CloudProviderFactory creates cloud provider instances
type CloudProviderFactory struct{}

// NewCloudProviderFactory creates a new cloud provider factory
func NewCloudProviderFactory() *CloudProviderFactory {
	return &CloudProviderFactory{}
}

// Create creates a new cloud provider instance based on the provider type
func (f *CloudProviderFactory) Create(providerType domain.CloudProvider, credentials map[string]string) (CloudProvider, error) {
	switch providerType {
	case domain.CloudProviderAWS:
		return NewAWSProvider(credentials)
	case domain.CloudProviderAzure:
		return NewAzureProvider(credentials)
	case domain.CloudProviderGCP:
		return NewGCPProvider(credentials)
	default:
		return nil, fmt.Errorf("unsupported cloud provider: %s", providerType)
	}
}

// GetSupportedProviders returns a list of supported cloud providers
func (f *CloudProviderFactory) GetSupportedProviders() []domain.CloudProvider {
	return []domain.CloudProvider{
		domain.CloudProviderAWS,
		domain.CloudProviderAzure,
		domain.CloudProviderGCP,
	}
}

// Credentials represents cloud provider credentials
type Credentials struct {
	AWS   *AWSCredentials   `json:"aws,omitempty"`
	Azure *AzureCredentials `json:"azure,omitempty"`
	GCP   *GCPCredentials   `json:"gcp,omitempty"`
}

type AWSCredentials struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	SessionToken    string `json:"session_token,omitempty"`
	Region          string `json:"region"`
}

type AzureCredentials struct {
	SubscriptionID string `json:"subscription_id"`
	TenantID       string `json:"tenant_id"`
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	Location       string `json:"location"`
}

type GCPCredentials struct {
	ProjectID      string `json:"project_id"`
	ServiceAccount string `json:"service_account"` // Base64 encoded service account JSON
	Region         string `json:"region"`
}

// ProviderCapabilities defines what each provider supports
type ProviderCapabilities struct {
	StaticSites        bool     `json:"static_sites"`
	ServerlessFunctions bool     `json:"serverless_functions"`
	Containers         bool     `json:"containers"`
	Databases          []string `json:"databases"`
	Networking         bool     `json:"networking"`
	Monitoring         bool     `json:"monitoring"`
	CostManagement     bool     `json:"cost_management"`
}

// GetProviderCapabilities returns the capabilities of a cloud provider
func GetProviderCapabilities(provider domain.CloudProvider) ProviderCapabilities {
	capabilities := map[domain.CloudProvider]ProviderCapabilities{
		domain.CloudProviderAWS: {
			StaticSites:         true,
			ServerlessFunctions: true,
			Containers:          true,
			Databases:          []string{"postgresql", "mysql", "mongodb", "redis", "dynamodb"},
			Networking:          true,
			Monitoring:          true,
			CostManagement:      true,
		},
		domain.CloudProviderAzure: {
			StaticSites:         true,
			ServerlessFunctions: true,
			Containers:          true,
			Databases:          []string{"postgresql", "mysql", "mongodb", "redis", "cosmosdb"},
			Networking:          true,
			Monitoring:          true,
			CostManagement:      true,
		},
		domain.CloudProviderGCP: {
			StaticSites:         true,
			ServerlessFunctions: true,
			Containers:          true,
			Databases:          []string{"postgresql", "mysql", "mongodb", "redis", "firestore"},
			Networking:          true,
			Monitoring:          true,
			CostManagement:      true,
		},
	}
	
	return capabilities[provider]
}

// Service mapping for cross-cloud compatibility
var ServiceMapping = map[domain.CloudProvider]map[string]string{
	domain.CloudProviderAWS: {
		"static_hosting":       "s3+cloudfront",
		"serverless_functions": "lambda",
		"containers":          "ecs-fargate",
		"sql_database":        "rds",
		"nosql_database":      "dynamodb",
		"cache":               "elasticache",
		"message_queue":       "sqs",
		"api_gateway":         "api-gateway",
		"cdn":                 "cloudfront",
		"dns":                 "route53",
		"ssl":                 "acm",
		"monitoring":          "cloudwatch",
		"build":               "codebuild",
	},
	domain.CloudProviderAzure: {
		"static_hosting":       "storage+cdn",
		"serverless_functions": "functions",
		"containers":          "container-apps",
		"sql_database":        "sql-database",
		"nosql_database":      "cosmosdb",
		"cache":               "redis",
		"message_queue":       "service-bus",
		"api_gateway":         "api-management",
		"cdn":                 "cdn",
		"dns":                 "dns-zone",
		"ssl":                 "key-vault",
		"monitoring":          "monitor",
		"build":               "devops",
	},
	domain.CloudProviderGCP: {
		"static_hosting":       "storage+cdn",
		"serverless_functions": "functions",
		"containers":          "cloud-run",
		"sql_database":        "sql",
		"nosql_database":      "firestore",
		"cache":               "memorystore",
		"message_queue":       "pubsub",
		"api_gateway":         "api-gateway",
		"cdn":                 "cdn",
		"dns":                 "dns",
		"ssl":                 "ssl-certificates",
		"monitoring":          "monitoring",
		"build":               "build",
	},
}