"use client";

import { FormEvent, useEffect, useState } from "react";
import { api, PracticeInvite } from "../../lib/api";

export default function AcceptInvitePage() {
  const [token, setToken] = useState("");
  const [invite, setInvite] = useState<PracticeInvite | null>(null);
  const [displayName, setDisplayName] = useState("");
  const [password, setPassword] = useState("ChangeMePlease123");
  const [error, setError] = useState("");
  const [done, setDone] = useState(false);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const inviteToken = params.get("token") ?? "";
    setToken(inviteToken);
    if (!inviteToken) {
      setError("Invite token is missing.");
      return;
    }
    api.getInvite(inviteToken)
      .then(setInvite)
      .catch((err) => setError(err.message));
  }, []);

  async function submit(event: FormEvent) {
    event.preventDefault();
    setError("");
    setLoading(true);
    try {
      await api.acceptInvite(token, { displayName, password });
      setDone(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not accept invite");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="auth-page">
      <section className="panel auth-card">
        <h1>Accept invite</h1>
        {invite && <p className="muted">Joining as {invite.email} with role {invite.role}.</p>}
        {done ? (
          <div className="stack">
            <p>Your account is ready.</p>
            <a className="button" href="/">Go to dashboard</a>
          </div>
        ) : (
          <form className="stack" onSubmit={submit}>
            <div className="field">
              <label>Name</label>
              <input value={displayName} onChange={(event) => setDisplayName(event.target.value)} />
            </div>
            <div className="field">
              <label>Password</label>
              <input
                type="password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
              />
            </div>
            {error && <div className="error">{error}</div>}
            <button className="button" disabled={loading || !invite}>Accept invite</button>
          </form>
        )}
      </section>
    </main>
  );
}
