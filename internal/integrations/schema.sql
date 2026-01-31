-- Integrations Service Schema
-- Manages custom integration projects and migration services

CREATE TABLE IF NOT EXISTS integration_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    source_platform VARCHAR(100) NOT NULL, -- 'stripe', 'paypal', 'square', 'adyen', 'custom'
    category VARCHAR(100) NOT NULL, -- 'payment_migration', 'marketplace_setup', 'webhook_integration', 'custom'
    estimated_days INTEGER NOT NULL,
    base_price BIGINT NOT NULL, -- in cents
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    requirements JSONB DEFAULT '[]',
    deliverables JSONB DEFAULT '[]',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS integration_projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_number VARCHAR(20) NOT NULL UNIQUE,
    org_id UUID NOT NULL,
    template_id UUID REFERENCES integration_templates(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    source_platform VARCHAR(100), -- if migration
    target_config JSONB DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'draft', -- 'draft', 'scoping', 'approved', 'in_progress', 'testing', 'completed', 'on_hold', 'canceled'
    priority VARCHAR(50) NOT NULL DEFAULT 'normal', -- 'low', 'normal', 'high', 'urgent'
    estimated_price BIGINT,
    final_price BIGINT,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    assigned_engineer VARCHAR(255),
    start_date TIMESTAMP WITH TIME ZONE,
    target_completion_date TIMESTAMP WITH TIME ZONE,
    actual_completion_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS project_milestones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES integration_projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    sequence INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'in_progress', 'completed', 'blocked'
    deliverables JSONB DEFAULT '[]',
    due_date TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS integration_consultations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    contact_name VARCHAR(255) NOT NULL,
    contact_email VARCHAR(255) NOT NULL,
    contact_phone VARCHAR(50),
    company_name VARCHAR(255),
    project_type VARCHAR(100) NOT NULL, -- 'migration', 'marketplace', 'custom'
    current_platform VARCHAR(100),
    monthly_volume VARCHAR(100), -- estimated transaction volume
    requirements TEXT,
    preferred_timeline VARCHAR(100),
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'scheduled', 'completed', 'canceled'
    scheduled_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    assigned_to VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS project_updates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES integration_projects(id) ON DELETE CASCADE,
    author VARCHAR(255) NOT NULL,
    update_type VARCHAR(50) NOT NULL, -- 'general', 'milestone', 'blocker', 'completion'
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    is_customer_visible BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_integration_projects_org_id ON integration_projects(org_id);
CREATE INDEX idx_integration_projects_status ON integration_projects(status);
CREATE INDEX idx_integration_projects_assigned ON integration_projects(assigned_engineer);
CREATE INDEX idx_project_milestones_project_id ON project_milestones(project_id);
CREATE INDEX idx_integration_consultations_status ON integration_consultations(status);
CREATE INDEX idx_integration_consultations_org_id ON integration_consultations(org_id);
CREATE INDEX idx_project_updates_project_id ON project_updates(project_id);

-- Insert default integration templates
INSERT INTO integration_templates (name, display_name, description, source_platform, category, estimated_days, base_price, currency, requirements, deliverables) VALUES
('stripe_full_migration', 'Stripe Full Migration', 'Complete migration from Stripe including customers, subscriptions, and historical data', 'stripe', 'payment_migration', 30, 2500000, 'USD',
 '["Stripe API keys (read-only)", "Customer consent documentation", "Historical data requirements"]',
 '["Customer data migration", "Subscription migration", "Payment method migration", "Historical transaction import", "Webhook reconfiguration", "Testing environment", "Go-live support"]'),
('stripe_basic_migration', 'Stripe Basic Migration', 'Essential migration from Stripe for customers and subscriptions only', 'stripe', 'payment_migration', 14, 1000000, 'USD',
 '["Stripe API keys (read-only)", "Customer list export"]',
 '["Customer data migration", "Active subscription migration", "Basic webhook setup", "Go-live support"]'),
('paypal_migration', 'PayPal Migration', 'Migration from PayPal including buyer and seller data', 'paypal', 'payment_migration', 21, 1500000, 'USD',
 '["PayPal API credentials", "Transaction history export", "Account documentation"]',
 '["Customer data migration", "Transaction history import", "Webhook integration", "Testing and validation", "Go-live support"]'),
('square_migration', 'Square Migration', 'Migration from Square for retail and e-commerce merchants', 'square', 'payment_migration', 21, 1500000, 'USD',
 '["Square API credentials", "Product catalog export", "Customer list export"]',
 '["Customer data migration", "Product catalog migration", "Inventory sync setup", "POS integration guidance", "Go-live support"]'),
('marketplace_standard', 'Standard Marketplace Setup', 'Complete marketplace setup with Connect-style payouts', 'custom', 'marketplace_setup', 45, 5000000, 'USD',
 '["Business requirements document", "Seller onboarding flow design", "Payout schedule requirements"]',
 '["Multi-seller architecture", "Seller onboarding flow", "Commission/fee structure", "Payout automation", "Seller dashboard", "Dispute handling workflow", "Testing environment", "Documentation"]'),
('marketplace_enterprise', 'Enterprise Marketplace', 'Enterprise-grade marketplace with advanced features and custom compliance', 'custom', 'marketplace_setup', 90, 15000000, 'USD',
 '["Detailed business requirements", "Compliance requirements", "Custom SLA agreement"]',
 '["Everything in Standard", "Custom compliance controls", "Advanced fraud detection", "Multi-currency support", "White-label dashboard", "Dedicated support channel", "Custom integrations", "SLA guarantees"]'),
('webhook_integration', 'Custom Webhook Integration', 'Integration with third-party systems via webhooks', 'custom', 'webhook_integration', 7, 500000, 'USD',
 '["Target system API documentation", "Event requirements", "Authentication details"]',
 '["Webhook endpoint configuration", "Event mapping", "Retry and failure handling", "Monitoring setup", "Documentation"]'),
('custom_integration', 'Custom Integration Project', 'Bespoke integration work tailored to your specific needs', 'custom', 'custom', 30, 0, 'USD',
 '["Detailed requirements document", "Technical specifications"]',
 '["Custom scoping", "Dedicated project manager", "Regular progress updates", "Testing and validation", "Documentation", "Post-launch support"]')
ON CONFLICT (name) DO NOTHING;
