import { useEffect, useMemo, useState } from "react";
import StatusPill from "./StatusPill.jsx";

export default function Cashier({ programs, catalogItems, cashier, onResolveMember, onCreateTransaction, onCreateRedemption, onValidateRedemption, onCaptureRedemption, onReleaseRedemption, embedded = false }) {
  const [lookupType, setLookupType] = useState("phone");
  const [lookupValue, setLookupValue] = useState("");
  const [programId, setProgramId] = useState(programs[0]?.id || "");
  const [amount, setAmount] = useState("12.00");
  const [category, setCategory] = useState("coffee");
  const [selectedRewardId, setSelectedRewardId] = useState("");

  const rewards = cashier.availableRewards.length ? cashier.availableRewards : catalogItems.filter((item) => item.status === "active");
  const selectedReward = useMemo(() => rewards.find((item) => item.id === selectedRewardId) || rewards[0], [rewards, selectedRewardId]);
  const availablePoints = cashier.balance.availablePoints || 0;
  const selectedRewardCost = selectedReward?.pointsCost || 0;
  const redemptionStatus = cashier.redemption?.status || "";
  const hasOpenRedemption = ["requested", "reserved", "validated"].includes(redemptionStatus);
  const canReserveReward = Boolean(cashier.member && selectedReward && availablePoints >= selectedRewardCost && !hasOpenRedemption);
  const canValidateRedemption = ["requested", "reserved"].includes(redemptionStatus);
  const canCaptureRedemption = ["reserved", "validated"].includes(redemptionStatus);
  const canReleaseRedemption = ["requested", "reserved", "validated"].includes(redemptionStatus);

  useEffect(() => {
    if (!programId && programs[0]?.id) {
      setProgramId(programs[0].id);
      return;
    }
    if (programId && !programs.some((program) => program.id === programId)) {
      setProgramId(programs[0]?.id || "");
    }
  }, [programId, programs]);

  useEffect(() => {
    if (!selectedRewardId && rewards[0]?.id) {
      setSelectedRewardId(rewards[0].id);
      return;
    }
    if (selectedRewardId && !rewards.some((reward) => reward.id === selectedRewardId)) {
      setSelectedRewardId(rewards[0]?.id || "");
    }
  }, [rewards, selectedRewardId]);

  function cents(value) {
    return Math.round(Number(value || 0) * 100);
  }

  function resolve() {
    onResolveMember({
      programId,
      identifiers: [{ type: lookupType, value: lookupValue }],
    });
  }

  function earn() {
    onCreateTransaction({
      memberId: cashier.member?.id || "",
      customer: {
        programId,
        identifiers: [{ type: lookupType, value: lookupValue }],
      },
      category,
      subtotalMinor: cents(amount),
      totalMinor: cents(amount),
      eligibleMinor: cents(amount),
      currency: "USD",
      occurredAt: new Date().toISOString(),
    });
  }

  return (
    <section className="view-stack">
      {!embedded ? (
        <div className="view-header">
          <div>
            <p className="eyebrow">Cashier</p>
            <h2>Checkout</h2>
          </div>
        </div>
      ) : null}

      <div className="cashier-layout">
        <section className="panel spacious">
          <div className="section-heading">
            <div>
              <p className="eyebrow">Customer</p>
              <h3>Member</h3>
            </div>
          </div>
          <div className="form-grid">
            <label>
              Identifier
              <select value={lookupType} onChange={(event) => setLookupType(event.target.value)}>
                <option value="phone">Phone</option>
                <option value="email">Email</option>
                <option value="qr_code">QR code</option>
              </select>
            </label>
            <label>
              Value
              <input value={lookupValue} onChange={(event) => setLookupValue(event.target.value)} placeholder="+1 517 555 1212" />
            </label>
            <label>
              Default program
              <select value={programId} onChange={(event) => setProgramId(event.target.value)}>
                <option value="">No enrollment</option>
                {programs.map((program) => <option key={program.id} value={program.id}>{program.name}</option>)}
              </select>
            </label>
          </div>
          <button className="primary" type="button" onClick={resolve} disabled={!lookupValue.trim()}>Resolve member</button>

          {cashier.member ? (
            <div className="cashier-member">
              <span>Active member</span>
              <strong>{cashier.member.externalCustomerId}</strong>
              <small>{availablePoints} available pts / {cashier.balance.reservedPoints || 0} reserved</small>
            </div>
          ) : null}
        </section>

        <section className="panel spacious">
          <div className="section-heading">
            <div>
              <p className="eyebrow">Earn</p>
              <h3>Purchase</h3>
            </div>
          </div>
          <div className="form-grid">
            <label>
              Amount
              <input value={amount} onChange={(event) => setAmount(event.target.value)} inputMode="decimal" />
            </label>
            <label>
              Category
              <input value={category} onChange={(event) => setCategory(event.target.value)} />
            </label>
          </div>
          <button type="button" onClick={earn} disabled={!cashier.member && !lookupValue.trim()}>Earn points</button>
        </section>

        <section className="panel spacious">
          <div className="section-heading">
            <div>
              <p className="eyebrow">Redeem</p>
              <h3>Reward</h3>
            </div>
          </div>
          <div className="form-grid">
            <label>
              Reward
              <select value={selectedReward?.id || ""} onChange={(event) => setSelectedRewardId(event.target.value)}>
                {rewards.map((item) => <option key={item.id} value={item.id}>{item.name} - {item.pointsCost} pts</option>)}
              </select>
            </label>
          </div>
          <div className="button-row">
            <button type="button" onClick={() => selectedReward && onCreateRedemption(selectedReward.id)} disabled={!canReserveReward}>Reserve</button>
            <button type="button" onClick={() => cashier.redemption && onValidateRedemption(cashier.redemption.id)} disabled={!canValidateRedemption}>Validate</button>
            <button className="primary" type="button" onClick={() => cashier.redemption && onCaptureRedemption(cashier.redemption.id)} disabled={!canCaptureRedemption}>Capture</button>
            <button type="button" onClick={() => cashier.redemption && onReleaseRedemption(cashier.redemption.id)} disabled={!canReleaseRedemption}>Release</button>
          </div>
          {cashier.member && selectedReward && !canReserveReward ? (
            <p className="field-hint">
              {hasOpenRedemption ? "Finish or release the open redemption before reserving another reward." : `Needs ${selectedRewardCost - availablePoints} more points to reserve this reward.`}
            </p>
          ) : null}
          {cashier.redemption ? (
            <div className="redemption-ticket">
              <span>Reward code</span>
              <strong>{cashier.redemption.code}</strong>
              <StatusPill value={cashier.redemption.status} />
            </div>
          ) : null}
        </section>
      </div>
    </section>
  );
}
