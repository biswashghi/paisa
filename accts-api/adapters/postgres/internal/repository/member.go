package repository

import (
	"context"
	"database/sql"

	"accts-api/domain"
)

func (s MemberStore) Create(ctx context.Context, partnerID, externalCustomerID string) (domain.Member, error) {
	var member domain.Member
	err := s.q.QueryRowContext(ctx, `
		INSERT INTO paisa.members (partner_id, external_customer_id)
		VALUES ($1, $2)
		RETURNING id, partner_id, external_customer_id, status, created_at, updated_at`,
		partnerID, externalCustomerID,
	).Scan(&member.ID, &member.PartnerID, &member.ExternalCustomerID, &member.Status, &member.CreatedAt, &member.UpdatedAt)
	return member, AppErrorFromDB(err)
}

func (s MemberStore) CreateAccount(ctx context.Context, partnerID, memberID string) (domain.MemberAccount, error) {
	var account domain.MemberAccount
	err := s.q.QueryRowContext(ctx, `
		INSERT INTO paisa.member_accounts (partner_id, member_id)
		VALUES ($1, $2)
		RETURNING id, partner_id, member_id, status`, partnerID, memberID,
	).Scan(&account.ID, &account.PartnerID, &account.MemberID, &account.Status)
	return account, AppErrorFromDB(err)
}

func (s MemberStore) InsertIdentifierHash(ctx context.Context, partnerID, memberID, identifierType, valueHash string) error {
	_, err := s.q.ExecContext(ctx, `
		INSERT INTO paisa.member_identifiers (partner_id, member_id, type, value_hash)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING`, partnerID, memberID, identifierType, valueHash)
	return AppErrorFromDB(err)
}

func (s MemberStore) List(ctx context.Context, partnerID string) ([]domain.Member, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT id, partner_id, external_customer_id, status, created_at, updated_at
		FROM paisa.members
		WHERE partner_id = $1
		ORDER BY created_at DESC`, partnerID)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	defer rows.Close()

	members := []domain.Member{}
	for rows.Next() {
		var member domain.Member
		if err := rows.Scan(&member.ID, &member.PartnerID, &member.ExternalCustomerID, &member.Status, &member.CreatedAt, &member.UpdatedAt); err != nil {
			return nil, AppErrorFromDB(err)
		}
		members = append(members, member)
	}
	return members, AppErrorFromDB(rows.Err())
}

func (s MemberStore) GetByID(ctx context.Context, partnerID, memberID string) (domain.Member, error) {
	member, err := MemberByID(ctx, s.q, partnerID, memberID)
	return member, AppErrorFromDB(err)
}

func (s MemberStore) GetByExternalID(ctx context.Context, partnerID, externalCustomerID string) (domain.Member, error) {
	member, err := MemberByExternalID(ctx, s.q, partnerID, externalCustomerID)
	return member, AppErrorFromDB(err)
}

func (s MemberStore) AccountID(ctx context.Context, partnerID, memberID string) (string, error) {
	accountID, err := MemberAccountID(ctx, s.q, partnerID, memberID)
	return accountID, AppErrorFromDB(err)
}

func (s MemberStore) EndActiveEnrollment(ctx context.Context, partnerID, memberID string) error {
	_, err := s.q.ExecContext(ctx, `
		UPDATE paisa.program_enrollments
		SET status = 'ended', ended_at = now()
		WHERE partner_id = $1 AND member_id = $2 AND status = 'active'`, partnerID, memberID)
	return AppErrorFromDB(err)
}

func (s MemberStore) CreateEnrollment(ctx context.Context, partnerID, memberID, programID string, body domain.EnrollmentRequest) (domain.ProgramEnrollment, error) {
	var enrollment domain.ProgramEnrollment
	row := s.q.QueryRowContext(ctx, `
		INSERT INTO paisa.program_enrollments (
			partner_id, member_id, program_id, effective_at, change_reason, created_by_type, created_by_id
		)
		VALUES ($1, $2, $3, NULLIF($4, '')::timestamptz, NULLIF($5, ''), COALESCE(NULLIF($6, ''), 'system'), NULLIF($7, '')::uuid)
		RETURNING id, partner_id, member_id, program_id, status, started_at, ended_at, effective_at,
			COALESCE(change_reason, ''), created_by_type, COALESCE(created_by_id::text, ''), created_at`,
		partnerID, memberID, programID, body.EffectiveAt, body.ChangeReason, body.CreatedByType, body.CreatedByID)
	if err := scanEnrollment(row, &enrollment); err != nil {
		return domain.ProgramEnrollment{}, AppErrorFromDB(err)
	}
	return enrollment, nil
}

func (s MemberStore) ActiveEnrollment(ctx context.Context, partnerID, memberID string) (domain.ProgramEnrollment, error) {
	var enrollment domain.ProgramEnrollment
	row := s.q.QueryRowContext(ctx, `
		SELECT pe.id, pe.partner_id, pe.member_id, pe.program_id, p.name, pe.status, pe.started_at, pe.ended_at, pe.effective_at,
			COALESCE(pe.change_reason, ''), pe.created_by_type, COALESCE(pe.created_by_id::text, ''), pe.created_at
		FROM paisa.program_enrollments pe
		JOIN paisa.programs p ON p.id = pe.program_id
		WHERE pe.partner_id = $1 AND pe.member_id = $2 AND pe.status = 'active'`, partnerID, memberID)
	if err := scanEnrollmentWithProgram(row, &enrollment); err != nil {
		return domain.ProgramEnrollment{}, AppErrorFromDB(err)
	}
	return enrollment, nil
}

func (s MemberStore) ListEnrollments(ctx context.Context, partnerID, memberID string) ([]domain.ProgramEnrollment, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT pe.id, pe.partner_id, pe.member_id, pe.program_id, p.name, pe.status, pe.started_at, pe.ended_at, pe.effective_at,
			COALESCE(pe.change_reason, ''), pe.created_by_type, COALESCE(pe.created_by_id::text, ''), pe.created_at
		FROM paisa.program_enrollments pe
		JOIN paisa.programs p ON p.id = pe.program_id
		WHERE pe.partner_id = $1 AND pe.member_id = $2
		ORDER BY pe.started_at DESC`, partnerID, memberID)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	defer rows.Close()
	enrollments := []domain.ProgramEnrollment{}
	for rows.Next() {
		var enrollment domain.ProgramEnrollment
		if err := scanEnrollmentWithProgram(rows, &enrollment); err != nil {
			return nil, AppErrorFromDB(err)
		}
		enrollments = append(enrollments, enrollment)
	}
	return enrollments, AppErrorFromDB(rows.Err())
}

func (s MemberStore) CreateRuleAssignment(ctx context.Context, partnerID, memberID string, body domain.MemberRuleAssignmentRequest) (domain.MemberRuleAssignment, error) {
	row := s.q.QueryRowContext(ctx, `
		WITH active_enrollment AS (
			SELECT pe.id, pe.program_id
			FROM paisa.program_enrollments pe
			WHERE pe.partner_id = $1 AND pe.member_id = $2 AND pe.status = 'active'
		), selected_rule AS (
			SELECT prv.id, prv.program_id
			FROM paisa.program_rule_versions prv
			JOIN active_enrollment ae ON ae.program_id = prv.program_id
			WHERE prv.partner_id = $1 AND prv.id = $3 AND prv.scope = $4 AND prv.status = $5
		)
		INSERT INTO paisa.member_rule_assignments (
			partner_id, member_id, program_enrollment_id, rule_version_id, starts_at, reason, created_by_type, created_by_id
		)
		SELECT $1, $2, ae.id, sr.id, COALESCE(NULLIF($6, '')::timestamptz, now()), NULLIF($7, ''), COALESCE(NULLIF($8, ''), 'system'), NULLIF($9, '')::uuid
		FROM active_enrollment ae
		JOIN selected_rule sr ON sr.program_id = ae.program_id
		RETURNING id`, partnerID, memberID, body.RuleVersionID, domain.RuleScopeMemberAddOn, domain.RulePublished, body.StartsAt, body.Reason, body.CreatedByType, body.CreatedByID)
	var assignmentID string
	if err := row.Scan(&assignmentID); err != nil {
		return domain.MemberRuleAssignment{}, AppErrorFromDB(err)
	}
	return s.ruleAssignmentByID(ctx, partnerID, memberID, assignmentID)
}

func (s MemberStore) UpdateRuleAssignment(ctx context.Context, partnerID, memberID, assignmentID string, body domain.MemberRuleAssignmentUpdateRequest) (domain.MemberRuleAssignment, error) {
	status := body.Status
	if status == "" {
		status = "ended"
	}
	_, err := s.q.ExecContext(ctx, `
		UPDATE paisa.member_rule_assignments
		SET status = $1, ends_at = COALESCE(NULLIF($2, '')::timestamptz, ends_at, now()), reason = COALESCE(NULLIF($3, ''), reason), updated_at = now()
		WHERE id = $4 AND partner_id = $5 AND member_id = $6`, status, body.EndsAt, body.Reason, assignmentID, partnerID, memberID)
	if err != nil {
		return domain.MemberRuleAssignment{}, AppErrorFromDB(err)
	}
	return s.ruleAssignmentByID(ctx, partnerID, memberID, assignmentID)
}

func (s MemberStore) ActiveRuleAssignments(ctx context.Context, partnerID, memberID string) ([]domain.MemberRuleAssignment, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT mra.id, mra.partner_id, mra.member_id, mra.program_enrollment_id, mra.rule_version_id,
			COALESCE(prv.rule_set_key, ''), COALESCE(prv.name, ''), COALESCE(prv.description, ''), prv.program_id,
			mra.status, mra.starts_at, mra.ends_at, COALESCE(mra.reason, ''), mra.created_by_type,
			COALESCE(mra.created_by_id::text, ''), mra.created_at, mra.updated_at
		FROM paisa.member_rule_assignments mra
		JOIN paisa.program_rule_versions prv ON prv.id = mra.rule_version_id
		JOIN paisa.program_enrollments pe ON pe.id = mra.program_enrollment_id
		WHERE mra.partner_id = $1 AND mra.member_id = $2 AND mra.status = 'active'
			AND pe.status = 'active' AND mra.starts_at <= now() AND (mra.ends_at IS NULL OR mra.ends_at > now())
		ORDER BY mra.starts_at, mra.created_at`, partnerID, memberID)
	if err != nil {
		return nil, AppErrorFromDB(err)
	}
	defer rows.Close()
	assignments := []domain.MemberRuleAssignment{}
	for rows.Next() {
		var assignment domain.MemberRuleAssignment
		if err := scanRuleAssignment(rows, &assignment); err != nil {
			return nil, AppErrorFromDB(err)
		}
		assignments = append(assignments, assignment)
	}
	return assignments, AppErrorFromDB(rows.Err())
}

func (s MemberStore) ruleAssignmentByID(ctx context.Context, partnerID, memberID, assignmentID string) (domain.MemberRuleAssignment, error) {
	var assignment domain.MemberRuleAssignment
	row := s.q.QueryRowContext(ctx, `
		SELECT mra.id, mra.partner_id, mra.member_id, mra.program_enrollment_id, mra.rule_version_id,
			COALESCE(prv.rule_set_key, ''), COALESCE(prv.name, ''), COALESCE(prv.description, ''), prv.program_id,
			mra.status, mra.starts_at, mra.ends_at, COALESCE(mra.reason, ''), mra.created_by_type,
			COALESCE(mra.created_by_id::text, ''), mra.created_at, mra.updated_at
		FROM paisa.member_rule_assignments mra
		JOIN paisa.program_rule_versions prv ON prv.id = mra.rule_version_id
		WHERE mra.id = $1 AND mra.partner_id = $2 AND mra.member_id = $3`, assignmentID, partnerID, memberID)
	if err := scanRuleAssignment(row, &assignment); err != nil {
		return domain.MemberRuleAssignment{}, AppErrorFromDB(err)
	}
	return assignment, nil
}

func MemberByID(ctx context.Context, q Queryer, partnerID, memberID string) (domain.Member, error) {
	var member domain.Member
	err := q.QueryRowContext(ctx, `
		SELECT id, partner_id, external_customer_id, status, created_at, updated_at
		FROM paisa.members
		WHERE id = $1 AND partner_id = $2`, memberID, partnerID,
	).Scan(&member.ID, &member.PartnerID, &member.ExternalCustomerID, &member.Status, &member.CreatedAt, &member.UpdatedAt)
	return member, err
}

func MemberByExternalID(ctx context.Context, q Queryer, partnerID, externalCustomerID string) (domain.Member, error) {
	var member domain.Member
	err := q.QueryRowContext(ctx, `
		SELECT id, partner_id, external_customer_id, status, created_at, updated_at
		FROM paisa.members
		WHERE partner_id = $1 AND external_customer_id = $2`, partnerID, externalCustomerID,
	).Scan(&member.ID, &member.PartnerID, &member.ExternalCustomerID, &member.Status, &member.CreatedAt, &member.UpdatedAt)
	return member, err
}

func MemberAccountID(ctx context.Context, q Queryer, partnerID, memberID string) (string, error) {
	var accountID string
	err := q.QueryRowContext(ctx, `
		SELECT ma.id
		FROM paisa.member_accounts ma
		JOIN paisa.members m ON m.id = ma.member_id
		WHERE ma.partner_id = $1 AND ma.member_id = $2 AND ma.status = 'active' AND m.status = 'active'`, partnerID, memberID,
	).Scan(&accountID)
	return accountID, err
}

func EnrollMember(ctx context.Context, tx Queryer, partnerID, memberID, programID string) error {
	if err := EnsureProgramPartner(ctx, tx, partnerID, programID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE paisa.program_enrollments
		SET status = 'ended', ended_at = now()
		WHERE partner_id = $1 AND member_id = $2 AND status = 'active'`, partnerID, memberID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO paisa.program_enrollments (partner_id, member_id, program_id)
		VALUES ($1, $2, $3)`, partnerID, memberID, programID)
	return err
}

type scanner interface {
	Scan(...interface{}) error
}

func scanEnrollment(row scanner, enrollment *domain.ProgramEnrollment) error {
	var endedAt, effectiveAt sql.NullTime
	if err := row.Scan(
		&enrollment.ID, &enrollment.PartnerID, &enrollment.MemberID, &enrollment.ProgramID, &enrollment.Status,
		&enrollment.StartedAt, &endedAt, &effectiveAt, &enrollment.ChangeReason, &enrollment.CreatedByType,
		&enrollment.CreatedByID, &enrollment.CreatedAt,
	); err != nil {
		return err
	}
	if endedAt.Valid {
		enrollment.EndedAt = &endedAt.Time
	}
	if effectiveAt.Valid {
		enrollment.EffectiveAt = &effectiveAt.Time
	}
	return nil
}

func scanEnrollmentWithProgram(row scanner, enrollment *domain.ProgramEnrollment) error {
	var endedAt, effectiveAt sql.NullTime
	if err := row.Scan(
		&enrollment.ID, &enrollment.PartnerID, &enrollment.MemberID, &enrollment.ProgramID, &enrollment.ProgramName,
		&enrollment.Status, &enrollment.StartedAt, &endedAt, &effectiveAt, &enrollment.ChangeReason, &enrollment.CreatedByType,
		&enrollment.CreatedByID, &enrollment.CreatedAt,
	); err != nil {
		return err
	}
	if endedAt.Valid {
		enrollment.EndedAt = &endedAt.Time
	}
	if effectiveAt.Valid {
		enrollment.EffectiveAt = &effectiveAt.Time
	}
	return nil
}

func scanRuleAssignment(row scanner, assignment *domain.MemberRuleAssignment) error {
	var endsAt sql.NullTime
	if err := row.Scan(
		&assignment.ID, &assignment.PartnerID, &assignment.MemberID, &assignment.ProgramEnrollmentID, &assignment.RuleVersionID,
		&assignment.RuleSetKey, &assignment.Name, &assignment.Description, &assignment.ProgramID, &assignment.Status,
		&assignment.StartsAt, &endsAt, &assignment.Reason, &assignment.CreatedByType, &assignment.CreatedByID,
		&assignment.CreatedAt, &assignment.UpdatedAt,
	); err != nil {
		return err
	}
	if endsAt.Valid {
		assignment.EndsAt = &endsAt.Time
	}
	return nil
}
