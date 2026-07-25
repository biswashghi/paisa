import PaisaLogo from "./PaisaLogo.jsx";

export default function Login({ partner, partnerKey, apiBaseUrl, error, loading, onLogin, onPartnerKeyChange }) {
  return (
    <main className="login-page">
      <section className="login-card">
        <PaisaLogo variant="login" />
        <p className="eyebrow">Default partner login</p>
        <h1>Paisa Partner Admin</h1>
        <p className="login-copy">
          Manage loyalty programs, earning rules, enrolled members, and reward activity from one partner workspace.
        </p>
        <div className="login-partner">
          <span>Partner</span>
          <strong>{partner.name}</strong>
          <small>{apiBaseUrl}</small>
        </div>
        <label className="login-field">
          Partner key
          <input value={partnerKey} onChange={(event) => onPartnerKeyChange(event.target.value)} placeholder="acme-retail" />
        </label>
        {error ? <p className="login-error">{error}</p> : null}
        <button className="primary wide" type="button" onClick={() => onLogin(partnerKey)} disabled={loading || !partnerKey.trim()}>
          {loading ? "Connecting..." : `Open ${partnerKey || "partner"} workspace`}
        </button>
      </section>
    </main>
  );
}
