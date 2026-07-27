import PaisaLogo from "./PaisaLogo.jsx";

const items = [
  ["dashboard", "01", "Dashboard", "Status"],
  ["setup", "02", "Setup", "Checklist"],
  ["programs", "03", "Programs", "Rules + rewards"],
  ["members", "04", "Members", "Enrollments"],
  ["activity", "05", "Activity", "Transactions"],
  ["settings", "06", "Settings", "Keys + integrations"],
];

export default function Sidebar({ activeView, onChangeView }) {
  return (
    <aside className="sidebar">
      <div className="sidebar-brand">
        <PaisaLogo />
        <div>
          <strong>Paisa</strong>
          <span>Partner Admin</span>
        </div>
      </div>
      <nav>
        {items.map(([id, index, label, detail]) => (
          <button className={activeView === id ? "active" : ""} key={id} type="button" onClick={() => onChangeView(id)}>
            <span className="nav-index" aria-hidden="true">{index}</span>
            <span>
              <strong>{label}</strong>
              <small>{detail}</small>
            </span>
          </button>
        ))}
      </nav>
      <div className="sidebar-footer">
        <span>Workspace</span>
        <strong>Local</strong>
        <small>Partner session</small>
      </div>
    </aside>
  );
}
