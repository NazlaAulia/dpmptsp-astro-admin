package application

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"dpmptsp/api/internal/domain"
)

// ErrRateLimited is returned when one address submits too often.
var ErrRateLimited = errors.New("too many submissions")

const (
	maxPerWindow   = 5
	rateWindow     = time.Hour
	ticketAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no I, O, 0 or 1
)

type ContactService struct {
	messages domain.ContactRepository
}

func NewContactService(m domain.ContactRepository) *ContactService {
	return &ContactService{messages: m}
}

// Submit validates, rate limits, and stores a message.
func (s *ContactService) Submit(ctx context.Context, m *domain.ContactMessage) error {
	m.Nama = strings.TrimSpace(m.Nama)
	m.Pesan = strings.TrimSpace(m.Pesan)

	if m.Nama == "" {
		return fmt.Errorf("%w: nama wajib diisi", domain.ErrInvalid)
	}
	if m.Pesan == "" {
		return fmt.Errorf("%w: pesan wajib diisi", domain.ErrInvalid)
	}

	// The public form is unauthenticated, so the address is the only thing to
	// limit on.
	if m.IPAddress != "" {
		n, err := s.messages.CountRecentFromIP(ctx, m.IPAddress, time.Now().Add(-rateWindow))
		if err != nil {
			return err
		}
		if n >= maxPerWindow {
			return ErrRateLimited
		}
	}

	code, err := newTicketCode()
	if err != nil {
		return err
	}
	m.TicketCode = code
	m.Status = "baru"
	return s.messages.Create(ctx, m)
}

// Track looks a message up by its ticket code.
func (s *ContactService) Track(ctx context.Context, code string) (*domain.ContactMessage, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return nil, domain.ErrNotFound
	}
	return s.messages.ByTicket(ctx, code)
}

// newTicketCode builds an unguessable ticket reference.
//
// The code is the only credential for reading a message back, so it is random
// rather than sequential: a counter would let anyone read other people's
// complaints by incrementing it.
func newTicketCode() (string, error) {
	var b strings.Builder
	b.WriteString("DPM-")
	b.WriteString(time.Now().UTC().Format("20060102"))
	b.WriteByte('-')

	max := big.NewInt(int64(len(ticketAlphabet)))
	for i := 0; i < 8; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("generate ticket code: %w", err)
		}
		b.WriteByte(ticketAlphabet[n.Int64()])
	}
	return b.String(), nil
}
