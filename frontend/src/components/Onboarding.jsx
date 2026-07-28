import { useMemo, useState } from "react";
import Cashier from "./Cashier.jsx";
import Rewards from "./Rewards.jsx";
import StatusPill from "./StatusPill.jsx";

export default function Onboarding({
  partner,
  programs,
  transactions,
  dashboardSummary,
  catalogItems,
  cashier,
  selectedProgram,
  redemptions = [],
  onCreateProgram,
  onUpdateProgram,
  onPublishProgramRules,
  onCreateCatalogItem,
  onResolveMember,
  onCreateTransaction,
  onCreateRedemption,
  onValidateRedemption,
  onCaptureRedemption,
  onReleaseRedemption,
  onLogout,
}) {
  const publishedPrograms = programs.filter((program) => program.status === "published").length;
  const activeCatalogItems = dashboardSummary?.activeCatalogItems || 0;
  const allDone = programs.length > 0 && publishedPrograms > 0 && activeCatalogItems > 0;
  const currentStep = !programs.length
    ? "program"
    : publishedPrograms === 0
      ? "rules"
      : activeCatalogItems === 0
        ? "reward"
        : "ready";
  const steps = [
    { key: "program", index: "01", label: "Program", detail: programs.length ? `${programs.length} created` : "Create the first program", done: programs.length > 0 },
    { key: "rules", index: "02", label: "Earn rules", detail: publishedPrograms ? `${publishedPrograms} published` : "Publish the base rate", done: publishedPrograms > 0, locked: programs.length === 0 },
    { key: "reward", index: "03", label: "Rewards", detail: activeCatalogItems ? `${activeCatalogItems} active` : "Add a redeemable item", done: activeCatalogItems > 0, locked: publishedPrograms === 0 },
  ];

  return (
    <section className="view-stack">
      <div className="view-header">
        <div>
          <p className="eyebrow">{partner.name}</p>
          <h2>Partner setup</h2>
        </div>
        <div className="button-row">
          {onLogout ? <button type="button" onClick={onLogout}>Sign out</button> : null}
        </div>
      </div>

      <div className="onboarding-layout">
        <aside className="panel onboarding-rail">
          <div>
            <p className="eyebrow">Onboarding</p>
            <h3>{currentStep === "ready" ? "Ready to launch" : "Complete setup"}</h3>
          </div>
          <div className="onboarding-track" aria-label="Required onboarding steps">
            {steps.map((step) => (
              <article className={step.key === currentStep ? "onboarding-step active" : step.done ? "onboarding-step done" : step.locked ? "onboarding-step locked" : "onboarding-step"} key={step.key}>
                <span className="step-orb">{step.done ? "✓" : step.index}</span>
                <div>
                  <strong>{step.label}</strong>
                  <small>{step.detail}</small>
                </div>
              </article>
            ))}
          </div>
        </aside>
        <main className="onboarding-main">
          {currentStep === "program" ? (
            <ProgramSetup programs={programs} onCreateProgram={onCreateProgram} />
          ) : null}

          {currentStep === "rules" && selectedProgram ? (
            <RulesSetup program={selectedProgram} onUpdateProgram={onUpdateProgram} onPublishProgramRules={onPublishProgramRules} />
          ) : null}

          {currentStep === "reward" ? (
            <section className="panel spacious setup-focus-panel">
              <div className="section-heading">
                <div>
                  <p className="eyebrow">Reward catalog</p>
                  <h3>Create the first redeemable reward</h3>
                </div>
                <StatusPill value="todo" />
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

          {allDone ? (
            <section className="panel spacious setup-focus-panel">
              <div className="section-heading">
                <div>
                  <p className="eyebrow">Ready</p>
                  <h3>Setup is complete</h3>
                </div>
                <StatusPill value="ready" />
              </div>
              <p className="helper-copy">The partner workspace is unlocked. Continue into programs, members, activity, and settings from the sidebar.</p>
              <details className="optional-checkout" open={transactions.length === 0}>
                <summary>Optional checkout verification</summary>
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
              </details>
            </section>
          ) : null}
        </main>
      </div>
    </section>
  );
}

function ProgramSetup({ programs, onCreateProgram }) {
  const [name, setName] = useState(programs[0]?.name || "");
  const [tierCode, setTierCode] = useState(programs[0]?.tierCode || "base");
  const canCreate = name.trim() && tierCode.trim();

  return (
    <section className="panel spacious setup-focus-panel">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Program</p>
          <h3>Create the first rewards program</h3>
        </div>
        <StatusPill value="todo" />
      </div>
      <div className="setup-editor-card">
        <div>
          <strong>Program details</strong>
          <span>This defines the member group that earns under one set of rules.</span>
        </div>
        <div className="form-grid">
          <label>
            Program name
            <input value={name} onChange={(event) => setName(event.target.value)} placeholder="Everyday Rewards" />
          </label>
          <label>
            Tier code
            <input value={tierCode} onChange={(event) => setTierCode(event.target.value)} placeholder="base" />
          </label>
        </div>
        <button className="primary" type="button" disabled={!canCreate} onClick={() => onCreateProgram({ name, tierCode })}>
          Create program
        </button>
      </div>
    </section>
  );
}

function RulesSetup({ program, onUpdateProgram, onPublishProgramRules }) {
  const currentBaseRule = program.rules?.groups?.[0]?.rules?.[0];
  const [pointsPerDollar, setPointsPerDollar] = useState(decimalPoints(currentBaseRule?.pointsPerDollar || 1));
  const [earnBasis, setEarnBasis] = useState(program.rules?.earnBasis || "total");
  const normalizedPoints = decimalPoints(pointsPerDollar);
  const explanation = earnBasisExplanation(earnBasis, normalizedPoints);
  const draftProgram = useMemo(() => ({
    ...program,
    status: "draft",
    rules: buildBaseRules(normalizedPoints, earnBasis),
  }), [earnBasis, normalizedPoints, program]);

  function saveDraft() {
    onUpdateProgram(program.id, { rules: draftProgram.rules, status: "draft" });
  }

  function publish() {
    saveDraft();
    onPublishProgramRules(program.id, draftProgram);
  }

  return (
    <section className="panel spacious setup-focus-panel">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Earning rules</p>
          <h3>Set the base earn rate for {program.name}</h3>
        </div>
        <StatusPill value="todo" />
      </div>
      <div className="setup-editor-card">
        <div>
          <strong>Base earn</strong>
          <span>
            This is the default earning rule for every purchase in this program. Bonuses, category rules,
            member add-ons, and caps can be layered in after setup.
          </span>
        </div>
        <div className="form-grid">
          <label>
            Points per dollar
            <input type="number" min="0.01" step="0.01" value={pointsPerDollar} onChange={(event) => setPointsPerDollar(decimalPoints(event.target.value))} />
          </label>
          <label>
            Earn basis
            <select value={earnBasis} onChange={(event) => setEarnBasis(event.target.value)}>
              <option value="total">Total</option>
              <option value="subtotal">Subtotal</option>
              <option value="eligible">Eligible amount</option>
            </select>
          </label>
        </div>
        <div className="rule-preview-line">
          <span>Every qualifying purchase earns</span>
          <strong>{normalizedPoints} point{normalizedPoints === 1 ? "" : "s"} per dollar</strong>
          <small>Calculated from {explanation.basisLabel.toLowerCase()}.</small>
        </div>
        <div className="earn-explainer-grid">
          <article className="earn-formula-card">
            <span>Formula</span>
            <strong>{explanation.basisLabel} x {normalizedPoints} points per dollar</strong>
            <small>{explanation.formula}</small>
          </article>
          <article className="earn-formula-card">
            <span>Plain English</span>
            <strong>{explanation.summary}</strong>
            <small>{explanation.edge}</small>
          </article>
        </div>
        <div className="earn-example-panel">
          <div>
            <p className="eyebrow">Example purchase</p>
            <h4>{explanation.exampleTitle}</h4>
          </div>
          <div className="earn-example-grid">
            {explanation.amounts.map((item) => (
              <div className={item.active ? "active" : ""} key={item.label}>
                <span>{item.label}</span>
                <strong>{formatMoney(item.amount)}</strong>
              </div>
            ))}
          </div>
          <div className="earn-result">
            <span>Points posted</span>
            <strong>{explanation.points} point{explanation.points === 1 ? "" : "s"}</strong>
            <small>{explanation.result}</small>
          </div>
          <div className="edge-case-list">
            {explanation.cases.map((item) => (
              <article key={item.title}>
                <strong>{item.title}</strong>
                <span>{item.detail}</span>
              </article>
            ))}
          </div>
        </div>
        <div className="button-row">
          <button type="button" onClick={saveDraft}>Save draft</button>
          <button className="primary" type="button" onClick={publish} disabled={normalizedPoints <= 0}>Publish rules</button>
        </div>
      </div>
    </section>
  );
}

function buildBaseRules(pointsPerDollar, earnBasis) {
  return {
    earnBasis,
    groups: [{
      id: "setup-base-group",
      name: "Base earn",
      strategy: "stack",
      status: "draft",
      rules: [{
        id: "setup-base-rule",
        key: "base_earn",
        name: "Base earn",
        type: "points_per_dollar",
        pointsPerDollar: Number(pointsPerDollar),
        category: "All transactions",
        cap: "",
        status: "active",
      }],
    }],
  };
}

function decimalPoints(value) {
  const parsed = Number(value);
  if (!Number.isFinite(parsed) || parsed <= 0) return 0.01;
  return Number(parsed.toFixed(4));
}

function earnBasisExplanation(earnBasis, rate) {
  const amounts = {
    eligible: 9000,
    subtotal: 12000,
    tax: 770,
    total: 12770,
  };
  const basisAmount = amounts[earnBasis] || amounts.eligible;
  const points = Math.floor((basisAmount / 100) * rate);
  const commonCases = [
    {
      title: "Refunds",
      detail: "Refunds reverse points from the original calculation, not whatever rules are active later.",
    },
    {
      title: "Whole-point ledger",
      detail: "Fractional results round down per transaction. At 0.01 points per dollar, a $25 basis earns 0 points.",
    },
  ];
  const basisSpecific = {
    eligible: {
      basisLabel: "Eligible amount",
      summary: "Only the transaction amount sent as eligible earns points.",
      formula: "Your POS or transaction API must send eligibleMinor. If it is omitted today, the backend defaults eligible amount to subtotal.",
      edge: "This is not a percentage setting. Use it when your integration can calculate exclusions before sending the transaction.",
      exampleTitle: "$120 subtotal, $30 excluded item, $7.70 tax",
      result: `${formatMoney(basisAmount)} eligible basis x ${rate} = ${points} posted point${points === 1 ? "" : "s"}.`,
      cases: [
        {
          title: "Who sets it",
          detail: "The partner integration sends a specific eligible amount per transaction or line item.",
        },
        {
          title: "Not configured here",
          detail: "Setup does not yet define exclusion formulas, percentages, or category rules. Those belong in advanced rule configuration.",
        },
        {
          title: "Fallback",
          detail: "If eligible amount is missing during ingestion, Paisa currently uses subtotal as the eligible amount.",
        },
      ],
    },
    subtotal: {
      basisLabel: "Subtotal",
      summary: "Points are based on merchandise subtotal before tax.",
      formula: "Example: the pre-tax item subtotal is used, even when tax changes by location.",
      edge: "Use this when every item in the subtotal should earn and taxes should never earn.",
      exampleTitle: "$120 subtotal, $7.70 tax",
      result: `${formatMoney(basisAmount)} subtotal basis x ${rate} = ${points} posted point${points === 1 ? "" : "s"}.`,
      cases: [
        {
          title: "Tax ignored",
          detail: "Tax does not increase earning, which keeps rewards consistent across jurisdictions.",
        },
        ...commonCases,
      ],
    },
    total: {
      basisLabel: "Total",
      summary: "Points are based on the final charged amount including tax.",
      formula: "Example: subtotal after discount plus tax becomes the earning basis.",
      edge: "Use this only if you want taxes and final charges included in rewards.",
      exampleTitle: "$120 subtotal, $7.70 tax",
      result: `${formatMoney(basisAmount)} total basis x ${rate} = ${points} posted point${points === 1 ? "" : "s"}.`,
      cases: [
        {
          title: "Tax earns too",
          detail: "Customers earn on taxes and included charges, so liability may be higher than subtotal or eligible-basis programs.",
        },
        ...commonCases,
      ],
    },
  };
  const selected = basisSpecific[earnBasis] || basisSpecific.eligible;
  return {
    ...selected,
    points,
    amounts: [
      { label: "Total", amount: amounts.total, active: earnBasis === "total" },
      { label: "Subtotal", amount: amounts.subtotal, active: earnBasis === "subtotal" },
      { label: "Eligible", amount: amounts.eligible, active: earnBasis === "eligible" },
      { label: "Tax", amount: amounts.tax, active: false },
    ],
  };
}

function formatMoney(minor) {
  return `$${(minor / 100).toFixed(2)}`;
}
