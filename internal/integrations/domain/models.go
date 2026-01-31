package domain

import (
	"time"
)

// ProjectStatus defines the current state of an integration project.
type ProjectStatus string

const (
	ProjectStatusDraft      ProjectStatus = "draft"
	ProjectStatusScoping    ProjectStatus = "scoping"
	ProjectStatusApproved   ProjectStatus = "approved"
	ProjectStatusInProgress ProjectStatus = "in_progress"
	ProjectStatusTesting    ProjectStatus = "testing"
	ProjectStatusCompleted  ProjectStatus = "completed"
	ProjectStatusOnHold     ProjectStatus = "on_hold"
	ProjectStatusCanceled   ProjectStatus = "canceled"
)

// ProjectPriority defines the urgency level of a project.
type ProjectPriority string

const (
	ProjectPriorityLow    ProjectPriority = "low"
	ProjectPriorityNormal ProjectPriority = "normal"
	ProjectPriorityHigh   ProjectPriority = "high"
	ProjectPriorityUrgent ProjectPriority = "urgent"
)

// MilestoneStatus defines the current state of a milestone.
type MilestoneStatus string

const (
	MilestoneStatusPending    MilestoneStatus = "pending"
	MilestoneStatusInProgress MilestoneStatus = "in_progress"
	MilestoneStatusCompleted  MilestoneStatus = "completed"
	MilestoneStatusBlocked    MilestoneStatus = "blocked"
)

// ConsultationStatus defines the state of a consultation request.
type ConsultationStatus string

const (
	ConsultationStatusPending   ConsultationStatus = "pending"
	ConsultationStatusScheduled ConsultationStatus = "scheduled"
	ConsultationStatusCompleted ConsultationStatus = "completed"
	ConsultationStatusCanceled  ConsultationStatus = "canceled"
)

// TemplateCategory defines the type of integration template.
type TemplateCategory string

const (
	CategoryPaymentMigration   TemplateCategory = "payment_migration"
	CategoryMarketplaceSetup   TemplateCategory = "marketplace_setup"
	CategoryWebhookIntegration TemplateCategory = "webhook_integration"
	CategoryCustom             TemplateCategory = "custom"
)

// IntegrationTemplate represents a pre-built integration offering.
type IntegrationTemplate struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	DisplayName    string           `json:"display_name"`
	Description    string           `json:"description,omitempty"`
	SourcePlatform string           `json:"source_platform"`
	Category       TemplateCategory `json:"category"`
	EstimatedDays  int              `json:"estimated_days"`
	BasePrice      int64            `json:"base_price"` // in cents
	Currency       string           `json:"currency"`
	Requirements   []string         `json:"requirements,omitempty"`
	Deliverables   []string         `json:"deliverables,omitempty"`
	IsActive       bool             `json:"is_active"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// IntegrationProject represents a custom integration or migration project.
type IntegrationProject struct {
	ID                   string                 `json:"id"`
	ProjectNumber        string                 `json:"project_number"`
	OrgID                string                 `json:"org_id"`
	TemplateID           *string                `json:"template_id,omitempty"`
	Template             *IntegrationTemplate   `json:"template,omitempty"`
	Name                 string                 `json:"name"`
	Description          string                 `json:"description,omitempty"`
	SourcePlatform       string                 `json:"source_platform,omitempty"`
	TargetConfig         map[string]interface{} `json:"target_config,omitempty"`
	Status               ProjectStatus          `json:"status"`
	Priority             ProjectPriority        `json:"priority"`
	EstimatedPrice       *int64                 `json:"estimated_price,omitempty"`
	FinalPrice           *int64                 `json:"final_price,omitempty"`
	Currency             string                 `json:"currency"`
	AssignedEngineer     *string                `json:"assigned_engineer,omitempty"`
	StartDate            *time.Time             `json:"start_date,omitempty"`
	TargetCompletionDate *time.Time             `json:"target_completion_date,omitempty"`
	ActualCompletionDate *time.Time             `json:"actual_completion_date,omitempty"`
	Milestones           []*ProjectMilestone    `json:"milestones,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

// ProjectMilestone represents a milestone in an integration project.
type ProjectMilestone struct {
	ID           string          `json:"id"`
	ProjectID    string          `json:"project_id"`
	Name         string          `json:"name"`
	Description  string          `json:"description,omitempty"`
	Sequence     int             `json:"sequence"`
	Status       MilestoneStatus `json:"status"`
	Deliverables []string        `json:"deliverables,omitempty"`
	DueDate      *time.Time      `json:"due_date,omitempty"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	Notes        string          `json:"notes,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// Consultation represents a consultation request for custom integrations.
type Consultation struct {
	ID                string             `json:"id"`
	OrgID             string             `json:"org_id"`
	ContactName       string             `json:"contact_name"`
	ContactEmail      string             `json:"contact_email"`
	ContactPhone      string             `json:"contact_phone,omitempty"`
	CompanyName       string             `json:"company_name,omitempty"`
	ProjectType       string             `json:"project_type"` // 'migration', 'marketplace', 'custom'
	CurrentPlatform   string             `json:"current_platform,omitempty"`
	MonthlyVolume     string             `json:"monthly_volume,omitempty"`
	Requirements      string             `json:"requirements,omitempty"`
	PreferredTimeline string             `json:"preferred_timeline,omitempty"`
	Status            ConsultationStatus `json:"status"`
	ScheduledAt       *time.Time         `json:"scheduled_at,omitempty"`
	CompletedAt       *time.Time         `json:"completed_at,omitempty"`
	Notes             string             `json:"notes,omitempty"`
	AssignedTo        *string            `json:"assigned_to,omitempty"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

// ProjectUpdate represents an update or note on a project.
type ProjectUpdate struct {
	ID                string    `json:"id"`
	ProjectID         string    `json:"project_id"`
	Author            string    `json:"author"`
	UpdateType        string    `json:"update_type"` // 'general', 'milestone', 'blocker', 'completion'
	Title             string    `json:"title"`
	Content           string    `json:"content"`
	IsCustomerVisible bool      `json:"is_customer_visible"`
	CreatedAt         time.Time `json:"created_at"`
}

// CreateProjectInput represents input for creating a new integration project.
type CreateProjectInput struct {
	OrgID          string          `json:"org_id"`
	TemplateID     *string         `json:"template_id,omitempty"`
	Name           string          `json:"name"`
	Description    string          `json:"description,omitempty"`
	SourcePlatform string          `json:"source_platform,omitempty"`
	Priority       ProjectPriority `json:"priority,omitempty"`
}

// CreateConsultationInput represents input for requesting a consultation.
type CreateConsultationInput struct {
	OrgID             string `json:"org_id"`
	ContactName       string `json:"contact_name"`
	ContactEmail      string `json:"contact_email"`
	ContactPhone      string `json:"contact_phone,omitempty"`
	CompanyName       string `json:"company_name,omitempty"`
	ProjectType       string `json:"project_type"`
	CurrentPlatform   string `json:"current_platform,omitempty"`
	MonthlyVolume     string `json:"monthly_volume,omitempty"`
	Requirements      string `json:"requirements,omitempty"`
	PreferredTimeline string `json:"preferred_timeline,omitempty"`
}

// TemplateRepository defines persistence operations for integration templates.
type TemplateRepository interface {
	ListTemplates(activeOnly bool, category *TemplateCategory) ([]*IntegrationTemplate, error)
	GetTemplateByID(id string) (*IntegrationTemplate, error)
	GetTemplateByName(name string) (*IntegrationTemplate, error)
}

// ProjectRepository defines persistence operations for integration projects.
type ProjectRepository interface {
	CreateProject(p *IntegrationProject) error
	GetProjectByID(id string) (*IntegrationProject, error)
	GetProjectByNumber(number string) (*IntegrationProject, error)
	ListProjectsByOrg(orgID string, status *ProjectStatus, limit, offset int) ([]*IntegrationProject, int, error)
	UpdateProject(p *IntegrationProject) error

	CreateMilestone(m *ProjectMilestone) error
	GetMilestoneByID(id string) (*ProjectMilestone, error)
	ListMilestonesByProject(projectID string) ([]*ProjectMilestone, error)
	UpdateMilestone(m *ProjectMilestone) error

	CreateUpdate(u *ProjectUpdate) error
	ListUpdatesByProject(projectID string, customerVisibleOnly bool) ([]*ProjectUpdate, error)
}

// ConsultationRepository defines persistence operations for consultations.
type ConsultationRepository interface {
	CreateConsultation(c *Consultation) error
	GetConsultationByID(id string) (*Consultation, error)
	ListConsultationsByOrg(orgID string) ([]*Consultation, error)
	ListPendingConsultations() ([]*Consultation, error)
	UpdateConsultation(c *Consultation) error
}
