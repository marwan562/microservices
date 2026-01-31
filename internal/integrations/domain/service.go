package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProjectNotFound      = errors.New("integration project not found")
	ErrTemplateNotFound     = errors.New("integration template not found")
	ErrConsultationNotFound = errors.New("consultation not found")
	ErrInvalidInput         = errors.New("invalid input")
	ErrInvalidTransition    = errors.New("invalid status transition")
)

// Service provides integration project and consultation management operations.
type Service struct {
	templateRepo     TemplateRepository
	projectRepo      ProjectRepository
	consultationRepo ConsultationRepository
}

// NewService creates a new integrations service.
func NewService(templateRepo TemplateRepository, projectRepo ProjectRepository, consultationRepo ConsultationRepository) *Service {
	return &Service{
		templateRepo:     templateRepo,
		projectRepo:      projectRepo,
		consultationRepo: consultationRepo,
	}
}

// ListTemplates returns available integration templates.
func (s *Service) ListTemplates(ctx context.Context, activeOnly bool, category *TemplateCategory) ([]*IntegrationTemplate, error) {
	return s.templateRepo.ListTemplates(activeOnly, category)
}

// GetTemplate retrieves a template by ID.
func (s *Service) GetTemplate(ctx context.Context, templateID string) (*IntegrationTemplate, error) {
	template, err := s.templateRepo.GetTemplateByID(templateID)
	if err != nil {
		return nil, ErrTemplateNotFound
	}
	return template, nil
}

// CreateProject creates a new integration project.
func (s *Service) CreateProject(ctx context.Context, input CreateProjectInput) (*IntegrationProject, error) {
	if input.OrgID == "" || input.Name == "" {
		return nil, ErrInvalidInput
	}

	if input.Priority == "" {
		input.Priority = ProjectPriorityNormal
	}

	project := &IntegrationProject{
		ID:             uuid.New().String(),
		ProjectNumber:  generateProjectNumber(),
		OrgID:          input.OrgID,
		TemplateID:     input.TemplateID,
		Name:           input.Name,
		Description:    input.Description,
		SourcePlatform: input.SourcePlatform,
		Status:         ProjectStatusDraft,
		Priority:       input.Priority,
		Currency:       "USD",
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	// If using a template, populate defaults
	if input.TemplateID != nil {
		template, err := s.templateRepo.GetTemplateByID(*input.TemplateID)
		if err == nil && template != nil {
			project.Template = template
			project.SourcePlatform = template.SourcePlatform
			estimatedPrice := template.BasePrice
			project.EstimatedPrice = &estimatedPrice

			// Calculate target completion date
			targetDate := time.Now().UTC().AddDate(0, 0, template.EstimatedDays)
			project.TargetCompletionDate = &targetDate
		}
	}

	if err := s.projectRepo.CreateProject(project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Create default milestones if using a template
	if project.Template != nil {
		s.createDefaultMilestones(ctx, project)
	}

	return project, nil
}

// GetProject retrieves a project by ID.
func (s *Service) GetProject(ctx context.Context, projectID string) (*IntegrationProject, error) {
	project, err := s.projectRepo.GetProjectByID(projectID)
	if err != nil {
		return nil, ErrProjectNotFound
	}

	// Load milestones
	milestones, _ := s.projectRepo.ListMilestonesByProject(projectID)
	project.Milestones = milestones

	return project, nil
}

// GetProjectByNumber retrieves a project by its human-readable number.
func (s *Service) GetProjectByNumber(ctx context.Context, number string) (*IntegrationProject, error) {
	project, err := s.projectRepo.GetProjectByNumber(number)
	if err != nil {
		return nil, ErrProjectNotFound
	}
	return project, nil
}

// ListProjects returns projects for an organization.
func (s *Service) ListProjects(ctx context.Context, orgID string, status *ProjectStatus, limit, offset int) ([]*IntegrationProject, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.projectRepo.ListProjectsByOrg(orgID, status, limit, offset)
}

// UpdateProjectStatus updates the status of a project.
func (s *Service) UpdateProjectStatus(ctx context.Context, projectID string, status ProjectStatus) (*IntegrationProject, error) {
	project, err := s.projectRepo.GetProjectByID(projectID)
	if err != nil {
		return nil, ErrProjectNotFound
	}

	// Validate status transition
	if !isValidStatusTransition(project.Status, status) {
		return nil, ErrInvalidTransition
	}

	project.Status = status
	project.UpdatedAt = time.Now().UTC()

	if status == ProjectStatusInProgress && project.StartDate == nil {
		now := time.Now().UTC()
		project.StartDate = &now
	}

	if status == ProjectStatusCompleted {
		now := time.Now().UTC()
		project.ActualCompletionDate = &now
	}

	if err := s.projectRepo.UpdateProject(project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return project, nil
}

// AssignEngineer assigns an engineer to a project.
func (s *Service) AssignEngineer(ctx context.Context, projectID, engineer string) (*IntegrationProject, error) {
	project, err := s.projectRepo.GetProjectByID(projectID)
	if err != nil {
		return nil, ErrProjectNotFound
	}

	project.AssignedEngineer = &engineer
	project.UpdatedAt = time.Now().UTC()

	if err := s.projectRepo.UpdateProject(project); err != nil {
		return nil, fmt.Errorf("failed to assign engineer: %w", err)
	}

	return project, nil
}

// UpdateMilestone updates a milestone's status.
func (s *Service) UpdateMilestone(ctx context.Context, milestoneID string, status MilestoneStatus, notes string) (*ProjectMilestone, error) {
	milestone, err := s.projectRepo.GetMilestoneByID(milestoneID)
	if err != nil {
		return nil, errors.New("milestone not found")
	}

	milestone.Status = status
	milestone.Notes = notes
	milestone.UpdatedAt = time.Now().UTC()

	if status == MilestoneStatusCompleted {
		now := time.Now().UTC()
		milestone.CompletedAt = &now
	}

	if err := s.projectRepo.UpdateMilestone(milestone); err != nil {
		return nil, fmt.Errorf("failed to update milestone: %w", err)
	}

	return milestone, nil
}

// AddProjectUpdate adds an update to a project.
func (s *Service) AddProjectUpdate(ctx context.Context, projectID, author, updateType, title, content string, customerVisible bool) (*ProjectUpdate, error) {
	_, err := s.projectRepo.GetProjectByID(projectID)
	if err != nil {
		return nil, ErrProjectNotFound
	}

	update := &ProjectUpdate{
		ID:                uuid.New().String(),
		ProjectID:         projectID,
		Author:            author,
		UpdateType:        updateType,
		Title:             title,
		Content:           content,
		IsCustomerVisible: customerVisible,
		CreatedAt:         time.Now().UTC(),
	}

	if err := s.projectRepo.CreateUpdate(update); err != nil {
		return nil, fmt.Errorf("failed to add update: %w", err)
	}

	return update, nil
}

// ListProjectUpdates returns updates for a project.
func (s *Service) ListProjectUpdates(ctx context.Context, projectID string, customerVisibleOnly bool) ([]*ProjectUpdate, error) {
	return s.projectRepo.ListUpdatesByProject(projectID, customerVisibleOnly)
}

// RequestConsultation creates a new consultation request.
func (s *Service) RequestConsultation(ctx context.Context, input CreateConsultationInput) (*Consultation, error) {
	if input.OrgID == "" || input.ContactName == "" || input.ContactEmail == "" || input.ProjectType == "" {
		return nil, ErrInvalidInput
	}

	consultation := &Consultation{
		ID:                uuid.New().String(),
		OrgID:             input.OrgID,
		ContactName:       input.ContactName,
		ContactEmail:      input.ContactEmail,
		ContactPhone:      input.ContactPhone,
		CompanyName:       input.CompanyName,
		ProjectType:       input.ProjectType,
		CurrentPlatform:   input.CurrentPlatform,
		MonthlyVolume:     input.MonthlyVolume,
		Requirements:      input.Requirements,
		PreferredTimeline: input.PreferredTimeline,
		Status:            ConsultationStatusPending,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	if err := s.consultationRepo.CreateConsultation(consultation); err != nil {
		return nil, fmt.Errorf("failed to create consultation: %w", err)
	}

	return consultation, nil
}

// GetConsultation retrieves a consultation by ID.
func (s *Service) GetConsultation(ctx context.Context, consultationID string) (*Consultation, error) {
	consultation, err := s.consultationRepo.GetConsultationByID(consultationID)
	if err != nil {
		return nil, ErrConsultationNotFound
	}
	return consultation, nil
}

// ListConsultations returns consultations for an organization.
func (s *Service) ListConsultations(ctx context.Context, orgID string) ([]*Consultation, error) {
	return s.consultationRepo.ListConsultationsByOrg(orgID)
}

// ScheduleConsultation schedules a consultation.
func (s *Service) ScheduleConsultation(ctx context.Context, consultationID string, scheduledAt time.Time, assignedTo string) (*Consultation, error) {
	consultation, err := s.consultationRepo.GetConsultationByID(consultationID)
	if err != nil {
		return nil, ErrConsultationNotFound
	}

	consultation.Status = ConsultationStatusScheduled
	consultation.ScheduledAt = &scheduledAt
	consultation.AssignedTo = &assignedTo
	consultation.UpdatedAt = time.Now().UTC()

	if err := s.consultationRepo.UpdateConsultation(consultation); err != nil {
		return nil, fmt.Errorf("failed to schedule consultation: %w", err)
	}

	return consultation, nil
}

// CompleteConsultation marks a consultation as completed.
func (s *Service) CompleteConsultation(ctx context.Context, consultationID string, notes string) (*Consultation, error) {
	consultation, err := s.consultationRepo.GetConsultationByID(consultationID)
	if err != nil {
		return nil, ErrConsultationNotFound
	}

	now := time.Now().UTC()
	consultation.Status = ConsultationStatusCompleted
	consultation.CompletedAt = &now
	consultation.Notes = notes
	consultation.UpdatedAt = now

	if err := s.consultationRepo.UpdateConsultation(consultation); err != nil {
		return nil, fmt.Errorf("failed to complete consultation: %w", err)
	}

	return consultation, nil
}

// Helper function to generate project number
func generateProjectNumber() string {
	return fmt.Sprintf("INT-%d", time.Now().UnixNano()%1000000000)
}

// Helper function to create default milestones for templated projects
func (s *Service) createDefaultMilestones(ctx context.Context, project *IntegrationProject) {
	defaultMilestones := []struct {
		Name        string
		Description string
	}{
		{"Discovery & Planning", "Initial assessment, requirements gathering, and project planning"},
		{"Environment Setup", "Configure development and staging environments"},
		{"Data Migration", "Migrate customer data, transactions, and configurations"},
		{"Integration Development", "Implement API integrations and webhooks"},
		{"Testing & Validation", "Comprehensive testing and stakeholder validation"},
		{"Go-Live & Handoff", "Production deployment and knowledge transfer"},
	}

	for i, m := range defaultMilestones {
		milestone := &ProjectMilestone{
			ID:          uuid.New().String(),
			ProjectID:   project.ID,
			Name:        m.Name,
			Description: m.Description,
			Sequence:    i + 1,
			Status:      MilestoneStatusPending,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}
		_ = s.projectRepo.CreateMilestone(milestone)
	}
}

// Helper function to validate status transitions
func isValidStatusTransition(from, to ProjectStatus) bool {
	validTransitions := map[ProjectStatus][]ProjectStatus{
		ProjectStatusDraft:      {ProjectStatusScoping, ProjectStatusCanceled},
		ProjectStatusScoping:    {ProjectStatusApproved, ProjectStatusDraft, ProjectStatusCanceled},
		ProjectStatusApproved:   {ProjectStatusInProgress, ProjectStatusOnHold, ProjectStatusCanceled},
		ProjectStatusInProgress: {ProjectStatusTesting, ProjectStatusOnHold, ProjectStatusCanceled},
		ProjectStatusTesting:    {ProjectStatusCompleted, ProjectStatusInProgress, ProjectStatusOnHold},
		ProjectStatusOnHold:     {ProjectStatusInProgress, ProjectStatusCanceled},
		ProjectStatusCompleted:  {},
		ProjectStatusCanceled:   {},
	}

	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}

	for _, status := range allowed {
		if status == to {
			return true
		}
	}
	return false
}
