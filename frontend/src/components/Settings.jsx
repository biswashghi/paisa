import { useState } from "react";
import StatusPill from "./StatusPill.jsx";

export default function Settings({ apiKeys, latestApiToken, onCreateApiKey }) {
  const [keyName, setKeyName] = useState("Partner API key");

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
            <p className="eyebrow">API keys</p>
            <h3>Access keys</h3>
          </div>
        </div>
        <div className="form-grid single">
          <label>Key name<input value={keyName} onChange={(event) => setKeyName(event.target.value)} /></label>
        </div>
        <button className="primary" type="button" onClick={() => onCreateApiKey(keyName)}>Create key</button>
        {latestApiToken ? (
          <div className="secret-box">
            <span>Shown once</span>
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

    </section>
  );
}
