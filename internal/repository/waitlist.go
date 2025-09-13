package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type WaitlistRepository struct {
	redis *redis.Client
}

type WaitlistEntry struct {
	UserID    uint      `json:"user_id"`
	EventID   uint      `json:"event_id"`
	JoinedAt  time.Time `json:"joined_at"`
	Position  int       `json:"position"`
	NotifiedAt *time.Time `json:"notified_at,omitempty"`
}

func NewWaitlistRepository(redis *redis.Client) *WaitlistRepository {
	return &WaitlistRepository{
		redis: redis,
	}
}

// JoinWaitlist adds a user to the event waitlist queue
func (r *WaitlistRepository) JoinWaitlist(ctx context.Context, userID, eventID uint) (*WaitlistEntry, error) {
	queueKey := fmt.Sprintf("waitlist:event:%d", eventID)
	userKey := fmt.Sprintf("waitlist:user:%d:event:%d", userID, eventID)
	
	// Check if user is already in the waitlist
	exists, err := r.redis.Exists(ctx, userKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to check if user exists in waitlist: %w", err)
	}
	
	if exists > 0 {
		// User already in waitlist, return current position
		return r.GetWaitlistPosition(ctx, userID, eventID)
	}
	
	entry := &WaitlistEntry{
		UserID:   userID,
		EventID:  eventID,
		JoinedAt: time.Now(),
	}
	
	// Serialize entry
	entryJSON, err := json.Marshal(entry)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal waitlist entry: %w", err)
	}
	
	// Use Redis pipeline for atomic operations
	pipe := r.redis.Pipeline()
	
	// Add to the end of the queue (FIFO)
	pipe.RPush(ctx, queueKey, string(entryJSON))
	
	// Set user-specific key for quick lookups
	pipe.Set(ctx, userKey, string(entryJSON), 24*time.Hour) // Expire after 24 hours
	
	// Get current queue length to determine position
	pipe.LLen(ctx, queueKey)
	
	results, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to join waitlist: %w", err)
	}
	
	// Get the queue length (position)
	queueLength := results[2].(*redis.IntCmd).Val()
	entry.Position = int(queueLength)
	
	// Update the entry with position
	entry.Position = int(queueLength)
	entryJSON, _ = json.Marshal(entry)
	r.redis.Set(ctx, userKey, string(entryJSON), 24*time.Hour)
	
	return entry, nil
}

// GetWaitlistPosition returns the current position of a user in the waitlist
func (r *WaitlistRepository) GetWaitlistPosition(ctx context.Context, userID, eventID uint) (*WaitlistEntry, error) {
	userKey := fmt.Sprintf("waitlist:user:%d:event:%d", userID, eventID)
	
	entryJSON, err := r.redis.Get(ctx, userKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("user not found in waitlist")
		}
		return nil, fmt.Errorf("failed to get waitlist position: %w", err)
	}
	
	var entry WaitlistEntry
	if err := json.Unmarshal([]byte(entryJSON), &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal waitlist entry: %w", err)
	}
	
	// Recalculate position by finding user in the queue
	queueKey := fmt.Sprintf("waitlist:event:%d", eventID)
	queueEntries, err := r.redis.LRange(ctx, queueKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get queue entries: %w", err)
	}
	
	for i, entryStr := range queueEntries {
		var queueEntry WaitlistEntry
		if err := json.Unmarshal([]byte(entryStr), &queueEntry); err != nil {
			continue
		}
		
		if queueEntry.UserID == userID {
			entry.Position = i + 1 // 1-based position
			break
		}
	}
	
	return &entry, nil
}

// RemoveFromWaitlist removes a user from the waitlist
func (r *WaitlistRepository) RemoveFromWaitlist(ctx context.Context, userID, eventID uint) error {
	queueKey := fmt.Sprintf("waitlist:event:%d", eventID)
	userKey := fmt.Sprintf("waitlist:user:%d:event:%d", userID, eventID)
	
	// Get user's entry first
	entryJSON, err := r.redis.Get(ctx, userKey).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("user not found in waitlist")
		}
		return fmt.Errorf("failed to get user waitlist entry: %w", err)
	}
	
	// Remove from queue and user key
	pipe := r.redis.Pipeline()
	pipe.LRem(ctx, queueKey, 1, entryJSON)
	pipe.Del(ctx, userKey)
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to remove from waitlist: %w", err)
	}
	
	return nil
}

// GetNextInWaitlist gets the next user in line for an event
func (r *WaitlistRepository) GetNextInWaitlist(ctx context.Context, eventID uint) (*WaitlistEntry, error) {
	queueKey := fmt.Sprintf("waitlist:event:%d", eventID)
	
	// Get the first entry in the queue (FIFO)
	entryJSON, err := r.redis.LIndex(ctx, queueKey, 0).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Empty queue
		}
		return nil, fmt.Errorf("failed to get next in waitlist: %w", err)
	}
	
	var entry WaitlistEntry
	if err := json.Unmarshal([]byte(entryJSON), &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal waitlist entry: %w", err)
	}
	
	entry.Position = 1
	return &entry, nil
}

// PopFromWaitlist removes and returns the first user in the waitlist
func (r *WaitlistRepository) PopFromWaitlist(ctx context.Context, eventID uint) (*WaitlistEntry, error) {
	queueKey := fmt.Sprintf("waitlist:event:%d", eventID)
	
	// Pop the first entry from the queue
	entryJSON, err := r.redis.LPop(ctx, queueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Empty queue
		}
		return nil, fmt.Errorf("failed to pop from waitlist: %w", err)
	}
	
	var entry WaitlistEntry
	if err := json.Unmarshal([]byte(entryJSON), &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal waitlist entry: %w", err)
	}
	
	// Remove user-specific key
	userKey := fmt.Sprintf("waitlist:user:%d:event:%d", entry.UserID, eventID)
	r.redis.Del(ctx, userKey)
	
	return &entry, nil
}

// GetWaitlistSize returns the number of people waiting for an event
func (r *WaitlistRepository) GetWaitlistSize(ctx context.Context, eventID uint) (int, error) {
	queueKey := fmt.Sprintf("waitlist:event:%d", eventID)
	
	size, err := r.redis.LLen(ctx, queueKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get waitlist size: %w", err)
	}
	
	return int(size), nil
}

// NotifyWaitlistUsers marks users as notified when seats become available
func (r *WaitlistRepository) NotifyWaitlistUsers(ctx context.Context, eventID uint, count int) ([]*WaitlistEntry, error) {
	queueKey := fmt.Sprintf("waitlist:event:%d", eventID)
	
	// Get the first 'count' entries without removing them
	entryJSONs, err := r.redis.LRange(ctx, queueKey, 0, int64(count-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get waitlist entries for notification: %w", err)
	}
	
	var notifiedUsers []*WaitlistEntry
	now := time.Now()
	
	for _, entryJSON := range entryJSONs {
		var entry WaitlistEntry
		if err := json.Unmarshal([]byte(entryJSON), &entry); err != nil {
			continue
		}
		
		// Mark as notified
		entry.NotifiedAt = &now
		notifiedUsers = append(notifiedUsers, &entry)
		
		// Update user-specific key with notification time
		userKey := fmt.Sprintf("waitlist:user:%d:event:%d", entry.UserID, eventID)
		updatedJSON, _ := json.Marshal(entry)
		r.redis.Set(ctx, userKey, string(updatedJSON), 24*time.Hour)
	}
	
	return notifiedUsers, nil
}

// CleanupExpiredNotifications removes users who were notified but didn't book within the time limit
func (r *WaitlistRepository) CleanupExpiredNotifications(ctx context.Context, eventID uint, notificationTTL time.Duration) error {
	queueKey := fmt.Sprintf("waitlist:event:%d", eventID)
	
	// Get all entries in the queue
	entryJSONs, err := r.redis.LRange(ctx, queueKey, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to get waitlist entries for cleanup: %w", err)
	}
	
	cutoffTime := time.Now().Add(-notificationTTL)
	
	for _, entryJSON := range entryJSONs {
		var entry WaitlistEntry
		if err := json.Unmarshal([]byte(entryJSON), &entry); err != nil {
			continue
		}
		
		// If user was notified and the notification has expired, remove them
		if entry.NotifiedAt != nil && entry.NotifiedAt.Before(cutoffTime) {
			userKey := fmt.Sprintf("waitlist:user:%d:event:%d", entry.UserID, eventID)
			
			// Remove from both queue and user key
			pipe := r.redis.Pipeline()
			pipe.LRem(ctx, queueKey, 1, entryJSON)
			pipe.Del(ctx, userKey)
			pipe.Exec(ctx)
		}
	}
	
	return nil
}
