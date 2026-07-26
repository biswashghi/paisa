const themeOptions = [
  ["graphite", "Graphite"],
  ["porcelain", "Porcelain"],
  ["aubergine", "Aubergine"],
];

export default function TopBar({ partner, selectedProgram, apiBaseUrl, loading, theme, onThemeChange, onRefresh, onLogout }) {
  return (
    <header className="topbar">
      <div className="topbar-title">
        <p className="eyebrow">Partner</p>
        <h1>{partner.name}</h1>
      </div>
      <div className="context-strip">
        <label className="command-search">
          <span>Search</span>
          <input readOnly value="Members, programs, rules" />
        </label>
        <div>
          <span>Program</span>
          <strong>{selectedProgram?.name || "No program selected"}</strong>
        </div>
        <div>
          <span>API</span>
          <strong>{partner.apiEnvironment}</strong>
        </div>
        <div>
          <span>Status</span>
          <strong className="good-text">{loading ? "Syncing" : "Connected"}</strong>
        </div>
        <div>
          <span>Backend</span>
          <strong>{apiBaseUrl.replace(/^https?:\/\//, "")}</strong>
        </div>
        <label className="theme-select">
          <span>Theme</span>
          <select value={theme} onChange={(event) => onThemeChange(event.target.value)}>
            {themeOptions.map(([value, label]) => (
              <option key={value} value={value}>{label}</option>
            ))}
          </select>
        </label>
        <button type="button" onClick={onRefresh} disabled={loading}>Refresh</button>
        <button type="button" onClick={onLogout}>Sign out</button>
      </div>
    </header>
  );
}
