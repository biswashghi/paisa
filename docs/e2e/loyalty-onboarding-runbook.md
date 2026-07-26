# Loyalty Onboarding E2E Runbook

Last run: 2026-07-26

## Goal

Exercise the onboarding flow as a store that has no rewards program yet, then
prove the backend can support a larger partner/member/transaction/redemption
suite.

## Local Stack

The run used:

- Postgres from `postgres-docker/docker-compose.yml`
- API at `http://127.0.0.1:8081`
- Frontend at `http://127.0.0.1:5173`
- Frontend env var: `VITE_PAISA_API_URL=http://127.0.0.1:8081`

## Browser Walkthrough

1. Opened the partner admin UI.
2. Signed out of the existing default workspace.
3. Logged in with a new partner key:
   `porcelain-store-1785103998258`.
4. Confirmed the empty partner state showed no programs and offered a create
   action.
5. Created the first program from the UI.
6. Opened Rule Studio.
7. Used the default `max_of` template:
   - Base earn: `1 pt / $`
   - Grocery bonus: `5 pt / $`
8. Published the rule version through the backend validator.
9. Opened Rewards and created a `Free coffee` catalog item for `100` points.
10. Opened Cashier mode.
11. Resolved a new phone number into a member enrolled in the active program.
12. Submitted a `100.00` grocery purchase.
13. Confirmed the member balance became `500 available / 0 reserved`.
14. Reserved the `100` point reward.
15. Confirmed the member balance became `400 available / 100 reserved`.
16. Validated and captured the redemption.
17. Confirmed the final member balance became `400 available / 0 reserved`.

Screenshots:

- `frontend/.visual-review/e2e-cashier-captured-porcelain-store-1785103998258.png`
- `frontend/.visual-review/e2e-cashier-reserve-disabled-porcelain-store-1785103998258.png`

## Refinement Made

The browser walkthrough exposed several partner-workflow issues:

- The dashboard mixed first-run setup with operating metrics. A new partner now
  sees a focused launch path until the minimum setup is complete.
- The empty dashboard previously showed placeholder rule-health numbers. It now
  hides the normal cockpit until at least one program exists, and rule health is
  derived from loaded rule data.
- The Onboarding view marked cashier setup as ready by default. It now shows
  real readiness for location, program, published rules, reward catalog, and
  cashier test.
- The cashier UI originally enabled `Reserve` for a zero-balance member. The
  backend ledger guard rejected overspending, but the UI created a bad cashier
  affordance. The Cashier view now disables `Reserve` when the selected reward
  cost is greater than available points and shows the missing point amount.
- Cashier redemption actions are now status-aware: validate/capture/release only
  enable when the current redemption state supports that action.
- The frontend now accepts both `VITE_PAISA_API_URL` and `VITE_API_BASE` for the
  backend URL to avoid local setup confusion.

Additional screenshots:

- `frontend/.visual-review/workflow-launch-path-readable-workflow-store-1785106587912.png`
- `frontend/.visual-review/workflow-onboarding-checklist-workflow-store-1785106587912.png`

## Full Load Rehearsal

Command:

```bash
PAISA_API_BASE=http://127.0.0.1:8081 \
PAISA_E2E_RUN_ID=full-20260726-1840 \
node scripts/e2e-loyalty-load.mjs
```

Scenario:

- `3` partners
- `2` programs per partner
- `6` published rule versions
- `6` active catalog rewards
- `1000` enrolled members
- `10000` purchase transactions
- `1000` reserve/validate/capture redemptions

Result:

- Partners created: `3`
- Programs created: `6`
- Rule versions published: `6`
- Catalog items created: `6`
- Members created: `1000`
- Transactions ingested: `10000`
- Transactions processed: `10000`
- Processing failures: `0`
- Redemptions reserved: `1000`
- Redemptions validated: `1000`
- Redemptions captured: `1000`
- Verified transaction count: `10000`
- Verified redemption count: `1000`
- Verified captured redemptions: `1000`
- Sampled reserved balances after capture: all `0`

Timings:

- Setup: `1.894s`
- Ingestion: `16.339s`
- Reward processing: `79.695s`
- Redemptions: `4.204s`
- Verification: `21.308s`
- Total: `123.465s`

Report:

- `docs/e2e/loyalty-load-full-20260726-1840.json`

## Observations

- The domain flow works end to end: partner setup, program creation, published
  rules, member enrollment, earning, ledger balance updates, redemption reserve,
  validation, capture, and final balance verification.
- Reward processing is stable but intentionally batch-bound at `50` events per
  job call, so `10000` transactions required `200` productive batches plus one
  empty final check.
- Full verification currently lists all transactions and redemptions by partner;
  this is acceptable for a local rehearsal but should become paginated before
  large partner admin screens depend on it.
