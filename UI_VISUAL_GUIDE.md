# Authentication Implementation - Visual Guide

## UI Screenshots & Descriptions

### 1. Login Page (`/login`)

**Location:** `http://localhost:3000/login`

**Visual Description:**
```
┌─────────────────────────────────────────────────────┐
│  V-Insight                          [Login]         │
├─────────────────────────────────────────────────────┤
│                                                      │
│              ┌─────────────────────────┐            │
│              │                         │            │
│              │  Login                  │            │
│              │  Sign in to access your │            │
│              │  monitoring dashboard   │            │
│              │                         │            │
│              │  [Email Address]        │            │
│              │  you@example.com        │            │
│              │                         │            │
│              │  [Password]             │            │
│              │  ••••••••               │            │
│              │                         │            │
│              │  [     Sign In     ]    │            │
│              │                         │            │
│              │  Don't have an account? │            │
│              │  Sign up                │            │
│              │                         │            │
│              └─────────────────────────┘            │
│                                                      │
└─────────────────────────────────────────────────────┘
```

**Features:**
- Clean, centered form design
- Email and password input fields
- "Sign In" button (blue, full width)
- Loading state ("Signing In...")
- Error message display (red alert box)
- Link to registration page
- Responsive design

**Form Fields:**
- Email: `type="email"`, required
- Password: `type="password"`, required

**Behavior:**
- Submit → POST to `/api/v1/auth/login`
- Success → Store JWT → Redirect to `/dashboard`
- Error → Display error message

---

### 2. Registration Page (`/register`)

**Location:** `http://localhost:3000/register`

**Visual Description:**
```
┌─────────────────────────────────────────────────────┐
│  V-Insight                          [Login]         │
├─────────────────────────────────────────────────────┤
│                                                      │
│              ┌─────────────────────────┐            │
│              │                         │            │
│              │  Create Account         │            │
│              │  Sign up to start       │            │
│              │  monitoring your domains│            │
│              │                         │            │
│              │  [Email Address]        │            │
│              │  you@example.com        │            │
│              │                         │            │
│              │  [Organization Name]    │            │
│              │  Your Company           │            │
│              │                         │            │
│              │  [Password]             │            │
│              │  ••••••••               │            │
│              │  Minimum 6 characters   │            │
│              │                         │            │
│              │  [Confirm Password]     │            │
│              │  ••••••••               │            │
│              │                         │            │
│              │  [     Sign Up     ]    │            │
│              │                         │            │
│              │  Already have account?  │            │
│              │  Sign in                │            │
│              │                         │            │
│              └─────────────────────────┘            │
│                                                      │
└─────────────────────────────────────────────────────┘
```

**Features:**
- Clean, centered form design
- Four input fields (email, org name, password, confirm)
- "Sign Up" button (blue, full width)
- Loading state ("Creating Account...")
- Error message display (red alert box)
- Password validation hint
- Link to login page
- Responsive design

**Form Fields:**
- Email: `type="email"`, required
- Organization Name: `type="text"`, required
- Password: `type="password"`, required, minlength="6"
- Confirm Password: `type="password"`, required, minlength="6"

**Validation:**
- Passwords must match
- Password minimum 6 characters
- Email must be valid format

**Behavior:**
- Submit → Validate → POST to `/api/v1/auth/register`
- Success → Store JWT → Redirect to `/dashboard`
- Error → Display error message

---

### 3. Dashboard Page (Protected)

**Location:** `http://localhost:3000/dashboard`

**Visual Description (After Login):**
```
┌─────────────────────────────────────────────────────┐
│  V-Insight  Home Dashboard Domains Alerts [Logout] │
├─────────────────────────────────────────────────────┤
│                                                      │
│  Dashboard                                          │
│  Monitor your domains and view system metrics      │
│                                                      │
│  ┌────────────┐ ┌────────────┐ ┌────────────┐     │
│  │ Total      │ │ Active     │ │ Uptime     │     │
│  │ Domains    │ │ Alerts     │ │            │     │
│  │     0      │ │     0      │ │    0%      │     │
│  └────────────┘ └────────────┘ └────────────┘     │
│                                                      │
│  ┌──────────────────────────────────────────┐      │
│  │ Recent Activity                          │      │
│  │                                          │      │
│  │ No recent activity to display            │      │
│  │                                          │      │
│  └──────────────────────────────────────────┘      │
│                                                      │
└─────────────────────────────────────────────────────┘
```

**Navigation (Authenticated):**
- Home
- Dashboard
- Domains
- Alerts
- Settings
- Logout button

**Protection:**
- Requires authentication
- Redirects to `/login` if not logged in

---

### 4. Navigation States

**Unauthenticated Navigation:**
```
┌─────────────────────────────────────┐
│  V-Insight           [Login]        │
└─────────────────────────────────────┘
```

**Authenticated Navigation:**
```
┌──────────────────────────────────────────────────────────┐
│  V-Insight  [Home] [Dashboard] [Domains] [Alerts]       │
│             [Settings] [Logout]                          │
└──────────────────────────────────────────────────────────┘
```

**Active State:**
- Current page has darker blue background
- Hover states with lighter blue

---

## User Flows

### Registration Flow

```
1. User visits /register
   ↓
2. Fills form:
   - Email: test@example.com
   - Organization: Test Company
   - Password: password123
   - Confirm: password123
   ↓
3. Clicks "Sign Up"
   ↓
4. Frontend validates:
   - Passwords match? ✓
   - Password >= 6 chars? ✓
   ↓
5. POST /api/v1/auth/register
   ↓
6. Backend creates user & tenant
   ↓
7. Backend returns JWT token
   ↓
8. Frontend:
   - Stores token in localStorage
   - Calls /api/v1/auth/me
   - Fetches user data
   - Updates authStore
   ↓
9. Redirects to /dashboard
   ↓
10. Navigation shows authenticated menu
```

### Login Flow

```
1. User visits /login
   ↓
2. Enters credentials:
   - Email: test@example.com
   - Password: password123
   ↓
3. Clicks "Sign In"
   ↓
4. POST /api/v1/auth/login
   ↓
5. Backend validates credentials
   ↓
6. Backend returns JWT token
   ↓
7. Frontend:
   - Stores token in localStorage
   - Calls /api/v1/auth/me
   - Fetches user data
   - Updates authStore
   ↓
8. Redirects to /dashboard
   ↓
9. Navigation shows authenticated menu
```

### Session Persistence

```
1. User logged in previously
   ↓
2. User opens app / refreshes page
   ↓
3. +layout.svelte runs checkAuth()
   ↓
4. Retrieves token from localStorage
   ↓
5. GET /api/v1/auth/me with token
   ↓
6. Backend validates token
   ↓
7. If valid:
   - Returns user data
   - Updates authStore
   - User stays on current page
   ↓
8. If invalid:
   - Clears token
   - Redirects to /login (if protected route)
```

### Logout Flow

```
1. User clicks "Logout" in nav
   ↓
2. authStore.logout() called
   ↓
3. Removes token from localStorage
   ↓
4. Resets authStore state
   ↓
5. Redirects to /login
   ↓
6. Navigation shows public menu
```

### Protected Route Access

```
1. Unauthenticated user visits /dashboard
   ↓
2. +layout.svelte reactive statement runs:
   $: if (browser && !$authStore.isAuthenticated && !isPublicRoute($page.url.pathname))
   ↓
3. Condition is true
   ↓
4. window.location.href = '/login'
   ↓
5. User redirected to login page
```

---

## Color Scheme

**Primary Colors:**
- Blue: `bg-blue-600` (#2563EB) - Primary buttons, navigation
- Blue Hover: `bg-blue-700` (#1D4ED8)
- Blue Light: `bg-blue-500` (#3B82F6)

**Background Colors:**
- Page Background: `bg-gray-50` (#F9FAFB)
- Card Background: `bg-white` (#FFFFFF)

**Text Colors:**
- Primary: `text-gray-900` (#111827)
- Secondary: `text-gray-600` (#4B5563)
- Muted: `text-gray-500` (#6B7280)

**Alert Colors:**
- Error Background: `bg-red-100` (#FEE2E2)
- Error Border: `border-red-400` (#F87171)
- Error Text: `text-red-700` (#B91C1C)

**Form Elements:**
- Border: `border-gray-300` (#D1D5DB)
- Focus Ring: `ring-blue-500` (#3B82F6)

---

## Responsive Design

### Desktop (≥768px)
- Forms: `max-w-md` (28rem / 448px wide)
- Centered horizontally
- Full padding and spacing

### Mobile (<768px)
- Forms: Full width with padding
- Touch-friendly buttons
- Stacked inputs
- Readable font sizes

### Container
- `container mx-auto px-4` - Responsive container with padding

---

## Accessibility

**Keyboard Navigation:**
- Tab through form fields
- Enter to submit forms
- Focus states visible

**Screen Readers:**
- Semantic HTML elements
- Label associations
- Error announcements

**Form Labels:**
- All inputs have associated labels
- `for` attribute matches input `id`
- Clear, descriptive text

**Button States:**
- Disabled state during loading
- Visual feedback on hover
- Clear active states

---

## Design Principles

1. **Minimalist:** Clean, uncluttered interface
2. **Centered:** Forms centered on screen
3. **Responsive:** Works on all device sizes
4. **Consistent:** Same patterns throughout
5. **Accessible:** WCAG compliant
6. **Professional:** Modern, polished look
7. **User-Friendly:** Clear feedback and states

---

## Technical Implementation

**Form Handling:**
- Svelte's two-way binding (`bind:value`)
- `on:submit|preventDefault` for form submission
- Async/await for API calls
- Loading states during submission
- Error state management

**Styling:**
- Tailwind CSS utility classes
- Consistent spacing scale
- Responsive breakpoints
- Hover and focus states
- Shadow and rounded corners

**State Management:**
- Svelte stores for auth state
- Reactive statements (`$:`)
- Component-level state for forms
- Global navigation state

---

## Files Structure

```
frontend/src/
├── routes/
│   ├── +layout.svelte          # Global layout with Nav and route protection
│   ├── +page.svelte            # Home page (public)
│   ├── login/
│   │   └── +page.svelte        # Login page
│   ├── register/
│   │   └── +page.svelte        # Registration page
│   └── dashboard/
│       └── +page.svelte        # Dashboard (protected)
├── lib/
│   ├── stores/
│   │   └── auth.ts             # Auth state management
│   ├── api/
│   │   └── client.ts           # API client with auto auth
│   └── components/
│       └── Nav.svelte          # Navigation component
└── app.css                     # Tailwind CSS imports
```

---

This implementation provides a complete, production-ready authentication system with modern UI/UX, comprehensive error handling, and excellent developer experience.
