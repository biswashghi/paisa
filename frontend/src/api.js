const API_BASE = import.meta.env.VITE_PAISA_API_URL || "http://localhost:8080";

async function request(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
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

  health: () => request("/health"),

  createPartner: (body) => request("/v1/partners", { method: "POST", body: JSON.stringify(body) }),
  getPartner: (partnerKey) => request(`/v1/partners/${partnerKey}`),

  createProgram: (partnerKey, body) => request(`/v1/partners/${partnerKey}/programs`, { method: "POST", body: JSON.stringify(body) }),
  listPrograms: (partnerKey) => request(`/v1/partners/${partnerKey}/programs`),

  createRuleVersion: (partnerKey, programId, body) => request(`/v1/partners/${partnerKey}/programs/${programId}/rule-versions`, { method: "POST", body: JSON.stringify(body) }),
  publishRuleVersion: (partnerKey, programId, versionId) => request(`/v1/partners/${partnerKey}/programs/${programId}/rule-versions/${versionId}/publish`, { method: "POST" }),
  listRuleVersions: (partnerKey, programId) => request(`/v1/partners/${partnerKey}/programs/${programId}/rule-versions`),
  getRuleVersionReview: (partnerKey, programId, versionId) => request(`/v1/partners/${partnerKey}/programs/${programId}/rule-versions/${versionId}`),
  listRulePackages: (partnerKey, programId) => request(`/v1/partners/${partnerKey}/programs/${programId}/rule-packages`),
  createRulePackage: (partnerKey, programId, body) => request(`/v1/partners/${partnerKey}/programs/${programId}/rule-packages`, { method: "POST", body: JSON.stringify(body) }),

  createMember: (partnerKey, body) => request(`/v1/partners/${partnerKey}/members`, { method: "POST", body: JSON.stringify(body) }),
  listMembers: (partnerKey) => request(`/v1/partners/${partnerKey}/members`),
  getRewardsProfile: (partnerKey, memberId) => request(`/v1/partners/${partnerKey}/members/${memberId}/rewards-profile`),
  updateEnrollment: (partnerKey, memberId, body) => request(`/v1/partners/${partnerKey}/members/${memberId}/program-enrollment`, { method: "PUT", body: JSON.stringify(body) }),
  createRuleAssignment: (partnerKey, memberId, body) => request(`/v1/partners/${partnerKey}/members/${memberId}/rule-assignments`, { method: "POST", body: JSON.stringify(body) }),
  updateRuleAssignment: (partnerKey, memberId, assignmentId, body) => request(`/v1/partners/${partnerKey}/members/${memberId}/rule-assignments/${assignmentId}`, { method: "PATCH", body: JSON.stringify(body) }),

  ingestTransaction: (partnerKey, body) => request(`/v1/partners/${partnerKey}/ingest/transactions`, { method: "POST", body: JSON.stringify(body) }),
  processTransactions: () => request("/v1/jobs/process-transaction-events", { method: "POST" }),
  listTransactions: (partnerKey) => request(`/v1/partners/${partnerKey}/transactions`),
  getCalculation: (partnerKey, transactionId) => request(`/v1/partners/${partnerKey}/transactions/${transactionId}/calculation`),
};
