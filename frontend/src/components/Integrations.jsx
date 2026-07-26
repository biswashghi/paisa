import { useState } from "react";
import StatusPill from "./StatusPill.jsx";

export default function Integrations({ apiKeys, latestApiToken, connections, onCreateApiKey, onStartSquare, onSyncConnection }) {
  const [name, setName] = useState("Cashier tablet key");

  return (
    <section className="view-stack">
      <div className="view-header">
        <div>
          <p className="eyebrow">Integrations</p>
          <h2>Issue API keys and prepare Square import connections.</h2>
        </div>
      </div>

      <section className="panel spacious">
        <div className="section-heading">
          <div>
            <p className="eyebrow">API keys</p>
            <h3>Partner-scoped credentials</h3>
          </div>
        </div>
        <div className="form-grid">
          <label>
            Key name
            <input value={name} onChange={(event) => setName(event.target.value)} />
          </label>
        </div>
        <button className="primary" type="button" onClick={() => onCreateApiKey(name)}>Create API key</button>
        {latestApiToken ? (
          <div className="secret-box">
            <span>Copy now; this token is shown once.</span>
            <strong>{latestApiToken}</strong>
          </div>
        ) : null}
        <div className="compact-table">
          {apiKeys.map((key) => (
            <div className="table-row" key={key.id}>
              <strong>{key.name}</strong>
              <span>{key.keyPrefix}</span>
              <span>{key.scopes?.join(", ")}</span>
              <StatusPill value={key.status} />
            </div>
          ))}
        </div>
      </section>

      <section className="panel spacious">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Square</p>
            <h3>Import-only connection</h3>
          </div>
        </div>
        <div className="button-row">
          <button type="button" onClick={onStartSquare}>Start Square connection</button>
        </div>
        <div className="card-grid">
          {connections.map((connection) => (
            <article className="mini-card" key={connection.id}>
              <div><strong>{connection.provider}</strong><StatusPill value={connection.status} /></div>
              <span>{connection.externalMerchantId || "No merchant ID yet"}</span>
              <small>{connection.lastSyncAt ? `Last synced ${new Date(connection.lastSyncAt).toLocaleString()}` : "No sync yet"}</small>
              <button type="button" onClick={() => onSyncConnection(connection.id)}>Sync</button>
            </article>
          ))}
        </div>
      </section>
    </section>
  );
}
