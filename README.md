# DentalDesk AI

AI call answering for dental clinics. This repo starts the product as a multi-service SaaS with a Next.js frontend, modular Go API, separate Go worker, and Postgres migration files.

## Apps

- `apps/web`: Next.js dashboard and onboarding UI
- `apps/api`: Go API with first-party auth, tenant isolation, practice APIs, and initial voice tool routes
- `apps/worker`: Go worker skeleton for background jobs
- `infra/migrations`: SQL schema migrations
- `docs`: architecture notes

## Local Setup

Install web dependencies:

```powershell
npm.cmd install
```

Run the API:

```powershell
go run -buildvcs=false ./apps/api/cmd/api
```

Run the worker:

```powershell
go run -buildvcs=false ./apps/worker/cmd/worker
```

Run the web app:

```powershell
npm.cmd run dev:web
```

## Local Postgres

Start Postgres:

```powershell
docker compose up -d postgres
```

Apply migrations:

```powershell
go run -buildvcs=false ./apps/api/cmd/migrate
```

Run the API against Postgres:

```powershell
$env:STORE_DRIVER='postgres'
$env:DATABASE_URL='postgres://postgres:postgres@localhost:5432/dentaldesk?sslmode=disable'
go run -buildvcs=false ./apps/api/cmd/api
```

The web app expects the API at `http://localhost:8080` by default. Override it with:

```text
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

The API currently defaults to the in-memory store:

```text
STORE_DRIVER=memory
```

The Postgres driver and SQL repository methods are available behind:

```text
STORE_DRIVER=postgres
DATABASE_URL=postgres://postgres:postgres@localhost:5432/dentaldesk?sslmode=disable
```

Run migrations before starting the API with `STORE_DRIVER=postgres`.

## Verification

```powershell
go test ./...
go build -buildvcs=false ./...
npm.cmd run build:web
```

## Production Container Build

The repo includes Dockerfiles for the API, worker, and web app plus a production compose file:

```powershell
docker compose -f docker-compose.production.yml build
docker compose -f docker-compose.production.yml up -d
```

Required production environment values include:

```text
APP_BASE_URL
API_BASE_URL
NEXT_PUBLIC_API_BASE_URL
DATABASE_URL
POSTGRES_PASSWORD
SESSION_SECRET
CALENDAR_TOKEN_SECRET
VAPI_WEBHOOK_SECRET
WORKER_TOKEN
```

Optional integrations:

```text
GOOGLE_CLIENT_ID
GOOGLE_CLIENT_SECRET
GOOGLE_REDIRECT_URL
STRIPE_SECRET_KEY
STRIPE_WEBHOOK_SECRET
STRIPE_PRICE_ID
SMTP_*
```

## Render Free Deploy

This repo includes `render.yaml` for a Render Blueprint deploy:

- `dentaldesk-api`: Docker web service
- `dentaldesk-web`: Docker web service
- `dentaldesk-db`: Render Postgres

Use Render Dashboard -> New -> Blueprint, connect the GitHub repo, and select `render.yaml`.

The blueprint assumes these Render hostnames:

```text
https://dentaldesk-api.onrender.com
https://dentaldesk-web.onrender.com
```

If Render assigns different hostnames, update these environment variables in Render and redeploy:

```text
APP_BASE_URL
API_BASE_URL
GOOGLE_REDIRECT_URL
NEXT_PUBLIC_API_BASE_URL
```

For Vapi/Retell, use the deployed API URL:

```text
Base URL: https://dentaldesk-api.onrender.com
Header: X-DentalDesk-Webhook-Secret: <Render VAPI_WEBHOOK_SECRET>
Bootstrap: POST /v1/voice/bootstrap
Tools:
POST /v1/voice/practice-info
POST /v1/voice/appointment-request
POST /v1/voice/call-summary
```

Render free tier is for testing only: free web services can spin down after idle time, and free Postgres expires after 30 days.

## Current State

Implemented:

- First-party auth skeleton with secure session cookies
- Practice creation and listing
- Practice employee invite flow with accept-invite password setup
- Gmail/SMTP invite email delivery with dev-log fallback
- Role, permission, member, and location admin scaffolding
- Dental onboarding form and assistant configuration editing
- Voice assistant bootstrap endpoint with dental prompt and tool schemas
- Per-practice voice provider configuration for Vapi, Retell, or custom providers
- Local voice test harness that creates appointment requests and call summaries
- Staff appointment request queue with status workflow, internal notes, and dashboard filters
- Calendar configuration for request-only, booking-link, and Google Calendar-ready scheduling modes
- Google Calendar OAuth connection and event creation endpoints with encrypted token storage
- Appointment queue scheduling action that creates Google Calendar events and marks requests scheduled
- Manual billing subscription records for pilot pricing, minute limits, overage rate, and future Stripe IDs
- Stripe Checkout session creation and signed webhook handling for completed checkout sessions
- Docker production build files and GitHub Actions CI
- Staff email notifications for appointment requests and call summaries
- Audit logs and dashboard activity feed
- Auth hardening: password policy, rate limiting, CSRF protection, email verification, and password reset
- Tenant-scoped assistant config, call summaries, and appointment request reads
- Voice webhook/tool endpoints protected by a shared secret
- Modular Go service structure
- Next.js dashboard shell
- Worker skeleton
- Initial Postgres schema migration

Next implementation steps:

- Add migration runner instead of applying SQL manually
- Add production email templates and retry tracking
- Connect a live Vapi or Retell assistant and test real phone calls
- Add Resend/Twilio notification providers
- Add Stripe customer portal and subscription update/cancel webhook sync
- Add host-specific deploy target after choosing infrastructure
- Add automatic appointment duration suggestions and availability lookup
