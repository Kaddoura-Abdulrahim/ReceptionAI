"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { api, AppointmentRequest, AssistantConfig, AuditLog, BillingSubscription, CalendarConfig, CallSummary, Location, Member, Practice, PracticeInvite, Role, VoiceProviderConfig } from "../lib/api";

export default function Home() {
  const [mode, setMode] = useState<"login" | "register">("register");
  const [email, setEmail] = useState("owner@example.com");
  const [displayName, setDisplayName] = useState("Dental Owner");
  const [password, setPassword] = useState("ChangeMePlease123");
  const [error, setError] = useState("");
  const [authMessage, setAuthMessage] = useState("");
  const [loading, setLoading] = useState(false);
  const [practices, setPractices] = useState<Practice[]>([]);
  const [selectedPracticeId, setSelectedPracticeId] = useState("");
  const [practiceName, setPracticeName] = useState("Downtown Dental");
  const [assistantConfig, setAssistantConfig] = useState<AssistantConfig | null>(null);
  const [voiceProviderConfig, setVoiceProviderConfig] = useState<VoiceProviderConfig | null>(null);
  const [calendarConfig, setCalendarConfig] = useState<CalendarConfig | null>(null);
  const [billing, setBilling] = useState<BillingSubscription | null>(null);
  const [greeting, setGreeting] = useState("");
  const [escalationPhone, setEscalationPhone] = useState("");
  const [notificationEmail, setNotificationEmail] = useState("");
  const [officeHours, setOfficeHours] = useState("Monday-Friday 8:00 AM-5:00 PM");
  const [services, setServices] = useState("cleanings, exams, whitening, emergency dental visits");
  const [acceptedInsurance, setAcceptedInsurance] = useState("Aetna, Delta Dental, Cigna");
  const [newPatientRules, setNewPatientRules] = useState("Accepting new patients. Capture preferred days and callback number.");
  const [emergencyRules, setEmergencyRules] = useState("Escalate severe pain, swelling, heavy bleeding, trauma, trouble breathing, or post-surgery complications.");
  const [cancellationPolicy, setCancellationPolicy] = useState("Ask patients to call at least 24 hours before cancellation when possible.");
  const [intakeFormLink, setIntakeFormLink] = useState("");
  const [voiceTone, setVoiceTone] = useState("warm, concise, professional");
  const [voiceProvider, setVoiceProvider] = useState("vapi");
  const [voicePhoneNumber, setVoicePhoneNumber] = useState("");
  const [voiceAssistantId, setVoiceAssistantId] = useState("");
  const [calendarMode, setCalendarMode] = useState("request_only");
  const [calendarProvider, setCalendarProvider] = useState("none");
  const [bookingUrl, setBookingUrl] = useState("");
  const [calendarId, setCalendarId] = useState("");
  const [calendarTimezone, setCalendarTimezone] = useState("America/New_York");
  const [calendarInstructions, setCalendarInstructions] = useState("");
  const [billingPlan, setBillingPlan] = useState("pilot");
  const [billingStatus, setBillingStatus] = useState("manual");
  const [includedMinutes, setIncludedMinutes] = useState(300);
  const [overageCents, setOverageCents] = useState(25);
  const [stripeCustomerId, setStripeCustomerId] = useState("");
  const [stripeSubscriptionId, setStripeSubscriptionId] = useState("");
  const [testCallerName, setTestCallerName] = useState("Sarah Ahmed");
  const [testCallerPhone, setTestCallerPhone] = useState("+1555010100");
  const [testRequestType, setTestRequestType] = useState("new patient cleaning");
  const [testPreferredTime, setTestPreferredTime] = useState("weekday morning");
  const [testInsurance, setTestInsurance] = useState("Aetna");
  const [testNotes, setTestNotes] = useState("Caller asked whether the office is accepting new patients.");
  const [testUrgency, setTestUrgency] = useState("routine");
  const [appointments, setAppointments] = useState<AppointmentRequest[]>([]);
  const [appointmentFilter, setAppointmentFilter] = useState("open");
  const [appointmentNotes, setAppointmentNotes] = useState<Record<string, string>>({});
  const [appointmentSchedule, setAppointmentSchedule] = useState<Record<string, { start: string; end: string; attendee: string }>>({});
  const [summaries, setSummaries] = useState<CallSummary[]>([]);
  const [locations, setLocations] = useState<Location[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [members, setMembers] = useState<Member[]>([]);
  const [invites, setInvites] = useState<PracticeInvite[]>([]);
  const [activity, setActivity] = useState<AuditLog[]>([]);
  const [locationName, setLocationName] = useState("Main Office");
  const [locationCity, setLocationCity] = useState("Austin");
  const [memberEmail, setMemberEmail] = useState("frontdesk@example.com");
  const [memberRole, setMemberRole] = useState("staff");
  const [roleName, setRoleName] = useState("scheduler");
  const selectedPractice = useMemo(
    () => practices.find((practice) => practice.id === selectedPracticeId),
    [practices, selectedPracticeId],
  );
  const filteredAppointments = useMemo(() => {
    if (appointmentFilter === "all") return appointments;
    if (appointmentFilter === "open") {
      return appointments.filter((item) => item.status === "new" || item.status === "contacted");
    }
    return appointments.filter((item) => item.status === appointmentFilter);
  }, [appointments, appointmentFilter]);

  useEffect(() => {
    if (!selectedPracticeId) return;
    void Promise.all([
      api.assistantConfig(selectedPracticeId).then(setAssistantConfig),
      api.voiceProviderConfig(selectedPracticeId).then(setVoiceProviderConfig),
      api.calendarConfig(selectedPracticeId).then(setCalendarConfig),
      api.billing(selectedPracticeId).then(setBilling),
      api.appointmentRequests(selectedPracticeId).then(setAppointments),
      api.callSummaries(selectedPracticeId).then(setSummaries),
      api.locations(selectedPracticeId).then(setLocations),
      api.roles(selectedPracticeId).then(setRoles),
      api.members(selectedPracticeId).then(setMembers),
      api.invites(selectedPracticeId).then(setInvites),
      api.activity(selectedPracticeId).then(setActivity),
    ]).catch((err) => setError(err.message));
  }, [selectedPracticeId]);

  useEffect(() => {
    setAppointmentNotes((current) => {
      const next = { ...current };
      for (const appointment of appointments) {
        if (next[appointment.id] === undefined) {
          next[appointment.id] = appointment.staffNote ?? "";
        }
      }
      return next;
    });
    setAppointmentSchedule((current) => {
      const next = { ...current };
      for (const appointment of appointments) {
        if (next[appointment.id] === undefined) {
          next[appointment.id] = { start: "", end: "", attendee: "" };
        }
      }
      return next;
    });
  }, [appointments]);

  async function refreshActivity() {
    if (!selectedPracticeId) return;
    try {
      setActivity(await api.activity(selectedPracticeId));
    } catch {
      // Activity should not block the primary workflow.
    }
  }

  useEffect(() => {
    if (!assistantConfig) return;
    const settings = assistantConfig.settings ?? {};
    setGreeting(assistantConfig.greeting ?? "");
    setEscalationPhone(assistantConfig.escalationPhone ?? "");
    setNotificationEmail(assistantConfig.notificationEmail ?? "");
    setOfficeHours(settings.officeHours ?? "Monday-Friday 8:00 AM-5:00 PM");
    setServices(settings.services ?? "cleanings, exams, whitening, emergency dental visits");
    setAcceptedInsurance(settings.acceptedInsurance ?? "Aetna, Delta Dental, Cigna");
    setNewPatientRules(settings.newPatientRules ?? "Accepting new patients. Capture preferred days and callback number.");
    setEmergencyRules(settings.emergencyRules ?? "Escalate severe pain, swelling, heavy bleeding, trauma, trouble breathing, or post-surgery complications.");
    setCancellationPolicy(settings.cancellationPolicy ?? "Ask patients to call at least 24 hours before cancellation when possible.");
    setIntakeFormLink(settings.intakeFormLink ?? "");
    setVoiceTone(settings.voiceTone ?? "warm, concise, professional");
  }, [assistantConfig]);

  useEffect(() => {
    if (!voiceProviderConfig) return;
    setVoiceProvider(voiceProviderConfig.provider || "vapi");
    setVoicePhoneNumber(voiceProviderConfig.phoneNumber || "");
    setVoiceAssistantId(voiceProviderConfig.assistantId || "");
  }, [voiceProviderConfig]);

  useEffect(() => {
    if (!calendarConfig) return;
    setCalendarMode(calendarConfig.mode || "request_only");
    setCalendarProvider(calendarConfig.provider || "none");
    setBookingUrl(calendarConfig.bookingUrl || "");
    setCalendarId(calendarConfig.calendarId || "");
    setCalendarTimezone(calendarConfig.timezone || "America/New_York");
    setCalendarInstructions(calendarConfig.instructions || "");
  }, [calendarConfig]);

  useEffect(() => {
    if (!billing) return;
    setBillingPlan(billing.plan || "pilot");
    setBillingStatus(billing.status || "manual");
    setIncludedMinutes(billing.includedMinutes || 300);
    setOverageCents(billing.overageCents || 25);
    setStripeCustomerId(billing.stripeCustomerId || "");
    setStripeSubscriptionId(billing.stripeSubscriptionId || "");
  }, [billing]);

  async function submitAuth(event: FormEvent) {
    event.preventDefault();
    setError("");
    setAuthMessage("");
    setLoading(true);
    try {
      if (mode === "register") {
        await api.register({ email, displayName, password });
      } else {
        await api.login({ email, password });
      }
      const nextPractices = await api.practices();
      setPractices(nextPractices);
      setSelectedPracticeId(nextPractices[0]?.id ?? "");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Authentication failed");
    } finally {
      setLoading(false);
    }
  }

  async function requestPasswordReset() {
    setError("");
    setAuthMessage("");
    setLoading(true);
    try {
      await api.requestPasswordReset(email);
      setAuthMessage("If that email exists, a reset link has been sent.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not request password reset");
    } finally {
      setLoading(false);
    }
  }

  async function createPractice(event: FormEvent) {
    event.preventDefault();
    setError("");
    setLoading(true);
    try {
      const practice = await api.createPractice(practiceName);
      setPractices((current) => [practice, ...current]);
      setSelectedPracticeId(practice.id);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not create practice");
    } finally {
      setLoading(false);
    }
  }

  async function createLocation(event: FormEvent) {
    event.preventDefault();
    if (!selectedPracticeId) return;
    setError("");
    setLoading(true);
    try {
      const location = await api.createLocation(selectedPracticeId, {
        name: locationName,
        addressLine1: "Address pending",
        city: locationCity,
        region: "TX",
        postalCode: "00000",
        country: "US",
        timezone: "America/Chicago",
      });
      setLocations((current) => [location, ...current]);
      await refreshActivity();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not create location");
    } finally {
      setLoading(false);
    }
  }

  async function createInvite(event: FormEvent) {
    event.preventDefault();
    if (!selectedPracticeId) return;
    setError("");
    setLoading(true);
    try {
      const invite = await api.createInvite(selectedPracticeId, {
        email: memberEmail,
        role: memberRole,
      });
      setInvites((current) => [invite, ...current]);
      await refreshActivity();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not create invite");
    } finally {
      setLoading(false);
    }
  }

  async function createRole(event: FormEvent) {
    event.preventDefault();
    if (!selectedPracticeId) return;
    setError("");
    setLoading(true);
    try {
      const role = await api.createRole(selectedPracticeId, {
        name: roleName,
        description: "Can manage appointment requests",
        permissions: ["call:read", "appointment:read", "appointment:update", "assistant_config:read"],
      });
      setRoles((current) => [role, ...current]);
      await refreshActivity();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not create role");
    } finally {
      setLoading(false);
    }
  }

  async function saveAssistantConfig(event: FormEvent) {
    event.preventDefault();
    if (!selectedPracticeId || !assistantConfig) return;
    setError("");
    setLoading(true);
    try {
      const updated = await api.updateAssistantConfig(selectedPracticeId, {
        greeting,
        escalationPhone,
        notificationEmail,
        settings: {
          ...(assistantConfig.settings ?? {}),
          officeHours,
          services,
          acceptedInsurance,
          newPatientRules,
          emergencyRules,
          cancellationPolicy,
          intakeFormLink,
          voiceTone,
        },
      });
      setAssistantConfig(updated);
      await refreshActivity();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not save assistant config");
    } finally {
      setLoading(false);
    }
  }

  async function saveVoiceProviderConfig(event: FormEvent) {
    event.preventDefault();
    if (!selectedPracticeId) return;
    setError("");
    setLoading(true);
    try {
      const updated = await api.updateVoiceProviderConfig(selectedPracticeId, {
        provider: voiceProvider,
        phoneNumber: voicePhoneNumber,
        assistantId: voiceAssistantId,
      });
      setVoiceProviderConfig(updated);
      await refreshActivity();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not save voice provider config");
    } finally {
      setLoading(false);
    }
  }

  async function saveCalendarConfig(event: FormEvent) {
    event.preventDefault();
    if (!selectedPracticeId) return;
    setError("");
    setLoading(true);
    try {
      const updated = await api.updateCalendarConfig(selectedPracticeId, {
        mode: calendarMode,
        provider: calendarProvider,
        bookingUrl,
        calendarId,
        timezone: calendarTimezone,
        instructions: calendarInstructions,
      });
      setCalendarConfig(updated);
      await refreshActivity();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not save calendar config");
    } finally {
      setLoading(false);
    }
  }

  async function connectGoogleCalendar() {
    if (!selectedPracticeId) return;
    setError("");
    setLoading(true);
    try {
      const result = await api.startGoogleCalendarOAuth(selectedPracticeId);
      window.location.href = result.authorizationUrl;
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not start Google Calendar connection");
      setLoading(false);
    }
  }

  async function scheduleAppointment(item: AppointmentRequest) {
    if (!selectedPracticeId) return;
    const schedule = appointmentSchedule[item.id] ?? { start: "", end: "", attendee: "" };
    setError("");
    setLoading(true);
    try {
      await api.createCalendarEvent(selectedPracticeId, {
        summary: `${item.requestType || "Appointment"} - ${item.callerName || "Patient"}`,
        description: [
          `Caller: ${item.callerName}`,
          `Phone: ${item.callerPhone}`,
          `Preferred time: ${item.preferredTime || "Not provided"}`,
          `Insurance: ${item.insurance || "Not provided"}`,
          `Notes: ${item.notes || "None"}`,
          `Staff note: ${appointmentNotes[item.id] ?? item.staffNote ?? ""}`,
        ].join("\n"),
        start: schedule.start,
        end: schedule.end,
        attendee: schedule.attendee,
      });
      const updated = await api.updateAppointmentRequest(selectedPracticeId, item.id, {
        status: "scheduled",
        staffNote: appointmentNotes[item.id] ?? item.staffNote ?? "",
      });
      setAppointments((current) => current.map((candidate) => (candidate.id === updated.id ? updated : candidate)));
      await refreshActivity();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not schedule appointment");
    } finally {
      setLoading(false);
    }
  }

  async function saveBilling(event: FormEvent) {
    event.preventDefault();
    if (!selectedPracticeId) return;
    setError("");
    setLoading(true);
    try {
      const updated = await api.updateBilling(selectedPracticeId, {
        plan: billingPlan,
        status: billingStatus,
        includedMinutes,
        overageCents,
        stripeCustomerId,
        stripeSubscriptionId,
      });
      setBilling(updated);
      await refreshActivity();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not save billing");
    } finally {
      setLoading(false);
    }
  }

  async function openStripeCheckout() {
    if (!selectedPracticeId) return;
    setError("");
    setLoading(true);
    try {
      const session = await api.createCheckoutSession(selectedPracticeId);
      window.location.href = session.url;
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not create Stripe checkout session");
      setLoading(false);
    }
  }

  async function runVoiceTestCall(event: FormEvent) {
    event.preventDefault();
    if (!selectedPracticeId) return;
    setError("");
    setLoading(true);
    try {
      const result = await api.runVoiceTestCall(selectedPracticeId, {
        callerName: testCallerName,
        callerPhone: testCallerPhone,
        requestType: testRequestType,
        preferredTime: testPreferredTime,
        insurance: testInsurance,
        notes: testNotes,
        urgency: testUrgency,
      });
      setAppointments((current) => [result.call, ...current]);
      setSummaries((current) => [result.summary, ...current]);
      await refreshActivity();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not run voice test call");
    } finally {
      setLoading(false);
    }
  }

  async function updateAppointment(item: AppointmentRequest, status?: string) {
    if (!selectedPracticeId) return;
    setError("");
    setLoading(true);
    try {
      const updated = await api.updateAppointmentRequest(selectedPracticeId, item.id, {
        status,
        staffNote: appointmentNotes[item.id] ?? item.staffNote ?? "",
      });
      setAppointments((current) => current.map((candidate) => (candidate.id === updated.id ? updated : candidate)));
      setAppointmentNotes((current) => ({ ...current, [updated.id]: updated.staffNote ?? "" }));
      await refreshActivity();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not update appointment request");
    } finally {
      setLoading(false);
    }
  }

  if (practices.length === 0) {
    return (
      <main className="auth-page">
        <section className="panel auth-card">
          <h1>DentalDesk AI</h1>
          <p className="muted">AI call answering for dental clinics.</p>
          <form className="stack" onSubmit={submitAuth}>
            <div className="field">
              <label>Email</label>
              <input value={email} onChange={(event) => setEmail(event.target.value)} />
            </div>
            {mode === "register" && (
              <div className="field">
                <label>Name</label>
                <input value={displayName} onChange={(event) => setDisplayName(event.target.value)} />
              </div>
            )}
            <div className="field">
              <label>Password</label>
              <input
                type="password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
              />
            </div>
            {error && <div className="error">{error}</div>}
            {authMessage && <div className="muted">{authMessage}</div>}
            <button className="button" disabled={loading}>
              {mode === "register" ? "Create account" : "Sign in"}
            </button>
            {mode === "login" && (
              <button className="button secondary" type="button" onClick={requestPasswordReset} disabled={loading}>
                Send reset email
              </button>
            )}
            <button
              className="button secondary"
              type="button"
              onClick={() => setMode(mode === "register" ? "login" : "register")}
            >
              {mode === "register" ? "Use existing account" : "Create new account"}
            </button>
          </form>
        </section>
      </main>
    );
  }

  return (
    <main className="shell">
      <aside className="sidebar">
        <div className="brand">DentalDesk AI</div>
        <p className="muted">Multi-tenant dental receptionist dashboard.</p>
        <div className="stack">
          {practices.map((practice) => (
            <button
              className={`button ${practice.id === selectedPracticeId ? "" : "secondary"}`}
              key={practice.id}
              onClick={() => setSelectedPracticeId(practice.id)}
            >
              {practice.name}
            </button>
          ))}
        </div>
      </aside>
      <section className="main">
        <div className="topbar">
          <div>
            <h1>{selectedPractice?.name ?? "Practice"}</h1>
            <div className="muted">Dental MVP workspace</div>
          </div>
          <span className="status">Call forwarding MVP</span>
        </div>

        <div className="grid">
          <section className="panel">
            <h2>Create practice</h2>
            <form className="stack" onSubmit={createPractice}>
              <div className="field">
                <label>Practice name</label>
                <input value={practiceName} onChange={(event) => setPracticeName(event.target.value)} />
              </div>
              <button className="button" disabled={loading}>Add practice</button>
            </form>
          </section>

          <section className="panel">
            <h2>Assistant status</h2>
            <div className="stack">
              <div>
                <strong>Greeting</strong>
                <p className="muted">{assistantConfig?.greeting ?? "No config loaded."}</p>
              </div>
              <div>
                <strong>Escalation</strong>
                <p className="muted">{assistantConfig?.escalationPhone || "Not configured yet"}</p>
              </div>
              <div>
                <strong>Notification email</strong>
                <p className="muted">{assistantConfig?.notificationEmail || "Not configured yet"}</p>
              </div>
            </div>
          </section>

          <section className="panel">
            <h2>Metrics</h2>
            <div className="stack">
              <div>Appointment requests: {appointments.length}</div>
              <div>Call summaries: {summaries.length}</div>
              <div>Urgent escalations: {summaries.filter((item) => item.urgency === "urgent").length}</div>
            </div>
          </section>
        </div>

        <section className="panel" style={{ marginTop: 16 }}>
          <h2>Voice provider</h2>
          <form className="grid" onSubmit={saveVoiceProviderConfig}>
            <div className="field">
              <label>Provider</label>
              <select value={voiceProvider} onChange={(event) => setVoiceProvider(event.target.value)}>
                <option value="vapi">Vapi</option>
                <option value="retell">Retell</option>
                <option value="custom">Custom</option>
              </select>
            </div>
            <div className="field">
              <label>Phone number</label>
              <input value={voicePhoneNumber} onChange={(event) => setVoicePhoneNumber(event.target.value)} />
            </div>
            <div className="field">
              <label>Assistant ID</label>
              <input value={voiceAssistantId} onChange={(event) => setVoiceAssistantId(event.target.value)} />
            </div>
            <div className="field">
              <label>Webhook status</label>
              <input value={voiceProviderConfig?.webhookStatus ?? "not_configured"} readOnly />
            </div>
            <div className="field">
              <label>Bootstrap endpoint</label>
              <input value="/v1/voice/bootstrap" readOnly />
            </div>
            <div className="field">
              <label>Webhook header</label>
              <input value="X-DentalDesk-Webhook-Secret" readOnly />
            </div>
            <button className="button" disabled={loading}>Save voice provider</button>
          </form>
          <form className="grid" onSubmit={runVoiceTestCall} style={{ marginTop: 16 }}>
            <div className="field">
              <label>Test caller name</label>
              <input value={testCallerName} onChange={(event) => setTestCallerName(event.target.value)} />
            </div>
            <div className="field">
              <label>Test caller phone</label>
              <input value={testCallerPhone} onChange={(event) => setTestCallerPhone(event.target.value)} />
            </div>
            <div className="field">
              <label>Request type</label>
              <input value={testRequestType} onChange={(event) => setTestRequestType(event.target.value)} />
            </div>
            <div className="field">
              <label>Preferred time</label>
              <input value={testPreferredTime} onChange={(event) => setTestPreferredTime(event.target.value)} />
            </div>
            <div className="field">
              <label>Insurance</label>
              <input value={testInsurance} onChange={(event) => setTestInsurance(event.target.value)} />
            </div>
            <div className="field">
              <label>Urgency</label>
              <select value={testUrgency} onChange={(event) => setTestUrgency(event.target.value)}>
                <option value="routine">Routine</option>
                <option value="urgent">Urgent</option>
              </select>
            </div>
            <div className="field wide">
              <label>Notes</label>
              <textarea value={testNotes} onChange={(event) => setTestNotes(event.target.value)} />
            </div>
            <button className="button secondary" disabled={loading}>Run local voice test</button>
          </form>
        </section>

        <section className="panel" style={{ marginTop: 16 }}>
          <h2>Calendar</h2>
          <form className="grid" onSubmit={saveCalendarConfig}>
            <div className="field">
              <label>Scheduling mode</label>
              <select value={calendarMode} onChange={(event) => setCalendarMode(event.target.value)}>
                <option value="request_only">Request only</option>
                <option value="booking_link">Booking link</option>
                <option value="google">Google Calendar ready</option>
              </select>
            </div>
            <div className="field">
              <label>Provider</label>
              <select value={calendarProvider} onChange={(event) => setCalendarProvider(event.target.value)}>
                <option value="none">None</option>
                <option value="calendly">Calendly</option>
                <option value="google">Google</option>
                <option value="custom">Custom link</option>
              </select>
            </div>
            <div className="field">
              <label>Status</label>
              <input value={calendarConfig?.status ?? "not_configured"} readOnly />
            </div>
            <div className="field">
              <label>Google OAuth</label>
              <input value={calendarConfig?.oauthConnected ? "connected" : "not connected"} readOnly />
            </div>
            <div className="field">
              <label>Timezone</label>
              <input value={calendarTimezone} onChange={(event) => setCalendarTimezone(event.target.value)} />
            </div>
            <div className="field wide">
              <label>Booking link</label>
              <input value={bookingUrl} onChange={(event) => setBookingUrl(event.target.value)} />
            </div>
            <div className="field wide">
              <label>Google calendar ID</label>
              <input value={calendarId} onChange={(event) => setCalendarId(event.target.value)} />
            </div>
            <div className="field wide">
              <label>Scheduling instructions</label>
              <textarea value={calendarInstructions} onChange={(event) => setCalendarInstructions(event.target.value)} />
            </div>
            <button className="button" disabled={loading}>Save calendar config</button>
            <button className="button secondary" type="button" disabled={loading || calendarMode !== "google"} onClick={connectGoogleCalendar}>
              Connect Google Calendar
            </button>
          </form>
        </section>

        <section className="panel" style={{ marginTop: 16 }}>
          <h2>Billing</h2>
          <form className="grid" onSubmit={saveBilling}>
            <div className="field">
              <label>Plan</label>
              <select value={billingPlan} onChange={(event) => setBillingPlan(event.target.value)}>
                <option value="pilot">Pilot</option>
                <option value="starter">Starter</option>
                <option value="growth">Growth</option>
                <option value="custom">Custom</option>
              </select>
            </div>
            <div className="field">
              <label>Status</label>
              <select value={billingStatus} onChange={(event) => setBillingStatus(event.target.value)}>
                <option value="manual">Manual</option>
                <option value="trialing">Trialing</option>
                <option value="active">Active</option>
                <option value="past_due">Past due</option>
                <option value="canceled">Canceled</option>
              </select>
            </div>
            <div className="field">
              <label>Included minutes</label>
              <input type="number" min="1" value={includedMinutes} onChange={(event) => setIncludedMinutes(Number(event.target.value))} />
            </div>
            <div className="field">
              <label>Overage cents/min</label>
              <input type="number" min="1" value={overageCents} onChange={(event) => setOverageCents(Number(event.target.value))} />
            </div>
            <div className="field wide">
              <label>Stripe customer ID</label>
              <input value={stripeCustomerId} onChange={(event) => setStripeCustomerId(event.target.value)} />
            </div>
            <div className="field wide">
              <label>Stripe subscription ID</label>
              <input value={stripeSubscriptionId} onChange={(event) => setStripeSubscriptionId(event.target.value)} />
            </div>
            <button className="button" disabled={loading}>Save billing</button>
            <button className="button secondary" type="button" disabled={loading} onClick={openStripeCheckout}>
              Open Stripe Checkout
            </button>
          </form>
        </section>

        <section className="panel" style={{ marginTop: 16 }}>
          <h2>Dental onboarding</h2>
          <form className="grid" onSubmit={saveAssistantConfig}>
            <div className="field wide">
              <label>Greeting</label>
              <input value={greeting} onChange={(event) => setGreeting(event.target.value)} />
            </div>
            <div className="field">
              <label>Staff handoff phone</label>
              <input value={escalationPhone} onChange={(event) => setEscalationPhone(event.target.value)} />
            </div>
            <div className="field">
              <label>Staff notification email</label>
              <input value={notificationEmail} onChange={(event) => setNotificationEmail(event.target.value)} />
            </div>
            <div className="field">
              <label>Voice tone</label>
              <input value={voiceTone} onChange={(event) => setVoiceTone(event.target.value)} />
            </div>
            <div className="field">
              <label>Office hours</label>
              <textarea value={officeHours} onChange={(event) => setOfficeHours(event.target.value)} />
            </div>
            <div className="field">
              <label>Services</label>
              <textarea value={services} onChange={(event) => setServices(event.target.value)} />
            </div>
            <div className="field">
              <label>Accepted insurance</label>
              <textarea value={acceptedInsurance} onChange={(event) => setAcceptedInsurance(event.target.value)} />
            </div>
            <div className="field">
              <label>New patient rules</label>
              <textarea value={newPatientRules} onChange={(event) => setNewPatientRules(event.target.value)} />
            </div>
            <div className="field">
              <label>Emergency escalation rules</label>
              <textarea value={emergencyRules} onChange={(event) => setEmergencyRules(event.target.value)} />
            </div>
            <div className="field">
              <label>Cancellation policy</label>
              <textarea value={cancellationPolicy} onChange={(event) => setCancellationPolicy(event.target.value)} />
            </div>
            <div className="field wide">
              <label>Intake form link</label>
              <input value={intakeFormLink} onChange={(event) => setIntakeFormLink(event.target.value)} />
            </div>
            <button className="button" disabled={loading || !assistantConfig}>Save assistant config</button>
          </form>
        </section>

        <div className="grid" style={{ marginTop: 16 }}>
          <section className="panel">
            <div className="section-header">
              <h3>Appointment requests</h3>
              <select value={appointmentFilter} onChange={(event) => setAppointmentFilter(event.target.value)}>
                <option value="open">Open</option>
                <option value="all">All</option>
                <option value="new">New</option>
                <option value="contacted">Contacted</option>
                <option value="scheduled">Scheduled</option>
                <option value="closed">Closed</option>
                <option value="spam">Spam</option>
              </select>
            </div>
            <div className="list">
              {filteredAppointments.length === 0 && <p className="muted">No appointment requests match this filter.</p>}
              {filteredAppointments.map((item) => (
                <div className="item" key={item.id}>
                  <div className="item-title">
                    <strong>{item.callerName}</strong>
                    <span className="status">{item.status}</span>
                  </div>
                  <div className="muted">{item.requestType} - {item.preferredTime || "No preference"}</div>
                  <div>{item.callerPhone}</div>
                  {item.insurance && <div className="muted">Insurance: {item.insurance}</div>}
                  {item.notes && <p>{item.notes}</p>}
                  <div className="field">
                    <label>Staff note</label>
                    <textarea
                      value={appointmentNotes[item.id] ?? item.staffNote ?? ""}
                      onChange={(event) => setAppointmentNotes((current) => ({ ...current, [item.id]: event.target.value }))}
                    />
                  </div>
                  <div className="appointment-schedule">
                    <div className="field">
                      <label>Start RFC3339</label>
                      <input
                        placeholder="2026-07-11T09:00:00-05:00"
                        value={appointmentSchedule[item.id]?.start ?? ""}
                        onChange={(event) =>
                          setAppointmentSchedule((current) => ({
                            ...current,
                            [item.id]: { ...(current[item.id] ?? { start: "", end: "", attendee: "" }), start: event.target.value },
                          }))
                        }
                      />
                    </div>
                    <div className="field">
                      <label>End RFC3339</label>
                      <input
                        placeholder="2026-07-11T09:30:00-05:00"
                        value={appointmentSchedule[item.id]?.end ?? ""}
                        onChange={(event) =>
                          setAppointmentSchedule((current) => ({
                            ...current,
                            [item.id]: { ...(current[item.id] ?? { start: "", end: "", attendee: "" }), end: event.target.value },
                          }))
                        }
                      />
                    </div>
                    <div className="field">
                      <label>Attendee email</label>
                      <input
                        value={appointmentSchedule[item.id]?.attendee ?? ""}
                        onChange={(event) =>
                          setAppointmentSchedule((current) => ({
                            ...current,
                            [item.id]: { ...(current[item.id] ?? { start: "", end: "", attendee: "" }), attendee: event.target.value },
                          }))
                        }
                      />
                    </div>
                  </div>
                  <div className="actions">
                    <button className="button secondary" type="button" disabled={loading} onClick={() => updateAppointment(item, "contacted")}>
                      Contacted
                    </button>
                    <button className="button secondary" type="button" disabled={loading} onClick={() => updateAppointment(item, "scheduled")}>
                      Scheduled
                    </button>
                    <button className="button secondary" type="button" disabled={loading} onClick={() => updateAppointment(item, "closed")}>
                      Close
                    </button>
                    <button className="button secondary" type="button" disabled={loading} onClick={() => updateAppointment(item, "spam")}>
                      Spam
                    </button>
                    <button className="button" type="button" disabled={loading} onClick={() => updateAppointment(item)}>
                      Save note
                    </button>
                    <button className="button" type="button" disabled={loading || !calendarConfig?.oauthConnected} onClick={() => scheduleAppointment(item)}>
                      Create calendar event
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </section>

          <section className="panel">
            <h3>Call summaries</h3>
            <div className="list">
              {summaries.length === 0 && <p className="muted">No call summaries yet.</p>}
              {summaries.map((item) => (
                <div className="item" key={item.id}>
                  <strong>{item.callerName || "Unknown caller"}</strong>
                  <div className="muted">{item.reason} - {item.urgency}</div>
                  <p>{item.summary}</p>
                </div>
              ))}
            </div>
          </section>

          <section className="panel">
            <h3>Next build items</h3>
            <div className="list">
              <div className="item">Practice onboarding form</div>
              <div className="item">Vapi/Retell assistant wiring</div>
              <div className="item">Staff email/SMS notifications</div>
              <div className="item">Postgres persistence</div>
            </div>
          </section>
        </div>

        <section className="panel" style={{ marginTop: 16 }}>
          <h2>Activity</h2>
          <div className="list">
            {activity.length === 0 && <p className="muted">No activity yet.</p>}
            {activity.map((item) => (
              <div className="item" key={item.id}>
                <strong>{item.action}</strong>
                <div className="muted">
                  {item.actorType} - {item.targetType} - {new Date(item.createdAt).toLocaleString()}
                </div>
                {Object.keys(item.metadata ?? {}).length > 0 && (
                  <div className="muted">
                    {Object.entries(item.metadata).map(([key, value]) => `${key}: ${value}`).join(" - ")}
                  </div>
                )}
              </div>
            ))}
          </div>
        </section>

        <div className="grid" style={{ marginTop: 16 }}>
          <section className="panel">
            <h3>Locations</h3>
            <form className="stack" onSubmit={createLocation}>
              <div className="field">
                <label>Location name</label>
                <input value={locationName} onChange={(event) => setLocationName(event.target.value)} />
              </div>
              <div className="field">
                <label>City</label>
                <input value={locationCity} onChange={(event) => setLocationCity(event.target.value)} />
              </div>
              <button className="button" disabled={loading}>Add location</button>
            </form>
            <div className="list" style={{ marginTop: 12 }}>
              {locations.map((location) => (
                <div className="item" key={location.id}>
                  <strong>{location.name}</strong>
                  <div className="muted">{location.city}, {location.region}</div>
                </div>
              ))}
            </div>
          </section>

          <section className="panel">
            <h3>Employees</h3>
            <form className="stack" onSubmit={createInvite}>
              <div className="field">
                <label>Email</label>
                <input value={memberEmail} onChange={(event) => setMemberEmail(event.target.value)} />
              </div>
              <div className="field">
                <label>Role</label>
                <input value={memberRole} onChange={(event) => setMemberRole(event.target.value)} />
              </div>
              <button className="button" disabled={loading}>Send invite</button>
            </form>
            <div className="list" style={{ marginTop: 12 }}>
              {invites.map((invite) => (
                <div className="item" key={invite.id}>
                  <strong>{invite.email}</strong>
                  <div className="muted">
                    Pending invite - {invite.role} - {invite.emailSent === false ? "email failed" : "email sent"}
                  </div>
                  {invite.emailError && <div className="error">{invite.emailError}</div>}
                  {invite.inviteUrl && <div className="muted">{invite.inviteUrl}</div>}
                </div>
              ))}
              {members.map((member) => (
                <div className="item" key={member.userId}>
                  <strong>{member.displayName || member.email}</strong>
                  <div className="muted">{member.email} - {member.role}</div>
                </div>
              ))}
            </div>
          </section>

          <section className="panel">
            <h3>Roles</h3>
            <form className="stack" onSubmit={createRole}>
              <div className="field">
                <label>Role name</label>
                <input value={roleName} onChange={(event) => setRoleName(event.target.value)} />
              </div>
              <button className="button" disabled={loading}>Create role</button>
            </form>
            <div className="list" style={{ marginTop: 12 }}>
              {roles.map((role) => (
                <div className="item" key={role.id}>
                  <strong>{role.name}</strong>
                  <div className="muted">{role.permissions.length} permissions{role.system ? " - system" : ""}</div>
                </div>
              ))}
            </div>
          </section>
        </div>
      </section>
    </main>
  );
}
