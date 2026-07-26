import { useState } from "react";
import Cashier from "./Cashier.jsx";
import StatusPill from "./StatusPill.jsx";

export default function Onboarding({
  partner,
  programs,
  transactions,
  locations,
  dashboardSummary,
  catalogItems,
  cashier,
  onCreateLocation,
  onCreateProgram,
  onChangeView,
  onResolveMember,
  onCreateTransaction,
  onCreateRedemption,
  onValidateRedemption,
  onCaptureRedemption,
  onReleaseRedemption,
}) {
  const [location, setLocation] = useState({ name: "Main counter", address: "", timezone: "America/Detroit" });
  const publishedPrograms = programs.filter((program) => program.status === "published").length;
  const checklist = [
    {
      label: "Partner session",
      done: true,
      detail: "Signed in.",
    },
    {
      label: "Location",
      done: locations.length > 0,
      detail: locations.length ? `${locations.length} saved.` : "Add the first checkout location below.",
    },
    {
      label: "Program",
      done: programs.length > 0,
      detail: programs.length ? `${programs.length} created.` : "Create the first loyalty program.",
      action: "Create program",
      onClick: onCreateProgram,
    },
    {
      label: "Rules",
      done: publishedPrograms > 0,
      detail: publishedPrograms ? `${publishedPrograms} published.` : "Publish the earning rules.",
      action: "Edit rules",
      onClick: () => onChangeView("programs"),
    },
    {
      label: "Reward",
      done: (dashboardSummary?.activeCatalogItems || 0) > 0,
      detail: (dashboardSummary?.activeCatalogItems || 0) > 0 ? `${dashboardSummary.activeCatalogItems} active.` : "Add a redeemable reward.",
      action: "Add reward",
      onClick: () => onChangeView("programs"),
    },
    {
      label: "Checkout test",
      done: transactions.length > 0,
      detail: transactions.length ? `${transactions.length} transactions recorded.` : "Run an earn and redeem test.",
      action: "Open cashier",
      onClick: () => onChangeView("setup"),
    },
  ];
  const nextStepIndex = checklist.findIndex((step) => !step.done);
  const nextStep = nextStepIndex >= 0 ? checklist[nextStepIndex] : null;

  return (
    <section className="view-stack">
      <div className="view-header">
        <div>
          <p className="eyebrow">{partner.name}</p>
          <h2>Setup</h2>
        </div>
      </div>
      <section className="panel spacious">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Next</p>
            <h3>{nextStep?.label || "Ready"}</h3>
          </div>
          {nextStep?.action ? (
            <button type="button" onClick={nextStep.onClick}>{nextStep.action}</button>
          ) : null}
        </div>
        <div className="setup-flow">
          {checklist.map((step, index) => (
            <article className={index === nextStepIndex ? "setup-step active" : step.done ? "setup-step done" : "setup-step"} key={step.label}>
              <span className="setup-step-index">{String(index + 1).padStart(2, "0")}</span>
              <div>
                <strong>{step.label}</strong>
                <small>{step.detail}</small>
              </div>
              <StatusPill value={step.done ? "ready" : "todo"} />
              {index === nextStepIndex && step.action ? (
                <button type="button" onClick={step.onClick}>{step.action}</button>
              ) : null}
            </article>
          ))}
        </div>
      </section>
      <section className="panel spacious">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Location</p>
            <h3>Checkout location</h3>
          </div>
        </div>
        <div className="form-grid">
          <label>
            Name
            <input value={location.name} onChange={(event) => setLocation({ ...location, name: event.target.value })} />
          </label>
          <label>
            Address
            <input value={location.address} onChange={(event) => setLocation({ ...location, address: event.target.value })} />
          </label>
          <label>
            Timezone
            <input value={location.timezone} onChange={(event) => setLocation({ ...location, timezone: event.target.value })} />
          </label>
        </div>
        <button className="primary" type="button" onClick={() => onCreateLocation(location)}>Save location</button>
        <div className="card-grid">
          {locations.map((item) => (
            <article className="mini-card" key={item.id}>
              <div><strong>{item.name}</strong><StatusPill value={item.status} /></div>
              <span>{item.address || "No address set"}</span>
              <small>{item.timezone}</small>
            </article>
          ))}
        </div>
      </section>
      <section className="setup-test-section">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Test</p>
            <h3>Checkout</h3>
          </div>
        </div>
        <Cashier
          programs={programs}
          catalogItems={catalogItems}
          cashier={cashier}
          onResolveMember={onResolveMember}
          onCreateTransaction={onCreateTransaction}
          onCreateRedemption={onCreateRedemption}
          onValidateRedemption={onValidateRedemption}
          onCaptureRedemption={onCaptureRedemption}
          onReleaseRedemption={onReleaseRedemption}
          embedded
        />
      </section>
    </section>
  );
}
