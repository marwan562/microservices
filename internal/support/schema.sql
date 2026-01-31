-- Support Service Schema
-- Manages support tiers, SLA definitions, and support tickets

CREATE TABLE IF NOT EXISTS support_tiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    price_monthly BIGINT NOT NULL, -- in cents
    price_yearly BIGINT NOT NULL,  -- in cents
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    features JSONB NOT NULL DEFAULT '[]',
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sla_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tier_id UUID NOT NULL REFERENCES support_tiers(id) ON DELETE CASCADE,
    priority VARCHAR(50) NOT NULL, -- 'critical', 'high', 'normal', 'low'
    first_response_minutes INTEGER NOT NULL,
    resolution_target_minutes INTEGER,
    uptime_percentage DECIMAL(5,2), -- e.g., 99.99
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tier_id, priority)
);

CREATE TABLE IF NOT EXISTS support_contracts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    tier_id UUID NOT NULL REFERENCES support_tiers(id),
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- 'active', 'suspended', 'canceled', 'expired'
    billing_cycle VARCHAR(20) NOT NULL, -- 'monthly', 'yearly'
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE,
    auto_renew BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS support_tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_number VARCHAR(20) NOT NULL UNIQUE,
    org_id UUID NOT NULL,
    contract_id UUID REFERENCES support_contracts(id),
    requester_email VARCHAR(255) NOT NULL,
    requester_name VARCHAR(255),
    subject VARCHAR(500) NOT NULL,
    description TEXT NOT NULL,
    priority VARCHAR(50) NOT NULL DEFAULT 'normal', -- 'critical', 'high', 'normal', 'low'
    status VARCHAR(50) NOT NULL DEFAULT 'open', -- 'open', 'in_progress', 'pending_customer', 'resolved', 'closed'
    category VARCHAR(100), -- 'billing', 'technical', 'integration', 'security', 'general'
    assigned_to VARCHAR(255),
    sla_first_response_at TIMESTAMP WITH TIME ZONE,
    sla_resolution_at TIMESTAMP WITH TIME ZONE,
    first_responded_at TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    sla_breached BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS support_ticket_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    author_email VARCHAR(255) NOT NULL,
    author_name VARCHAR(255),
    is_internal BOOLEAN NOT NULL DEFAULT false, -- internal notes vs customer-visible
    is_staff BOOLEAN NOT NULL DEFAULT false,
    content TEXT NOT NULL,
    attachments JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS support_escalations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    escalated_by VARCHAR(255) NOT NULL,
    escalated_to VARCHAR(255) NOT NULL,
    reason TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_support_tickets_org_id ON support_tickets(org_id);
CREATE INDEX idx_support_tickets_status ON support_tickets(status);
CREATE INDEX idx_support_tickets_priority ON support_tickets(priority);
CREATE INDEX idx_support_tickets_created_at ON support_tickets(created_at);
CREATE INDEX idx_support_tickets_sla_breached ON support_tickets(sla_breached);
CREATE INDEX idx_support_contracts_org_id ON support_contracts(org_id);
CREATE INDEX idx_support_contracts_status ON support_contracts(status);
CREATE INDEX idx_support_ticket_comments_ticket_id ON support_ticket_comments(ticket_id);

-- Insert default support tiers
INSERT INTO support_tiers (name, display_name, description, price_monthly, price_yearly, currency, features, sort_order) VALUES
('essential', 'Essential', 'Community-level support with basic SLAs', 0, 0, 'USD', 
 '["Community forum access", "Email support (48h response)", "Documentation access", "Standard business hours"]', 1),
('professional', 'Professional', 'Priority support for growing teams', 49900, 499000, 'USD',
 '["Priority email support (24h response)", "Phone support (business hours)", "Dedicated Slack channel", "Quarterly business reviews", "99.9% uptime SLA"]', 2),
('enterprise', 'Enterprise', 'Mission-critical support with guaranteed SLAs', 199900, 1999000, 'USD',
 '["24/7 phone and email support", "1h response for critical issues", "Dedicated support engineer", "Custom SLA agreements", "Executive escalation path", "99.99% uptime SLA", "Priority bugfix queue"]', 3)
ON CONFLICT (name) DO NOTHING;

-- Insert default SLA definitions
INSERT INTO sla_definitions (tier_id, priority, first_response_minutes, resolution_target_minutes, uptime_percentage)
SELECT id, 'critical', 2880, NULL, 99.00 FROM support_tiers WHERE name = 'essential'
UNION ALL
SELECT id, 'high', 2880, NULL, 99.00 FROM support_tiers WHERE name = 'essential'
UNION ALL
SELECT id, 'normal', 2880, NULL, 99.00 FROM support_tiers WHERE name = 'essential'
UNION ALL
SELECT id, 'low', 2880, NULL, 99.00 FROM support_tiers WHERE name = 'essential'
UNION ALL
SELECT id, 'critical', 240, 1440, 99.90 FROM support_tiers WHERE name = 'professional'
UNION ALL
SELECT id, 'high', 480, 2880, 99.90 FROM support_tiers WHERE name = 'professional'
UNION ALL
SELECT id, 'normal', 1440, 4320, 99.90 FROM support_tiers WHERE name = 'professional'
UNION ALL
SELECT id, 'low', 2880, NULL, 99.90 FROM support_tiers WHERE name = 'professional'
UNION ALL
SELECT id, 'critical', 60, 480, 99.99 FROM support_tiers WHERE name = 'enterprise'
UNION ALL
SELECT id, 'high', 120, 1440, 99.99 FROM support_tiers WHERE name = 'enterprise'
UNION ALL
SELECT id, 'normal', 480, 2880, 99.99 FROM support_tiers WHERE name = 'enterprise'
UNION ALL
SELECT id, 'low', 1440, 4320, 99.99 FROM support_tiers WHERE name = 'enterprise'
ON CONFLICT DO NOTHING;
