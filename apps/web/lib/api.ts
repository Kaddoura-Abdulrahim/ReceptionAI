const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
let csrfToken = "";

export type Practice = {
  id: string;
  name: string;
  specialty: string;
  createdAt: string;
};

export type AssistantConfig = {
  id: string;
  practiceId: string;
  greeting: string;
  escalationPhone: string;
  notificationEmail: string;
  settings: Record<string, string>;
};

export type VoiceProviderConfig = {
  id: string;
  practiceId: string;
  provider: string;
  phoneNumber: string;
  assistantId: string;
  webhookStatus: string;
  lastWebhookAt?: string;
  createdAt: string;
  updatedAt: string;
};

export type CalendarConfig = {
  id: string;
  practiceId: string;
  mode: string;
  provider: string;
  bookingUrl: string;
  calendarId: string;
  timezone: string;
  status: string;
  instructions: string;
  oauthConnected: boolean;
  oauthTokenExpiresAt?: string;
  createdAt: string;
  updatedAt: string;
};

export type BillingSubscription = {
  id: string;
  practiceId: string;
  plan: string;
  status: string;
  includedMinutes: number;
  overageCents: number;
  stripeCustomerId: string;
  stripeSubscriptionId: string;
  trialEndsAt?: string;
  currentPeriodEndsAt?: string;
  createdAt: string;
  updatedAt: string;
};

export type AppointmentRequest = {
  id: string;
  callerName: string;
  callerPhone: string;
  requestType: string;
  preferredTime: string;
  insurance: string;
  notes: string;
  staffNote: string;
  status: string;
  createdAt: string;
};

export type CallSummary = {
  id: string;
  callerName: string;
  reason: string;
  urgency: string;
  aiAction: string;
  followUpNeeded: string;
  summary: string;
  createdAt: string;
};

export type Location = {
  id: string;
  practiceId: string;
  name: string;
  addressLine1: string;
  addressLine2: string;
  city: string;
  region: string;
  postalCode: string;
  country: string;
  timezone: string;
};

export type Role = {
  id: string;
  practiceId: string;
  name: string;
  description: string;
  system: boolean;
  permissions: string[];
};

export type Member = {
  practiceId: string;
  userId: string;
  role: string;
  email: string;
  displayName: string;
  active: boolean;
};

export type PracticeInvite = {
  id: string;
  practiceId: string;
  email: string;
  role: string;
  invitedBy: string;
  acceptedAt?: string;
  expiresAt: string;
  createdAt: string;
  inviteUrl?: string;
  emailSent?: boolean;
  emailError?: string;
};

export type AuditLog = {
  id: string;
  practiceId: string;
  actorType: string;
  actorId: string;
  action: string;
  targetType: string;
  targetId: string;
  metadata: Record<string, string>;
  createdAt: string;
};

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const method = (init.method ?? "GET").toUpperCase();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...((init.headers as Record<string, string> | undefined) ?? {}),
  };
  if (!["GET", "HEAD", "OPTIONS"].includes(method)) {
    csrfToken = csrfToken || (await getCSRFToken());
    headers["X-CSRF-Token"] = csrfToken;
  }
  const res = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    credentials: "include",
    headers,
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: "Request failed" }));
    throw new Error(body.error ?? "Request failed");
  }

  return res.json() as Promise<T>;
}

async function getCSRFToken(): Promise<string> {
  const res = await fetch(`${API_BASE_URL}/v1/csrf`, {
    credentials: "include",
  });
  if (!res.ok) {
    throw new Error("Could not create CSRF token");
  }
  const body = (await res.json()) as { csrfToken: string };
  return body.csrfToken;
}

export const api = {
  register: (payload: { email: string; displayName: string; password: string }) =>
    request<{ id: string; email: string; displayName: string }>("/v1/auth/register", {
      method: "POST",
      body: JSON.stringify(payload),
    }),
  login: (payload: { email: string; password: string }) =>
    request<{ id: string; email: string; displayName: string }>("/v1/auth/login", {
      method: "POST",
      body: JSON.stringify(payload),
    }),
  verifyEmail: (token: string) =>
    request<{ id: string; email: string; displayName: string; emailVerifiedAt?: string }>("/v1/auth/verify-email", {
      method: "POST",
      body: JSON.stringify({ token }),
    }),
  requestPasswordReset: (email: string) =>
    request<{ status: string }>("/v1/auth/request-password-reset", {
      method: "POST",
      body: JSON.stringify({ email }),
    }),
  resetPassword: (token: string, password: string) =>
    request<{ id: string; email: string; displayName: string }>("/v1/auth/reset-password", {
      method: "POST",
      body: JSON.stringify({ token, password }),
    }),
  practices: () => request<Practice[]>("/v1/practices"),
  createPractice: (name: string) =>
    request<Practice>("/v1/practices", {
      method: "POST",
      body: JSON.stringify({ name }),
    }),
  assistantConfig: (practiceId: string) =>
    request<AssistantConfig>(`/v1/practices/${practiceId}/assistant-config`),
  updateAssistantConfig: (practiceId: string, payload: Partial<AssistantConfig>) =>
    request<AssistantConfig>(`/v1/practices/${practiceId}/assistant-config`, {
      method: "PATCH",
      body: JSON.stringify(payload),
    }),
  voiceProviderConfig: (practiceId: string) =>
    request<VoiceProviderConfig>(`/v1/practices/${practiceId}/voice-provider`),
  updateVoiceProviderConfig: (practiceId: string, payload: Partial<VoiceProviderConfig>) =>
    request<VoiceProviderConfig>(`/v1/practices/${practiceId}/voice-provider`, {
      method: "PATCH",
      body: JSON.stringify(payload),
    }),
  calendarConfig: (practiceId: string) =>
    request<CalendarConfig>(`/v1/practices/${practiceId}/calendar-config`),
  updateCalendarConfig: (practiceId: string, payload: Partial<CalendarConfig>) =>
    request<CalendarConfig>(`/v1/practices/${practiceId}/calendar-config`, {
      method: "PATCH",
      body: JSON.stringify(payload),
    }),
  startGoogleCalendarOAuth: (practiceId: string) =>
    request<{ authorizationUrl: string }>(`/v1/practices/${practiceId}/calendar/oauth/start`, {
      method: "POST",
      body: JSON.stringify({}),
    }),
  createCalendarEvent: (
    practiceId: string,
    payload: { summary: string; description: string; start: string; end: string; attendee?: string },
  ) =>
    request<Record<string, unknown>>(`/v1/practices/${practiceId}/calendar/events`, {
      method: "POST",
      body: JSON.stringify(payload),
    }),
  billing: (practiceId: string) => request<BillingSubscription>(`/v1/practices/${practiceId}/billing`),
  updateBilling: (practiceId: string, payload: Partial<BillingSubscription>) =>
    request<BillingSubscription>(`/v1/practices/${practiceId}/billing`, {
      method: "PATCH",
      body: JSON.stringify(payload),
    }),
  createCheckoutSession: (practiceId: string) =>
    request<{ id: string; url: string }>(`/v1/practices/${practiceId}/billing/checkout-session`, {
      method: "POST",
      body: JSON.stringify({}),
    }),
  runVoiceTestCall: (
    practiceId: string,
    payload: {
      callerName: string;
      callerPhone: string;
      requestType: string;
      preferredTime: string;
      insurance: string;
      notes: string;
      urgency: string;
    },
  ) =>
    request<{ call: AppointmentRequest; summary: CallSummary }>(`/v1/practices/${practiceId}/voice-test-call`, {
      method: "POST",
      body: JSON.stringify(payload),
    }),
  appointmentRequests: (practiceId: string) =>
    request<AppointmentRequest[]>(`/v1/practices/${practiceId}/appointment-requests`),
  updateAppointmentRequest: (practiceId: string, requestId: string, payload: { status?: string; staffNote?: string }) =>
    request<AppointmentRequest>(`/v1/practices/${practiceId}/appointment-requests/${requestId}`, {
      method: "PATCH",
      body: JSON.stringify(payload),
    }),
  callSummaries: (practiceId: string) =>
    request<CallSummary[]>(`/v1/practices/${practiceId}/call-summaries`),
  activity: (practiceId: string) => request<AuditLog[]>(`/v1/practices/${practiceId}/activity`),
  locations: (practiceId: string) => request<Location[]>(`/v1/practices/${practiceId}/locations`),
  createLocation: (practiceId: string, payload: Partial<Location>) =>
    request<Location>(`/v1/practices/${practiceId}/locations`, {
      method: "POST",
      body: JSON.stringify(payload),
    }),
  roles: (practiceId: string) => request<Role[]>(`/v1/practices/${practiceId}/roles`),
  createRole: (practiceId: string, payload: { name: string; description: string; permissions: string[] }) =>
    request<Role>(`/v1/practices/${practiceId}/roles`, {
      method: "POST",
      body: JSON.stringify(payload),
    }),
  members: (practiceId: string) => request<Member[]>(`/v1/practices/${practiceId}/members`),
  invites: (practiceId: string) => request<PracticeInvite[]>(`/v1/practices/${practiceId}/invites`),
  createInvite: (practiceId: string, payload: { email: string; role: string }) =>
    request<PracticeInvite>(`/v1/practices/${practiceId}/invites`, {
      method: "POST",
      body: JSON.stringify(payload),
    }),
  getInvite: (token: string) => request<PracticeInvite>(`/v1/invites/${token}`),
  acceptInvite: (token: string, payload: { displayName: string; password: string }) =>
    request<PracticeInvite>(`/v1/invites/${token}/accept`, {
      method: "POST",
      body: JSON.stringify(payload),
    }),
};
