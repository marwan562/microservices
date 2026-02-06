# 🚀 Sapliy Production Deployment Checklist

> Step-by-step guide to deploy Sapliy to production with confidence

---

## Pre-Deployment (Week 1)

### Code Quality & Testing

- [ ] **All tests passing**
  ```bash
  npm run test:all
  # Expected: All unit, integration, E2E tests pass
  ```

- [ ] **Code coverage >80%**
  ```bash
  npm run test -- --coverage
  # Expected: branches 80%, functions 80%, lines 80%
  ```

- [ ] **TypeScript strict mode enabled**
  ```json
  {
    "compilerOptions": {
      "strict": true,
      "noImplicitAny": true,
      "strictNullChecks": true
    }
  }
  ```

- [ ] **No linting errors**
  ```bash
  npm run lint
  # Expected: No errors
  ```

- [ ] **Security vulnerabilities fixed**
  ```bash
  npm run security-check
  # Expected: 0 vulnerabilities
  ```

- [ ] **Dependencies up-to-date**
  ```bash
  npm audit fix
  npm update
  ```

### Documentation

- [ ] **API documentation complete** (`/docs/api`)
- [ ] **Deployment guide finalized** (`ENTERPRISE_GUIDE.md`)
- [ ] **Runbooks created** (`/ops/runbooks`)
- [ ] **Troubleshooting guide done** (`/docs/troubleshooting`)
- [ ] **Architecture diagram updated** (`ARCHITECTURE.md`)
- [ ] **CLI documentation complete** (`QUICK_REFERENCE.md`)

### Security Hardening

- [ ] **Secrets management configured**
  - [ ] All secrets in `.env.production` (not in code)
  - [ ] AWS Secrets Manager / HashiCorp Vault ready
  - [ ] Secret rotation policy defined

- [ ] **Database security**
  - [ ] SSL/TLS enabled for DB connections
  - [ ] Database backups configured
  - [ ] Point-in-time recovery tested
  - [ ] Least privilege database user configured

- [ ] **API security**
  - [ ] Rate limiting configured
  - [ ] CORS properly restricted
  - [ ] API key rotation policy ready
  - [ ] Webhook signature verification enabled

- [ ] **Infrastructure security**
  - [ ] Firewall rules configured
  - [ ] Network segmentation done
  - [ ] VPC setup completed
  - [ ] Security group rules reviewed

---

## Docker & Container Preparation (Week 1-2)

### Docker Images

- [ ] **All Dockerfile's optimized**
  ```dockerfile
  # Multi-stage build
  FROM node:18 AS builder
  # ... build stage
  FROM node:18-alpine
  # ... runtime stage (smaller image)
  ```

- [ ] **Images scanned for vulnerabilities**
  ```bash
  docker scan sapliy:latest
  # Expected: 0 critical vulnerabilities
  
  trivy image sapliy:latest
  # Expected: 0 critical, <5 high
  ```

- [ ] **Images tagged properly**
  ```bash
  docker tag sapliy:latest sapliy:1.0.0
  docker tag sapliy:latest sapliy:stable
  docker push sapliy:1.0.0
  docker push sapliy:stable
  ```

- [ ] **Image sizes optimized** (<200MB base)
  ```bash
  docker images | grep sapliy
  # Expected: sapliy:1.0.0 ~150MB
  ```

### Docker Compose

- [ ] **docker-compose.yml reviewed for production**
  - [ ] Resource limits set
  - [ ] Restart policies configured
  - [ ] Health checks defined
  - [ ] Environment variables externalized

- [ ] **All services have health checks**
  ```yaml
  services:
    api:
      healthcheck:
        test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
        interval: 10s
        timeout: 5s
        retries: 5
  ```

- [ ] **Networking configured**
  - [ ] Services on internal network
  - [ ] Only API exposed to internet
  - [ ] Database not directly exposed

- [ ] **Volumes & persistence**
  - [ ] Database persistence configured
  - [ ] Backup volumes mounted
  - [ ] Data encryption enabled

---

## Infrastructure Setup (Week 2)

### Cloud Provider Selection

- [ ] **AWS / GCP / Azure selected**
- [ ] **Budget approved & allocated**
- [ ] **Resource quotas requested** (if needed)

### Database Setup

- [ ] **Production PostgreSQL instance created**
  ```bash
  # AWS RDS example
  aws rds create-db-instance \
    --db-instance-identifier sapliy-prod \
    --db-instance-class db.t3.medium \
    --engine postgres \
    --engine-version 15.2 \
    --allocated-storage 100 \
    --storage-type gp3 \
    --backup-retention-period 30
  ```

- [ ] **Database backups configured**
  - [ ] Automated daily backups enabled
  - [ ] Point-in-time recovery (PITR) enabled
  - [ ] Backup retention: 30 days
  - [ ] Test restore procedure

- [ ] **Database replication** (if high-availability)
  - [ ] Read replicas created
  - [ ] Failover tested
  - [ ] Synchronous replication enabled

- [ ] **Database security**
  - [ ] Encryption at rest enabled
  - [ ] SSL/TLS connections enforced
  - [ ] Network access restricted to app servers
  - [ ] Database audit logging enabled

### Caching Layer (Redis)

- [ ] **Redis cluster created** (production)
  ```bash
  # AWS ElastiCache example
  aws elasticache create-cache-cluster \
    --cache-cluster-id sapliy-redis-prod \
    --cache-node-type cache.r6g.large \
    --engine redis \
    --engine-version 7.0 \
    --num-cache-nodes 3 \
    --automatic-failover-enabled
  ```

- [ ] **Redis persistence configured**
  - [ ] RDB snapshots enabled
  - [ ] AOF (Append-Only File) enabled
  - [ ] Backup location secured

- [ ] **Redis security**
  - [ ] AUTH token configured
  - [ ] Network isolation enforced
  - [ ] Encryption in transit (TLS) enabled

### Message Queue (Kafka)

- [ ] **Kafka cluster provisioned**
  ```bash
  # AWS MSK example
  aws kafka create-cluster \
    --cluster-name sapliy-events-prod \
    --broker-node-group-info \
      NumberOfBrokerNodes=3
  ```

- [ ] **Topics created & configured**
  ```bash
  kafka-topics --bootstrap-server localhost:9092 \
    --create --topic events \
    --partitions 10 \
    --replication-factor 3
  ```

- [ ] **Kafka security**
  - [ ] TLS encryption enabled
  - [ ] SASL authentication configured
  - [ ] Network policies enforced

### Kubernetes Setup (Optional)

- [ ] **EKS / GKE / AKS cluster created**
  ```bash
  # AWS EKS example
  eksctl create cluster \
    --name sapliy-prod \
    --region us-east-1 \
    --nodes 3 \
    --node-type t3.large
  ```

- [ ] **Ingress controller installed** (nginx)
- [ ] **Service mesh configured** (optional, Istio)
- [ ] **Persistent storage provisioned** (EBS)
- [ ] **Node auto-scaling configured**

---

## Application Deployment (Week 2-3)

### Environment Variables

- [ ] **All .env variables configured**
  ```bash
  # .env.production
  NODE_ENV=production
  PORT=8080
  DATABASE_URL=postgresql://user:pass@db:5432/sapliy
  REDIS_URL=redis://user:pass@cache:6379
  KAFKA_BROKERS=kafka:9092
  JWT_SECRET=<secret>
  WEBHOOK_SECRET=<secret>
  ```

- [ ] **Secrets stored securely**
  - [ ] AWS Secrets Manager
  - [ ] HashiCorp Vault
  - [ ] Azure Key Vault
  - [ ] Not in code or .env files

### API Server Deployment

- [ ] **Docker Compose deployment**
  ```bash
  docker-compose -f docker-compose.prod.yml up -d
  docker-compose logs -f api
  ```

- [ ] **Health checks passing**
  ```bash
  curl http://localhost:8080/health
  # Expected: { status: "ok" }
  ```

- [ ] **Graceful shutdown configured**
  ```javascript
  process.on('SIGTERM', async () => {
    await server.close();
    await db.disconnect();
    process.exit(0);
  });
  ```

### Database Migrations

- [ ] **All migrations tested**
  ```bash
  npm run migrate:up -- --dry-run
  npm run migrate:up
  npm run migrate:status
  ```

- [ ] **Rollback tested**
  ```bash
  npm run migrate:down
  npm run migrate:up
  ```

- [ ] **Data consistency verified**
  ```sql
  SELECT COUNT(*) FROM users; -- should match expected count
  ```

### Service Initialization

- [ ] **API server started**
- [ ] **Event processor started**
- [ ] **Webhook scheduler started**
- [ ] **Background job workers started**

---

## Monitoring & Logging Setup (Week 3)

### Logging

- [ ] **Centralized logging configured**
  - [ ] CloudWatch / ELK Stack / Datadog
  - [ ] Log rotation enabled
  - [ ] Log retention: 30 days

- [ ] **Log levels configured**
  ```javascript
  logger.setLevel(process.env.LOG_LEVEL || 'info');
  // Production: info level (not debug)
  ```

- [ ] **Structured logging implemented**
  ```json
  {
    "timestamp": "2024-01-15T10:30:00Z",
    "level": "error",
    "service": "api",
    "message": "Payment processing failed",
    "context": {
      "orderId": "ord_123",
      "error": "Payment gateway timeout"
    }
  }
  ```

### Monitoring & Metrics

- [ ] **Metrics collection configured** (Prometheus)
  - [ ] API response times
  - [ ] Event processing latency
  - [ ] Error rates
  - [ ] Database connection pool
  - [ ] Cache hit rates

- [ ] **Dashboards created** (Grafana)
  - [ ] System health dashboard
  - [ ] API performance dashboard
  - [ ] Business metrics dashboard

- [ ] **Alerts configured**
  - [ ] High error rate (>5%)
  - [ ] High API latency (>500ms p95)
  - [ ] Database connection exhaustion
  - [ ] Cache failure
  - [ ] Disk space running low (<20%)

### Error Tracking

- [ ] **Error tracking configured** (Sentry)
  ```javascript
  Sentry.init({
    dsn: process.env.SENTRY_DSN,
    environment: 'production',
    tracesSampleRate: 0.1,
  });
  ```

- [ ] **Alerts for critical errors**
- [ ] **Error grouping & deduplication** enabled
- [ ] **Source maps uploaded**

---

## Backup & Disaster Recovery (Week 3)

### Backup Strategy

- [ ] **Database backups automated**
  - [ ] Daily backups enabled
  - [ ] Retention: 30 days
  - [ ] Cross-region replication (optional)

- [ ] **Backup testing scheduled**
  ```bash
  # Monthly restore test
  aws rds restore-db-instance-from-db-snapshot \
    --db-instance-identifier sapliy-restore-test \
    --db-snapshot-identifier sapliy-prod-backup
  ```

- [ ] **Backup storage secured**
  - [ ] Encryption enabled
  - [ ] Access restricted
  - [ ] MFA required for restore

### Disaster Recovery

- [ ] **RTO defined** (Recovery Time Objective)
  - SaaS: <1 hour
  - Enterprise: <15 minutes

- [ ] **RPO defined** (Recovery Point Objective)
  - SaaS: <1 hour data loss acceptable
  - Enterprise: <5 minutes

- [ ] **Failover tested**
  - [ ] Database failover tested
  - [ ] Automatic failover working
  - [ ] Load balancer failover tested

- [ ] **Runbook for recovery documented**

---

## Security Validation (Week 3)

### SSL/TLS Certificates

- [ ] **Valid SSL certificate obtained**
  ```bash
  # Let's Encrypt via Certbot
  certbot certonly --standalone -d api.sapliy.io
  ```

- [ ] **Certificate auto-renewal configured**
  ```bash
  # Cron job for renewal
  0 0 1 * * certbot renew --quiet
  ```

- [ ] **TLS 1.3 minimum**
  ```nginx
  ssl_protocols TLSv1.3;
  ssl_ciphers HIGH:!aNULL:!MD5;
  ```

### API Security

- [ ] **API key validation working**
  ```typescript
  app.use(validateApiKey);
  ```

- [ ] **Rate limiting active**
  ```typescript
  app.use(rateLimit({
    windowMs: 15 * 60 * 1000,
    max: 100,
  }));
  ```

- [ ] **CORS configured**
  ```javascript
  app.use(cors({
    origin: process.env.ALLOWED_ORIGINS.split(','),
    credentials: true,
  }));
  ```

- [ ] **CSRF protection enabled** (if web UI)
- [ ] **Input validation on all endpoints**
- [ ] **Output encoding enabled**

### Infrastructure Security

- [ ] **Firewall rules reviewed**
  - [ ] Only necessary ports open (80, 443)
  - [ ] Database not exposed publicly
  - [ ] SSH only from specific IPs

- [ ] **Security groups configured**
  ```bash
  # Example: Allow only from load balancer
  aws ec2 authorize-security-group-ingress \
    --group-id sg-xxxxx \
    --protocol tcp \
    --port 8080 \
    --source-security-group-id sg-load-balancer
  ```

- [ ] **DDoS protection enabled** (AWS Shield, Cloudflare)
- [ ] **Web Application Firewall (WAF) configured** (AWS WAF)

---

## Load Testing & Performance (Week 3-4)

### Load Testing

- [ ] **Load tests run successfully**
  ```bash
  k6 run tests/load-tests/events-load-test.js
  # Expected: 10K events/sec, <100ms p99
  ```

- [ ] **Performance benchmarks met**
  - [ ] API response: <100ms p95
  - [ ] Event processing: <50ms p95
  - [ ] Webhook delivery: >99% success

- [ ] **Scalability verified**
  - [ ] Auto-scaling policies tested
  - [ ] Load balancer tested
  - [ ] Database connection pooling verified

### Production Readiness Test

- [ ] **Full system test in production-like environment**
  ```bash
  npm run test:production-like
  ```

- [ ] **Canary deployment tested** (if available)
- [ ] **Blue/green deployment tested** (if available)

---

## Pre-Launch Final Checks (Day Before)

### 24-Hour Checklist

- [ ] **Final code review completed**
- [ ] **All PR comments addressed**
- [ ] **Release notes prepared**
- [ ] **Communication plan ready**
  - [ ] Status page updated
  - [ ] Slack notifications configured
  - [ ] Email templates prepared

### Deployment Team Readiness

- [ ] **Deployment runbook reviewed by team**
- [ ] **Rollback plan reviewed**
- [ ] **On-call rotation confirmed**
- [ ] **Incident response team briefed**
- [ ] **Stakeholder notification plan ready**

### Infrastructure Verification

- [ ] **All services accessible**
  ```bash
  curl https://api.sapliy.io/health
  curl https://app.sapliy.io
  ```

- [ ] **DNS records correct**
  ```bash
  nslookup api.sapliy.io
  # Should resolve to correct IP
  ```

- [ ] **SSL certificate valid**
  ```bash
  openssl s_client -connect api.sapliy.io:443
  # Check cert expiry and validity
  ```

- [ ] **Monitoring dashboards accessible**
- [ ] **Alert channels tested**

---

## Launch Day (Day 0)

### Pre-Launch (6 Hours Before)

- [ ] **Status page set to "Scheduled Maintenance"**
- [ ] **Team in communication channel**
- [ ] **All monitoring dashboards open**
- [ ] **Runbook accessible**
- [ ] **Rollback plan ready**

### Deployment

```bash
#!/bin/bash
# Deploy to production

set -e

echo "🚀 Starting Sapliy production deployment..."

# 1. Tag release
git tag -a v1.0.0 -m "Production release"
git push origin v1.0.0

# 2. Build & push images
docker build -t sapliy:1.0.0 -t sapliy:latest .
docker push sapliy:1.0.0
docker push sapliy:latest

# 3. Deploy to Kubernetes / Docker Swarm
kubectl apply -f k8s/sapliy-prod.yaml

# 4. Verify deployment
kubectl rollout status deployment/sapliy-api -n production

# 5. Run smoke tests
npm run smoke-tests

# 6. Update status page
curl https://status.sapliy.io/api/v1/incidents \
  -X POST \
  -d '{"status":"investigating"}'

echo "✅ Deployment complete"
```

### Post-Deployment Verification

- [ ] **API responding** 
  ```bash
  curl https://api.sapliy.io/health
  # Expected: {"status":"ok"}
  ```

- [ ] **Databases accessible**
  ```bash
  npm run db:status
  # Expected: all green
  ```

- [ ] **Event processing working**
  ```bash
  # Emit test event
  curl -X POST https://api.sapliy.io/events \
    -H "Authorization: Bearer sk_test_123" \
    -d '{"eventType":"test.event","data":{}}'
  ```

- [ ] **Webhooks delivering**
- [ ] **Logging working**
- [ ] **Monitoring data flowing**
- [ ] **Alerts firing normally**

### Communication

- [ ] **Status page updated to "Operational"**
- [ ] **Launch announcement sent**
  - [ ] Blog post
  - [ ] Twitter/social media
  - [ ] Email to users
  - [ ] Slack announcement

- [ ] **Team celebration** 🎉

---

## Post-Launch (Week 1)

### Ongoing Monitoring

- [ ] **24/7 monitoring for 7 days**
  - [ ] Check dashboards hourly
  - [ ] Review error logs
  - [ ] Monitor performance metrics

- [ ] **Daily standup** (first week)
  - [ ] No critical issues
  - [ ] Performance metrics nominal
  - [ ] User feedback positive

### Issue Response

- [ ] **Bug fix process tested**
  - [ ] Can hotfix within 30 minutes
  - [ ] Rollback capability verified

- [ ] **Incident response procedures tested**
  - [ ] Team knows escalation path
  - [ ] Communication templates ready

### User Feedback Collection

- [ ] **Set up feedback collection**
  - [ ] In-app survey
  - [ ] Email to beta users
  - [ ] Community forum monitoring

- [ ] **First week metrics review**
  - [ ] Sign-up rate
  - [ ] Activation rate
  - [ ] Error rate
  - [ ] Performance metrics

---

## Success Metrics (First 30 Days)

✅ **Zero critical production issues**  
✅ **99.95% uptime** (SaaS target)  
✅ **<100ms API response time** (p95)  
✅ **10K+ events processed successfully**  
✅ **Zero data loss incidents**  
✅ **Positive user feedback**  

---

## Rollback Plan (If Needed)

```bash
#!/bin/bash
# Emergency rollback

set -e

echo "🚨 Rolling back to previous version..."

# 1. Revert database
npm run migrate:rollback

# 2. Deploy previous container
docker pull sapliy:1.0.0-prev
kubectl set image deployment/sapliy-api \
  sapliy=sapliy:1.0.0-prev

# 3. Verify
kubectl rollout status deployment/sapliy-api

# 4. Verify services
curl https://api.sapliy.io/health

echo "✅ Rollback complete"
```

---

## Post-Mortem (If Issues Occurred)

After deployment, if issues occurred:

1. **Timeline**: What happened and when
2. **Impact**: How many users affected
3. **Root cause**: Why it happened
4. **Resolution**: How we fixed it
5. **Prevention**: How to prevent next time
6. **Action items**: Owner + deadline

---

## Sign-Off

- [ ] **CTO approved**:  ________________  Date: ________
- [ ] **Ops lead approved**:  ________________  Date: ________
- [ ] **Security approved**:  ________________  Date: ________
- [ ] **Product approved**:  ________________  Date: ________

---

**Go-live date**: ________________  
**Deployed by**: ________________  
**Deployment time**: ________ minutes  
**Issues encountered**: ☐ None ☐ Minor ☐ Major  
**Status**: ☐ Success ☐ Rollback  
