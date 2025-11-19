package networking

import (
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
)

type Network struct {
	id        NetworkID
	name      NetworkName
	projectID uuid.UUID
	cidr      string
	gateway   string
	dns       []string
	isolated  bool
	status    NetworkStatus
	createdAt time.Time
	updatedAt time.Time
}

type NetworkID struct {
	value string
}

func NewNetworkID() NetworkID {
	return NetworkID{value: uuid.New().String()}
}

func NetworkIDFromString(s string) (NetworkID, error) {
	if s == "" {
		return NetworkID{}, fmt.Errorf("network ID cannot be empty")
	}
	return NetworkID{value: s}, nil
}

func (id NetworkID) String() string {
	return id.value
}

type NetworkName struct {
	value string
}

func NewNetworkName(name string) (NetworkName, error) {
	if name == "" {
		return NetworkName{}, fmt.Errorf("network name cannot be empty")
	}
	if len(name) > 64 {
		return NetworkName{}, fmt.Errorf("network name cannot exceed 64 characters")
	}
	return NetworkName{value: name}, nil
}

func (n NetworkName) String() string {
	return n.value
}

type NetworkStatus string

const (
	NetworkStatusCreating NetworkStatus = "creating"
	NetworkStatusActive   NetworkStatus = "active"
	NetworkStatusDeleting NetworkStatus = "deleting"
	NetworkStatusError    NetworkStatus = "error"
)

type NetworkRule struct {
	id          NetworkRuleID
	networkID   NetworkID
	protocol    Protocol
	port        *int
	portRange   *PortRange
	source      string
	target      string
	action      RuleAction
	description string
	createdAt   time.Time
}

type NetworkRuleID struct {
	value string
}

func NewNetworkRuleID() NetworkRuleID {
	return NetworkRuleID{value: uuid.New().String()}
}

func (id NetworkRuleID) String() string {
	return id.value
}

type Protocol string

const (
	ProtocolTCP Protocol = "tcp"
	ProtocolUDP Protocol = "udp"
	ProtocolAny Protocol = "any"
)

type PortRange struct {
	Start int
	End   int
}

func (pr PortRange) Valid() error {
	if pr.Start < 1 || pr.Start > 65535 {
		return fmt.Errorf("invalid start port: %d", pr.Start)
	}
	if pr.End < 1 || pr.End > 65535 {
		return fmt.Errorf("invalid end port: %d", pr.End)
	}
	if pr.Start > pr.End {
		return fmt.Errorf("start port cannot be greater than end port")
	}
	return nil
}

type RuleAction string

const (
	RuleActionAllow RuleAction = "allow"
	RuleActionDeny  RuleAction = "deny"
)

func NewNetwork(
	name NetworkName,
	projectID uuid.UUID,
	cidr string,
	isolated bool,
) (*Network, error) {
	if _, _, err := net.ParseCIDR(cidr); err != nil {
		return nil, fmt.Errorf("invalid CIDR: %w", err)
	}

	now := time.Now()
	return &Network{
		id:        NewNetworkID(),
		name:      name,
		projectID: projectID,
		cidr:      cidr,
		isolated:  isolated,
		status:    NetworkStatusCreating,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func (n *Network) ID() NetworkID {
	return n.id
}

func (n *Network) Name() NetworkName {
	return n.name
}

func (n *Network) ProjectID() uuid.UUID {
	return n.projectID
}

func (n *Network) CIDR() string {
	return n.cidr
}

func (n *Network) Gateway() string {
	return n.gateway
}

func (n *Network) DNS() []string {
	return n.dns
}

func (n *Network) Isolated() bool {
	return n.isolated
}

func (n *Network) Status() NetworkStatus {
	return n.status
}

func (n *Network) CreatedAt() time.Time {
	return n.createdAt
}

func (n *Network) UpdatedAt() time.Time {
	return n.updatedAt
}

func (n *Network) SetGateway(gateway string) error {
	if gateway != "" {
		if ip := net.ParseIP(gateway); ip == nil {
			return fmt.Errorf("invalid gateway IP: %s", gateway)
		}
	}
	n.gateway = gateway
	n.updatedAt = time.Now()
	return nil
}

func (n *Network) SetDNS(dns []string) error {
	for _, d := range dns {
		if ip := net.ParseIP(d); ip == nil {
			return fmt.Errorf("invalid DNS IP: %s", d)
		}
	}
	n.dns = dns
	n.updatedAt = time.Now()
	return nil
}

func (n *Network) ChangeStatus(status NetworkStatus) {
	n.status = status
	n.updatedAt = time.Now()
}

func NewNetworkRule(
	networkID NetworkID,
	protocol Protocol,
	port *int,
	portRange *PortRange,
	source, target string,
	action RuleAction,
	description string,
) (*NetworkRule, error) {
	if port != nil && (*port < 1 || *port > 65535) {
		return nil, fmt.Errorf("invalid port: %d", *port)
	}
	if portRange != nil {
		if err := portRange.Valid(); err != nil {
			return nil, err
		}
	}
	if port != nil && portRange != nil {
		return nil, fmt.Errorf("cannot specify both port and port range")
	}

	return &NetworkRule{
		id:          NewNetworkRuleID(),
		networkID:   networkID,
		protocol:    protocol,
		port:        port,
		portRange:   portRange,
		source:      source,
		target:      target,
		action:      action,
		description: description,
		createdAt:   time.Now(),
	}, nil
}

func (nr *NetworkRule) ID() NetworkRuleID {
	return nr.id
}

func (nr *NetworkRule) NetworkID() NetworkID {
	return nr.networkID
}

func (nr *NetworkRule) Protocol() Protocol {
	return nr.protocol
}

func (nr *NetworkRule) Port() *int {
	return nr.port
}

func (nr *NetworkRule) PortRange() *PortRange {
	return nr.portRange
}

func (nr *NetworkRule) Source() string {
	return nr.source
}

func (nr *NetworkRule) Target() string {
	return nr.target
}

func (nr *NetworkRule) Action() RuleAction {
	return nr.action
}

func (nr *NetworkRule) Description() string {
	return nr.description
}

func (nr *NetworkRule) CreatedAt() time.Time {
	return nr.createdAt
}

func ReconstructNetwork(
	id NetworkID,
	name NetworkName,
	projectID uuid.UUID,
	cidr, gateway string,
	dns []string,
	isolated bool,
	status NetworkStatus,
	createdAt, updatedAt time.Time,
) *Network {
	return &Network{
		id:        id,
		name:      name,
		projectID: projectID,
		cidr:      cidr,
		gateway:   gateway,
		dns:       dns,
		isolated:  isolated,
		status:    status,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func ReconstructNetworkRule(
	id NetworkRuleID,
	networkID NetworkID,
	protocol Protocol,
	port *int,
	portRange *PortRange,
	source, target string,
	action RuleAction,
	description string,
	createdAt time.Time,
) *NetworkRule {
	return &NetworkRule{
		id:          id,
		networkID:   networkID,
		protocol:    protocol,
		port:        port,
		portRange:   portRange,
		source:      source,
		target:      target,
		action:      action,
		description: description,
		createdAt:   createdAt,
	}
}
