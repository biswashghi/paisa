export default function TopBar({ partner, selectedProgram, loading, onRefresh, onLogout }) {
  return (
    <header className="topbar">
      <div className="topbar-title">
        <p className="eyebrow">Partner</p>
        <h1>{partner.name}</h1>
      </div>
      <div className="context-strip">
        <div>
          <span>Program</span>
          <strong>{selectedProgram?.name || "No program selected"}</strong>
        </div>
        <div>
          <span>Status</span>
          <strong className="good-text">{loading ? "Syncing" : "Connected"}</strong>
        </div>
        <button type="button" onClick={onRefresh} disabled={loading}>Refresh</button>
        <button type="button" onClick={onLogout}>Sign out</button>
      </div>
    </header>
  );
}
