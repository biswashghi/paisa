import Cashier from "./Cashier.jsx";
import Rewards from "./Rewards.jsx";
import RuleStudio from "./RuleStudio.jsx";
import StatusPill from "./StatusPill.jsx";

export default function Onboarding({
  partner,
  programs,
  transactions,
  dashboardSummary,
  catalogItems,
  cashier,
  setupLocked = false,
  onCreateProgram,
  selectedProgram,
  redemptions = [],
  onUpdateProgram,
  onCreateRulePackage,
  onPublishProgramRules,
  onUpdateRulePackage,
  onPublishRulePackage,
  onCreateCatalogItem,
  onChangeView,
  onResolveMember,
  onCreateTransaction,
  onCreateRedemption,
  onValidateRedemption,
  onCaptureRedemption,
  onReleaseRedemption,
  onLogout,
}) {
  const publishedPrograms = programs.filter((program) => program.status === "published").length;
  const checklist = [
    {
      label: "Account access",
      done: true,
      detail: "Signed in.",
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
      action: programs.length ? null : undefined,
    },
    {
      label: "Reward",
      done: (dashboardSummary?.activeCatalogItems || 0) > 0,
      detail: (dashboardSummary?.activeCatalogItems || 0) > 0 ? `${dashboardSummary.activeCatalogItems} active.` : "Add a redeemable reward.",
      action: publishedPrograms > 0 ? null : undefined,
    },
    {
      label: "Checkout test",
      done: transactions.length > 0,
      detail: transactions.length ? `${transactions.length} transactions recorded.` : "Resolve a test member, then run a purchase.",
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
        <div className="button-row">
          {onLogout ? <button type="button" onClick={onLogout}>Sign out</button> : null}
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
        {setupLocked ? (
          <p className="setup-gate-copy">
            Finish these launch steps before the dashboard, programs, members, activity, and settings workspaces unlock.
          </p>
        ) : null}
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
      {programs.length ? (
        <section className="panel spacious setup-config-section">
          <div className="section-heading">
            <div>
              <p className="eyebrow">Rules</p>
              <h3>Configure earning</h3>
            </div>
            <StatusPill value={publishedPrograms > 0 ? "ready" : "todo"} />
          </div>
          <RuleStudio
            program={selectedProgram || programs[0]}
            embedded
            onUpdateProgram={onUpdateProgram}
            onPublishProgramRules={onPublishProgramRules}
            onCreateRulePackage={onCreateRulePackage}
            onUpdateRulePackage={onUpdateRulePackage}
            onPublishRulePackage={onPublishRulePackage}
          />
        </section>
      ) : null}
      {publishedPrograms > 0 ? (
        <section className="panel spacious setup-config-section">
          <div className="section-heading">
            <div>
              <p className="eyebrow">Reward</p>
              <h3>Add a redeemable item</h3>
            </div>
            <StatusPill value={(dashboardSummary?.activeCatalogItems || 0) > 0 ? "ready" : "todo"} />
          </div>
          <Rewards
            catalogItems={catalogItems}
            redemptions={redemptions}
            programs={programs}
            onCreateCatalogItem={onCreateCatalogItem}
            embedded
          />
        </section>
      ) : null}
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
