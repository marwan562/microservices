# Sapliy Business Model & Customer Journey

> Complete overview of Sapliy's hybrid SaaS + self-hosted business model and customer acquisition strategy

---

## Table of Contents

1. [Business Model Overview](#business-model-overview)
2. [Customer Segments](#customer-segments)
3. [SaaS Pricing & Growth Strategy](#saas-pricing--growth-strategy)
4. [Self-Hosted Enterprise Licensing](#self-hosted-enterprise-licensing)
5. [Customer Acquisition Journey](#customer-acquisition-journey)
6. [Revenue Projections](#revenue-projections)
7. [Competitive Positioning](#competitive-positioning)

---

## Business Model Overview

### Hybrid SaaS + Self-Hosted Strategy

Sapliy operates a **dual revenue model** serving two distinct customer segments with the same core technology:

#### 🌐 SaaS Segment (Managed Service)

- **Target**: Startups, SMBs, fast-growing companies
- **Entry point**: Free tier (1 zone, 1K events/month, test mode)
- **Growth path**: Starter → Pro → Enterprise
- **Revenue model**: Usage-based + subscription
- **Customer LTV**: $5K - $50K annually (average)
- **Churn rate**: <5% (industry target)
- **Growth rate**: 15% MoM in year 1

#### 💼 Self-Hosted Segment (Enterprise)

- **Target**: Large enterprises, regulated industries, data-sovereignty requirements
- **Entry point**: $1,999/year startup license
- **Growth path**: Startup → Growth → Enterprise
- **Revenue model**: Annual license + support
- **Customer LTV**: $10K - $500K+ (depending on scale)
- **Churn rate**: <2% (sticky, long contract terms)
- **Growth rate**: 30% YoY (enterprise sales cycles are longer)

### Cross-Selling Opportunities

Customers often start with **SaaS** and graduate to **Self-Hosted**:

1. Dev team uses SaaS (test mode) for rapid prototyping
2. Company gains compliance/scale requirements
3. Migration path: Export flows → Deploy self-hosted → Use same SDKs
4. Potential upgrade from SaaS ($99/mo) to Enterprise ($50K+/year)

---

## Customer Segments

### Segment 1: Early-Stage Startups

**Profile**:

- 5-50 employees
- Series A/B funded
- Moving fast, shipping weekly
- Low compliance requirements
- Budget: $500-$5K/month for tools

**Use Cases**:

- Payment automation & reconciliation
- User onboarding workflows
- Notification & alerting systems
- Ledger & accounting automation

**Acquisition Channel**:

- Product Hunt, Hacker News, Dev.to
- Startup communities (Y Combinator, Techstars)
- Developer marketing (Twitter, LinkedIn)
- Referrals

**Success Metrics**:

- Time to first event: < 15 minutes
- Time to first flow: < 30 minutes
- Free-to-paid conversion: 10-15%
- Upgrade rate (free → starter): 5-10%

**Pricing Sensitivity**: Very high

- Start with free tier
- Upgrade when hitting limits
- Price increase triggers churn (need smooth tier transitions)

### Segment 2: Growth-Stage Companies

**Profile**:

- 50-500 employees
- Series B/C funded
- Established product, scaling operations
- Increasing compliance needs (GDPR, PCI-DSS)
- Budget: $10K-$50K/month for tools

**Use Cases**:

- Complex payment processing & routing
- Multi-channel notifications (email, SMS, WhatsApp)
- Advanced analytics & reporting
- Audit logging & compliance
- Custom policy enforcement

**Acquisition Channel**:

- Sales-assisted (SDR outreach)
- Content marketing (blog, whitepapers)
- Industry conferences
- Customer references

**Success Metrics**:

- Implementation time: 4-6 weeks
- ROI within 90 days
- Adoption across 3+ teams
- Data-driven decision making

**Pricing Sensitivity**: Medium

- Willing to pay for advanced features
- Value time savings & reduced maintenance
- Annual contracts preferred (10-15% discount)

### Segment 3: Enterprise Companies

**Profile**:

- 500+ employees
- Public or large private companies
- Mission-critical systems
- Strict compliance & regulatory requirements (HIPAA, FedRAMP, SOX)
- Budget: $50K-$500K+/year for platform

**Use Cases**:

- Enterprise-grade event automation
- Regulatory compliance (audit trails, data residency)
- High-availability, disaster recovery
- Custom integrations & policies
- White-label deployments

**Acquisition Channel**:

- Enterprise sales (AE + CSM)
- RFP/RFQ process
- Direct outreach to CTOs
- Analyst reports (Gartner, Forrester)

**Success Metrics**:

- Deployment within target SLA
- 99.95% uptime guarantee maintained
- Compliance certifications achieved
- Executive sponsorship & adoption

**Pricing Sensitivity**: Low

- Focused on TCO, not unit cost
- Value reliability, support, compliance
- Multi-year contracts (3-5 years)
- Budget available in capital expenditure

---

## SaaS Pricing & Growth Strategy

### Pricing Tiers

| Tier           | Monthly Price | Annual Price | Events/mo | Zones     | Live Mode | Use Case               |
| -------------- | ------------- | ------------ | --------- | --------- | --------- | ---------------------- |
| **Free**       | $0            | $0           | 1K        | 1         | ❌ Test   | Getting started, hobby |
| **Starter**    | $29           | $290         | 10K       | 3         | ✅ Yes    | Early-stage startups   |
| **Pro**        | $99           | $990         | 100K      | Unlimited | ✅ Yes    | Growth-stage companies |
| **Enterprise** | Custom        | Custom       | Unlimited | Unlimited | ✅ Yes    | Large companies, SaaS  |

### Revenue Drivers (SaaS)

#### 1. Event Overage Charges

```
Pricing:
- Included in tier: 1K-100K events/month
- Overage: $0.10 per 1M events (decreasing scale)

Example:
- Pro tier: 100K events/month included
- Customer uses 250K: 150K overages = $0.015
- Total: $99 + $1.50 = $100.50/month
```

#### 2. Add-On Features (Premium)

```
Pricing (monthly):
- SMS/WhatsApp notifications: $29 (+ per message)
- Custom policies (OPA): $49
- Advanced analytics: $49
- Priority support: $99
- White-label dashboard: $299
```

#### 3. Professional Services (Optional)

```
Hourly Rates:
- Setup & configuration: $200/hour
- Custom integrations: $250/hour
- Compliance consulting: $300/hour
- Architecture review: $350/hour
```

### Acquisition Strategy (SaaS)

#### Month 1-3: Product Validation

- Launch free tier
- Target 50 early adopters
- Focus on developer experience
- Collect feedback for improvements

#### Month 4-6: Growth

- Optimize onboarding funnel
- Achieve 500+ free tier users
- 50+ paying customers
- Free-to-paid conversion rate: 10%
- MRR: $5K

#### Month 7-12: Scale

- Launch self-serve upgrade flow
- Implement automated onboarding
- Target 5K+ free users
- 500+ paying customers
- MRR: $25K
- Customer acquisition cost (CAC): $150
- CAC payback period: 2-3 months

### Retention & Expansion

#### Churn Prevention

- Monitor for inactive zones
- Proactive outreach at 30-day mark
- Success guides for each use case
- Community support (Discord, GitHub)

#### Expansion Revenue (Upsell)

- Email campaign when approaching limits
- "Upgrade suggestion" in dashboard
- Feature gating (advanced features need upgrade)
- Target: 25% of Pro users upgrade to Enterprise

---

## Self-Hosted Enterprise Licensing

### Licensing Model

| License        | Annual Price | Employees | Deployment   | Support   |
| -------------- | ------------ | --------- | ------------ | --------- |
| **Startup**    | $1,999       | <50       | Single AZ    | Community |
| **Growth**     | $9,999       | <500      | Multi-AZ     | Standard  |
| **Enterprise** | Custom       | Unlimited | Multi-region | Premium   |

### License Terms

```
Standard Enterprise License:
├── Term: 3 years (automatic renewal)
├── Usage: Unlimited events, zones, flows
├── Deployment: Dedicated infrastructure
├── Support: 24/7/365 with 1-hour response SLA
├── Upgrades: 2x per year (security, features)
├── Training: Annual on-site training (Startup: virtual only)
└── Consulting: 40 hours/year included

Optional Add-ons:
├── Multi-region failover: +$25K/year
├── Advanced compliance (FedRAMP): +$50K/year
├── Custom development: $250/hour
├── Managed services: +$15K/month
└── White-label deployment: +$75K/year
```

### Deployment Scenarios & Pricing

#### Scenario 1: Single-Region (AWS)

```
License: $9,999/year (Growth tier)
Infrastructure: ~$10K/month
  - RDS PostgreSQL: $3K/month
  - EKS Kubernetes: $4K/month
  - MSK Kafka: $2K/month
  - Data transfer: $1K/month
Total Annual Cost: ~$129K
```

#### Scenario 2: Multi-Region (AWS)

```
License: $35K/year (Enterprise)
Infrastructure: ~$25K/month (3 regions)
  - Multi-region RDS: $9K/month
  - Multi-region EKS: $10K/month
  - Multi-region Kafka: $4K/month
  - Data transfer: $2K/month
Total Annual Cost: ~$335K
```

#### Scenario 3: On-Premise

```
License: $25K/year (Enterprise)
Infrastructure: ~$30K/month (capital expense, amortized)
  - Hardware: $200K (amortized 5 years = $3.3K/month)
  - Personnel: 2 FTE ops = $20K/month
  - Networking/SAN: $3K/month
  - Maintenance: $3.7K/month
Total Annual Cost: ~$445K
```

### Sales Strategy (Enterprise)

#### Lead Generation

- Analyst reports & research (Gartner G2)
- Industry conferences (Fintech, Enterprise Software)
- Inbound marketing (SEO, content)
- Outbound outreach (LinkedIn, cold email)
- Customer references & case studies

#### Sales Process (120-180 day cycle)

```
Week 1-4: Discovery & Qualification
├── Initial call with procurement/CTO
├── Identify pain points & requirements
├── Provide ROI calculator
└── Send requirements questionnaire

Week 5-8: Solution Design
├── Detailed technical review
├── Security & compliance assessment
├── Infrastructure planning
├── Cost estimation

Week 9-12: RFP/RFQ Response
├── Formal proposal submission
├── Legal & procurement negotiations
├── Contract review with counsel

Week 13-16: Deployment & Training
├── Infrastructure provisioning
├── Software deployment
├── Security validation & penetration testing
├── Staff training

Week 17+: Go-Live & Managed Growth
├── Production cutover
├── 30-60-90 day check-ins
├── Optimization recommendations
```

---

## Customer Acquisition Journey

### SaaS Funnel (Month 1)

```
Marketing → Landing Page → Sign Up → Onboarding → First Event → First Flow
100%         20%            5%         80%        60%            40%

Target Metrics:
- Landing page conversion: 20% (CTA click)
- Sign-up completion: 25% (create account + zone)
- Activation: 60% (emit first event)
- Retention: 40% (build first flow)
```

### SaaS Customer Lifecycle

```
Month 1: Onboarding
├── Welcome email series
├── Guided setup tour
├── Usage tracking & alerts
└── Success metrics: 60% activation

Month 2-3: Growth
├── Feature education (flows, integrations)
├── Best practices guide
├── Community engagement
└── Success metrics: 40% retention

Month 4-6: Expansion
├── Email about tier limits
├── Product webinars
├── Premium feature trials
└── Success metrics: 25% free-to-paid

Month 7+: Retention & Advocacy
├── Regular feature releases
├── Dedicated support
├── Upgrade opportunities
├── Referral program
└── Success metrics: <5% monthly churn
```

### Enterprise Sales Cycle

```
T+0 Days: First Inquiry
├── Sales development rep (SDR) qualifies lead
├── Schedule discovery call with account executive (AE)
└── Send case studies & references

T+7 Days: Discovery Call
├── Present 3-5 use cases relevant to prospect
├── Understand technical requirements
├── Identify decision makers
└── Schedule demo with technical team

T+14 Days: Technical Demo
├── Live deployment demo (AWS/on-prem)
├── Security & compliance walkthroughs
├── Q&A with engineering
└── Provide technical requirements doc

T+30 Days: Proposal & RFP
├── Submit formal proposal
├── Detailed pricing & license terms
├── Compliance documentation
├── 30-day negotiation window

T+60 Days: Legal & Contracting
├── Standard enterprise agreement
├── DPA (Data Processing Agreement) for GDPR
├── BAA (Business Associate Agreement) for HIPAA
└── Executive sign-off

T+75 Days: Project Kickoff
├── Implementation plan & timeline
├── Infrastructure provisioning
├── Security validation
└── Training schedule

T+120+ Days: Go-Live
├── Soft launch to staging
├── Gradual traffic migration
├── Production validation
└── Executive handoff to customer success
```

---

## Revenue Projections

### Year 1 SaaS Projections

```
Q1 (Launch)
├── Free users: 500
├── Paying customers: 20
├── ARPU (Average Revenue Per User): $45
└── MRR: $900

Q2
├── Free users: 2K
├── Paying customers: 75
├── ARPU: $55
└── MRR: $4,125

Q3
├── Free users: 5K
├── Paying customers: 250
├── ARPU: $65
└── MRR: $16,250

Q4
├── Free users: 10K
├── Paying customers: 500
├── ARPU: $75
└── MRR: $37,500

Year 1 Total SaaS Revenue: ~$120K (from Q1-Q4 average MRR)
```

### Year 1 Self-Hosted Projections

```
Q1 (Launch)
├── Enterprise customers: 0
└── ARR: $0

Q2
├── Enterprise customers: 1 (Growth tier)
└── ARR: $10K

Q3
├── Enterprise customers: 2 (1x Growth, 1x Enterprise)
└── ARR: $45K

Q4
├── Enterprise customers: 3 (1x Growth, 2x Enterprise)
└── ARR: $100K

Year 1 Total Self-Hosted Revenue: ~$155K ARR
```

### Combined Year 1 Revenue

```
SaaS MRR (end of year): $37.5K
SaaS ARR (end of year): $450K

Self-Hosted ARR: $100K

Total YoY Revenue: ~$550K (Year 1)

Gross Margin (SaaS): 75% (infrastructure + personnel)
Gross Margin (Self-Hosted): 85% (license only, no infrastructure)

Blended Gross Margin: 78%
```

### Year 2-3 Projections

```
Year 2:
├── SaaS: $1.5M ARR (3.3x growth)
├── Self-Hosted: $500K ARR (5x growth)
└── Total: $2M ARR

Year 3:
├── SaaS: $4M ARR (2.7x growth)
├── Self-Hosted: $1.5M ARR (3x growth)
└── Total: $5.5M ARR
```

---

## Competitive Positioning

### Competitive Landscape

| Competitor          | SaaS Model | Self-Hosted | Fintech-Focus | Open-Source | Pricing Model   |
| ------------------- | ---------- | ----------- | ------------- | ----------- | --------------- |
| **Zapier**          | ✅ Yes     | ❌ No       | ❌ No         | ❌ No       | Usage-based     |
| **Make.com**        | ✅ Yes     | ✅ Yes      | ❌ No         | ❌ No       | Usage-based     |
| **IFTTT**           | ✅ Yes     | ❌ No       | ❌ No         | ❌ No       | Freemium        |
| **Stripe Webhooks** | ✅ Yes     | ✅ Native   | ✅ Yes        | ❌ No       | Per transaction |
| **n8n**             | ✅ Yes     | ✅ Yes      | ❌ No         | ✅ Yes      | Open-source     |
| **Sapliy**          | ✅ Yes     | ✅ Yes      | ✅ Yes        | ✅ Yes      | Hybrid          |

### Sapliy's Unique Value Proposition

```
1. Hybrid-First Architecture
   └─ One codebase works SaaS + Self-hosted
   └─ Customers don't need to choose upfront
   └─ Easy migration path as company scales

2. Fintech-Optimized
   └─ Built for payment workflows
   └─ Ledger & audit trail first-class citizens
   └─ Compliance by design (HIPAA, PCI-DSS ready)

3. Open-Source Foundation
   └─ Community trust & contributions
   └─ Transparent development
   └─ No vendor lock-in

4. Developer-First
   └─ SDKs in Node, Python, Go
   └─ Simple API (emit → build flow)
   └─ Test/live mode built-in

5. Zone Isolation
   └─ Separate credentials, logs, flows per zone
   └─ No risk of mixing test/prod
   └─ Industry-standard approach
```

### Positioning Against Competitors

#### vs. Zapier

```
Zapier: $20-$600/month (SaaS only)
Sapliy: $0-$99/month SaaS, $2K-$500K+ Self-hosted

Advantage Sapliy:
✅ Self-hosted option for enterprises
✅ Fintech-specific features (ledger, approvals)
✅ Open-source foundation
✅ Lower cost for high-volume users

Advantage Zapier:
✅ 7000+ pre-built connectors
✅ Larger community & ecosystem
✅ 10+ year track record
```

#### vs. n8n

```
n8n: Free (self-hosted), $20-$250/month (cloud)
Sapliy: Free-$99/month SaaS, $2K-$500K+ Self-hosted

Advantage Sapliy:
✅ Managed SaaS option
✅ Enterprise support & SLA
✅ Fintech-specific features
✅ Easier compliance path (HIPAA-ready)

Advantage n8n:
✅ Lower self-hosted cost
✅ More open-source community
✅ Larger connector library
```

---

---

## CLI-First Developer Experience

### The sapliy-cli Vision

The **sapliy-cli** is the unified entry point for all developer interactions with Sapliy. It provides:

- 🔐 Authentication & key management
- 🏃 Local automation server (self-hosted)
- 🎨 Frontend UI launcher (flow builder)
- 🧪 Testing & debugging tools
- 🔌 Webhook inspection & replay
- 📊 Log streaming & monitoring

**Goal**: Developers can build, test, and deploy flows without switching between 5+ different tools.

### Command Structure

#### Core Commands

```bash
# Authentication
sapliy login                          # Authenticate with Sapliy (SaaS or self-hosted)
sapliy logout                         # Remove stored credentials
sapliy whoami                         # Show current user & zone

# Development Server
sapliy run                            # Start local backend + event processor
sapliy frontend                       # Launch flow builder UI (http://localhost:3000)
sapliy dev                            # Start backend + frontend + watch mode (all-in-one)

# Zone Management
sapliy zones list                     # List all zones
sapliy zones create --name="my-app"   # Create new zone
sapliy zones switch --zone="prod"     # Switch active zone
sapliy zones export                   # Export zone configuration

# Flow Management
sapliy flows list                     # List flows in current zone
sapliy flows create --name="checkout" # Create flow interactively
sapliy flows deploy --flow="checkout" # Deploy flow to live mode
sapliy flows test --flow="checkout"   # Test flow locally
sapliy flows logs --flow="checkout"   # Stream flow execution logs

# Event Management
sapliy events emit "payment.completed" '{"amount":100}' # Emit test event
sapliy events listen "payment.*"      # Listen to events in real-time
sapliy events replay --after="2024-01-15" # Replay historical events

# Webhook Management
sapliy webhooks listen                # Start webhook inspector (http://localhost:9000)
sapliy webhooks test --url="http://..." # Test webhook delivery
sapliy webhooks replay --id="evt_xxx" # Replay webhook delivery

# Testing
sapliy test --flow="checkout"         # Run flow tests
sapliy test --all                     # Run all flow tests
sapliy test --coverage                # Show code coverage

# Monitoring
sapliy logs --follow                  # Stream all service logs
sapliy metrics                        # Show performance metrics
sapliy health                         # Check service health

# Configuration
sapliy config get <key>               # Get config value
sapliy config set <key> <value>       # Set config value
sapliy env --export                   # Export environment variables
```

#### Advanced Options

```bash
# Port/Host configuration
sapliy run --port 8080 --host 0.0.0.0

# Multi-service control
sapliy run --services postgres,redis,kafka # Start specific services only
sapliy run --skip-postgres                 # Skip specific services

# Frontend options
sapliy frontend --port 3000 --auto-open    # Auto-open browser
sapliy frontend --prod-endpoint "https://api.sapliy.io" # Use SaaS backend

# Development mode
sapliy dev --watch                    # Watch for file changes, auto-reload
sapliy dev --debug                    # Enable debug logging
sapliy dev --seed                     # Load sample data on startup
```

### Architecture: CLI Entry Point Design

```
sapliy-cli/
│
├─ bin/
│   └─ sapliy                   # Main CLI executable
│
├─ commands/
│   ├─ auth/
│   │   ├─ login.js
│   │   ├─ logout.js
│   │   └─ whoami.js
│   │
│   ├─ dev/
│   │   ├─ run.js               # Start backend services
│   │   ├─ frontend.js          # Launch flow builder UI
│   │   ├─ dev.js               # Combined backend + frontend
│   │   ├─ test.js              # Run flow tests
│   │   └─ logs.js              # Stream logs
│   │
│   ├─ zones/
│   │   ├─ list.js
│   │   ├─ create.js
│   │   ├─ switch.js
│   │   └─ export.js
│   │
│   ├─ flows/
│   │   ├─ list.js
│   │   ├─ create.js
│   │   ├─ deploy.js
│   │   └─ test.js
│   │
│   ├─ events/
│   │   ├─ emit.js
│   │   ├─ listen.js
│   │   └─ replay.js
│   │
│   ├─ webhooks/
│   │   ├─ listen.js
│   │   ├─ test.js
│   │   └─ replay.js
│   │
│   └─ config/
│       ├─ get.js
│       └─ set.js
│
├─ services/
│   ├─ docker.js                # Docker/container management
│   ├─ ports.js                 # Port detection & management
│   ├─ logger.js                # Unified logging
│   ├─ config-loader.js         # Load sapliy.json config
│   └─ subprocess-manager.js    # Manage child processes
│
├─ utils/
│   ├─ auth.js                  # JWT/key management
│   ├─ http-client.js           # API calls to backend
│   ├─ ws-client.js             # WebSocket for event streaming
│   ├─ docker-compose.js        # Docker Compose helper
│   ├─ spinner.js               # CLI animations
│   └─ table.js                 # Formatted table output
│
├─ config/
│   └─ defaults.json            # Default ports, endpoints
│
├─ templates/
│   ├─ sapliy.json              # Config template
│   ├─ Dockerfile.dev           # Development Docker setup
│   └─ docker-compose.dev.yml   # Multi-service compose
│
└─ package.json
```

### Key Implementation Details

#### 1. Backend Launcher (run.js)

```javascript
// commands/dev/run.js
const { spawn } = require("child_process");
const path = require("path");
const fs = require("fs");
const { findFreePort } = require("../../utils/ports");
const { Logger } = require("../../services/logger");

const logger = new Logger("sapliy:run");

async function runBackend(options = {}) {
  const {
    port = 8080,
    host = "localhost",
    services = ["postgres", "redis", "kafka"],
    skipDocker = false,
    watch = false,
  } = options;

  logger.info("🚀 Starting Sapliy backend services...");

  // 1. Check if docker-compose is available
  if (!skipDocker && !hasDockerCompose()) {
    logger.error("Docker Compose not found. Install it or use --skip-docker");
    return;
  }

  // 2. Find available ports
  const ports = {
    api: await findFreePort(port),
    postgres: await findFreePort(5432),
    redis: await findFreePort(6379),
    kafka: await findFreePort(9092),
  };

  // 3. Load or create .env.local
  const envPath = path.resolve(process.cwd(), ".env.local");
  const env = loadEnv(envPath);
  env.API_PORT = ports.api;
  env.DATABASE_URL = `postgresql://postgres:password@localhost:${ports.postgres}/sapliy`;
  env.REDIS_URL = `redis://localhost:${ports.redis}`;
  env.KAFKA_BROKERS = `localhost:${ports.kafka}`;
  saveEnv(envPath, env);

  // 4. Start docker-compose
  if (!skipDocker) {
    const dockerProcess = spawn(
      "docker-compose",
      ["-f", "docker-compose.dev.yml", "up", "--remove-orphans"],
      {
        stdio: "inherit",
        env: { ...process.env, ...env },
      },
    );

    dockerProcess.on("error", (err) => {
      logger.error(`Docker Compose failed: ${err.message}`);
    });
  }

  // 5. Start API server
  logger.info(`✅ Starting API server on http://${host}:${ports.api}`);

  const apiProcess = spawn("npm", ["run", "start:api"], {
    cwd: path.resolve(__dirname, "../../..", "fintech-ecosystem"),
    stdio: "inherit",
    env: { ...process.env, ...env },
  });

  // 6. Handle graceful shutdown
  process.on("SIGINT", () => {
    logger.info("Shutting down services...");
    apiProcess.kill();
    if (!skipDocker) {
      spawn("docker-compose", ["down"], { stdio: "inherit" });
    }
    process.exit(0);
  });

  logger.info("✨ All services running!");
  logger.info("");
  logger.info("  API:      http://${host}:${ports.api}");
  logger.info("  Postgres: localhost:${ports.postgres}");
  logger.info("  Redis:    localhost:${ports.redis}");
  logger.info("  Kafka:    localhost:${ports.kafka}");
  logger.info("");
  logger.info("Run `sapliy frontend` in another terminal to launch the UI");
}

module.exports = { runBackend };
```

#### 2. Frontend Launcher (frontend.js)

```javascript
// commands/dev/frontend.js
const { spawn } = require("child_process");
const path = require("path");
const open = require("open");
const { Logger } = require("../../services/logger");
const { findFreePort } = require("../../utils/ports");

const logger = new Logger("sapliy:frontend");

async function launchFrontend(options = {}) {
  const {
    port = 3000,
    host = "localhost",
    autoOpen = true,
    prodEndpoint = null,
  } = options;

  logger.info("🎨 Launching Sapliy Flow Builder...");

  // 1. Find available port
  const availablePort = await findFreePort(port);

  // 2. Set environment variables
  const env = {
    ...process.env,
    PORT: availablePort,
    VITE_API_ENDPOINT: prodEndpoint || `http://localhost:8080`,
    VITE_MODE: "development",
  };

  // 3. Launch frontend dev server
  const frontendPath = path.resolve(
    __dirname,
    "../../..",
    "fintech-automation",
  );

  logger.info(`Starting dev server on http://${host}:${availablePort}`);

  const frontendProcess = spawn("npm", ["run", "dev"], {
    cwd: frontendPath,
    stdio: "inherit",
    env,
  });

  // 4. Auto-open browser
  if (autoOpen) {
    setTimeout(() => {
      logger.info(`🌐 Opening browser...`);
      open(`http://${host}:${availablePort}`);
    }, 2000);
  }

  // 5. Handle shutdown
  frontendProcess.on("error", (err) => {
    logger.error(`Frontend process failed: ${err.message}`);
  });

  frontendProcess.on("close", (code) => {
    if (code !== 0) {
      logger.error(`Frontend exited with code ${code}`);
    }
  });

  logger.info("✨ Frontend running!");
  logger.info(`   → http://${host}:${availablePort}`);
}

module.exports = { launchFrontend };
```

#### 3. Combined Dev Command (dev.js)

```javascript
// commands/dev/dev.js
const { runBackend } = require("./run");
const { launchFrontend } = require("./frontend");
const { Logger } = require("../../services/logger");

const logger = new Logger("sapliy:dev");

async function devMode(options = {}) {
  logger.info("🚀 Starting Sapliy in development mode...");
  logger.info("");

  const backendOptions = {
    port: options.apiPort || 8080,
    watch: options.watch !== false,
    ...options,
  };

  const frontendOptions = {
    port: options.uiPort || 3000,
    autoOpen: options.autoOpen !== false,
    prodEndpoint: options.prodEndpoint,
  };

  try {
    // Start both in parallel
    await Promise.all([
      runBackend(backendOptions),
      new Promise((resolve) => setTimeout(resolve, 3000)) // Wait for backend to start
        .then(() => launchFrontend(frontendOptions)),
    ]);
  } catch (error) {
    logger.error(`Development mode failed: ${error.message}`);
    process.exit(1);
  }
}

module.exports = { devMode };
```

### Professional Enhancements

#### 1. Automatic Port Detection

```javascript
// utils/ports.js
const net = require("net");

async function findFreePort(preferredPort = 3000) {
  return new Promise((resolve) => {
    const server = net.createServer();
    server.listen(preferredPort, () => {
      const { port } = server.address();
      server.close(() => resolve(port));
    });
    server.on("error", () => {
      // Port in use, try next one
      resolve(findFreePort(preferredPort + 1));
    });
  });
}
```

#### 2. Unified Logging System

```javascript
// services/logger.js
class Logger {
  constructor(namespace) {
    this.namespace = namespace;
  }

  info(message) {
    console.log(`  ${this.namespace} ℹ️  ${message}`);
  }

  success(message) {
    console.log(`  ${this.namespace} ✅ ${message}`);
  }

  warn(message) {
    console.warn(`  ${this.namespace} ⚠️  ${message}`);
  }

  error(message) {
    console.error(`  ${this.namespace} ❌ ${message}`);
  }

  debug(message) {
    if (process.env.DEBUG) {
      console.log(`  ${this.namespace} 🐛 ${message}`);
    }
  }

  // Pretty print tables
  table(data) {
    console.table(data);
  }
}
```

#### 3. Configuration Management

```javascript
// sapliy.json (in project root)
{
  "name": "my-sapliy-app",
  "version": "1.0.0",
  "sapliy": {
    "apiPort": 8080,
    "uiPort": 3000,
    "services": ["postgres", "redis", "kafka"],
    "environment": {
      "LOG_LEVEL": "info",
      "TEST_MODE": true
    },
    "integrations": [
      {
        "type": "stripe",
        "testKey": "sk_test_..."
      }
    ],
    "flows": [
      "./flows/payment.json",
      "./flows/notifications.json"
    ]
  }
}
```

### Developer Experience: Real-World Workflow

#### Scenario 1: Fresh Start (No Backend Running)

```bash
$ cd my-sapliy-app
$ sapliy dev

✨ Starting Sapliy in development mode...

  sapliy:run ✅ Checking Docker...
  sapliy:run ✅ Creating .env.local
  sapliy:run 🐳 Starting containers...
  sapliy:run ✅ PostgreSQL running on localhost:5432
  sapliy:run ✅ Redis running on localhost:6379
  sapliy:run ✅ Kafka running on localhost:9092
  sapliy:run ✅ API server running on http://localhost:8080

  sapliy:frontend ✅ Starting dev server...
  sapliy:frontend 🌐 Opening browser to http://localhost:3000

✨ All systems running!
   API:      http://localhost:8080
   Frontend: http://localhost:3000

Press Ctrl+C to stop
```

#### Scenario 2: Test an Event

```bash
$ sapliy events emit "payment.completed" '{
  "orderId": "12345",
  "amount": 99.99,
  "currency": "USD"
}'

✅ Event emitted: evt_abc123
   Event ID: evt_abc123
   Zone: zone_test_xxx
   Type: payment.completed

🔄 Executing flows...
   Flow: send_confirmation_email (success)
   Flow: update_accounting (success)

📊 Results:
   ✅ 2/2 flows executed successfully
```

#### Scenario 3: Stream Logs

```bash
$ sapliy logs --follow

sapliy:api [10:23:45] GET /health 200
sapliy:api [10:23:46] POST /events 201
sapliy:flow-engine [10:23:46] Executing flow: send_confirmation_email
sapliy:flow-engine [10:23:47] Webhook sent to https://example.com/webhook (200 OK)
sapliy:api [10:23:47] POST /webhooks/callback 200
```

### Revenue Implications

#### For SaaS

- ✅ Lower onboarding friction (5 min to first event vs. 15 min without CLI)
- ✅ Improved developer experience → higher conversion (free → paid)
- ✅ Faster feedback loop → product-market fit
- ✅ Community contributions to CLI

#### For Self-Hosted

- ✅ Professional developer experience (comparable to Node-RED)
- ✅ Easier evaluation for enterprises
- ✅ Faster time-to-value
- ✅ CLI-first enables infrastructure automation

#### For Monetization

- ✅ **Advanced CLI features** (enterprise flag):
  - `sapliy audit-export` (compliance)
  - `sapliy performance-profile` (analytics)
  - `sapliy multi-region-sync` (HA)
  - Pricing: $49-$299/month add-on
- ✅ **Professional Services**: Setup & configuration via CLI
  - Pricing: $200-$350/hour

---

## Conclusion

Sapliy's hybrid SaaS + self-hosted model creates multiple revenue streams while serving the complete market:

- **SaaS** captures fast-moving startups & SMBs (high volume, low margin)
- **Self-Hosted** captures enterprises & regulated companies (low volume, high margin)
- **One codebase** reduces engineering burden & increases agility
- **Open-source** builds community trust & accelerates adoption
- **Fintech focus** carves out a differentiated niche
- **CLI-first** approach drives adoption & professional positioning

**Target metrics for success:**

- Year 1: $550K revenue, 500+ SaaS customers, 3 Enterprise customers
- Year 3: $5.5M revenue, 5K+ SaaS customers, 20+ Enterprise customers
- CAC Payback: 3-4 months (SaaS), 12-18 months (Enterprise)
- Gross Margin: 78% blended
- CLI adoption: 80% of developers use sapliy-cli for development
