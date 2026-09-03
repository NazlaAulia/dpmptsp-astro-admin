package application

import (
	"context"
	"strings"
	"testing"
	"time"

	"dpmptsp/api/internal/domain"
)

type fakeContactRepo struct {
	saved  []*domain.ContactMessage
	recent int64
}

func (f *fakeContactRepo) Create(_ context.Context, m *domain.ContactMessage) error {
	f.saved = append(f.saved, m)
	return nil
}

func (f *fakeContactRepo) ByTicket(_ context.Context, code string) (*domain.ContactMessage, error) {
	for _, m := range f.saved {
		if m.TicketCode == code {
			return m, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (f *fakeContactRepo) CountRecentFromIP(_ context.Context, _ string, _ time.Time) (int64, error) {
	return f.recent, nil
}

func TestSubmitRequiresNamaAndPesan(t *testing.T) {
	svc := NewContactService(&fakeContactRepo{})

	for _, m := range []*domain.ContactMessage{
		{Nama: "  ", Pesan: "halo"},
		{Nama: "Budi", Pesan: "   "},
	} {
		if err := svc.Submit(context.Background(), m); err == nil {
			t.Fatalf("expected a validation error for %+v", m)
		}
	}
}

func TestSubmitAssignsTicketAndStatus(t *testing.T) {
	repo := &fakeContactRepo{}
	svc := NewContactService(repo)

	m := &domain.ContactMessage{Nama: "Budi", Pesan: "Lampu jalan mati"}
	if err := svc.Submit(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	if m.Status != "baru" {
		t.Errorf("status = %q, want baru", m.Status)
	}
	if !strings.HasPrefix(m.TicketCode, "DPM-") || len(m.TicketCode) != 21 {
		t.Errorf("ticket %q is not the expected DPM-YYYYMMDD-XXXXXXXX shape", m.TicketCode)
	}
}

// The ticket code is the only credential for reading a message back, so two
// submissions must never be able to guess each other's.
func TestTicketCodesAreUnique(t *testing.T) {
	seen := make(map[string]bool, 500)
	for i := 0; i < 500; i++ {
		code, err := newTicketCode()
		if err != nil {
			t.Fatal(err)
		}
		if seen[code] {
			t.Fatalf("duplicate ticket code after %d draws: %s", i, code)
		}
		seen[code] = true
	}
}

// A confusable alphabet turns a phoned-in ticket number into a support call.
func TestTicketAlphabetExcludesConfusableCharacters(t *testing.T) {
	for _, c := range "IO01" {
		if strings.ContainsRune(ticketAlphabet, c) {
			t.Errorf("alphabet contains confusable %q", c)
		}
	}
}

func TestSubmitRateLimitsByAddress(t *testing.T) {
	repo := &fakeContactRepo{recent: maxPerWindow}
	svc := NewContactService(repo)

	err := svc.Submit(context.Background(), &domain.ContactMessage{
		Nama: "Budi", Pesan: "spam", IPAddress: "203.0.113.7",
	})
	if err != ErrRateLimited {
		t.Fatalf("err = %v, want ErrRateLimited", err)
	}
	if len(repo.saved) != 0 {
		t.Error("a rate-limited submission was stored anyway")
	}
}

// An empty address must not become a shared bucket that limits everyone at once.
func TestSubmitWithoutAddressIsNotRateLimited(t *testing.T) {
	repo := &fakeContactRepo{recent: maxPerWindow}
	svc := NewContactService(repo)

	if err := svc.Submit(context.Background(), &domain.ContactMessage{Nama: "Budi", Pesan: "halo"}); err != nil {
		t.Fatal(err)
	}
}

func TestTrackNormalisesCase(t *testing.T) {
	repo := &fakeContactRepo{}
	svc := NewContactService(repo)

	m := &domain.ContactMessage{Nama: "Budi", Pesan: "halo"}
	if err := svc.Submit(context.Background(), m); err != nil {
		t.Fatal(err)
	}

	got, err := svc.Track(context.Background(), "  "+strings.ToLower(m.TicketCode)+"  ")
	if err != nil {
		t.Fatal(err)
	}
	if got.TicketCode != m.TicketCode {
		t.Errorf("got %q, want %q", got.TicketCode, m.TicketCode)
	}
}
