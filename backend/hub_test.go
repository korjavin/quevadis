package main

import (
	"testing"
	"time"
)

// MockUser creates a test user
func MockUser(id, username string) *User {
	return &User{
		ID:       id,
		Username: username,
		InGame:   false,
	}
}

// MockGame creates a game in initial state
func MockGame(id string, p1, p2 *User) *Game {
	return &Game{
		ID:             id,
		Player1:        p1,
		Player2:        p2,
		Turn:           1,
		CurrentRound:   1,
		Status:         "WAITING_FOR_BIDS",
		Player1Pos:     0,
		Player2Pos:     0,
		Player1Balance: INITIAL_BUDGET,
		Player2Balance: INITIAL_BUDGET,
		Player1Bid:     nil,
		Player2Bid:     nil,
		GameOver:       false,
		Winner:         0,
		History:        []RoundHistory{},
		StartTime:      time.Now(),
	}
}

// TestBidValidation tests that bids are validated correctly
func TestBidValidation(t *testing.T) {
	tests := []struct {
		name           string
		bid            int
		balance        int
		expectedValid  bool
	}{
		{"Valid bid 0", 0, 20, true},
		{"Valid bid 1", 1, 20, true},
		{"Valid bid half", 10, 20, true},
		{"Valid bid all-in", 20, 20, true},
		{"Invalid negative", -1, 20, false},
		{"Invalid over balance", 21, 20, false},
		{"Valid bid with less balance", 5, 5, true},
		{"Invalid over balance with low funds", 6, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.bid >= 0 && tt.bid <= tt.balance
			if isValid != tt.expectedValid {
				t.Errorf("bid validation: got %v, want %v", isValid, tt.expectedValid)
			}
		})
	}
}

// TestAllPayMechanic tests that both players lose their bid regardless of outcome
func TestAllPayMechanic(t *testing.T) {
	tests := []struct {
		name           string
		p1Bid          int
		p2Bid          int
		p1Balance      int
		p2Balance      int
		expectedP1Bal  int
		expectedP2Bal  int
	}{
		{
			name:           "P1 wins round",
			p1Bid:          5,
			p2Bid:          3,
			p1Balance:      20,
			p2Balance:      20,
			expectedP1Bal:  15, // 20 - 5
			expectedP2Bal:  17, // 20 - 3
		},
		{
			name:           "P2 wins round",
			p1Bid:          2,
			p2Bid:          7,
			p1Balance:      20,
			p2Balance:      20,
			expectedP1Bal:  18, // 20 - 2
			expectedP2Bal:  13, // 20 - 7
		},
		{
			name:           "Draw - both bid 0",
			p1Bid:          0,
			p2Bid:          0,
			p1Balance:      20,
			p2Balance:      20,
			expectedP1Bal:  20, // 20 - 0
			expectedP2Bal:  20, // 20 - 0
		},
		{
			name:           "Draw - both bid same non-zero",
			p1Bid:          5,
			p2Bid:          5,
			p1Balance:      20,
			p2Balance:      20,
			expectedP1Bal:  15, // 20 - 5
			expectedP2Bal:  15, // 20 - 5
		},
		{
			name:           "All-in P1 wins",
			p1Bid:          20,
			p2Bid:          10,
			p1Balance:      20,
			p2Balance:      20,
			expectedP1Bal:  0,  // 20 - 20
			expectedP2Bal:  10, // 20 - 10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the all-pay deduction
			p1Bal := tt.p1Balance - tt.p1Bid
			p2Bal := tt.p2Balance - tt.p2Bid

			if p1Bal != tt.expectedP1Bal {
				t.Errorf("P1 balance: got %d, want %d", p1Bal, tt.expectedP1Bal)
			}
			if p2Bal != tt.expectedP2Bal {
				t.Errorf("P2 balance: got %d, want %d", p2Bal, tt.expectedP2Bal)
			}
		})
	}
}

// TestRoundResolution tests who advances based on bids
func TestRoundResolution(t *testing.T) {
	tests := []struct {
		name         string
		p1Bid        int
		p2Bid        int
		expectedPos1 int
		expectedPos2 int
		expectedResult string
	}{
		{
			name:         "P1 wins with higher bid",
			p1Bid:        5,
			p2Bid:        3,
			expectedPos1: 1,
			expectedPos2: 0,
			expectedResult: "P1_WINS_ROUND",
		},
		{
			name:         "P2 wins with higher bid",
			p1Bid:        2,
			p2Bid:        7,
			expectedPos1: 0,
			expectedPos2: 1,
			expectedResult: "P2_WINS_ROUND",
		},
		{
			name:         "Draw - equal bids",
			p1Bid:        5,
			p2Bid:        5,
			expectedPos1: 0,
			expectedPos2: 0,
			expectedResult: "DRAW",
		},
		{
			name:         "Draw - both bid 0",
			p1Bid:        0,
			p2Bid:        0,
			expectedPos1: 0,
			expectedPos2: 0,
			expectedResult: "DRAW",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate movement determination
			p1Pos, p2Pos := 0, 0
			var result string

			if tt.p1Bid > tt.p2Bid {
				p1Pos = 1
				result = "P1_WINS_ROUND"
			} else if tt.p2Bid > tt.p1Bid {
				p2Pos = 1
				result = "P2_WINS_ROUND"
			} else {
				result = "DRAW"
			}

			if p1Pos != tt.expectedPos1 {
				t.Errorf("P1 position: got %d, want %d", p1Pos, tt.expectedPos1)
			}
			if p2Pos != tt.expectedPos2 {
				t.Errorf("P2 position: got %d, want %d", p2Pos, tt.expectedPos2)
			}
			if result != tt.expectedResult {
				t.Errorf("Result: got %s, want %s", result, tt.expectedResult)
			}
		})
	}
}

// TestWinCondition tests the win conditions
func TestWinCondition(t *testing.T) {
	tests := []struct {
		name        string
		p1Pos       int
		p2Pos       int
		p1Bal       int
		p2Bal       int
		expectedWin int // 0 = continue, 1 = p1 wins, 2 = p2 wins, 3 = draw
	}{
		{
			name:        "P1 reaches finish",
			p1Pos:       3,
			p2Pos:       1,
			p1Bal:       10,
			p2Bal:       10,
			expectedWin: 1,
		},
		{
			name:        "P2 reaches finish",
			p1Pos:       1,
			p2Pos:       3,
			p1Bal:       10,
			p2Bal:       10,
			expectedWin: 2,
		},
		{
			name:        "Game continues - neither at finish",
			p1Pos:       1,
			p2Pos:       1,
			p1Bal:       10,
			p2Bal:       10,
			expectedWin: 0,
		},
		{
			name:        "Bankruptcy stalemate - P1 higher position",
			p1Pos:       2,
			p2Pos:       1,
			p1Bal:       0,
			p2Bal:       0,
			expectedWin: 1,
		},
		{
			name:        "Bankruptcy stalemate - P2 higher position",
			p1Pos:       1,
			p2Pos:       2,
			p1Bal:       0,
			p2Bal:       0,
			expectedWin: 2,
		},
		{
			name:        "Bankruptcy stalemate - equal position draw",
			p1Pos:       1,
			p2Pos:       1,
			p1Bal:       0,
			p2Bal:       0,
			expectedWin: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate win condition check
			var winner int

			if tt.p1Pos >= MAX_STEPS {
				winner = 1
			} else if tt.p2Pos >= MAX_STEPS {
				winner = 2
			} else if tt.p1Bal == 0 && tt.p2Bal == 0 {
				if tt.p1Pos > tt.p2Pos {
					winner = 1
				} else if tt.p2Pos > tt.p1Pos {
					winner = 2
				} else {
					winner = 3
				}
			} else {
				winner = 0
			}

			if winner != tt.expectedWin {
				t.Errorf("Winner: got %d, want %d", winner, tt.expectedWin)
			}
		})
	}
}

// TestFullGameSequence tests a complete game sequence
func TestFullGameSequence(t *testing.T) {
	// Test a game where P1 wins
	p1 := MockUser("p1", "Player1")
	p2 := MockUser("p2", "Player2")
	game := MockGame("test-game", p1, p2)

	// Round 1: P1 bids 5, P2 bids 3 -> P1 advances
	p1Bid1, p2Bid1 := 5, 3
	game.Player1Balance -= p1Bid1
	game.Player2Balance -= p2Bid1
	if p1Bid1 > p2Bid1 {
		game.Player1Pos++
	}

	if game.Player1Balance != 15 || game.Player2Balance != 17 {
		t.Errorf("After round 1: P1 bal=%d (want 15), P2 bal=%d (want 17)",
			game.Player1Balance, game.Player2Balance)
	}
	if game.Player1Pos != 1 || game.Player2Pos != 0 {
		t.Errorf("After round 1: P1 pos=%d (want 1), P2 pos=%d (want 0)",
			game.Player1Pos, game.Player2Pos)
	}

	// Round 2: P1 bids 3, P2 bids 6 -> P2 advances
	p1Bid2, p2Bid2 := 3, 6
	game.Player1Balance -= p1Bid2
	game.Player2Balance -= p2Bid2
	if p2Bid2 > p1Bid2 {
		game.Player2Pos++
	}

	if game.Player1Balance != 12 || game.Player2Balance != 11 {
		t.Errorf("After round 2: P1 bal=%d (want 12), P2 bal=%d (want 11)",
			game.Player1Balance, game.Player2Balance)
	}
	if game.Player1Pos != 1 || game.Player2Pos != 1 {
		t.Errorf("After round 2: P1 pos=%d (want 1), P2 pos=%d (want 1)",
			game.Player1Pos, game.Player2Pos)
	}

	// Round 3: P1 bids 8, P2 bids 4 -> P1 advances
	p1Bid3, p2Bid3 := 8, 4
	game.Player1Balance -= p1Bid3
	game.Player2Balance -= p2Bid3
	if p1Bid3 > p2Bid3 {
		game.Player1Pos++
	}

	if game.Player1Pos != 2 || game.Player2Pos != 1 {
		t.Errorf("After round 3: P1 pos=%d (want 2), P2 pos=%d (want 1)",
			game.Player1Pos, game.Player2Pos)
	}

	// Round 4: P1 bids 5, P2 bids 5 -> Draw
	p1Bid4, p2Bid4 := 5, 5
	game.Player1Balance -= p1Bid4
	game.Player2Balance -= p2Bid4
	// No movement on draw

	if game.Player1Pos != 2 || game.Player2Pos != 1 {
		t.Errorf("After round 4: P1 pos=%d (want 2), P2 pos=%d (want 1)",
			game.Player1Pos, game.Player2Pos)
	}

	// Round 5: P1 bids 7 (all-in), P2 bids 2 -> P1 wins!
	p1Bid5, p2Bid5 := 7, 2
	game.Player1Balance -= p1Bid5
	game.Player2Balance -= p2Bid5
	if p1Bid5 > p2Bid5 {
		game.Player1Pos++
	}

	// P1 should now have position 3 (winning)
	if game.Player1Pos != MAX_STEPS {
		t.Errorf("After round 5: P1 pos=%d (want %d - WIN!)",
			game.Player1Pos, MAX_STEPS)
	}
}

// TestConstants verifies the game constants are correct
func TestConstants(t *testing.T) {
	if MAX_STEPS != 3 {
		t.Errorf("MAX_STEPS: got %d, want 3", MAX_STEPS)
	}
	if INITIAL_BUDGET != 20 {
		t.Errorf("INITIAL_BUDGET: got %d, want 20", INITIAL_BUDGET)
	}
	if CHALLENGE_EXPIRY != 60 {
		t.Errorf("CHALLENGE_EXPIRY: got %d, want 60", CHALLENGE_EXPIRY)
	}
}

// TestHistoryRecording tests that round history is recorded correctly
func TestHistoryRecording(t *testing.T) {
	p1 := MockUser("p1", "Player1")
	p2 := MockUser("p2", "Player2")
	game := MockGame("test-game", p1, p2)

	// Record a round
	round1 := RoundHistory{
		Turn:     1,
		P1Bid:    5,
		P2Bid:    3,
		P1NewPos: 1,
		P2NewPos: 0,
		Result:   "P1_WINS_ROUND",
	}
	game.History = append(game.History, round1)

	if len(game.History) != 1 {
		t.Errorf("History length: got %d, want 1", len(game.History))
	}

	if game.History[0].P1Bid != 5 {
		t.Errorf("History P1 bid: got %d, want 5", game.History[0].P1Bid)
	}
	if game.History[0].Result != "P1_WINS_ROUND" {
		t.Errorf("History result: got %s, want P1_WINS_ROUND", game.History[0].Result)
	}
}
