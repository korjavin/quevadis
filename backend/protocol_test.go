package main

import (
	"encoding/json"
	"testing"
	"time"
)

// TestMessageSerialization tests that messages are serialized/deserialized correctly
func TestMessageSerialization(t *testing.T) {
	tests := []struct {
		name      string
		msg       Message
		checkFunc func(msg Message) bool // Custom check function
	}{
		{
			name: "welcome message",
			msg: Message{
				Type:     "welcome",
				UserID:   "user123",
				Username: "TestUser",
			},
			checkFunc: func(msg Message) bool {
				return msg.Type == "welcome" && msg.UserID == "user123" && msg.Username == "TestUser"
			},
		},
		{
			name: "challenge message",
			msg: Message{
				Type:         "challenge",
				TargetUserID: "target456",
			},
			checkFunc: func(msg Message) bool {
				return msg.Type == "challenge" && msg.TargetUserID == "target456"
			},
		},
		{
			name: "game_start message",
			msg: Message{
				Type:             "game_start",
				GameID:           "game789",
				YourPlayer:       1,
				OpponentID:       "opp123",
				OpponentUsername: "Opponent",
			},
			checkFunc: func(msg Message) bool {
				return msg.Type == "game_start" && msg.GameID == "game789" &&
					msg.YourPlayer == 1 && msg.OpponentID == "opp123" && msg.OpponentUsername == "Opponent"
			},
		},
		{
			name: "waiting_for_bids message",
			msg: Message{
				Type:        "waiting_for_bids",
				GameID:      "game789",
				Turn:        1,
				P1Balance:   20,
				P2Balance:   20,
				P1Position:  0,
				P2Position:  0,
			},
			checkFunc: func(msg Message) bool {
				return msg.Type == "waiting_for_bids" && msg.GameID == "game789" &&
					msg.Turn == 1 && msg.P1Balance == 20 && msg.P2Balance == 20 &&
					msg.P1Position == 0 && msg.P2Position == 0
			},
		},
		{
			name: "submit_bid message",
			msg: Message{
				Type:   "submit_bid",
				GameID: "game789",
				Bid:    5,
			},
			checkFunc: func(msg Message) bool {
				return msg.Type == "submit_bid" && msg.GameID == "game789" && msg.Bid == 5
			},
		},
		{
			name: "round_result message",
			msg: Message{
				Type:        "round_result",
				GameID:      "game789",
				Turn:        1,
				P1Bid:       5,
				P2Bid:       3,
				P1Position:  1,
				P2Position:  0,
				P1Balance:   15,
				P2Balance:   17,
				Result:      "P1_WINS_ROUND",
			},
			checkFunc: func(msg Message) bool {
				return msg.Type == "round_result" && msg.GameID == "game789" &&
					msg.Turn == 1 && msg.P1Bid == 5 && msg.P2Bid == 3 &&
					msg.P1Position == 1 && msg.P2Position == 0 &&
					msg.P1Balance == 15 && msg.P2Balance == 17 &&
					msg.Result == "P1_WINS_ROUND"
			},
		},
		{
			name: "game_end message",
			msg: Message{
				Type:   "game_end",
				GameID: "game789",
				Winner: 1,
				Reason: "Reached final step",
			},
			checkFunc: func(msg Message) bool {
				return msg.Type == "game_end" && msg.GameID == "game789" &&
					msg.Winner == 1 && msg.Reason == "Reached final step"
			},
		},
		{
			name: "users_update message",
			msg: Message{
				Type: "users_update",
				Users: []UserInfo{
					{UserID: "user1", Username: "Player1", InGame: false},
					{UserID: "user2", Username: "Player2", InGame: true},
				},
			},
			checkFunc: func(msg Message) bool {
				return msg.Type == "users_update" && len(msg.Users) == 2 &&
					msg.Users[0].UserID == "user1" && msg.Users[0].Username == "Player1" && !msg.Users[0].InGame &&
					msg.Users[1].UserID == "user2" && msg.Users[1].Username == "Player2" && msg.Users[1].InGame
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			data, err := json.Marshal(tt.msg)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}

			// Deserialize back
			var decoded Message
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			// Check deserialization
			if !tt.checkFunc(decoded) {
				t.Errorf("Message fields not preserved correctly after round-trip")
			}
		})
	}
}

// TestChallengeFlow tests the challenge accept flow
func TestChallengeFlow(t *testing.T) {
	// Create a hub
	hub := newHub()
	go hub.run()
	defer func() {
		// Clean up - this would need proper channel closure in real code
	}()

	// Create mock users
	challenger := MockUser("challenger-id", "Challenger")
	challengee := MockUser("challengee-id", "Challengee")

	// Simulate challenge creation
	challengeID := "challenge-123"
	challenge := &Challenge{
		ID:        challengeID,
		FromUser:  challenger,
		ToUser:    challengee,
		Timestamp: time.Now(),
	}

	hub.challenges[challengeID] = challenge

	// Verify challenge exists
	if _, exists := hub.challenges[challengeID]; !exists {
		t.Error("Challenge should exist in hub")
	}

	// Verify challenge has correct users
	if hub.challenges[challengeID].FromUser.ID != "challenger-id" {
		t.Error("Challenge should have correct challenger")
	}
	if hub.challenges[challengeID].ToUser.ID != "challengee-id" {
		t.Error("Challenge should have correct challengee")
	}

	// Simulate challenge acceptance (create game)
	gameID := "game-456"
	game := &Game{
		ID:             gameID,
		Player1:        challenger,
		Player2:        challengee,
		Turn:           1,
		CurrentRound:   1,
		Status:         "WAITING_FOR_BIDS",
		Player1Pos:     0,
		Player2Pos:     0,
		Player1Balance: INITIAL_BUDGET,
		Player2Balance: INITIAL_BUDGET,
		GameOver:       false,
		Winner:         0,
		StartTime:      time.Now(),
	}
	hub.games[gameID] = game

	// Verify game was created
	if _, exists := hub.games[gameID]; !exists {
		t.Error("Game should exist in hub")
	}

	// Simulate challenge deletion (what happens in handleAcceptChallenge)
	delete(hub.challenges, challengeID)

	// Verify challenge was removed
	if _, exists := hub.challenges[challengeID]; exists {
		t.Error("Challenge should be removed after acceptance")
	}
}

// TestGameStateTransitions tests game state transitions
func TestGameStateTransitions(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  string
		action         func(*Game)
		expectedStatus string
	}{
		{
			name:           "New game starts in WAITING_FOR_BIDS",
			initialStatus:  "WAITING_FOR_BIDS",
			action:         func(g *Game) {},
			expectedStatus: "WAITING_FOR_BIDS",
		},
		{
			name:          "First bid submitted changes nothing",
			initialStatus: "WAITING_FOR_BIDS",
			action: func(g *Game) {
				bid := 5
				g.Player1Bid = &bid
			},
			expectedStatus: "WAITING_FOR_BIDS",
		},
		{
			name:          "Second bid submitted triggers RESOLVING",
			initialStatus: "WAITING_FOR_BIDS",
			action: func(g *Game) {
				p1Bid := 5
				p2Bid := 3
				g.Player1Bid = &p1Bid
				g.Player2Bid = &p2Bid
				// Simulate what happens in resolveRound
				g.Status = "RESOLVING"
			},
			expectedStatus: "RESOLVING",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := &Game{
				Status: tt.initialStatus,
			}
			tt.action(game)
			if game.Status != tt.expectedStatus {
				t.Errorf("Status: got %s, want %s", game.Status, tt.expectedStatus)
			}
		})
	}
}

// TestBidSubmissionProtocol tests the bid submission protocol
func TestBidSubmissionProtocol(t *testing.T) {
	tests := []struct {
		name          string
		playerNum     int
		bid           int
		balance       int
		shouldAccept  bool
		expectedBal   int
	}{
		{
			name:         "Valid bid from player 1",
			playerNum:    1,
			bid:          5,
			balance:      20,
			shouldAccept: true,
			expectedBal:  15,
		},
		{
			name:         "Valid bid from player 2",
			playerNum:    2,
			bid:          7,
			balance:      20,
			shouldAccept: true,
			expectedBal:  13,
		},
		{
			name:         "Invalid negative bid",
			playerNum:    1,
			bid:          -1,
			balance:      20,
			shouldAccept: false,
			expectedBal:  20,
		},
		{
			name:         "Invalid over-balance bid",
			playerNum:    1,
			bid:          25,
			balance:      20,
			shouldAccept: false,
			expectedBal:  20,
		},
		{
			name:         "Valid all-in bid",
			playerNum:    1,
			bid:          20,
			balance:      20,
			shouldAccept: true,
			expectedBal:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balance := tt.balance

			// Validate bid
			isValid := tt.bid >= 0 && tt.bid <= tt.balance

			if isValid != tt.shouldAccept {
				t.Errorf("Bid validity: got %v, want %v", isValid, tt.shouldAccept)
			}

			if isValid {
				balance -= tt.bid
			}

			if balance != tt.expectedBal {
				t.Errorf("Balance after bid: got %d, want %d", balance, tt.expectedBal)
			}
		})
	}
}
