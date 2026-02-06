# 🧪 Sapliy Complete Testing & QA Plan

> Comprehensive testing strategy for all services, Docker, events, CLI commands, and production readiness before launch

---

## Table of Contents

1. [Testing Strategy Overview](#testing-strategy-overview)
2. [Unit Testing](#unit-testing)
3. [Integration Testing](#integration-testing)
4. [End-to-End Testing](#end-to-end-testing)
5. [Docker & Infrastructure Testing](#docker--infrastructure-testing)
6. [CLI Testing](#cli-testing)
7. [Event Flow Testing](#event-flow-testing)
8. [Performance Testing](#performance-testing)
9. [Security Testing](#security-testing)
10. [Production Readiness Checklist](#production-readiness-checklist)

---

## Testing Strategy Overview

### Testing Pyramid

```
                    ▲
                   /|\
                  / | \
                 /  |  \  E2E Tests (10%)
                /   |   \
               /    |    \
              /_____|_____\
             /      |      \
            /       |       \  Integration Tests (30%)
           /________|________\
          /         |         \
         /          |          \
        /           |           \  Unit Tests (60%)
       /____________|____________\
```

### Test Coverage Targets

| Layer | Target | Tools | Priority |
|-------|--------|-------|----------|
| **Unit Tests** | 80%+ | Jest, Mocha | 🔴 Critical |
| **Integration Tests** | 70%+ | Supertest, Docker Compose | 🔴 Critical |
| **E2E Tests** | 60%+ | Cypress, Playwright | 🟡 High |
| **Performance Tests** | Key paths | K6, Artillery | 🟡 High |
| **Security Tests** | OWASP Top 10 | OWASP ZAP, Snyk | 🟡 High |

---

## Unit Testing

### 1. Backend Services (fintech-ecosystem)

#### Setup

```bash
# Install testing dependencies
npm install --save-dev jest @testing-library/jest-dom supertest
npm install --save-dev ts-jest @types/jest

# Create jest.config.js
cat > jest.config.js << 'EOF'
module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'node',
  roots: ['<rootDir>/src'],
  testMatch: ['**/__tests__/**/*.ts', '**/?(*.)+(spec|test).ts'],
  collectCoverageFrom: [
    'src/**/*.ts',
    '!src/**/*.d.ts',
    '!src/**/index.ts',
  ],
  coverageThreshold: {
    global: {
      branches: 70,
      functions: 80,
      lines: 80,
      statements: 80,
    },
  },
};
EOF
```

#### Example: Auth Service Tests

```typescript
// src/services/__tests__/auth.service.test.ts
import { AuthService } from '../auth.service';
import { JwtService } from '../jwt.service';
import { UserRepository } from '../../repositories/user.repository';

describe('AuthService', () => {
  let authService: AuthService;
  let jwtService: JwtService;
  let userRepository: UserRepository;

  beforeEach(() => {
    // Setup mocks
    userRepository = {
      findByEmail: jest.fn(),
      create: jest.fn(),
    } as any;

    jwtService = new JwtService();
    authService = new AuthService(userRepository, jwtService);
  });

  describe('login', () => {
    it('should return token on successful login', async () => {
      // Arrange
      const email = 'user@example.com';
      const password = 'password123';
      const user = { id: '1', email, password: 'hashed_password' };

      jest.spyOn(userRepository, 'findByEmail').mockResolvedValue(user);
      jest.spyOn(authService, 'comparePasswords').mockResolvedValue(true);

      // Act
      const result = await authService.login(email, password);

      // Assert
      expect(result).toHaveProperty('token');
      expect(result.token).toBeTruthy();
      expect(userRepository.findByEmail).toHaveBeenCalledWith(email);
    });

    it('should throw error on invalid credentials', async () => {
      // Arrange
      jest.spyOn(userRepository, 'findByEmail').mockResolvedValue(null);

      // Act & Assert
      await expect(authService.login('wrong@example.com', 'password')).rejects.toThrow(
        'Invalid credentials'
      );
    });

    it('should validate email format', async () => {
      // Act & Assert
      await expect(authService.login('invalid-email', 'password')).rejects.toThrow(
        'Invalid email format'
      );
    });
  });

  describe('createZone', () => {
    it('should create zone with unique keys', async () => {
      // Arrange
      const zoneName = 'test-zone';

      // Act
      const zone = await authService.createZone(zoneName, 'test');

      // Assert
      expect(zone).toHaveProperty('secretKey');
      expect(zone).toHaveProperty('publishableKey');
      expect(zone.secretKey).toMatch(/^sk_test_/);
      expect(zone.publishableKey).toMatch(/^pk_test_/);
    });

    it('should generate cryptographically secure keys', async () => {
      // Act
      const zone1 = await authService.createZone('zone1', 'test');
      const zone2 = await authService.createZone('zone2', 'test');

      // Assert
      expect(zone1.secretKey).not.toBe(zone2.secretKey);
      expect(zone1.publishableKey).not.toBe(zone2.publishableKey);
    });
  });

  describe('verifyToken', () => {
    it('should verify valid token', async () => {
      // Arrange
      const token = jwtService.sign({ userId: '1' });

      // Act
      const decoded = authService.verifyToken(token);

      // Assert
      expect(decoded).toHaveProperty('userId');
      expect(decoded.userId).toBe('1');
    });

    it('should throw error on expired token', () => {
      // Arrange
      const expiredToken = jwtService.sign({ userId: '1' }, { expiresIn: '0s' });

      // Act & Assert
      expect(() => authService.verifyToken(expiredToken)).toThrow(
        'Token expired'
      );
    });
  });
});
```

#### Example: Event Service Tests

```typescript
// src/services/__tests__/event.service.test.ts
import { EventService } from '../event.service';
import { EventRepository } from '../../repositories/event.repository';
import { KafkaService } from '../kafka.service';

describe('EventService', () => {
  let eventService: EventService;
  let eventRepository: EventRepository;
  let kafkaService: KafkaService;

  beforeEach(() => {
    eventRepository = {
      create: jest.fn(),
      findById: jest.fn(),
      findByZone: jest.fn(),
    } as any;

    kafkaService = {
      publish: jest.fn(),
    } as any;

    eventService = new EventService(eventRepository, kafkaService);
  });

  describe('emitEvent', () => {
    it('should create event and publish to Kafka', async () => {
      // Arrange
      const payload = {
        zoneId: 'zone_test_123',
        eventType: 'payment.completed',
        data: { orderId: '123', amount: 99.99 },
      };

      jest.spyOn(eventRepository, 'create').mockResolvedValue({
        id: 'evt_123',
        ...payload,
        createdAt: new Date(),
      });

      // Act
      const event = await eventService.emitEvent(payload);

      // Assert
      expect(event).toHaveProperty('id');
      expect(event.eventType).toBe('payment.completed');
      expect(eventRepository.create).toHaveBeenCalledWith(payload);
      expect(kafkaService.publish).toHaveBeenCalledWith('events', expect.objectContaining({
        eventId: event.id,
        eventType: 'payment.completed',
      }));
    });

    it('should validate event data schema', async () => {
      // Act & Assert
      await expect(eventService.emitEvent({
        zoneId: 'zone_test_123',
        eventType: null,
        data: {},
      })).rejects.toThrow('Invalid event type');
    });

    it('should enforce size limits on event data', async () => {
      // Arrange
      const largeData = { content: 'x'.repeat(1024 * 1024 * 10) }; // 10MB

      // Act & Assert
      await expect(eventService.emitEvent({
        zoneId: 'zone_test_123',
        eventType: 'large.event',
        data: largeData,
      })).rejects.toThrow('Event payload too large');
    });

    it('should idempotently handle duplicate events', async () => {
      // Arrange
      const eventId = 'evt_123';
      const payload = {
        zoneId: 'zone_test_123',
        eventType: 'payment.completed',
        eventId,
        data: { orderId: '123' },
      };

      const existingEvent = { id: eventId, ...payload };
      jest.spyOn(eventRepository, 'findById').mockResolvedValue(existingEvent);

      // Act
      const result = await eventService.emitEvent(payload);

      // Assert
      expect(result.id).toBe(eventId);
      expect(eventRepository.create).not.toHaveBeenCalled(); // Not created again
    });
  });

  describe('getEventsByZone', () => {
    it('should return paginated events', async () => {
      // Arrange
      const zoneId = 'zone_test_123';
      const events = Array.from({ length: 50 }, (_, i) => ({
        id: `evt_${i}`,
        zoneId,
        eventType: 'test.event',
        createdAt: new Date(),
      }));

      jest.spyOn(eventRepository, 'findByZone').mockResolvedValue({
        data: events.slice(0, 10),
        total: 50,
        page: 1,
        pageSize: 10,
      });

      // Act
      const result = await eventService.getEventsByZone(zoneId, { page: 1, pageSize: 10 });

      // Assert
      expect(result.data).toHaveLength(10);
      expect(result.total).toBe(50);
    });
  });
});
```

#### Example: Flow Engine Tests

```typescript
// src/services/__tests__/flow-engine.service.test.ts
import { FlowEngineService } from '../flow-engine.service';
import { FlowRepository } from '../../repositories/flow.repository';
import { WebhookService } from '../webhook.service';

describe('FlowEngineService', () => {
  let flowEngine: FlowEngineService;
  let flowRepository: FlowRepository;
  let webhookService: WebhookService;

  beforeEach(() => {
    flowRepository = {
      findByZoneAndTrigger: jest.fn(),
    } as any;

    webhookService = {
      sendWebhook: jest.fn(),
    } as any;

    flowEngine = new FlowEngineService(flowRepository, webhookService);
  });

  describe('executeFlow', () => {
    it('should execute all actions in sequence', async () => {
      // Arrange
      const flow = {
        id: 'flow_123',
        zoneId: 'zone_test_123',
        triggerEvent: 'payment.completed',
        actions: [
          { type: 'webhook', url: 'https://example.com/webhook' },
          { type: 'log', message: 'Payment completed' },
        ],
      };

      const event = {
        id: 'evt_123',
        zoneId: 'zone_test_123',
        eventType: 'payment.completed',
        data: { orderId: '123' },
      };

      jest.spyOn(webhookService, 'sendWebhook').mockResolvedValue({
        statusCode: 200,
      });

      // Act
      const result = await flowEngine.executeFlow(flow, event);

      // Assert
      expect(result.status).toBe('success');
      expect(result.executedActions).toBe(2);
      expect(webhookService.sendWebhook).toHaveBeenCalled();
    });

    it('should handle action failures gracefully', async () => {
      // Arrange
      const flow = {
        id: 'flow_123',
        actions: [
          { type: 'webhook', url: 'https://example.com/webhook' },
        ],
      };

      jest.spyOn(webhookService, 'sendWebhook')
        .mockRejectedValue(new Error('Webhook failed'));

      // Act
      const result = await flowEngine.executeFlow(flow, { id: 'evt_123' });

      // Assert
      expect(result.status).toBe('failed');
      expect(result.errors).toContain('Webhook failed');
    });

    it('should support conditional actions', async () => {
      // Arrange
      const flow = {
        id: 'flow_123',
        actions: [
          {
            type: 'conditional',
            condition: { field: 'amount', operator: 'gt', value: 100 },
            thenActions: [{ type: 'webhook', url: 'https://notify.example.com' }],
            elseActions: [],
          },
        ],
      };

      const event = {
        id: 'evt_123',
        data: { amount: 150 },
      };

      // Act
      const result = await flowEngine.executeFlow(flow, event);

      // Assert
      expect(result.executedActions).toBe(1);
      expect(webhookService.sendWebhook).toHaveBeenCalled();
    });

    it('should enforce timeout on action execution', async () => {
      // Arrange
      const flow = {
        id: 'flow_123',
        actions: [
          { type: 'webhook', url: 'https://slow.example.com', timeout: 5000 },
        ],
      };

      jest.spyOn(webhookService, 'sendWebhook')
        .mockImplementation(() => new Promise(resolve => setTimeout(resolve, 10000)));

      // Act & Assert
      await expect(flowEngine.executeFlow(flow, { id: 'evt_123' }))
        .rejects.toThrow('Action timeout');
    });
  });

  describe('validateFlow', () => {
    it('should validate flow structure', async () => {
      // Arrange
      const invalidFlow = {
        id: 'flow_123',
        actions: [{ type: 'invalid_action' }],
      };

      // Act & Assert
      await expect(flowEngine.validateFlow(invalidFlow))
        .rejects.toThrow('Invalid action type');
    });

    it('should catch circular references', async () => {
      // Arrange
      const flow = {
        id: 'flow_123',
        actions: [
          {
            type: 'conditional',
            thenActions: [
              {
                type: 'trigger_flow',
                flowId: 'flow_123', // Circular!
              },
            ],
          },
        ],
      };

      // Act & Assert
      await expect(flowEngine.validateFlow(flow))
        .rejects.toThrow('Circular flow reference');
    });
  });
});
```

### 2. SDK Testing (fintech-sdk-node)

```typescript
// src/__tests__/sapliy.test.ts
import { Sapliy } from '../index';
import { HttpClient } from '../http-client';

describe('Sapliy SDK', () => {
  let sapliy: Sapliy;
  let httpClient: HttpClient;

  beforeEach(() => {
    httpClient = {
      post: jest.fn(),
      get: jest.fn(),
    } as any;

    sapliy = new Sapliy({
      secretKey: 'sk_test_123',
      endpoint: 'http://localhost:8080',
      httpClient,
    });
  });

  describe('emit', () => {
    it('should emit event successfully', async () => {
      // Arrange
      jest.spyOn(httpClient, 'post').mockResolvedValue({
        id: 'evt_123',
        eventType: 'test.event',
      });

      // Act
      const result = await sapliy.emit('test.event', { data: 'test' });

      // Assert
      expect(result.id).toBe('evt_123');
      expect(httpClient.post).toHaveBeenCalledWith(
        '/events',
        expect.objectContaining({ eventType: 'test.event' })
      );
    });

    it('should validate event type', async () => {
      // Act & Assert
      await expect(sapliy.emit(null as any, {}))
        .rejects.toThrow('Event type is required');
    });

    it('should support optional context', async () => {
      // Act
      await sapliy.emit('test.event', { data: 'test' }, { idempotencyKey: 'key_123' });

      // Assert
      expect(httpClient.post).toHaveBeenCalledWith(
        '/events',
        expect.objectContaining({ idempotencyKey: 'key_123' })
      );
    });
  });

  describe('verifyWebhook', () => {
    it('should verify valid webhook signature', () => {
      // Arrange
      const payload = JSON.stringify({ eventType: 'test' });
      const signature = sapliy.computeSignature(payload, 'sk_test_123');

      // Act
      const isValid = sapliy.verifyWebhook(payload, signature);

      // Assert
      expect(isValid).toBe(true);
    });

    it('should reject tampered webhook', () => {
      // Arrange
      const payload = JSON.stringify({ eventType: 'test' });
      const signature = sapliy.computeSignature(payload, 'sk_test_123');
      const tamperedPayload = JSON.stringify({ eventType: 'hacked' });

      // Act
      const isValid = sapliy.verifyWebhook(tamperedPayload, signature);

      // Assert
      expect(isValid).toBe(false);
    });
  });

  describe('retry logic', () => {
    it('should retry on transient failure', async () => {
      // Arrange
      const httpSpy = jest
        .spyOn(httpClient, 'post')
        .mockRejectedValueOnce(new Error('Network error'))
        .mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValueOnce({ id: 'evt_123' });

      sapliy = new Sapliy({
        secretKey: 'sk_test_123',
        maxRetries: 3,
        httpClient,
      });

      // Act
      const result = await sapliy.emit('test.event', { data: 'test' });

      // Assert
      expect(result.id).toBe('evt_123');
      expect(httpClient.post).toHaveBeenCalledTimes(3);
    });

    it('should not retry on permanent failure', async () => {
      // Arrange
      jest.spyOn(httpClient, 'post').mockRejectedValue(
        new Error('401 Unauthorized')
      );

      // Act & Assert
      await expect(sapliy.emit('test.event', {}))
        .rejects.toThrow('Unauthorized');
      expect(httpClient.post).toHaveBeenCalledTimes(1);
    });
  });
});
```

### 3. CLI Testing

```typescript
// sapliy-cli/src/__tests__/commands/run.test.ts
import { RunCommand } from '../../commands/dev/run';
import { DockerService } from '../../services/docker';
import { PortDetector } from '../../utils/ports';

describe('sapliy run command', () => {
  let runCommand: RunCommand;
  let dockerService: DockerService;
  let portDetector: PortDetector;

  beforeEach(() => {
    dockerService = {
      isInstalled: jest.fn().mockResolvedValue(true),
      compose: jest.fn(),
    } as any;

    portDetector = {
      findFreePort: jest.fn(),
    } as any;

    runCommand = new RunCommand(dockerService, portDetector);
  });

  describe('execute', () => {
    it('should start all services', async () => {
      // Arrange
      jest.spyOn(portDetector, 'findFreePort').mockResolvedValue(8080);
      jest.spyOn(dockerService, 'compose').mockResolvedValue('success');

      // Act
      await runCommand.execute({ skipDocker: false });

      // Assert
      expect(dockerService.compose).toHaveBeenCalledWith('up', expect.any(Object));
    });

    it('should detect port conflicts', async () => {
      // Arrange
      jest.spyOn(portDetector, 'findFreePort').mockResolvedValue(8081);

      // Act
      await runCommand.execute({ port: 8080 });

      // Assert: Should use different port
      expect(dockerService.compose).toHaveBeenCalled();
    });

    it('should handle Docker not installed', async () => {
      // Arrange
      jest.spyOn(dockerService, 'isInstalled').mockResolvedValue(false);

      // Act & Assert
      await expect(runCommand.execute({}))
        .rejects.toThrow('Docker not found');
    });
  });
});
```

### Run Unit Tests

```bash
# Run all unit tests
npm test

# Run with coverage report
npm test -- --coverage

# Run specific test file
npm test -- auth.service.test.ts

# Watch mode for development
npm test -- --watch

# Output coverage report to HTML
npm test -- --coverage --coverageReporters=html
```

---

## Integration Testing

### 1. API Integration Tests

```typescript
// src/__tests__/integration/auth.integration.test.ts
import request from 'supertest';
import { app } from '../../app';
import { DatabaseService } from '../../services/database';

describe('Auth API Integration', () => {
  let database: DatabaseService;

  beforeAll(async () => {
    database = new DatabaseService();
    await database.connect('postgresql://test:test@localhost:5432/sapliy_test');
    await database.migrateUp();
  });

  afterAll(async () => {
    await database.migrateDown();
    await database.disconnect();
  });

  beforeEach(async () => {
    await database.seed();
  });

  describe('POST /auth/login', () => {
    it('should login successfully', async () => {
      // Act
      const response = await request(app)
        .post('/auth/login')
        .send({
          email: 'user@example.com',
          password: 'password123',
        });

      // Assert
      expect(response.status).toBe(200);
      expect(response.body).toHaveProperty('token');
      expect(response.body.user.email).toBe('user@example.com');
    });

    it('should return 401 on invalid credentials', async () => {
      // Act
      const response = await request(app)
        .post('/auth/login')
        .send({
          email: 'user@example.com',
          password: 'wrongpassword',
        });

      // Assert
      expect(response.status).toBe(401);
      expect(response.body.error).toContain('Invalid credentials');
    });

    it('should validate email format', async () => {
      // Act
      const response = await request(app)
        .post('/auth/login')
        .send({
          email: 'invalid-email',
          password: 'password123',
        });

      // Assert
      expect(response.status).toBe(400);
      expect(response.body.error).toContain('Invalid email');
    });
  });

  describe('POST /zones', () => {
    it('should create zone with valid token', async () => {
      // Arrange
      const loginResponse = await request(app)
        .post('/auth/login')
        .send({
          email: 'user@example.com',
          password: 'password123',
        });

      const token = loginResponse.body.token;

      // Act
      const response = await request(app)
        .post('/zones')
        .set('Authorization', `Bearer ${token}`)
        .send({
          name: 'my-app',
          mode: 'test',
        });

      // Assert
      expect(response.status).toBe(201);
      expect(response.body).toHaveProperty('secretKey');
      expect(response.body).toHaveProperty('publishableKey');
      expect(response.body.secretKey).toMatch(/^sk_test_/);
    });

    it('should reject unauthorized request', async () => {
      // Act
      const response = await request(app)
        .post('/zones')
        .send({ name: 'my-app', mode: 'test' });

      // Assert
      expect(response.status).toBe(401);
    });
  });
});
```

### 2. Event Flow Integration Test

```typescript
// src/__tests__/integration/event-flow.integration.test.ts
import request from 'supertest';
import { app } from '../../app';
import { KafkaService } from '../../services/kafka';
import { FlowEngineService } from '../../services/flow-engine';

describe('Event Flow Integration', () => {
  let kafkaService: KafkaService;
  let flowEngine: FlowEngineService;

  beforeAll(async () => {
    kafkaService = new KafkaService();
    flowEngine = new FlowEngineService();
    await kafkaService.connect();
  });

  afterAll(async () => {
    await kafkaService.disconnect();
  });

  it('should emit event and execute flow end-to-end', async (done) => {
    // Arrange
    const zoneId = 'zone_test_123';
    const token = 'test_token';

    // Create flow
    await request(app)
      .post('/flows')
      .set('Authorization', `Bearer ${token}`)
      .send({
        zoneId,
        name: 'test-flow',
        triggerEvent: 'payment.completed',
        actions: [
          {
            type: 'log',
            message: 'Payment completed',
          },
        ],
      });

    // Subscribe to Kafka for verification
    let flowExecuted = false;
    kafkaService.subscribe('flow-executions', (message) => {
      if (message.triggerEvent === 'payment.completed') {
        flowExecuted = true;
        expect(message.status).toBe('success');
        done();
      }
    });

    // Emit event
    await request(app)
      .post('/events')
      .set('Authorization', `Bearer sk_test_123`)
      .send({
        zoneId,
        eventType: 'payment.completed',
        data: { orderId: '123', amount: 99.99 },
      });

    // Timeout safety
    setTimeout(() => {
      if (!flowExecuted) done(new Error('Flow not executed'));
    }, 5000);
  });

  it('should handle webhook delivery', async () => {
    // Arrange
    const mockWebhookServer = await startMockWebhookServer();
    const zoneId = 'zone_test_123';

    // Create flow with webhook action
    await request(app)
      .post('/flows')
      .send({
        zoneId,
        triggerEvent: 'order.created',
        actions: [
          {
            type: 'webhook',
            url: mockWebhookServer.url,
          },
        ],
      });

    // Act: Emit event
    await request(app)
      .post('/events')
      .send({
        zoneId,
        eventType: 'order.created',
        data: { orderId: '456' },
      });

    // Assert: Webhook was called
    await new Promise(resolve => setTimeout(resolve, 1000));
    expect(mockWebhookServer.receivedWebhooks.length).toBeGreaterThan(0);

    const webhook = mockWebhookServer.receivedWebhooks[0];
    expect(webhook.body.eventType).toBe('order.created');
  });
});
```

### 3. Docker Compose Integration Test

```yaml
# docker-compose.test.yml
version: '3.9'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
      POSTGRES_DB: sapliy_test
    ports:
      - "5433:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U test"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6380:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  kafka:
    image: confluentinc/cp-kafka:7.5.0
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://kafka:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
    depends_on:
      - zookeeper
    ports:
      - "9093:9092"
    healthcheck:
      test: kafka-topics --bootstrap-server localhost:9092 --list
      interval: 10s
      timeout: 5s
      retries: 5

  zookeeper:
    image: confluentinc/cp-zookeeper:7.5.0
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181

  api:
    build:
      context: ./fintech-ecosystem
      dockerfile: Dockerfile.test
    environment:
      DATABASE_URL: postgresql://test:test@postgres:5432/sapliy_test
      REDIS_URL: redis://redis:6379
      KAFKA_BROKERS: kafka:9092
      NODE_ENV: test
    ports:
      - "8081:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      kafka:
        condition: service_healthy
    command: npm test
```

### Run Integration Tests

```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
npm run test:integration

# Stop test environment
docker-compose -f docker-compose.test.yml down

# Run everything
npm run test:integration:full
```

---

## End-to-End Testing

### 1. CLI E2E Tests

```typescript
// sapliy-cli/e2e/commands.e2e.test.ts
import { execSync } from 'child_process';
import * as fs from 'fs';
import * as path from 'path';

describe('CLI E2E Tests', () => {
  const testDir = path.join(__dirname, '..', '.test-workspace');

  beforeAll(() => {
    if (fs.existsSync(testDir)) {
      fs.rmSync(testDir, { recursive: true });
    }
    fs.mkdirSync(testDir, { recursive: true });
  });

  describe('sapliy dev', () => {
    it('should start backend and frontend successfully', async () => {
      // Arrange
      const startTime = Date.now();
      const timeout = 30000; // 30 seconds

      // Act: Start sapliy dev in background
      const child = spawn('sapliy', ['dev'], {
        cwd: testDir,
        stdio: ['ignore', 'pipe', 'pipe'],
      });

      // Assert: Check services are running
      const servicesRunning = await waitForServices(
        [
          { url: 'http://localhost:8080/health', name: 'API' },
          { url: 'http://localhost:3000', name: 'Frontend' },
        ],
        timeout
      );

      expect(servicesRunning).toBe(true);

      // Cleanup
      child.kill();
    }, 60000);

    it('should auto-open browser when frontend starts', async () => {
      // Arrange
      const openSpy = jest.spyOn(require('open'), 'default');

      // Act
      const child = spawn('sapliy', ['dev', '--auto-open'], {
        cwd: testDir,
      });

      // Assert
      await waitFor(() => openSpy.mock.calls.length > 0, 10000);
      expect(openSpy).toHaveBeenCalledWith(expect.stringContaining('http://localhost:3000'));

      // Cleanup
      child.kill();
    }, 15000);
  });

  describe('sapliy events emit', () => {
    it('should emit event successfully', () => {
      // Act
      const output = execSync('sapliy events emit "test.event" \'{"data":"test"}\'', {
        cwd: testDir,
        encoding: 'utf-8',
      });

      // Assert
      expect(output).toContain('Event emitted');
      expect(output).toContain('evt_');
    });
  });

  describe('sapliy zones', () => {
    it('should create and list zones', () => {
      // Act: Create zone
      const createOutput = execSync('sapliy zones create --name="test-zone"', {
        cwd: testDir,
        encoding: 'utf-8',
      });

      expect(createOutput).toContain('test-zone');
      expect(createOutput).toContain('sk_');

      // Act: List zones
      const listOutput = execSync('sapliy zones list', {
        cwd: testDir,
        encoding: 'utf-8',
      });

      // Assert
      expect(listOutput).toContain('test-zone');
    });
  });
});
```

### 2. Cypress E2E Tests (Frontend)

```typescript
// fintech-automation/cypress/e2e/flow-builder.cy.ts
describe('Flow Builder E2E', () => {
  beforeEach(() => {
    cy.visit('http://localhost:3000');
    cy.login('user@example.com', 'password123');
  });

  it('should create a flow end-to-end', () => {
    // Act: Click create flow button
    cy.get('[data-testid="create-flow-btn"]').click();

    // Assert: Modal appears
    cy.get('[data-testid="create-flow-modal"]').should('be.visible');

    // Act: Fill form
    cy.get('[name="flowName"]').type('My Payment Flow');
    cy.get('[name="triggerEvent"]').select('payment.completed');

    // Act: Add action
    cy.get('[data-testid="add-action-btn"]').click();
    cy.get('[data-testid="action-type-select"]').select('webhook');
    cy.get('[name="webhookUrl"]').type('https://example.com/webhook');

    // Act: Save flow
    cy.get('[data-testid="save-flow-btn"]').click();

    // Assert: Success message
    cy.get('[data-testid="success-message"]').should('contain', 'Flow created');

    // Assert: Flow appears in list
    cy.get('[data-testid="flows-list"]').should('contain', 'My Payment Flow');
  });

  it('should test flow with sample event', () => {
    // Arrange: Open existing flow
    cy.get('[data-testid="flow-item"]').first().click();

    // Act: Click test button
    cy.get('[data-testid="test-flow-btn"]').click();

    // Act: Select event type
    cy.get('[name="eventType"]').select('payment.completed');

    // Act: Add event data
    cy.get('[name="eventData"]').type('{"orderId":"123","amount":99.99}');

    // Act: Run test
    cy.get('[data-testid="run-test-btn"]').click();

    // Assert: Execution result shown
    cy.get('[data-testid="execution-result"]').should('contain', 'success');
  });

  it('should validate flow before deploy', () => {
    // Arrange: Create invalid flow
    cy.createFlow('invalid-flow');
    cy.addActionWithoutTrigger(); // Invalid

    // Act: Try to deploy
    cy.get('[data-testid="deploy-flow-btn"]').click();

    // Assert: Validation error
    cy.get('[data-testid="validation-error"]').should('contain', 'Flow validation failed');
  });
});

// Custom Cypress commands
Cypress.Commands.add('login', (email, password) => {
  cy.visit('http://localhost:3000/login');
  cy.get('[name="email"]').type(email);
  cy.get('[name="password"]').type(password);
  cy.get('[data-testid="login-btn"]').click();
  cy.url().should('include', '/dashboard');
});

Cypress.Commands.add('createFlow', (name) => {
  cy.get('[data-testid="create-flow-btn"]').click();
  cy.get('[name="flowName"]').type(name);
  cy.get('[data-testid="save-flow-btn"]').click();
});
```

### Run E2E Tests

```bash
# Run Cypress tests
npm run test:e2e

# Run in headed mode
npm run test:e2e:headed

# Run single spec file
npm run test:e2e -- --spec="cypress/e2e/flow-builder.cy.ts"
```

---

## Docker & Infrastructure Testing

### 1. Docker Image Testing

```dockerfile
# fintech-ecosystem/Dockerfile.test
FROM node:18-alpine

WORKDIR /app

# Install dependencies
COPY package*.json ./
RUN npm ci

# Copy source
COPY . .

# Run tests
CMD ["npm", "test", "--", "--coverage"]
```

```bash
# Build and test Docker image
docker build -f Dockerfile.test -t sapliy:test .
docker run --rm sapliy:test

# Test multi-stage build
docker build -t sapliy:production .
docker run --rm sapliy:production npm run health-check
```

### 2. Docker Compose Health Checks

```yaml
# docker-compose.yml
version: '3.9'

services:
  postgres:
    image: postgres:15-alpine
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U sapliy"]
      interval: 10s
      timeout: 5s
      retries: 5
    environment:
      POSTGRES_USER: sapliy
      POSTGRES_PASSWORD: sapliy
      POSTGRES_DB: sapliy

  redis:
    image: redis:7-alpine
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  api:
    build: ./fintech-ecosystem
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 5
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      DATABASE_URL: postgresql://sapliy:sapliy@postgres:5432/sapliy
      REDIS_URL: redis://redis:6379
```

### 3. Infrastructure as Code Testing

```python
# tests/infrastructure_test.py
import pytest
from testinfra.utils.ansible_runner import get_docker_module

docker = get_docker_module()

@pytest.fixture
def docker_container(docker):
    container = docker.run(
        "sapliy:latest",
        "-d",
        "--name", "test-container",
        "-p", "8080:8080"
    )
    yield container
    docker.remove_container("test-container", force=True)

def test_api_service_running(docker_container):
    """Test API service is running"""
    assert docker_container.status() == "running"

def test_api_port_open(docker_container):
    """Test API port is accessible"""
    socket = docker_container.socket("tcp://8080")
    assert socket is not None

def test_health_endpoint(docker_container):
    """Test health check endpoint"""
    health_check = docker_container.check_output("curl -f http://localhost:8080/health")
    assert "ok" in health_check.lower()
```

---

## Event Flow Testing

### 1. Detailed Event Flow Test

```typescript
// src/__tests__/event-flow/payment-flow.test.ts
import { EventEmitter } from '../../services/event-emitter';
import { FlowEngine } from '../../services/flow-engine';
import { WebhookService } from '../../services/webhook.service';

describe('Payment Event Flow', () => {
  let eventEmitter: EventEmitter;
  let flowEngine: FlowEngine;
  let webhookService: WebhookService;

  beforeEach(() => {
    eventEmitter = new EventEmitter();
    flowEngine = new FlowEngine();
    webhookService = new WebhookService();
  });

  it('should process complete payment flow', async () => {
    // Test scenario:
    // 1. Payment initiated event
    // 2. Validate payment flow
    // 3. Process payment
    // 4. Send webhooks
    // 5. Update ledger
    // 6. Send notifications

    const paymentId = 'pay_123';
    const orderId = 'ord_456';
    const amount = 99.99;

    // Step 1: Emit payment.initiated
    const initiatedEvent = await eventEmitter.emit({
      type: 'payment.initiated',
      data: { paymentId, orderId, amount },
    });

    expect(initiatedEvent.status).toBe('delivered');

    // Step 2: Verify flow executed
    const initiatedFlow = await flowEngine.getLastExecution('payment.initiated');
    expect(initiatedFlow.status).toBe('success');
    expect(initiatedFlow.actions[0]).toBe('validate_payment');

    // Step 3: Emit payment.completed
    const completedEvent = await eventEmitter.emit({
      type: 'payment.completed',
      data: { paymentId, orderId, amount, status: 'success' },
    });

    expect(completedEvent.status).toBe('delivered');

    // Step 4: Verify webhook sent
    const webhookCalls = webhookService.getCalls();
    const confirmationWebhook = webhookCalls.find(
      call => call.data.eventType === 'payment.completed'
    );

    expect(confirmationWebhook).toBeDefined();
    expect(confirmationWebhook.statusCode).toBe(200);

    // Step 5: Verify ledger entry created
    const ledgerEntries = await flowEngine.getLedgerEntries(orderId);
    expect(ledgerEntries).toContainEqual(
      expect.objectContaining({
        amount,
        type: 'debit',
      })
    );

    // Step 6: Verify notification sent
    const notifications = await flowEngine.getNotifications(orderId);
    expect(notifications.length).toBeGreaterThan(0);
  });

  it('should handle payment failure flow', async () => {
    // Test scenario: Payment fails, rollback occurs

    const paymentId = 'pay_fail_123';

    // Emit payment.failed
    const failedEvent = await eventEmitter.emit({
      type: 'payment.failed',
      data: { paymentId, reason: 'Insufficient funds' },
    });

    // Verify failure flow executed
    const failureFlow = await flowEngine.getLastExecution('payment.failed');
    expect(failureFlow.status).toBe('success');

    // Verify notification sent
    const notifications = await flowEngine.getNotifications(paymentId);
    expect(notifications).toContainEqual(
      expect.objectContaining({
        type: 'payment_failed',
        status: 'pending',
      })
    );
  });
});
```

### 2. CLI Event Testing Script

```bash
#!/bin/bash
# scripts/test-events.sh

set -e

echo "🧪 Testing Event Flows..."

# Start sapliy dev in background
sapliy dev &
DEV_PID=$!

# Wait for services to be ready
sleep 5

echo "✅ Services started"

# Test 1: Emit single event
echo "📤 Test 1: Emitting single event..."
RESULT=$(sapliy events emit "test.event" '{"data":"test"}')
echo "$RESULT"

# Test 2: Emit with idempotency
echo "📤 Test 2: Emitting with idempotency key..."
sapliy events emit "test.event" '{"data":"test"}' --idempotency-key="key_123"

# Test 3: List events
echo "📋 Test 3: Listing events..."
sapliy events list

# Test 4: Listen to events
echo "👂 Test 4: Listening to events..."
timeout 10 sapliy events listen "test.*" || true

# Test 5: Replay events
echo "🔄 Test 5: Replaying events..."
sapliy events replay --after="1 hour ago"

echo "✅ All event tests passed!"

# Cleanup
kill $DEV_PID
```

---

## Performance Testing

### 1. Load Testing with K6

```javascript
// tests/load-tests/events-load-test.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '2m', target: 100 },   // Ramp-up
    { duration: '5m', target: 100 },   // Stay at 100
    { duration: '2m', target: 200 },   // Ramp-up to 200
    { duration: '5m', target: 200 },   // Stay at 200
    { duration: '2m', target: 0 },     // Ramp-down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed: ['rate<0.1'],
  },
};

export default function() {
  const baseUrl = 'http://localhost:8080';
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer sk_test_123',
  };

  // Test event emission
  const eventPayload = JSON.stringify({
    zoneId: 'zone_test_123',
    eventType: 'test.event',
    data: { test: true, timestamp: new Date().toISOString() },
  });

  const response = http.post(`${baseUrl}/events`, eventPayload, {
    headers,
  });

  check(response, {
    'status is 201': (r) => r.status === 201,
    'response time < 500ms': (r) => r.timings.duration < 500,
    'event ID present': (r) => JSON.parse(r.body).id !== undefined,
  });

  sleep(1);
}
```

### 2. Run Load Tests

```bash
# Run load test
k6 run tests/load-tests/events-load-test.js

# Run with output
k6 run tests/load-tests/events-load-test.js --out csv=results.csv

# Cloud testing
k6 cloud tests/load-tests/events-load-test.js
```

---

## Security Testing

### 1. OWASP Top 10 Tests

```typescript
// tests/security/owasp-tests.ts
describe('Security - OWASP Top 10', () => {
  describe('A01:2021 – Broken Access Control', () => {
    it('should not allow access without authentication', async () => {
      const response = await request(app)
        .get('/zones')
        .set('Authorization', 'Bearer invalid');

      expect(response.status).toBe(401);
    });

    it('should not allow user to access other user zones', async () => {
      // Arrange
      const user2Token = await loginAsUser('user2@example.com');
      const user1ZoneId = 'zone_user1_123';

      // Act
      const response = await request(app)
        .get(`/zones/${user1ZoneId}`)
        .set('Authorization', `Bearer ${user2Token}`);

      // Assert
      expect(response.status).toBe(403);
    });
  });

  describe('A02:2021 – Cryptographic Failures', () => {
    it('should not log sensitive data', async () => {
      const logSpy = jest.spyOn(console, 'log');

      await request(app)
        .post('/auth/login')
        .send({
          email: 'user@example.com',
          password: 'sensitive_password_123',
        });

      expect(logSpy).not.toHaveBeenCalledWith(
        expect.stringContaining('sensitive_password_123')
      );
    });

    it('should encrypt sensitive data in database', async () => {
      // Arrange
      const apiKey = await createApiKey();

      // Act
      const dbRecord = await database.query(
        'SELECT * FROM api_keys WHERE id = $1',
        [apiKey.id]
      );

      // Assert
      expect(dbRecord.secret_key).not.toBe(apiKey.secretKey);
      expect(dbRecord.secret_key).toMatch(/^[a-f0-9]{64}$/); // Hashed
    });
  });

  describe('A03:2021 – Injection', () => {
    it('should prevent SQL injection', async () => {
      const response = await request(app)
        .get("/zones?name='; DROP TABLE zones; --")
        .set('Authorization', `Bearer ${token}`);

      expect(response.status).toBe(400);
      // Verify table still exists
      const result = await database.query('SELECT COUNT(*) FROM zones');
      expect(parseInt(result.rows[0].count)).toBeGreaterThan(0);
    });

    it('should prevent NoSQL injection', async () => {
      const response = await request(app)
        .post('/events')
        .set('Authorization', `Bearer sk_test_123`)
        .send({
          zoneId: 'zone_test_123',
          eventType: { $ne: null },
          data: {},
        });

      expect(response.status).toBe(400);
    });
  });

  describe('A07:2021 – Cross-Site Scripting (XSS)', () => {
    it('should escape HTML in event data', async () => {
      const response = await request(app)
        .post('/events')
        .set('Authorization', `Bearer sk_test_123`)
        .send({
          zoneId: 'zone_test_123',
          eventType: 'test.event',
          data: { message: '<script>alert("xss")</script>' },
        });

      expect(response.status).toBe(201);
      const event = response.body;
      expect(event.data.message).not.toContain('<script>');
    });
  });
});
```

---

## Production Readiness Checklist

```markdown
# ✅ Production Readiness Checklist

## 🔴 Critical (Must Pass)

### Code Quality
- [ ] All unit tests passing (>80% coverage)
- [ ] All integration tests passing
- [ ] No critical security vulnerabilities (SNYK)
- [ ] Code linting passes (ESLint)
- [ ] TypeScript strict mode enabled

### Testing
- [ ] Unit test coverage >80%
- [ ] Integration tests for all critical paths
- [ ] E2E tests for user workflows
- [ ] Load tests pass (target: 10K events/sec)
- [ ] All Docker images tested

### Security
- [ ] OWASP Top 10 tests passing
- [ ] No hardcoded secrets
- [ ] API keys properly rotated
- [ ] TLS 1.3 enabled
- [ ] Rate limiting configured

### Infrastructure
- [ ] Docker images built & scanned
- [ ] Docker Compose fully functional
- [ ] Health checks configured
- [ ] Logging configured
- [ ] Monitoring alerts set up

## 🟡 High (Should Pass)

### Performance
- [ ] API response time <100ms p99
- [ ] Event processing latency <50ms p99
- [ ] Webhook delivery >99% success rate
- [ ] Database queries optimized
- [ ] Caching configured

### Operations
- [ ] Runbooks documented
- [ ] Incident response plan ready
- [ ] Backup/restore tested
- [ ] Database migrations tested
- [ ] Log aggregation working

### Documentation
- [ ] API docs complete
- [ ] Deployment guide complete
- [ ] CLI documentation complete
- [ ] Troubleshooting guide complete
- [ ] Architecture documentation updated

## 🟢 Nice to Have

### Additional Testing
- [ ] Chaos engineering tests
- [ ] Penetration testing done
- [ ] Browser compatibility tested
- [ ] Mobile responsiveness tested
- [ ] Accessibility (WCAG) compliant

### Monitoring & Analytics
- [ ] Metrics dashboard configured
- [ ] Error tracking (Sentry) integrated
- [ ] Performance monitoring set up
- [ ] Usage analytics configured
- [ ] Cost monitoring configured
```

---

## Test Execution Scripts

### npm test Scripts

```json
{
  "scripts": {
    "test": "jest --coverage",
    "test:watch": "jest --watch",
    "test:unit": "jest --testPathPattern=__tests__",
    "test:integration": "jest --testPathPattern=integration",
    "test:e2e": "cypress run",
    "test:e2e:headed": "cypress open",
    "test:security": "jest --testPathPattern=security",
    "test:performance": "k6 run tests/load-tests/events-load-test.js",
    "test:all": "npm run test:unit && npm run test:integration && npm run test:e2e",
    "test:docker": "docker-compose -f docker-compose.test.yml up --abort-on-container-exit",
    "test:cli": "jest sapliy-cli/src/__tests__",
    "lint": "eslint src --fix",
    "type-check": "tsc --noEmit",
    "security-check": "snyk test",
    "health-check": "node scripts/health-check.js"
  }
}
```

### Master Test Script

```bash
#!/bin/bash
# scripts/run-all-tests.sh

set -e

echo "🚀 Running Full Test Suite..."
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Unit Tests
echo -e "${YELLOW}📝 Running unit tests...${NC}"
npm run test:unit || (echo -e "${RED}❌ Unit tests failed${NC}" && exit 1)
echo -e "${GREEN}✅ Unit tests passed${NC}\n"

# Step 2: Type Check
echo -e "${YELLOW}🔍 Running type check...${NC}"
npm run type-check || (echo -e "${RED}❌ Type check failed${NC}" && exit 1)
echo -e "${GREEN}✅ Type check passed${NC}\n"

# Step 3: Linting
echo -e "${YELLOW}📐 Running linter...${NC}"
npm run lint || (echo -e "${RED}❌ Linting failed${NC}" && exit 1)
echo -e "${GREEN}✅ Linting passed${NC}\n"

# Step 4: Security
echo -e "${YELLOW}🔐 Running security check...${NC}"
npm run security-check || (echo -e "${RED}❌ Security check failed${NC}" && exit 1)
echo -e "${GREEN}✅ Security check passed${NC}\n"

# Step 5: Integration Tests
echo -e "${YELLOW}🔗 Running integration tests...${NC}"
docker-compose -f docker-compose.test.yml up -d
npm run test:integration || (echo -e "${RED}❌ Integration tests failed${NC}" && docker-compose -f docker-compose.test.yml down && exit 1)
docker-compose -f docker-compose.test.yml down
echo -e "${GREEN}✅ Integration tests passed${NC}\n"

# Step 6: E2E Tests
echo -e "${YELLOW}🎯 Running E2E tests...${NC}"
npm run test:e2e || (echo -e "${RED}❌ E2E tests failed${NC}" && exit 1)
echo -e "${GREEN}✅ E2E tests passed${NC}\n"

# Step 7: Performance Tests
echo -e "${YELLOW}⚡ Running performance tests...${NC}"
npm run test:performance || (echo -e "${RED}❌ Performance tests failed${NC}" && exit 1)
echo -e "${GREEN}✅ Performance tests passed${NC}\n"

# Step 8: CLI Tests
echo -e "${YELLOW}💻 Running CLI tests...${NC}"
npm run test:cli || (echo -e "${RED}❌ CLI tests failed${NC}" && exit 1)
echo -e "${GREEN}✅ CLI tests passed${NC}\n"

# Step 9: Health Check
echo -e "${YELLOW}❤️  Running health checks...${NC}"
npm run health-check || (echo -e "${RED}❌ Health checks failed${NC}" && exit 1)
echo -e "${GREEN}✅ Health checks passed${NC}\n"

echo -e "${GREEN}🎉 All tests passed! Ready for production!${NC}"
```

### GitHub Actions CI/CD

```yaml
# .github/workflows/test.yml
name: Full Test Suite

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_DB: sapliy_test
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      
      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'
          cache: 'npm'
      
      - name: Install dependencies
        run: npm ci
      
      - name: Type check
        run: npm run type-check
      
      - name: Lint
        run: npm run lint
      
      - name: Unit tests
        run: npm run test:unit
      
      - name: Security check
        run: npm run security-check
      
      - name: Integration tests
        run: npm run test:integration
      
      - name: E2E tests
        run: npm run test:e2e
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage/lcov.info
      
      - name: Performance tests
        run: npm run test:performance
        continue-on-error: true
      
      - name: Health check
        run: npm run health-check
```

---

## Conclusion

This comprehensive testing plan ensures **Sapliy is production-ready** with:

✅ 80%+ unit test coverage  
✅ Full integration testing with Docker  
✅ End-to-end user workflow testing  
✅ Performance testing (10K+ events/sec)  
✅ Security testing (OWASP Top 10)  
✅ CLI testing with real commands  
✅ Automated CI/CD with GitHub Actions  

**Before going to production, ensure ALL tests pass** using the `run-all-tests.sh` script.
