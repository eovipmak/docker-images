import { test, expect } from '@playwright/test';

// Test configuration
const BACKEND_URL = process.env.BACKEND_API_URL || 'http://localhost:8080';
const API_BASE = '/api/v1';

// Test data
const TEST_USER = {
	email: `test-${Date.now()}@example.com`,
	password: 'testpassword123',
	tenantName: `Test Tenant ${Date.now()}`,
};

let authToken = '';

test.describe('Authentication E2E Tests', () => {
	test.describe.configure({ mode: 'serial' });

	test('✅ User can register', async ({ request }) => {
		const response = await request.post(`${BACKEND_URL}${API_BASE}/auth/register`, {
			data: {
				email: TEST_USER.email,
				password: TEST_USER.password,
				tenant_name: TEST_USER.tenantName,
			},
		});

		expect(response.ok()).toBeTruthy();
		expect(response.status()).toBe(201);

		const responseBody = await response.json();
		expect(responseBody).toHaveProperty('token');
		expect(typeof responseBody.token).toBe('string');
		expect(responseBody.token.length).toBeGreaterThan(0);

		// Store token for later tests
		authToken = responseBody.token;
		console.log('✅ Registration successful, token received');
	});

	test('✅ User can login and receive JWT token', async ({ request }) => {
		const response = await request.post(`${BACKEND_URL}${API_BASE}/auth/login`, {
			data: {
				email: TEST_USER.email,
				password: TEST_USER.password,
			},
		});

		expect(response.ok()).toBeTruthy();
		expect(response.status()).toBe(200);

		const responseBody = await response.json();
		expect(responseBody).toHaveProperty('token');
		expect(typeof responseBody.token).toBe('string');
		expect(responseBody.token.length).toBeGreaterThan(0);

		// Update token with login token
		authToken = responseBody.token;
		console.log('✅ Login successful, JWT token received');
	});

	test('✅ Protected endpoint requires valid token', async ({ request }) => {
		// Test 1: Access protected endpoint WITHOUT token (should fail)
		const responseWithoutToken = await request.get(`${BACKEND_URL}${API_BASE}/auth/me`);
		expect(responseWithoutToken.status()).toBe(401);

		const errorBody = await responseWithoutToken.json();
		expect(errorBody).toHaveProperty('error');
		console.log('✅ Protected endpoint correctly rejects request without token');

		// Test 2: Access protected endpoint WITH invalid token (should fail)
		const responseWithInvalidToken = await request.get(`${BACKEND_URL}${API_BASE}/auth/me`, {
			headers: {
				Authorization: 'Bearer invalid-token-here',
			},
		});
		expect(responseWithInvalidToken.status()).toBe(401);

		const invalidTokenError = await responseWithInvalidToken.json();
		expect(invalidTokenError).toHaveProperty('error');
		console.log('✅ Protected endpoint correctly rejects invalid token');

		// Test 3: Access protected endpoint WITH valid token (should succeed)
		const responseWithValidToken = await request.get(`${BACKEND_URL}${API_BASE}/auth/me`, {
			headers: {
				Authorization: `Bearer ${authToken}`,
			},
		});

		expect(responseWithValidToken.ok()).toBeTruthy();
		expect(responseWithValidToken.status()).toBe(200);

		const userData = await responseWithValidToken.json();
		expect(userData).toHaveProperty('id');
		expect(userData).toHaveProperty('email');
		expect(userData.email).toBe(TEST_USER.email);
		console.log('✅ Protected endpoint accepts valid token and returns user data');
	});

	test('❌ Registration fails with invalid email', async ({ request }) => {
		const response = await request.post(`${BACKEND_URL}${API_BASE}/auth/register`, {
			data: {
				email: 'invalid-email',
				password: TEST_USER.password,
				tenant_name: TEST_USER.tenantName,
			},
		});

		expect(response.status()).toBe(400);
		const errorBody = await response.json();
		expect(errorBody).toHaveProperty('error');
		console.log('✅ Registration correctly rejects invalid email');
	});

	test('❌ Registration fails with short password', async ({ request }) => {
		const response = await request.post(`${BACKEND_URL}${API_BASE}/auth/register`, {
			data: {
				email: `test-short-pass-${Date.now()}@example.com`,
				password: '12345',
				tenant_name: TEST_USER.tenantName,
			},
		});

		expect(response.status()).toBe(400);
		const errorBody = await response.json();
		expect(errorBody).toHaveProperty('error');
		console.log('✅ Registration correctly rejects short password');
	});

	test('❌ Login fails with wrong password', async ({ request }) => {
		const response = await request.post(`${BACKEND_URL}${API_BASE}/auth/login`, {
			data: {
				email: TEST_USER.email,
				password: 'wrongpassword',
			},
		});

		expect(response.status()).toBe(401);
		const errorBody = await response.json();
		expect(errorBody).toHaveProperty('error');
		console.log('✅ Login correctly rejects wrong password');
	});

	test('❌ Login fails with non-existent user', async ({ request }) => {
		const response = await request.post(`${BACKEND_URL}${API_BASE}/auth/login`, {
			data: {
				email: 'nonexistent@example.com',
				password: 'anypassword',
			},
		});

		expect(response.status()).toBe(401);
		const errorBody = await response.json();
		expect(errorBody).toHaveProperty('error');
		console.log('✅ Login correctly rejects non-existent user');
	});
});
