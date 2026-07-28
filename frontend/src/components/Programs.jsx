import { useEffect, useMemo, useState } from "react";
import Rewards from "./Rewards.jsx";
import RuleStudio from "./RuleStudio.jsx";
import StatusPill from "./StatusPill.jsx";
import { validateProgramRules } from "../utils/rules.js";

const buildTabs = ["Basics", "Rules", "Rewards"];

export default function Programs({
  programs,
  selectedProgramId,
  enrollments,
  catalogItems,
  redemptions,
  onSelectProgram,
  onCreateProgram,
  onUpdateProgram,
  onSaveProgramDetails,
  onDeleteDraftProgram,
  onCreateRulePackage,
  onUpdateRulePackage,
  onPublishProgramRules,
  onPublishRulePackage,
  onCreateCatalogItem,
}) {
  const [activeTab, setActiveTab] = useState("Basics");
  const selected = programs.find((program) => program.id === selectedProgramId) || programs[0];

  useEffect(() => {
    setActiveTab("Basics");
  }, [selectedProgramId]);

  async function createDraftProgram() {
    setActiveTab("Basics");
    await onCreateProgram();
  }

  if (!selected) {
    return (
      <section className="view-stack">
        <div className="view-header">
          <div>
            <p className="eyebrow">Programs</p>
            <h2>Create the first program</h2>
          </div>
          <button className="primary" type="button" onClick={createDraftProgram}>Create draft program</button>
        </div>
      </section>
    );
  }

  const programEnrollments = enrollments.filter((enrollment) => enrollment.programId === selected.id);
  const assignedPackages = (selected.rulePackages || []).reduce((sum, pkg) => (
    sum + enrollments.filter((enrollment) => (enrollment.addOns || []).includes(pkg.id)).length
  ), 0);
  const ruleIssues = validateProgramRules(selected);
  const rulesReady = ruleIssues.length === 0 && Boolean(selected.rules?.groups?.some((group) => group.rules?.length));
  const rewardsReady = catalogItems.some((item) => item.status === "active" && (!item.programId || item.programId === selected.id));
  const canPublish = selected.status !== "published" && rulesReady && rewardsReady;
  const canDeleteDraft = selected.status !== "published" && programEnrollments.length === 0;
  const setupSteps = [
    { label: "Basics", done: Boolean(selected.name && selected.tierCode), tab: "Basics" },
    { label: "Rules", done: rulesReady, tab: "Rules" },
    { label: "Rewards", done: rewardsReady, tab: "Rewards" },
  ];

  function publishProgram() {
    if (!canPublish) return;
    onPublishProgramRules(selected.id, selected);
  }

  function deleteDraft() {
    if (!canDeleteDraft) return;
    if (window.confirm(`Delete draft program "${selected.name}"? This removes its draft rules and program rewards.`)) {
      onDeleteDraftProgram(selected.id);
    }
  }

  return (
    <section className="view-stack">
      <div className="view-header">
        <div>
          <p className="eyebrow">Programs</p>
          <h2>Program builder</h2>
        </div>
        <button className="primary" type="button" onClick={createDraftProgram}>Create draft program</button>
      </div>

      <section className="panel program-tab-strip" aria-label="Programs">
        {programs.map((program) => (
          <button className={program.id === selected.id ? "program-tab selected" : "program-tab"} key={program.id} type="button" onClick={() => onSelectProgram(program.id)}>
            <strong>{program.name}</strong>
            <span>{program.tierCode || "no tier"} / {enrollments.filter((enrollment) => enrollment.programId === program.id).length} active members</span>
            <StatusPill value={program.status} />
          </button>
        ))}
        <button className="program-tab create-tab" type="button" onClick={createDraftProgram}>
          <strong>+ New draft</strong>
          <span>Add another program and configure it here.</span>
        </button>
      </section>

      <section className="panel spacious">
        <div className="program-command-card">
          <div>
            <p className="eyebrow">Selected draft</p>
            <h3>{selected.name}</h3>
            <small>{selected.tierCode || "no tier"} / {programEnrollments.length} active enrollments / {assignedPackages} add-on assignments</small>
          </div>
          <div className="program-command-actions">
            <StatusPill value={selected.status} />
            {canDeleteDraft ? <button type="button" onClick={deleteDraft}>Delete draft</button> : null}
          </div>
        </div>

        <div className="program-builder-grid">
          <aside className="program-build-steps" aria-label="Program build steps">
            {setupSteps.map((step, index) => (
              <button className={activeTab === step.tab ? "active" : step.done ? "done" : ""} key={step.label} type="button" onClick={() => setActiveTab(step.tab)}>
                <span>{step.done ? "✓" : String(index + 1).padStart(2, "0")}</span>
                <strong>{step.label}</strong>
                <small>{step.done ? "Complete" : step.tab === "Basics" ? "Name and tier" : step.tab === "Rules" ? "Create earn logic" : "Add redeemable item"}</small>
              </button>
            ))}
          </aside>

          <main className="program-build-main">
            {activeTab === "Basics" ? (
              <ProgramBasics selected={selected} onSaveProgramDetails={onSaveProgramDetails} onContinue={() => setActiveTab("Rules")} />
            ) : null}

            {activeTab === "Rules" ? (
              <section className="program-build-panel">
                <div className="section-heading">
                  <div>
                    <p className="eyebrow">Step 2</p>
                    <h3>Create earning rules</h3>
                  </div>
                  <StatusPill value={rulesReady ? "ready" : "todo"} />
                </div>
                <RuleStudio
                  program={selected}
                  embedded
                  showPublish={false}
                  onUpdateProgram={onUpdateProgram}
                  onPublishProgramRules={onPublishProgramRules}
                  onCreateRulePackage={onCreateRulePackage}
                  onUpdateRulePackage={onUpdateRulePackage}
                  onPublishRulePackage={onPublishRulePackage}
                />
                <div className="program-step-footer">
                  <div>
                    <strong>{rulesReady ? "Rules are ready." : "Rules need changes."}</strong>
                    <span>{rulesReady ? "Next, create the first reward customers can redeem." : ruleIssues.join(" ")}</span>
                  </div>
                  <button className="primary" type="button" disabled={!rulesReady} onClick={() => setActiveTab("Rewards")}>Continue to rewards</button>
                </div>
              </section>
            ) : null}

            {activeTab === "Rewards" ? (
              <section className="program-build-panel">
                <div className="section-heading">
                  <div>
                    <p className="eyebrow">Step 3</p>
                    <h3>Create rewards</h3>
                  </div>
                  <StatusPill value={rewardsReady ? "ready" : "todo"} />
                </div>
                <Rewards
                  catalogItems={catalogItems}
                  redemptions={redemptions}
                  programs={programs}
                  selectedProgramId={selected.id}
                  setupMode
                  onCreateCatalogItem={onCreateCatalogItem}
                  embedded
                />
              </section>
            ) : null}
          </main>
        </div>

        {canPublish ? (
          <div className="program-publish-bar">
            <div>
              <p className="eyebrow">Ready</p>
              <strong>Rules and rewards are complete.</strong>
              <span>Publish this program to make it available for member enrollment and earning.</span>
            </div>
            <button className="primary" type="button" onClick={publishProgram}>Publish program</button>
          </div>
        ) : (
          <div className="program-publish-bar muted">
            <div>
              <p className="eyebrow">Publish locked</p>
              <strong>Complete rules and rewards first.</strong>
              <span>Enrollment and activity live in their own sidebar views after the program is published.</span>
            </div>
          </div>
        )}
      </section>
    </section>
  );
}

function ProgramBasics({ selected, onSaveProgramDetails, onContinue }) {
  const [name, setName] = useState(selected.name);
  const [tierCode, setTierCode] = useState(selected.tierCode);
  const basicsReady = Boolean(name && tierCode);

  useEffect(() => {
    setName(selected.name);
    setTierCode(selected.tierCode);
  }, [selected.id, selected.name, selected.tierCode]);

  async function saveDetails() {
    if (!basicsReady) return;
    await onSaveProgramDetails(selected.id, { name, tierCode });
  }

  return (
    <section className="program-build-panel">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Step 1</p>
          <h3>Edit program basics</h3>
        </div>
        <StatusPill value={basicsReady ? "ready" : "draft"} />
      </div>
      <div className="form-grid">
        <label>Program name<input value={name} onChange={(event) => setName(event.target.value)} /></label>
        <label>Tier code<input value={tierCode} onChange={(event) => setTierCode(event.target.value)} /></label>
      </div>
      <div className="program-step-footer">
        <div>
          <strong>Start as draft.</strong>
          <span>This program will not be treated as published until rules and rewards are complete.</span>
        </div>
        <div className="button-row">
          <button type="button" disabled={!basicsReady} onClick={saveDetails}>Save details</button>
          <button className="primary" type="button" disabled={!basicsReady} onClick={onContinue}>Continue to rules</button>
        </div>
      </div>
    </section>
  );
}
