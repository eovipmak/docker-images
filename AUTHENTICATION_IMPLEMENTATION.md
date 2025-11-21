# Authentication Implementation Guide

This document describes the complete authentication implementation for V-Insight frontend.

## Overview

The authentication system implements:
- User registration and login with JWT tokens
- Automatic token management in localStorage
- Protected routes with automatic redirect
- User session management
- Automatic token injection in API requests
- 401 error handling with auto-redirect

## Architecture

### 1. Auth Store (`/frontend/src/lib/stores/auth.ts`)

The auth store manages authentication state using Svelte's writable store pattern.

**State Structure:**
```typescript
interface AuthState {
  isAuthenticated: boolean;
  currentUser: User | null;
}

interface User {
  id: number;
  email: string;
  tenant_id: number;
}
```

**Key Methods:**

- `login(token: string)`: Stores JWT token and fetches user data
  - Saves token to localStorage
  - Updates isAuthenticated to true
  - Fetches user info from `/api/v1/auth/me`
  - Updates currentUser state

- `logout()`: Clears authentication
  - Removes token from localStorage
  - Resets state to unauthenticated

- `checkAuth()`: Validates existing session
  - Called on app mount
  - Checks localStorage for token
  - Validates token with backend
  - Updates state accordingly

- `getToken()`: Returns current JWT token or null

**Backward Compatibility:**
The `isAuthenticated` export is maintained for compatibility with existing code.

### 2. API Client (`/frontend/src/lib/api/client.ts`)

Enhanced fetch wrapper with automatic authentication.

**Features:**
- Automatically adds `Authorization: Bearer <token>` header
- Retrieves token from localStorage
- Handles 401 responses by redirecting to login
- `skipAuth` option for public endpoints

**Usage:**
```typescript
// Authenticated request (automatic)
const response = await fetchAPI('/api/v1/monitors');

// Public request (skip auth)
const response = await fetchAPI('/api/v1/auth/login', { 
  skipAuth: true,
  method: 'POST',
  body: JSON.stringify({ email, password })
});
```

### 3. Route Protection (`/frontend/src/routes/+layout.svelte`)

Layout component implements global route protection.

**Public Routes:**
- `/` - Home page
- `/login` - Login page
- `/register` - Registration page

**Protected Routes:**
All other routes require authentication.

**Implementation:**
- Uses reactive statement to check auth on route changes
- Redirects to `/login` if accessing protected route while unauthenticated
- Calls `checkAuth()` on mount to validate session

### 4. Login Page (`/frontend/src/routes/login/+page.svelte`)

**Features:**
- Email and password input fields
- Client-side validation
- Error message display
- Loading states during submission
- Redirects to `/dashboard` on success

**Flow:**
1. User enters credentials
2. POST to `/api/v1/auth/login`
3. Receive JWT token
4. Call `authStore.login(token)` to store and fetch user
5. Redirect to dashboard

### 5. Registration Page (`/frontend/src/routes/register/+page.svelte`)

**Features:**
- Email, password, confirm password, and tenant name fields
- Password validation (minimum 6 characters)
- Password match validation
- Error message display
- Loading states during submission
- Redirects to `/dashboard` on success

**Flow:**
1. User fills registration form
2. Validate password match and length
3. POST to `/api/v1/auth/register`
4. Receive JWT token
5. Call `authStore.login(token)` to store and fetch user
6. Redirect to dashboard

### 6. Navigation Component (`/frontend/src/lib/components/Nav.svelte`)

**Features:**
- Shows different menu items based on auth state
- Authenticated users see: Home, Dashboard, Domains, Alerts, Settings, Logout
- Unauthenticated users see: Login
- Calls `checkAuth()` on mount
- Logout button clears session and redirects to login

## API Endpoints

### Public Endpoints

**POST /api/v1/auth/register**
```json
Request:
{
  "email": "user@example.com",
  "password": "password123",
  "tenant_name": "My Company"
}

Response:
{
  "token": "eyJhbGc..."
}
```

**POST /api/v1/auth/login**
```json
Request:
{
  "email": "user@example.com",
  "password": "password123"
}

Response:
{
  "token": "eyJhbGc..."
}
```

### Protected Endpoints

**GET /api/v1/auth/me**
```
Headers:
Authorization: Bearer <token>

Response:
{
  "id": 1,
  "email": "user@example.com",
  "tenant_id": 1
}
```

## User Flow

### Registration Flow
1. User navigates to `/register`
2. Fills out registration form
3. Submits form
4. Backend creates user and tenant
5. Backend returns JWT token
6. Frontend stores token and fetches user data
7. User redirected to `/dashboard`

### Login Flow
1. User navigates to `/login`
2. Enters email and password
3. Submits form
4. Backend validates credentials
5. Backend returns JWT token
6. Frontend stores token and fetches user data
7. User redirected to `/dashboard`

### Session Validation Flow
1. User opens app
2. Layout component calls `checkAuth()`
3. Auth store retrieves token from localStorage
4. Sends request to `/api/v1/auth/me`
5. If valid: Store user data, stay on current page
6. If invalid: Clear token, redirect to login (if on protected route)

### Logout Flow
1. User clicks logout button
2. Auth store clears localStorage
3. Auth state reset to unauthenticated
4. User redirected to `/login`

### Protected Route Access
1. User tries to access protected route (e.g., `/dashboard`)
2. Layout reactive statement checks authentication
3. If authenticated: Allow access
4. If not authenticated: Redirect to `/login`

## Security Considerations

1. **Token Storage**: JWT tokens stored in localStorage
   - Accessible only to same-origin scripts
   - Cleared on logout
   - Validated on every app load

2. **Automatic Token Injection**: All API requests automatically include token
   - Reduces chance of forgetting to add auth header
   - Consistent authentication across app

3. **401 Handling**: Invalid/expired tokens trigger automatic logout
   - Prevents access with invalid credentials
   - Graceful degradation to login page

4. **Route Protection**: Server-side and client-side protection
   - Client-side redirect for UX
   - Backend validates all requests

5. **CORS-Free Architecture**: Proxy eliminates CORS issues
   - All requests appear from same origin
   - No CORS headers needed

## Testing

### Manual Testing Steps

1. **Registration:**
   - Navigate to `/register`
   - Fill form with valid data
   - Verify redirect to dashboard
   - Check localStorage for token

2. **Login:**
   - Logout if logged in
   - Navigate to `/login`
   - Enter credentials
   - Verify redirect to dashboard
   - Check localStorage for token

3. **Route Protection:**
   - Logout
   - Try accessing `/dashboard` directly
   - Verify redirect to `/login`

4. **Session Persistence:**
   - Login
   - Refresh page
   - Verify still logged in
   - Check user info displayed in nav

5. **Logout:**
   - Click logout in nav
   - Verify redirect to login
   - Check localStorage cleared
   - Try accessing protected route

### E2E Tests

Run existing Playwright tests:
```bash
cd frontend
npm run test:e2e
```

Tests cover:
- Registration with valid data
- Login with valid credentials
- Protected endpoint access
- Invalid token rejection
- Validation errors

## Troubleshooting

### Token not being sent with requests
- Check browser localStorage for `auth_token`
- Verify `fetchAPI` is being used instead of plain `fetch`
- Check if `skipAuth: true` was accidentally set

### 401 errors on protected endpoints
- Verify token exists in localStorage
- Check token format (should be JWT)
- Verify backend is running and accessible
- Check token expiration

### Infinite redirect loops
- Check if route is in `publicRoutes` array
- Verify `checkAuth()` is working properly
- Check browser console for errors

### User not redirected after login
- Check for JavaScript errors in console
- Verify `window.location.href = '/dashboard'` executes
- Check if backend returned token properly

## Future Enhancements

Potential improvements:
- Token refresh mechanism
- Remember me functionality
- Password reset flow
- Email verification
- Multi-factor authentication
- Session timeout warnings
- Secure cookie storage (alternative to localStorage)
