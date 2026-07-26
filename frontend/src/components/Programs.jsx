import { useState } from "react";
import Rewards from "./Rewards.jsx";
import RuleStudio from "./RuleStudio.jsx";
import StatusPill from "./StatusPill.jsx";

const tabs = ["Overview", "Rules", "Rewards", "Enrollments", "Activity"];

export default function Programs({
  programs,
  selectedProgramId,
  enrollments,
  transactions,
  catalogItems,
  redemptions,
  campaigns,
  onSelectProgram,
  onCreateProgram,
  onUpdateProgram,
  onCreateRulePackage,
  onUpdateRulePackage,
  onPublishProgramRules,
  onPublishRulePackage,
  onCreateCatalogItem,
  onCreateCampaign,
}) {
  const [activeTab, setActiveTab] = useState("Overview");
  const selected = programs.find((program) => program.id === selectedProgramId) || programs[0];
  if (!selected) {
    return (
      <section className="view-stack">
        <div className="view-header">
          <div>
            <p className="eyebrow">Program management</p>
            <h2>No programs exist for this partner yet.</h2>
          </div>
          <button className="primary" type="button" onClick={onCreateProgram}>Create program</button>
        </div>
      </section>
    );
  }
  const programEnrollments = enrollments.filter((enrollment) => enrollment.programId === selected.id);
  const programTransactions = transactions.filter((transaction) => transaction.programId === selected.id);
  const assignedPackages = (selected.rulePackages || []).reduce((sum, pkg) => (
    sum + enrollments.filter((enrollment) => (enrollment.addOns || []).includes(pkg.id)).length
  ), 0);

  return (
    <section className="view-stack">
      <div className="view-header">
        <div>
          <p className="eyebrow">Programs</p>
          <h2>Program workspace</h2>
        </div>
        <button className="primary" type="button" onClick={onCreateProgram}>Create program</button>
      </div>

      <section className="panel program-selector-strip">
        {programs.map((program) => (
          <button className={program.id === selectedProgramId ? "list-item selected" : "list-item"} key={program.id} type="button" onClick={() => onSelectProgram(program.id)}>
            <strong>{program.name}</strong>
            <span>{program.tierCode} / {enrollments.filter((enrollment) => enrollment.programId === program.id).length} active demo members</span>
            <StatusPill value={program.status} />
          </button>
        ))}
      </section>

      <section className="panel spacious">
          <div className="program-command-card">
            <div>
              <p className="eyebrow">Selected program</p>
              <h3>{selected.name}</h3>
              <small>{selected.tierCode} / {programEnrollments.length} active demo enrollments / {assignedPackages} add-on assignments</small>
            </div>
            <StatusPill value={selected.status} />
          </div>
          <div className="program-stat-row">
            <div><span>Validation</span><strong>{selected.validationScore.toFixed(1)}%</strong></div>
            <div><span>Liability</span><strong>${(selected.liabilityPoints / 100).toLocaleString()}</strong></div>
            <div><span>Packages</span><strong>{selected.rulePackages?.length || 0}</strong></div>
          </div>
          <div className="form-grid">
            <label>Program name<input value={selected.name} onChange={(event) => onUpdateProgram(selected.id, { name: event.target.value })} /></label>
            <label>Tier code<input value={selected.tierCode} onChange={(event) => onUpdateProgram(selected.id, { tierCode: event.target.value })} /></label>
            <label>
              Status
              <select value={selected.status} onChange={(event) => onUpdateProgram(selected.id, { status: event.target.value })}>
                <option value="draft">Draft</option>
                <option value="published">Published</option>
                <option value="paused">Paused</option>
              </select>
            </label>
          </div>

          <div className="tab-row">
            {tabs.map((tab) => (
              <button className={activeTab === tab ? "active" : ""} type="button" key={tab} onClick={() => setActiveTab(tab)}>{tab}</button>
            ))}
          </div>

          {activeTab === "Overview" ? (
            <div className="program-overview-grid">
              <div className="rule-summary-list rule-card-grid">
                {selected.rules.groups.map((group) => (
                  <article key={group.id}>
                    <div><strong>{group.name}</strong><StatusPill value={group.status} /></div>
                    <span>Strategy: {group.strategy}</span>
                    <small>{group.rules.map((rule) => `${rule.name}: ${rule.type === "points_per_dollar" ? `${rule.pointsPerDollar} pt / $` : `${rule.points} pts`}`).join(" | ")}</small>
                  </article>
                ))}
              </div>
              <div className="package-toolbar">
                <div>
                  <strong>Add-on packages</strong>
                  <span>Assign published packages from member detail.</span>
                </div>
                <button type="button" onClick={() => onCreateRulePackage(selected.id)}>Create member add-on package</button>
              </div>
              <div className="package-board">
                {(selected.rulePackages || []).map((pkg) => (
                  <article key={pkg.id}>
                    <div><strong>{pkg.name}</strong><StatusPill value={pkg.status} /></div>
                    <span>{pkg.description}</span>
                  <small>{pkg.rules.length} rules / Assigned to {enrollments.filter((enrollment) => (enrollment.addOns || []).includes(pkg.id)).length} members</small>
                  </article>
                ))}
              </div>
            </div>
          ) : null}

          {activeTab === "Rules" ? (
            <RuleStudio
              program={selected}
              embedded
              onUpdateProgram={onUpdateProgram}
              onPublishProgramRules={onPublishProgramRules}
              onCreateRulePackage={onCreateRulePackage}
              onUpdateRulePackage={onUpdateRulePackage}
              onPublishRulePackage={onPublishRulePackage}
            />
          ) : null}

          {activeTab === "Rewards" ? (
            <Rewards
              catalogItems={catalogItems}
              redemptions={redemptions}
              campaigns={campaigns}
              programs={programs}
              onCreateCatalogItem={onCreateCatalogItem}
              onCreateCampaign={onCreateCampaign}
              embedded
            />
          ) : null}

          {activeTab === "Enrollments" ? (
            <div className="compact-table program-enrollment-table">
              {programEnrollments.map((enrollment) => (
                <div className="table-row" key={enrollment.id}>
                  <strong>{enrollment.member}</strong>
                  <span>{enrollment.email}</span>
                  <span>{enrollment.points.toLocaleString()} pts</span>
                  <span>{(enrollment.addOns || []).length} add-ons</span>
                  <StatusPill value={enrollment.status} />
                </div>
              ))}
            </div>
          ) : null}

          {activeTab === "Activity" ? (
            <div className="compact-table program-activity-table">
              {programTransactions.map((transaction) => (
                <div className="table-row" key={transaction.id}>
                  <span>{transaction.occurredAt}</span>
                  <strong>{transaction.member}</strong>
                  <span>{transaction.category}</span>
                  <strong className={transaction.points >= 0 ? "good-text" : "bad-text"}>{transaction.points > 0 ? "+" : ""}{transaction.points}</strong>
                  <span>{transaction.ruleSource || "Base program rules"}</span>
                  <StatusPill value={transaction.status} />
                </div>
              ))}
            </div>
          ) : null}
      </section>
    </section>
  );
}
