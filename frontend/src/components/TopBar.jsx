export default function TopBar({ partner, selectedProgram, apiBaseUrl, loading, onRefresh, onLogout }) {
  return (
    <header className="topbar">
      <div className="topbar-title">
        <p className="eyebrow">Partner workspace</p>
        <h1>{partner.name}</h1>
      </div>
      <div className="context-strip">
        <label className="command-search">
          <span>Command</span>
          <input readOnly value="Search members, programs, rule packages" />
        </label>
        <div>
          <span>Program</span>
          <strong>{selectedProgram?.name || "No program selected"}</strong>
        </div>
        <div>
          <span>API Environment</span>
          <strong>{partner.apiEnvironment}</strong>
        </div>
        <div>
          <span>API Status</span>
          <strong className="good-text">{loading ? "Syncing" : "Connected"}</strong>
        </div>
        <div>
          <span>Backend</span>
          <strong>{apiBaseUrl.replace(/^https?:\/\//, "")}</strong>
        </div>
        <button type="button" onClick={onRefresh} disabled={loading}>Refresh</button>
        <button type="button" onClick={onLogout}>Sign out</button>
      </div>
    </header>
  );
}
