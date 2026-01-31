# PCI-DSS Compliance (SAQ A-EP)

As a Service Provider that impacts the security of payment transactions (even if not storing raw PAN data directly), we align with PCI-DSS Self-Assessment Questionnaire A-EP.

## 1. Build and Maintain a Secure Network

- **Requirement 1: Install and maintain a firewall configuration.**
    - Implemented via AWS Security Groups and Network ACLs.
    - Default policy: Deny All. Explicitly allow only necessary ports (443).
    - Review of rules every 6 months.

- **Requirement 2: Do not use vendor-supplied defaults.**
    - All default passwords/accounts removed from containers.
    - Infrastructure provisioning (Terraform) sets custom unique passwords/keys.

## 2. Protect Cardholder Data

- **Requirement 3: Protect stored cardholder data.**
    - **We do not store PAN (Primary Account Number) or CVV.**
    - Tokens and benign data (Last 4 digits, expiration) allow for reconciliation without risk.
    - Any sensitive config (API keys) is encrypted via AWS KMS.

- **Requirement 4: Encrypt transmission of cardholder data across open, public networks.**
    - TLS 1.2 or 1.3 is enforced for all public endpoints.
    - Strong ciphers only (no weak/deprecated suites).

## 3. Maintain a Vulnerability Management Program

- **Requirement 5: Protect all systems against malware.**
    - Linux-based containerized environment minimizes malware risk.
    - Endpoint protection (e.g., CrowdStrike/SentinelOne) on nodes if managed.

- **Requirement 6: Develop and maintain secure systems and applications.**
    - Security training for developers (OWASP Top 10).
    - Code reviews mandatory for all changes.
    - Dependencies scanned for known CVEs (Dependabot/Snyk).

## 4. Implement Strong Access Control Measures

- **Requirement 7: Restrict access to cardholder data by business need to know.**
    - RBAC implemented in Kubernetes and AWS IAM.

- **Requirement 8: Identify and authenticate access to system components.**
    - Unique ID for each person.
    - MFA required for all access.
    - Sessions time out after 15 minutes of inactivity.

- **Requirement 9: Restrict physical access to cardholder data.**
    - We rely on AWS's physical security (SOC 2 Type II / PCI certified data centers).

## 5. Regularly Monitor and Test Networks

- **Requirement 10: Track and monitor all access to network resources and cardholder data.**
    - Audit logs enabled for all critical systems (AWS CloudTrail, K8s Audit Logs, Application Logs).
    - Logs sent to centralized aggregation (CloudWatch/Loki) and retained for 1 year.

- **Requirement 11: Regularly test security systems and processes.**
    - Quarterly internal vulnerability scans.
    - Annual Penetration Test by a third party.

## 6. Maintain an Information Security Policy

- **Requirement 12: Maintain a policy that addresses information security.**
    - Information Security Policy published and reviewed annually.
    - Incident Response Plan defined.
