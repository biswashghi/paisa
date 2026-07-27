import PaisaLogo from "./PaisaLogo.jsx";

export default function Login({ email, password, error, loading, onLogin, onEmailChange, onPasswordChange }) {
  return (
    <main className="login-page">
      <section className="login-card">
        <PaisaLogo variant="login" />
        <p className="eyebrow">Partner login</p>
        <h1>Paisa</h1>
        <div className="login-partner">
          <span>Secure workspace</span>
          <strong>Email and password required</strong>
        </div>
        <label className="login-field">
          Email
          <input value={email} onChange={(event) => onEmailChange(event.target.value)} placeholder="admin@partner.com" autoComplete="email" />
        </label>
        <label className="login-field">
          Password
          <input type="password" value={password} onChange={(event) => onPasswordChange(event.target.value)} placeholder="Password" autoComplete="current-password" />
        </label>
        {error ? <p className="login-error">{error}</p> : null}
        <button className="primary wide" type="button" onClick={() => onLogin({ email, password })} disabled={loading || !email.trim() || !password.trim()}>
          {loading ? "Signing in..." : "Sign in"}
        </button>
      </section>
    </main>
  );
}
