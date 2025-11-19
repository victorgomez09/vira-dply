package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/users"
)

// Session represents an authenticated user session
type Session struct {
	id        SessionID
	userID    users.UserID
	token     string
	expiresAt time.Time
	createdAt time.Time
	isRevoked bool
}

type SessionID struct {
	value string
}

func NewSessionID() SessionID {
	return SessionID{value: generateID()}
}

func SessionIDFromString(s string) (SessionID, error) {
	if s == "" {
		return SessionID{}, errors.New("session ID cannot be empty")
	}
	return SessionID{value: s}, nil
}

func (id SessionID) String() string {
	return id.value
}

// RefreshToken represents a refresh token for session renewal
type RefreshToken struct {
	id        RefreshTokenID
	userID    users.UserID
	sessionID SessionID
	token     string
	expiresAt time.Time
	createdAt time.Time
	isUsed    bool
}

type RefreshTokenID struct {
	value string
}

func NewRefreshTokenID() RefreshTokenID {
	return RefreshTokenID{value: generateID()}
}

func RefreshTokenIDFromString(s string) (RefreshTokenID, error) {
	if s == "" {
		return RefreshTokenID{}, errors.New("refresh token ID cannot be empty")
	}
	return RefreshTokenID{value: s}, nil
}

func (id RefreshTokenID) String() string {
	return id.value
}

// NewSession creates a new user session
func NewSession(userID users.UserID, token string, expirationDuration time.Duration) *Session {
	now := time.Now()
	return &Session{
		id:        NewSessionID(),
		userID:    userID,
		token:     token,
		expiresAt: now.Add(expirationDuration),
		createdAt: now,
		isRevoked: false,
	}
}

// Session methods
func (s *Session) ID() SessionID {
	return s.id
}

func (s *Session) UserID() users.UserID {
	return s.userID
}

func (s *Session) Token() string {
	return s.token
}

func (s *Session) ExpiresAt() time.Time {
	return s.expiresAt
}

func (s *Session) CreatedAt() time.Time {
	return s.createdAt
}

func (s *Session) IsRevoked() bool {
	return s.isRevoked
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.expiresAt)
}

func (s *Session) IsValid() bool {
	return !s.isRevoked && !s.IsExpired()
}

func (s *Session) Revoke() {
	s.isRevoked = true
}

// NewRefreshToken creates a new refresh token
func NewRefreshToken(userID users.UserID, sessionID SessionID, token string, expirationDuration time.Duration) *RefreshToken {
	now := time.Now()
	return &RefreshToken{
		id:        NewRefreshTokenID(),
		userID:    userID,
		sessionID: sessionID,
		token:     token,
		expiresAt: now.Add(expirationDuration),
		createdAt: now,
		isUsed:    false,
	}
}

// RefreshToken methods
func (rt *RefreshToken) ID() RefreshTokenID {
	return rt.id
}

func (rt *RefreshToken) UserID() users.UserID {
	return rt.userID
}

func (rt *RefreshToken) SessionID() SessionID {
	return rt.sessionID
}

func (rt *RefreshToken) Token() string {
	return rt.token
}

func (rt *RefreshToken) ExpiresAt() time.Time {
	return rt.expiresAt
}

func (rt *RefreshToken) CreatedAt() time.Time {
	return rt.createdAt
}

func (rt *RefreshToken) IsUsed() bool {
	return rt.isUsed
}

func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.expiresAt)
}

func (rt *RefreshToken) IsValid() bool {
	return !rt.isUsed && !rt.IsExpired()
}

func (rt *RefreshToken) MarkAsUsed() {
	rt.isUsed = true
}

// Reconstruction functions for loading from database
func ReconstructSession(
	id SessionID,
	userID users.UserID,
	token string,
	expiresAt, createdAt time.Time,
	isRevoked bool,
) *Session {
	return &Session{
		id:        id,
		userID:    userID,
		token:     token,
		expiresAt: expiresAt,
		createdAt: createdAt,
		isRevoked: isRevoked,
	}
}

func ReconstructRefreshToken(
	id RefreshTokenID,
	userID users.UserID,
	sessionID SessionID,
	token string,
	expiresAt, createdAt time.Time,
	isUsed bool,
) *RefreshToken {
	return &RefreshToken{
		id:        id,
		userID:    userID,
		sessionID: sessionID,
		token:     token,
		expiresAt: expiresAt,
		createdAt: createdAt,
		isUsed:    isUsed,
	}
}

// generateID is a helper function for generating unique IDs
func generateID() string {
	return uuid.New().String()
}
