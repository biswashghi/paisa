CREATE SCHEMA IF NOT EXISTS paisa;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS paisa.partners (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_key TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS paisa.partner_users (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    email TEXT NOT NULL,
    name TEXT NOT NULL,
    password_hash TEXT,
    role TEXT NOT NULL DEFAULT 'admin',
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (partner_id, email)
);

ALTER TABLE paisa.partner_users
ADD COLUMN IF NOT EXISTS password_hash TEXT;

CREATE TABLE IF NOT EXISTS paisa.sessions (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    partner_user_id UUID NOT NULL REFERENCES paisa.partner_users(id),
    token_hash TEXT UNIQUE NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS paisa.api_keys (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    name TEXT NOT NULL,
    key_prefix TEXT NOT NULL,
    secret_hash TEXT UNIQUE NOT NULL,
    scopes TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    status TEXT NOT NULL DEFAULT 'active',
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS paisa.partner_locations (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    name TEXT NOT NULL,
    address TEXT,
    timezone TEXT NOT NULL DEFAULT 'America/Detroit',
    status TEXT NOT NULL DEFAULT 'active',
    external_location_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS paisa.integration_connections (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    provider TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    external_merchant_id TEXT,
    external_location_id TEXT,
    metadata JSONB NOT NULL DEFAULT '{}',
    last_sync_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS paisa.programs (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    name TEXT NOT NULL,
    tier_code TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    priority INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (partner_id, tier_code)
);

CREATE TABLE IF NOT EXISTS paisa.program_rule_versions (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    program_id UUID NOT NULL REFERENCES paisa.programs(id),
    version INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    scope TEXT NOT NULL DEFAULT 'program_base',
    rule_set_key TEXT,
    name TEXT,
    description TEXT,
    earn_basis TEXT NOT NULL DEFAULT 'eligible',
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (program_id, scope, version)
);

ALTER TABLE paisa.program_rule_versions
DROP CONSTRAINT IF EXISTS program_rule_versions_program_id_version_key;

CREATE UNIQUE INDEX IF NOT EXISTS program_rule_versions_program_scope_version_key
ON paisa.program_rule_versions(program_id, scope, version);

ALTER TABLE paisa.program_rule_versions
ADD COLUMN IF NOT EXISTS scope TEXT NOT NULL DEFAULT 'program_base';

ALTER TABLE paisa.program_rule_versions
ADD COLUMN IF NOT EXISTS rule_set_key TEXT;

ALTER TABLE paisa.program_rule_versions
ADD COLUMN IF NOT EXISTS name TEXT;

ALTER TABLE paisa.program_rule_versions
ADD COLUMN IF NOT EXISTS description TEXT;

CREATE TABLE IF NOT EXISTS paisa.rule_groups (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    rule_version_id UUID NOT NULL REFERENCES paisa.program_rule_versions(id),
    name TEXT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 0,
    resolution_strategy TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS paisa.earning_rules (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    rule_group_id UUID NOT NULL REFERENCES paisa.rule_groups(id),
    rule_key TEXT NOT NULL,
    name TEXT NOT NULL,
    rule_type TEXT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'active',
    eligibility_config JSONB NOT NULL DEFAULT '{}',
    formula_config JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (rule_group_id, rule_key)
);

CREATE TABLE IF NOT EXISTS paisa.rule_limits (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    rule_id UUID NOT NULL REFERENCES paisa.earning_rules(id),
    scope TEXT NOT NULL DEFAULT 'member',
    period TEXT NOT NULL DEFAULT 'lifetime',
    max_points INTEGER,
    max_basis_amount_minor INTEGER,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS paisa.rule_dependencies (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    rule_id UUID NOT NULL REFERENCES paisa.earning_rules(id),
    depends_on_rule_id UUID NOT NULL REFERENCES paisa.earning_rules(id),
    dependency_type TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (rule_id, depends_on_rule_id, dependency_type)
);

CREATE TABLE IF NOT EXISTS paisa.members (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    external_customer_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (partner_id, external_customer_id)
);

CREATE TABLE IF NOT EXISTS paisa.member_identifiers (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    member_id UUID NOT NULL REFERENCES paisa.members(id),
    type TEXT NOT NULL,
    value_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (partner_id, type, value_hash)
);

CREATE TABLE IF NOT EXISTS paisa.member_accounts (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    member_id UUID NOT NULL REFERENCES paisa.members(id),
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (member_id)
);

CREATE TABLE IF NOT EXISTS paisa.program_enrollments (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    member_id UUID NOT NULL REFERENCES paisa.members(id),
    program_id UUID NOT NULL REFERENCES paisa.programs(id),
    status TEXT NOT NULL DEFAULT 'active',
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at TIMESTAMPTZ,
    effective_at TIMESTAMPTZ,
    change_reason TEXT,
    created_by_type TEXT NOT NULL DEFAULT 'system',
    created_by_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE paisa.program_enrollments
ADD COLUMN IF NOT EXISTS effective_at TIMESTAMPTZ;

ALTER TABLE paisa.program_enrollments
ADD COLUMN IF NOT EXISTS change_reason TEXT;

ALTER TABLE paisa.program_enrollments
ADD COLUMN IF NOT EXISTS created_by_type TEXT NOT NULL DEFAULT 'system';

ALTER TABLE paisa.program_enrollments
ADD COLUMN IF NOT EXISTS created_by_id UUID;

ALTER TABLE paisa.program_enrollments
ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now();

CREATE UNIQUE INDEX IF NOT EXISTS program_enrollments_one_active_member
ON paisa.program_enrollments(member_id)
WHERE status = 'active';

CREATE TABLE IF NOT EXISTS paisa.member_rule_assignments (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    member_id UUID NOT NULL REFERENCES paisa.members(id),
    program_enrollment_id UUID NOT NULL REFERENCES paisa.program_enrollments(id),
    rule_version_id UUID NOT NULL REFERENCES paisa.program_rule_versions(id),
    status TEXT NOT NULL DEFAULT 'active',
    starts_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ends_at TIMESTAMPTZ,
    reason TEXT,
    created_by_type TEXT NOT NULL DEFAULT 'system',
    created_by_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS member_rule_assignments_one_active_rule_version
ON paisa.member_rule_assignments(member_id, rule_version_id)
WHERE status = 'active';

CREATE INDEX IF NOT EXISTS member_rule_assignments_member_active_idx
ON paisa.member_rule_assignments(partner_id, member_id, status, starts_at);

CREATE TABLE IF NOT EXISTS paisa.transaction_events (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    member_id UUID NOT NULL REFERENCES paisa.members(id),
    external_transaction_id TEXT NOT NULL,
    original_external_transaction_id TEXT,
    type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'accepted',
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    subtotal_minor INTEGER,
    tax_minor INTEGER,
    total_minor INTEGER,
    eligible_minor INTEGER,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    raw_payload JSONB NOT NULL,
    payload_hash TEXT NOT NULL,
    source_system TEXT NOT NULL DEFAULT 'legacy',
    source_connection_id UUID,
    source_location_id TEXT,
    external_event_type TEXT,
    idempotency_key TEXT,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (partner_id, external_transaction_id)
);

ALTER TABLE paisa.transaction_events
ADD COLUMN IF NOT EXISTS source_system TEXT NOT NULL DEFAULT 'legacy';

ALTER TABLE paisa.transaction_events
ADD COLUMN IF NOT EXISTS source_connection_id UUID;

ALTER TABLE paisa.transaction_events
ADD COLUMN IF NOT EXISTS source_location_id TEXT;

ALTER TABLE paisa.transaction_events
ADD COLUMN IF NOT EXISTS external_event_type TEXT;

ALTER TABLE paisa.transaction_events
ADD COLUMN IF NOT EXISTS idempotency_key TEXT;

ALTER TABLE paisa.transaction_events
ADD COLUMN IF NOT EXISTS received_at TIMESTAMPTZ NOT NULL DEFAULT now();

CREATE TABLE IF NOT EXISTS paisa.transaction_line_items (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    transaction_event_id UUID NOT NULL REFERENCES paisa.transaction_events(id),
    external_line_id TEXT,
    sku TEXT,
    category TEXT,
    quantity INTEGER NOT NULL DEFAULT 1,
    subtotal_minor INTEGER,
    tax_minor INTEGER,
    total_minor INTEGER,
    eligible_minor INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS paisa.reward_calculations (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    transaction_event_id UUID NOT NULL REFERENCES paisa.transaction_events(id),
    program_id UUID REFERENCES paisa.programs(id),
    rule_version_id UUID REFERENCES paisa.program_rule_versions(id),
    status TEXT NOT NULL,
    points_delta INTEGER NOT NULL DEFAULT 0,
    basis_amount_minor INTEGER,
    calculation_data JSONB NOT NULL DEFAULT '{}',
    failure_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (transaction_event_id)
);

CREATE TABLE IF NOT EXISTS paisa.catalog_items (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    program_id UUID REFERENCES paisa.programs(id),
    location_id UUID REFERENCES paisa.partner_locations(id),
    name TEXT NOT NULL,
    description TEXT,
    points_cost INTEGER NOT NULL,
    reward_type TEXT NOT NULL DEFAULT 'manual_discount',
    status TEXT NOT NULL DEFAULT 'active',
    expires_after_minutes INTEGER NOT NULL DEFAULT 15,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS paisa.redemption_codes (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    catalog_item_id UUID NOT NULL REFERENCES paisa.catalog_items(id),
    code TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'available',
    redemption_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (partner_id, code)
);

CREATE TABLE IF NOT EXISTS paisa.redemptions (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    member_id UUID NOT NULL REFERENCES paisa.members(id),
    member_account_id UUID NOT NULL REFERENCES paisa.member_accounts(id),
    catalog_item_id UUID NOT NULL REFERENCES paisa.catalog_items(id),
    code TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'requested',
    points_cost INTEGER NOT NULL,
    reservation_expires_at TIMESTAMPTZ,
    failure_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (partner_id, code)
);

CREATE TABLE IF NOT EXISTS paisa.redemption_events (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    redemption_id UUID NOT NULL REFERENCES paisa.redemptions(id),
    status TEXT NOT NULL,
    details JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS paisa.campaigns (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    name TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'draft',
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    required_visit_count INTEGER NOT NULL DEFAULT 3,
    reward_catalog_item_id UUID REFERENCES paisa.catalog_items(id),
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS paisa.rule_limit_usage (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    rule_limit_id UUID NOT NULL REFERENCES paisa.rule_limits(id),
    member_id UUID NOT NULL REFERENCES paisa.members(id),
    period_key TEXT NOT NULL DEFAULT 'lifetime',
    used_points INTEGER NOT NULL DEFAULT 0,
    used_basis_amount_minor INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (rule_limit_id, member_id, period_key)
);

CREATE TABLE IF NOT EXISTS paisa.ledger_entries (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    member_account_id UUID NOT NULL REFERENCES paisa.member_accounts(id),
    program_id UUID REFERENCES paisa.programs(id),
    entry_type TEXT NOT NULL,
    available_delta INTEGER NOT NULL DEFAULT 0,
    reserved_delta INTEGER NOT NULL DEFAULT 0,
    expired_delta INTEGER NOT NULL DEFAULT 0,
    source_type TEXT NOT NULL,
    source_id UUID NOT NULL,
    reason TEXT,
    created_by_type TEXT NOT NULL DEFAULT 'system',
    created_by_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE paisa.ledger_entries
ADD COLUMN IF NOT EXISTS created_by_type TEXT NOT NULL DEFAULT 'system';

ALTER TABLE paisa.ledger_entries
ADD COLUMN IF NOT EXISTS created_by_id UUID;

CREATE TABLE IF NOT EXISTS paisa.balance_snapshots (
    member_account_id UUID PRIMARY KEY REFERENCES paisa.member_accounts(id),
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    available_points INTEGER NOT NULL DEFAULT 0,
    reserved_points INTEGER NOT NULL DEFAULT 0,
    expired_points INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS paisa.ledger_exports (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    partner_id UUID NOT NULL REFERENCES paisa.partners(id),
    business_date DATE NOT NULL,
    status TEXT NOT NULL DEFAULT 'complete',
    file_path TEXT,
    summary JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (partner_id, business_date)
);

CREATE INDEX IF NOT EXISTS transaction_events_status_created_idx
ON paisa.transaction_events(status, created_at);

CREATE INDEX IF NOT EXISTS transaction_events_accepted_claim_idx
ON paisa.transaction_events(status, created_at, id)
WHERE status = 'accepted';

CREATE INDEX IF NOT EXISTS transaction_events_partner_member_occurred_idx
ON paisa.transaction_events(partner_id, member_id, occurred_at);

CREATE INDEX IF NOT EXISTS transaction_events_partner_member_type_status_occurred_idx
ON paisa.transaction_events(partner_id, member_id, type, status, occurred_at);

CREATE INDEX IF NOT EXISTS transaction_events_partner_original_type_idx
ON paisa.transaction_events(partner_id, original_external_transaction_id, type);

CREATE INDEX IF NOT EXISTS program_enrollments_partner_member_active_idx
ON paisa.program_enrollments(partner_id, member_id)
WHERE status = 'active';

CREATE INDEX IF NOT EXISTS program_rule_versions_partner_program_scope_status_published_idx
ON paisa.program_rule_versions(partner_id, program_id, scope, status, published_at DESC);

CREATE INDEX IF NOT EXISTS member_rule_assignments_partner_member_status_window_idx
ON paisa.member_rule_assignments(partner_id, member_id, status, starts_at, ends_at);

CREATE INDEX IF NOT EXISTS rule_groups_version_status_priority_idx
ON paisa.rule_groups(rule_version_id, status, priority);

CREATE INDEX IF NOT EXISTS earning_rules_group_status_priority_idx
ON paisa.earning_rules(rule_group_id, status, priority);

CREATE INDEX IF NOT EXISTS rule_limits_rule_status_idx
ON paisa.rule_limits(rule_id, status);

CREATE INDEX IF NOT EXISTS rule_dependencies_rule_idx
ON paisa.rule_dependencies(rule_id);

CREATE INDEX IF NOT EXISTS reward_calculations_event_status_idx
ON paisa.reward_calculations(transaction_event_id, status);

CREATE INDEX IF NOT EXISTS balance_snapshots_partner_account_idx
ON paisa.balance_snapshots(partner_id, member_account_id);

CREATE INDEX IF NOT EXISTS sessions_token_hash_idx
ON paisa.sessions(token_hash, status);

CREATE INDEX IF NOT EXISTS api_keys_secret_hash_idx
ON paisa.api_keys(secret_hash, status);

CREATE INDEX IF NOT EXISTS member_identifiers_lookup_idx
ON paisa.member_identifiers(partner_id, type, value_hash);

CREATE INDEX IF NOT EXISTS catalog_items_partner_status_idx
ON paisa.catalog_items(partner_id, status);

CREATE INDEX IF NOT EXISTS redemptions_partner_member_idx
ON paisa.redemptions(partner_id, member_id, created_at);

CREATE INDEX IF NOT EXISTS integration_connections_partner_provider_idx
ON paisa.integration_connections(partner_id, provider, status);

CREATE INDEX IF NOT EXISTS reward_calculations_partner_status_created_idx
ON paisa.reward_calculations(partner_id, status, created_at);

CREATE INDEX IF NOT EXISTS ledger_entries_partner_account_created_idx
ON paisa.ledger_entries(partner_id, member_account_id, created_at);

CREATE INDEX IF NOT EXISTS ledger_entries_partner_source_idx
ON paisa.ledger_entries(partner_id, source_type, source_id);

CREATE UNIQUE INDEX IF NOT EXISTS ledger_entries_one_per_source_entry
ON paisa.ledger_entries(partner_id, source_type, source_id, entry_type);

CREATE INDEX IF NOT EXISTS ledger_exports_partner_status_idx
ON paisa.ledger_exports(partner_id, status);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'transaction_events_type_check') THEN
        ALTER TABLE paisa.transaction_events
        ADD CONSTRAINT transaction_events_type_check CHECK (type IN ('purchase', 'refund'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'rule_groups_resolution_strategy_check') THEN
        ALTER TABLE paisa.rule_groups
        ADD CONSTRAINT rule_groups_resolution_strategy_check CHECK (resolution_strategy IN ('stack', 'max_of', 'waterfall'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'program_rule_versions_scope_check') THEN
        ALTER TABLE paisa.program_rule_versions
        ADD CONSTRAINT program_rule_versions_scope_check CHECK (scope IN ('program_base', 'member_add_on'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'member_rule_assignments_status_check') THEN
        ALTER TABLE paisa.member_rule_assignments
        ADD CONSTRAINT member_rule_assignments_status_check CHECK (status IN ('active', 'ended'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'earning_rules_rule_type_check') THEN
        ALTER TABLE paisa.earning_rules
        ADD CONSTRAINT earning_rules_rule_type_check CHECK (rule_type IN ('points_per_dollar', 'fixed_per_transaction', 'first_purchase_bonus', 'spend_window_bonus'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'rule_limits_scope_check') THEN
        ALTER TABLE paisa.rule_limits
        ADD CONSTRAINT rule_limits_scope_check CHECK (scope IN ('member'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'rule_limits_period_check') THEN
        ALTER TABLE paisa.rule_limits
        ADD CONSTRAINT rule_limits_period_check CHECK (period IN ('lifetime', 'day', 'calendar_month', 'calendar_year'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ledger_entries_entry_type_check') THEN
        ALTER TABLE paisa.ledger_entries
        ADD CONSTRAINT ledger_entries_entry_type_check CHECK (entry_type IN ('earn', 'refund', 'redemption_reserve', 'redemption_capture', 'reservation_release', 'adjustment', 'points_expiration'));
    END IF;
END $$;
