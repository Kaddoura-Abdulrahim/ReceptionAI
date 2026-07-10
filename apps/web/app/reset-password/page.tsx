"use client";

import { FormEvent, useEffect, useState } from "react";
import { api } from "../../lib/api";

export default function ResetPasswordPage() {
  const [token, setToken] = useState("");
  const [password, setPassword] = useState("ChangeMePlease123");
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setToken(new URLSearchParams(window.location.search).get("token") ?? "");
  }, []);

  async function submit(event: FormEvent) {
    event.preventDefault();
    setLoading(true);
    setMessage("");
    try {
      await api.resetPassword(token, password);
      setMessage("Password reset. You can sign in with the new password.");
    } catch (err) {
      setMessage(err instanceof Error ? err.message : "Could not reset password");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="auth-page">
      <section className="panel auth-card">
        <h1>Reset password</h1>
        <form className="stack" onSubmit={submit}>
          <div className="field">
            <label>New password</label>
            <input type="password" value={password} onChange={(event) => setPassword(event.target.value)} />
          </div>
          {message && <p className="muted">{message}</p>}
          <button className="button" disabled={loading || !token}>Reset password</button>
        </form>
      </section>
    </main>
  );
}
