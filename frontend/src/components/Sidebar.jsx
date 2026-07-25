import PaisaLogo from "./PaisaLogo.jsx";

const items = [
  ["dashboard", "Dashboard", "Portfolio health"],
  ["programs", "Programs", "Base + packages"],
  ["rules", "Rule Studio", "Graph editor"],
  ["members", "Enrollments", "Moves + add-ons"],
  ["transactions", "Earned Activity", "Point events"],
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
      <div className="sidebar-snapshot">
        <span>Today</span>
        <strong>Rewards live</strong>
        <small>Base rules and member packages are evaluating in production.</small>
      </div>
      <nav>
        {items.map(([id, label, detail]) => (
          <button className={activeView === id ? "active" : ""} key={id} type="button" onClick={() => onChangeView(id)}>
            <span className="nav-dot" aria-hidden="true" />
            <span>
              <strong>{label}</strong>
              <small>{detail}</small>
            </span>
          </button>
        ))}
      </nav>
      <div className="sidebar-footer">
        <span>Workspace</span>
        <strong>Production</strong>
        <small>Default partner session</small>
      </div>
    </aside>
  );
}
