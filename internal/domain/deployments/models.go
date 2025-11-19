package deployments

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/applications"
	"github.com/mikrocloud/mikrocloud/internal/domain/users"
)

type Deployment struct {
	id                    DeploymentID
	applicationID         applications.ApplicationID
	deploymentNumber      int
	isProduction          bool
	triggeredBy           *users.UserID
	triggerType           TriggerType
	status                DeploymentStatus
	containerID           string
	imageTag              string
	imageDigest           string
	gitCommitHash         string
	gitCommitMessage      string
	gitBranch             string
	gitAuthorName         string
	buildLogs             string
	deployLogs            string
	errorMessage          string
	startedAt             time.Time
	buildStartedAt        *time.Time
	buildCompletedAt      *time.Time
	deployStartedAt       *time.Time
	deployCompletedAt     *time.Time
	stoppedAt             *time.Time
	buildDurationSeconds  *int
	deployDurationSeconds *int
	updatedAt             time.Time
}

type DeploymentID struct {
	value string
}

func NewDeploymentID() DeploymentID {
	return DeploymentID{value: uuid.New().String()}
}

func DeploymentIDFromString(s string) (DeploymentID, error) {
	if s == "" {
		return DeploymentID{}, fmt.Errorf("deployment ID cannot be empty")
	}
	return DeploymentID{value: s}, nil
}

func (id DeploymentID) String() string {
	return id.value
}

type TriggerType string

const (
	TriggerTypeManual    TriggerType = "manual"
	TriggerTypeGitPush   TriggerType = "git_push"
	TriggerTypeAPI       TriggerType = "api"
	TriggerTypeScheduled TriggerType = "scheduled"
	TriggerTypeRollback  TriggerType = "rollback"
)

type DeploymentStatus string

const (
	DeploymentStatusPending   DeploymentStatus = "pending"
	DeploymentStatusQueued    DeploymentStatus = "queued"
	DeploymentStatusBuilding  DeploymentStatus = "building"
	DeploymentStatusDeploying DeploymentStatus = "deploying"
	DeploymentStatusRunning   DeploymentStatus = "running"
	DeploymentStatusFailed    DeploymentStatus = "failed"
	DeploymentStatusCancelled DeploymentStatus = "cancelled"
	DeploymentStatusStopped   DeploymentStatus = "stopped"
)

func NewDeployment(
	applicationID applications.ApplicationID,
	deploymentNumber int,
	isProduction bool,
	triggeredBy *users.UserID,
	triggerType TriggerType,
	imageTag string,
) *Deployment {
	now := time.Now()
	return &Deployment{
		id:               NewDeploymentID(),
		applicationID:    applicationID,
		deploymentNumber: deploymentNumber,
		isProduction:     isProduction,
		triggeredBy:      triggeredBy,
		triggerType:      triggerType,
		status:           DeploymentStatusPending,
		imageTag:         imageTag,
		startedAt:        now,
		updatedAt:        now,
	}
}

func (d *Deployment) ID() DeploymentID {
	return d.id
}

func (d *Deployment) ApplicationID() applications.ApplicationID {
	return d.applicationID
}

func (d *Deployment) DeploymentNumber() int {
	return d.deploymentNumber
}

func (d *Deployment) IsProduction() bool {
	return d.isProduction
}

func (d *Deployment) TriggeredBy() *users.UserID {
	return d.triggeredBy
}

func (d *Deployment) TriggerType() TriggerType {
	return d.triggerType
}

func (d *Deployment) Status() DeploymentStatus {
	return d.status
}

func (d *Deployment) ContainerID() string {
	return d.containerID
}

func (d *Deployment) ImageTag() string {
	return d.imageTag
}

func (d *Deployment) ImageDigest() string {
	return d.imageDigest
}

func (d *Deployment) GitCommitHash() string {
	return d.gitCommitHash
}

func (d *Deployment) GitCommitMessage() string {
	return d.gitCommitMessage
}

func (d *Deployment) GitBranch() string {
	return d.gitBranch
}

func (d *Deployment) GitAuthorName() string {
	return d.gitAuthorName
}

func (d *Deployment) BuildLogs() string {
	return d.buildLogs
}

func (d *Deployment) DeployLogs() string {
	return d.deployLogs
}

func (d *Deployment) ErrorMessage() string {
	return d.errorMessage
}

func (d *Deployment) StartedAt() time.Time {
	return d.startedAt
}

func (d *Deployment) BuildStartedAt() *time.Time {
	return d.buildStartedAt
}

func (d *Deployment) BuildCompletedAt() *time.Time {
	return d.buildCompletedAt
}

func (d *Deployment) DeployStartedAt() *time.Time {
	return d.deployStartedAt
}

func (d *Deployment) DeployCompletedAt() *time.Time {
	return d.deployCompletedAt
}

func (d *Deployment) StoppedAt() *time.Time {
	return d.stoppedAt
}

func (d *Deployment) BuildDurationSeconds() *int {
	return d.buildDurationSeconds
}

func (d *Deployment) DeployDurationSeconds() *int {
	return d.deployDurationSeconds
}

func (d *Deployment) UpdatedAt() time.Time {
	return d.updatedAt
}

func (d *Deployment) ChangeStatus(status DeploymentStatus) {
	d.status = status
	d.updatedAt = time.Now()
}

func (d *Deployment) SetContainerID(containerID string) {
	d.containerID = containerID
	d.updatedAt = time.Now()
}

func (d *Deployment) SetImageDigest(digest string) {
	d.imageDigest = digest
	d.updatedAt = time.Now()
}

func (d *Deployment) SetGitInfo(commitHash, commitMessage, branch, authorName string) {
	d.gitCommitHash = commitHash
	d.gitCommitMessage = commitMessage
	d.gitBranch = branch
	d.gitAuthorName = authorName
	d.updatedAt = time.Now()
}

func (d *Deployment) StartBuild() {
	now := time.Now()
	d.status = DeploymentStatusBuilding
	d.buildStartedAt = &now
	d.updatedAt = now
}

func (d *Deployment) CompleteBuild() {
	now := time.Now()
	d.buildCompletedAt = &now
	if d.buildStartedAt != nil {
		duration := int(now.Sub(*d.buildStartedAt).Seconds())
		d.buildDurationSeconds = &duration
	}
	d.updatedAt = now
}

func (d *Deployment) StartDeploy() {
	now := time.Now()
	d.status = DeploymentStatusDeploying
	d.deployStartedAt = &now
	d.updatedAt = now
}

func (d *Deployment) CompleteDeploy() {
	now := time.Now()
	d.status = DeploymentStatusRunning
	d.deployCompletedAt = &now
	if d.deployStartedAt != nil {
		duration := int(now.Sub(*d.deployStartedAt).Seconds())
		d.deployDurationSeconds = &duration
	}
	d.updatedAt = now
}

func (d *Deployment) Fail(errorMessage string) {
	d.status = DeploymentStatusFailed
	d.errorMessage = errorMessage
	d.updatedAt = time.Now()
}

func (d *Deployment) Cancel() {
	d.status = DeploymentStatusCancelled
	d.updatedAt = time.Now()
}

func (d *Deployment) Stop() {
	now := time.Now()
	d.status = DeploymentStatusStopped
	d.stoppedAt = &now
	d.updatedAt = now
}

func (d *Deployment) AppendBuildLogs(logs string) {
	d.buildLogs += logs
	d.updatedAt = time.Now()
}

func (d *Deployment) AppendDeployLogs(logs string) {
	d.deployLogs += logs
	d.updatedAt = time.Now()
}

func ReconstructDeployment(
	id DeploymentID,
	applicationID applications.ApplicationID,
	deploymentNumber int,
	isProduction bool,
	triggeredBy *users.UserID,
	triggerType TriggerType,
	status DeploymentStatus,
	containerID, imageTag, imageDigest string,
	gitCommitHash, gitCommitMessage, gitBranch, gitAuthorName string,
	buildLogs, deployLogs, errorMessage string,
	startedAt time.Time,
	buildStartedAt, buildCompletedAt, deployStartedAt, deployCompletedAt, stoppedAt *time.Time,
	buildDurationSeconds, deployDurationSeconds *int,
	updatedAt time.Time,
) *Deployment {
	return &Deployment{
		id:                    id,
		applicationID:         applicationID,
		deploymentNumber:      deploymentNumber,
		isProduction:          isProduction,
		triggeredBy:           triggeredBy,
		triggerType:           triggerType,
		status:                status,
		containerID:           containerID,
		imageTag:              imageTag,
		imageDigest:           imageDigest,
		gitCommitHash:         gitCommitHash,
		gitCommitMessage:      gitCommitMessage,
		gitBranch:             gitBranch,
		gitAuthorName:         gitAuthorName,
		buildLogs:             buildLogs,
		deployLogs:            deployLogs,
		errorMessage:          errorMessage,
		startedAt:             startedAt,
		buildStartedAt:        buildStartedAt,
		buildCompletedAt:      buildCompletedAt,
		deployStartedAt:       deployStartedAt,
		deployCompletedAt:     deployCompletedAt,
		stoppedAt:             stoppedAt,
		buildDurationSeconds:  buildDurationSeconds,
		deployDurationSeconds: deployDurationSeconds,
		updatedAt:             updatedAt,
	}
}
