# JWT Authentication System Documentation

## Overview

The SSL Monitor application now includes a complete JWT (JSON Web Token) authentication system built with FastAPI Users. This system provides secure multi-user support with email-based registration, login, email verification, and password reset functionality.

## Features

### Core Authentication
- **Email-based authentication**: Users register and login with email addresses
- **JWT tokens**: Secure, stateless authentication with 1-hour token expiration
- **Password hashing**: Bcrypt algorithm for secure password storage
- **Protected routes**: Both frontend and backend routes require authentication
- **Automatic token management**: Frontend automatically includes tokens in API requests

### User Management
- **User roles**: Support for regular users and superusers
- **Account status**: Active/inactive user management
- **Email verification**: Built-in verification system (hooks ready for email integration)
- **Password reset**: Secure password reset flow with tokens

### Security Features
- **CORS protection**: Configured for frontend-backend communication
- **Token expiration**: 1-hour access token lifetime
- **401 handling**: Automatic redirect to login on unauthorized access
- **Password validation**: Minimum password requirements
- **SQL injection protection**: SQLAlchemy ORM prevents SQL injection

## Architecture

### Backend Stack

```
FastAPI Users (13.0.0)
    └── JWT Strategy (python-jose)
    └── User Manager
    └── SQLAlchemy Adapter
            └── Async SQLite (aiosqlite)
                └── User Model
```

### Frontend Stack

```
React App
    └── AuthProvider (Context)
        └── useAuth Hook
            └── authService
                └── Axios Interceptors
                    └── API Calls
```

## API Endpoints

### Authentication Endpoints

#### POST /auth/register
Register a new user account.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "your_password"
}
```

**Response:**
```json
{
  "id": 1,
  "email": "user@example.com",
  "is_active": true,
  "is_superuser": false,
  "is_verified": false
}
```

#### POST /auth/jwt/login
Login and receive a JWT token.

**Request:**
```
Content-Type: application/x-www-form-urlencoded

username=user@example.com&password=your_password
```

**Response:**
```json
{
  "access_token": "eyJhbGci...",
  "token_type": "bearer"
}
```

#### POST /auth/jwt/logout
Logout (invalidates client-side token).

**Headers:**
```
Authorization: Bearer <token>
```

#### GET /auth/me
Get current authenticated user information.

**Headers:**
```
Authorization: Bearer <token>
```

**Response:**
```json
{
  "id": 1,
  "email": "user@example.com",
  "is_active": true,
  "is_superuser": false,
  "is_verified": false
}
```

#### POST /auth/forgot-password
Request password reset.

**Request:**
```json
{
  "email": "user@example.com"
}
```

#### POST /auth/reset-password
Reset password with token.

**Request:**
```json
{
  "token": "reset_token_here",
  "password": "new_password"
}
```

#### POST /auth/request-verify-token
Request email verification token.

**Headers:**
```
Authorization: Bearer <token>
```

#### POST /auth/verify
Verify email address.

**Request:**
```json
{
  "token": "verification_token_here"
}
```

## Frontend Implementation

### AuthContext

The `AuthContext` provides global authentication state:

```typescript
{
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  isAuthenticated: boolean;
}
```

### useAuth Hook

Access authentication state and methods in any component:

```typescript
import { useAuth } from '../hooks/useAuth';

function MyComponent() {
  const { user, isAuthenticated, login, logout } = useAuth();
  
  // Use auth state and methods
}
```

### Protected Routes

Protect routes that require authentication:

```typescript
<Route
  path="/dashboard"
  element={
    <ProtectedRoute>
      <Dashboard />
    </ProtectedRoute>
  }
/>
```

### API Service

The API service automatically includes auth tokens:

```typescript
import api from '../services/api';

// Token is automatically included
const response = await api.get('/auth/me');
```

## Database Schema

### Users Table

| Column | Type | Description |
|--------|------|-------------|
| id | INTEGER | Primary key |
| email | VARCHAR | Unique email address (indexed) |
| hashed_password | VARCHAR | Bcrypt hashed password |
| is_active | BOOLEAN | Account active status |
| is_superuser | BOOLEAN | Superuser flag |
| is_verified | BOOLEAN | Email verification status |
| created_at | DATETIME | Account creation timestamp |

## Configuration

### Environment Variables

Set these environment variables in production:

```bash
# JWT Secret Key (REQUIRED in production)
JWT_SECRET_KEY=your-very-secure-random-secret-key-min-32-characters

# SSL Checker URL
SSL_CHECKER_URL=http://ssl-checker:8000
```

### Generate Secret Key

```bash
python -c "import secrets; print(secrets.token_urlsafe(32))"
```

## Migration Guide

### Upgrading from Old User Model

The migration automatically updates the users table:

```bash
cd api
alembic upgrade head
```

**Note:** The migration drops and recreates the users table. Existing users will need to re-register.

## Security Best Practices

### Production Deployment

1. **Use HTTPS**: Always deploy with HTTPS in production
2. **Set SECRET_KEY**: Use a strong, random JWT secret key
3. **Configure CORS**: Update `allow_origins` in CORS middleware
4. **Enable Email**: Integrate SMTP for email verification
5. **Rate Limiting**: Add rate limiting to auth endpoints
6. **Token Storage**: Consider using httpOnly cookies for enhanced security

### Password Requirements

Current requirements:
- Minimum length: 3 characters (update to 8+ in production)
- No complexity requirements (add in production)

Update in `frontend/src/pages/Login.tsx`:

```typescript
if (password.length < 8) {
  setError('Password must be at least 8 characters long');
  return;
}
```

## Email Integration

The backend includes hooks for email notifications. To enable:

1. Configure SMTP settings:

```python
# In auth.py, add SMTP configuration
SMTP_HOST = os.getenv("SMTP_HOST", "smtp.gmail.com")
SMTP_PORT = int(os.getenv("SMTP_PORT", "587"))
SMTP_USER = os.getenv("SMTP_USER", "")
SMTP_PASSWORD = os.getenv("SMTP_PASSWORD", "")
```

2. Update hooks in `auth.py`:

```python
async def on_after_register(self, user: User, request: Optional[Request] = None):
    # Send verification email
    await send_verification_email(user.email, token)

async def on_after_forgot_password(self, user: User, token: str, request: Optional[Request] = None):
    # Send password reset email
    await send_reset_email(user.email, token)
```

## Testing

### Manual Testing

1. **Register a user:**
```bash
curl -X POST http://localhost:8001/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123"}'
```

2. **Login:**
```bash
curl -X POST http://localhost:8001/auth/jwt/login \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=test@example.com&password=test123"
```

3. **Access protected endpoint:**
```bash
TOKEN="your_token_here"
curl -X GET http://localhost:8001/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

### Frontend Testing

Tests are located in `frontend/src/tests/`:
```bash
cd frontend
npm run test
```

## Troubleshooting

### Common Issues

**Issue:** 401 Unauthorized on protected routes
- **Solution:** Check if token is stored in localStorage
- **Check:** Browser DevTools → Application → Local Storage → auth_token

**Issue:** CORS errors
- **Solution:** Verify CORS configuration in `main.py`
- **Check:** Browser DevTools → Console for CORS errors

**Issue:** Database locked
- **Solution:** SQLite doesn't support concurrent writes well
- **Fix:** Use connection pooling or PostgreSQL for production

**Issue:** Token expired
- **Solution:** Tokens expire after 1 hour
- **Fix:** Re-login or implement refresh token mechanism

## Future Enhancements

1. **Refresh Tokens**: Implement token refresh for longer sessions
2. **OAuth2**: Add social login (Google, GitHub, etc.)
3. **2FA**: Two-factor authentication with TOTP
4. **Session Management**: View and revoke active sessions
5. **Password Complexity**: Add complexity requirements
6. **Rate Limiting**: Prevent brute force attacks
7. **Remember Me**: Optional longer session duration

## Support

For issues or questions:
- Check the API documentation: http://localhost:8001/docs
- Review FastAPI Users docs: https://fastapi-users.github.io/fastapi-users/
- Check application logs for error details
