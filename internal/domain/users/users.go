package users

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type User struct {
	id              UserID
	email           Email
	passwordHash    string
	name            string
	username        *Username
	status          UserStatus
	emailVerifiedAt *time.Time
	lastLoginAt     *time.Time
	timezone        string
	createdAt       time.Time
	updatedAt       time.Time
}

type UserID struct {
	value string
}

func NewUserID() UserID {
	return UserID{value: uuid.New().String()}
}

func UserIDFromString(s string) (UserID, error) {
	if s == "" {
		return UserID{}, fmt.Errorf("user ID cannot be empty")
	}
	return UserID{value: s}, nil
}

func (id UserID) String() string {
	return id.value
}

type Email struct {
	value string
}

func NewEmail(email string) (Email, error) {
	if email == "" {
		return Email{}, fmt.Errorf("email cannot be empty")
	}
	return Email{value: email}, nil
}

func (e Email) String() string {
	return e.value
}

type Username struct {
	value string
}

func NewUsername(username string) (*Username, error) {
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	return &Username{value: username}, nil
}

func (u *Username) String() string {
	if u == nil {
		return ""
	}
	return u.value
}

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusPending   UserStatus = "pending"
)

func NewUser(email Email, passwordHash string) *User {
	now := time.Now()
	return &User{
		id:           NewUserID(),
		email:        email,
		passwordHash: passwordHash,
		status:       UserStatusPending,
		timezone:     "UTC",
		createdAt:    now,
		updatedAt:    now,
	}
}

func NewUserWithName(email Email, passwordHash, name string) *User {
	now := time.Now()
	return &User{
		id:           NewUserID(),
		email:        email,
		passwordHash: passwordHash,
		name:         name,
		status:       UserStatusPending,
		timezone:     "UTC",
		createdAt:    now,
		updatedAt:    now,
	}
}

func (u *User) ID() UserID {
	return u.id
}

func (u *User) Email() Email {
	return u.email
}

func (u *User) PasswordHash() string {
	return u.passwordHash
}

func (u *User) Name() string {
	return u.name
}

func (u *User) Username() *Username {
	return u.username
}

func (u *User) Status() UserStatus {
	return u.status
}

func (u *User) EmailVerifiedAt() *time.Time {
	return u.emailVerifiedAt
}

func (u *User) LastLoginAt() *time.Time {
	return u.lastLoginAt
}

func (u *User) Timezone() string {
	return u.timezone
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

func (u *User) SetName(name string) {
	u.name = name
	u.updatedAt = time.Now()
}

func (u *User) SetUsername(username *Username) {
	u.username = username
	u.updatedAt = time.Now()
}

func (u *User) VerifyEmail() {
	now := time.Now()
	u.emailVerifiedAt = &now
	u.status = UserStatusActive
	u.updatedAt = now
}

func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.lastLoginAt = &now
	u.updatedAt = now
}

func (u *User) UpdatePassword(passwordHash string) {
	u.passwordHash = passwordHash
	u.updatedAt = time.Now()
}

func (u *User) ChangeStatus(status UserStatus) {
	u.status = status
	u.updatedAt = time.Now()
}

func (u *User) SetTimezone(timezone string) {
	u.timezone = timezone
	u.updatedAt = time.Now()
}

func ReconstructUser(
	id UserID,
	email Email,
	passwordHash string,
	name string,
	username *Username,
	status UserStatus,
	emailVerifiedAt *time.Time,
	lastLoginAt *time.Time,
	timezone string,
	createdAt time.Time,
	updatedAt time.Time,
) *User {
	return &User{
		id:              id,
		email:           email,
		passwordHash:    passwordHash,
		name:            name,
		username:        username,
		status:          status,
		emailVerifiedAt: emailVerifiedAt,
		lastLoginAt:     lastLoginAt,
		timezone:        timezone,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}
}

type Organization struct {
	id           OrganizationID
	name         string
	slug         string
	description  string
	ownerID      UserID
	billingEmail string
	plan         OrganizationPlan
	status       OrganizationStatus
	createdAt    time.Time
	updatedAt    time.Time
}

type OrganizationID struct {
	value string
}

func NewOrganizationID() OrganizationID {
	return OrganizationID{value: uuid.New().String()}
}

func OrganizationIDFromString(s string) (OrganizationID, error) {
	if s == "" {
		return OrganizationID{}, fmt.Errorf("organization ID cannot be empty")
	}
	return OrganizationID{value: s}, nil
}

func (id OrganizationID) String() string {
	return id.value
}

type OrganizationPlan string

const (
	PlanFree       OrganizationPlan = "free"
	PlanPro        OrganizationPlan = "pro"
	PlanEnterprise OrganizationPlan = "enterprise"
)

type OrganizationStatus string

const (
	OrgStatusActive    OrganizationStatus = "active"
	OrgStatusSuspended OrganizationStatus = "suspended"
	OrgStatusDeleted   OrganizationStatus = "deleted"
)

func NewOrganization(name, slug string, ownerID UserID) *Organization {
	now := time.Now()
	return &Organization{
		id:        NewOrganizationID(),
		name:      name,
		slug:      slug,
		ownerID:   ownerID,
		plan:      PlanFree,
		status:    OrgStatusActive,
		createdAt: now,
		updatedAt: now,
	}
}

func (o *Organization) ID() OrganizationID {
	return o.id
}

func (o *Organization) Name() string {
	return o.name
}

func (o *Organization) Slug() string {
	return o.slug
}

func (o *Organization) Description() string {
	return o.description
}

func (o *Organization) OwnerID() UserID {
	return o.ownerID
}

func (o *Organization) BillingEmail() string {
	return o.billingEmail
}

func (o *Organization) Plan() OrganizationPlan {
	return o.plan
}

func (o *Organization) Status() OrganizationStatus {
	return o.status
}

func (o *Organization) CreatedAt() time.Time {
	return o.createdAt
}

func (o *Organization) UpdatedAt() time.Time {
	return o.updatedAt
}

func (o *Organization) UpdateDescription(description string) {
	o.description = description
	o.updatedAt = time.Now()
}

func (o *Organization) SetBillingEmail(email string) {
	o.billingEmail = email
	o.updatedAt = time.Now()
}

func (o *Organization) ChangePlan(plan OrganizationPlan) {
	o.plan = plan
	o.updatedAt = time.Now()
}

func (o *Organization) ChangeStatus(status OrganizationStatus) {
	o.status = status
	o.updatedAt = time.Now()
}

func ReconstructOrganization(
	id OrganizationID,
	name, slug, description string,
	ownerID UserID,
	billingEmail string,
	plan OrganizationPlan,
	status OrganizationStatus,
	createdAt, updatedAt time.Time,
) *Organization {
	return &Organization{
		id:           id,
		name:         name,
		slug:         slug,
		description:  description,
		ownerID:      ownerID,
		billingEmail: billingEmail,
		plan:         plan,
		status:       status,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}
