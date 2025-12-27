package main

import (
	"time"
)

// Game Constants
const (
	MAX_STEPS       = 3  // Target position to win (positions 0, 1, 2, 3)
	INITIAL_BUDGET  = 20 // Starting points/stones
	CHALLENGE_EXPIRY = 60 // seconds
)

// Message types sent between client and server
type Message struct {
	Type             string      `json:"type"`
	UserID           string      `json:"userId,omitempty"`
	Username         string      `json:"username,omitempty"`
	TargetUserID     string      `json:"targetUserId,omitempty"`
	ChallengeID      string      `json:"challengeId,omitempty"`
	GameID           string      `json:"gameId,omitempty"`
	FromUserID       string      `json:"fromUserId,omitempty"`
	FromUsername     string      `json:"fromUsername,omitempty"`
	OpponentID       string      `json:"opponentId,omitempty"`
	OpponentUsername string      `json:"opponentUsername,omitempty"`
	YourPlayer       int         `json:"yourPlayer,omitempty"`
	Bid              int         `json:"bid,omitempty"`
	Users            []UserInfo  `json:"users,omitempty"`
	// Game state fields
	Turn             int         `json:"turn,omitempty"`
	P1Balance        int         `json:"p1Balance,omitempty"`
	P2Balance        int         `json:"p2Balance,omitempty"`
	P1Bid            int         `json:"p1Bid,omitempty"`
	P2Bid            int         `json:"p2Bid,omitempty"`
	P1Position       int         `json:"p1Position,omitempty"`
	P2Position       int         `json:"p2Position,omitempty"`
	Winner           int         `json:"winner,omitempty"`
	Reason           string      `json:"reason,omitempty"`
	Result           string      `json:"result,omitempty"` // "P1_WINS", "P2_WINS", "DRAW"
}

type UserInfo struct {
	UserID    string `json:"userId"`
	Username  string `json:"username"`
	InGame    bool   `json:"inGame"`
}

// User represents a connected client
type User struct {
	ID      string
	Username string
	Client   *Client
	InGame   bool
	GameID   string // ID of game user is in
}

// Challenge represents a game challenge between two users
type Challenge struct {
	ID        string
	FromUser  *User
	ToUser    *User
	Timestamp time.Time
}

// Game represents an active game session
type Game struct {
	ID          string
	Player1     *User
	Player2     *User
	Turn        int
	CurrentRound int
	Status      string // "WAITING_FOR_BIDS", "RESOLVING", "GAME_OVER"
	Player1Pos  int
	Player2Pos  int
	Player1Balance int
	Player2Balance int
	Player1Bid  *int
	Player2Bid  *int
	GameOver    bool
	Winner      int // 0 = none, 1 = player1, 2 = player2, 3 = draw
	History     []RoundHistory
	StartTime   time.Time
	EndTime     time.Time
}

type RoundHistory struct {
	Turn        int
	P1Bid       int
	P2Bid       int
	P1NewPos    int
	P2NewPos    int
	Result      string
}

// MessageWrapper wraps a message with its client
type MessageWrapper struct {
	client  *Client
	message *Message
}
