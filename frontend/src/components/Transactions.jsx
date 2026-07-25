import SummaryCard from "./SummaryCard.jsx";
import StatusPill from "./StatusPill.jsx";

export default function Transactions({ transactions, programs }) {
  const earned = transactions.filter((transaction) => transaction.points > 0).reduce((sum, transaction) => sum + transaction.points, 0);
  const burned = transactions.filter((transaction) => transaction.points < 0).reduce((sum, transaction) => sum + Math.abs(transaction.points), 0);
  const posted = transactions.filter((transaction) => transaction.status === "posted").length;

  return (
    <section className="view-stack">
      <div className="view-header">
        <div>
          <p className="eyebrow">Earned activity</p>
          <h2>Transactions that changed points</h2>
        </div>
      </div>
      <div className="summary-grid">
        <SummaryCard label="Points earned" value={earned.toLocaleString()} detail="Earn events" tone="green" />
        <SummaryCard label="Points burned" value={burned.toLocaleString()} detail="Burn and reversal events" tone="amber" />
        <SummaryCard label="Posted events" value={posted} detail={`${transactions.length} total`} tone="blue" />
      </div>
      <section className="panel spacious">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Activity ledger</p>
            <h3>Compact event stream</h3>
          </div>
        </div>
        <div className="compact-table activity-table">
          <div className="table-row table-head">
            <span>Time</span>
            <span>Member</span>
            <span>Program</span>
            <span>Type</span>
            <span>Category</span>
            <span>Points</span>
            <span>Rule source</span>
            <span>Status</span>
          </div>
          {transactions.map((transaction) => (
            <div className="table-row" key={transaction.id}>
              <span>{transaction.occurredAt}</span>
              <strong>{transaction.member}</strong>
              <span>{programs.find((program) => program.id === transaction.programId)?.name}</span>
              <span>{transaction.type}</span>
              <span>{transaction.category}</span>
              <strong className={transaction.points >= 0 ? "good-text" : "bad-text"}>{transaction.points > 0 ? "+" : ""}{transaction.points}</strong>
              <span>{transaction.ruleSource || "Base program rules"}</span>
              <StatusPill value={transaction.status} />
            </div>
          ))}
        </div>
      </section>
    </section>
  );
}
