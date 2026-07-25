import { useMemo, useState } from "react";
import Dashboard from "./components/Dashboard.jsx";
import Login from "./components/Login.jsx";
import Members from "./components/Members.jsx";
import Programs from "./components/Programs.jsx";
import RuleStudio from "./components/RuleStudio.jsx";
import Sidebar from "./components/Sidebar.jsx";
import TopBar from "./components/TopBar.jsx";
import Transactions from "./components/Transactions.jsx";
import { defaultPartner, initialEnrollments, initialPrograms, initialTransactions } from "./data/mockData.js";
import { createProgramDraft } from "./utils/rules.js";

const saved = JSON.parse(localStorage.getItem("paisa.partnerPortal") || "null");

export default function App() {
  const [isLoggedIn, setIsLoggedIn] = useState(Boolean(saved?.isLoggedIn));
  const [activeView, setActiveView] = useState("dashboard");
  const [programs, setPrograms] = useStoredState("programs", initialPrograms);
  const [enrollments, setEnrollments] = useStoredState("enrollments", initialEnrollments);
  const [transactions] = useStoredState("transactions", initialTransactions);
  const [selectedProgramId, setSelectedProgramId] = useState(saved?.selectedProgramId || initialPrograms[0].id);

  const selectedProgram = useMemo(
    () => programs.find((program) => program.id === selectedProgramId) || programs[0],
    [programs, selectedProgramId],
  );

  function persistSession(next = {}) {
    localStorage.setItem("paisa.partnerPortal", JSON.stringify({
      isLoggedIn,
      selectedProgramId,
      ...next,
    }));
  }

  function login() {
    setIsLoggedIn(true);
    persistSession({ isLoggedIn: true });
  }

  function logout() {
    setIsLoggedIn(false);
    persistSession({ isLoggedIn: false });
  }

  function createProgram() {
    const program = createProgramDraft(programs.length + 1);
    setPrograms([...programs, program]);
    setSelectedProgramId(program.id);
    setActiveView("programs");
    persistSession({ selectedProgramId: program.id });
  }

  function updateProgram(programId, patch) {
    setPrograms(programs.map((program) => program.id === programId ? { ...program, ...patch } : program));
  }

  function updateEnrollment(enrollmentId, patch) {
    setEnrollments(enrollments.map((enrollment) => enrollment.id === enrollmentId ? { ...enrollment, ...patch } : enrollment));
  }

  function moveEnrollment(enrollmentId, programId, reason) {
    updateEnrollment(enrollmentId, {
      programId,
      status: "active",
      addOns: [],
      lastChangeReason: reason || "Program move",
      movedAt: new Date().toISOString().slice(0, 10),
    });
  }

  function assignRulePackage(enrollmentId, packageId) {
    setEnrollments(enrollments.map((enrollment) => {
      if (enrollment.id !== enrollmentId || enrollment.addOns.includes(packageId)) return enrollment;
      return { ...enrollment, addOns: [...enrollment.addOns, packageId], lastChangeReason: "Rule package assigned" };
    }));
  }

  function removeRulePackage(enrollmentId, packageId) {
    setEnrollments(enrollments.map((enrollment) => (
      enrollment.id === enrollmentId
        ? { ...enrollment, addOns: enrollment.addOns.filter((id) => id !== packageId), lastChangeReason: "Rule package removed" }
        : enrollment
    )));
  }

  function createRulePackage(programId) {
    const rulePackage = {
      id: `pkg-${Date.now()}`,
      ruleVersionId: `rules-${Date.now()}`,
      key: `member_add_on_${Date.now()}`,
      name: "New member add-on package",
      status: "draft",
      description: "Supplemental rule package assignable to selected members.",
      rules: [
        { id: `pkg-rule-${Date.now()}`, key: "bonus_earn", name: "Bonus earn", type: "points_per_dollar", pointsPerDollar: 1, category: "All transactions", cap: "", status: "active" },
      ],
    };
    setPrograms(programs.map((program) => (
      program.id === programId ? { ...program, rulePackages: [...(program.rulePackages || []), rulePackage] } : program
    )));
  }

  function updateRulePackage(programId, packageId, patch) {
    setPrograms(programs.map((program) => (
      program.id === programId
        ? { ...program, rulePackages: program.rulePackages.map((pkg) => pkg.id === packageId ? { ...pkg, ...patch } : pkg) }
        : program
    )));
  }

  function selectProgram(programId) {
    setSelectedProgramId(programId);
    persistSession({ selectedProgramId: programId });
  }

  if (!isLoggedIn) {
    return <Login partner={defaultPartner} onLogin={login} />;
  }

  return (
    <div className="app-shell">
      <Sidebar activeView={activeView} onChangeView={setActiveView} />
      <div className="main-shell">
        <TopBar partner={defaultPartner} selectedProgram={selectedProgram} onLogout={logout} />
        <main className="content">
          {activeView === "dashboard" ? (
            <Dashboard
              programs={programs}
              enrollments={enrollments}
              transactions={transactions}
              selectedProgramId={selectedProgramId}
              onSelectProgram={selectProgram}
              onCreateProgram={createProgram}
              onChangeView={setActiveView}
            />
          ) : null}
          {activeView === "programs" ? (
            <Programs
              programs={programs}
              selectedProgramId={selectedProgramId}
              onSelectProgram={selectProgram}
              onCreateProgram={createProgram}
              onUpdateProgram={updateProgram}
              enrollments={enrollments}
              transactions={transactions}
              onCreateRulePackage={createRulePackage}
            />
          ) : null}
          {activeView === "rules" ? (
            <RuleStudio program={selectedProgram} onUpdateProgram={updateProgram} onCreateRulePackage={createRulePackage} onUpdateRulePackage={updateRulePackage} />
          ) : null}
          {activeView === "members" ? (
            <Members
              enrollments={enrollments}
              programs={programs}
              transactions={transactions}
              onUpdateEnrollment={updateEnrollment}
              onMoveEnrollment={moveEnrollment}
              onAssignRulePackage={assignRulePackage}
              onRemoveRulePackage={removeRulePackage}
            />
          ) : null}
          {activeView === "transactions" ? (
            <Transactions transactions={transactions} programs={programs} />
          ) : null}
        </main>
      </div>
    </div>
  );
}

function useStoredState(key, initialValue) {
  const storageKey = `paisa.partnerPortal.${key}`;
  const [value, setValue] = useState(() => JSON.parse(localStorage.getItem(storageKey) || "null") || initialValue);

  function setStoredValue(nextValue) {
    localStorage.setItem(storageKey, JSON.stringify(nextValue));
    setValue(nextValue);
  }

  return [value, setStoredValue];
}
