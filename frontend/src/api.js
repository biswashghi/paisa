const API_BASE = import.meta.env.VITE_PAISA_API_URL || import.meta.env.VITE_API_BASE || "http://localhost:8080";
let authToken = localStorage.getItem("paisa.partnerPortal.token") || "";

export function setAuthToken(token) {
  authToken = token || "";
  if (authToken) {
    localStorage.setItem("paisa.partnerPortal.token", authToken);
  } else {
    localStorage.removeItem("paisa.partnerPortal.token");
  }
}

async function request(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(authToken ? { Authorization: `Bearer ${authToken}` } : {}),
      ...(options.headers || {}),
    },
  });
  const text = await response.text();
  const payload = text ? JSON.parse(text) : null;
  if (!response.ok) {
    throw new Error(payload?.error || `Request failed with ${response.status}`);
  }
  return payload;
}

export const api = {
  baseUrl: API_BASE,
  setAuthToken,

  health: () => request("/health"),

  login: (body) => request("/partner/v1/auth/login", { method: "POST", body: JSON.stringify(body) }),
  logout: () => request("/partner/v1/auth/logout", { method: "POST" }),
  me: () => request("/partner/v1/me"),

  createProgram: (body) => request("/partner/v1/programs", { method: "POST", body: JSON.stringify(body) }),
  listPrograms: () => request("/partner/v1/programs"),

  createRuleVersion: (programId, body) => request(`/partner/v1/programs/${programId}/rule-versions`, { method: "POST", body: JSON.stringify(body) }),
  publishRuleVersion: (programId, versionId) => request(`/partner/v1/programs/${programId}/rule-versions/${versionId}/publish`, { method: "POST" }),
  listRuleVersions: (programId) => request(`/partner/v1/programs/${programId}/rule-versions`),
  getRuleVersionReview: (programId, versionId) => request(`/partner/v1/programs/${programId}/rule-versions/${versionId}`),
  listRulePackages: (programId) => request(`/partner/v1/programs/${programId}/rule-packages`),
  createRulePackage: (programId, body) => request(`/partner/v1/programs/${programId}/rule-packages`, { method: "POST", body: JSON.stringify(body) }),

  createMember: (body) => request("/partner/v1/members", { method: "POST", body: JSON.stringify(body) }),
  listMembers: () => request("/partner/v1/members"),
  getRewardsProfile: (memberId) => request(`/partner/v1/members/${memberId}/rewards-profile`),
  updateEnrollment: (memberId, body) => request(`/partner/v1/members/${memberId}/program-enrollment`, { method: "PUT", body: JSON.stringify(body) }),
  createRuleAssignment: (memberId, body) => request(`/partner/v1/members/${memberId}/rule-assignments`, { method: "POST", body: JSON.stringify(body) }),
  updateRuleAssignment: (memberId, assignmentId, body) => request(`/partner/v1/members/${memberId}/rule-assignments/${assignmentId}`, { method: "PATCH", body: JSON.stringify(body) }),

  ingestTransaction: (body) => request("/partner/v1/ingest/transactions", { method: "POST", body: JSON.stringify(body) }),
  processTransactions: () => request("/partner/v1/jobs/process-transaction-events", { method: "POST" }),
  listTransactions: () => request("/partner/v1/transactions"),
  getCalculation: (transactionId) => request(`/partner/v1/transactions/${transactionId}/calculation`),

  createApiKey: (body) => request("/partner/v1/api-keys", { method: "POST", body: JSON.stringify(body) }),
  listApiKeys: () => request("/partner/v1/api-keys"),
  revokeApiKey: (keyId) => request(`/partner/v1/api-keys/${keyId}`, { method: "DELETE" }),

  createLocation: (body) => request("/partner/v1/locations", { method: "POST", body: JSON.stringify(body) }),
  listLocations: () => request("/partner/v1/locations"),
  dashboardSummary: () => request("/partner/v1/dashboard"),

  createCatalogItem: (body) => request("/partner/v1/catalog-items", { method: "POST", body: JSON.stringify(body) }),
  listCatalogItems: () => request("/partner/v1/catalog-items"),
  updateCatalogItem: (itemId, body) => request(`/partner/v1/catalog-items/${itemId}`, { method: "PATCH", body: JSON.stringify(body) }),

  resolveMember: (body) => request("/pos/v1/members/resolve", { method: "POST", body: JSON.stringify(body) }),
  createManualTransaction: (body) => request("/pos/v1/manual-transactions", { method: "POST", body: JSON.stringify(body) }),
  getPOSBalance: (memberId) => request(`/pos/v1/members/${memberId}/balance`),
  availableRewards: (memberId) => request(`/pos/v1/members/${memberId}/available-rewards`),
  createRedemption: (body) => request("/pos/v1/redemptions", { method: "POST", body: JSON.stringify(body) }),
  validateRedemption: (redemptionId) => request(`/pos/v1/redemptions/${redemptionId}/validate`, { method: "POST" }),
  captureRedemption: (redemptionId) => request(`/pos/v1/redemptions/${redemptionId}/capture`, { method: "POST" }),
  releaseRedemption: (redemptionId) => request(`/pos/v1/redemptions/${redemptionId}/release`, { method: "POST" }),
  listRedemptions: () => request("/partner/v1/redemptions"),

  listIntegrationConnections: () => request("/partner/v1/integration-connections"),
  startSquareOAuth: () => request("/partner/v1/integration-connections/square/oauth-start", { method: "POST" }),
  completeSquareOAuth: (code) => request(`/partner/v1/integration-connections/square/oauth-callback?code=${encodeURIComponent(code)}`),
  syncIntegrationConnection: (connectionId) => request(`/partner/v1/integration-connections/${connectionId}/sync`, { method: "POST" }),

  createCampaign: (body) => request("/partner/v1/campaigns", { method: "POST", body: JSON.stringify(body) }),
  listCampaigns: () => request("/partner/v1/campaigns"),
};
