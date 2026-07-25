export function createProgramDraft(index) {
  return {
    id: `program-${Date.now()}`,
    name: `Rewards Program ${index}`,
    tierCode: `tier-${index}`,
    status: "draft",
    members: 0,
    liabilityPoints: 0,
    validationScore: 0,
    rules: createRulesTemplate("max_of"),
    rulePackages: [],
  };
}

export function createRulesTemplate(kind) {
  if (kind === "stack") {
    return {
      earnBasis: "eligible",
      groups: [
        {
          id: `group-${Date.now()}`,
          name: "Stacked earn",
          strategy: "stack",
          status: "draft",
          rules: [
            createRule("Base earn", "base_earn", 1, "All transactions", ""),
            { ...createRule("First purchase bonus", "first_purchase_bonus", 0, "first purchase", "once / member"), type: "fixed_per_transaction", points: 75 },
          ],
        },
      ],
    };
  }

  if (kind === "waterfall") {
    return {
      earnBasis: "eligible",
      groups: [
        {
          id: `group-${Date.now()}`,
          name: "Capped category waterfall",
          strategy: "waterfall",
          status: "draft",
          rules: [
            createRule("Category boost capped", "category_cap", 5, "grocery", "5000 basis / month"),
            createRule("Remainder earn", "remainder_earn", 1, "grocery", "after cap"),
          ],
        },
      ],
    };
  }

  return {
    earnBasis: "eligible",
    groups: [
      {
        id: `group-${Date.now()}`,
        name: "Max earn comparison",
        strategy: "max_of",
        status: "draft",
        rules: [
          createRule("Base earn", "base_earn", 1, "All transactions", ""),
          createRule("Category bonus", "category_bonus", 5, "grocery", "300 pts / month"),
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
        const eligibilityConfig = {};
        if (rule.category && rule.category !== "All transactions" && rule.category !== "first purchase") {
          eligibilityConfig.categories = [rule.category];
        }
        if (rule.category === "first purchase") {
          eligibilityConfig.firstPurchase = true;
        }
        return {
          ruleKey: rule.key,
          name: rule.name,
          ruleType: rule.type,
          priority: ruleIndex + 1,
          status: rule.status,
          eligibilityConfig,
          formulaConfig: rule.type === "points_per_dollar" ? { pointsPerDollar: Number(rule.pointsPerDollar) } : { points: Number(rule.points) },
        };
      }),
    })),
  };
}
