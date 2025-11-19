package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mikrocloud/mikrocloud/internal/domain/proxy"
	"github.com/mikrocloud/mikrocloud/pkg/containers/manager"
	"gopkg.in/yaml.v3"
)

type TraefikService struct {
	containerManager manager.ContainerManager
	configDir        string
	containerName    string
	containerID      string
	isRunning        bool
	networkMode      string
}

type TraefikConfig struct {
	Global      *proxy.TraefikGlobalConfig `json:"global"`
	HTTP        *HTTPConfig                `json:"http,omitempty"`
	EntryPoints map[string]EntryPoint      `json:"entryPoints"`
	Providers   *ProvidersConfig           `json:"providers,omitempty"`
	API         APIConfig                  `json:"api"`
	Log         LogConfig                  `json:"log"`
	AccessLog   AccessLogConfig            `json:"accessLog"`
}

type ProvidersConfig struct {
	Docker *DockerProviderConfig `json:"docker,omitempty"`
	File   *FileProviderConfig   `json:"file,omitempty"`
}

type DockerProviderConfig struct {
	Endpoint         string `json:"endpoint,omitempty"`
	ExposedByDefault bool   `json:"exposedByDefault"`
}

type FileProviderConfig struct {
	Directory string `json:"directory,omitempty"`
	Watch     bool   `json:"watch,omitempty"`
}

type HTTPConfig struct {
	Routers     map[string]Router     `json:"routers"`
	Services    map[string]Service    `json:"services"`
	Middlewares map[string]Middleware `json:"middlewares,omitempty"`
}

type Router struct {
	Rule        string     `json:"rule"`
	Service     string     `json:"service"`
	Middlewares []string   `json:"middlewares,omitempty"`
	TLS         *RouterTLS `json:"tls,omitempty"`
	Priority    int        `json:"priority,omitempty"`
}

type RouterTLS struct {
	CertResolver string                 `json:"certResolver,omitempty"`
	Domains      []TLSDomain            `json:"domains,omitempty"`
	Options      map[string]interface{} `json:"options,omitempty"`
}

type TLSDomain struct {
	Main string   `json:"main"`
	SANs []string `json:"sans,omitempty"`
}

type Service struct {
	LoadBalancer LoadBalancer `json:"loadBalancer"`
}

type LoadBalancer struct {
	Servers     []Server     `json:"servers"`
	HealthCheck *HealthCheck `json:"healthCheck,omitempty"`
	Sticky      *Sticky      `json:"sticky,omitempty"`
}

type Server struct {
	URL    string `json:"url"`
	Weight int    `json:"weight,omitempty"`
}

type HealthCheck struct {
	Path     string `json:"path"`
	Interval string `json:"interval"`
	Timeout  string `json:"timeout"`
	Retries  int    `json:"retries,omitempty"`
}

type Sticky struct {
	Cookie *StickyCookie `json:"cookie,omitempty"`
}

type StickyCookie struct {
	Name string `json:"name"`
}

type Middleware struct {
	StripPrefix *StripPrefixMiddleware `json:"stripPrefix,omitempty"`
	AddPrefix   *AddPrefixMiddleware   `json:"addPrefix,omitempty"`
	Headers     *HeadersMiddleware     `json:"headers,omitempty"`
	RateLimit   *RateLimitMiddleware   `json:"rateLimit,omitempty"`
	BasicAuth   *BasicAuthMiddleware   `json:"basicAuth,omitempty"`
	Compress    *CompressMiddleware    `json:"compress,omitempty"`
	CORS        *CORSMiddleware        `json:"cors,omitempty"`
}

type StripPrefixMiddleware struct {
	Prefixes []string `json:"prefixes"`
}

type AddPrefixMiddleware struct {
	Prefix string `json:"prefix"`
}

type HeadersMiddleware struct {
	CustomRequestHeaders  map[string]string `json:"customRequestHeaders,omitempty"`
	CustomResponseHeaders map[string]string `json:"customResponseHeaders,omitempty"`
}

type RateLimitMiddleware struct {
	Average int `json:"average"`
	Burst   int `json:"burst,omitempty"`
}

type BasicAuthMiddleware struct {
	Users []string `json:"users"`
}

type CompressMiddleware struct{}

type CORSMiddleware struct {
	AccessControlAllowOriginList []string `json:"accessControlAllowOriginList,omitempty"`
	AccessControlAllowMethods    []string `json:"accessControlAllowMethods,omitempty"`
	AccessControlAllowHeaders    []string `json:"accessControlAllowHeaders,omitempty"`
}

type EntryPoint struct {
	Address string    `json:"address"`
	TLS     *EntryTLS `json:"tls,omitempty"`
}

type EntryTLS struct {
	CertResolver string                 `json:"certResolver,omitempty"`
	Domains      []TLSDomain            `json:"domains,omitempty"`
	Options      map[string]interface{} `json:"options,omitempty"`
}

type APIConfig struct {
	Dashboard bool `json:"dashboard"`
	Debug     bool `json:"debug,omitempty"`
	Insecure  bool `json:"insecure,omitempty"`
}

type LogConfig struct {
	Level  string `json:"level"`
	Format string `json:"format,omitempty"`
}

type AccessLogConfig struct {
	FilePath string `json:"filePath,omitempty"`
	Format   string `json:"format,omitempty"`
}

func NewTraefikService(containerManager manager.ContainerManager, configDir string, networkMode string) *TraefikService {
	if networkMode == "" {
		networkMode = "bridge"
	}
	return &TraefikService{
		containerManager: containerManager,
		configDir:        configDir,
		containerName:    "mikrocloud-traefik",
		isRunning:        false,
		networkMode:      networkMode,
	}
}

func (ts *TraefikService) Start(ctx context.Context, globalConfig *proxy.TraefikGlobalConfig) error {
	if ts.isRunning {
		return nil
	}

	if err := ts.ensureConfigDir(); err != nil {
		return fmt.Errorf("failed to ensure config directory: %w", err)
	}

	if err := ts.writeGlobalConfig(globalConfig); err != nil {
		return fmt.Errorf("failed to write global config: %w", err)
	}

	// Check if container already exists
	containers, err := ts.containerManager.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	for _, container := range containers {
		if container.Name == ts.containerName {
			// Container exists, check its state
			if container.State == "running" {
				ts.containerID = container.ID
				ts.isRunning = true
				return nil
			}
			// Container exists but not running, remove it and recreate
			_ = ts.containerManager.Delete(ctx, container.ID)
			break
		}
	}

	image := "traefik:v3.0"
	if err := ts.containerManager.PullImage(ctx, image); err != nil {
		return fmt.Errorf("failed to pull Traefik image: %w", err)
	}

	containerConfig := manager.ContainerConfig{
		Name:          ts.containerName,
		Image:         image,
		RestartPolicy: "unless-stopped",
		NetworkMode:   ts.networkMode,
		Ports: map[string]string{
			"80":   "80/tcp",
			"443":  "443/tcp",
			"8080": "8080/tcp", // Dashboard
		},
		Volumes: map[string]string{
			ts.configDir:           "/etc/traefik",
			"/var/run/docker.sock": "/var/run/docker.sock:ro",
		},
		Environment: map[string]string{
			"TRAEFIK_CONFIGFILE": "/etc/traefik/traefik.yml",
		},
		Command: []string{
			"--configfile=/etc/traefik/traefik.yml",
		},
	}

	containerID, err := ts.containerManager.Create(ctx, containerConfig)
	if err != nil {
		return fmt.Errorf("failed to create Traefik container: %w", err)
	}

	if err := ts.containerManager.Start(ctx, containerID); err != nil {
		return fmt.Errorf("failed to start Traefik container: %w", err)
	}

	ts.containerID = containerID
	ts.isRunning = true

	return nil
}

func (ts *TraefikService) Stop(ctx context.Context) error {
	if !ts.isRunning || ts.containerID == "" {
		return nil
	}

	if err := ts.containerManager.Stop(ctx, ts.containerID); err != nil {
		return fmt.Errorf("failed to stop Traefik container: %w", err)
	}

	ts.isRunning = false
	return nil
}

func (ts *TraefikService) Restart(ctx context.Context, globalConfig *proxy.TraefikGlobalConfig) error {
	if err := ts.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop Traefik: %w", err)
	}

	if err := ts.Start(ctx, globalConfig); err != nil {
		return fmt.Errorf("failed to start Traefik: %w", err)
	}

	return nil
}

func (ts *TraefikService) UpdateDynamicConfig(ctx context.Context, configs []*proxy.ProxyConfig) error {
	if !ts.isRunning {
		return fmt.Errorf("Traefik is not running")
	}

	httpConfig := &HTTPConfig{
		Routers:  make(map[string]Router),
		Services: make(map[string]Service),
	}

	for _, config := range configs {
		if config.Status() != proxy.ProxyStatusActive {
			continue
		}

		routerName := config.GetRouterName()
		serviceName := config.GetServiceName()

		router := Router{
			Rule:    config.GetRuleHost(),
			Service: serviceName,
		}

		if config.PathPrefix() != "" {
			router.Rule = fmt.Sprintf("%s && PathPrefix(`%s`)", router.Rule, config.PathPrefix())
		}

		if config.Protocol() == proxy.ProxyProtocolHTTPS || (config.TLS() != nil && config.TLS().Enabled()) {
			router.TLS = &RouterTLS{}
		}

		httpConfig.Routers[routerName] = router

		service := Service{
			LoadBalancer: LoadBalancer{
				Servers: []Server{
					{
						URL: config.TargetURL(),
					},
				},
			},
		}

		if config.HealthCheck() != nil && config.HealthCheck().Enabled() {
			service.LoadBalancer.HealthCheck = &HealthCheck{
				Path:     config.HealthCheck().Path(),
				Interval: config.HealthCheck().Interval().String(),
				Timeout:  config.HealthCheck().Timeout().String(),
				Retries:  config.HealthCheck().Retries(),
			}
		}

		httpConfig.Services[serviceName] = service
	}

	return ts.writeDynamicConfig(httpConfig)
}

func (ts *TraefikService) IsRunning() bool {
	return ts.isRunning
}

func (ts *TraefikService) GetContainerID() string {
	return ts.containerID
}

func (ts *TraefikService) GetDashboardURL() string {
	return "http://localhost:8080"
}

func (ts *TraefikService) ensureConfigDir() error {
	if err := os.MkdirAll(ts.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	dynamicDir := filepath.Join(ts.configDir, "dynamic")
	if err := os.MkdirAll(dynamicDir, 0755); err != nil {
		return fmt.Errorf("failed to create dynamic config directory: %w", err)
	}

	return nil
}

func (ts *TraefikService) writeGlobalConfig(globalConfig *proxy.TraefikGlobalConfig) error {
	config := TraefikConfig{
		EntryPoints: map[string]EntryPoint{
			"web": {
				Address: ":80",
			},
			"websecure": {
				Address: ":443",
			},
		},
		Providers: &ProvidersConfig{
			Docker: &DockerProviderConfig{
				Endpoint:         "unix:///var/run/docker.sock",
				ExposedByDefault: false,
			},
			File: &FileProviderConfig{
				Directory: "/etc/traefik/dynamic",
				Watch:     true,
			},
		},
		API: APIConfig{
			Dashboard: globalConfig.API().Dashboard(),
			Debug:     globalConfig.API().Debug(),
			Insecure:  true,
		},
		Log: LogConfig{
			Level: "INFO",
		},
		AccessLog: AccessLogConfig{
			Format: "json",
		},
	}

	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal global config: %w", err)
	}

	configPath := filepath.Join(ts.configDir, "traefik.yml")
	if err := os.WriteFile(configPath, configBytes, 0644); err != nil {
		return fmt.Errorf("failed to write global config: %w", err)
	}

	return nil
}

func (ts *TraefikService) writeDynamicConfig(httpConfig *HTTPConfig) error {
	dynamicConfig := map[string]interface{}{
		"http": httpConfig,
	}

	configBytes, err := json.MarshalIndent(dynamicConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal dynamic config: %w", err)
	}

	configPath := filepath.Join(ts.configDir, "dynamic", "mikrocloud.json")
	if err := os.WriteFile(configPath, configBytes, 0644); err != nil {
		return fmt.Errorf("failed to write dynamic config: %w", err)
	}

	return nil
}

func (ts *TraefikService) RemoveDynamicConfig() error {
	configPath := filepath.Join(ts.configDir, "dynamic", "mikrocloud.json")
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove dynamic config: %w", err)
	}
	return nil
}
