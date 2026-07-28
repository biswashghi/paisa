export function createProgramDraft(index) {
  return {
    id: `program-${Date.now()}`,
    name: `Rewards Program ${index}`,
    tierCode: `tier-${index}`,
    status: "draft",
    members: 0,
    liabilityPoints: 0,
    validationScore: 0,
    rules: createRulesTemplate("base"),
    rulePackages: [],
  };
}

export function createRulesTemplate(kind) {
  if (kind === "base") {
    return {
      earnBasis: "total",
      groups: [
        {
          id: `group-${Date.now()}`,
          name: "Every purchase earns",
          strategy: "stack",
          status: "draft",
          rules: [
            createRule("Base earn", "base_earn", 1, "All transactions", ""),
          ],
        },
      ],
    };
  }

  if (kind === "stack") {
    return {
      earnBasis: "total",
      groups: [
        {
          id: `group-${Date.now()}`,
          name: "Stacked earn",
          strategy: "stack",
          status: "draft",
          rules: [
            createRule("Base earn", "base_earn", 1, "All transactions", ""),
            { ...createRule("First purchase bonus", "first_purchase_bonus", 0, "first purchase", "once / member"), type: "fixed_per_transaction", points: 75, interaction: { mode: "adds" } },
          ],
        },
      ],
    };
  }

  if (kind === "waterfall") {
    return {
      earnBasis: "total",
      groups: [
        {
          id: `group-${Date.now()}`,
          name: "Capped category waterfall",
          strategy: "waterfall",
          status: "draft",
          rules: [
            { ...createRule("Category boost capped", "category_cap", 5, "grocery", "5000 basis / month"), basis: "eligible" },
            {
              ...createRule("Remainder earn", "remainder_earn", 1, "grocery", "after cap"),
              basis: "eligible",
              interaction: { mode: "overflow_after_cap", dependsOnRuleKey: "category_cap" },
              dependencies: [{ dependsOnRuleKey: "category_cap", dependencyType: "requires_exhausted" }],
            },
          ],
        },
      ],
    };
  }

  return {
    earnBasis: "total",
    groups: [
      {
        id: `group-${Date.now()}`,
        name: "Max earn comparison",
        strategy: "max_of",
        status: "draft",
        rules: [
          createRule("Base earn", "base_earn", 1, "All transactions", ""),
          { ...createRule("Category bonus", "category_bonus", 5, "grocery", "300 pts / month"), basis: "eligible", interaction: { mode: "better_of" } },
        ],
      },
    ],
  };
}

export function createRule(name, key, pointsPerDollar, category, cap) {
  return {
    id: `rule-${Date.now()}-${Math.random().toString(16).slice(2)}`,
    key,
    name,
    type: "points_per_dollar",
    pointsPerDollar,
    category,
    cap,
    limit: limitFromCap(cap),
    interaction: { mode: "adds" },
    dependencies: [],
    status: "active",
  };
}

export function validateProgramRules(program) {
  const issues = [];
  program.rules.groups.forEach((group) => {
    const activeRules = group.rules.filter((rule) => rule.status === "active");
    if (group.strategy === "max_of" && activeRules.length < 2) {
      issues.push(`${group.name} needs at least two active rules for max_of.`);
    }
    group.rules.forEach((rule) => {
      if (!rule.name.trim()) issues.push("A rule is missing a display name.");
      if (!rule.key.trim()) issues.push(`${rule.name || "A rule"} is missing a rule key.`);
      if (rule.type === "points_per_dollar" && Number(rule.pointsPerDollar) <= 0) {
        issues.push(`${rule.name} needs points per dollar above zero.`);
      }
      if (rule.type !== "points_per_dollar" && Number(rule.points) <= 0) {
        issues.push(`${rule.name} needs fixed points above zero.`);
      }
    });
  });
  return issues;
}

export function rulesToPayload(program) {
  return {
    earnBasis: program.rules.earnBasis,
    ruleGroups: program.rules.groups.map((group, groupIndex) => ({
      name: group.name,
      resolutionStrategy: group.strategy,
      priority: groupIndex + 1,
      rules: group.rules.map((rule, ruleIndex) => {
        const eligibilityConfig = { basis: rule.basis || program.rules.earnBasis };
        if (rule.category && rule.category !== "All transactions" && rule.category !== "first purchase") {
          eligibilityConfig.categories = [rule.category];
        }
        if (rule.category === "first purchase") {
          eligibilityConfig.firstPurchase = true;
        }
        const limits = limitsFromRule(rule);
        return {
          ruleKey: rule.key,
          name: rule.name,
          ruleType: rule.type,
          priority: ruleIndex + 1,
          status: rule.status,
          eligibilityConfig,
          formulaConfig: rule.type === "points_per_dollar" ? { pointsPerDollar: Number(rule.pointsPerDollar) } : { points: Number(rule.points) },
          limits,
          dependencies: dependenciesFromRule(rule, group.rules),
        };
      }),
    })),
  };
}

export function limitsFromRule(rule) {
  if (rule.limit?.enabled) return limitsFromLimit(rule.limit);
  return limitsFromCap(rule.cap);
}

export function limitsFromCap(cap = "") {
  return limitsFromLimit(limitFromCap(cap));
}

export function limitFromCap(cap = "") {
  const normalized = String(cap).trim().toLowerCase();
  if (!normalized || normalized === "after cap" || normalized.includes("once")) {
    return { enabled: false, metric: "points", amount: 300, period: "calendar_month", scope: "member" };
  }
  const amount = Number(normalized.match(/\d+(?:\.\d+)?/)?.[0] || 0);
  if (!amount) return { enabled: false, metric: "points", amount: 300, period: "calendar_month", scope: "member" };
  const period = normalized.includes("month") ? "calendar_month" : normalized.includes("day") ? "day" : "lifetime";
  const metric = normalized.includes("basis") ? "basis" : "points";
  return { enabled: true, metric, amount, period, scope: "member" };
}

export function capFromLimit(limit) {
  if (!limit?.enabled) return "";
  const metricLabel = limit.metric === "basis" ? "basis" : "pts";
  const periodLabel = limit.period === "calendar_month" ? "month" : limit.period;
  return `${Number(limit.amount || 0)} ${metricLabel} / ${periodLabel}`;
}

export function limitsFromLimit(limit) {
  if (!limit?.enabled || Number(limit.amount) <= 0) return [];
  const output = {
    scope: limit.scope || "member",
    period: limit.period || "calendar_month",
  };
  if (limit.metric === "basis") {
    output.maxBasisAmountMinor = Math.round(Number(limit.amount));
  } else {
    output.maxPoints = Math.round(Number(limit.amount));
  }
  return [output];
}

export function dependenciesFromRule(rule, rules) {
  if (rule.interaction?.mode === "overflow_after_cap" && rule.interaction.dependsOnRuleKey) {
    return [{ dependsOnRuleKey: rule.interaction.dependsOnRuleKey, dependencyType: "requires_exhausted" }];
  }
  if (rule.dependencies?.length) return rule.dependencies;
  return dependenciesFromCap(rule, rules);
}

export function dependenciesFromCap(rule, rules) {
  if (String(rule.cap || "").trim().toLowerCase() !== "after cap") return [];
  const priorCappedRule = [...rules].reverse().find((candidate) => candidate.key !== rule.key && limitsFromCap(candidate.cap).some((limit) => limit.maxBasisAmountMinor));
  if (!priorCappedRule) return [];
  return [{ dependsOnRuleKey: priorCappedRule.key, dependencyType: "requires_exhausted" }];
}
