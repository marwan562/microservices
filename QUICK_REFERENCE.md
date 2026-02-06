# Sapliy Quick Reference Guide

> Fast reference for common development and operational tasks

---

## Development Quick Start

### 1. Local Setup (Using sapliy-cli)

```bash
# Install CLI
npm install -g @sapliyio/sapliy-cli

# Start everything with one command
sapliy dev

# OR start separately
sapliy run       # Backend + services in one terminal
sapliy frontend  # Frontend in another terminal

# Verify everything is running
sapliy health
```

### 1b. Manual Setup (Without CLI)

```bash
# Clone and setup
git clone https://github.com/sapliy/fintech-ecosystem.git
cd fintech-ecosystem
docker-compose up -d

# Verify services
docker-compose ps
curl http://localhost:8080/health
```

### 2. Create First Zone

```bash
# Login
sapliy login --endpoint=http://localhost:8080

# Create zone
sapliy zones create --name="my-app" --mode=test

# Copy API keys
export SAPLIY_SK=sk_test_xxxxx
export SAPLIY_PK=pk_test_xxxxx
```

### 3. Emit Events

```javascript
// Node.js
const Sapliy = require("@sapliyio/fintech");

const sapliy = new Sapliy({ secretKey: process.env.SAPLIY_SK });

await sapliy.emit("checkout.completed", {
  orderId: "12345",
  amount: 99.99,
  currency: "USD",
});
```

```python
# Python
from sapliyio_fintech import Sapliy

sapliy = Sapliy(secret_key=os.environ['SAPLIY_SK'])

sapliy.emit('checkout.completed', {
    'orderId': '12345',
    'amount': 99.99,
    'currency': 'USD'
})
```

```go
// Go
package main

import "github.com/sapliy/fintech-sdk-go"

func main() {
  client := fintech.NewClient(os.Getenv("SAPLIY_SK"))

  client.Emit(context.Background(), "checkout.completed", map[string]interface{}{
    "orderId": "12345",
    "amount": 99.99,
    "currency": "USD",
  })
}
```

### 4. Build a Flow

Using Flow Builder UI:

1. Go to `http://localhost:3000`
2. Create new flow
3. Trigger: Select `checkout.completed`
4. Action: Send webhook to your endpoint
5. Deploy to test mode

### 5. Verify Webhook Signature

```javascript
// Node.js
const isValid = sapliy.verifyWebhook(
  req.body,
  req.headers["x-sapliy-signature"],
);

if (!isValid) {
  return res.status(401).send("Invalid signature");
}
```

---

## Testing Flows

### Using sapliy-cli

#### Quick Commands

```bash
# Emit a test event
sapliy events emit "checkout.completed" '{"orderId":"12345","amount":99.99}'

# Listen for events in real-time
sapliy events listen "checkout.*"

# View flow executions
sapliy flows list
sapliy flows logs --flow=send-confirmation-email --follow

# Test a specific flow
sapliy test --flow=send-confirmation-email

# Replay an event
sapliy events replay --after="2024-01-15T10:00:00Z"
```

#### Webhook Testing

```bash
# Start webhook inspector (see incoming webhooks)
sapliy webhooks listen

# Test webhook delivery to URL
sapliy webhooks test --url="https://example.com/webhook"

# Replay failed webhook
sapliy webhooks replay --id="evt_abc123"
```

#### Complete Dev Workflow

```bash
# Terminal 1: Start backend + frontend
sapliy dev

# Terminal 2: Watch logs
sapliy logs --follow

# Terminal 3: Listen to events
sapliy events listen "*"

# Terminal 4: Emit test events
sapliy events emit "checkout.completed" '{"orderId":"test"}'
```

### Using Testing Toolkit

```python
from sapliyio_fintech.testing import TestClient

client = TestClient(endpoint='http://localhost:8080')

# Emit test event
response = client.emit_event(
  zone_id='zone_test_xxx',
  event_type='checkout.completed',
  data={
    'orderId': '12345',
    'amount': 99.99,
    'currency': 'USD'
  }
)

# Assert flow executed
assert response.status_code == 200
assert len(response.flow_executions) > 0
assert response.flow_executions[0]['status'] == 'success'
```

### CI/CD Integration

```yaml
# GitHub Actions
name: Test Flows
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      sapliy:
        image: sapliyio/fintech:latest
        ports:
          - 8080:8080
    steps:
      - uses: actions/checkout@v2
      - name: Run flow tests
        run: |
          npm install -g @sapliyio/fintech-testing
          fintech-test --endpoint=http://localhost:8080
```

---

## API Reference (Common Endpoints)

### Authentication

```bash
# POST /auth/login
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}'

# Response
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {"id":"user_xxx","email":"user@example.com"}
}
```

### Zones

```bash
# GET /zones
curl http://localhost:8080/zones \
  -H "Authorization: Bearer $TOKEN"

# POST /zones
curl -X POST http://localhost:8080/zones \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"my-app","mode":"test"}'

# Response
{
  "id": "zone_xxx",
  "name": "my-app",
  "mode": "test",
  "secret_key": "sk_test_xxx",
  "publishable_key": "pk_test_xxx"
}
```

### Events

```bash
# POST /events
curl -X POST http://localhost:8080/events \
  -H "Authorization: Bearer sk_test_xxx" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "checkout.completed",
    "zone_id": "zone_xxx",
    "data": {"orderId":"12345","amount":99.99}
  }'

# Response
{
  "id": "evt_xxx",
  "event_type": "checkout.completed",
  "zone_id": "zone_xxx",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### Flows

```bash
# GET /flows
curl http://localhost:8080/flows?zone_id=zone_xxx \
  -H "Authorization: Bearer $TOKEN"

# POST /flows
curl -X POST http://localhost:8080/flows \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "send-email",
    "zone_id": "zone_xxx",
    "trigger_event": "checkout.completed",
    "definition": {
      "actions": [{
        "type": "webhook",
        "url": "https://example.com/webhook"
      }]
    }
  }'

# PATCH /flows/{flowId}
curl -X PATCH http://localhost:8080/flows/flow_xxx \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"enabled":true}'
```

---

## Environment Variables

### Development

```bash
SAPLIY_ENDPOINT=http://localhost:8080
SAPLIY_SECRET_KEY=sk_test_xxxxx
SAPLIY_PUBLISHABLE_KEY=pk_test_xxxxx
LOG_LEVEL=debug
DATABASE_URL=postgresql://postgres:password@localhost:5432/sapliy
REDIS_URL=redis://localhost:6379
KAFKA_BROKERS=localhost:9092
```

### Production (SaaS)

```bash
SAPLIY_ENDPOINT=https://api.sapliy.io
SAPLIY_SECRET_KEY=sk_live_xxxxx
SAPLIY_PUBLISHABLE_KEY=pk_live_xxxxx
LOG_LEVEL=info
```

### Production (Self-Hosted)

```bash
SAPLIY_ENDPOINT=https://fintech.yourcompany.com
SAPLIY_SECRET_KEY=sk_live_xxxxx
SAPLIY_SECRET_ENCRYPTION_KEY=<base64-encoded-key>
DATABASE_URL=postgresql://user:pass@prod-db:5432/sapliy_prod
REDIS_URL=redis://prod-redis:6379
KAFKA_BROKERS=kafka1:9092,kafka2:9092,kafka3:9092
VAULT_ADDR=https://vault.yourcompany.com
VAULT_TOKEN=<token>
```

---

## Common Troubleshooting

### Event Not Processing?

```bash
# 1. Check zone exists
sapliy zones list

# 2. Emit test event
sapliy flows trigger --zone=test --flow=<flowId> --data='{}'

# 3. Check flow logs
kubectl logs -n sapliy-prod -l app=flow-engine --tail=50

# 4. Check database
psql -c "SELECT * FROM flow_executions ORDER BY created_at DESC LIMIT 10;"
```

### Webhook Not Firing?

```bash
# 1. Check webhook queue
psql -c "SELECT * FROM webhook_queue WHERE status = 'failed';"

# 2. Verify endpoint is reachable
curl -v https://your-endpoint.com/webhook

# 3. Check retry attempts
psql -c "SELECT event_id, retry_count FROM webhook_queue WHERE status = 'failed';"

# 4. Manually retry
psql -c "UPDATE webhook_queue SET status = 'pending', retry_count = 0 WHERE status = 'failed';"
```

### Performance Issues?

```bash
# 1. Check database connections
psql -c "SELECT count(*) FROM pg_stat_activity;"

# 2. Check slow queries
psql -c "SELECT query, mean_time FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;"

# 3. Check Kafka lag
kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --group sapliy-flow-executor --describe

# 4. Scale services
kubectl scale deployment flow-engine -n sapliy-prod --replicas=15
```

---

## Useful Links

- 📚 [Full Documentation](https://docs.sapliy.io)
- 🐙 [GitHub Organization](https://github.com/sapliy)
- 💬 [Discord Community](https://discord.gg/sapliy)
- 🐛 [Report Issues](https://github.com/sapliy/fintech-ecosystem/issues)
- 🔒 [Security](https://security.sapliy.io)
- 📧 [Enterprise Support](mailto:contact@sapliy.io)
