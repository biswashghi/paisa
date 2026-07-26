import { useState } from "react";
import StatusPill from "./StatusPill.jsx";

export default function Campaigns({ campaigns, catalogItems, onCreateCampaign, embedded = false }) {
  const [form, setForm] = useState({ name: "Downtown Passport", description: "Visit three local partners to unlock a reward.", requiredVisitCount: 3, rewardCatalogItemId: "" });

  function submit() {
    onCreateCampaign({
      ...form,
      status: "draft",
      requiredVisitCount: Number(form.requiredVisitCount || 3),
      metadata: { campaignType: "local_passport" },
    });
  }

  return (
    <section className="view-stack">
      {!embedded ? (
        <div className="view-header">
          <div>
            <p className="eyebrow">Campaigns</p>
            <h2>Local passport campaigns</h2>
          </div>
        </div>
      ) : null}
      <section className="panel spacious">
        <div className="form-grid">
          <label>
            Campaign name
            <input value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} />
          </label>
          <label>
            Visits required
            <input value={form.requiredVisitCount} onChange={(event) => setForm({ ...form, requiredVisitCount: event.target.value })} inputMode="numeric" />
          </label>
          <label>
            Reward
            <select value={form.rewardCatalogItemId} onChange={(event) => setForm({ ...form, rewardCatalogItemId: event.target.value })}>
              <option value="">No reward selected</option>
              {catalogItems.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
            </select>
          </label>
          <label>
            Description
            <input value={form.description} onChange={(event) => setForm({ ...form, description: event.target.value })} />
          </label>
        </div>
        <button className="primary" type="button" onClick={submit}>Create campaign</button>
      </section>
      <section className="panel spacious">
        <div className="card-grid">
          {campaigns.map((campaign) => (
            <article className="mini-card" key={campaign.id}>
              <div><strong>{campaign.name}</strong><StatusPill value={campaign.status} /></div>
              <span>{campaign.description}</span>
              <small>{campaign.requiredVisitCount} visits required</small>
            </article>
          ))}
        </div>
      </section>
    </section>
  );
}
