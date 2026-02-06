# 🏆 Sapliy Professional Quality Standards

> Enterprise-grade quality standards ensuring production readiness

---

## Code Quality Standards

### 1. TypeScript Strictness

```typescript
// tsconfig.json - Strict mode enabled
{
  "compilerOptions": {
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true,
    "strictFunctionTypes": true,
    "strictPropertyInitialization": true,
    "noImplicitThis": true,
    "alwaysStrict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitOverride": true
  }
}
```

### 2. ESLint Configuration

```javascript
// .eslintrc.js
module.exports = {
  root: true,
  parser: "@typescript-eslint/parser",
  extends: [
    "eslint:recommended",
    "plugin:@typescript-eslint/recommended",
    "plugin:prettier/recommended",
  ],
  rules: {
    "no-console": ["warn", { allow: ["warn", "error"] }],
    "no-var": "error",
    "prefer-const": "error",
    eqeqeq: ["error", "always"],
    "@typescript-eslint/explicit-function-return-types": "error",
    "@typescript-eslint/no-explicit-any": "error",
    "no-unused-vars": "off",
    "@typescript-eslint/no-unused-vars": [
      "error",
      {
        argsIgnorePattern: "^_",
      },
    ],
  },
};
```

### 3. Code Formatting (Prettier)

```json
// .prettierrc
{
  "semi": true,
  "trailingComma": "es5",
  "singleQuote": true,
  "printWidth": 100,
  "tabWidth": 2,
  "useTabs": false,
  "bracketSpacing": true,
  "arrowParens": "always"
}
```

### 4. Code Complexity Rules

```bash
# Maximum cyclomatic complexity: 10
# Maximum lines per function: 50
# Maximum parameters: 5
# Nesting depth: 3 levels max

# Check with complexity checker
npm install -g complexity-report
# or
npm run check-complexity
```

### 5. Git Pre-Commit Hooks

```bash
# .husky/pre-commit
#!/bin/sh
. "$(dirname "$0")/_/husky.sh"

echo "🔍 Linting..."
npm run lint --fix

echo "🔍 Type checking..."
npm run type-check

echo "✅ Pre-commit checks passed"
```

---

## Testing Standards

### 1. Test Coverage Requirements

```
Minimum coverage requirements:
├── Overall coverage: 80%
├── Branches: 75%
├── Functions: 80%
├── Lines: 80%
├── Statements: 80%
└── Critical paths: 100%
```

### 2. Test File Organization

```
src/
├── services/
│   ├── auth.service.ts
│   └── __tests__/
│       ├── auth.service.unit.test.ts
│       ├── auth.service.integration.test.ts
│       └── fixtures/
├── repositories/
│   ├── user.repository.ts
│   └── __tests__/
│       ├── user.repository.test.ts
│       └── mocks/
└── ...
```

### 3. Test Naming Convention

```typescript
describe("AuthService", () => {
  describe("login", () => {
    // ✅ Good: Clear, follows Given-When-Then pattern
    it("should return token when valid credentials provided", () => {});
    it("should throw error when user not found", () => {});
    it("should enforce password minimum length", () => {});

    // ❌ Bad: Unclear, too generic
    it("works", () => {});
    it("handles login", () => {});
  });
});
```

### 4. Mock & Stub Standards

```typescript
// ✅ Good: Using jest.fn() with mocks
const mockRepository = {
  findByEmail: jest.fn().mockResolvedValue(user),
  create: jest.fn().mockResolvedValue(newUser),
};

// ✅ Good: Clear mock reset between tests
beforeEach(() => {
  jest.clearAllMocks();
  jest.resetAllMocks();
});

// ❌ Bad: Global state pollution
let globalMockUser;
```

### 5. Assertion Standards

```typescript
// ✅ Good: Specific, descriptive assertions
expect(response.status).toBe(200);
expect(response.body).toHaveProperty("token");
expect(response.body.token).toMatch(/^tk_/);

// ❌ Bad: Generic, hard to debug failures
expect(response).toBeTruthy();
expect(data).toEqual(expect.anything());
```

---

## Security Standards

### 1. Secret Management

```typescript
// ✅ Good: Use environment variables
const dbPassword = process.env.DB_PASSWORD;
if (!dbPassword) {
  throw new Error("DB_PASSWORD not set");
}

// ❌ Bad: Hardcoded secrets
const dbPassword = "mySecurePassword123";
```

### 2. Input Validation

```typescript
// ✅ Good: Validate all inputs
import { z } from "zod";

const loginSchema = z.object({
  email: z.string().email(),
  password: z.string().min(8).regex(/[A-Z]/),
});

const login = (data: unknown) => {
  const validated = loginSchema.parse(data);
  // Use validated data
};

// ❌ Bad: Trust user input
const login = (email: string, password: string) => {
  return db.query(`SELECT * FROM users WHERE email = '${email}'`);
};
```

### 3. Error Handling

```typescript
// ✅ Good: Don't expose internal details
catch (error) {
  logger.error('Payment processing failed', { orderId, error });
  res.status(500).json({ error: 'Payment processing failed' });
}

// ❌ Bad: Leak sensitive information
catch (error) {
  res.status(500).json({
    error: error.message, // Exposes database details
    stack: error.stack,    // Exposes source code paths
  });
}
```

### 4. Dependency Security

```bash
# Check for vulnerabilities regularly
npm audit
npm audit fix

# Use lockfiles
npm ci  # Instead of npm install

# Update dependencies securely
npm outdated
npm update

# Use Snyk for continuous monitoring
snyk test
snyk monitor
```

### 5. Authentication & Authorization

```typescript
// ✅ Good: Implement proper RBAC
class UserController {
  async updateUser(req: Request, res: Response) {
    // Verify user owns the resource
    const user = await auth.getUser(req);
    const targetUser = await db.getUser(req.params.id);

    if (user.id !== targetUser.id && !user.isAdmin) {
      throw new ForbiddenError('Unauthorized');
    }

    // Update user
  }
}

// ❌ Bad: No permission checks
async updateUser(req: Request, res: Response) {
  await db.updateUser(req.params.id, req.body);
}
```

---

## API Standards

### 1. RESTful Design

```typescript
// ✅ Good: Proper HTTP methods and status codes
GET    /api/v1/zones              // 200 OK - list resources
POST   /api/v1/zones              // 201 Created - create resource
GET    /api/v1/zones/:id          // 200 OK - get single resource
PUT    /api/v1/zones/:id          // 200 OK - update resource
DELETE /api/v1/zones/:id          // 204 No Content - delete resource

// ✅ Good: Proper error codes
400 Bad Request    - Invalid input
401 Unauthorized   - Missing authentication
403 Forbidden      - Lacks permission
404 Not Found      - Resource doesn't exist
409 Conflict       - Resource already exists
429 Too Many Req   - Rate limited
500 Server Error   - Internal error
```

### 2. Request/Response Format

```typescript
// ✅ Good: Consistent, documented response format
{
  "success": true,
  "data": {
    "id": "zone_123",
    "name": "My Zone",
    "createdAt": "2024-01-15T10:30:00Z"
  },
  "pagination": {
    "page": 1,
    "pageSize": 10,
    "total": 100
  }
}

// Error response
{
  "success": false,
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Email format is invalid",
    "details": {
      "field": "email",
      "received": "invalid-email"
    }
  }
}
```

### 3. API Versioning

```typescript
// ✅ Good: Version in URL path
GET / api / v1 / zones;
GET / api / v2 / zones; // Different implementation if needed

// Accept both with deprecation warning
// But prefer URL versioning over Accept headers
```

### 4. API Documentation (OpenAPI/Swagger)

```typescript
/**
 * Create a new zone
 *
 * @openapi
 * /zones:
 *   post:
 *     summary: Create a new zone
 *     requestBody:
 *       required: true
 *       content:
 *         application/json:
 *           schema:
 *             type: object
 *             properties:
 *               name:
 *                 type: string
 *               mode:
 *                 type: string
 *                 enum: [test, live]
 *     responses:
 *       201:
 *         description: Zone created successfully
 *       400:
 *         description: Invalid request
 */
app.post("/zones", createZoneController);
```

---

## Database Standards

### 1. Query Performance

```typescript
// ✅ Good: Use indexes, limit results
const users = await db
  .query('SELECT * FROM users WHERE status = $1 LIMIT $2', ['active', 100])
  .explain(); // Check query plan

// Create indexes
CREATE INDEX idx_users_status ON users(status);

// ❌ Bad: N+1 queries
users.forEach(async (user) => {
  user.orders = await db.query('SELECT * FROM orders WHERE user_id = $1', [user.id]);
});
```

### 2. Database Transactions

```typescript
// ✅ Good: ACID transactions
async function transferMoney(fromId: string, toId: string, amount: number) {
  const tx = await db.transaction();

  try {
    await tx.query("UPDATE accounts SET balance = balance - $1 WHERE id = $2", [
      amount,
      fromId,
    ]);
    await tx.query("UPDATE accounts SET balance = balance + $1 WHERE id = $2", [
      amount,
      toId,
    ]);
    await tx.commit();
  } catch (error) {
    await tx.rollback();
    throw error;
  }
}

// ✅ Good: Using ORM transaction helpers
await db.transaction(async (trx) => {
  await trx("accounts").where("id", fromId).decrement("balance", amount);
  await trx("accounts").where("id", toId).increment("balance", amount);
});
```

### 3. Data Encryption

```typescript
// ✅ Good: Encrypt sensitive data
const crypto = require('crypto');

function encryptField(plaintext: string, key: string): string {
  const iv = crypto.randomBytes(16);
  const cipher = crypto.createCipheriv('aes-256-cbc', Buffer.from(key), iv);
  const encrypted = Buffer.concat([cipher.update(plaintext), cipher.final()]);
  return iv.toString('hex') + ':' + encrypted.toString('hex');
}

// ✅ Good: Use database-level encryption
// PostgreSQL: PGP_SYM_ENCRYPT, pgcrypto extension
SELECT pgp_sym_encrypt('sensitive_data', 'encryption_key');
```

### 4. Backup Strategy

```bash
# ✅ Good: Regular automated backups
# Daily backup at 2 AM UTC
0 2 * * * pg_dump $DATABASE_URL > /backups/sapliy-$(date +%Y%m%d).sql

# Test restore monthly
0 0 1 * * pg_restore --create /backups/sapliy-latest.sql -U test_user
```

---

## Logging Standards

### 1. Log Levels

```typescript
// ✅ Good: Appropriate log levels
logger.debug("User login attempt", { userId, timestamp }); // Dev only
logger.info("User logged in successfully", { userId }); // Normal operations
logger.warn("Slow query detected", { duration: "5s" }); // Potential issue
logger.error("Payment processing failed", { error, orderId }); // Error
logger.fatal("Database connection lost"); // Critical
```

### 2. Structured Logging

```typescript
// ✅ Good: Structured, searchable logs
logger.info("Event processed", {
  eventId: "evt_123",
  eventType: "payment.completed",
  duration: 250,
  status: "success",
  userId: "usr_456",
  timestamp: new Date().toISOString(),
});

// ❌ Bad: Unstructured, hard to search
console.log(`Event evt_123 processed in 250ms`);
```

### 3. Sensitive Data Masking

```typescript
// ✅ Good: Mask sensitive data in logs
function maskSensitiveData(data: Record<string, any>) {
  return {
    ...data,
    password: "***",
    ssn: data.ssn?.slice(-4).padStart(data.ssn.length, "*"),
    creditCard: data.creditCard?.slice(-4).padStart(16, "*"),
  };
}

logger.info("User created", maskSensitiveData(userData));
```

---

## Documentation Standards

### 1. Code Comments

```typescript
// ✅ Good: Explain WHY, not WHAT
// Use database transaction to ensure consistency
// if payment fails after debit, rollback prevents data corruption
async function processPayment(orderId: string) {
  const tx = await db.transaction();
  // ...
}

// ❌ Bad: Obvious, redundant comments
// Get user by ID
const user = await db.getUser(id);

// ✅ Good: Comment complex logic
// Exponential backoff: 100ms, 200ms, 400ms, 800ms...
// Prevents thundering herd when service recovers
const delay = baseDelay * Math.pow(2, attemptNumber);
```

### 2. Function Documentation

```typescript
/**
 * Process a payment event and execute associated flows
 *
 * This function is the main entry point for payment processing.
 * It validates the payment, executes all matching flows, and
 * updates the ledger atomically.
 *
 * @param eventId - The ID of the payment event
 * @param data - Payment event data
 * @returns Promise<ProcessedEvent> The processed event result
 * @throws PaymentValidationError if payment data is invalid
 * @throws FlowExecutionError if any flow fails
 *
 * @example
 * const result = await processPaymentEvent('evt_123', {
 *   orderId: 'ord_456',
 *   amount: 99.99,
 * });
 *
 * @see {@link FlowEngine.execute} for flow execution details
 */
async function processPaymentEvent(
  eventId: string,
  data: PaymentEventData,
): Promise<ProcessedEvent> {
  // Implementation
}
```

### 3. README Standards

Every module should have a README:

````markdown
# Event Service

Description of what this module does.

## Installation

```bash
npm install @sapliyio/event-service
```
````

## Quick Start

```typescript
import { EventService } from "@sapliyio/event-service";

const service = new EventService(config);
await service.emit("event.type", data);
```

## API Reference

### `emit(eventType: string, data: any): Promise<Event>`

Emit an event...

### Error Handling

Possible errors thrown...

## Examples

Real-world examples...

## Contributing

How to contribute to this module...

```

---

## Performance Standards

### 1. Response Time Targets

```

API Endpoints:

- GET requests: <100ms p95
- POST requests: <200ms p95
- Heavy queries: <500ms p95

Event Processing:

- Event emit: <50ms p95
- Flow execution: <500ms p95
- Webhook delivery: <5s timeout

Database:

- Queries: <50ms p95
- Transactions: <100ms p95

````

### 2. Caching Strategy

```typescript
// ✅ Good: Cache appropriate data
const user = await cache.get(`user:${userId}`, () =>
  db.getUser(userId),
  { ttl: 3600 } // 1 hour cache
);

// ✅ Good: Invalidate on updates
async function updateUser(id: string, data: any) {
  await db.updateUser(id, data);
  await cache.delete(`user:${id}`); // Invalidate
}
````

### 3. Database Connection Pooling

```javascript
// ✅ Good: Configure pool limits
const pool = new Pool({
  max: 20, // Max connections
  idleTimeoutMillis: 30000,
  connectionTimeoutMillis: 2000,
});
```

---

## Monitoring & Observability Standards

### 1. Metrics to Track

```
Application Metrics:
- Request rate (req/sec)
- Request duration (p50, p95, p99)
- Error rate (%)
- Error types (500, 403, etc)

Business Metrics:
- Events processed (per second)
- Webhooks delivered (success rate)
- Flows executed (per hour)

Infrastructure Metrics:
- CPU usage (%)
- Memory usage (%)
- Disk usage (%)
- Network I/O (MB/s)
- Database connections (active)
```

### 2. Alerting Rules

```yaml
# ✅ Good: Actionable alerts
- name: HighErrorRate
  condition: error_rate > 5%
  duration: 5m
  action: Page on-call engineer

- name: SlowQueries
  condition: query_duration_p95 > 500ms
  duration: 10m
  action: Send to team Slack channel
```

---

## Deployment Standards

### 1. Version Control

```bash
# ✅ Good: Semantic versioning
v1.0.0  # MAJOR.MINOR.PATCH
v1.1.0  # Feature release
v1.0.1  # Bug fix

# Good commit messages
✅ "feat: add event replay functionality"
✅ "fix: prevent webhook signature verification bypass"
✅ "docs: update API documentation"
❌ "update"
❌ "fix bug"
```

### 2. Release Process

```bash
# Create release branch
git checkout -b release/v1.1.0

# Update version in package.json
npm version minor

# Create release notes
# Update CHANGELOG.md

# Tag release
git tag -a v1.1.0 -m "Release v1.1.0"

# Merge to main
git checkout main
git merge release/v1.1.0

# Publish
npm publish
git push origin main --tags
```

---

## Conclusion

These standards ensure:
✅ Code quality and maintainability  
✅ Security and reliability  
✅ Performance and scalability  
✅ Ease of debugging and monitoring  
✅ Team consistency and collaboration

**All code must pass these standards before being merged to production.**
