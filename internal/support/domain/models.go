package domain

import (
	"time"
)

// TicketPriority defines the urgency level of a support ticket.
type TicketPriority string

const (
	PriorityCritical TicketPriority = "critical"
	PriorityHigh     TicketPriority = "high"
	PriorityNormal   TicketPriority = "normal"
	PriorityLow      TicketPriority = "low"
)

// TicketStatus defines the current state of a support ticket.
type TicketStatus string

const (
	StatusOpen            TicketStatus = "open"
	StatusInProgress      TicketStatus = "in_progress"
	StatusPendingCustomer TicketStatus = "pending_customer"
	StatusResolved        TicketStatus = "resolved"
	StatusClosed          TicketStatus = "closed"
)

// ContractStatus defines the state of a support contract.
type ContractStatus string

const (
	ContractStatusActive    ContractStatus = "active"
	ContractStatusSuspended ContractStatus = "suspended"
	ContractStatusCanceled  ContractStatus = "canceled"
	ContractStatusExpired   ContractStatus = "expired"
)

// TicketCategory categorizes support tickets.
type TicketCategory string

const (
	CategoryBilling     TicketCategory = "billing"
	CategoryTechnical   TicketCategory = "technical"
	CategoryIntegration TicketCategory = "integration"
	CategorySecurity    TicketCategory = "security"
	CategoryGeneral     TicketCategory = "general"
)

// SupportTier represents a support package offering.
type SupportTier struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	DisplayName  string    `json:"display_name"`
	Description  string    `json:"description"`
	PriceMonthly int64     `json:"price_monthly"` // in cents
	PriceYearly  int64     `json:"price_yearly"`  // in cents
	Currency     string    `json:"currency"`
	Features     []string  `json:"features"`
	IsActive     bool      `json:"is_active"`
	SortOrder    int       `json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SLADefinition defines SLA commitments for a tier and priority combination.
type SLADefinition struct {
	ID                      string         `json:"id"`
	TierID                  string         `json:"tier_id"`
	Priority                TicketPriority `json:"priority"`
	FirstResponseMinutes    int            `json:"first_response_minutes"`
	ResolutionTargetMinutes *int           `json:"resolution_target_minutes,omitempty"`
	UptimePercentage        *float64       `json:"uptime_percentage,omitempty"`
	CreatedAt               time.Time      `json:"created_at"`
}

// SupportContract represents an organization's active support agreement.
type SupportContract struct {
	ID           string         `json:"id"`
	OrgID        string         `json:"org_id"`
	TierID       string         `json:"tier_id"`
	Tier         *SupportTier   `json:"tier,omitempty"`
	Status       ContractStatus `json:"status"`
	BillingCycle string         `json:"billing_cycle"` // "monthly" or "yearly"
	StartDate    time.Time      `json:"start_date"`
	EndDate      *time.Time     `json:"end_date,omitempty"`
	AutoRenew    bool           `json:"auto_renew"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// SupportTicket represents a customer support request.
type SupportTicket struct {
	ID                 string         `json:"id"`
	TicketNumber       string         `json:"ticket_number"`
	OrgID              string         `json:"org_id"`
	ContractID         *string        `json:"contract_id,omitempty"`
	RequesterEmail     string         `json:"requester_email"`
	RequesterName      string         `json:"requester_name,omitempty"`
	Subject            string         `json:"subject"`
	Description        string         `json:"description"`
	Priority           TicketPriority `json:"priority"`
	Status             TicketStatus   `json:"status"`
	Category           TicketCategory `json:"category,omitempty"`
	AssignedTo         *string        `json:"assigned_to,omitempty"`
	SLAFirstResponseAt *time.Time     `json:"sla_first_response_at,omitempty"`
	SLAResolutionAt    *time.Time     `json:"sla_resolution_at,omitempty"`
	FirstRespondedAt   *time.Time     `json:"first_responded_at,omitempty"`
	ResolvedAt         *time.Time     `json:"resolved_at,omitempty"`
	SLABreached        bool           `json:"sla_breached"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

// TicketComment represents a comment on a support ticket.
type TicketComment struct {
	ID          string    `json:"id"`
	TicketID    string    `json:"ticket_id"`
	AuthorEmail string    `json:"author_email"`
	AuthorName  string    `json:"author_name,omitempty"`
	IsInternal  bool      `json:"is_internal"` // internal notes vs customer-visible
	IsStaff     bool      `json:"is_staff"`
	Content     string    `json:"content"`
	Attachments []string  `json:"attachments,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Escalation represents a ticket escalation event.
type Escalation struct {
	ID          string    `json:"id"`
	TicketID    string    `json:"ticket_id"`
	EscalatedBy string    `json:"escalated_by"`
	EscalatedTo string    `json:"escalated_to"`
	Reason      string    `json:"reason"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateTicketInput represents input for creating a new ticket.
type CreateTicketInput struct {
	OrgID          string         `json:"org_id"`
	RequesterEmail string         `json:"requester_email"`
	RequesterName  string         `json:"requester_name,omitempty"`
	Subject        string         `json:"subject"`
	Description    string         `json:"description"`
	Priority       TicketPriority `json:"priority"`
	Category       TicketCategory `json:"category,omitempty"`
}

// AddCommentInput represents input for adding a comment to a ticket.
type AddCommentInput struct {
	TicketID    string   `json:"ticket_id"`
	AuthorEmail string   `json:"author_email"`
	AuthorName  string   `json:"author_name,omitempty"`
	IsInternal  bool     `json:"is_internal"`
	IsStaff     bool     `json:"is_staff"`
	Content     string   `json:"content"`
	Attachments []string `json:"attachments,omitempty"`
}

// TicketRepository defines persistence operations for support tickets.
type TicketRepository interface {
	CreateTicket(t *SupportTicket) error
	GetTicketByID(id string) (*SupportTicket, error)
	GetTicketByNumber(number string) (*SupportTicket, error)
	ListTicketsByOrg(orgID string, status *TicketStatus, limit, offset int) ([]*SupportTicket, int, error)
	UpdateTicket(t *SupportTicket) error

	AddComment(c *TicketComment) error
	ListCommentsByTicket(ticketID string) ([]*TicketComment, error)

	AddEscalation(e *Escalation) error
}

// TierRepository defines persistence operations for support tiers.
type TierRepository interface {
	ListTiers(activeOnly bool) ([]*SupportTier, error)
	GetTierByID(id string) (*SupportTier, error)
	GetTierByName(name string) (*SupportTier, error)
	GetSLADefinitions(tierID string) ([]*SLADefinition, error)
}

// ContractRepository defines persistence operations for support contracts.
type ContractRepository interface {
	CreateContract(c *SupportContract) error
	GetContractByID(id string) (*SupportContract, error)
	GetActiveContractByOrg(orgID string) (*SupportContract, error)
	UpdateContract(c *SupportContract) error
	ListContractsByOrg(orgID string) ([]*SupportContract, error)
}
