import { useEffect, useMemo, useState } from "react";
import { api } from "./api.js";
import Dashboard from "./components/Dashboard.jsx";
import Login from "./components/Login.jsx";
import Members from "./components/Members.jsx";
import Programs from "./components/Programs.jsx";
import RuleStudio from "./components/RuleStudio.jsx";
import Sidebar from "./components/Sidebar.jsx";
import TopBar from "./components/TopBar.jsx";
import Transactions from "./components/Transactions.jsx";
import { defaultPartner } from "./data/mockData.js";
import { createRulesTemplate, rulesToPayload } from "./utils/rules.js";

const saved = JSON.parse(localStorage.getItem("paisa.partnerPortal") || "null");

export default function App() {
  const [isLoggedIn, setIsLoggedIn] = useState(Boolean(saved?.isLoggedIn));
  const [activeView, setActiveView] = useState("dashboard");
  const [partnerKey, setPartnerKey] = useState(saved?.partnerKey || defaultPartner.partnerKey);
  const [partner, setPartner] = useState(defaultPartner);
  const [programs, setPrograms] = useState([]);
  const [enrollments, setEnrollments] = useState([]);
  const [transactions, setTransactions] = useState([]);
  const [selectedProgramId, setSelectedProgramId] = useState(saved?.selectedProgramId || "");
  const [loading, setLoading] = useState(false);
  const [notice, setNotice] = useState("");
  const [error, setError] = useState("");

  const selectedProgram = useMemo(
    () => programs.find((program) => program.id === selectedProgramId) || programs[0],
    [programs, selectedProgramId],
  );

  useEffect(() => {
    if (isLoggedIn) {
      loadWorkspace(partnerKey);
    }
  }, [isLoggedIn]);

  function persistSession(next = {}) {
    localStorage.setItem("paisa.partnerPortal", JSON.stringify({
      isLoggedIn,
      selectedProgramId,
      partnerKey,
      ...next,
    }));
  }

  async function runAction(message, action) {
    setLoading(true);
    setError("");
    setNotice("");
    try {
      const result = await action();
      setNotice(message);
      return result;
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  }

  async function login(nextPartnerKey) {
    const key = nextPartnerKey || partnerKey;
    await runAction("Partner workspace loaded from API.", async () => {
      let apiPartner;
      try {
        apiPartner = await api.getPartner(key);
      } catch (err) {
        apiPartner = await api.createPartner({ partnerKey: key, name: titleizePartnerKey(key) });
      }
      setPartnerKey(key);
      setPartner(toUiPartner(apiPartner));
      setIsLoggedIn(true);
      persistSession({ isLoggedIn: true, partnerKey: key });
      await loadWorkspace(key, apiPartner);
    });
  }

  function logout() {
    setIsLoggedIn(false);
    persistSession({ isLoggedIn: false });
  }

  async function loadWorkspace(key = partnerKey, existingPartner = null) {
    setLoading(true);
    setError("");
    try {
      const apiPartner = existingPartner || await api.getPartner(key);
      const apiPrograms = await api.listPrograms(key);
      const apiMembers = await api.listMembers(key);
      const programModels = await Promise.all(apiPrograms.map((program) => hydrateProgram(key, program)));
      const memberModels = await Promise.all(apiMembers.map((member) => hydrateMember(key, member)));
      const transactionModels = await hydrateTransactions(key, memberModels);

      setPartner(toUiPartner(apiPartner));
      setPrograms(withProgramDerivedFields(programModels, memberModels, transactionModels));
      setEnrollments(memberModels);
      setTransactions(transactionModels);
      const nextSelected = selectedProgramId && programModels.some((program) => program.id === selectedProgramId)
        ? selectedProgramId
        : programModels[0]?.id || "";
      setSelectedProgramId(nextSelected);
      persistSession({ partnerKey: key, selectedProgramId: nextSelected, isLoggedIn: true });
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  async function createProgram() {
    await runAction("Program created in the API.", async () => {
      const index = programs.length + 1;
      const program = await api.createProgram(partnerKey, {
        name: `Rewards Program ${index}`,
        tierCode: `tier-${index}`,
        priority: index,
      });
      setSelectedProgramId(program.id);
      await loadWorkspace(partnerKey);
      setActiveView("programs");
    });
  }

  function updateProgram(programId, patch) {
    setPrograms(programs.map((program) => program.id === programId ? { ...program, ...patch } : program));
  }

  async function publishProgramRules(programId, draftProgram) {
    await runAction("Rule version published through the backend validator.", async () => {
      const body = rulesToPayload(draftProgram);
      const version = await api.createRuleVersion(partnerKey, programId, body);
      await api.publishRuleVersion(partnerKey, programId, version.id);
      await loadWorkspace(partnerKey);
    });
  }

  async function createRulePackage(programId) {
    await runAction("Rule package draft created in the API.", async () => {
      const body = {
        ...rulesToPayload({ rules: createRulesTemplate("stack") }),
        ruleSetKey: `member_add_on_${Date.now()}`,
        name: "New member add-on package",
        description: "Supplemental rule package assignable to selected members.",
      };
      await api.createRulePackage(partnerKey, programId, body);
      await loadWorkspace(partnerKey);
      setActiveView("rules");
    });
  }

  function updateRulePackage(programId, packageId, patch) {
    setPrograms(programs.map((program) => (
      program.id === programId
        ? { ...program, rulePackages: program.rulePackages.map((pkg) => pkg.id === packageId ? { ...pkg, ...patch } : pkg) }
        : program
    )));
  }

  async function publishRulePackage(programId, rulePackage, draftProgram) {
    await runAction("Member add-on package published through the backend.", async () => {
      const body = {
        ...rulesToPayload(draftProgram),
        ruleSetKey: rulePackage.key,
        name: rulePackage.name,
        description: rulePackage.description,
      };
      const version = await api.createRulePackage(partnerKey, programId, body);
      await api.publishRuleVersion(partnerKey, programId, version.id);
      await loadWorkspace(partnerKey);
    });
  }

  async function moveEnrollment(memberId, programId, reason) {
    await runAction("Member program enrollment updated in the API.", async () => {
      await api.updateEnrollment(partnerKey, memberId, {
        programId,
        changeReason: reason || "Program move",
      });
      await loadWorkspace(partnerKey);
    });
  }

  async function assignRulePackage(memberId, packageId) {
    await runAction("Rule package assigned to member in the API.", async () => {
      await api.createRuleAssignment(partnerKey, memberId, {
        ruleVersionId: packageId,
        reason: "Partner admin assignment",
      });
      await loadWorkspace(partnerKey);
    });
  }

  async function removeRulePackage(memberId, packageId) {
    const enrollment = enrollments.find((item) => item.id === memberId);
    const assignmentId = enrollment?.addOnAssignments?.[packageId];
    if (!assignmentId) return;
    await runAction("Rule package assignment ended in the API.", async () => {
      await api.updateRuleAssignment(partnerKey, memberId, assignmentId, {
        status: "ended",
        reason: "Partner admin removal",
      });
      await loadWorkspace(partnerKey);
    });
  }

  async function seedDemoSuite() {
    await runAction("API demo partner suite created.", async () => {
      const suffix = Date.now();
      const gold = await api.createProgram(partnerKey, { name: "Gold Rewards", tierCode: `gold-${suffix}`, priority: 1 });
      const everyday = await api.createProgram(partnerKey, { name: "Everyday Rewards", tierCode: `everyday-${suffix}`, priority: 2 });

      const goldVersion = await api.createRuleVersion(partnerKey, gold.id, rulesToPayload({ rules: createRulesTemplate("max_of") }));
      await api.publishRuleVersion(partnerKey, gold.id, goldVersion.id);
      const everydayVersion = await api.createRuleVersion(partnerKey, everyday.id, rulesToPayload({ rules: createRulesTemplate("stack") }));
      await api.publishRuleVersion(partnerKey, everyday.id, everydayVersion.id);
      const addOn = await api.createRulePackage(partnerKey, gold.id, {
        ...rulesToPayload({ rules: createRulesTemplate("stack") }),
        ruleSetKey: `vip_grocery_${suffix}`,
        name: "VIP grocery accelerator",
        description: "Adds extra earn for selected grocery-heavy members.",
      });
      await api.publishRuleVersion(partnerKey, gold.id, addOn.id);

      const member = await api.createMember(partnerKey, {
        externalCustomerId: `member-${suffix}`,
        programId: gold.id,
        identifiers: [{ type: "email", value: `member-${suffix}@example.invalid` }],
      });
      await api.createRuleAssignment(partnerKey, member.member.id, {
        ruleVersionId: addOn.id,
        reason: "Demo VIP assignment",
      });
      const purchase = await api.ingestTransaction(partnerKey, {
        externalTransactionId: `txn-${suffix}`,
        externalCustomerId: `member-${suffix}`,
        type: "purchase",
        currency: "USD",
        subtotalMinor: 10000,
        taxMinor: 600,
        totalMinor: 10600,
        eligibleMinor: 10000,
        occurredAt: new Date().toISOString(),
        lineItems: [{
          externalLineId: "line-1",
          sku: "sku-grocery",
          category: "grocery",
          quantity: 1,
          subtotalMinor: 10000,
          taxMinor: 600,
          totalMinor: 10600,
          eligibleMinor: 10000,
        }],
      });
      await api.processTransactions();
      setSelectedProgramId(gold.id);
      await loadWorkspace(partnerKey);
      setActiveView("dashboard");
      return purchase;
    });
  }

  function selectProgram(programId) {
    setSelectedProgramId(programId);
    persistSession({ selectedProgramId: programId });
  }

  if (!isLoggedIn) {
    return <Login partner={partner} partnerKey={partnerKey} apiBaseUrl={api.baseUrl} error={error} loading={loading} onLogin={login} onPartnerKeyChange={setPartnerKey} />;
  }

  return (
    <div className="app-shell">
      <Sidebar activeView={activeView} onChangeView={setActiveView} />
      <div className="main-shell">
        <TopBar partner={partner} selectedProgram={selectedProgram} apiBaseUrl={api.baseUrl} loading={loading} onRefresh={() => loadWorkspace(partnerKey)} onLogout={logout} />
        <main className="content">
          {notice || error ? (
            <div className={error ? "notice-bar error" : "notice-bar"}>
              <span>{error || notice}</span>
            </div>
          ) : null}
          {programs.length === 0 && !loading ? (
            <section className="panel spacious empty-api-state">
              <p className="eyebrow">API workspace</p>
              <h2>No programs exist for {partner.partnerKey} yet.</h2>
              <p>Create a program manually or seed a demo suite to exercise partner, program, rules, member, transaction, calculation, balance, and ledger APIs.</p>
              <div className="button-row">
                <button type="button" onClick={createProgram}>Create program</button>
                <button className="primary" type="button" onClick={seedDemoSuite}>Seed API demo suite</button>
              </div>
            </section>
          ) : null}
          {activeView === "dashboard" ? (
            <Dashboard
              programs={programs}
              enrollments={enrollments}
              transactions={transactions}
              selectedProgramId={selectedProgramId}
              onSelectProgram={selectProgram}
              onCreateProgram={createProgram}
              onSeedDemoSuite={seedDemoSuite}
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
          {activeView === "rules" && selectedProgram ? (
            <RuleStudio
              program={selectedProgram}
              onUpdateProgram={updateProgram}
              onPublishProgramRules={publishProgramRules}
              onCreateRulePackage={createRulePackage}
              onUpdateRulePackage={updateRulePackage}
              onPublishRulePackage={publishRulePackage}
            />
          ) : null}
          {activeView === "members" ? (
            <Members
              enrollments={enrollments}
              programs={programs}
              transactions={transactions}
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

async function hydrateProgram(partnerKey, program) {
  const versions = await api.listRuleVersions(partnerKey, program.id);
  const packages = await api.listRulePackages(partnerKey, program.id);
  const baseVersions = versions.filter((version) => version.scope !== "member_add_on");
  const selectedBase = baseVersions.find((version) => version.status === "published") || baseVersions[0];
  const review = selectedBase ? await api.getRuleVersionReview(partnerKey, program.id, selectedBase.id) : null;
  const packageModels = await Promise.all(packages.map(async (pkg) => {
    const packageReview = await api.getRuleVersionReview(partnerKey, program.id, pkg.id);
    return ruleVersionToPackage(pkg, packageReview);
  }));
  return {
    id: program.id,
    name: program.name,
    tierCode: program.tierCode || "base",
    status: selectedBase?.status || program.status,
    members: 0,
    liabilityPoints: 0,
    validationScore: review?.validation?.valid ? 100 : 0,
    ruleVersionId: selectedBase?.id || "",
    rules: review ? reviewToRules(review) : createRulesTemplate("max_of"),
    rulePackages: packageModels,
  };
}

async function hydrateMember(partnerKey, member) {
  try {
    const profile = await api.getRewardsProfile(partnerKey, member.id);
    const addOnAssignments = {};
    profile.addOns.forEach((assignment) => {
      addOnAssignments[assignment.ruleVersionId] = assignment.id;
    });
    return {
      id: member.id,
      member: member.externalCustomerId,
      email: member.externalCustomerId,
      programId: profile.enrollment.programId,
      status: member.status,
      points: profile.balance.availablePoints || 0,
      earnedPoints: profile.transactions.reduce((sum, event) => sum + Math.max(0, event.eligibleMinor || 0), 0),
      joinedAt: new Date(member.createdAt).toISOString().slice(0, 10),
      addOns: profile.addOns.map((assignment) => assignment.ruleVersionId),
      addOnAssignments,
      lastChangeReason: profile.enrollment.changeReason || "Active enrollment",
    };
  } catch (err) {
    return {
      id: member.id,
      member: member.externalCustomerId,
      email: member.externalCustomerId,
      programId: "",
      status: member.status,
      points: 0,
      earnedPoints: 0,
      joinedAt: new Date(member.createdAt).toISOString().slice(0, 10),
      addOns: [],
      addOnAssignments: {},
      lastChangeReason: "No active enrollment",
    };
  }
}

async function hydrateTransactions(partnerKey, members) {
  const memberNames = Object.fromEntries(members.map((member) => [member.id, member.member]));
  const events = await api.listTransactions(partnerKey);
  return Promise.all(events.map(async (event) => {
    let calculation = null;
    try {
      calculation = await api.getCalculation(partnerKey, event.id);
    } catch (err) {
      calculation = null;
    }
    const firstLine = event.lineItems?.[0];
    const points = calculation?.pointsDelta || 0;
    return {
      id: event.id,
      member: memberNames[event.memberId] || event.memberId,
      memberId: event.memberId,
      programId: calculation?.programId || "",
      type: event.type === "refund" ? "Refund" : "Earn",
      category: firstLine?.category || "transaction",
      amount: (event.totalMinor || 0) / 100,
      points,
      status: event.status === "processed" ? "posted" : event.status,
      occurredAt: new Date(event.occurredAt).toLocaleString(),
      ruleSource: calculation?.calculationData?.addOnRuleVersionIds?.length ? "Base program + member add-ons" : "Base program rules",
    };
  }));
}

function withProgramDerivedFields(programs, enrollments, transactions) {
  return programs.map((program) => ({
    ...program,
    members: enrollments.filter((enrollment) => enrollment.programId === program.id).length,
    liabilityPoints: enrollments
      .filter((enrollment) => enrollment.programId === program.id)
      .reduce((sum, enrollment) => sum + enrollment.points, 0),
    transactionCount: transactions.filter((transaction) => transaction.programId === program.id).length,
  }));
}

function toUiPartner(apiPartner) {
  return {
    ...defaultPartner,
    id: apiPartner.id,
    name: apiPartner.name,
    partnerKey: apiPartner.partnerKey,
    adminName: "API operator",
    adminEmail: `${apiPartner.partnerKey}@example.invalid`,
    apiEnvironment: "Local API",
  };
}

function reviewToRules(review) {
  return {
    earnBasis: review.ruleVersion.earnBasis,
    groups: review.groups.map((group) => ({
      id: group.id,
      name: group.name,
      strategy: group.resolutionStrategy,
      status: group.status,
      rules: group.rules.map(reviewRuleToUiRule),
    })),
  };
}

function reviewRuleToUiRule(rule) {
  const category = rule.eligibilityConfig?.categories?.[0] || (rule.eligibilityConfig?.firstPurchase ? "first purchase" : "All transactions");
  return {
    id: rule.id,
    key: rule.ruleKey,
    name: rule.name,
    type: rule.ruleType,
    pointsPerDollar: rule.formulaConfig?.pointsPerDollar || rule.formulaConfig?.points_per_dollar || 0,
    points: rule.formulaConfig?.points || rule.formulaConfig?.fixedPoints || 0,
    category,
    cap: rule.limits?.[0] ? `${rule.limits[0].maxPoints || rule.limits[0].maxBasisAmountMinor} / ${rule.limits[0].period}` : "",
    status: rule.status,
  };
}

function ruleVersionToPackage(version, review) {
  const rules = review.groups.flatMap((group) => group.rules.map(reviewRuleToUiRule));
  return {
    id: version.id,
    ruleVersionId: version.id,
    key: version.ruleSetKey || version.id,
    name: version.name || version.ruleSetKey || "Member add-on package",
    status: version.status,
    description: version.description || "Member-specific supplemental earning rules.",
    rules,
  };
}

function titleizePartnerKey(key) {
  return key.split("-").filter(Boolean).map((part) => part.charAt(0).toUpperCase() + part.slice(1)).join(" ") || "Demo Partner";
}
