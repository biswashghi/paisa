import { createRule, createRulesTemplate, rulesToPayload, validateProgramRules } from "../utils/rules.js";
import StatusPill from "./StatusPill.jsx";
import { useState } from "react";

export default function RuleStudio({ program, onUpdateProgram, onPublishProgramRules, onCreateRulePackage, onUpdateRulePackage, onPublishRulePackage, embedded = false }) {
  const [appliesTo, setAppliesTo] = useState("program_base");
  const [selectedPackageId, setSelectedPackageId] = useState(program.rulePackages?.[0]?.id || "");
  const selectedPackage = (program.rulePackages || []).find((pkg) => pkg.id === selectedPackageId) || program.rulePackages?.[0];
  const editingPackage = appliesTo === "member_add_on" && selectedPackage;
  const editableRules = editingPackage ? packageToRules(selectedPackage) : program.rules;
  const editableProgram = editingPackage ? { ...program, rules: editableRules } : program;
  const issues = validateProgramRules(editableProgram);
  const group = editableRules.groups[0];

  function updateRules(nextRules) {
    if (editingPackage) {
      onUpdateRulePackage(program.id, selectedPackage.id, { ...rulesToPackagePatch(nextRules), status: "draft" });
      return;
    }
    onUpdateProgram(program.id, { rules: nextRules, status: "draft" });
  }

  function updateGroup(patch) {
    updateRules({ ...editableRules, groups: [{ ...group, ...patch }] });
  }

  function updateRule(ruleId, patch) {
    updateGroup({ rules: group.rules.map((rule) => rule.id === ruleId ? { ...rule, ...patch } : rule) });
  }

  function addRule() {
    updateGroup({ rules: [...group.rules, createRule("New earn rule", `rule_${group.rules.length + 1}`, 1, "All transactions", "")] });
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
      {!embedded ? (
        <div className="view-header">
          <div>
            <p className="eyebrow">Rules</p>
            <h2>{editingPackage ? selectedPackage.name : program.name}</h2>
            <div className="studio-scope-line">
              <span>{appliesTo === "program_base" ? "Base graph" : "Member add-on"}</span>
              <span>{group?.rules.length || 0} rules</span>
              <span>{issues.length ? `${issues.length} issues` : "Valid"}</span>
            </div>
          </div>
          <div className="button-row">
            <button type="button" onClick={() => updateRules(createRulesTemplate("max_of"))}>Max-of</button>
            <button type="button" onClick={() => updateRules(createRulesTemplate("stack"))}>Stack</button>
            <button type="button" onClick={() => updateRules(createRulesTemplate("waterfall"))}>Waterfall</button>
            <button className="primary" type="button" onClick={editingPackage ? publishPackage : publish} disabled={issues.length > 0}>Publish</button>
          </div>
        </div>
      ) : (
        <div className="button-row">
          <button type="button" onClick={() => updateRules(createRulesTemplate("max_of"))}>Max-of</button>
          <button type="button" onClick={() => updateRules(createRulesTemplate("stack"))}>Stack</button>
          <button type="button" onClick={() => updateRules(createRulesTemplate("waterfall"))}>Waterfall</button>
          <button className="primary" type="button" onClick={editingPackage ? publishPackage : publish} disabled={issues.length > 0}>Publish</button>
        </div>
      )}

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

      <div className="rule-workspace">
        <section className="panel spacious rule-canvas-panel">
          <div className="section-heading">
            <div>
              <p className="eyebrow">Graph</p>
              <h3>{group?.name}</h3>
            </div>
            <StatusPill value={issues.length ? "Needs changes" : "Validated"} />
          </div>
          <div className="graph-legend">
            <span>Group resolver</span>
            <span>Earn candidate</span>
            <span>Cap or dependency</span>
          </div>
          <div className="rule-canvas">
            <div className="graph-node parent">
              <span>Rule group</span>
              <strong>{group?.name}</strong>
              <small>Strategy: {group?.strategy}</small>
            </div>
            <div className="graph-branch">
              {group?.rules.map((rule) => (
                <article className="graph-node rule-node" key={rule.id}>
                  <div>
                    <strong>{rule.name}</strong>
                    <StatusPill value={rule.status} />
                  </div>
                  <span>{rule.type === "points_per_dollar" ? `${rule.pointsPerDollar} pt / $` : `${rule.points} pts fixed`}</span>
                  <small>{rule.category || "All transactions"}</small>
                  {rule.cap ? <em>{rule.cap}</em> : null}
                </article>
              ))}
            </div>
          </div>
        </section>

        <aside className="panel spacious">
          <div className="section-heading">
            <div>
              <p className="eyebrow">Config</p>
              <h3>Rule group</h3>
            </div>
          </div>
          <div className="form-grid single">
            <label>
              Group name
              <input value={group?.name || ""} onChange={(event) => updateGroup({ name: event.target.value })} />
            </label>
            <label>
              Strategy
              <select value={group?.strategy || "stack"} onChange={(event) => updateGroup({ strategy: event.target.value })}>
                <option value="max_of">Max of</option>
                <option value="stack">Stack</option>
                <option value="waterfall">Waterfall</option>
              </select>
            </label>
            <label>
              Earn basis
              <select value={editableRules.earnBasis} onChange={(event) => updateRules({ ...editableRules, earnBasis: event.target.value })}>
                <option value="eligible">Eligible amount</option>
                <option value="subtotal">Subtotal</option>
                <option value="total">Total</option>
              </select>
            </label>
          </div>
          <button type="button" onClick={addRule}>Add rule</button>
          <div className="issue-box">
            <strong>{issues.length ? "Validation issues" : "Ready to publish"}</strong>
            {issues.length ? issues.map((issue) => <span key={issue}>{issue}</span>) : <span>No issues detected.</span>}
          </div>
        </aside>
      </div>

      <section className="panel spacious">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Rules</p>
            <h3>Configure earn candidates</h3>
          </div>
        </div>
        <div className="rule-editor-list">
          {group?.rules.map((rule) => (
            <article className="rule-edit-card" key={rule.id}>
              <div className="rule-card-title">
                <strong>{rule.name}</strong>
                <StatusPill value={rule.status} />
              </div>
              <label>Name<input value={rule.name} onChange={(event) => updateRule(rule.id, { name: event.target.value })} /></label>
              <label>Rule key<input value={rule.key} onChange={(event) => updateRule(rule.id, { key: event.target.value })} /></label>
              <label>
                Type
                <select value={rule.type} onChange={(event) => updateRule(rule.id, { type: event.target.value })}>
                  <option value="points_per_dollar">Points per dollar</option>
                  <option value="fixed_per_transaction">Fixed per transaction</option>
                  <option value="first_purchase_bonus">First purchase bonus</option>
                </select>
              </label>
              <label>
                Value
                <input type="number" value={rule.type === "points_per_dollar" ? rule.pointsPerDollar : rule.points || 0} onChange={(event) => updateRule(rule.id, rule.type === "points_per_dollar" ? { pointsPerDollar: Number(event.target.value) } : { points: Number(event.target.value) })} />
              </label>
              <label>Category<input value={rule.category} onChange={(event) => updateRule(rule.id, { category: event.target.value })} /></label>
              <label>Cap<input value={rule.cap} onChange={(event) => updateRule(rule.id, { cap: event.target.value })} /></label>
            </article>
          ))}
        </div>
      </section>

    </section>
  );
}

function packageToRules(pkg) {
  return {
    earnBasis: "eligible",
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
