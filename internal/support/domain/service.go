package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTicketNotFound   = errors.New("support ticket not found")
	ErrContractNotFound = errors.New("support contract not found")
	ErrTierNotFound     = errors.New("support tier not found")
	ErrInvalidInput     = errors.New("invalid input")
)

// Service provides support ticket and contract management operations.
type Service struct {
	ticketRepo   TicketRepository
	tierRepo     TierRepository
	contractRepo ContractRepository
}

// NewService creates a new support service.
func NewService(ticketRepo TicketRepository, tierRepo TierRepository, contractRepo ContractRepository) *Service {
	return &Service{
		ticketRepo:   ticketRepo,
		tierRepo:     tierRepo,
		contractRepo: contractRepo,
	}
}

// CreateTicket creates a new support ticket with SLA calculations.
func (s *Service) CreateTicket(ctx context.Context, input CreateTicketInput) (*SupportTicket, error) {
	if input.OrgID == "" || input.RequesterEmail == "" || input.Subject == "" || input.Description == "" {
		return nil, ErrInvalidInput
	}

	if input.Priority == "" {
		input.Priority = PriorityNormal
	}

	if input.Category == "" {
		input.Category = CategoryGeneral
	}

	ticket := &SupportTicket{
		ID:             uuid.New().String(),
		TicketNumber:   generateTicketNumber(),
		OrgID:          input.OrgID,
		RequesterEmail: input.RequesterEmail,
		RequesterName:  input.RequesterName,
		Subject:        input.Subject,
		Description:    input.Description,
		Priority:       input.Priority,
		Status:         StatusOpen,
		Category:       input.Category,
		SLABreached:    false,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	// Look up org's support contract and calculate SLA deadlines
	contract, err := s.contractRepo.GetActiveContractByOrg(input.OrgID)
	if err == nil && contract != nil {
		ticket.ContractID = &contract.ID

		slas, err := s.tierRepo.GetSLADefinitions(contract.TierID)
		if err == nil {
			for _, sla := range slas {
				if sla.Priority == input.Priority {
					responseDeadline := ticket.CreatedAt.Add(time.Duration(sla.FirstResponseMinutes) * time.Minute)
					ticket.SLAFirstResponseAt = &responseDeadline

					if sla.ResolutionTargetMinutes != nil && *sla.ResolutionTargetMinutes > 0 {
						resolutionDeadline := ticket.CreatedAt.Add(time.Duration(*sla.ResolutionTargetMinutes) * time.Minute)
						ticket.SLAResolutionAt = &resolutionDeadline
					}
					break
				}
			}
		}
	}

	if err := s.ticketRepo.CreateTicket(ticket); err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	return ticket, nil
}

// GetTicket retrieves a ticket by ID.
func (s *Service) GetTicket(ctx context.Context, ticketID string) (*SupportTicket, error) {
	ticket, err := s.ticketRepo.GetTicketByID(ticketID)
	if err != nil {
		return nil, ErrTicketNotFound
	}
	return ticket, nil
}

// GetTicketByNumber retrieves a ticket by its human-readable number.
func (s *Service) GetTicketByNumber(ctx context.Context, number string) (*SupportTicket, error) {
	ticket, err := s.ticketRepo.GetTicketByNumber(number)
	if err != nil {
		return nil, ErrTicketNotFound
	}
	return ticket, nil
}

// ListTickets returns tickets for an organization.
func (s *Service) ListTickets(ctx context.Context, orgID string, status *TicketStatus, limit, offset int) ([]*SupportTicket, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.ticketRepo.ListTicketsByOrg(orgID, status, limit, offset)
}

// UpdateTicketStatus updates the status of a ticket.
func (s *Service) UpdateTicketStatus(ctx context.Context, ticketID string, status TicketStatus) (*SupportTicket, error) {
	ticket, err := s.ticketRepo.GetTicketByID(ticketID)
	if err != nil {
		return nil, ErrTicketNotFound
	}

	ticket.Status = status
	ticket.UpdatedAt = time.Now().UTC()

	if status == StatusResolved || status == StatusClosed {
		now := time.Now().UTC()
		ticket.ResolvedAt = &now
	}

	if err := s.ticketRepo.UpdateTicket(ticket); err != nil {
		return nil, fmt.Errorf("failed to update ticket: %w", err)
	}

	return ticket, nil
}

// AssignTicket assigns a ticket to a support agent.
func (s *Service) AssignTicket(ctx context.Context, ticketID, assignee string) (*SupportTicket, error) {
	ticket, err := s.ticketRepo.GetTicketByID(ticketID)
	if err != nil {
		return nil, ErrTicketNotFound
	}

	ticket.AssignedTo = &assignee
	ticket.UpdatedAt = time.Now().UTC()

	if ticket.Status == StatusOpen {
		ticket.Status = StatusInProgress
	}

	if err := s.ticketRepo.UpdateTicket(ticket); err != nil {
		return nil, fmt.Errorf("failed to assign ticket: %w", err)
	}

	return ticket, nil
}

// AddComment adds a comment to a ticket.
func (s *Service) AddComment(ctx context.Context, input AddCommentInput) (*TicketComment, error) {
	if input.TicketID == "" || input.AuthorEmail == "" || input.Content == "" {
		return nil, ErrInvalidInput
	}

	ticket, err := s.ticketRepo.GetTicketByID(input.TicketID)
	if err != nil {
		return nil, ErrTicketNotFound
	}

	comment := &TicketComment{
		ID:          uuid.New().String(),
		TicketID:    input.TicketID,
		AuthorEmail: input.AuthorEmail,
		AuthorName:  input.AuthorName,
		IsInternal:  input.IsInternal,
		IsStaff:     input.IsStaff,
		Content:     input.Content,
		Attachments: input.Attachments,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.ticketRepo.AddComment(comment); err != nil {
		return nil, fmt.Errorf("failed to add comment: %w", err)
	}

	// Mark first response time if this is a staff reply and it's the first one
	if input.IsStaff && ticket.FirstRespondedAt == nil && !input.IsInternal {
		now := time.Now().UTC()
		ticket.FirstRespondedAt = &now
		ticket.UpdatedAt = now

		// Check SLA breach
		if ticket.SLAFirstResponseAt != nil && now.After(*ticket.SLAFirstResponseAt) {
			ticket.SLABreached = true
		}

		_ = s.ticketRepo.UpdateTicket(ticket)
	}

	return comment, nil
}

// ListComments returns all comments for a ticket.
func (s *Service) ListComments(ctx context.Context, ticketID string) ([]*TicketComment, error) {
	return s.ticketRepo.ListCommentsByTicket(ticketID)
}

// EscalateTicket escalates a ticket to a higher tier.
func (s *Service) EscalateTicket(ctx context.Context, ticketID, escalatedBy, escalatedTo, reason string) error {
	ticket, err := s.ticketRepo.GetTicketByID(ticketID)
	if err != nil {
		return ErrTicketNotFound
	}

	escalation := &Escalation{
		ID:          uuid.New().String(),
		TicketID:    ticketID,
		EscalatedBy: escalatedBy,
		EscalatedTo: escalatedTo,
		Reason:      reason,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.ticketRepo.AddEscalation(escalation); err != nil {
		return fmt.Errorf("failed to add escalation: %w", err)
	}

	// Upgrade priority if not already critical
	if ticket.Priority != PriorityCritical {
		newPriority := upgradeTicketPriority(ticket.Priority)
		ticket.Priority = newPriority
		ticket.AssignedTo = &escalatedTo
		ticket.UpdatedAt = time.Now().UTC()
		_ = s.ticketRepo.UpdateTicket(ticket)
	}

	return nil
}

// ListTiers returns available support tiers.
func (s *Service) ListTiers(ctx context.Context, activeOnly bool) ([]*SupportTier, error) {
	return s.tierRepo.ListTiers(activeOnly)
}

// GetTier retrieves a tier by ID.
func (s *Service) GetTier(ctx context.Context, tierID string) (*SupportTier, error) {
	tier, err := s.tierRepo.GetTierByID(tierID)
	if err != nil {
		return nil, ErrTierNotFound
	}
	return tier, nil
}

// CreateContract creates a new support contract for an organization.
func (s *Service) CreateContract(ctx context.Context, orgID, tierID, billingCycle string) (*SupportContract, error) {
	tier, err := s.tierRepo.GetTierByID(tierID)
	if err != nil {
		return nil, ErrTierNotFound
	}

	// Check for existing active contract
	existing, _ := s.contractRepo.GetActiveContractByOrg(orgID)
	if existing != nil {
		return nil, errors.New("organization already has an active support contract")
	}

	now := time.Now().UTC()
	var endDate time.Time
	if billingCycle == "yearly" {
		endDate = now.AddDate(1, 0, 0)
	} else {
		billingCycle = "monthly"
		endDate = now.AddDate(0, 1, 0)
	}

	contract := &SupportContract{
		ID:           uuid.New().String(),
		OrgID:        orgID,
		TierID:       tierID,
		Tier:         tier,
		Status:       ContractStatusActive,
		BillingCycle: billingCycle,
		StartDate:    now,
		EndDate:      &endDate,
		AutoRenew:    true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.contractRepo.CreateContract(contract); err != nil {
		return nil, fmt.Errorf("failed to create contract: %w", err)
	}

	return contract, nil
}

// GetContract retrieves a contract by ID.
func (s *Service) GetContract(ctx context.Context, contractID string) (*SupportContract, error) {
	contract, err := s.contractRepo.GetContractByID(contractID)
	if err != nil {
		return nil, ErrContractNotFound
	}
	return contract, nil
}

// GetOrgContract retrieves the active contract for an organization.
func (s *Service) GetOrgContract(ctx context.Context, orgID string) (*SupportContract, error) {
	return s.contractRepo.GetActiveContractByOrg(orgID)
}

// CancelContract cancels an active contract.
func (s *Service) CancelContract(ctx context.Context, contractID string) error {
	contract, err := s.contractRepo.GetContractByID(contractID)
	if err != nil {
		return ErrContractNotFound
	}

	if contract.Status != ContractStatusActive {
		return errors.New("only active contracts can be canceled")
	}

	contract.Status = ContractStatusCanceled
	contract.AutoRenew = false
	contract.UpdatedAt = time.Now().UTC()

	return s.contractRepo.UpdateContract(contract)
}

// CheckSLABreaches checks for SLA breaches across all open tickets.
func (s *Service) CheckSLABreaches(ctx context.Context) ([]*SupportTicket, error) {
	// This would typically be run by a background worker
	// For now, return empty - actual implementation would query tickets
	// where SLA deadlines have passed and SLABreached is false
	return []*SupportTicket{}, nil
}

// Helper function to generate ticket number
func generateTicketNumber() string {
	return fmt.Sprintf("TKT-%d", time.Now().UnixNano()%1000000000)
}

// Helper function to upgrade ticket priority
func upgradeTicketPriority(current TicketPriority) TicketPriority {
	switch current {
	case PriorityLow:
		return PriorityNormal
	case PriorityNormal:
		return PriorityHigh
	case PriorityHigh:
		return PriorityCritical
	default:
		return current
	}
}
