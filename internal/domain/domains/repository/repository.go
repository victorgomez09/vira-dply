package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/domains"
)

type DomainRepository interface {
	Create(ctx context.Context, domain *domains.Domain) error
	GetByID(ctx context.Context, id domains.DomainID) (*domains.Domain, error)
	GetByName(ctx context.Context, name domains.DomainName) (*domains.Domain, error)
	GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*domains.Domain, error)
	GetByServiceID(ctx context.Context, serviceID uuid.UUID) ([]*domains.Domain, error)
	Update(ctx context.Context, domain *domains.Domain) error
	Delete(ctx context.Context, id domains.DomainID) error
}

type CertificateRepository interface {
	Create(ctx context.Context, certificate *domains.Certificate) error
	GetByID(ctx context.Context, id domains.CertificateID) (*domains.Certificate, error)
	GetByDomainID(ctx context.Context, domainID domains.DomainID) (*domains.Certificate, error)
	GetExpiringBefore(ctx context.Context, days int) ([]*domains.Certificate, error)
	Update(ctx context.Context, certificate *domains.Certificate) error
	Delete(ctx context.Context, id domains.CertificateID) error
}

type DNSRecordRepository interface {
	Create(ctx context.Context, record *domains.DNSRecord) error
	GetByID(ctx context.Context, id domains.DNSRecordID) (*domains.DNSRecord, error)
	GetByDomainID(ctx context.Context, domainID domains.DomainID) ([]*domains.DNSRecord, error)
	Update(ctx context.Context, record *domains.DNSRecord) error
	Delete(ctx context.Context, id domains.DNSRecordID) error
}
