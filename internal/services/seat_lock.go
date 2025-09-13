package services

import (
	"api/constants"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SeatLockService struct {
	redis *redis.Client
}

// Ensure SeatLockService implements SeatLockServiceInterface
var _ SeatLockServiceInterface = (*SeatLockService)(nil)

func NewSeatLockService(redisClient *redis.Client) *SeatLockService {
	return &SeatLockService{
		redis: redisClient,
	}
}

// LockSeat creates a lock for a specific seat with TTL
func (s *SeatLockService) LockSeat(ctx context.Context, seatID uint, userID uint, intentID string) error {
	key := fmt.Sprintf("%s%d", constants.SeatLockPrefix, seatID)
	value := fmt.Sprintf("%d:%s", userID, intentID)

	// Try to set the lock with NX (only if not exists) and TTL
	result := s.redis.SetNX(ctx, key, value, time.Duration(constants.SeatLockDuration)*time.Minute)
	if result.Err() != nil {
		return fmt.Errorf("failed to create seat lock: %w", result.Err())
	}

	if !result.Val() {
		return fmt.Errorf("seat is already locked")
	}

	return nil
}

// UnlockSeat removes the lock for a specific seat
func (s *SeatLockService) UnlockSeat(ctx context.Context, seatID uint, userID uint, intentID string) error {
	key := fmt.Sprintf("%s%d", constants.SeatLockPrefix, seatID)
	expectedValue := fmt.Sprintf("%d:%s", userID, intentID)

	// Lua script to atomically check and delete
	script := `
		local key = KEYS[1]
		local expected = ARGV[1]
		local current = redis.call('GET', key)
		if current == expected then
			return redis.call('DEL', key)
		else
			return 0
		end
	`

	result := s.redis.Eval(ctx, script, []string{key}, expectedValue)
	if result.Err() != nil {
		return fmt.Errorf("failed to unlock seat: %w", result.Err())
	}

	return nil
}

// IsLocked checks if a seat is currently locked
func (s *SeatLockService) IsLocked(ctx context.Context, seatID uint) (bool, string, error) {
	key := fmt.Sprintf("%s%d", constants.SeatLockPrefix, seatID)

	result := s.redis.Get(ctx, key)
	if result.Err() == redis.Nil {
		return false, "", nil
	}
	if result.Err() != nil {
		return false, "", fmt.Errorf("failed to check seat lock: %w", result.Err())
	}

	return true, result.Val(), nil
}

// ExtendLock extends the TTL of an existing lock
func (s *SeatLockService) ExtendLock(ctx context.Context, seatID uint, userID uint, intentID string) error {
	key := fmt.Sprintf("%s%d", constants.SeatLockPrefix, seatID)
	expectedValue := fmt.Sprintf("%d:%s", userID, intentID)

	// Lua script to atomically check and extend TTL
	script := `
		local key = KEYS[1]
		local expected = ARGV[1]
		local ttl = ARGV[2]
		local current = redis.call('GET', key)
		if current == expected then
			return redis.call('EXPIRE', key, ttl)
		else
			return 0
		end
	`

	ttlSeconds := constants.SeatLockDuration * 60
	result := s.redis.Eval(ctx, script, []string{key}, expectedValue, ttlSeconds)
	if result.Err() != nil {
		return fmt.Errorf("failed to extend seat lock: %w", result.Err())
	}

	if result.Val().(int64) == 0 {
		return fmt.Errorf("lock not found or not owned by user")
	}

	return nil
}

// GetLockTTL returns the remaining TTL for a seat lock
func (s *SeatLockService) GetLockTTL(ctx context.Context, seatID uint) (time.Duration, error) {
	key := fmt.Sprintf("%s%d", constants.SeatLockPrefix, seatID)

	result := s.redis.TTL(ctx, key)
	if result.Err() != nil {
		return 0, fmt.Errorf("failed to get lock TTL: %w", result.Err())
	}

	return result.Val(), nil
}

// CleanupExpiredLocks removes expired locks (this should be called periodically)
func (s *SeatLockService) CleanupExpiredLocks(ctx context.Context) error {
	pattern := constants.SeatLockPrefix + "*"

	keys, err := s.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get lock keys: %w", err)
	}

	for _, key := range keys {
		ttl, err := s.redis.TTL(ctx, key).Result()
		if err != nil {
			continue
		}

		// If TTL is -1 (no expiry) or -2 (key doesn't exist), clean it up
		if ttl < 0 {
			s.redis.Del(ctx, key)
		}
	}

	return nil
}
