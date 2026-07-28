import { useEffect, useMemo, useState } from "react";
import { api } from "./api.js";
import Dashboard from "./components/Dashboard.jsx";
import Login from "./components/Login.jsx";
import Members from "./components/Members.jsx";
import Onboarding from "./components/Onboarding.jsx";
import PartnerWelcome from "./components/PartnerWelcome.jsx";
import Programs from "./components/Programs.jsx";
import Settings from "./components/Settings.jsx";
import Sidebar from "./components/Sidebar.jsx";
import TopBar from "./components/TopBar.jsx";
import Transactions from "./components/Transactions.jsx";
import { defaultPartner } from "./data/mockData.js";
import { createRulesTemplate, limitFromCap, rulesToPayload } from "./utils/rules.js";

const saved = JSON.parse(localStorage.getItem("paisa.partnerPortal") || "null");

export default function App() {
  const [isLoggedIn, setIsLoggedIn] = useState(Boolean(saved?.isLoggedIn && localStorage.getItem("paisa.partnerPortal.token")));
  const [activeView, setActiveView] = useState("dashboard");
  const [partnerKey, setPartnerKey] = useState(saved?.partnerKey || defaultPartner.partnerKey);
  const [loginEmail, setLoginEmail] = useState(saved?.loginEmail || "");
  const [loginPassword, setLoginPassword] = useState("");
  const [partner, setPartner] = useState(defaultPartner);
  const [programs, setPrograms] = useState([]);
  const [enrollments, setEnrollments] = useState([]);
  const [transactions, setTransactions] = useState([]);
  const [catalogItems, setCatalogItems] = useState([]);
  const [redemptions, setRedemptions] = useState([]);
  const [apiKeys, setApiKeys] = useState([]);
  const [latestApiToken, setLatestApiToken] = useState("");
  const [dashboardSummary, setDashboardSummary] = useState(null);
  const [cashier, setCashier] = useState({ member: null, balance: {}, availableRewards: [], redemption: null });
  const [selectedProgramId, setSelectedProgramId] = useState(saved?.selectedProgramId || "");
  const [theme, setTheme] = useState(saved?.theme || "porcelain");
  const [loading, setLoading] = useState(false);
  const [notice, setNotice] = useState("");
  const [error, setError] = useState("");
  const [showWelcome, setShowWelcome] = useState(false);

  const selectedProgram = useMemo(
    () => programs.find((program) => program.id === selectedProgramId) || programs[0],
    [programs, selectedProgramId],
  );
  const setupStatus = useMemo(() => {
    const publishedPrograms = programs.filter((program) => program.status === "published").length;
    return {
      hasProgram: programs.length > 0,
      hasRules: publishedPrograms > 0,
      hasReward: (dashboardSummary?.activeCatalogItems || 0) > 0,
    };
  }, [dashboardSummary, programs]);
  const setupComplete = setupStatus.hasProgram
    && setupStatus.hasRules
    && setupStatus.hasReward;
  const setupLocked = isLoggedIn && !setupComplete;

  useEffect(() => {
    if (isLoggedIn) {
      loadWorkspace(partnerKey);
    }
  }, [isLoggedIn]);

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
    persistSession({ theme });
  }, [theme]);

  useEffect(() => {
    if (setupLocked && activeView !== "setup") {
      setActiveView("setup");
      return;
    }
    if (!setupLocked && activeView === "setup") {
      setActiveView("dashboard");
    }
  }, [activeView, setupLocked]);

  useEffect(() => {
    if (!setupLocked) {
      setShowWelcome(false);
      return;
    }
    setShowWelcome(!localStorage.getItem(welcomeStorageKey(partnerKey)));
  }, [partnerKey, setupLocked]);

  function persistSession(next = {}) {
    localStorage.setItem("paisa.partnerPortal", JSON.stringify({
      isLoggedIn,
      selectedProgramId,
      partnerKey,
      theme,
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

  async function login(credentials) {
    await runAction("Workspace loaded.", async () => {
      const loginResult = await api.login({
        email: credentials.email,
        password: credentials.password,
      });
      api.setAuthToken(loginResult.token);
      const apiPartner = loginResult.partner;
      setPartnerKey(apiPartner.partnerKey);
      setPartner(toUiPartner(apiPartner));
      setIsLoggedIn(true);
      setLoginPassword("");
      persistSession({ isLoggedIn: true, partnerKey: apiPartner.partnerKey, loginEmail: credentials.email, token: true, theme });
      await loadWorkspace(apiPartner.partnerKey, apiPartner);
    });
  }

  function logout() {
    api.logout().catch(() => {});
    api.setAuthToken("");
    setIsLoggedIn(false);
    setShowWelcome(false);
    persistSession({ isLoggedIn: false });
  }

  function beginSetup() {
    localStorage.setItem(welcomeStorageKey(partnerKey), "true");
    setShowWelcome(false);
  }

  async function loadWorkspace(key = partnerKey, existingPartner = null) {
    setLoading(true);
    setError("");
    try {
      const ops = await loadMerchantOps();
      const apiPartner = existingPartner || ops?.summary?.partner || partner;
      const apiPrograms = await api.listPrograms();
      const apiMembers = await api.listMembers();
      const programModels = await Promise.all(apiPrograms.map((program) => hydrateProgram(program)));
      const memberModels = await Promise.all(apiMembers.map((member) => hydrateMember(member)));
      const transactionModels = await hydrateTransactions(memberModels);

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

  async function loadMerchantOps() {
    const [summary, items, redemptionList, keys] = await Promise.all([
      api.dashboardSummary().catch(() => null),
      api.listCatalogItems().catch(() => []),
      api.listRedemptions().catch(() => []),
      api.listApiKeys().catch(() => []),
    ]);
    setDashboardSummary(summary);
    setCatalogItems(items);
    setRedemptions(redemptionList);
    setApiKeys(keys);
    return { summary, items, redemptionList, keys };
  }

  async function createProgram(input = {}) {
    await runAction("Program created.", async () => {
      const index = programs.length + 1;
      const program = await api.createProgram({
        name: input.name?.trim() || `Rewards Program ${index}`,
        tierCode: input.tierCode?.trim() || `tier-${index}`,
        priority: index,
      });
      await loadWorkspace(partnerKey);
      setSelectedProgramId(program.id);
      persistSession({ selectedProgramId: program.id });
      setActiveView("programs");
    });
  }

  function updateProgram(programId, patch) {
    setPrograms(programs.map((program) => program.id === programId ? { ...program, ...patch } : program));
  }

  async function saveProgramDetails(programId, patch) {
    const current = programs.find((program) => program.id === programId);
    if (!current) return;
    await runAction("Program details saved.", async () => {
      await api.updateProgram(programId, {
        name: patch.name?.trim() || current.name,
        tierCode: patch.tierCode?.trim() || current.tierCode,
        priority: current.priority || programs.findIndex((program) => program.id === programId) + 1,
      });
      await loadWorkspace(partnerKey);
      setSelectedProgramId(programId);
    });
  }

  async function deleteDraftProgram(programId) {
    const current = programs.find((program) => program.id === programId);
    if (!current) return;
    await runAction("Draft program deleted.", async () => {
      await api.deleteProgram(programId);
      const remaining = programs.filter((program) => program.id !== programId);
      const nextSelected = remaining[0]?.id || "";
      setSelectedProgramId(nextSelected);
      persistSession({ selectedProgramId: nextSelected });
      await loadWorkspace(partnerKey);
    });
  }

  async function publishProgramRules(programId, draftProgram) {
    await runAction("Rule version published.", async () => {
      const body = rulesToPayload(draftProgram);
      const version = await api.createRuleVersion(programId, body);
      await api.publishRuleVersion(programId, version.id);
      await loadWorkspace(partnerKey);
    });
  }

  async function createRulePackage(programId) {
    await runAction("Rule package draft created.", async () => {
      const body = {
        ...rulesToPayload({ rules: createRulesTemplate("stack") }),
        ruleSetKey: `member_add_on_${Date.now()}`,
        name: "New member add-on package",
        description: "Supplemental rule package assignable to selected members.",
      };
      await api.createRulePackage(programId, body);
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
    await runAction("Member add-on package published.", async () => {
      const body = {
        ...rulesToPayload(draftProgram),
        ruleSetKey: rulePackage.key,
        name: rulePackage.name,
        description: rulePackage.description,
      };
      const version = await api.createRulePackage(programId, body);
      await api.publishRuleVersion(programId, version.id);
      await loadWorkspace(partnerKey);
    });
  }

  async function moveEnrollment(memberId, programId, reason) {
    await runAction("Member program enrollment updated.", async () => {
      await api.updateEnrollment(memberId, {
        programId,
        changeReason: reason || "Program move",
      });
      await loadWorkspace(partnerKey);
    });
  }

  async function assignRulePackage(memberId, packageId) {
    await runAction("Rule package assigned to member.", async () => {
      await api.createRuleAssignment(memberId, {
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
    await runAction("Rule package assignment ended.", async () => {
      await api.updateRuleAssignment(memberId, assignmentId, {
        status: "ended",
        reason: "Partner admin removal",
      });
      await loadWorkspace(partnerKey);
    });
  }

  async function resolveCashierMember(body) {
    await runAction("Member resolved for cashier mode.", async () => {
      const result = await api.resolveMember(body);
      const [balance, rewards] = await Promise.all([
        api.getPOSBalance(result.member.id).catch(() => ({})),
        api.availableRewards(result.member.id).catch(() => []),
      ]);
      setCashier({ member: result.member, balance, availableRewards: rewards, redemption: null });
      await loadWorkspace(partnerKey);
    });
  }

  async function createCashierTransaction(body) {
    await runAction("Manual purchase submitted.", async () => {
      const event = await api.createManualTransaction(body);
      await api.processTransactions();
      const memberId = body.memberId || cashier.member?.id;
      const balance = memberId ? await api.getPOSBalance(memberId).catch(() => cashier.balance) : cashier.balance;
      setCashier((current) => ({ ...current, balance }));
      await loadWorkspace(partnerKey);
      return event;
    });
  }

  async function createCatalogItem(body) {
    await runAction("Reward catalog item created.", async () => {
      await api.createCatalogItem(body);
      await loadMerchantOps();
    });
  }

  async function createRedemption(catalogItemId) {
    if (!cashier.member) return;
    await runAction("Reward reserved.", async () => {
      const result = await api.createRedemption({ memberId: cashier.member.id, catalogItemId });
      setCashier((current) => ({ ...current, redemption: result.redemption, balance: result.balance }));
      await loadMerchantOps();
    });
  }

  async function validateRedemption(redemptionId) {
    await runAction("Reward validated.", async () => {
      const result = await api.validateRedemption(redemptionId);
      setCashier((current) => ({ ...current, redemption: result.redemption, balance: result.balance }));
      await loadMerchantOps();
    });
  }

  async function captureRedemption(redemptionId) {
    await runAction("Reward captured.", async () => {
      const result = await api.captureRedemption(redemptionId);
      setCashier((current) => ({ ...current, redemption: result.redemption, balance: result.balance }));
      await loadMerchantOps();
      await loadWorkspace(partnerKey);
    });
  }

  async function releaseRedemption(redemptionId) {
    await runAction("Reward released.", async () => {
      const result = await api.releaseRedemption(redemptionId);
      setCashier((current) => ({ ...current, redemption: result.redemption, balance: result.balance }));
      await loadMerchantOps();
    });
  }

  async function createApiKey(name) {
    await runAction("Access key created.", async () => {
      const result = await api.createApiKey({ name });
      setLatestApiToken(result.token);
      await loadMerchantOps();
    });
  }

  function selectProgram(programId) {
    setSelectedProgramId(programId);
    persistSession({ selectedProgramId: programId });
  }

  if (!isLoggedIn) {
    return (
      <Login
        email={loginEmail}
        password={loginPassword}
        error={error}
        loading={loading}
        onLogin={login}
        onEmailChange={setLoginEmail}
        onPasswordChange={setLoginPassword}
      />
    );
  }

  if (setupLocked) {
    return (
      <div className="setup-only-shell">
        <main className="content setup-only-content">
          {notice || error ? (
            <div className={error ? "notice-bar error" : "notice-bar"}>
              <span>{error || notice}</span>
            </div>
          ) : null}
          {showWelcome ? (
            <PartnerWelcome partner={partner} onBegin={beginSetup} onLogout={logout} />
          ) : (
            <Onboarding
              partner={partner}
              programs={programs}
              transactions={transactions}
              dashboardSummary={dashboardSummary}
              catalogItems={catalogItems}
              cashier={cashier}
              setupLocked
              selectedProgram={selectedProgram}
              redemptions={redemptions}
              onCreateProgram={createProgram}
              onUpdateProgram={updateProgram}
              onCreateRulePackage={createRulePackage}
              onPublishProgramRules={publishProgramRules}
              onUpdateRulePackage={updateRulePackage}
              onPublishRulePackage={publishRulePackage}
              onCreateCatalogItem={createCatalogItem}
              onChangeView={setActiveView}
              onResolveMember={resolveCashierMember}
              onCreateTransaction={createCashierTransaction}
              onCreateRedemption={createRedemption}
              onValidateRedemption={validateRedemption}
              onCaptureRedemption={captureRedemption}
              onReleaseRedemption={releaseRedemption}
              onLogout={logout}
            />
          )}
        </main>
      </div>
    );
  }

  return (
    <div className="app-shell">
      <Sidebar activeView={activeView} onChangeView={setActiveView} />
      <div className="main-shell">
        <TopBar partner={partner} selectedProgram={selectedProgram} loading={loading} onRefresh={() => loadWorkspace(partnerKey)} onLogout={logout} />
        <main className="content">
          {notice || error ? (
            <div className={error ? "notice-bar error" : "notice-bar"}>
              <span>{error || notice}</span>
            </div>
          ) : null}
          {activeView === "dashboard" ? (
            <Dashboard
              programs={programs}
              enrollments={enrollments}
              transactions={transactions}
              dashboardSummary={dashboardSummary}
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
              onSaveProgramDetails={saveProgramDetails}
              onDeleteDraftProgram={deleteDraftProgram}
              enrollments={enrollments}
              transactions={transactions}
              catalogItems={catalogItems}
              redemptions={redemptions}
              onCreateRulePackage={createRulePackage}
              onPublishProgramRules={publishProgramRules}
              onUpdateRulePackage={updateRulePackage}
              onPublishRulePackage={publishRulePackage}
              onCreateCatalogItem={createCatalogItem}
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
          {activeView === "activity" ? (
            <Transactions transactions={transactions} programs={programs} redemptions={redemptions} />
          ) : null}
          {activeView === "settings" ? (
            <Settings
              apiKeys={apiKeys}
              latestApiToken={latestApiToken}
              onCreateApiKey={createApiKey}
            />
          ) : null}
        </main>
      </div>
    </div>
  );
}

async function hydrateProgram(program) {
  const versions = await api.listRuleVersions(program.id);
  const packages = await api.listRulePackages(program.id);
  const baseVersions = versions.filter((version) => version.scope !== "member_add_on");
  const publishedBase = baseVersions.find((version) => version.status === "published");
  const selectedBase = publishedBase || baseVersions[0];
  const review = selectedBase ? await api.getRuleVersionReview(program.id, selectedBase.id) : null;
  const packageModels = await Promise.all(packages.map(async (pkg) => {
    const packageReview = await api.getRuleVersionReview(program.id, pkg.id);
    return ruleVersionToPackage(pkg, packageReview);
  }));
  return {
    id: program.id,
    name: program.name,
    tierCode: program.tierCode || "base",
    status: publishedBase ? "published" : "draft",
    members: 0,
    liabilityPoints: 0,
    validationScore: review?.validation?.valid ? 100 : 0,
    ruleVersionId: selectedBase?.id || "",
    rules: review ? reviewToRules(review) : createRulesTemplate("base"),
    rulePackages: packageModels,
  };
}

async function hydrateMember(member) {
  try {
    const profile = await api.getRewardsProfile(member.id);
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

async function hydrateTransactions(members) {
  const memberNames = Object.fromEntries(members.map((member) => [member.id, member.member]));
  const events = await api.listTransactions();
  return Promise.all(events.map(async (event) => {
    let calculation = null;
    try {
      calculation = await api.getCalculation(event.id);
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
    adminName: "Partner admin",
    adminEmail: "",
    apiEnvironment: "",
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
  const cap = rule.limits?.[0] ? ruleLimitToCap(rule.limits[0]) : "";
  return {
    id: rule.id,
    key: rule.ruleKey,
    name: rule.name,
    type: rule.ruleType,
    pointsPerDollar: rule.formulaConfig?.pointsPerDollar || rule.formulaConfig?.points_per_dollar || 0,
    points: rule.formulaConfig?.points || rule.formulaConfig?.fixedPoints || 0,
    category,
    basis: rule.eligibilityConfig?.basis || "",
    cap,
    limit: limitFromCap(cap),
    interaction: { mode: rule.dependencies?.some((dep) => dep.dependencyType === "requires_exhausted") ? "overflow_after_cap" : "adds", dependsOnRuleKey: rule.dependencies?.[0]?.dependsOnRuleKey || "" },
    dependencies: rule.dependencies || [],
    status: rule.status,
  };
}

function ruleLimitToCap(limit) {
  const period = limit.period === "calendar_month" ? "month" : limit.period;
  if (limit.maxBasisAmountMinor) return `${limit.maxBasisAmountMinor} basis / ${period}`;
  if (limit.maxPoints) return `${limit.maxPoints} pts / ${period}`;
  return "";
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
  return key.split("-").filter(Boolean).map((part) => part.charAt(0).toUpperCase() + part.slice(1)).join(" ") || "Partner";
}

function welcomeStorageKey(key) {
  return `paisa.partnerPortal.welcome.${key || "partner"}`;
}
