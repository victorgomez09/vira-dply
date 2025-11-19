package domains

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"strings"
)

const (
	charset  = "abcdefghijklmnopqrstuvwxyz0123456789"
	idLength = 24
)

type DomainGenerator struct {
	serverIP string
}

func NewDomainGenerator(serverIP string) *DomainGenerator {
	return &DomainGenerator{
		serverIP: serverIP,
	}
}

func (dg *DomainGenerator) GenerateRandomDomain() (string, error) {
	randomID, err := generateRandomID(idLength)
	if err != nil {
		return "", fmt.Errorf("failed to generate random ID: %w", err)
	}

	ip := dg.serverIP
	if ip == "" {
		detectedIP, err := detectPublicIP()
		if err != nil {
			return "", fmt.Errorf("failed to detect public IP: %w", err)
		}
		ip = detectedIP
	}

	domain := fmt.Sprintf("%s.%s.sslip.io", randomID, strings.ReplaceAll(ip, ".", "-"))
	return domain, nil
}

func (dg *DomainGenerator) GenerateDomainWithSubdomain(subdomain string) (string, error) {
	if subdomain == "" {
		return "", fmt.Errorf("subdomain cannot be empty")
	}

	subdomain = sanitizeSubdomain(subdomain)

	ip := dg.serverIP
	if ip == "" {
		detectedIP, err := detectPublicIP()
		if err != nil {
			return "", fmt.Errorf("failed to detect public IP: %w", err)
		}
		ip = detectedIP
	}

	domain := fmt.Sprintf("%s.%s.sslip.io", subdomain, strings.ReplaceAll(ip, ".", "-"))
	return domain, nil
}

func (dg *DomainGenerator) ValidateDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid domain format")
	}

	return nil
}

func generateRandomID(length int) (string, error) {
	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		result[i] = charset[randomIndex.Int64()]
	}

	return string(result), nil
}

func sanitizeSubdomain(subdomain string) string {
	subdomain = strings.ToLower(subdomain)
	subdomain = strings.ReplaceAll(subdomain, " ", "-")
	subdomain = strings.ReplaceAll(subdomain, "_", "-")

	var result strings.Builder
	for _, char := range subdomain {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			result.WriteRune(char)
		}
	}

	sanitized := result.String()
	sanitized = strings.Trim(sanitized, "-")

	if len(sanitized) > 63 {
		sanitized = sanitized[:63]
	}

	return sanitized
}

func detectPublicIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "127.0.0.1", nil
}
