#!/usr/bin/env node

const API_BASE = process.env.PAISA_API_BASE || "http://127.0.0.1:8081";
const RUN_ID = process.env.PAISA_E2E_RUN_ID || new Date().toISOString().replace(/[-:.TZ]/g, "").slice(0, 14);
const PARTNER_COUNT = Number(process.env.PAISA_E2E_PARTNERS || 3);
const PROGRAMS_PER_PARTNER = Number(process.env.PAISA_E2E_PROGRAMS_PER_PARTNER || 2);
const TOTAL_MEMBERS = Number(process.env.PAISA_E2E_MEMBERS || 1000);
const TOTAL_TRANSACTIONS = Number(process.env.PAISA_E2E_TRANSACTIONS || 10000);
const TOTAL_REDEMPTIONS = Number(process.env.PAISA_E2E_REDEMPTIONS || 1000);
const CREATE_CONCURRENCY = Number(process.env.PAISA_E2E_CREATE_CONCURRENCY || 24);
const INGEST_CONCURRENCY = Number(process.env.PAISA_E2E_INGEST_CONCURRENCY || 40);
const REDEEM_CONCURRENCY = Number(process.env.PAISA_E2E_REDEEM_CONCURRENCY || 24);

const started = Date.now();

const summary = {
  apiBase: API_BASE,
  runId: RUN_ID,
  requested: {
    partners: PARTNER_COUNT,
    programsPerPartner: PROGRAMS_PER_PARTNER,
    members: TOTAL_MEMBERS,
    transactions: TOTAL_TRANSACTIONS,
    redemptions: TOTAL_REDEMPTIONS,
  },
  created: {
    partners: 0,
    programs: 0,
    ruleVersions: 0,
    catalogItems: 0,
    members: 0,
    transactions: 0,
    redemptionsReserved: 0,
    redemptionsValidated: 0,
    redemptionsCaptured: 0,
  },
  processed: {
    events: 0,
    failed: 0,
    iterations: 0,
  },
  verified: {
    transactions: 0,
    redemptions: 0,
    capturedRedemptions: 0,
    sampleBalances: [],
  },
  timingsMs: {},
  notes: [],
};

function log(message) {
  const elapsed = ((Date.now() - started) / 1000).toFixed(1).padStart(6, " ");
  console.log(`[${elapsed}s] ${message}`);
}

async function request(path, { method = "GET", token = "", body } = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    method,
    headers: {
      "content-type": "application/json",
      ...(token ? { authorization: `Bearer ${token}` } : {}),
    },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  const text = await response.text();
  const payload = text ? JSON.parse(text) : null;
  if (!response.ok) {
    const error = new Error(payload?.error || `HTTP ${response.status} ${response.statusText}`);
    error.status = response.status;
    error.payload = payload;
    error.path = path;
    throw error;
  }
  return payload;
}

async function mapLimit(items, limit, worker) {
  const results = new Array(items.length);
  let index = 0;
  const workers = Array.from({ length: Math.min(limit, items.length) }, async () => {
    for (;;) {
      const current = index;
      index += 1;
      if (current >= items.length) {
        break;
      }
      results[current] = await worker(items[current], current);
    }
  });
  await Promise.all(workers);
  return results;
}

function distribute(total, buckets) {
  return Array.from({ length: buckets }, (_, index) => {
    const base = Math.floor(total / buckets);
    return base + (index < total % buckets ? 1 : 0);
  });
}

function ruleVersionPayload(partnerIndex, programIndex) {
  const rate = 10 + partnerIndex + programIndex;
  return {
    scope: "program_base",
    earnBasis: "eligible",
    name: `Published earn graph ${partnerIndex + 1}-${programIndex + 1}`,
    description: "E2E load-test graph: one active stack rule for deterministic earning.",
    ruleGroups: [
      {
        name: "Everyday earn",
        resolutionStrategy: "stack",
        priority: 1,
        rules: [
          {
            ruleKey: `everyday_${partnerIndex + 1}_${programIndex + 1}`,
            name: `${rate} points per dollar`,
            ruleType: "points_per_dollar",
            priority: 1,
            status: "active",
            eligibilityConfig: {},
            formulaConfig: { pointsPerDollar: rate },
            limits: [],
            dependencies: [],
          },
        ],
      },
    ],
  };
}

function purchasePayload(member, txIndex) {
  const category = txIndex % 3 === 0 ? "grocery" : txIndex % 3 === 1 ? "coffee" : "general";
  const subtotalMinor = 10000;
  const taxMinor = 600;
  return {
    externalTransactionId: `e2e-${RUN_ID}-${member.partnerIndex}-${member.memberIndex}-${txIndex}`,
    externalCustomerId: member.externalCustomerId,
    type: "purchase",
    currency: "USD",
    subtotalMinor,
    taxMinor,
    totalMinor: subtotalMinor + taxMinor,
    eligibleMinor: subtotalMinor,
    occurredAt: new Date(Date.now() - txIndex * 1000).toISOString(),
    lineItems: [
      {
        externalLineId: `line-${txIndex}`,
        sku: `sku-${category}`,
        category,
        quantity: 1,
        subtotalMinor,
        taxMinor,
        totalMinor: subtotalMinor + taxMinor,
        eligibleMinor: subtotalMinor,
      },
    ],
  };
}

async function setupPartners() {
  const t0 = Date.now();
  const memberDistribution = distribute(TOTAL_MEMBERS, PARTNER_COUNT);
  const partners = [];

  for (let partnerIndex = 0; partnerIndex < PARTNER_COUNT; partnerIndex += 1) {
    const partnerKey = `e2e-store-${RUN_ID}-${partnerIndex + 1}`;
    log(`Logging in partner ${partnerKey}`);
    const login = await request("/partner/v1/auth/login", {
      method: "POST",
      body: {
        partnerKey,
        email: `admin+${partnerKey}@paisa.local`,
        name: `E2E Store ${partnerIndex + 1} Admin`,
      },
    });

    const partner = {
      partnerIndex,
      partnerKey,
      token: login.token,
      partner: login.partner,
      programs: [],
      catalogItems: [],
      members: [],
    };
    summary.created.partners += 1;

    await request("/partner/v1/locations", {
      method: "POST",
      token: partner.token,
      body: {
        name: `Flagship ${partnerIndex + 1}`,
        address: `${100 + partnerIndex} Main Street`,
        timezone: "America/Detroit",
        externalLocationId: `loc-${RUN_ID}-${partnerIndex + 1}`,
      },
    });

    for (let programIndex = 0; programIndex < PROGRAMS_PER_PARTNER; programIndex += 1) {
      const program = await request("/partner/v1/programs", {
        method: "POST",
        token: partner.token,
        body: {
          name: `E2E Program ${partnerIndex + 1}-${programIndex + 1}`,
          tierCode: `tier-${programIndex + 1}`,
          priority: programIndex + 1,
        },
      });
      summary.created.programs += 1;

      const draft = await request(`/partner/v1/programs/${program.id}/rule-versions`, {
        method: "POST",
        token: partner.token,
        body: ruleVersionPayload(partnerIndex, programIndex),
      });
      await request(`/partner/v1/programs/${program.id}/rule-versions/${draft.id}/publish`, {
        method: "POST",
        token: partner.token,
      });
      summary.created.ruleVersions += 1;

      const catalogItem = await request("/partner/v1/catalog-items", {
        method: "POST",
        token: partner.token,
        body: {
          programId: program.id,
          name: `E2E $5 Reward ${partnerIndex + 1}-${programIndex + 1}`,
          description: "Load-test coupon reward",
          pointsCost: 100,
          rewardType: "coupon_code",
          status: "active",
          expiresAfterMinutes: 30,
        },
      });
      summary.created.catalogItems += 1;

      partner.programs.push(program);
      partner.catalogItems.push(catalogItem);
    }

    const targetMembers = memberDistribution[partnerIndex];
    const memberInputs = Array.from({ length: targetMembers }, (_, localIndex) => {
      const globalIndex = partners.reduce((sum, current) => sum + current.members.length, 0) + localIndex;
      const program = partner.programs[localIndex % partner.programs.length];
      return {
        partnerIndex,
        memberIndex: globalIndex,
        externalCustomerId: `customer-${RUN_ID}-${partnerIndex + 1}-${localIndex + 1}`,
        program,
        payload: {
          externalCustomerId: `customer-${RUN_ID}-${partnerIndex + 1}-${localIndex + 1}`,
          programId: program.id,
          identifiers: [
            { type: "email", value: `customer-${partnerIndex + 1}-${localIndex + 1}-${RUN_ID}@example.test` },
            { type: "phone", value: `+1517${String(partnerIndex + 1).padStart(2, "0")}${String(localIndex + 1).padStart(6, "0")}` },
          ],
        },
      };
    });

    log(`Creating ${targetMembers} members for ${partnerKey}`);
    const createdMembers = await mapLimit(memberInputs, CREATE_CONCURRENCY, async (input) => {
      const result = await request("/partner/v1/members", {
        method: "POST",
        token: partner.token,
        body: input.payload,
      });
      summary.created.members += 1;
      return {
        ...input,
        id: result.member.id,
        accountId: result.account.id,
        catalogItem: partner.catalogItems[partner.programs.findIndex((program) => program.id === input.program.id)],
        token: partner.token,
        partnerKey,
      };
    });
    partner.members = createdMembers;
    partners.push(partner);
  }

  summary.timingsMs.setup = Date.now() - t0;
  return partners;
}

async function ingestPurchases(partners) {
  const t0 = Date.now();
  const allMembers = partners.flatMap((partner) => partner.members);
  const txDistribution = distribute(TOTAL_TRANSACTIONS, allMembers.length);
  const txInputs = [];

  for (let memberIndex = 0; memberIndex < allMembers.length; memberIndex += 1) {
    const member = allMembers[memberIndex];
    for (let txIndex = 0; txIndex < txDistribution[memberIndex]; txIndex += 1) {
      txInputs.push({ member, txIndex });
    }
  }

  log(`Ingesting ${txInputs.length} purchase transactions`);
  await mapLimit(txInputs, INGEST_CONCURRENCY, async ({ member, txIndex }, absoluteIndex) => {
    await request("/partner/v1/ingest/transactions", {
      method: "POST",
      token: member.token,
      body: purchasePayload(member, txIndex),
    });
    summary.created.transactions += 1;
    if ((absoluteIndex + 1) % 1000 === 0) {
      log(`Ingested ${absoluteIndex + 1}/${txInputs.length} transactions`);
    }
  });

  summary.timingsMs.ingest = Date.now() - t0;
}

async function processTransactions(partners) {
  const t0 = Date.now();
  const token = partners[0].token;
  log("Processing accepted transaction events");

  for (let iteration = 1; iteration <= 500; iteration += 1) {
    const result = await request("/partner/v1/jobs/process-transaction-events", {
      method: "POST",
      token,
    });
    const processed = Number(result.processed || 0);
    const failed = Number(result.failed || 0);
    summary.processed.events += processed;
    summary.processed.failed += failed;
    summary.processed.iterations = iteration;

    if (processed || failed) {
      log(`Processing batch ${iteration}: processed=${processed}, failed=${failed}, total=${summary.processed.events}`);
    }
    if (failed > 0) {
      throw new Error(`Reward processing failed for ${failed} events in batch ${iteration}`);
    }
    if (processed === 0 && failed === 0) {
      break;
    }
  }

  if (summary.processed.events !== TOTAL_TRANSACTIONS) {
    throw new Error(`Expected ${TOTAL_TRANSACTIONS} processed events, got ${summary.processed.events}`);
  }

  summary.timingsMs.process = Date.now() - t0;
}

async function redeemRewards(partners) {
  const t0 = Date.now();
  const members = partners.flatMap((partner) => partner.members).slice(0, TOTAL_REDEMPTIONS);
  log(`Creating, validating, and capturing ${members.length} redemptions`);

  await mapLimit(members, REDEEM_CONCURRENCY, async (member, index) => {
    const redemption = await request("/pos/v1/redemptions", {
      method: "POST",
      token: member.token,
      body: {
        memberId: member.id,
        catalogItemId: member.catalogItem.id,
      },
    });
    summary.created.redemptionsReserved += 1;

    await request(`/pos/v1/redemptions/${redemption.redemption.id}/validate`, {
      method: "POST",
      token: member.token,
    });
    summary.created.redemptionsValidated += 1;

    await request(`/pos/v1/redemptions/${redemption.redemption.id}/capture`, {
      method: "POST",
      token: member.token,
    });
    summary.created.redemptionsCaptured += 1;

    if ((index + 1) % 100 === 0) {
      log(`Captured ${index + 1}/${members.length} redemptions`);
    }
  });

  summary.timingsMs.redeem = Date.now() - t0;
}

async function verify(partners) {
  const t0 = Date.now();
  log("Verifying partner transaction, redemption, and sample balance state");

  for (const partner of partners) {
    const transactions = await request("/partner/v1/transactions", { token: partner.token });
    const redemptions = await request("/partner/v1/redemptions", { token: partner.token });
    summary.verified.transactions += transactions.length;
    summary.verified.redemptions += redemptions.length;
    summary.verified.capturedRedemptions += redemptions.filter((redemption) => redemption.status === "captured").length;

    const samples = [
      partner.members[0],
      partner.members[Math.floor(partner.members.length / 2)],
      partner.members[partner.members.length - 1],
    ].filter(Boolean);

    for (const member of samples) {
      const balance = await request(`/pos/v1/members/${member.id}/balance`, {
        token: member.token,
      });
      summary.verified.sampleBalances.push({
        partnerKey: partner.partnerKey,
        memberId: member.id,
        availablePoints: balance.availablePoints,
        reservedPoints: balance.reservedPoints,
        expiredPoints: balance.expiredPoints,
      });
    }
  }

  if (summary.verified.transactions !== TOTAL_TRANSACTIONS) {
    throw new Error(`Expected ${TOTAL_TRANSACTIONS} verified transactions, got ${summary.verified.transactions}`);
  }
  if (summary.verified.redemptions !== TOTAL_REDEMPTIONS) {
    throw new Error(`Expected ${TOTAL_REDEMPTIONS} verified redemptions, got ${summary.verified.redemptions}`);
  }
  if (summary.verified.capturedRedemptions !== TOTAL_REDEMPTIONS) {
    throw new Error(`Expected ${TOTAL_REDEMPTIONS} captured redemptions, got ${summary.verified.capturedRedemptions}`);
  }
  if (summary.verified.sampleBalances.some((balance) => balance.reservedPoints !== 0)) {
    throw new Error("Expected all sampled balances to have zero reserved points after capture");
  }

  summary.timingsMs.verify = Date.now() - t0;
}

async function writeReport() {
  const fs = await import("node:fs/promises");
  const path = await import("node:path");
  summary.timingsMs.total = Date.now() - started;
  const reportDir = path.resolve("docs/e2e");
  await fs.mkdir(reportDir, { recursive: true });
  const reportPath = path.join(reportDir, `loyalty-load-${RUN_ID}.json`);
  await fs.writeFile(reportPath, `${JSON.stringify(summary, null, 2)}\n`);
  log(`Wrote report ${reportPath}`);
  return reportPath;
}

async function main() {
  await request("/health");
  log(`Starting loyalty onboarding/load rehearsal against ${API_BASE}`);

  const partners = await setupPartners();
  await ingestPurchases(partners);
  await processTransactions(partners);
  await redeemRewards(partners);
  await verify(partners);

  const reportPath = await writeReport();
  console.log(JSON.stringify({ ok: true, reportPath, summary }, null, 2));
}

main().catch(async (error) => {
  summary.error = {
    message: error.message,
    status: error.status,
    path: error.path,
    payload: error.payload,
  };
  try {
    await writeReport();
  } catch {
    // Ignore report failures while surfacing the original failure.
  }
  console.error(JSON.stringify({ ok: false, error: summary.error, summary }, null, 2));
  process.exit(1);
});
