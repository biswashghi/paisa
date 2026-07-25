import { useState } from "react";
import StatusPill from "./StatusPill.jsx";

const tabs = ["Base Rules", "Rule Packages", "Enrollments", "Earned Activity"];

export default function Programs({ programs, selectedProgramId, enrollments, transactions, onSelectProgram, onCreateProgram, onUpdateProgram, onCreateRulePackage }) {
  const [activeTab, setActiveTab] = useState("Base Rules");
  const selected = programs.find((program) => program.id === selectedProgramId) || programs[0];
  const programEnrollments = enrollments.filter((enrollment) => enrollment.programId === selected.id);
  const programTransactions = transactions.filter((transaction) => transaction.programId === selected.id);
  const assignedPackages = (selected.rulePackages || []).reduce((sum, pkg) => (
    sum + enrollments.filter((enrollment) => enrollment.addOns.includes(pkg.id)).length
  ), 0);

  return (
    <section className="view-stack">
      <div className="view-header">
        <div>
          <p className="eyebrow">Program management</p>
          <h2>Programs connect base rules, add-on packages, enrolled members, and earned activity.</h2>
        </div>
        <button className="primary" type="button" onClick={onCreateProgram}>Create program</button>
      </div>

      <div className="split-layout">
        <section className="panel list-panel">
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

          {activeTab === "Base Rules" ? (
            <div className="rule-summary-list rule-card-grid">
              {selected.rules.groups.map((group) => (
                <article key={group.id}>
                  <div><strong>{group.name}</strong><StatusPill value={group.status} /></div>
                  <span>Strategy: {group.strategy}</span>
                  <small>{group.rules.map((rule) => `${rule.name}: ${rule.type === "points_per_dollar" ? `${rule.pointsPerDollar} pt / $` : `${rule.points} pts`}`).join(" | ")}</small>
                </article>
              ))}
            </div>
          ) : null}

          {activeTab === "Rule Packages" ? (
            <div className="rule-summary-list">
              <div className="package-toolbar">
                <div>
                  <strong>Reusable member add-ons</strong>
                  <span>Publish packages here, then assign them from member detail.</span>
                </div>
                <button type="button" onClick={() => onCreateRulePackage(selected.id)}>Create member add-on package</button>
              </div>
              <div className="package-board">
                {(selected.rulePackages || []).map((pkg) => (
                  <article key={pkg.id}>
                    <div><strong>{pkg.name}</strong><StatusPill value={pkg.status} /></div>
                    <span>{pkg.description}</span>
                    <small>{pkg.rules.length} rules / Assigned to {enrollments.filter((enrollment) => enrollment.addOns.includes(pkg.id)).length} members</small>
                  </article>
                ))}
              </div>
            </div>
          ) : null}

          {activeTab === "Enrollments" ? (
            <div className="compact-table program-enrollment-table">
              {programEnrollments.map((enrollment) => (
                <div className="table-row" key={enrollment.id}>
                  <strong>{enrollment.member}</strong>
                  <span>{enrollment.email}</span>
                  <span>{enrollment.points.toLocaleString()} pts</span>
                  <span>{enrollment.addOns.length} add-ons</span>
                  <StatusPill value={enrollment.status} />
                </div>
              ))}
            </div>
          ) : null}

          {activeTab === "Earned Activity" ? (
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
      </div>
    </section>
  );
}
