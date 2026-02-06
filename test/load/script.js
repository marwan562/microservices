import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const throughput = new Counter('throughput');

export const options = {
    scenarios: {
        // 1. Smoke test: Verify system is healthy
        smoke: {
            executor: 'constant-vus',
            vus: 1,
            duration: '30s',
            tags: { test_type: 'smoke' },
        },
        // 2. Load test: Simulate peak traffic
        load: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '2m', target: 50 },  // Ramp up
                { duration: '5m', target: 50 },  // Stay at peak
                { duration: '2m', target: 0 },   // Ramp down
            ],
            startTime: '30s',
            tags: { test_type: 'load' },
        },
        // 3. Stress test: Find breaking point
        stress: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '2m', target: 100 },
                { duration: '5m', target: 200 },
                { duration: '5m', target: 300 }, // Push hard
                { duration: '2m', target: 0 },
            ],
            startTime: '8m', // Run after load test
            tags: { test_type: 'stress' },
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95% of requests must be < 500ms
        errors: ['rate<0.01'],            // Error rate must be < 1%
    },
};

const BASE_URL = 'http://localhost:8080/api/v1';

export default function () {
    const params = {
        headers: {
            'Content-Type': 'application/json',
            'X-Tenant-ID': 'tenant_123', // Simulate multi-tenancy
        },
    };

    // 1. Health check
    const healthRes = http.get(`${BASE_URL}/health`, params);
    check(healthRes, {
        'health status is 200': (r) => r.status === 200,
    });

    // 2. Secrets access (simulating core feature)
    const secretRes = http.get(`${BASE_URL}/secrets/stripe_key`, params);
    const success = check(secretRes, {
        'GET secret is 200': (r) => r.status === 200,
    });

    if (!success) {
        errorRate.add(1);
    } else {
        throughput.add(1);
    }

    // 3. Rate limited endpoint
    // We expect some 429s here during stress test
    const billingRes = http.get(`${BASE_URL}/billing/summary`, params);
    check(billingRes, {
        'billing status is 200 or 429': (r) => r.status === 200 || r.status === 429,
    });

    sleep(1);
}
