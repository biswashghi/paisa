export const defaultPartner = {
  id: "partner-acme",
  name: "Acme Retail",
  partnerKey: "acme-retail",
  adminName: "Partner admin",
  adminEmail: "",
  apiEnvironment: "Production",
};

export const initialPrograms = [
  {
    id: "program-gold",
    name: "Acme Rewards Gold",
    tierCode: "gold",
    status: "published",
    members: 84210,
    liabilityPoints: 12845630,
    validationScore: 98.7,
    rules: {
      earnBasis: "eligible",
      groups: [
        {
          id: "group-gold-max",
          name: "Gold max earn",
          strategy: "max_of",
          status: "published",
          rules: [
            {
              id: "rule-base",
              key: "base_earn",
              name: "Base earn",
              type: "points_per_dollar",
              pointsPerDollar: 1,
              category: "All transactions",
              cap: "",
              status: "active",
            },
            {
              id: "rule-grocery",
              key: "grocery_bonus",
              name: "Grocery bonus",
              type: "points_per_dollar",
              pointsPerDollar: 5,
              category: "grocery",
              cap: "300 pts / month",
              status: "active",
            },
          ],
        },
      ],
    },
    rulePackages: [
      {
        id: "pkg-vip-grocery",
        ruleVersionId: "rules-vip-grocery",
        key: "vip_grocery_boost",
        name: "VIP grocery accelerator",
        status: "published",
        description: "Adds 2 extra points per dollar on grocery transactions for selected members.",
        rules: [
          { id: "pkg-rule-grocery", key: "vip_grocery_extra", name: "VIP grocery extra", type: "points_per_dollar", pointsPerDollar: 2, category: "grocery", cap: "200 pts / month", status: "active" },
        ],
      },
      {
        id: "pkg-service-recovery",
        ruleVersionId: "rules-service-recovery",
        key: "service_recovery",
        name: "Service recovery boost",
        status: "draft",
        description: "Temporary fixed earn bonus used by support teams after escalations.",
        rules: [
          { id: "pkg-rule-fixed", key: "support_makegood", name: "Support make-good", type: "fixed_per_transaction", points: 250, category: "all", cap: "once / assignment", status: "active" },
        ],
      },
    ],
  },
  {
    id: "program-everyday",
    name: "Everyday Rewards",
    tierCode: "everyday",
    status: "draft",
    members: 31450,
    liabilityPoints: 4250400,
    validationScore: 91.2,
    rules: {
      earnBasis: "eligible",
      groups: [
        {
          id: "group-everyday-stack",
          name: "Everyday stack",
          strategy: "stack",
          status: "draft",
          rules: [
            {
              id: "rule-everyday-base",
              key: "everyday_base",
              name: "Everyday base earn",
              type: "points_per_dollar",
              pointsPerDollar: 1,
              category: "All transactions",
              cap: "",
              status: "active",
            },
            {
              id: "rule-first",
              key: "first_purchase_bonus",
              name: "First purchase bonus",
              type: "fixed_per_transaction",
              points: 75,
              category: "first purchase",
              cap: "once / member",
              status: "active",
            },
          ],
        },
      ],
    },
    rulePackages: [
      {
        id: "pkg-new-member",
        ruleVersionId: "rules-new-member",
        key: "new_member_apparel",
        name: "New member apparel boost",
        status: "published",
        description: "Adds a short-lived 3x apparel bonus for newly enrolled members.",
        rules: [
          { id: "pkg-rule-apparel", key: "apparel_boost", name: "Apparel boost", type: "points_per_dollar", pointsPerDollar: 3, category: "apparel", cap: "500 pts / 30 days", status: "active" },
        ],
      },
    ],
  },
];

export const initialEnrollments = [
  { id: "enr-001", member: "Member 1001", email: "", programId: "program-gold", status: "active", points: 18450, earnedPoints: 1040, joinedAt: "2026-07-01", addOns: ["pkg-vip-grocery"], lastChangeReason: "High value grocery household" },
  { id: "enr-002", member: "Member 1002", email: "", programId: "program-gold", status: "active", points: 9120, earnedPoints: 342, joinedAt: "2026-07-05", addOns: [], lastChangeReason: "Initial enrollment" },
  { id: "enr-003", member: "Member 1003", email: "", programId: "program-everyday", status: "pending_review", points: 0, earnedPoints: 0, joinedAt: "2026-07-18", addOns: [], lastChangeReason: "Imported from partner CRM" },
  { id: "enr-004", member: "Member 1004", email: "", programId: "program-gold", status: "active", points: 2230, earnedPoints: 210, joinedAt: "2026-07-20", addOns: ["pkg-service-recovery"], lastChangeReason: "Support recovery" },
  { id: "enr-005", member: "Member 1005", email: "", programId: "program-everyday", status: "active", points: 640, earnedPoints: 150, joinedAt: "2026-07-22", addOns: ["pkg-new-member"], lastChangeReason: "New member rule" },
];

export const initialTransactions = [
  { id: "txn-1001", member: "Member 1001", programId: "program-gold", type: "Earn", category: "grocery", amount: 100.0, points: 500, status: "posted", occurredAt: "2026-07-24 10:24 AM", ruleSource: "Gold base + VIP grocery accelerator" },
  { id: "txn-1002", member: "Member 1002", programId: "program-gold", type: "Earn", category: "pharmacy", amount: 42.0, points: 42, status: "posted", occurredAt: "2026-07-24 10:18 AM" },
  { id: "txn-1003", member: "Member 1001", programId: "program-gold", type: "Burn", category: "coupon", amount: 0, points: -500, status: "posted", occurredAt: "2026-07-24 09:41 AM" },
  { id: "txn-1004", member: "Member 1005", programId: "program-everyday", type: "Earn", category: "apparel", amount: 75.0, points: 150, status: "processing", occurredAt: "2026-07-23 04:03 PM" },
  { id: "txn-1005", member: "Member 1004", programId: "program-gold", type: "Adjust", category: "support", amount: 0, points: 250, status: "posted", occurredAt: "2026-07-23 01:32 PM", ruleSource: "Service recovery boost" },
];
