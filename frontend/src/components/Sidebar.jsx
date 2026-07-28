import PaisaLogo from "./PaisaLogo.jsx";

const items = [
  ["dashboard", "01", "Dashboard", "Status"],
  ["programs", "02", "Programs", "Rules + rewards"],
  ["members", "03", "Members", "Enrollments"],
  ["activity", "04", "Activity", "Transactions"],
  ["settings", "05", "Settings", "Access"],
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
        <strong>Partner portal</strong>
        <small>Signed in</small>
      </div>
    </aside>
  );
}
