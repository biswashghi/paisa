# Loyalty Platform Rule Engine Design

## Problem

Partner loyalty rules will not stay as simple as "if purchase, award X points." A realistic platform needs to support:

- A base earn rule plus one or more bonuses.
- Independent bonuses that can stack.
- Base-vs-bonus logic where the member earns the max of either result.
- Bonus caps where the capped bonus may lose to base earn.
- Dependent rules where a second lower earn rate is triggered only after a first rule reaches a limit.
- Rules that are narrow in eligibility but still feed later overflow behavior.
- Purchase, refund, and adjustment flows without directly mutating balances.

The rule engine should model calculation as a constrained graph of rule groups and candidates, not as arbitrary partner-provided code.

## Design Principle

A rule should not directly mutate points. A rule produces an earn candidate.

The engine:

1. Loads the member, member account, active program, and published rule-set version.
2. Normalizes the transaction into calculation facts.
3. Evaluates rule groups by priority.
4. Evaluates eligible rules within each group.
5. Produces candidate awards.
6. Applies limits and dependencies.
7. Resolves selected candidates through the group's strategy.
8. Persists the calculation trace.
9. Posts selected point movement through the ledger.

```text
Transaction event
  -> calculation facts
  -> eligible rule groups
  -> eligible rules
  -> candidate awards
  -> limits/dependencies
  -> selected awards
  -> reward calculation
  -> ledger entries
```

## Core Concepts

### Rule Set Version

A published immutable configuration snapshot for a program.

Responsibilities:

- Groups related rule groups under one version.
- Defines the active earning configuration used for transaction calculation.
- Gives every calculation a stable reference point.

Invariants:

- Draft rule versions can be edited.
- Published rule versions are immutable.
- A calculation always references the published version used.

### Rule Group

A group defines how related rules interact.

Examples:

- Base earn vs promo bonus.
- Category bonus plus overflow earn.
- First-purchase bonus plus other lifecycle bonuses.
- Campaign bonus rules that stack independently.

Key fields:

- `rule_version_id`
- `name`
- `priority`
- `resolution_strategy`
- optional `description`

Resolution strategies:

- `stack`: award every selected candidate.
- `max_of`: award only the highest candidate.
- `base_plus_bonus`: award a required base candidate plus selected bonuses.
- `waterfall`: apply ordered rules against an eligible basis, passing remainder to dependent rules.
- `first_match`: award the first eligible rule by priority.

### Rule

A rule defines eligibility and formula.

Rules should be explicit, typed, and constrained. Do not let partners upload executable expressions in v1.

Key fields:

- `rule_group_id`
- `name`
- `rule_type`
- `priority`
- `status`
- `eligibility_config`
- `formula_config`

V1 rule types:

- `points_per_dollar`
- `fixed_per_transaction`
- `first_purchase_bonus`
- `spend_window_bonus`

Future rule types:

- SKU bonus.
- Category multiplier.
- Merchant/location multiplier.
- Day/time campaign.
- Tier-transition bonus.
- Streak bonus.

### Eligibility

Eligibility decides whether the rule has a basis to calculate against.

Examples:

```json
{
  "transaction_types": ["purchase"],
  "categories": ["grocery"],
  "basis": "line_item_eligible"
}
```

```json
{
  "transaction_types": ["purchase"],
  "channels": ["online"],
  "basis": "eligible"
}
```

```json
{
  "transaction_types": ["purchase"],
  "member_predicate": "first_purchase"
}
```

Supported basis values:

- `subtotal`
- `tax`
- `total`
- `eligible`
- `line_item_subtotal`
- `line_item_total`
- `line_item_eligible`

### Formula

Formula turns an eligible basis into candidate points.

Examples:

```json
{
  "type": "points_per_dollar",
  "points_per_dollar": 5,
  "rounding": "floor"
}
```

```json
{
  "type": "fixed_points",
  "points": 500
}
```

Formula invariants:

- Points are integers.
- Money is calculated from minor units.
- Rounding is explicit.
- Formula output is a candidate only, not a ledger entry.

### Limit

A limit caps candidate output or eligible basis.

Limits must be first-class records so usage can be tracked independently from rule definitions.

Limit dimensions:

- Scope: member, account, partner, program.
- Period: transaction, day, month, year, lifetime, rolling window.
- Metric: points, basis amount, redemption count.

Examples:

```json
{
  "scope": "member",
  "period": "calendar_month",
  "max_points": 1000
}
```

```json
{
  "scope": "member",
  "period": "calendar_month",
  "max_basis_amount_minor": 50000
}
```

Limit invariants:

- Limit usage is keyed by `rule_limit_id`, scope identity, and period.
- Limit usage is committed only for selected candidates.
- In a `max_of` group, losing candidates do not consume cap.
- Limit usage updates must happen in the same transaction as reward calculation and ledger posting.

### Dependency

A dependency controls whether one rule can use the result or remainder of another rule.

Dependency types:

- `requires_match`: rule B can apply only if rule A matched eligibility.
- `requires_award`: rule B can apply only if rule A awarded points.
- `requires_exhausted`: rule B can apply only if rule A hit its limit.
- `applies_to_remainder`: rule B applies only to basis not consumed by rule A.
- `blocked_if_awarded`: rule B is skipped if rule A awarded points.

Dependency invariants:

- Dependencies only reference rules in the same rule-set version.
- Cycles are invalid.
- Dependencies should usually stay within the same rule group.
- Cross-group dependencies should be avoided in v1 unless there is a strong product need.

### Candidate Award

Candidate awards are intermediate calculation outputs.

Candidate fields:

- `rule_id`
- `rule_group_id`
- `eligible`
- `basis_amount_minor`
- `raw_points`
- `limited_points`
- `limit_usage_delta`
- `selected`
- `selection_reason`
- `rejection_reason`

Candidates are useful because support and operations need to understand:

- Which rules matched.
- Which rules did not match.
- Which candidate lost to another candidate.
- Which caps reduced points.
- Which dependency blocked a rule.

### Calculation Trace

The `reward_calculations.calculation_data` JSON should store a compact structured trace.

Example:

```json
{
  "rule_version_id": "uuid",
  "groups": [
    {
      "rule_group_id": "uuid",
      "strategy": "max_of",
      "candidates": [
        {
          "rule_id": "base",
          "basis_amount_minor": 10000,
          "raw_points": 100,
          "limited_points": 100,
          "selected": false,
          "rejection_reason": "lower_than_selected_candidate"
        },
        {
          "rule_id": "online_bonus",
          "basis_amount_minor": 10000,
          "raw_points": 500,
          "limited_points": 300,
          "selected": true,
          "selection_reason": "highest_candidate_after_limits"
        }
      ]
    }
  ],
  "total_points": 300
}
```

## Resolution Strategies

### `stack`

Awards all eligible candidates.

Good for:

- First purchase bonus.
- Spend-window bonus.
- Partner campaigns intended to add on top of base earn.

Important behavior:

- Every selected candidate consumes its own limit usage.
- The group can produce multiple selected candidates.

### `max_of`

Awards only the highest candidate after limits.

Good for:

- "Earn the better of base earn or promo bonus."
- "Earn max of category earn or global campaign earn."

Important behavior:

- Limits are applied before comparison.
- Only the selected candidate consumes limit usage.
- Losing capped candidates do not consume cap.

Example:

```text
Base rule: 1 point per dollar on all purchases.
Bonus rule: 5 points per dollar online, capped at 300 points/month.

$100 online purchase:
- Base candidate: 100 points.
- Bonus candidate: 500 raw, capped to 300.
- Selected: bonus, 300 points.

Second $100 online purchase with only 50 bonus cap remaining:
- Base candidate: 100 points.
- Bonus candidate: 500 raw, capped to 50.
- Selected: base, 100 points.
- Bonus cap is not consumed because bonus lost.
```

### `base_plus_bonus`

Awards a designated base candidate plus selected bonus candidates.

Good for:

- Partners who always want base earn, plus additional bonuses.

This can be implemented as a specialization of `stack`, but naming it explicitly makes partner configuration easier to reason about.

### `waterfall`

Applies ordered rules to an eligible basis. Earlier rules can consume part of the basis. Later dependent rules can receive the remainder.

Good for:

- "5 points per dollar on grocery for the first $500/month, then 1 point per dollar on remaining grocery spend."
- "High promotional earn until a cap, then lower fallback earn."

Important behavior:

- Rules run by priority.
- A rule can consume eligible basis.
- A dependent rule can require the prior rule to be exhausted.
- Remainder should preserve the original eligibility scope.

Example:

```text
Rule A: Grocery premium
- Applies only to grocery line items.
- 5 points per dollar.
- Cap: first $500 grocery eligible spend/month.

Rule B: Grocery overflow
- Depends on Rule A.
- Dependency: applies_to_remainder + requires_exhausted.
- 1 point per dollar.

$700 grocery purchase:
- Rule A consumes $500 and awards 2,500 points.
- Rule B receives $200 remainder and awards 200 points.
- Total: 2,700 points.
```

### `first_match`

Awards the first eligible candidate by priority.

Good for:

- Mutually exclusive campaign tiers.
- Simple partner configuration where priority should decide.

## Proposed Table Extensions

These tables extend the broader loyalty data model.

### `rule_groups`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | uuid pk | Rule group ID. |
| `partner_id` | uuid fk | Tenant scope. |
| `rule_version_id` | uuid fk | FK to `program_rule_versions.id`. |
| `name` | text | Partner/admin label. |
| `priority` | integer | Group evaluation order. |
| `resolution_strategy` | text | `stack`, `max_of`, `base_plus_bonus`, `waterfall`, `first_match`. |
| `status` | text | `active`, `disabled`. |
| `created_at` | timestamptz | Required. |
| `updated_at` | timestamptz | Required. |

Constraints:

- Index on `(partner_id, rule_version_id, status, priority)`.

### `earning_rules`

Replace the earlier direct `rule_version_id` relationship with `rule_group_id`.

| Column | Type | Notes |
| --- | --- | --- |
| `id` | uuid pk | Rule ID. |
| `partner_id` | uuid fk | Tenant scope. |
| `rule_group_id` | uuid fk | FK to `rule_groups.id`. |
| `name` | text | Partner/admin label. |
| `rule_type` | text | Constrained type. |
| `priority` | integer | Rule evaluation order within group. |
| `status` | text | `active`, `disabled`. |
| `eligibility_config` | jsonb | Eligibility predicates and basis selection. |
| `formula_config` | jsonb | Formula-specific config. |
| `created_at` | timestamptz | Required. |
| `updated_at` | timestamptz | Required. |

Constraints:

- Index on `(partner_id, rule_group_id, status, priority)`.

### `rule_limits`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | uuid pk | Limit ID. |
| `partner_id` | uuid fk | Tenant scope. |
| `rule_id` | uuid fk | FK to `earning_rules.id`. |
| `scope` | text | `member`, `account`, `program`, `partner`. |
| `period` | text | `transaction`, `day`, `calendar_month`, `calendar_year`, `lifetime`, `rolling_window`. |
| `period_config` | jsonb | Window details if needed. |
| `max_points` | integer nullable | Cap by points. |
| `max_basis_amount_minor` | integer nullable | Cap by eligible amount. |
| `status` | text | `active`, `disabled`. |
| `created_at` | timestamptz | Required. |
| `updated_at` | timestamptz | Required. |

Constraints:

- Index on `(partner_id, rule_id, status)`.

### `rule_dependencies`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | uuid pk | Dependency ID. |
| `partner_id` | uuid fk | Tenant scope. |
| `rule_id` | uuid fk | Dependent rule. |
| `depends_on_rule_id` | uuid fk | Prerequisite rule. |
| `dependency_type` | text | `requires_match`, `requires_award`, `requires_exhausted`, `applies_to_remainder`, `blocked_if_awarded`. |
| `created_at` | timestamptz | Required. |

Constraints:

- Unique index on `(rule_id, depends_on_rule_id, dependency_type)`.
- Validate no dependency cycles before publishing a rule version.

### `rule_limit_usage`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | uuid pk | Usage row ID. |
| `partner_id` | uuid fk | Tenant scope. |
| `rule_limit_id` | uuid fk | FK to `rule_limits.id`. |
| `scope_type` | text | Mirrors limit scope. |
| `scope_id` | uuid | Member/account/program/partner ID. |
| `period_start` | timestamptz nullable | Null for lifetime. |
| `period_end` | timestamptz nullable | Null for lifetime. |
| `used_points` | integer | Points consumed against limit. |
| `used_basis_amount_minor` | integer | Basis consumed against limit. |
| `updated_at` | timestamptz | Required. |

Constraints:

- Unique index on `(rule_limit_id, scope_type, scope_id, period_start, period_end)`.
- Usage rows should be updated transactionally with selected candidate persistence and ledger posting.

## Rule Publishing Validation

Before publishing a rule version:

- Every group has a valid resolution strategy.
- Every active rule belongs to an active group.
- Every active rule has valid eligibility and formula config for its `rule_type`.
- Dependency graph has no cycles.
- `waterfall` dependencies reference earlier-priority rules in the same group.
- Limit periods and scopes are valid.
- A `max_of` group has at least two active rules.
- A `waterfall` group has at least one remainder/dependency path if any rule has a basis cap.
- Published version would not orphan members with no eligible active rule version.

## Evaluation Algorithm

Pseudocode:

```text
calculate(transaction_event):
  member = load_member(transaction_event.member_id)
  enrollment = load_active_program_enrollment(member)
  rule_version = load_published_rule_version(enrollment.program_id)
  facts = normalize_transaction(transaction_event)

  calculation = start_reward_calculation(transaction_event, rule_version)

  for group in rule_version.groups ordered by priority:
    candidates = evaluate_group(group, facts, member)
    selected = resolve_candidates(group.strategy, candidates)
    record group trace
    append selected awards

  in one db transaction:
    persist calculation trace
    update selected rule_limit_usage rows
    insert ledger_entries for selected awards
    update balance_snapshots
    mark transaction_event processed
```

For `waterfall`:

```text
remaining_basis = eligible basis by scope/category/line

for rule in group.rules ordered by priority:
  if dependency is not satisfied:
    record skipped candidate
    continue

  basis = select basis from remaining_basis
  candidate = formula(rule, basis)
  candidate = apply_limit(candidate)

  if candidate selected:
    consume candidate.basis_amount_minor from remaining_basis
    record whether rule exhausted its limit
```

## Refund Handling

Refund behavior should not blindly re-run the current rule version, because rules may have changed since the purchase.

Recommended v1 behavior:

- Refund events should reference the original transaction when available.
- If original transaction is known, create compensating ledger entries from the original selected calculation candidates.
- If partial refund is line-item specific, reverse proportional selected awards for the refunded line/basis when trace data supports it.
- If original transaction is unknown, calculate a negative refund using the member's current active program only as a fallback and mark the calculation trace as `original_transaction_missing`.

Refund invariants:

- Refunds can make available balance negative.
- Refunds write compensating ledger entries.
- Refunds should not consume positive earn caps.
- Refunds may release prior limit usage only if the partner explicitly wants caps restored; default v1 should not restore cap usage automatically.

## Example Configurations

### Independent Lifecycle Bonuses

```json
{
  "group": "Lifecycle bonuses",
  "resolution_strategy": "stack",
  "rules": [
    {
      "name": "First purchase bonus",
      "rule_type": "first_purchase_bonus",
      "eligibility_config": {
        "transaction_types": ["purchase"],
        "member_predicate": "first_purchase"
      },
      "formula_config": {
        "type": "fixed_points",
        "points": 500
      }
    },
    {
      "name": "Spend $500 in 30 days",
      "rule_type": "spend_window_bonus",
      "eligibility_config": {
        "transaction_types": ["purchase"],
        "window_days": 30,
        "threshold_minor": 50000
      },
      "formula_config": {
        "type": "fixed_points",
        "points": 1000
      }
    }
  ]
}
```

### Base Earn Or Capped Bonus

```json
{
  "group": "Everyday earn",
  "resolution_strategy": "max_of",
  "rules": [
    {
      "name": "Base earn",
      "rule_type": "points_per_dollar",
      "eligibility_config": {
        "transaction_types": ["purchase"],
        "basis": "eligible"
      },
      "formula_config": {
        "type": "points_per_dollar",
        "points_per_dollar": 1,
        "rounding": "floor"
      }
    },
    {
      "name": "Online promo",
      "rule_type": "points_per_dollar",
      "eligibility_config": {
        "transaction_types": ["purchase"],
        "channels": ["online"],
        "basis": "eligible"
      },
      "formula_config": {
        "type": "points_per_dollar",
        "points_per_dollar": 5,
        "rounding": "floor"
      },
      "limits": [
        {
          "scope": "member",
          "period": "calendar_month",
          "max_points": 300
        }
      ]
    }
  ]
}
```

### Dependent Waterfall

```json
{
  "group": "Grocery earn",
  "resolution_strategy": "waterfall",
  "rules": [
    {
      "name": "Premium grocery earn",
      "rule_type": "points_per_dollar",
      "eligibility_config": {
        "transaction_types": ["purchase"],
        "categories": ["grocery"],
        "basis": "line_item_eligible"
      },
      "formula_config": {
        "type": "points_per_dollar",
        "points_per_dollar": 5
      },
      "limits": [
        {
          "scope": "member",
          "period": "calendar_month",
          "max_basis_amount_minor": 50000
        }
      ]
    },
    {
      "name": "Grocery overflow earn",
      "rule_type": "points_per_dollar",
      "eligibility_config": {
        "transaction_types": ["purchase"],
        "categories": ["grocery"],
        "basis": "line_item_eligible"
      },
      "formula_config": {
        "type": "points_per_dollar",
        "points_per_dollar": 1
      },
      "dependencies": [
        {
          "depends_on": "Premium grocery earn",
          "type": "requires_exhausted"
        },
        {
          "depends_on": "Premium grocery earn",
          "type": "applies_to_remainder"
        }
      ]
    }
  ]
}
```

## Backend Validation Coverage

Backend domain/application tests should validate:

- `stack` groups for independent lifecycle bonuses.
- `max_of` groups for base-vs-bonus selection.
- Bonus caps applied before max selection.
- Losing candidates do not consume cap.
- `waterfall` groups for dependent overflow earn.
- Remainder basis feeds correctly from the first rule into the dependent rule.
- Publish validation rejects dependency cycles.
- Publish validation rejects waterfall dependencies that point forward to later-priority rules.
- Publish validation rejects `max_of` groups with fewer than two active rules.
- Published rule versions reject further mutation.
- Draft rule versions cannot calculate live transactions.
