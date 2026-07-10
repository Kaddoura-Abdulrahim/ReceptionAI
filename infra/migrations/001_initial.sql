CREATE TABLE practices (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    specialty TEXT NOT NULL DEFAULT 'dental',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    email_verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_credentials (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE email_verification_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE practice_members (
    practice_id UUID NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (practice_id, user_id)
);

CREATE TABLE roles (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    system BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (practice_id, name)
);

CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission TEXT NOT NULL,
    PRIMARY KEY (role_id, permission)
);

CREATE TABLE member_permissions (
    practice_id UUID NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission TEXT NOT NULL,
    granted BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (practice_id, user_id, permission)
);

CREATE TABLE practice_invites (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    role TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    invited_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    accepted_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE locations (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    address_line1 TEXT NOT NULL,
    address_line2 TEXT NOT NULL DEFAULT '',
    city TEXT NOT NULL,
    region TEXT NOT NULL,
    postal_code TEXT NOT NULL,
    country TEXT NOT NULL DEFAULT 'US',
    timezone TEXT NOT NULL DEFAULT 'America/New_York',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE assistant_configs (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    greeting TEXT NOT NULL,
    escalation_phone TEXT NOT NULL DEFAULT '',
    notification_email TEXT NOT NULL DEFAULT '',
    config_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE voice_provider_configs (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    provider TEXT NOT NULL DEFAULT 'vapi',
    phone_number TEXT NOT NULL DEFAULT '',
    assistant_id TEXT NOT NULL DEFAULT '',
    webhook_status TEXT NOT NULL DEFAULT 'not_configured',
    last_webhook_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (practice_id)
);

CREATE TABLE calendar_configs (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    mode TEXT NOT NULL DEFAULT 'request_only',
    provider TEXT NOT NULL DEFAULT 'none',
    booking_url TEXT NOT NULL DEFAULT '',
    calendar_id TEXT NOT NULL DEFAULT '',
    timezone TEXT NOT NULL DEFAULT 'America/New_York',
    status TEXT NOT NULL DEFAULT 'not_configured',
    instructions TEXT NOT NULL DEFAULT '',
    oauth_connected BOOLEAN NOT NULL DEFAULT false,
    oauth_access_token_enc TEXT NOT NULL DEFAULT '',
    oauth_refresh_token_enc TEXT NOT NULL DEFAULT '',
    oauth_token_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (practice_id)
);

CREATE TABLE billing_subscriptions (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    plan TEXT NOT NULL DEFAULT 'pilot',
    status TEXT NOT NULL DEFAULT 'manual',
    included_minutes INTEGER NOT NULL DEFAULT 300,
    overage_cents INTEGER NOT NULL DEFAULT 25,
    stripe_customer_id TEXT NOT NULL DEFAULT '',
    stripe_subscription_id TEXT NOT NULL DEFAULT '',
    trial_ends_at TIMESTAMPTZ,
    current_period_ends_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (practice_id)
);

CREATE TABLE call_sessions (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    provider_call_id TEXT NOT NULL,
    caller_phone TEXT NOT NULL,
    status TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at TIMESTAMPTZ
);

CREATE TABLE call_summaries (
    id UUID PRIMARY KEY,
    call_session_id UUID REFERENCES call_sessions(id) ON DELETE SET NULL,
    practice_id UUID NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    caller_name TEXT NOT NULL DEFAULT '',
    reason TEXT NOT NULL DEFAULT '',
    urgency TEXT NOT NULL DEFAULT 'routine',
    ai_action TEXT NOT NULL DEFAULT '',
    follow_up_needed TEXT NOT NULL DEFAULT '',
    summary TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE appointment_requests (
    id UUID PRIMARY KEY,
    practice_id UUID NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
    call_session_id UUID REFERENCES call_sessions(id) ON DELETE SET NULL,
    caller_name TEXT NOT NULL,
    caller_phone TEXT NOT NULL,
    request_type TEXT NOT NULL,
    preferred_time TEXT NOT NULL DEFAULT '',
    insurance TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    staff_note TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'new',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    practice_id UUID REFERENCES practices(id) ON DELETE SET NULL,
    actor_type TEXT NOT NULL,
    actor_id TEXT NOT NULL,
    action TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_id TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
