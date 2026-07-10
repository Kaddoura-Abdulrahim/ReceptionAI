"use client";

import { useEffect, useState } from "react";
import { api } from "../../lib/api";

export default function VerifyEmailPage() {
  const [status, setStatus] = useState("Verifying email...");

  useEffect(() => {
    const token = new URLSearchParams(window.location.search).get("token") ?? "";
    if (!token) {
      setStatus("Verification token is missing.");
      return;
    }
    api.verifyEmail(token)
      .then(() => setStatus("Email verified. You can return to the dashboard."))
      .catch((err) => setStatus(err.message));
  }, []);

  return (
    <main className="auth-page">
      <section className="panel auth-card">
        <h1>Verify email</h1>
        <p className="muted">{status}</p>
        <a className="button" href="/">Go to dashboard</a>
      </section>
    </main>
  );
}
