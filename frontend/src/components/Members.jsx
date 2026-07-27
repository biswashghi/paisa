import { useMemo, useState } from "react";
import StatusPill from "./StatusPill.jsx";

export default function Members({ enrollments, programs, transactions, onUpdateEnrollment, onMoveEnrollment, onAssignRulePackage, onRemoveRulePackage }) {
  const [selectedId, setSelectedId] = useState(enrollments[0]?.id || "");
  const [moveProgramId, setMoveProgramId] = useState("");
  const [moveReason, setMoveReason] = useState("");
  const selected = enrollments.find((enrollment) => enrollment.id === selectedId) || enrollments[0];
  if (!selected) {
    return (
      <section className="view-stack">
        <div className="view-header">
          <div>
            <p className="eyebrow">Member rewards profile</p>
            <h2>No members exist for this partner yet.</h2>
          </div>
        </div>
      </section>
    );
  }
  const program = programs.find((item) => item.id === selected.programId);
  const availablePackages = program?.rulePackages || [];
  const selectedAddOns = selected.addOns || [];
  const activePackages = availablePackages.filter((pkg) => selectedAddOns.includes(pkg.id));
  const memberTransactions = useMemo(
    () => transactions.filter((transaction) => transaction.member === selected.member),
    [transactions, selected.member],
  );
  const candidatePackages = availablePackages.filter((pkg) => !selectedAddOns.includes(pkg.id));

  function moveMember() {
    if (!moveProgramId) return;
    onMoveEnrollment(selected.id, moveProgramId, moveReason);
    setMoveProgramId("");
    setMoveReason("");
  }

  return (
    <section className="view-stack">
      <div className="view-header">
        <div>
          <p className="eyebrow">Member rewards profile</p>
          <h2>Move members between programs or layer add-on rules over their current program.</h2>
        </div>
      </div>

      <div className="member-detail-layout">
        <section className="panel list-panel">
          {enrollments.map((enrollment) => (
            <button className={enrollment.id === selected.id ? "list-item selected" : "list-item"} type="button" key={enrollment.id} onClick={() => setSelectedId(enrollment.id)}>
              <strong>{enrollment.member}</strong>
              <span>{programs.find((item) => item.id === enrollment.programId)?.name}</span>
              <small>{(enrollment.addOns || []).length} add-ons / {enrollment.points.toLocaleString()} pts</small>
              <StatusPill value={enrollment.status} />
            </button>
          ))}
        </section>

        <section className="view-stack">
          <section className="panel spacious">
            <div className="member-profile-header">
              <div>
                <p className="eyebrow">Active enrollment</p>
                <h3>{selected.member}</h3>
                {selected.email ? <span>{selected.email}</span> : null}
              </div>
              <StatusPill value={selected.status} />
            </div>
            <div className="summary-grid">
              <div><span>Program</span><strong>{program?.name}</strong></div>
              <div><span>Balance</span><strong>{selected.points.toLocaleString()}</strong></div>
              <div><span>Earned</span><strong>{selected.earnedPoints.toLocaleString()}</strong></div>
              <div><span>Add-ons</span><strong>{activePackages.length}</strong></div>
            </div>
            <p className="helper-copy">Last change: {selected.lastChangeReason || "Initial enrollment"}</p>
            <div className="member-rule-map">
              <article>
                <span>Base program</span>
                <strong>{program?.name}</strong>
                <small>Future transactions evaluate this program first.</small>
              </article>
              <article>
                <span>Active add-ons</span>
                <strong>{activePackages.length ? activePackages.map((pkg) => pkg.name).join(", ") : "No add-ons"}</strong>
                <small>Layered rule packages assigned only to this member.</small>
              </article>
              <article>
                <span>Available packages</span>
                <strong>{candidatePackages.length}</strong>
                <small>Published packages can be added below.</small>
              </article>
            </div>
          </section>

          <section className="panel spacious">
            <div className="section-heading">
              <div>
                <p className="eyebrow">Move program</p>
                <h3>Future transactions use the selected program</h3>
              </div>
            </div>
            <div className="form-grid">
              <label>
                Target program
                <select value={moveProgramId} onChange={(event) => setMoveProgramId(event.target.value)}>
                  <option value="">Select program</option>
                  {programs.filter((item) => item.id !== selected.programId).map((item) => <option value={item.id} key={item.id}>{item.name}</option>)}
                </select>
              </label>
              <label>
                Effective date
                <input type="date" defaultValue={new Date().toISOString().slice(0, 10)} />
              </label>
              <label>
                Reason
                <input value={moveReason} onChange={(event) => setMoveReason(event.target.value)} placeholder="Upgrade, tier change, support action" />
              </label>
            </div>
            <button className="primary" type="button" onClick={moveMember}>Move program</button>
          </section>

          <section className="panel spacious">
            <div className="section-heading">
              <div>
                <p className="eyebrow">Rule add-ons</p>
                <h3>Member-specific packages layered over {program?.name}</h3>
              </div>
            </div>
            <div className="addon-grid">
              {availablePackages.map((pkg) => {
                const active = selectedAddOns.includes(pkg.id);
                return (
                  <article className={active ? "addon-card active" : "addon-card"} key={pkg.id}>
                    <div><strong>{pkg.name}</strong><StatusPill value={pkg.status} /></div>
                    <span>{pkg.description}</span>
                    <small>{pkg.rules.map((rule) => rule.name).join(", ")}</small>
                    {active ? (
                      <button type="button" onClick={() => onRemoveRulePackage(selected.id, pkg.id)}>Remove package</button>
                    ) : (
                      <button type="button" onClick={() => onAssignRulePackage(selected.id, pkg.id)} disabled={pkg.status !== "published"}>Add package</button>
                    )}
                  </article>
                );
              })}
            </div>
          </section>

          <section className="panel spacious">
            <div className="section-heading">
              <div>
                <p className="eyebrow">Earning history</p>
                <h3>Transactions that earned or changed points</h3>
              </div>
            </div>
            <div className="compact-table member-history-table">
              {memberTransactions.map((transaction) => (
                <div className="table-row" key={transaction.id}>
                  <span>{transaction.occurredAt}</span>
                  <span>{transaction.category}</span>
                  <strong className={transaction.points >= 0 ? "good-text" : "bad-text"}>{transaction.points > 0 ? "+" : ""}{transaction.points}</strong>
                  <span>{transaction.ruleSource || "Base program rules"}</span>
                  <StatusPill value={transaction.status} />
                </div>
              ))}
            </div>
          </section>
        </section>
      </div>
    </section>
  );
}
