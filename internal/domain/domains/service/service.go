package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/domains"
	"github.com/mikrocloud/mikrocloud/internal/domain/domains/repository"
)

type DomainService struct {
	domainRepo      repository.DomainRepository
	certificateRepo repository.CertificateRepository
	dnsRecordRepo   repository.DNSRecordRepository
}

func NewDomainService(
	domainRepo repository.DomainRepository,
	certificateRepo repository.CertificateRepository,
	dnsRecordRepo repository.DNSRecordRepository,
) *DomainService {
	return &DomainService{
		domainRepo:      domainRepo,
		certificateRepo: certificateRepo,
		dnsRecordRepo:   dnsRecordRepo,
	}
}

func (s *DomainService) CreateDomain(
	ctx context.Context,
	name domains.DomainName,
	projectID uuid.UUID,
) (*domains.Domain, error) {
	existing, err := s.domainRepo.GetByName(ctx, name)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("domain %s already exists", name)
	}

	domain := domains.NewDomain(name, projectID)
	if err := s.domainRepo.Create(ctx, domain); err != nil {
		return nil, fmt.Errorf("failed to create domain: %w", err)
	}

	return domain, nil
}

func (s *DomainService) GetDomain(ctx context.Context, id domains.DomainID) (*domains.Domain, error) {
	return s.domainRepo.GetByID(ctx, id)
}

func (s *DomainService) GetDomainByName(ctx context.Context, name domains.DomainName) (*domains.Domain, error) {
	return s.domainRepo.GetByName(ctx, name)
}

func (s *DomainService) GetDomainsByProject(ctx context.Context, projectID uuid.UUID) ([]*domains.Domain, error) {
	return s.domainRepo.GetByProjectID(ctx, projectID)
}

func (s *DomainService) AttachDomainToService(ctx context.Context, domainID domains.DomainID, serviceID uuid.UUID) error {
	domain, err := s.domainRepo.GetByID(ctx, domainID)
	if err != nil {
		return err
	}

	domain.AttachToService(serviceID)
	return s.domainRepo.Update(ctx, domain)
}

func (s *DomainService) DetachDomainFromService(ctx context.Context, domainID domains.DomainID) error {
	domain, err := s.domainRepo.GetByID(ctx, domainID)
	if err != nil {
		return err
	}

	domain.DetachFromService()
	return s.domainRepo.Update(ctx, domain)
}

func (s *DomainService) SetDomainRedirect(ctx context.Context, domainID domains.DomainID, redirectTo string) error {
	domain, err := s.domainRepo.GetByID(ctx, domainID)
	if err != nil {
		return err
	}

	if err := domain.SetRedirect(redirectTo); err != nil {
		return err
	}

	return s.domainRepo.Update(ctx, domain)
}

func (s *DomainService) VerifyDomain(ctx context.Context, domainID domains.DomainID) error {
	domain, err := s.domainRepo.GetByID(ctx, domainID)
	if err != nil {
		return err
	}

	domain.MarkVerified()
	return s.domainRepo.Update(ctx, domain)
}

func (s *DomainService) DeleteDomain(ctx context.Context, domainID domains.DomainID) error {
	certificate, err := s.certificateRepo.GetByDomainID(ctx, domainID)
	if err == nil && certificate != nil {
		if err := s.certificateRepo.Delete(ctx, certificate.ID()); err != nil {
			return fmt.Errorf("failed to delete certificate: %w", err)
		}
	}

	dnsRecords, err := s.dnsRecordRepo.GetByDomainID(ctx, domainID)
	if err != nil {
		return fmt.Errorf("failed to get DNS records: %w", err)
	}

	for _, record := range dnsRecords {
		if err := s.dnsRecordRepo.Delete(ctx, record.ID()); err != nil {
			return fmt.Errorf("failed to delete DNS record: %w", err)
		}
	}

	return s.domainRepo.Delete(ctx, domainID)
}

func (s *DomainService) CreateCertificate(
	ctx context.Context,
	domainID domains.DomainID,
	issuer domains.CertificateIssuer,
	expiresAt time.Time,
) (*domains.Certificate, error) {
	certificate := domains.NewCertificate(domainID, issuer, expiresAt)
	if err := s.certificateRepo.Create(ctx, certificate); err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	domain, err := s.domainRepo.GetByID(ctx, domainID)
	if err != nil {
		return nil, err
	}

	domain.AttachCertificate(certificate.ID())
	if err := s.domainRepo.Update(ctx, domain); err != nil {
		return nil, fmt.Errorf("failed to attach certificate to domain: %w", err)
	}

	return certificate, nil
}

func (s *DomainService) GetCertificate(ctx context.Context, id domains.CertificateID) (*domains.Certificate, error) {
	return s.certificateRepo.GetByID(ctx, id)
}

func (s *DomainService) GetExpiringCertificates(ctx context.Context, days int) ([]*domains.Certificate, error) {
	return s.certificateRepo.GetExpiringBefore(ctx, days)
}

func (s *DomainService) RenewCertificate(ctx context.Context, certificateID domains.CertificateID, newExpiresAt time.Time) error {
	certificate, err := s.certificateRepo.GetByID(ctx, certificateID)
	if err != nil {
		return err
	}

	certificate.ChangeStatus(domains.CertificateStatusIssued)
	return s.certificateRepo.Update(ctx, certificate)
}

func (s *DomainService) CreateDNSRecord(
	ctx context.Context,
	domainID domains.DomainID,
	name string,
	recordType domains.DNSRecordType,
	value string,
	ttl int,
	priority *int,
) (*domains.DNSRecord, error) {
	record, err := domains.NewDNSRecord(domainID, name, recordType, value, ttl, priority)
	if err != nil {
		return nil, fmt.Errorf("failed to create DNS record: %w", err)
	}

	if err := s.dnsRecordRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to save DNS record: %w", err)
	}

	return record, nil
}

func (s *DomainService) GetDNSRecord(ctx context.Context, id domains.DNSRecordID) (*domains.DNSRecord, error) {
	return s.dnsRecordRepo.GetByID(ctx, id)
}

func (s *DomainService) GetDNSRecords(ctx context.Context, domainID domains.DomainID) ([]*domains.DNSRecord, error) {
	return s.dnsRecordRepo.GetByDomainID(ctx, domainID)
}

func (s *DomainService) UpdateDNSRecord(ctx context.Context, recordID domains.DNSRecordID, value string, ttl int) error {
	record, err := s.dnsRecordRepo.GetByID(ctx, recordID)
	if err != nil {
		return err
	}

	record.UpdateValue(value)
	record.UpdateTTL(ttl)
	return s.dnsRecordRepo.Update(ctx, record)
}

func (s *DomainService) DeleteDNSRecord(ctx context.Context, recordID domains.DNSRecordID) error {
	return s.dnsRecordRepo.Delete(ctx, recordID)
}
