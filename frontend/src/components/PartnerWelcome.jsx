import PaisaLogo from "./PaisaLogo.jsx";

export default function PartnerWelcome({ partner, onBegin, onLogout }) {
  return (
    <section className="welcome-page">
      <div className="welcome-hero">
        <div className="welcome-copy">
          <PaisaLogo variant="login" />
          <p className="eyebrow">{partner.name}</p>
          <h1>Launch rewards without building the machinery yourself.</h1>
          <p>
            Set up one program, publish earning rules, and add the first reward.
            Paisa handles the account, balance, and redemption rails behind it.
          </p>
          <div className="button-row">
            <button className="primary" type="button" onClick={onBegin}>Start setup</button>
            {onLogout ? <button type="button" onClick={onLogout}>Sign out</button> : null}
          </div>
        </div>
        <div className="welcome-orbit" aria-hidden="true">
          <div className="orbit-ring ring-one" />
          <div className="orbit-ring ring-two" />
          <article className="orbit-card card-program">
            <span>01</span>
            <strong>Program</strong>
            <small>Define who earns</small>
          </article>
          <article className="orbit-card card-rules">
            <span>02</span>
            <strong>Earn rules</strong>
            <small>Set the rate</small>
          </article>
          <article className="orbit-card card-reward">
            <span>03</span>
            <strong>Rewards</strong>
            <small>Enable redemption</small>
          </article>
          <div className="orbit-core">
            <strong>Paisa</strong>
            <span>Loyalty hub</span>
          </div>
        </div>
      </div>
    </section>
  );
}
