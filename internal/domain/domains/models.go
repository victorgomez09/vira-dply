package domains

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Domain struct {
	id            DomainID
	name          DomainName
	projectID     uuid.UUID
	serviceID     *uuid.UUID
	certificateID *CertificateID
	status        DomainStatus
	verified      bool
	redirectTo    string
	createdAt     time.Time
	updatedAt     time.Time
}

type DomainID struct {
	value string
}

func NewDomainID() DomainID {
	return DomainID{value: uuid.New().String()}
}

func DomainIDFromString(s string) (DomainID, error) {
	if s == "" {
		return DomainID{}, fmt.Errorf("domain ID cannot be empty")
	}
	return DomainID{value: s}, nil
}

func (id DomainID) String() string {
	return id.value
}

type DomainName struct {
	value string
}

var domainNameRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

func NewDomainName(name string) (DomainName, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return DomainName{}, fmt.Errorf("domain name cannot be empty")
	}
	if len(name) > 253 {
		return DomainName{}, fmt.Errorf("domain name cannot exceed 253 characters")
	}
	if !domainNameRegex.MatchString(name) {
		return DomainName{}, fmt.Errorf("invalid domain name format")
	}
	return DomainName{value: name}, nil
}

func (n DomainName) String() string {
	return n.value
}

func (n DomainName) IsSubdomain() bool {
	return strings.Count(n.value, ".") > 1
}

func (n DomainName) RootDomain() string {
	parts := strings.Split(n.value, ".")
	if len(parts) < 2 {
		return n.value
	}
	return strings.Join(parts[len(parts)-2:], ".")
}

type DomainStatus string

const (
	DomainStatusPending   DomainStatus = "pending"
	DomainStatusActive    DomainStatus = "active"
	DomainStatusVerifying DomainStatus = "verifying"
	DomainStatusError     DomainStatus = "error"
	DomainStatusSuspended DomainStatus = "suspended"
)

type Certificate struct {
	id        CertificateID
	domainID  DomainID
	issuer    CertificateIssuer
	status    CertificateStatus
	expiresAt time.Time
	autoRenew bool
	createdAt time.Time
	updatedAt time.Time
}

type CertificateID struct {
	value string
}

func NewCertificateID() CertificateID {
	return CertificateID{value: uuid.New().String()}
}

func (id CertificateID) String() string {
	return id.value
}

type CertificateIssuer string

const (
	CertificateIssuerLetsEncrypt CertificateIssuer = "letsencrypt"
	CertificateIssuerSelfSigned  CertificateIssuer = "selfsigned"
	CertificateIssuerCustom      CertificateIssuer = "custom"
)

type CertificateStatus string

const (
	CertificateStatusPending  CertificateStatus = "pending"
	CertificateStatusIssued   CertificateStatus = "issued"
	CertificateStatusRenewing CertificateStatus = "renewing"
	CertificateStatusExpired  CertificateStatus = "expired"
	CertificateStatusRevoked  CertificateStatus = "revoked"
	CertificateStatusError    CertificateStatus = "error"
)

type DNSRecord struct {
	id        DNSRecordID
	domainID  DomainID
	name      string
	type_     DNSRecordType
	value     string
	ttl       int
	priority  *int
	createdAt time.Time
	updatedAt time.Time
}

type DNSRecordID struct {
	value string
}

func NewDNSRecordID() DNSRecordID {
	return DNSRecordID{value: uuid.New().String()}
}

func (id DNSRecordID) String() string {
	return id.value
}

type DNSRecordType string

const (
	DNSRecordTypeA     DNSRecordType = "A"
	DNSRecordTypeAAAA  DNSRecordType = "AAAA"
	DNSRecordTypeCNAME DNSRecordType = "CNAME"
	DNSRecordTypeMX    DNSRecordType = "MX"
	DNSRecordTypeTXT   DNSRecordType = "TXT"
	DNSRecordTypeNS    DNSRecordType = "NS"
	DNSRecordTypeSRV   DNSRecordType = "SRV"
)

func NewDomain(
	name DomainName,
	projectID uuid.UUID,
) *Domain {
	now := time.Now()
	return &Domain{
		id:        NewDomainID(),
		name:      name,
		projectID: projectID,
		status:    DomainStatusPending,
		verified:  false,
		createdAt: now,
		updatedAt: now,
	}
}

func (d *Domain) ID() DomainID {
	return d.id
}

func (d *Domain) Name() DomainName {
	return d.name
}

func (d *Domain) ProjectID() uuid.UUID {
	return d.projectID
}

func (d *Domain) ServiceID() *uuid.UUID {
	return d.serviceID
}

func (d *Domain) CertificateID() *CertificateID {
	return d.certificateID
}

func (d *Domain) Status() DomainStatus {
	return d.status
}

func (d *Domain) Verified() bool {
	return d.verified
}

func (d *Domain) RedirectTo() string {
	return d.redirectTo
}

func (d *Domain) CreatedAt() time.Time {
	return d.createdAt
}

func (d *Domain) UpdatedAt() time.Time {
	return d.updatedAt
}

func (d *Domain) AttachToService(serviceID uuid.UUID) {
	d.serviceID = &serviceID
	d.updatedAt = time.Now()
}

func (d *Domain) DetachFromService() {
	d.serviceID = nil
	d.updatedAt = time.Now()
}

func (d *Domain) SetRedirect(redirectTo string) error {
	if redirectTo != "" {
		if _, err := url.Parse(redirectTo); err != nil {
			return fmt.Errorf("invalid redirect URL: %w", err)
		}
	}
	d.redirectTo = redirectTo
	d.updatedAt = time.Now()
	return nil
}

func (d *Domain) ChangeStatus(status DomainStatus) {
	d.status = status
	d.updatedAt = time.Now()
}

func (d *Domain) MarkVerified() {
	d.verified = true
	d.status = DomainStatusActive
	d.updatedAt = time.Now()
}

func (d *Domain) AttachCertificate(certificateID CertificateID) {
	d.certificateID = &certificateID
	d.updatedAt = time.Now()
}

func (d *Domain) DetachCertificate() {
	d.certificateID = nil
	d.updatedAt = time.Now()
}

func NewCertificate(
	domainID DomainID,
	issuer CertificateIssuer,
	expiresAt time.Time,
) *Certificate {
	now := time.Now()
	return &Certificate{
		id:        NewCertificateID(),
		domainID:  domainID,
		issuer:    issuer,
		status:    CertificateStatusPending,
		expiresAt: expiresAt,
		autoRenew: issuer == CertificateIssuerLetsEncrypt,
		createdAt: now,
		updatedAt: now,
	}
}

func (c *Certificate) ID() CertificateID {
	return c.id
}

func (c *Certificate) DomainID() DomainID {
	return c.domainID
}

func (c *Certificate) Issuer() CertificateIssuer {
	return c.issuer
}

func (c *Certificate) Status() CertificateStatus {
	return c.status
}

func (c *Certificate) ExpiresAt() time.Time {
	return c.expiresAt
}

func (c *Certificate) AutoRenew() bool {
	return c.autoRenew
}

func (c *Certificate) CreatedAt() time.Time {
	return c.createdAt
}

func (c *Certificate) UpdatedAt() time.Time {
	return c.updatedAt
}

func (c *Certificate) ChangeStatus(status CertificateStatus) {
	c.status = status
	c.updatedAt = time.Now()
}

func (c *Certificate) SetAutoRenew(autoRenew bool) {
	c.autoRenew = autoRenew
	c.updatedAt = time.Now()
}

func (c *Certificate) IsExpired() bool {
	return time.Now().After(c.expiresAt)
}

func (c *Certificate) ExpiresWithinDays(days int) bool {
	return time.Now().Add(time.Duration(days) * 24 * time.Hour).After(c.expiresAt)
}

func NewDNSRecord(
	domainID DomainID,
	name string,
	recordType DNSRecordType,
	value string,
	ttl int,
	priority *int,
) (*DNSRecord, error) {
	if name == "" {
		return nil, fmt.Errorf("DNS record name cannot be empty")
	}
	if value == "" {
		return nil, fmt.Errorf("DNS record value cannot be empty")
	}
	if ttl <= 0 {
		ttl = 3600 // Default TTL
	}
	if recordType == DNSRecordTypeMX && priority == nil {
		return nil, fmt.Errorf("MX record requires priority")
	}

	now := time.Now()
	return &DNSRecord{
		id:        NewDNSRecordID(),
		domainID:  domainID,
		name:      name,
		type_:     recordType,
		value:     value,
		ttl:       ttl,
		priority:  priority,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func (r *DNSRecord) ID() DNSRecordID {
	return r.id
}

func (r *DNSRecord) DomainID() DomainID {
	return r.domainID
}

func (r *DNSRecord) Name() string {
	return r.name
}

func (r *DNSRecord) Type() DNSRecordType {
	return r.type_
}

func (r *DNSRecord) Value() string {
	return r.value
}

func (r *DNSRecord) TTL() int {
	return r.ttl
}

func (r *DNSRecord) Priority() *int {
	return r.priority
}

func (r *DNSRecord) CreatedAt() time.Time {
	return r.createdAt
}

func (r *DNSRecord) UpdatedAt() time.Time {
	return r.updatedAt
}

func (r *DNSRecord) UpdateValue(value string) {
	r.value = value
	r.updatedAt = time.Now()
}

func (r *DNSRecord) UpdateTTL(ttl int) {
	if ttl > 0 {
		r.ttl = ttl
		r.updatedAt = time.Now()
	}
}

func ReconstructDomain(
	id DomainID,
	name DomainName,
	projectID uuid.UUID,
	serviceID *uuid.UUID,
	certificateID *CertificateID,
	status DomainStatus,
	verified bool,
	redirectTo string,
	createdAt, updatedAt time.Time,
) *Domain {
	return &Domain{
		id:            id,
		name:          name,
		projectID:     projectID,
		serviceID:     serviceID,
		certificateID: certificateID,
		status:        status,
		verified:      verified,
		redirectTo:    redirectTo,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

func ReconstructCertificate(
	id CertificateID,
	domainID DomainID,
	issuer CertificateIssuer,
	status CertificateStatus,
	expiresAt time.Time,
	autoRenew bool,
	createdAt, updatedAt time.Time,
) *Certificate {
	return &Certificate{
		id:        id,
		domainID:  domainID,
		issuer:    issuer,
		status:    status,
		expiresAt: expiresAt,
		autoRenew: autoRenew,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func ReconstructDNSRecord(
	id DNSRecordID,
	domainID DomainID,
	name string,
	recordType DNSRecordType,
	value string,
	ttl int,
	priority *int,
	createdAt, updatedAt time.Time,
) *DNSRecord {
	return &DNSRecord{
		id:        id,
		domainID:  domainID,
		name:      name,
		type_:     recordType,
		value:     value,
		ttl:       ttl,
		priority:  priority,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}
