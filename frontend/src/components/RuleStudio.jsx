import {
  capFromLimit,
  createRule,
  createRulesTemplate,
  limitFromCap,
  limitsFromRule,
  validateProgramRules,
} from "../utils/rules.js";
import StatusPill from "./StatusPill.jsx";
import { useState } from "react";

const CATEGORY_OPTIONS = [
  { value: "All transactions", label: "All transactions" },
  { value: "grocery", label: "Grocery" },
  { value: "apparel", label: "Apparel" },
  { value: "fuel", label: "Fuel" },
  { value: "dining", label: "Dining" },
  { value: "travel", label: "Travel" },
  { value: "custom", label: "Custom category" },
];

const PATTERNS = [
  {
    kind: "base",
    title: "Every purchase earns",
    tag: "Start here",
    body: "A single base rate applies to the amount you choose.",
  },
  {
    kind: "stack",
    title: "Bonus adds on",
    tag: "Simple bonus",
    body: "Base earn still applies, and a matching bonus adds more points.",
  },
  {
    kind: "max_of",
    title: "Better rate wins",
    tag: "Compare offers",
    body: "Multiple candidates can match, but only the best award posts.",
  },
  {
    kind: "waterfall",
    title: "High rate, then overflow",
    tag: "Controlled spend",
    body: "Reward the first slice at a higher rate, then send the remainder to a lower rate.",
  },
];

export default function RuleStudio({ program, onUpdateProgram, onPublishProgramRules, onCreateRulePackage, onUpdateRulePackage, onPublishRulePackage, embedded = false, showPublish = true }) {
  const [appliesTo, setAppliesTo] = useState("program_base");
  const [selectedPackageId, setSelectedPackageId] = useState(program.rulePackages?.[0]?.id || "");
  const [selectedRuleId, setSelectedRuleId] = useState("");
  const [example, setExample] = useState({ total: "127.70", subtotal: "120.00", eligible: "90.00", grocery: "80.00" });
  const selectedPackage = (program.rulePackages || []).find((pkg) => pkg.id === selectedPackageId) || program.rulePackages?.[0];
  const editingPackage = appliesTo === "member_add_on" && selectedPackage;
  const editableRules = editingPackage ? packageToRules(selectedPackage) : program.rules;
  const editableProgram = editingPackage ? { ...program, rules: editableRules } : program;
  const issues = validateProgramRules(editableProgram);
  const group = editableRules.groups[0] || { name: "Every purchase earns", strategy: "stack", rules: [] };
  const selectedRule = group.rules.find((rule) => rule.id === selectedRuleId) || group.rules[0];
  const preview = ruleCalculationPreview(editableRules, example);

  function updateRules(nextRules) {
    if (editingPackage) {
      onUpdateRulePackage(program.id, selectedPackage.id, { ...rulesToPackagePatch(nextRules), status: "draft" });
      return;
    }
    onUpdateProgram(program.id, { rules: nextRules, status: "draft" });
  }

  function applyPattern(kind) {
    const nextRules = createRulesTemplate(kind);
    updateRules(nextRules);
    setSelectedRuleId(nextRules.groups[0]?.rules[0]?.id || "");
  }

  function updateGroup(patch) {
    updateRules({ ...editableRules, groups: [{ ...group, ...patch }] });
  }

  function updateRule(ruleId, patch) {
    updateGroup({ rules: group.rules.map((rule) => rule.id === ruleId ? { ...rule, ...patch } : rule) });
  }

  function updateSelectedRule(patch) {
    if (!selectedRule) return;
    updateRule(selectedRule.id, patch);
  }

  function addRule() {
    const nextRule = createRule("New earn rule", `rule_${group.rules.length + 1}`, 1, "All transactions", "");
    updateGroup({ rules: [...group.rules, nextRule] });
    setSelectedRuleId(nextRule.id);
  }

  function addOverflowRule() {
    if (!selectedRule) return;
    const nextRule = {
      ...createRule(`${selectedRule.name} overflow`, `${selectedRule.key}_overflow`, 1, selectedRule.category || "All transactions", "after cap"),
      basis: selectedRule.basis || editableRules.earnBasis,
      interaction: { mode: "overflow_after_cap", dependsOnRuleKey: selectedRule.key },
      dependencies: [{ dependsOnRuleKey: selectedRule.key, dependencyType: "requires_exhausted" }],
    };
    updateGroup({ strategy: "waterfall", rules: [...group.rules, nextRule] });
    setSelectedRuleId(nextRule.id);
  }

  function deleteSelectedRule() {
    if (!selectedRule || group.rules.length === 1) return;
    const remainingRules = group.rules.filter((rule) => rule.id !== selectedRule.id);
    updateGroup({ rules: remainingRules });
    setSelectedRuleId(remainingRules[0]?.id || "");
  }

  function moveSelectedRule(direction) {
    if (!selectedRule) return;
    const index = group.rules.findIndex((rule) => rule.id === selectedRule.id);
    const nextIndex = index + direction;
    if (nextIndex < 0 || nextIndex >= group.rules.length) return;
    const nextRules = [...group.rules];
    [nextRules[index], nextRules[nextIndex]] = [nextRules[nextIndex], nextRules[index]];
    updateGroup({ rules: nextRules });
  }

  function publish() {
    if (issues.length) return;
    const nextProgram = {
      ...program,
      status: "published",
      validationScore: 99.1,
      rules: { ...program.rules, groups: program.rules.groups.map((item) => ({ ...item, status: "published" })) },
    };
    onUpdateProgram(program.id, nextProgram);
    onPublishProgramRules(program.id, nextProgram);
  }

  function publishPackage() {
    if (!editingPackage || issues.length) return;
    onUpdateRulePackage(program.id, selectedPackage.id, { status: "published" });
    onPublishRulePackage(program.id, selectedPackage, editableProgram);
  }

  function createPackage() {
    onCreateRulePackage(program.id);
    setAppliesTo("member_add_on");
  }

  return (
    <section className="view-stack">
      <div className={embedded ? "rule-setup-banner rule-setup-banner-compact" : "view-header"}>
        <div>
          <p className="eyebrow">{embedded ? "Rule draft" : "Rules"}</p>
          <h2>{embedded ? "Choose how customers earn" : editingPackage ? selectedPackage.name : program.name}</h2>
          <div className="studio-scope-line">
            <span>{appliesTo === "program_base" ? "Program base rules" : "Member add-on package"}</span>
            <span>{group.rules.length} rules</span>
            <span>{issues.length ? `${issues.length} issues` : "Ready for review"}</span>
          </div>
        </div>
        {showPublish ? (
          <button className="primary" type="button" onClick={editingPackage ? publishPackage : publish} disabled={issues.length > 0}>
            Publish
          </button>
        ) : null}
      </div>

      <section className="panel spacious pattern-panel">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Rule patterns</p>
            <h3>Pick the behavior, not the engine term</h3>
          </div>
        </div>
        <div className="pattern-grid">
          {PATTERNS.map((pattern) => (
            <button className="pattern-card" type="button" key={pattern.kind} onClick={() => applyPattern(pattern.kind)}>
              <span>{pattern.tag}</span>
              <strong>{pattern.title}</strong>
              <small>{pattern.body}</small>
            </button>
          ))}
        </div>
      </section>

      <section className="panel spacious studio-control-bar">
        <div className="form-grid">
          <label>
            Applies to
            <select value={appliesTo} onChange={(event) => setAppliesTo(event.target.value)}>
              <option value="program_base">Entire program</option>
              <option value="member_add_on">Member add-on package</option>
            </select>
          </label>
          {appliesTo === "member_add_on" ? (
            <label>
              Rule package
              <select value={selectedPackage?.id || ""} onChange={(event) => setSelectedPackageId(event.target.value)}>
                {(program.rulePackages || []).map((pkg) => <option value={pkg.id} key={pkg.id}>{pkg.name}</option>)}
              </select>
            </label>
          ) : null}
        </div>
        {appliesTo === "member_add_on" ? <button type="button" onClick={createPackage}>Create add-on package</button> : null}
      </section>

      <div className="rule-studio-layout">
        <aside className="panel spacious rule-outline-panel">
          <div className="section-heading">
            <div>
              <p className="eyebrow">Rule outline</p>
              <h3>{group.name}</h3>
            </div>
            <StatusPill value={friendlyStrategy(group.strategy)} />
          </div>
          <div className="rule-outline-list">
            {group.rules.map((rule, index) => (
              <button
                className={`rule-outline-item ${selectedRule?.id === rule.id ? "active" : ""}`}
                type="button"
                key={rule.id}
                onClick={() => setSelectedRuleId(rule.id)}
              >
                <span>{String(index + 1).padStart(2, "0")}</span>
                <strong>{rule.name}</strong>
                <small>{ruleSummary(rule, editableRules)}</small>
              </button>
            ))}
          </div>
          <div className="button-row">
            <button type="button" onClick={() => moveSelectedRule(-1)} disabled={!selectedRule}>Move up</button>
            <button type="button" onClick={() => moveSelectedRule(1)} disabled={!selectedRule}>Move down</button>
          </div>
          <div className="button-row">
            <button className="primary" type="button" onClick={addRule}>Add earn rule</button>
            <button type="button" onClick={addOverflowRule} disabled={!selectedRule || !hasBasisCap(selectedRule)}>
              Add overflow rule
            </button>
          </div>
          <div className="issue-box">
            <strong>{issues.length ? "Review before publish" : "Rule graph is valid"}</strong>
            {issues.length ? issues.map((issue) => <span key={issue}>{issue}</span>) : <span>No validation issues detected.</span>}
          </div>
        </aside>

        <section className="panel spacious selected-rule-editor">
          <div className="section-heading">
            <div>
              <p className="eyebrow">Selected rule</p>
              <h3>{selectedRule?.name || "No rule selected"}</h3>
            </div>
            {selectedRule ? <StatusPill value={selectedRule.status} /> : null}
          </div>
          {selectedRule ? (
            <>
              <RuleBasics rule={selectedRule} updateRule={updateSelectedRule} />
              <RuleFormula rule={selectedRule} updateRule={updateSelectedRule} updateGroup={updateGroup} />
              <RuleLimit rule={selectedRule} updateRule={updateSelectedRule} group={group} />
              <div className="rule-danger-row">
                <button type="button" onClick={() => updateSelectedRule({ status: selectedRule.status === "active" ? "inactive" : "active" })}>
                  {selectedRule.status === "active" ? "Disable rule" : "Enable rule"}
                </button>
                <button type="button" onClick={deleteSelectedRule} disabled={group.rules.length === 1}>Delete draft rule</button>
              </div>
            </>
          ) : (
            <p className="helper-copy">Add an earn rule to start configuring this program.</p>
          )}
        </section>
      </div>

      <section className="panel spacious rule-example-panel">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Live review</p>
            <h3>{preview.title}</h3>
          </div>
          <StatusPill value={`${preview.totalPoints} pts`} />
        </div>
        <div className="example-input-grid">
          <label>Total<input value={example.total} onChange={(event) => setExample({ ...example, total: event.target.value })} /></label>
          <label>Subtotal<input value={example.subtotal} onChange={(event) => setExample({ ...example, subtotal: event.target.value })} /></label>
          <label>Eligible amount<input value={example.eligible} onChange={(event) => setExample({ ...example, eligible: event.target.value })} /></label>
          <label>Category amount<input value={example.grocery} onChange={(event) => setExample({ ...example, grocery: event.target.value })} /></label>
        </div>
        <div className="rule-example-summary">
          {preview.amounts.map((amount) => (
            <div key={amount.label}>
              <span>{amount.label}</span>
              <strong>{amount.value}</strong>
            </div>
          ))}
        </div>
        <div className="rule-example-steps">
          {preview.steps.map((step) => (
            <article className={step.selected ? "selected" : ""} key={step.name}>
              <div>
                <strong>{step.name}</strong>
                <StatusPill value={step.selected ? "selected" : step.status} />
              </div>
              <span>{step.formula}</span>
              <small>{step.detail}</small>
            </article>
          ))}
        </div>
        <p className="helper-copy">{preview.note}</p>
      </section>

      <details className="panel spacious review-graph-panel">
        <summary>
          <span>Review graph</span>
          <StatusPill value={friendlyStrategy(group.strategy)} />
        </summary>
        <div className="rule-canvas">
          <div className="graph-node parent">
            <span>Rule group</span>
            <strong>{group.name}</strong>
            <small>{friendlyStrategy(group.strategy)}</small>
          </div>
          <div className="graph-branch">
            {group.rules.map((rule) => (
              <article className="graph-node rule-node" key={rule.id}>
                <div>
                  <strong>{rule.name}</strong>
                  <StatusPill value={rule.status} />
                </div>
                <span>{rule.type === "points_per_dollar" ? `${rule.pointsPerDollar} pt / $` : `${rule.points} pts fixed`}</span>
                <small>{rule.category || "All transactions"}</small>
                {rule.limit?.enabled || rule.cap ? <em>{rule.limit?.enabled ? capFromLimit(rule.limit) : rule.cap}</em> : null}
              </article>
            ))}
          </div>
        </div>
      </details>
    </section>
  );
}

function RuleBasics({ rule, updateRule }) {
  const selectedCategory = CATEGORY_OPTIONS.some((option) => option.value === rule.category) ? rule.category : "custom";
  return (
    <div className="rule-editor-section">
      <div>
        <p className="eyebrow">Scope</p>
        <h4>What does this rule apply to?</h4>
      </div>
      <div className="form-grid">
        <label>Name<input value={rule.name} onChange={(event) => updateRule({ name: event.target.value })} /></label>
        <label>Rule key<input value={rule.key} onChange={(event) => updateRule({ key: event.target.value })} /></label>
        <label>
          Applies to
          <select value={selectedCategory} onChange={(event) => updateRule({ category: event.target.value === "custom" ? "" : event.target.value })}>
            {CATEGORY_OPTIONS.map((option) => <option value={option.value} key={option.value}>{option.label}</option>)}
          </select>
        </label>
        {selectedCategory === "custom" ? <label>Custom category<input value={rule.category || ""} onChange={(event) => updateRule({ category: event.target.value })} /></label> : null}
        <label>
          Calculation amount
          <select value={rule.basis || ""} onChange={(event) => updateRule({ basis: event.target.value })}>
            <option value="">Use program default</option>
            <option value="total">Total paid</option>
            <option value="subtotal">Subtotal before tax</option>
            <option value="eligible">Eligible amount sent by partner</option>
          </select>
        </label>
      </div>
    </div>
  );
}

function RuleFormula({ rule, updateRule, updateGroup }) {
  const valueLabel = rule.type === "points_per_dollar" ? "Points per dollar" : "Fixed points";
  function changeInteraction(mode) {
    updateRule({ interaction: { ...(rule.interaction || {}), mode } });
    updateGroup({ strategy: mode === "better_of" ? "max_of" : mode === "overflow_after_cap" ? "waterfall" : "stack" });
  }

  function changeType(type) {
    updateRule(type === "first_purchase_bonus" ? { type, category: "first purchase" } : { type });
  }

  return (
    <div className="rule-editor-section">
      <div>
        <p className="eyebrow">Formula</p>
        <h4>How many points should this create?</h4>
      </div>
      <div className="form-grid">
        <label>
          Earn formula
          <select value={rule.type} onChange={(event) => changeType(event.target.value)}>
            <option value="points_per_dollar">Points per dollar</option>
            <option value="fixed_per_transaction">Fixed per transaction</option>
            <option value="first_purchase_bonus">First purchase bonus</option>
          </select>
        </label>
        <label>
          {valueLabel}
          <input
            min="0"
            step="0.01"
            type="number"
            value={rule.type === "points_per_dollar" ? rule.pointsPerDollar : rule.points || 0}
            onChange={(event) => updateRule(rule.type === "points_per_dollar" ? { pointsPerDollar: event.target.value } : { points: Number(event.target.value) })}
          />
        </label>
        <label>
          When another rule also matches
          <select value={rule.interaction?.mode || "adds"} onChange={(event) => changeInteraction(event.target.value)}>
            <option value="adds">Add these points</option>
            <option value="better_of">Use only if this is better</option>
            <option value="overflow_after_cap">Run after another rule hits its cap</option>
          </select>
        </label>
      </div>
    </div>
  );
}

function RuleLimit({ rule, updateRule, group }) {
  const limit = rule.limit || limitFromCap(rule.cap);
  const cappedRuleOptions = group.rules.filter((candidate) => candidate.key !== rule.key && hasBasisCap(candidate));

  function updateLimit(patch) {
    const nextLimit = { ...limit, ...patch };
    updateRule({ limit: nextLimit, cap: capFromLimit(nextLimit) });
  }

  return (
    <div className="rule-editor-section">
      <div>
        <p className="eyebrow">Caps and dependencies</p>
        <h4>Control runaway earn before publishing</h4>
      </div>
      <label className="switch-row">
        <input type="checkbox" checked={Boolean(limit.enabled)} onChange={(event) => updateLimit({ enabled: event.target.checked })} />
        Limit how much this rule can award
      </label>
      {limit.enabled ? (
        <div className="form-grid cap-grid">
          <label>
            Cap type
            <select value={limit.metric} onChange={(event) => updateLimit({ metric: event.target.value })}>
              <option value="points">Points awarded</option>
              <option value="basis">Spend counted by this rule</option>
            </select>
          </label>
          <label>
            Amount
            <input type="number" min="0" step="1" value={limit.amount} onChange={(event) => updateLimit({ amount: Number(event.target.value) })} />
          </label>
          <label>
            Period
            <select value={limit.period} onChange={(event) => updateLimit({ period: event.target.value })}>
              <option value="calendar_month">Calendar month</option>
              <option value="day">Day</option>
              <option value="lifetime">Lifetime</option>
            </select>
          </label>
          <label>
            Scope
            <select value={limit.scope || "member"} onChange={(event) => updateLimit({ scope: event.target.value })}>
              <option value="member">Per member</option>
            </select>
          </label>
        </div>
      ) : null}
      {rule.interaction?.mode === "overflow_after_cap" ? (
        <label>
          Starts after
          <select
            value={rule.interaction.dependsOnRuleKey || cappedRuleOptions[0]?.key || ""}
            onChange={(event) => updateRule({
              cap: "after cap",
              interaction: { ...rule.interaction, dependsOnRuleKey: event.target.value },
              dependencies: [{ dependsOnRuleKey: event.target.value, dependencyType: "requires_exhausted" }],
            })}
          >
            {cappedRuleOptions.map((candidate) => <option value={candidate.key} key={candidate.key}>{candidate.name}</option>)}
          </select>
        </label>
      ) : null}
    </div>
  );
}

function packageToRules(pkg) {
  return {
    earnBasis: "total",
    groups: [{
      id: `${pkg.id}-group`,
      name: pkg.name,
      strategy: "stack",
      status: pkg.status,
      rules: pkg.rules,
    }],
  };
}

function rulesToPackagePatch(rules) {
  const group = rules.groups[0];
  return {
    name: group.name,
    rules: group.rules,
  };
}

function ruleCalculationPreview(rules, example) {
  const group = rules.groups?.[0] || { strategy: "stack", rules: [] };
  const amounts = {
    total: dollarsToMinor(example.total),
    subtotal: dollarsToMinor(example.subtotal),
    eligible: dollarsToMinor(example.eligible),
    grocery: dollarsToMinor(example.grocery),
  };
  const basisName = rules.earnBasis || "total";
  const baseBasis = amounts[basisName] || amounts.total;
  const exampleAmounts = [
    { label: "Total", value: formatMoney(amounts.total) },
    { label: "Subtotal", value: formatMoney(amounts.subtotal) },
    { label: "Eligible amount", value: formatMoney(amounts.eligible) },
    { label: "Category amount", value: formatMoney(amounts.grocery) },
  ];
  if (group.strategy === "waterfall") {
    let remainingCategory = amounts.grocery;
    let totalPoints = 0;
    let priorExhausted = false;
    const steps = group.rules.map((rule, index) => {
      const limit = limitsFromRule(rule).find((item) => item.maxBasisAmountMinor);
      const wantsAfterCap = rule.interaction?.mode === "overflow_after_cap" || String(rule.cap || "").toLowerCase() === "after cap" || (rule.dependencies || []).some((dep) => dep.dependencyType === "requires_exhausted");
      if (wantsAfterCap && !priorExhausted) {
        return {
          name: rule.name,
          selected: false,
          status: "waiting",
          points: 0,
          formula: "Runs after the prior spend cap is exhausted",
          detail: "This rule receives no basis until the capped rule before it has consumed its configured spend slice.",
        };
      }
      const rawBasis = categoryBasis(rule, amounts, baseBasis, rules);
      const availableBasis = rule.category && rule.category !== "All transactions" ? remainingCategory : rawBasis;
      const basis = limit ? Math.min(availableBasis, limit.maxBasisAmountMinor) : availableBasis;
      const points = pointsForRule(rule, basis);
      if (rule.category && rule.category !== "All transactions") remainingCategory = Math.max(0, remainingCategory - basis);
      priorExhausted = Boolean(limit && availableBasis > limit.maxBasisAmountMinor);
      totalPoints += points;
      return {
        name: rule.name,
        selected: points > 0,
        status: priorExhausted ? "cap hit" : "applied",
        points,
        formula: formulaForRule(rule, basis, points),
        detail: limit
          ? `${formatMoney(limit.maxBasisAmountMinor)} of spend can use this higher rate. ${formatMoney(remainingCategory)} remains for overflow rules.`
          : index > 0 ? "This receives the amount left after earlier waterfall rules consume their capped share." : "No spend cap on this rule.",
      };
    });
    return {
      title: "High rate first, then overflow",
      amounts: exampleAmounts,
      steps,
      totalPoints,
      note: "Spend caps limit the purchase amount a rule can use. Overflow rules run only after the capped rule is exhausted.",
    };
  }

  const steps = group.rules.map((rule) => {
    const basis = categoryBasis(rule, amounts, baseBasis, rules);
    const rawPoints = pointsForRule(rule, basis);
    const maxPoints = limitsFromRule(rule).find((item) => item.maxPoints)?.maxPoints;
    const points = maxPoints ? Math.min(rawPoints, maxPoints) : rawPoints;
    return {
      name: rule.name,
      selected: group.strategy === "stack" ? points > 0 : false,
      status: maxPoints && rawPoints > maxPoints ? "capped" : "candidate",
      points,
      formula: rule.type === "points_per_dollar"
        ? `${formatMoney(basis)} x ${rule.pointsPerDollar || 0} pt / $ = ${rawPoints} pts${maxPoints ? `, capped to ${points}` : ""}`
        : `${rule.points || 0} fixed pts`,
      detail: rule.limit?.enabled ? `Points cap: ${capFromLimit(rule.limit)}.` : `Uses ${friendlyBasis(rule.basis || basisName)} unless the rule targets a category amount.`,
    };
  });
  if (group.strategy === "max_of") {
    const bestIndex = steps.reduce((best, step, index) => (best === -1 || step.points > steps[best].points ? index : best), -1);
    steps.forEach((step, index) => {
      step.selected = index === bestIndex && step.points > 0;
      step.status = step.selected ? "highest" : step.status;
    });
  }
  const totalPoints = group.strategy === "max_of"
    ? steps.find((step) => step.selected)?.points || 0
    : steps.reduce((sum, step) => sum + step.points, 0);
  return {
    title: group.strategy === "max_of" ? "Better rate wins" : "Matching rules add together",
    amounts: exampleAmounts,
    steps,
    totalPoints,
    note: group.strategy === "max_of" ? "When more than one rule matches, only the highest award posts." : "Every matching active rule contributes points.",
  };
}

function hasBasisCap(rule) {
  return limitsFromRule(rule).some((limit) => limit.maxBasisAmountMinor);
}

function ruleSummary(rule, rules) {
  const basis = friendlyBasis(rule.basis || rules.earnBasis);
  const category = rule.category && rule.category !== "All transactions" ? rule.category : "all purchases";
  if (rule.type === "points_per_dollar") return `${rule.pointsPerDollar || 0} pt / $ on ${category} using ${basis}`;
  return `${rule.points || 0} fixed pts on ${category}`;
}

function categoryBasis(rule, amounts, baseBasis, rules) {
  if (rule.category && !["All transactions", "first purchase"].includes(rule.category)) return amounts.grocery;
  const basis = rule.basis || rules.earnBasis;
  return amounts[basis] || baseBasis;
}

function formulaForRule(rule, basis, points) {
  if (rule.type === "points_per_dollar") return `${formatMoney(basis)} x ${rule.pointsPerDollar || 0} pt / $ = ${points} pts`;
  return `${rule.points || 0} fixed pts`;
}

function pointsForRule(rule, basisMinor) {
  if (rule.type !== "points_per_dollar") return Number(rule.points || 0);
  return Math.floor((basisMinor / 100) * Number(rule.pointsPerDollar || 0));
}

function friendlyStrategy(strategy) {
  if (strategy === "max_of") return "Better rate wins";
  if (strategy === "waterfall") return "High rate then overflow";
  return "Adds together";
}

function friendlyBasis(basis) {
  if (basis === "subtotal") return "subtotal before tax";
  if (basis === "eligible") return "eligible amount";
  return "total paid";
}

function dollarsToMinor(value) {
  const amount = Number(String(value).replace(/[^0-9.]/g, ""));
  if (!Number.isFinite(amount)) return 0;
  return Math.round(amount * 100);
}

function formatMoney(minor) {
  return `$${(minor / 100).toFixed(2)}`;
}
