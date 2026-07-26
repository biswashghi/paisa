import { useState } from "react";
import StatusPill from "./StatusPill.jsx";

export default function Settings({ locations, apiKeys, latestApiToken, connections, onCreateLocation, onCreateApiKey, onStartSquare, onSyncConnection }) {
  const [location, setLocation] = useState({ name: "Main counter", address: "", timezone: "America/Detroit" });
  const [keyName, setKeyName] = useState("Cashier tablet key");

  return (
    <section className="view-stack">
      <div className="view-header">
        <div>
          <p className="eyebrow">Settings</p>
          <h2>Configuration</h2>
        </div>
      </div>

      <section className="panel spacious">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Locations</p>
            <h3>Checkout locations</h3>
          </div>
        </div>
        <div className="form-grid">
          <label>Name<input value={location.name} onChange={(event) => setLocation({ ...location, name: event.target.value })} /></label>
          <label>Address<input value={location.address} onChange={(event) => setLocation({ ...location, address: event.target.value })} /></label>
          <label>Timezone<input value={location.timezone} onChange={(event) => setLocation({ ...location, timezone: event.target.value })} /></label>
        </div>
        <button className="primary" type="button" onClick={() => onCreateLocation(location)}>Save location</button>
        <div className="card-grid compact-cards">
          {locations.map((item) => (
            <article className="mini-card" key={item.id}>
              <div><strong>{item.name}</strong><StatusPill value={item.status} /></div>
              <span>{item.address || "No address set"}</span>
              <small>{item.timezone}</small>
            </article>
          ))}
        </div>
      </section>

      <section className="panel spacious">
        <div className="section-heading">
          <div>
            <p className="eyebrow">API keys</p>
            <h3>Partner credentials</h3>
          </div>
        </div>
        <div className="form-grid single">
          <label>Key name<input value={keyName} onChange={(event) => setKeyName(event.target.value)} /></label>
        </div>
        <button className="primary" type="button" onClick={() => onCreateApiKey(keyName)}>Create key</button>
        {latestApiToken ? (
          <div className="secret-box">
            <span>Shown once.</span>
            <strong>{latestApiToken}</strong>
          </div>
        ) : null}
        <div className="compact-table settings-key-table">
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
            <p className="eyebrow">Integrations</p>
            <h3>Square import</h3>
          </div>
          <button type="button" onClick={onStartSquare}>Connect Square</button>
        </div>
        <div className="card-grid compact-cards">
          {connections.map((connection) => (
            <article className="mini-card" key={connection.id}>
              <div><strong>{connection.provider}</strong><StatusPill value={connection.status} /></div>
              <span>{connection.externalMerchantId || "No merchant ID"}</span>
              <small>{connection.lastSyncAt ? `Synced ${new Date(connection.lastSyncAt).toLocaleString()}` : "No sync"}</small>
              <button type="button" onClick={() => onSyncConnection(connection.id)}>Sync</button>
            </article>
          ))}
        </div>
      </section>
    </section>
  );
}
