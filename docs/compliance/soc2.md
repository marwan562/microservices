# SOC 2 Compliance Controls

This document maps the Trust Services Criteria (TSC) to our technical and procedural controls.

## 1. Security (Common Criteria)

### CC1. Control Environment
- **CC1.1**: We maintain a Code of Conduct and Ethics policy signed by all employees.
- **CC1.2**: Background checks are performed for all engineering and operations staff.

### CC6. Logical and Physical Access
- **CC6.1**: Access to production infrastructure (AWS, K8s) is restricted to the DevOps team via SSO.
- **CC6.2**: User access reviews are performed quarterly. Access is revoked within 24 hours of termination.
- **CC6.7**: We enforce MFA on all administrative accounts (AWS, GitHub, Google Workspace).

### CC7. System Operations
- **CC7.2**: Vulnerability scanning (AWS Inspector, Container Scanning) is automated in CI/CD.
- **CC7.4**: Incident Response Plan is tested annually.

### CC8. Change Management
- **CC8.1**: All code changes require a Pull Request with peer review and passing automated tests.
- **CC8.2**: Production deployments are gated by approval and performed via automated pipelines (ArgoCD).

## 2. Availability

- **A1.1**: We maintain an uptime SLA of 99.9%.
- **A1.2**: Architecture is deployed across multiple Availability Zones (AZs) in AWS.
- **A1.3**: Database backups are performed daily and tested for restoration monthly.
- **A1.4**: Capacity planning reviews occur quarterly to ensure scalability.

## 3. Confidentiality

- **C1.1**: All sensitive data is identified effectively classified.
- **C1.2**: Data is encrypted at rest (AES-256) and in transit (TLS 1.2+).
- **C1.3**: Secrets are managed via AWS Secrets Manager and never committed to code.

## 4. Processing Integrity

- **PI1.1**: The Ledger utilizes double-entry accounting to ensure mathematical correctness (Assets = Liabilities + Equity).
- **PI1.2**: Input validation (Joi/Validator schemas) is enforced on all API endpoints.
- **PI1.3**: Idempotency keys are required for all non-safe HTTP methods (POST, PUT) to prevent duplicate processing.

## 5. Privacy

- **P1.1**: Privacy Policy is published and versioned.
- **P2.1**: We collect only the minimum data required to process transactions.
- **P4.3**: Mechanisms exist to delete user data upon request (Right to be Forgotten/GDPR).
