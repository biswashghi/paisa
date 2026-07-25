import SummaryCard from "./SummaryCard.jsx";
import StatusPill from "./StatusPill.jsx";

export default function Dashboard({ programs, enrollments, transactions, selectedProgramId, onSelectProgram, onCreateProgram, onChangeView }) {
  const selected = programs.find((program) => program.id === selectedProgramId) || programs[0];
  const totalMembers = enrollments.length;
  const activeMembers = enrollments.filter((enrollment) => enrollment.status === "active").length;
  const earnedThisMonth = transactions.filter((transaction) => transaction.points > 0).reduce((sum, transaction) => sum + transaction.points, 0);
  const totalLiability = programs.reduce((sum, program) => sum + program.liabilityPoints, 0);
  const published = programs.filter((program) => program.status === "published").length;
  const pendingEnrollments = enrollments.filter((enrollment) => enrollment.status !== "active").length;
  const membersWithAddOns = enrollments.filter((enrollment) => enrollment.addOns.length > 0).length;
  const membershipByProgram = programs.map((program) => ({
    program,
    count: enrollments.filter((enrollment) => enrollment.programId === program.id).length,
  }));
  const maxMembership = Math.max(...membershipByProgram.map(({ count }) => count), 1);
  const selectedAddOns = selected.rulePackages?.length || 0;

  return (
    <section className="view-stack">
      <div className="hero-panel dashboard-command">
        <div className="hero-copy">
          <p className="eyebrow">Loyalty operating cockpit</p>
          <h2>Manage programs, rules, enrollment, and earned points without leaving the partner workspace.</h2>
          <div className="command-steps">
            <span>Programs</span>
            <span>Base rules</span>
            <span>Add-ons</span>
            <span>Members</span>
            <span>Ledger</span>
          </div>
        </div>
        <div className="hero-actions">
          <button type="button" onClick={() => onChangeView("rules")}>Open Rule Studio</button>
          <button className="primary" type="button" onClick={onCreateProgram}>Create program</button>
        </div>
      </div>

      <div className="summary-grid large">
        <SummaryCard label="Published programs" value={published} detail={`${programs.length} total programs`} tone="teal" />
        <SummaryCard label="Members enrolled" value={totalMembers.toLocaleString()} detail={`${activeMembers} active profiles`} />
        <SummaryCard label="With add-ons" value={membersWithAddOns} detail="Member-level rule packages" tone="blue" />
        <SummaryCard label="Points earned" value={earnedThisMonth.toLocaleString()} detail="Current activity window" tone="green" />
        <SummaryCard label="Liability" value={`$${(totalLiability / 100).toLocaleString()}`} detail="Point-backed estimate" tone="amber" />
        <SummaryCard label="Rule validation" value={`${Math.round(selected.validationScore)}%`} detail={`${selected.name} pass rate`} tone="blue" />
        <SummaryCard label="Needs review" value={pendingEnrollments} detail="Pending enrollments" tone="amber" />
      </div>

      <div className="dashboard-layout">
        <section className="panel spacious portfolio-panel">
          <div className="section-heading">
            <div>
              <p className="eyebrow">Program portfolio</p>
              <h3>Reward programs</h3>
            </div>
          </div>
          <div className="program-strip">
            {programs.map((program) => (
              <button className={program.id === selectedProgramId ? "program-tile selected" : "program-tile"} key={program.id} type="button" onClick={() => onSelectProgram(program.id)}>
                <span>{program.tierCode}</span>
                <strong>{program.name}</strong>
                <small>{program.members.toLocaleString()} partner members</small>
                <StatusPill value={program.status} />
              </button>
            ))}
          </div>
          <div className="distribution-list meter-list">
            {membershipByProgram.map(({ program, count }) => (
              <article key={program.id}>
                <div>
                  <span>{program.name}</span>
                  <strong>{count} demo members</strong>
                </div>
                <span className="meter-track"><span style={{ width: `${Math.max(12, (count / maxMembership) * 100)}%` }} /></span>
              </article>
            ))}
          </div>
        </section>

        <section className="panel spacious rule-health-panel">
          <div className="section-heading">
            <div>
              <p className="eyebrow">Rule health</p>
              <h3>{selected.name}</h3>
            </div>
            <StatusPill value={selected.status} />
          </div>
          <div className="health-grid">
            <div className="score-ring">
              <strong>{selected.validationScore.toFixed(1)}%</strong>
              <span>Validation pass rate</span>
            </div>
            <div className="issue-list">
              <div><span>Healthy rules</span><strong>18</strong></div>
              <div><span>Warnings</span><strong>4</strong></div>
              <div><span>Rule packages</span><strong>{selectedAddOns}</strong></div>
              <button type="button" onClick={() => onChangeView("rules")}>Open Rule Studio</button>
            </div>
          </div>
        </section>
      </div>

      <div className="ops-grid">
        <section className="panel spacious">
          <div className="section-heading">
            <div>
              <p className="eyebrow">Program architecture</p>
              <h3>How member earnings are selected</h3>
            </div>
          </div>
          <div className="architecture-map">
            <article>
              <span>01</span>
              <strong>Active enrollment</strong>
              <small>One base program per member.</small>
            </article>
            <article>
              <span>02</span>
              <strong>Published base rules</strong>
              <small>Program-wide graph evaluates first.</small>
            </article>
            <article>
              <span>03</span>
              <strong>Member add-ons</strong>
              <small>Supplemental packages stack for selected profiles.</small>
            </article>
            <article>
              <span>04</span>
              <strong>Trace + ledger</strong>
              <small>Point source is visible on earned activity.</small>
            </article>
          </div>
        </section>

        <section className="panel spacious action-rail">
          <div>
            <p className="eyebrow">Next best actions</p>
            <h3>Operate the loyalty surface</h3>
          </div>
          <button type="button" onClick={() => onChangeView("programs")}>Review program package assignments</button>
          <button type="button" onClick={() => onChangeView("members")}>Move a member or add a package</button>
          <button type="button" onClick={() => onChangeView("transactions")}>Audit earned activity trace</button>
        </section>
      </div>

      <section className="panel spacious">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Recent earned activity</p>
            <h3>Transactions that changed points</h3>
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
