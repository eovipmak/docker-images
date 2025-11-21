# Authentication Implementation - Testing Guide

## Summary of Changes

This PR implements complete authentication functionality for V-Insight frontend:

### Files Modified

1. **`/frontend/src/lib/stores/auth.ts`** - Enhanced auth store
   - Added `currentUser` state with User interface
   - `login()` now fetches user data from `/auth/me`
   - `checkAuth()` validates tokens and loads user data
   - Added `getToken()` helper method

2. **`/frontend/src/lib/api/client.ts`** - Enhanced API client
   - Automatic JWT token injection from localStorage
   - 401 response handling with auto-redirect
   - Added `skipAuth` option for public endpoints

3. **`/frontend/src/routes/+layout.svelte`** - Route protection
   - Reactive route protection using Svelte's `$:` syntax
   - Calls `checkAuth()` on mount
   - Redirects to login for protected routes

4. **`/frontend/src/lib/components/Nav.svelte`** - Updated navigation
   - Uses new `authStore` structure
   - Shows user-specific menu items

5. **`/frontend/src/routes/login/+page.svelte`** - Login page
   - Calls async `login()` method
   - Proper error handling

6. **`/frontend/src/routes/register/+page.svelte`** - Registration page
   - Calls async `login()` method
   - Proper error handling

### Files Created

1. **`/AUTHENTICATION_IMPLEMENTATION.md`** - Complete implementation documentation
2. **`/frontend/src/lib/api/README.md`** - API client usage guide

## Testing Instructions

### Prerequisites

1. Ensure Docker is installed
2. Clone the repository
3. Copy environment files:
   ```bash
   cp .env.example .env
   cp frontend/.env.example frontend/.env
   ```

### Running the Application

```bash
# Start all services
make up

# Wait ~30 seconds for PostgreSQL to initialize

# Check service status
make ps
```

Services should be running at:
- Frontend: http://localhost:3000
- Backend: http://localhost:8080
- Worker: http://localhost:8081
- PostgreSQL: localhost:5432

### Manual Testing Checklist

#### 1. Registration Flow
- [ ] Navigate to http://localhost:3000/register
- [ ] Fill in the form:
  - Email: `test@example.com`
  - Organization Name: `Test Company`
  - Password: `password123`
  - Confirm Password: `password123`
- [ ] Click "Sign Up"
- [ ] Verify redirect to `/dashboard`
- [ ] Check browser localStorage has `auth_token`
- [ ] Verify navigation shows authenticated menu items

#### 2. Logout Flow
- [ ] Click "Logout" in navigation
- [ ] Verify redirect to `/login`
- [ ] Verify localStorage `auth_token` is cleared
- [ ] Verify navigation shows public menu items

#### 3. Login Flow
- [ ] Navigate to http://localhost:3000/login
- [ ] Fill in credentials:
  - Email: `test@example.com`
  - Password: `password123`
- [ ] Click "Sign In"
- [ ] Verify redirect to `/dashboard`
- [ ] Check browser localStorage has `auth_token`
- [ ] Verify navigation shows authenticated menu items

#### 4. Route Protection
- [ ] Logout (if logged in)
- [ ] Try to access http://localhost:3000/dashboard directly
- [ ] Verify automatic redirect to `/login`
- [ ] Try to access http://localhost:3000/domains directly
- [ ] Verify automatic redirect to `/login`
- [ ] Login
- [ ] Navigate to `/dashboard`
- [ ] Verify access granted (no redirect)

#### 5. Session Persistence
- [ ] Login to the application
- [ ] Refresh the page (F5)
- [ ] Verify still logged in (no redirect to login)
- [ ] Verify user info still displayed in nav
- [ ] Navigate to different routes
- [ ] Verify authentication persists

#### 6. Invalid Token Handling
- [ ] Login to the application
- [ ] Open browser DevTools > Application > Local Storage
- [ ] Modify `auth_token` value to invalid string
- [ ] Navigate to any protected route (e.g., `/dashboard`)
- [ ] Verify automatic redirect to `/login`
- [ ] Verify localStorage `auth_token` is cleared

#### 7. Error Handling
- [ ] Navigate to `/login`
- [ ] Enter invalid credentials
- [ ] Verify error message displayed
- [ ] Navigate to `/register`
- [ ] Enter mismatched passwords
- [ ] Verify error message displayed
- [ ] Enter password < 6 characters
- [ ] Verify error message displayed

### Automated Testing

```bash
# TypeScript validation
cd frontend
npm run check

# E2E tests (requires services running)
cd frontend
npm run test:e2e
```

### Viewing Logs

```bash
# All services
make logs

# Backend only
make logs-backend

# Frontend only
make logs-frontend

# Worker only
make logs-worker
```

## Expected Behavior

### Successful Registration
1. Form submits to `/api/v1/auth/register`
2. Backend creates user and tenant
3. Backend returns JWT token
4. Frontend stores token in localStorage
5. Frontend calls `/api/v1/auth/me` to get user data
6. Frontend updates auth state
7. User redirected to `/dashboard`

### Successful Login
1. Form submits to `/api/v1/auth/login`
2. Backend validates credentials
3. Backend returns JWT token
4. Frontend stores token in localStorage
5. Frontend calls `/api/v1/auth/me` to get user data
6. Frontend updates auth state
7. User redirected to `/dashboard`

### Route Protection
1. User navigates to protected route
2. Layout checks `authStore.isAuthenticated`
3. If false and route not public → redirect to `/login`
4. If true → allow access

### API Request Authentication
1. Component calls `fetchAPI('/api/v1/monitors')`
2. Client retrieves token from localStorage
3. Client adds `Authorization: Bearer <token>` header
4. Request sent to backend via proxy
5. Backend validates token
6. If valid → return data
7. If invalid → return 401
8. Client catches 401 → clear token → redirect to login

## Troubleshooting

### Services not starting
```bash
# Check if .env exists
ls -la .env

# If not, copy from example
cp .env.example .env

# Restart services
make down
make up
```

### Permission errors
```bash
# Fix frontend permissions
sudo chown -R $USER:$USER frontend/node_modules frontend/.svelte-kit

# Rebuild
make rebuild
```

### Database connection errors
```bash
# Check PostgreSQL is healthy
docker compose ps postgres

# View PostgreSQL logs
make logs-postgres

# Wait for initialization (can take 30+ seconds on first run)
```

### Hot reload not working
```bash
# Check Air logs for backend
make logs-backend

# Check Vite logs for frontend
make logs-frontend

# Look for syntax errors
cat backend/build-errors.log
```

## Verification Commands

```bash
# Check backend is responding
curl http://localhost:8080/health

# Check API endpoint
curl http://localhost:8080/api/v1

# Test registration (replace with unique email)
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","tenant_name":"Test"}'

# Test login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Test protected endpoint (replace TOKEN with actual token)
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer TOKEN"
```

## Success Criteria

All of the following should be true:

✅ User can register with email, password, and tenant name  
✅ User can login with email and password  
✅ JWT token is stored in localStorage  
✅ Token is automatically sent with API requests  
✅ User info is fetched and stored on login  
✅ Protected routes redirect to login when not authenticated  
✅ Public routes (/, /login, /register) are accessible without auth  
✅ Session persists across page refreshes  
✅ Logout clears token and redirects to login  
✅ Invalid tokens trigger auto-redirect to login  
✅ Navigation shows correct menu items based on auth state  
✅ TypeScript checks pass (`npm run check`)  
✅ E2E tests pass (`npm run test:e2e`)

## Next Steps

After successful testing, consider:

1. Add password reset functionality
2. Implement token refresh mechanism
3. Add email verification
4. Implement remember me option
5. Add multi-factor authentication
6. Add session timeout warnings
7. Implement user profile page
8. Add password change functionality
