# AI Medical Front Desk

## 1. Product Direction

AI Medical Front Desk is an AI receptionist for healthcare and wellness practices. It answers inbound calls, handles routine administrative conversations, captures patient requests, books or routes appointments, and sends clean summaries to staff.

The product is not an AI doctor, therapist, nurse, or clinical triage system. It is an administrative front-desk assistant with strong escalation rules.

Initial public positioning should focus on dental clinics. The underlying platform should stay flexible enough to support other healthcare and wellness practices later.

## 2. Working Brand

The platform can use a broad internal name while the first market-facing product is dental-specific.

Working options:

- FrontDesk AI
- ClinicDesk AI
- DentalDesk AI
- AfterHours AI
- PracticeLine
- CallCare AI

Recommended first public brand:

> DentalDesk AI

This makes the first offer immediately understandable to dental clinics while the backend remains general enough for future verticals.

## 3. Simple Pitch

> Never miss a patient call again. DentalDesk AI answers, schedules, takes messages, and escalates urgent calls while your staff focuses on care.

Shorter version:

> AI call answering for dental clinics.

## 4. First Customer Segment

Start with dental clinics.

Why dental first:

- Dental clinics are appointment-heavy.
- Missed calls directly cost new patient bookings.
- Front desks handle many repetitive questions.
- Scheduling, rescheduling, cancellations, insurance questions, and new patient intake are common.
- The workflow is easier to understand than primary care or urgent care.
- It has medical-adjacent privacy needs without requiring the AI to perform clinical triage.

Initial target customers:

- Single-location dental offices
- Small dental groups with 2-5 locations
- Dental office managers
- Dental practice owners

Later sales channels:

- Dental marketing agencies
- Dental consultants
- Outsourced receptionist companies
- Practice management software partners

## 5. First Offer

Start with a simple, easy-to-understand package.

Example offer:

```text
DentalDesk AI
AI call answering for dental clinics.

$299/month
Includes 300 call minutes
$0.25/min after included minutes
Setup fee waived for the first 5 clinics
```

Do not offer unlimited calls early. Voice costs are usage-based, and unlimited pricing can create margin problems.

## 6. Core Workflows

Initial dental workflows:

- New patient appointment request
- Existing patient message taking
- Reschedule request
- Cancellation request
- Office hours and location questions
- Accepted insurance question
- Cleaning, exam, whitening, emergency dental visit, and consultation requests
- After-hours call handling
- Staff handoff for urgent pain, swelling, trauma, bleeding, caller distress, or AI uncertainty

The first MVP should favor appointment requests over confirmed booking unless the practice has a simple calendar workflow.

## 7. MVP Scope

The first sellable version should be narrow and reliable.

Core MVP:

- Dental-specific AI phone receptionist
- Call forwarding support
- New AI phone number option
- New patient appointment request flow
- Existing patient message-taking flow
- Reschedule and cancellation request flow
- Basic insurance information capture
- Office hours, location, and parking answers
- Emergency dental escalation rules
- Staff notification by email or SMS
- Call summaries
- Basic call transcript access if vendor supports it
- Practice onboarding form
- Per-practice assistant configuration
- Basic admin dashboard
- Basic billing

Implemented onboarding fields:

- Greeting
- Staff handoff phone
- Staff notification email
- Office hours
- Services
- Accepted insurance
- New patient rules
- Emergency escalation rules
- Cancellation policy
- Intake form link
- Voice tone

Avoid in v1:

- EHR integration
- Clinical advice
- Diagnosis or treatment recommendations
- Confirmed appointment booking if the practice workflow is not simple
- Complex insurance verification
- Prescription questions
- Multi-location enterprise routing
- On-premise deployments

## 8. Safety Boundaries

The AI should not:

- Diagnose conditions
- Recommend treatment
- Provide therapy
- Give medication advice
- Interpret symptoms clinically
- Decide whether someone needs urgent care
- Promise insurance coverage
- Discuss sensitive records unless the practice has authorized the workflow

The AI should:

- Help with scheduling and administrative questions
- Collect information for staff
- Use disclaimers when callers ask for medical or dental advice
- Escalate emergencies immediately
- Escalate uncertainty to staff
- Offer a human handoff when required

General medical boundary:

> I can help with scheduling and messages, but I cannot provide medical advice. If this is an emergency, call 911 or go to the nearest emergency room.

Dental-specific boundary:

> I can help schedule an appointment or take a message for the dental team, but I cannot provide dental advice. If you are experiencing severe pain, swelling, heavy bleeding, trauma, or trouble breathing, seek urgent medical care or call emergency services.

Mental health boundary for future templates:

> If you may harm yourself or someone else, call emergency services or a crisis line now. I can also take a message for the office, but I cannot provide crisis counseling.

## 9. Escalation Rules

The AI should escalate or hand off when the caller mentions:

- Severe dental pain
- Swelling
- Heavy bleeding
- Facial trauma
- Trouble breathing
- Post-surgery complication
- Medication question
- Prescription request
- Diagnosis request
- Treatment recommendation request
- Angry or distressed caller
- Billing dispute
- Request for a specific staff member
- Existing patient record details
- AI uncertainty
- Caller asks for a human

Escalation behavior should be configurable per practice:

- Warm transfer to staff
- Send urgent SMS to staff
- Take message and mark urgent
- Provide emergency disclaimer
- Provide after-hours instructions

## 10. Demo Script

The sales demo should use a real phone number and a simple dental call flow.

Example demo:

```text
Caller: Hi, are you accepting new patients?

AI: Yes, I can help with that. Are you looking for a cleaning, exam, emergency visit, or something else?

Caller: I need a cleaning.

AI: I can take your details and send an appointment request to the office. Can I get your name and the best phone number for a callback?

Caller: Sarah Ahmed, 555-0100.

AI: Thanks, Sarah. Do you prefer mornings, afternoons, or a specific day?

Caller: Weekday mornings.

AI: Got it. I will send this to the front desk and ask them to confirm a weekday morning appointment. Is there anything else you want them to know?
```

The demo should also show an escalation case:

```text
Caller: My face is swelling and I can barely breathe.

AI: I am sorry you are dealing with that. I cannot provide dental advice. Trouble breathing can be an emergency, so please call 911 or go to the nearest emergency room now. I can also notify the dental office that you called.
```

## 11. Onboarding Form

The onboarding form is the heart of the product. It should generate the practice-specific AI behavior.

Required fields:

- Practice name
- Specialty
- Main phone number
- Location address
- Office hours
- Holiday or after-hours instructions
- Services offered
- Appointment types
- New patient rules
- Existing patient rules
- Accepted insurance plans
- Payment policy notes
- Cancellation policy
- Emergency instructions
- Staff handoff phone number
- Staff notification email
- Staff notification SMS number
- Calendar link or booking process
- Intake form link
- Tone preferences
- Forbidden topics
- Recording preference
- Data retention preference

Dental-specific fields:

- New patient cleaning availability
- Emergency dental visit process
- Whitening or cosmetic consultation process
- Pediatric dentistry support
- Orthodontic consultation support
- Insurance plans commonly accepted
- Whether the office accepts walk-ins
- Whether after-hours emergencies are handled

## 12. Call Summary Format

Staff notifications should be short and useful.

Example summary:

```text
Caller: Sarah Ahmed
Phone: +1 555-0100
Caller type: New patient
Reason: Cleaning appointment request
Insurance: Aetna
Preferred time: Weekday mornings
Urgency: Routine
AI action: Appointment request captured
Follow-up needed: Call back to confirm a slot
Transcript: [link]
```

Urgent summary example:

```text
Caller: John Smith
Phone: +1 555-0130
Caller type: Existing patient
Reason: Facial swelling after dental procedure
Urgency: Urgent
AI action: Provided emergency disclaimer and notified staff
Follow-up needed: Staff review immediately
Transcript: [link]
```

## 13. Success Metrics

Track whether the product creates measurable value for the clinic.

Core metrics:

- Calls answered
- Missed calls recovered
- New patient appointment requests
- Existing patient messages captured
- Reschedule and cancellation requests handled
- Urgent escalations
- Human handoffs
- Average call duration
- Failed calls
- Caller sentiment
- Staff response time
- Estimated recovered revenue

Pilot success goal:

- 3 paying dental practices
- 100+ handled calls
- Measured missed-call reduction
- Measured appointment requests
- Clear list of unsafe or confusing call scenarios

## 14. Go-To-Market

Start with manual sales and a concierge onboarding process.

Early sales plan:

1. Build a working demo phone number.
2. Record sample dental calls.
3. Create a one-page landing page.
4. Contact local dental offices and small dental groups.
5. Offer a 14-day pilot.
6. Configure call forwarding for missed or after-hours calls.
7. Send a weekly report with calls answered, leads captured, and appointment requests.

The early product should feel done-for-you. Self-serve onboarding can come after repeated manual setups reveal the real workflow.

## 15. Compliance Considerations

For U.S. healthcare practices, the product will likely handle protected health information. Treat the product as HIPAA-sensitive from the beginning.

Important requirements:

- Use vendors that support HIPAA where protected health information is involved
- Sign Business Associate Agreements when required
- Encrypt data in transit and at rest
- Use access controls and role-based permissions
- Keep audit logs
- Limit data retention
- Avoid sending health data to advertising or tracking tools
- Do not use patient call data for model training unless explicitly allowed
- Handle call recording consent by state
- Provide a clear privacy policy and security posture
- Avoid protected health information in analytics, logs, and error traces

The first pilot can reduce risk by storing summaries by default and making recordings optional.

## 16. Technical Stack

Recommended MVP stack:

- Frontend: Next.js with TypeScript
- Backend API: Go
- Database: managed Postgres
- Auth: first-party auth in the Go backend
- Voice agent: Vapi or Retell
- Calendar: request-only first, then booking links, then Google Calendar OAuth/direct booking
- Notifications: email and SMS
- Email: Resend
- SMS: Twilio
- Billing: Stripe
- Background jobs: Inngest or Trigger.dev
- Hosting: Vercel
- Monitoring: Sentry

Fast MVP call flow:

```text
patient calls clinic
  -> clinic forwards missed or after-hours calls
  -> Vapi/Retell answers
  -> AI follows dental-specific practice config
  -> AI calls backend tools when needed
  -> backend saves call result
  -> staff receives summary
```

Tool-style actions the agent should use:

```text
getPracticeInfo()
createAppointmentRequest()
sendCallSummary()
```

Implemented voice bootstrap:

```text
POST /v1/voice/bootstrap
  -> returns dental system prompt
  -> returns first message
  -> returns voice tone
  -> returns provider-neutral tool endpoint schemas
```

Implemented voice provider config:

```text
provider: vapi | retell | custom
phone number
assistant ID
webhook status
last webhook timestamp
```

Implemented calendar config:

```text
GET   /v1/practices/{practiceID}/calendar-config
PATCH /v1/practices/{practiceID}/calendar-config
POST  /v1/practices/{practiceID}/calendar/oauth/start
GET   /v1/calendar/oauth/callback
POST  /v1/practices/{practiceID}/calendar/events
mode: request_only | booking_link | google
provider: none | calendly | google | custom
status: not_configured | needs_booking_url | needs_calendar_id | ready_for_oauth | connected | configured
Google tokens are encrypted at rest with CALENDAR_TOKEN_SECRET.
```

Implemented billing foundation:

```text
GET   /v1/practices/{practiceID}/billing
PATCH /v1/practices/{practiceID}/billing
POST  /v1/practices/{practiceID}/billing/checkout-session
POST  /v1/billing/stripe/webhook
plan: pilot | starter | growth | custom
status: manual | trialing | active | past_due | canceled
included minutes and overage cents per minute
Stripe customer/subscription IDs
```

Implemented local voice test harness:

```text
POST /v1/practices/{practiceID}/voice-test-call
  -> creates test call session
  -> creates appointment request
  -> creates call summary
```

Implemented appointment request workflow:

```text
GET   /v1/practices/{practiceID}/appointment-requests
PATCH /v1/practices/{practiceID}/appointment-requests/{requestID}
statuses: new, contacted, scheduled, closed, spam
staff note: internal front-desk note
dashboard filters: open, all, and per-status
```

Implemented staff notifications:

```text
appointment request -> notification email
call summary -> notification email
missing notification email -> log and skip
missing SMTP config -> dev-log fallback
```

Implemented activity feed:

```text
GET /v1/practices/{practiceID}/activity
  -> recent audit events
  -> dashboard activity panel
```

Implemented auth hardening:

```text
strong password policy
auth endpoint rate limiting
CSRF token for authenticated writes
production secure session cookies
email verification flow
password reset flow
```

Custom path later:

```text
phone call
  -> Twilio or SIP provider
  -> realtime voice model
  -> business rules and tools
  -> calendar / CRM / EHR / notification system
  -> transcript, summary, analytics
```

Start with a voice-agent platform to validate demand. Move custom only if margins, control, or compliance require it.

## 17. Service Architecture

Build the system as multiple deployable services so code changes do not require redeploying the entire product.

The MVP should still be simple, but it should not be a single monolith.

Recommended service split:

```text
web-app
  Next.js frontend, landing page, dashboard, onboarding UI

api-service
  Go backend API, auth, tenant isolation, practice config, dashboard APIs

voice-webhook-service
  Go service or separate Go binary for Vapi/Retell tool calls and call webhooks

worker-service
  Background jobs, summaries, notifications, billing sync, cleanup tasks

db
  Managed Postgres
```

The `api-service`, `voice-webhook-service`, and `worker-service` can share Go packages and live in one repository, but deploy as separate processes.

Why split services:

- Frontend changes do not require redeploying backend services.
- Dashboard/API changes do not require redeploying voice webhooks.
- Voice webhook changes can be deployed carefully without touching user auth.
- Worker changes can be deployed without affecting live calls.
- Dedicated single-tenant deployments later can reuse the same service boundaries.

Keep shared code in internal Go packages:

```text
internal/auth
internal/tenant
internal/db
internal/config
internal/audit
internal/notifications
internal/voice
internal/billing
```

Deployment units:

- `web-app`: Vercel
- `api-service`: Fly.io, Render, Railway, or AWS
- `voice-webhook-service`: same host as API or separate service
- `worker-service`: same host as API or separate worker process
- `db`: managed Postgres

For the first MVP, `api-service` and `voice-webhook-service` can be the same Go binary with separate route groups if that reduces setup work. The code should still be organized so they can be split later.

### Modular API Design

The Go API should be designed as a modular service layer before it is split into separate deployable microservices.

Start with one `api-service` deployment containing clear internal service boundaries:

```text
AuthService
PracticeService
TenantService
VoiceService
CalendarService
NotificationService
BillingService
AuditService
CallService
AppointmentService
```

Recommended internal package shape:

```text
apps/api/
  cmd/api/
  internal/
    auth/
    practices/
    tenants/
    voice/
    calendar/
    notifications/
    billing/
    audit/
    calls/
    appointments/
```

This gives the codebase clean domain boundaries without taking on early microservice overhead.

Avoid separate deployable services for auth, calendar, billing, and notifications in the MVP. Splitting too early adds service discovery, internal auth, retries, distributed tracing, config sprawl, deployment coordination, and more failure modes.

Split a module into its own deployable service only when there is a clear reason:

- Independent scaling
- Different security boundary
- Different uptime requirement
- Risky deployments
- High traffic
- Enterprise single-tenant requirements
- Clear operational benefit

Likely future split order:

1. `worker-service`, already separate for background jobs
2. `voice-webhook-service`, because live call webhooks need careful isolation
3. `notification-service`, if volume or provider complexity grows
4. `calendar-service`, if integrations become complex
5. `auth-service`, only if there is a strong security or enterprise reason

Recommended repo shape:

```text
apps/
  web/
  api/
  voice/
  worker/
packages/
  shared/
infra/
  migrations/
  deploy/
docs/
```

Pragmatic first implementation:

```text
apps/web        deploys independently
apps/api        Go modular API with auth, dashboard API, and voice routes
apps/worker     Go worker process
```

Later split:

```text
apps/voice      dedicated voice webhook service
```

## 18. Auth Architecture

Use first-party authentication for clinic users instead of Clerk, Auth0, or Supabase Auth.

Reason:

- Maximum control over identity and access
- Fewer third-party agreements for core user auth
- No patient or practice identity data stored in an external auth provider
- Easier future migration to dedicated single-tenant deployments
- Cleaner compliance story if implemented carefully

User auth should live in the Go backend.

Initial auth features:

- Email and password login
- Secure password hashing with Argon2id or bcrypt
- Email verification
- Password reset
- Session management
- Refresh token rotation
- MFA-ready design
- Practice membership and roles
- Admin invite flow
- Audit logs for login, logout, invite, role changes, and failed login attempts

Recommended internal roles:

- owner
- admin
- staff
- billing
- viewer

Account model:

```text
practice
  -> locations
  -> users / employees
  -> roles
  -> permissions
```

Each employee should have their own login. Do not use shared clinic logins.

Initial SaaS-style signup flow:

1. First user signs up.
2. They create a dental practice.
3. They become the practice `owner`.
4. Owner or admin creates locations.
5. Owner or admin invites employees.
6. Employees accept invite and set their own password.
7. Owner or admin assigns roles and permissions.
8. Owner or admin can deactivate users.

The first implementation may create users directly from the admin panel while email invites are being built, but the long-term flow should be email invitation based.

Role definitions:

- `owner`: full access, billing, user management, locations, roles, permissions, settings
- `admin`: manage practice settings, locations, users, calls, appointments, and assistant config
- `staff`: view calls, handle appointment requests, update call/request statuses
- `billing`: billing settings and invoices only
- `viewer`: read-only access

Permission examples:

- `practice:create`
- `practice:update`
- `practice:delete`
- `location:create`
- `location:update`
- `location:delete`
- `member:invite`
- `member:update`
- `member:disable`
- `role:create`
- `role:update`
- `role:delete`
- `call:read`
- `call:update`
- `appointment:read`
- `appointment:update`
- `assistant_config:read`
- `assistant_config:update`
- `billing:read`
- `billing:update`

Owners should be able to create and manage practices. Admins should be able to manage locations, users, roles, and permissions inside practices they administer. Cross-practice management must require membership and the right permission for the target practice.

Session model:

```text
browser
  -> httpOnly secure session cookie
  -> Go API validates session
  -> Go API loads user, practice membership, role, and scopes
```

The frontend should never be trusted as the source of permissions. Every backend request must check:

- Is the session valid?
- Is the user a member of this practice?
- Does the user's role allow this action?
- Is the requested record scoped to the same practice?

External systems still need their own standard verification:

- Vapi or Retell: signed request or per-assistant API token
- Stripe: webhook signature verification
- Twilio: request signature verification
- Resend: API key
- Google Calendar: OAuth only when the practice connects a calendar

Keep one internal authorization model even if external systems authenticate differently.

Internal actor shape:

```text
actor_type: user | integration | system | webhook
actor_id: internal user, integration, or service id
practice_id: tenant scope
role: role if applicable
scopes: allowed actions
```

Do not store protected health information in auth tables. Keep auth tables focused on users, credentials, sessions, memberships, roles, and audit events.

## 19. Deployment Model

The initial product should be a multi-tenant cloud SaaS.

```text
one production application
  -> many practices
  -> each practice has isolated users, locations, phone numbers, assistant config, calls, appointments, and messages
```

This is the right default for small and mid-sized practices because it keeps onboarding, updates, support, monitoring, and billing simple.

### Customer Setup

Businesses should get access through a standard account flow:

1. Create an account.
2. Create a practice profile.
3. Pick the dental template.
4. Add location, hours, services, appointment types, and escalation rules.
5. Connect a calendar or choose message-taking only.
6. Provision a new phone number or configure call forwarding.
7. Test the AI receptionist.
8. Go live.

### Phone Setup

The first version should support call forwarding and new AI numbers.

```text
patient calls clinic
  -> clinic forwards missed or after-hours calls
  -> AI receptionist answers
```

Later options:

- Port or import the clinic's existing number
- Configure staff-first routing
- Configure AI-first routing
- Add custom IVR behavior
- Add dedicated SIP trunking for larger customers

### Tenant Isolation

Every customer-facing record should belong to a practice.

Core multi-tenant pattern:

```text
practice_id on every tenant-owned table
role-based access for users
row-level security where supported
audit logs for sensitive actions
separate storage paths per practice
strict admin access controls
```

Core tenant-owned tables:

- practices
- practice_members
- locations
- phone_numbers
- assistant_configs
- call_sessions
- call_transcripts
- call_summaries
- callers
- appointment_requests
- messages
- escalations
- integrations
- audit_logs
- billing_subscriptions

Application code should never rely only on frontend filtering. Backend queries must always scope data by `practice_id`.

### Future Single-Tenant Path

The product should be built so larger customers can eventually run on a dedicated deployment without rewriting the application.

Future single-tenant model:

```text
same application codebase
  -> customer-specific environment
  -> customer-specific database
  -> customer-specific storage bucket
  -> customer-specific voice/telephony config
  -> customer-specific encryption and retention settings
```

This should be a premium enterprise option, not the default.

Good reasons to offer single-tenant later:

- Large clinic group
- Contractual data isolation requirements
- Custom compliance requirements
- Dedicated integration needs
- Higher uptime or support requirements
- Customer-specific retention and encryption policies

Avoid on-premise deployment early. It adds heavy support, update, monitoring, security, and telephony complexity. On-prem should only be considered for hospitals, government, or very large enterprise contracts.

## 20. Data Model

Initial tables:

- practices
- users
- user_credentials
- user_sessions
- password_reset_tokens
- email_verification_tokens
- practice_members
- roles
- role_permissions
- member_permissions
- locations
- phone_numbers
- assistant_configs
- call_sessions
- call_transcripts
- call_summaries
- callers
- appointment_requests
- messages
- escalations
- integrations
- audit_logs
- billing_subscriptions

Data model principles:

- Keep tenant context explicit.
- Keep business configuration separate from patient and call data.
- Keep background jobs tenant-aware.
- Keep audit logging centralized.
- Make data export and deletion possible per practice.
- Avoid protected health information in analytics, general logs, and error traces.

## 21. Product Roadmap

### Near-Term

- Build dental call flow
- Build onboarding form
- Build call summary pipeline
- Build demo number
- Split deployable frontend, API, and worker services
- Add staff email and SMS notifications
- Add confirmed calendar booking when workflow is simple
- Add SMS confirmations
- Add missed-call callback
- Add call outcome analytics
- Add admin controls for office staff
- Add recording retention settings
- Add dental-specific test call scripts

### Mid-Term

- Add support for med spas, therapy clinics, chiropractors, and physical therapy
- Add deeper calendar integrations
- Add lightweight CRM/lead pipeline
- Add form delivery and intake collection
- Add payment link support
- Add multi-location support
- Add richer human handoff routing
- Add quality scoring for calls
- Add configurable data retention per practice

### Long-Term

- Add EHR or practice management integrations where commercially justified
- Add custom Twilio/OpenAI realtime voice stack if provider costs or control become limiting
- Add dedicated single-tenant cloud deployments for larger clinic groups
- Add advanced compliance reporting
- Add role-based enterprise administration
- Add custom SLAs and priority support
- Add specialty-specific products under the same platform
- Add marketplace or partner integrations for agencies serving clinics

## 22. Build Order

Recommended implementation order:

1. Product spec
2. Dental call flows
3. Onboarding form
4. Database schema
5. First-party auth
6. Service skeletons for web, API, and worker
7. Vapi or Retell prototype
8. Dashboard
9. Call summaries
10. Staff notifications
11. Demo site
12. Billing

## 23. Starting Decisions

These are the baseline decisions for the first build.

- First niche: dental clinics
- Public first product: DentalDesk AI
- Architecture: multiple deployable services
- API design: modular Go API first, microservices later only when justified
- Deployment: multi-tenant cloud SaaS
- Future enterprise path: dedicated single-tenant cloud deployments from the same codebase
- On-premise: not supported early
- User auth: first-party auth in the Go backend
- Pilot phone setup: call forwarding first
- Voice provider: Vapi or Retell first, custom Twilio/OpenAI later if needed
- Scheduling model: appointment request first, confirmed booking later
- Integrations: avoid EHR integration in the first version
- Calendar: Google Calendar or Calendly-style scheduling first
- Data storage: call summaries by default, recordings optional
- AI scope: administrative receptionist only
- Clinical advice: not allowed
- Emergency handling: immediate escalation language and human handoff where possible
- Compliance posture: design as HIPAA-sensitive from the beginning
- Analytics: no protected health information in analytics, logs, or error traces
- Pricing model: subscription plus usage, not unlimited calls
