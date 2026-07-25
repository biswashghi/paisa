package domain

func ApplyLedgerDelta(balance BalanceSnapshot, entryType string, availableDelta, reservedDelta, expiredDelta int) (BalanceSnapshot, error) {
	next := balance
	next.AvailablePoints += availableDelta
	next.ReservedPoints += reservedDelta
	next.ExpiredPoints += expiredDelta

	if next.ReservedPoints < 0 || next.ExpiredPoints < 0 {
		return BalanceSnapshot{}, InvariantError("reserved/expired balances cannot become negative")
	}
	if entryType == EntryRedemptionReserve && next.AvailablePoints < 0 {
		return BalanceSnapshot{}, InvariantError("redemption reserve cannot make available balance negative")
	}
	if entryType == EntryPointsExpiration && next.AvailablePoints < 0 {
		return BalanceSnapshot{}, InvariantError("points expiration cannot make available balance negative")
	}
	return next, nil
}
