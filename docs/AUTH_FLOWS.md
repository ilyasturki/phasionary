# Phasionary - Authentication Flows

## Overview

Phasionary uses **better-auth** for authentication with email/password credentials. Authentication state is managed through secure HTTP-only session cookies, providing protection against XSS attacks while maintaining seamless user experience across page loads and SSR.

### Key Concepts

| Concept | Description |
|---------|-------------|
| **Session** | Server-managed authentication state tied to a secure cookie |
| **HTTP-only Cookie** | Session token stored in browser, inaccessible to JavaScript |
| **Authentication State** | Reactive client-side representation of user's logged-in status |

---

## 1. Sign Up Flow

New user registration with email and password.

### Flow Sequence

```
User                    App                     better-auth              Database
  │                      │                           │                       │
  ├─ Enter email ───────►│                           │                       │
  ├─ Enter password ────►│                           │                       │
  ├─ Enter name ────────►│                           │                       │
  ├─ Submit form ───────►│                           │                       │
  │                      ├─ Validate inputs ────────►│                       │
  │                      │                           ├─ Hash password ──────►│
  │                      │                           ├─ Create user record ─►│
  │                      │                           ├─ Create session ─────►│
  │                      │◄── Set session cookie ───┤                       │
  │◄─ Redirect to app ──┤                           │                       │
```

### Steps

| Step | Actor | Action |
|------|-------|--------|
| 1 | User | Navigates to sign-up page |
| 2 | User | Enters email address |
| 3 | User | Enters password (min 8 characters) |
| 4 | User | Enters display name |
| 5 | User | Submits registration form |
| 6 | App | Validates input format (email, password length) |
| 7 | better-auth | Checks if email already exists |
| 8 | better-auth | Hashes password with bcrypt |
| 9 | better-auth | Creates user record in database |
| 10 | better-auth | Creates session record |
| 11 | better-auth | Sets HTTP-only session cookie |
| 12 | App | Creates default project for user |
| 13 | App | Creates default categories in project |
| 14 | App | Redirects user to dashboard |

### Post-Registration State

- User is automatically signed in (no separate login required)
- Session cookie is set in browser
- Default project "My Project" is created
- Default categories (Feature, Fix, Ergonomy, Documentation, Research) are created
- User lands on empty task list in Current section

---

## 2. Sign In Flow

Existing user authentication with email and password.

### Flow Sequence

```
User                    App                     better-auth              Database
  │                      │                           │                       │
  ├─ Enter email ───────►│                           │                       │
  ├─ Enter password ────►│                           │                       │
  ├─ Submit form ───────►│                           │                       │
  │                      ├─ Send credentials ───────►│                       │
  │                      │                           ├─ Find user by email ─►│
  │                      │                           │◄─ User record ────────┤
  │                      │                           ├─ Verify password ────►│
  │                      │                           ├─ Create session ─────►│
  │                      │◄── Set session cookie ───┤                       │
  │◄─ Redirect to app ──┤                           │                       │
```

### Steps

| Step | Actor | Action |
|------|-------|--------|
| 1 | User | Navigates to sign-in page |
| 2 | User | Enters email address |
| 3 | User | Enters password |
| 4 | User | Optionally selects "Remember me" |
| 5 | User | Submits login form |
| 6 | App | Sends credentials to better-auth |
| 7 | better-auth | Looks up user by email |
| 8 | better-auth | Verifies password against stored hash |
| 9 | better-auth | Creates new session record |
| 10 | better-auth | Sets HTTP-only session cookie |
| 11 | App | Redirects to last visited page or dashboard |

### Remember Me Behavior

| Setting | Session Duration |
|---------|------------------|
| **Enabled** | Extended session (persists across browser restarts) |
| **Disabled** | Session expires when browser closes |

---

## 3. Sign Out Flow

User-initiated session termination.

### Flow Sequence

```
User                    App                     better-auth              Database
  │                      │                           │                       │
  ├─ Click sign out ────►│                           │                       │
  │                      ├─ Request sign out ───────►│                       │
  │                      │                           ├─ Delete session ─────►│
  │                      │◄── Clear session cookie ─┤                       │
  │◄─ Redirect to login ┤                           │                       │
```

### Steps

| Step | Actor | Action |
|------|-------|--------|
| 1 | User | Clicks "Sign Out" button |
| 2 | App | Sends sign-out request to better-auth |
| 3 | better-auth | Invalidates session in database |
| 4 | better-auth | Clears session cookie from browser |
| 5 | App | Clears local authentication state |
| 6 | App | Redirects user to sign-in page |

### Post-Sign-Out State

- Session cookie is removed
- All protected routes become inaccessible
- User must sign in again to access the app

---

## 4. Session Management

How authentication state persists and is validated.

### Session Lifecycle

| Phase | Description |
|-------|-------------|
| **Creation** | Session record created in database, cookie set in browser |
| **Validation** | Each request validates session cookie against database |
| **Refresh** | Session may be extended on activity (configurable) |
| **Expiration** | Session becomes invalid after timeout period |
| **Termination** | Explicit sign-out or manual invalidation |

### Session Validation Flow

```
Browser                 App                     better-auth              Database
  │                      │                           │                       │
  ├─ Request + cookie ──►│                           │                       │
  │                      ├─ Validate session ───────►│                       │
  │                      │                           ├─ Lookup session ─────►│
  │                      │                           │◄─ Session data ───────┤
  │                      │◄── User data ────────────┤                       │
  │◄─ Response ─────────┤                           │                       │
```

### SSR Session Handling (Nuxt-specific)

| Context | Behavior |
|---------|----------|
| **Server-side render** | Session fetched from cookie in request headers |
| **Client hydration** | Session state transferred from server to client |
| **Client navigation** | Session checked reactively via auth client |

### Session Expiration

| Event | Behavior |
|-------|----------|
| **Session expires** | Next request returns unauthorized |
| **App detects expiry** | User redirected to sign-in page |
| **User signs in** | New session created, previous data accessible |

---

## 5. Route Protection

How the app controls access to authenticated and public routes.

### Route Classification

| Route Type | Examples | Access |
|------------|----------|--------|
| **Public** | `/`, `/login`, `/signup` | Anyone |
| **Protected** | `/dashboard`, `/projects/*`, `/tasks/*` | Authenticated users only |

### Protection Flow

```
User                    Middleware               App
  │                         │                     │
  ├─ Navigate to route ────►│                     │
  │                         ├─ Check session ────►│
  │                         │◄── Session status ──┤
  │                         ├─ Route protected? ─►│
  │                         │                     │
  │  [If authenticated]     │                     │
  │◄─ Allow access ────────┤                     │
  │                         │                     │
  │  [If not authenticated] │                     │
  │◄─ Redirect to login ───┤                     │
```

### Middleware Behavior

| Scenario | Protected Route | Public Route |
|----------|-----------------|--------------|
| **User authenticated** | Allow access | Allow access |
| **User not authenticated** | Redirect to `/login` | Allow access |
| **Session expired** | Redirect to `/login` | Allow access |

### Redirect After Login

| Scenario | Redirect Target |
|----------|-----------------|
| **Direct login** | Dashboard (default) |
| **Redirect from protected route** | Original requested route |
| **Callback URL provided** | Specified callback URL |

---

## 6. Error Handling

Common authentication errors and their handling.

### Sign Up Errors

| Error | Cause | User Feedback |
|-------|-------|---------------|
| **Email already exists** | Account with email already registered | "An account with this email already exists" |
| **Invalid email format** | Email does not match expected pattern | "Please enter a valid email address" |
| **Password too short** | Password under minimum length | "Password must be at least 8 characters" |
| **Network error** | Connection failed | "Unable to connect. Please try again." |

### Sign In Errors

| Error | Cause | User Feedback |
|-------|-------|---------------|
| **Invalid credentials** | Wrong email or password | "Invalid email or password" |
| **Account not found** | No account with provided email | "Invalid email or password" |
| **Network error** | Connection failed | "Unable to connect. Please try again." |

### Session Errors

| Error | Cause | Behavior |
|-------|-------|----------|
| **Session expired** | Timeout reached | Redirect to sign-in with message |
| **Session invalid** | Session deleted or corrupted | Redirect to sign-in |
| **Cookie missing** | Browser cleared cookies | Redirect to sign-in |

### Security Notes

- Error messages for invalid credentials are intentionally vague to prevent user enumeration
- Failed login attempts should be rate-limited (implementation detail)
- Session tokens are never exposed to client-side JavaScript

---

## Summary

| Flow | Trigger | Result |
|------|---------|--------|
| **Sign Up** | User submits registration form | Account created, auto signed in, redirected to dashboard |
| **Sign In** | User submits login form | Session created, redirected to app |
| **Sign Out** | User clicks sign out | Session destroyed, redirected to login |
| **Session Check** | Any protected route access | Allow or redirect based on session validity |
| **Route Protection** | Navigation to protected route | Middleware validates session before allowing access |
