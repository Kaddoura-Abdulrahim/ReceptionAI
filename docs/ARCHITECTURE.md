# Architecture

DentalDesk AI starts as a multi-service SaaS with modular Go services.

## Deployable Units

- `apps/web`: Next.js frontend, landing page, onboarding UI, dashboard
- `apps/api`: Go API with first-party auth, tenant isolation, dashboard APIs, and initial voice routes
- `apps/worker`: Go worker for async summaries, notifications, cleanup, and billing sync
- `infra/migrations`: SQL migrations for managed Postgres

Container build files:

```text
apps/web/Dockerfile
apps/api/Dockerfile
apps/worker/Dockerfile
docker-compose.production.yml
.github/workflows/ci.yml
```

The API is a modular monolith first. Internal services are separated in code and can become separate deployables later when operational pressure justifies it.

## Persistence

The API uses a `store.Store` interface. `MemoryStore` is the zero-dependency development store. `PostgresStore` is the production persistence target and uses the initial schema in `infra/migrations/001_initial.sql`.

Store selection:

```text
STORE_DRIVER=memory
STORE_DRIVER=postgres
```

Run migrations before using `STORE_DRIVER=postgres`.

## Boundary Rules

- Every tenant-owned record is scoped by `practice_id`.
- The frontend is never trusted for authorization.
- PHI should not go into analytics, general logs, or error traces.
- External providers authenticate at the boundary with their native signatures or tokens.
- Internal authorization normalizes requests into an actor, role, scopes, and practice.

## Email

Invite emails are sent through the notification service. The current provider is SMTP-compatible and can use Gmail app-password credentials through:

```text
SMTP_HOST
SMTP_PORT
SMTP_USER
SMTP_PASS
SMTP_SECURE
SMTP_FROM_EMAIL
SMTP_FROM_NAME
```

When SMTP credentials are missing, the service logs email intent instead of failing local development.

## Voice Agent Wiring

Voice providers should call the API with:

```text
X-DentalDesk-Webhook-Secret: <VAPI_WEBHOOK_SECRET>
```

Bootstrap the assistant for a practice:

```http
POST /v1/voice/bootstrap
Content-Type: application/json

{
  "practiceId": "practice-id"
}
```

The response includes:

- `systemPrompt`: dental receptionist instructions generated from assistant config
- `firstMessage`: practice greeting
- `voiceTone`: configured tone
- `toolEndpoints`: provider-neutral tool definitions

Tool endpoints:

```text
POST /v1/voice/practice-info
POST /v1/voice/appointment-request
POST /v1/voice/call-summary
```

The assistant must stay administrative-only. It should create appointment requests, take messages, and save summaries, but it must not diagnose, recommend treatment, give medication advice, or promise insurance coverage.

Per-practice provider configuration is managed through:

```text
GET   /v1/practices/{practiceID}/voice-provider
PATCH /v1/practices/{practiceID}/voice-provider
```

Stored provider fields:

- provider: `vapi`, `retell`, or `custom`
- phone number
- assistant ID
- webhook status
- last webhook timestamp

## Calendar

Calendar configuration is tenant-scoped and intentionally separate from assistant configuration. The MVP supports three scheduling modes:

```text
request_only  -> AI captures the request and staff confirms manually
booking_link  -> AI can offer a configured booking link and still creates a staff-visible request
google        -> stores Google Calendar identifiers and is ready for OAuth/direct booking work
```

Calendar config endpoints:

```text
GET   /v1/practices/{practiceID}/calendar-config
PATCH /v1/practices/{practiceID}/calendar-config
POST  /v1/practices/{practiceID}/calendar/oauth/start
GET   /v1/calendar/oauth/callback
POST  /v1/practices/{practiceID}/calendar/events
```

The calendar status is derived from the mode:

```text
not_configured
needs_booking_url
needs_calendar_id
ready_for_oauth
connected
configured
```

Google OAuth tokens are encrypted before storage using `CALENDAR_TOKEN_SECRET`. The dashboard only receives connection status and token expiry metadata.

Appointment queue scheduling:

- Staff enter start/end times on a request.
- The dashboard creates a Google Calendar event through the API.
- The appointment request is marked `scheduled` after event creation succeeds.
- Event descriptions include caller phone, preferred time, insurance, caller notes, and staff notes.

## Billing

Billing starts as a tenant-scoped subscription record so pilots can be managed manually before Stripe is connected.

Billing endpoints:

```text
GET   /v1/practices/{practiceID}/billing
PATCH /v1/practices/{practiceID}/billing
POST  /v1/practices/{practiceID}/billing/checkout-session
POST  /v1/billing/stripe/webhook
```

Tracked billing fields:

```text
plan: pilot | starter | growth | custom
status: manual | trialing | active | past_due | canceled
included minutes
overage cents per minute
optional Stripe customer/subscription IDs
```

Stripe Checkout is created server-side using `STRIPE_SECRET_KEY` and `STRIPE_PRICE_ID`. Stripe webhook signatures are verified with `STRIPE_WEBHOOK_SECRET` when configured. Completed checkout sessions write customer and subscription IDs into the same billing subscription record.

Local test harness:

```text
POST /v1/practices/{practiceID}/voice-test-call
```

This endpoint is authenticated as a dashboard action and does not use the provider webhook secret. It simulates a completed voice call by creating:

- a `call_sessions` record
- an `appointment_requests` record
- a `call_summaries` record

Use it to verify the end-to-end dashboard flow before connecting a live provider.

Staff notifications:

- Appointment requests email the configured assistant `notificationEmail`.
- Call summaries email the configured assistant `notificationEmail`.
- If no notification email is configured, notification delivery is skipped and logged.
- In local development without SMTP credentials, delivery falls back to logs.

## Appointment Workflow

Appointment requests are saved as staff-review tasks instead of confirmed bookings. Staff can filter the dashboard queue, add internal notes, and move each request through:

```text
new -> contacted -> scheduled -> closed
```

Requests can also be marked `spam`. Updates are tenant-scoped and require `appointment:update`.

Dashboard/API endpoints:

```text
GET   /v1/practices/{practiceID}/appointment-requests
PATCH /v1/practices/{practiceID}/appointment-requests/{requestID}
```

## Activity Feed

Practice activity is stored in `audit_logs` and exposed through:

```text
GET /v1/practices/{practiceID}/activity
```

Tracked events include:

- practice creation
- assistant config updates
- voice provider updates
- calendar config updates
- Google Calendar OAuth connections
- Google Calendar event creation
- billing updates
- local voice test calls
- location creation
- role creation
- employee invites
- member role updates
- member disable actions
- appointment request updates
- voice-created appointment requests
- voice-created call summaries

## Auth Hardening

First-party auth includes:

- Secure session cookies, with `Secure` enabled in production
- CSRF token endpoint at `GET /v1/csrf`
- `X-CSRF-Token` enforcement for authenticated write requests
- Login/register/password-reset rate limiting
- Password policy: at least 12 characters with uppercase, lowercase, and a number
- Email verification tokens
- Password reset tokens

Auth endpoints:

```text
POST /v1/auth/register
POST /v1/auth/login
POST /v1/auth/logout
POST /v1/auth/verify-email
POST /v1/auth/request-password-reset
POST /v1/auth/reset-password
```

Live provider setup checklist:

1. Configure the practice in DentalDesk AI.
2. Save provider type, phone number, and assistant ID.
3. Set the provider webhook/tool header to `X-DentalDesk-Webhook-Secret`.
4. Set the provider secret value to `VAPI_WEBHOOK_SECRET`.
5. Bootstrap the assistant prompt from `POST /v1/voice/bootstrap`.
6. Register provider tools for practice info, appointment request, and call summary.
7. Place a test call and confirm the dashboard receives appointment requests and call summaries.
