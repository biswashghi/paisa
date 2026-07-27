import SummaryCard from "./SummaryCard.jsx";
import StatusPill from "./StatusPill.jsx";

export default function Dashboard({ programs, enrollments, transactions, dashboardSummary, selectedProgramId, onSelectProgram, onCreateProgram, onChangeView }) {
  const selected = programs.find((program) => program.id === selectedProgramId) || programs[0] || emptyProgram;
  const totalMembers = enrollments.length;
  const activeMembers = enrollments.filter((enrollment) => enrollment.status === "active").length;
  const earnedThisMonth = transactions.filter((transaction) => transaction.points > 0).reduce((sum, transaction) => sum + transaction.points, 0);
  const totalLiability = programs.reduce((sum, program) => sum + program.liabilityPoints, 0);
  const published = programs.filter((program) => program.status === "published").length;
  const activeCatalogItems = dashboardSummary?.activeCatalogItems || 0;
  const pendingEnrollments = enrollments.filter((enrollment) => enrollment.status !== "active").length;
  const membersWithAddOns = enrollments.filter((enrollment) => (enrollment.addOns || []).length > 0).length;
  const membershipByProgram = programs.map((program) => ({
    program,
    count: enrollments.filter((enrollment) => enrollment.programId === program.id).length,
  }));
  const maxMembership = Math.max(...membershipByProgram.map(({ count }) => count), 1);
  const selectedAddOns = selected.rulePackages?.length || 0;
  const selectedRules = selected.rules?.groups?.flatMap((group) => group.rules || []) || [];
  const healthyRules = selectedRules.filter((rule) => rule.status === "active").length;
  const ruleWarnings = selected.validationScore >= 100 ? 0 : Math.max(1, selectedRules.length - healthyRules);

  if (programs.length === 0) {
    return (
      <section className="view-stack">
        <LaunchChecklist
          partnerKey={dashboardSummary?.partner?.partnerKey}
          programs={programs}
          published={published}
          activeCatalogItems={activeCatalogItems}
          transactions={transactions}
          onCreateProgram={onCreateProgram}
          onChangeView={onChangeView}
        />
      </section>
    );
  }

  return (
    <section className="view-stack">
      {published === 0 || activeCatalogItems === 0 || transactions.length === 0 ? (
        <LaunchChecklist
          partnerKey={dashboardSummary?.partner?.partnerKey}
          programs={programs}
          published={published}
          activeCatalogItems={activeCatalogItems}
          transactions={transactions}
          onCreateProgram={onCreateProgram}
          onChangeView={onChangeView}
        />
      ) : null}

      <div className="dashboard-command-bar">
        <div>
          <p className="eyebrow">Dashboard</p>
          <h2>Status</h2>
        </div>
        <div className="command-flow">
          <span>Purchase</span>
          <span>Rules</span>
          <span>Ledger</span>
        </div>
        <div className="dashboard-actions">
          <button type="button" onClick={() => onChangeView("programs")}>Rules</button>
          <button className="primary" type="button" onClick={onCreateProgram}>Create program</button>
        </div>
      </div>

      <div className="summary-grid large">
        <SummaryCard label="Published" value={published} detail={`${programs.length} programs`} tone="teal" />
        <SummaryCard label="Members" value={totalMembers.toLocaleString()} detail={`${activeMembers} active`} />
        <SummaryCard label="Add-ons" value={membersWithAddOns} detail="Assigned members" tone="blue" />
        <SummaryCard label="Earned" value={earnedThisMonth.toLocaleString()} detail="Points" tone="green" />
        <SummaryCard label="Liability" value={`$${(totalLiability / 100).toLocaleString()}`} detail="Point-backed estimate" tone="amber" />
        <SummaryCard label="Rules" value={`${Math.round(selected.validationScore)}%`} detail={selected.name} tone="blue" />
        <SummaryCard label="Review" value={pendingEnrollments} detail="Pending" tone="amber" />
      </div>

      <div className="dashboard-layout">
        <section className="panel spacious portfolio-panel">
          <div className="section-heading">
            <div>
              <p className="eyebrow">Programs</p>
              <h3>Programs</h3>
            </div>
          </div>
          <div className="program-strip">
            {programs.map((program) => (
              <button className={program.id === selectedProgramId ? "program-tile selected" : "program-tile"} key={program.id} type="button" onClick={() => onSelectProgram(program.id)}>
                <span>{program.tierCode}</span>
                <strong>{program.name}</strong>
                <small>{program.members.toLocaleString()} members</small>
                <StatusPill value={program.status} />
              </button>
            ))}
          </div>
          <div className="distribution-list meter-list">
            {membershipByProgram.map(({ program, count }) => (
              <article key={program.id}>
                <div>
                  <span>{program.name}</span>
                  <strong>{count} members</strong>
                </div>
                <span className="meter-track"><span style={{ width: `${Math.max(12, (count / maxMembership) * 100)}%` }} /></span>
              </article>
            ))}
          </div>
        </section>

        <section className="panel spacious rule-health-panel">
          <div className="section-heading">
            <div>
              <p className="eyebrow">Rules</p>
              <h3>{selected.name}</h3>
            </div>
            <StatusPill value={selected.status} />
          </div>
          <div className="health-grid">
            <div className="score-ring">
              <strong>{selected.validationScore.toFixed(1)}%</strong>
              <span>Validation</span>
            </div>
            <div className="issue-list">
              <div><span>Active</span><strong>{healthyRules}</strong></div>
              <div><span>Warnings</span><strong>{ruleWarnings}</strong></div>
              <div><span>Packages</span><strong>{selectedAddOns}</strong></div>
              <button type="button" onClick={() => onChangeView("programs")}>Edit rules</button>
            </div>
          </div>
        </section>
      </div>

      <section className="panel spacious">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Activity</p>
            <h3>Recent transactions</h3>
          </div>
        </div>
        <div className="compact-table">
          {transactions.slice(0, 5).map((transaction) => (
            <div className="table-row" key={transaction.id}>
              <span>{transaction.occurredAt}</span>
              <strong>{transaction.member}</strong>
              <span>{programs.find((program) => program.id === transaction.programId)?.name}</span>
              <span>{transaction.category}</span>
              <strong className={transaction.points >= 0 ? "good-text" : "bad-text"}>{transaction.points > 0 ? "+" : ""}{transaction.points}</strong>
              <StatusPill value={transaction.status} />
            </div>
          ))}
        </div>
      </section>
    </section>
  );
}

function LaunchChecklist({ partnerKey, programs, published, activeCatalogItems, transactions, onCreateProgram, onChangeView }) {
  const steps = [
    {
      label: "Program",
      detail: programs.length ? `${programs.length} configured` : "Create the first program.",
      done: programs.length > 0,
      action: "Create program",
      onClick: onCreateProgram,
    },
    {
      label: "Rules",
      detail: published ? `${published} published` : "Publish earning rules.",
      done: published > 0,
      action: "Edit rules",
      onClick: () => onChangeView("programs"),
    },
    {
      label: "Reward",
      detail: activeCatalogItems ? `${activeCatalogItems} active` : "Add a reward.",
      done: activeCatalogItems > 0,
      action: "Add reward",
      onClick: () => onChangeView("programs"),
    },
    {
      label: "Test",
      detail: transactions.length ? `${transactions.length} recorded` : "Run checkout.",
      done: transactions.length > 0,
      action: "Open cashier",
      onClick: () => onChangeView("setup"),
    },
  ];
  const next = steps.find((step) => !step.done);

  return (
    <section className="launch-panel">
      <div className="launch-copy">
        <p className="eyebrow">{partnerKey || "Partner"}</p>
        <h2>Launch checklist</h2>
      </div>
      <div className="launch-steps">
        {steps.map((step, index) => (
          <article className={step.done ? "launch-step done" : "launch-step"} key={step.label}>
            <span>{String(index + 1).padStart(2, "0")}</span>
            <div>
              <strong>{step.label}</strong>
              <small>{step.detail}</small>
            </div>
            <StatusPill value={step.done ? "ready" : "todo"} />
          </article>
        ))}
      </div>
      <div className="launch-actions">
        <button className="primary" type="button" onClick={next?.onClick || (() => onChangeView("setup"))}>
          {next?.action || "Open setup"}
        </button>
      </div>
    </section>
  );
}

const emptyProgram = {
  id: "",
  name: "No program selected",
  tierCode: "",
  status: "draft",
  validationScore: 0,
  rulePackages: [],
};
