import PaisaLogo from "./PaisaLogo.jsx";

export default function Login({ partner, onLogin }) {
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
          <small>{partner.adminEmail}</small>
        </div>
        <button className="primary wide" type="button" onClick={onLogin}>
          Continue as {partner.adminName}
        </button>
      </section>
    </main>
  );
}
