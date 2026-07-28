import { useEffect, useState } from "react";
import StatusPill from "./StatusPill.jsx";

export default function Rewards({ catalogItems, redemptions, programs, onCreateCatalogItem, embedded = false, selectedProgramId = "", setupMode = false }) {
  const [form, setForm] = useState({ name: "Free coffee", pointsCost: 100, description: "Manual POS discount after Paisa validation.", programId: selectedProgramId });
  const selectedProgram = programs.find((program) => program.id === selectedProgramId);

  useEffect(() => {
    if (selectedProgramId) {
      setForm((current) => ({ ...current, programId: selectedProgramId }));
    }
  }, [selectedProgramId]);

  function submit() {
    onCreateCatalogItem({
      ...form,
      pointsCost: Number(form.pointsCost),
      rewardType: "manual_discount",
      status: "active",
      expiresAfterMinutes: 15,
    });
  }

  return (
    <section className="view-stack">
      {!embedded ? (
        <div className="view-header">
          <div>
            <p className="eyebrow">Rewards</p>
            <h2>Catalog</h2>
          </div>
        </div>
      ) : null}
      <section className="panel spacious">
        {setupMode ? (
          <div className="setup-inline-heading">
            <div>
              <p className="eyebrow">Reward draft</p>
              <h3>Create one redeemable reward</h3>
              <span>The program can be published after it has earning rules and at least one active reward.</span>
            </div>
            {selectedProgram ? <StatusPill value={selectedProgram.status} /> : null}
          </div>
        ) : null}
        <div className="form-grid">
          <label>
            Name
            <input value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} />
          </label>
          <label>
            Cost
            <input value={form.pointsCost} onChange={(event) => setForm({ ...form, pointsCost: event.target.value })} inputMode="numeric" />
          </label>
          <label>
            Program
            <select value={form.programId} onChange={(event) => setForm({ ...form, programId: event.target.value })}>
              <option value="">Any active member</option>
              {programs.map((program) => <option key={program.id} value={program.id}>{program.name}</option>)}
            </select>
          </label>
          <label>
            Description
            <input value={form.description} onChange={(event) => setForm({ ...form, description: event.target.value })} />
          </label>
        </div>
        <button className="primary" type="button" onClick={submit}>Add reward</button>
      </section>

      <section className="panel spacious">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Items</p>
            <h3>{catalogItems.length} rewards</h3>
          </div>
        </div>
        <div className="card-grid">
          {catalogItems.map((item) => (
            <article className="mini-card" key={item.id}>
              <div><strong>{item.name}</strong><StatusPill value={item.status} /></div>
              <span>{item.description || "Manual discount reward"}</span>
              <small>{item.pointsCost} pts / expires after {item.expiresAfterMinutes} min</small>
            </article>
          ))}
        </div>
      </section>

      <section className="panel spacious">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Redemptions</p>
            <h3>Recent</h3>
          </div>
        </div>
        <div className="compact-table redemption-table">
          {redemptions.map((redemption) => (
            <div className="table-row" key={redemption.id}>
              <strong>{redemption.catalogItemName || redemption.catalogItemId}</strong>
              <span>{redemption.code}</span>
              <span>{redemption.pointsCost} pts</span>
              <StatusPill value={redemption.status} />
            </div>
          ))}
        </div>
      </section>

    </section>
  );
}
