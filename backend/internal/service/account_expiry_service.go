package service

import (
	"context"
	"log"
	"sync"
	"time"
)

// AccountExpiryService periodically pauses expired accounts and groups.
type AccountExpiryService struct {
	accountRepo AccountRepository
	groupRepo   GroupRepository
	interval    time.Duration
	stopCh      chan struct{}
	stopOnce    sync.Once
	wg          sync.WaitGroup
}

func NewAccountExpiryService(accountRepo AccountRepository, groupRepo GroupRepository, interval time.Duration) *AccountExpiryService {
	return &AccountExpiryService{
		accountRepo: accountRepo,
		groupRepo:   groupRepo,
		interval:    interval,
		stopCh:      make(chan struct{}),
	}
}

func (s *AccountExpiryService) Start() {
	if s == nil || s.accountRepo == nil || s.interval <= 0 {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		s.runOnce()
		for {
			select {
			case <-ticker.C:
				s.runOnce()
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *AccountExpiryService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	s.wg.Wait()
}

func (s *AccountExpiryService) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	if s.accountRepo != nil {
		updated, err := s.accountRepo.AutoPauseExpiredAccounts(ctx, now)
		if err != nil {
			log.Printf("[AccountExpiry] Auto pause expired accounts failed: %v", err)
		} else if updated > 0 {
			log.Printf("[AccountExpiry] Auto paused %d expired accounts", updated)
		}
	}
	if s.groupRepo != nil {
		updated, err := s.groupRepo.AutoPauseExpiredGroups(ctx, now)
		if err != nil {
			log.Printf("[AccountExpiry] Auto pause expired groups failed: %v", err)
		} else if updated > 0 {
			log.Printf("[AccountExpiry] Auto paused %d expired groups", updated)
		}
	}
}
