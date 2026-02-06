# Disaster Recovery Runbook

## 🚨 Emergency Contacts
- **Incident Commander**: oncall-ic@sapliy.com
- **Database Team**: dba-emergency@sapliy.com
- **Infrastructure Team**: infra-sre@sapliy.com

## 🛑 Maintenance Mode
To enable maintenance mode during an incident:

```bash
# Enable maintenance mode
curl -X POST https://api.sapliy.com/admin/maintenance/enable \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"message": "System undergoing emergency maintenance", "duration_minutes": 60}'

# Whitelist your IP
curl -X POST https://api.sapliy.com/admin/maintenance/allow \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"ip": "YOUR_OFFICE_IP"}'
```

## 🔄 Database Failover (PostgreSQL)

### Scenario: Primary Region Failure

1. **Verify Primary is Down**
   Check CloudWatch/Datadog metrics for RDS Primary.

2. **Promote Read Replica**
   ```bash
   aws rds promote-read-replica \
     --db-instance-identifier sapliy-db-replica-us-west-2 \
     --region us-west-2
   ```

3. **Update Secrets**
   Update the secret in AWS Secrets Manager to point to the new writer endpoint.
   ```bash
   aws secretsmanager put-secret-value \
     --secret-id prod/sapliy/database \
     --secret-string '{"host": "sapliy-db-replica-us-west-2.cx...", ...}'
   ```

4. **Restart Services**
   Restart Kubernetes pods to pick up the new configuration (if not dynamic).
   ```bash
   kubectl rollout restart deployment/sapliy-api -n production
   ```

## 🔄 Region Failover (Active-Passive)

1. **Update DNS (Route53)**
   Switch the `api.sapliy.com` alias record to the secondary region load balancer.

2. **Scale Up Secondary Region**
   ```bash
   kubectl scale deployment/sapliy-api --replicas=50 -n production --context=arn:aws:eks:us-west-2...
   ```

3. **Enable Maintenance Mode in Secondary**
   Prevent inconsistent writes while verifying data integrity.

4. **Verify Data Integrity**
   Run the consistency check script:
   ```bash
   ./scripts/verify-consistency.sh --region us-west-2
   ```

5. **Disable Maintenance Mode**
   Open traffic to users.

## 📉 Recovery Point Objective (RPO) Validation

**Goal**: < 5 minutes data loss.

1. Check replication lag metric: `rds.replicaLag`.
2. If > 5 minutes, seek executive approval before failover unless automated.

## ⏱️ Recovery Time Objective (RTO) Validation

**Goal**: < 15 minutes downtime.

1. DNS TTL is set to 60s.
2. Database promotion takes ~5-10 mins.
3. Service restart takes ~2 mins.
