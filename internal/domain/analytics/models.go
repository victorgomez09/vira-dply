package analytics

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Metric struct {
	id        MetricID
	projectID uuid.UUID
	serviceID *uuid.UUID
	name      MetricName
	value     float64
	unit      MetricUnit
	tags      map[string]string
	timestamp time.Time
	createdAt time.Time
}

type MetricID struct {
	value string
}

func NewMetricID() MetricID {
	return MetricID{value: uuid.New().String()}
}

func MetricIDFromString(id string) MetricID {
	return MetricID{value: id}
}

func (id MetricID) String() string {
	return id.value
}

type MetricName struct {
	value string
}

func NewMetricName(name string) (MetricName, error) {
	if name == "" {
		return MetricName{}, fmt.Errorf("metric name cannot be empty")
	}
	if len(name) > 128 {
		return MetricName{}, fmt.Errorf("metric name cannot exceed 128 characters")
	}
	return MetricName{value: name}, nil
}

func (n MetricName) String() string {
	return n.value
}

type MetricUnit string

const (
	MetricUnitBytes          MetricUnit = "bytes"
	MetricUnitPercent        MetricUnit = "percent"
	MetricUnitCount          MetricUnit = "count"
	MetricUnitSeconds        MetricUnit = "seconds"
	MetricUnitMilliseconds   MetricUnit = "milliseconds"
	MetricUnitRequestsSecond MetricUnit = "requests_per_second"
	MetricUnitBytesSecond    MetricUnit = "bytes_per_second"
)

type Report struct {
	id          ReportID
	projectID   uuid.UUID
	name        ReportName
	description string
	reportType  ReportType
	period      ReportPeriod
	config      ReportConfig
	status      ReportStatus
	generatedAt *time.Time
	createdAt   time.Time
	updatedAt   time.Time
}

type ReportID struct {
	value string
}

func NewReportID() ReportID {
	return ReportID{value: uuid.New().String()}
}

func (id ReportID) String() string {
	return id.value
}

type ReportName struct {
	value string
}

func NewReportName(name string) (ReportName, error) {
	if name == "" {
		return ReportName{}, fmt.Errorf("report name cannot be empty")
	}
	if len(name) > 128 {
		return ReportName{}, fmt.Errorf("report name cannot exceed 128 characters")
	}
	return ReportName{value: name}, nil
}

func (n ReportName) String() string {
	return n.value
}

type ReportType string

const (
	ReportTypeUsage       ReportType = "usage"
	ReportTypePerformance ReportType = "performance"
	ReportTypeCost        ReportType = "cost"
	ReportTypeCustom      ReportType = "custom"
)

type ReportPeriod string

const (
	ReportPeriodHourly  ReportPeriod = "hourly"
	ReportPeriodDaily   ReportPeriod = "daily"
	ReportPeriodWeekly  ReportPeriod = "weekly"
	ReportPeriodMonthly ReportPeriod = "monthly"
)

type ReportConfig struct {
	Metrics   []string          `json:"metrics"`
	Services  []string          `json:"services,omitempty"`
	StartTime *time.Time        `json:"start_time,omitempty"`
	EndTime   *time.Time        `json:"end_time,omitempty"`
	Filters   map[string]string `json:"filters,omitempty"`
}

type ReportStatus string

const (
	ReportStatusPending    ReportStatus = "pending"
	ReportStatusGenerating ReportStatus = "generating"
	ReportStatusCompleted  ReportStatus = "completed"
	ReportStatusError      ReportStatus = "error"
)

type Dashboard struct {
	id        DashboardID
	projectID uuid.UUID
	name      DashboardName
	widgets   []DashboardWidget
	layout    DashboardLayout
	isDefault bool
	createdAt time.Time
	updatedAt time.Time
}

type DashboardID struct {
	value string
}

func NewDashboardID() DashboardID {
	return DashboardID{value: uuid.New().String()}
}

func (id DashboardID) String() string {
	return id.value
}

type DashboardName struct {
	value string
}

func NewDashboardName(name string) (DashboardName, error) {
	if name == "" {
		return DashboardName{}, fmt.Errorf("dashboard name cannot be empty")
	}
	if len(name) > 128 {
		return DashboardName{}, fmt.Errorf("dashboard name cannot exceed 128 characters")
	}
	return DashboardName{value: name}, nil
}

func (n DashboardName) String() string {
	return n.value
}

type DashboardWidget struct {
	ID         string           `json:"id"`
	Type       WidgetType       `json:"type"`
	Title      string           `json:"title"`
	Position   WidgetPosition   `json:"position"`
	Size       WidgetSize       `json:"size"`
	Config     map[string]any   `json:"config"`
	DataSource WidgetDataSource `json:"data_source"`
}

type WidgetType string

const (
	WidgetTypeChart   WidgetType = "chart"
	WidgetTypeGauge   WidgetType = "gauge"
	WidgetTypeCounter WidgetType = "counter"
	WidgetTypeTable   WidgetType = "table"
	WidgetTypeText    WidgetType = "text"
)

type WidgetPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type WidgetSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type WidgetDataSource struct {
	Type    string         `json:"type"`
	Query   string         `json:"query"`
	Params  map[string]any `json:"params,omitempty"`
	Refresh string         `json:"refresh,omitempty"`
}

type DashboardLayout struct {
	Columns int    `json:"columns"`
	Type    string `json:"type"`
}

type MetricQuery struct {
	ProjectID uuid.UUID
	ServiceID *uuid.UUID
	Name      *MetricName
	StartTime *time.Time
	EndTime   *time.Time
	Tags      map[string]string
	Limit     int
	Offset    int
}

func NewMetric(
	projectID uuid.UUID,
	serviceID *uuid.UUID,
	name MetricName,
	value float64,
	unit MetricUnit,
	tags map[string]string,
	timestamp time.Time,
) *Metric {
	if tags == nil {
		tags = make(map[string]string)
	}

	return &Metric{
		id:        NewMetricID(),
		projectID: projectID,
		serviceID: serviceID,
		name:      name,
		value:     value,
		unit:      unit,
		tags:      tags,
		timestamp: timestamp,
		createdAt: time.Now(),
	}
}

func (m *Metric) ID() MetricID {
	return m.id
}

func (m *Metric) ProjectID() uuid.UUID {
	return m.projectID
}

func (m *Metric) ServiceID() *uuid.UUID {
	return m.serviceID
}

func (m *Metric) Name() MetricName {
	return m.name
}

func (m *Metric) Value() float64 {
	return m.value
}

func (m *Metric) Unit() MetricUnit {
	return m.unit
}

func (m *Metric) Tags() map[string]string {
	return m.tags
}

func (m *Metric) Timestamp() time.Time {
	return m.timestamp
}

func (m *Metric) CreatedAt() time.Time {
	return m.createdAt
}

func NewReport(
	projectID uuid.UUID,
	name ReportName,
	description string,
	reportType ReportType,
	period ReportPeriod,
	config ReportConfig,
) *Report {
	now := time.Now()
	return &Report{
		id:          NewReportID(),
		projectID:   projectID,
		name:        name,
		description: description,
		reportType:  reportType,
		period:      period,
		config:      config,
		status:      ReportStatusPending,
		createdAt:   now,
		updatedAt:   now,
	}
}

func (r *Report) ID() ReportID {
	return r.id
}

func (r *Report) ProjectID() uuid.UUID {
	return r.projectID
}

func (r *Report) Name() ReportName {
	return r.name
}

func (r *Report) Description() string {
	return r.description
}

func (r *Report) Type() ReportType {
	return r.reportType
}

func (r *Report) Period() ReportPeriod {
	return r.period
}

func (r *Report) Config() ReportConfig {
	return r.config
}

func (r *Report) Status() ReportStatus {
	return r.status
}

func (r *Report) GeneratedAt() *time.Time {
	return r.generatedAt
}

func (r *Report) CreatedAt() time.Time {
	return r.createdAt
}

func (r *Report) UpdatedAt() time.Time {
	return r.updatedAt
}

func (r *Report) ChangeStatus(status ReportStatus) {
	r.status = status
	r.updatedAt = time.Now()
	if status == ReportStatusCompleted {
		now := time.Now()
		r.generatedAt = &now
	}
}

func (r *Report) UpdateConfig(config ReportConfig) {
	r.config = config
	r.updatedAt = time.Now()
}

func NewDashboard(
	projectID uuid.UUID,
	name DashboardName,
	isDefault bool,
) *Dashboard {
	now := time.Now()
	return &Dashboard{
		id:        NewDashboardID(),
		projectID: projectID,
		name:      name,
		widgets:   []DashboardWidget{},
		layout:    DashboardLayout{Columns: 12, Type: "grid"},
		isDefault: isDefault,
		createdAt: now,
		updatedAt: now,
	}
}

func (d *Dashboard) ID() DashboardID {
	return d.id
}

func (d *Dashboard) ProjectID() uuid.UUID {
	return d.projectID
}

func (d *Dashboard) Name() DashboardName {
	return d.name
}

func (d *Dashboard) Widgets() []DashboardWidget {
	return d.widgets
}

func (d *Dashboard) Layout() DashboardLayout {
	return d.layout
}

func (d *Dashboard) IsDefault() bool {
	return d.isDefault
}

func (d *Dashboard) CreatedAt() time.Time {
	return d.createdAt
}

func (d *Dashboard) UpdatedAt() time.Time {
	return d.updatedAt
}

func (d *Dashboard) AddWidget(widget DashboardWidget) {
	d.widgets = append(d.widgets, widget)
	d.updatedAt = time.Now()
}

func (d *Dashboard) RemoveWidget(widgetID string) {
	for i, widget := range d.widgets {
		if widget.ID == widgetID {
			d.widgets = append(d.widgets[:i], d.widgets[i+1:]...)
			d.updatedAt = time.Now()
			break
		}
	}
}

func (d *Dashboard) UpdateWidget(widgetID string, newWidget DashboardWidget) {
	for i, widget := range d.widgets {
		if widget.ID == widgetID {
			d.widgets[i] = newWidget
			d.updatedAt = time.Now()
			break
		}
	}
}

func (d *Dashboard) SetAsDefault() {
	d.isDefault = true
	d.updatedAt = time.Now()
}

func (d *Dashboard) UnsetAsDefault() {
	d.isDefault = false
	d.updatedAt = time.Now()
}

func ReconstructMetric(
	id MetricID,
	projectID uuid.UUID,
	serviceID *uuid.UUID,
	name MetricName,
	value float64,
	unit MetricUnit,
	tags map[string]string,
	timestamp, createdAt time.Time,
) *Metric {
	return &Metric{
		id:        id,
		projectID: projectID,
		serviceID: serviceID,
		name:      name,
		value:     value,
		unit:      unit,
		tags:      tags,
		timestamp: timestamp,
		createdAt: createdAt,
	}
}

func ReconstructReport(
	id ReportID,
	projectID uuid.UUID,
	name ReportName,
	description string,
	reportType ReportType,
	period ReportPeriod,
	config ReportConfig,
	status ReportStatus,
	generatedAt *time.Time,
	createdAt, updatedAt time.Time,
) *Report {
	return &Report{
		id:          id,
		projectID:   projectID,
		name:        name,
		description: description,
		reportType:  reportType,
		period:      period,
		config:      config,
		status:      status,
		generatedAt: generatedAt,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

func ReconstructDashboard(
	id DashboardID,
	projectID uuid.UUID,
	name DashboardName,
	widgets []DashboardWidget,
	layout DashboardLayout,
	isDefault bool,
	createdAt, updatedAt time.Time,
) *Dashboard {
	return &Dashboard{
		id:        id,
		projectID: projectID,
		name:      name,
		widgets:   widgets,
		layout:    layout,
		isDefault: isDefault,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}
