package application

import (
	"context"
	"testing"

	"accts-api/domain"
	"accts-api/ports"
)

func TestIngestTransactionReplaysSamePayload(t *testing.T) {
	fake := newFakeAppStore()
	body := domain.TransactionIngestRequest{
		ExternalTransactionID: "txn_1",
		ExternalCustomerID:    "cust_1",
		Type:                  domain.EventPurchase,
		Currency:              "USD",
		EligibleMinor:         10000,
	}
	fake.transactionIdentity = ports.TransactionIdentity{ID: "evt_existing", PayloadHash: domain.HashTransactionPayload(body)}
	fake.transactionEvent = domain.TransactionEvent{ID: "evt_existing"}

	event, err := TransactionIngestionService{app: testApp(fake)}.IngestTransaction(context.Background(), "acme-demo", body)
	if err != nil {
		t.Fatal(err)
	}
	if event.ID != "evt_existing" {
		t.Fatalf("expected replayed event, got %q", event.ID)
	}
	if fake.transactionCreates != 0 {
		t.Fatalf("expected no new transaction create, got %d", fake.transactionCreates)
	}
}

func TestIngestTransactionRejectsChangedIdempotencyPayload(t *testing.T) {
	fake := newFakeAppStore()
	fake.transactionIdentity = ports.TransactionIdentity{ID: "evt_existing", PayloadHash: "different"}

	_, err := TransactionIngestionService{app: testApp(fake)}.IngestTransaction(context.Background(), "acme-demo", domain.TransactionIngestRequest{
		ExternalTransactionID: "txn_1",
		ExternalCustomerID:    "cust_1",
		EligibleMinor:         10000,
	})
	if !domain.IsErrorKind(err, domain.ErrorKindConflict) {
		t.Fatalf("expected conflict error, got %v", err)
	}
}

func TestIngestTransactionRejectsUnsupportedType(t *testing.T) {
	fake := newFakeAppStore()

	_, err := TransactionIngestionService{app: testApp(fake)}.IngestTransaction(context.Background(), "acme-demo", domain.TransactionIngestRequest{
		ExternalTransactionID: "txn_1",
		ExternalCustomerID:    "cust_1",
		Type:                  "exchange",
		EligibleMinor:         10000,
	})
	if !domain.IsErrorKind(err, domain.ErrorKindInvalid) {
		t.Fatalf("expected invalid type error, got %v", err)
	}
}

func TestIngestTransactionRejectsSuspendedMember(t *testing.T) {
	fake := newFakeAppStore()
	fake.member.Status = domain.StatusSuspended

	_, err := TransactionIngestionService{app: testApp(fake)}.IngestTransaction(context.Background(), "acme-demo", domain.TransactionIngestRequest{
		ExternalTransactionID: "txn_1",
		ExternalCustomerID:    "cust_1",
		Type:                  domain.EventPurchase,
		EligibleMinor:         10000,
	})
	if !domain.IsErrorKind(err, domain.ErrorKindInvariant) {
		t.Fatalf("expected invariant error, got %v", err)
	}
}
