package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	clients      map[*Client]bool
	users        map[string]*User
	challenges   map[string]*Challenge
	games        map[string]*Game
	register     chan *Client
	unregister   chan *Client
	handleMessage chan *MessageWrapper
}

func newHub() *Hub {
	return &Hub{
		clients:      make(map[*Client]bool),
		users:        make(map[string]*User),
		challenges:   make(map[string]*Challenge),
		games:        make(map[string]*Game),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		handleMessage: make(chan *MessageWrapper, 256),
	}
}

func (h *Hub) run() {
	// Challenge expiration ticker - runs every 1 second
	challengeTicker := time.NewTicker(1 * time.Second)
	defer challengeTicker.Stop()

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.handleConnect(client)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.handleDisconnect(client)
				delete(h.clients, client)
				close(client.send)
			}
		case wrapper := <-h.handleMessage:
			h.handleClientMessage(wrapper.client, wrapper.message)
		case <-challengeTicker.C:
			h.checkExpiredChallenges()
		}
	}
}

func (h *Hub) handleConnect(client *Client) {
	username := GenerateRandomName()
	userID := uuid.New().String()

	user := &User{
		ID:       userID,
		Username: username,
		Client:   client,
		InGame:   false,
	}
	client.user = user
	h.users[userID] = user

	// Send welcome message
	msg := Message{
		Type:     "welcome",
		UserID:   userID,
		Username: username,
	}
	h.sendToClient(client, &msg)

	// Broadcast updated user list
	h.broadcastUserList()

	log.Printf("User connected: %s (%s)", username, userID)
}

func (h *Hub) handleDisconnect(client *Client) {
	if client.user == nil {
		return
	}

	user := client.user
	log.Printf("User disconnected: %s (%s)", user.Username, user.ID)

	// Remove user from active games
	for gameID, game := range h.games {
		if (game.Player1 != nil && game.Player1.ID == user.ID) || (game.Player2 != nil && game.Player2.ID == user.ID) {
			// Notify opponent
			var opponent *User
			if game.Player1 != nil && game.Player1.ID == user.ID {
				opponent = game.Player2
			} else {
				opponent = game.Player1
			}

			if opponent != nil && !game.GameOver {
				opponent.InGame = false
				msg := Message{
					Type:   "opponent_disconnected",
					GameID: gameID,
				}
				h.sendToUser(opponent, &msg)
			}

			delete(h.games, gameID)
		}
	}

	// Remove pending challenges
	for challengeID, challenge := range h.challenges {
		if challenge.FromUser.ID == user.ID || challenge.ToUser.ID == user.ID {
			// Notify the other party if it's the recipient
			if challenge.FromUser.ID == user.ID && challenge.ToUser != nil {
				expireMsg := Message{
					Type:     "challenge_expired",
					ChallengeID: challengeID,
					Username: challenge.ToUser.Username,
				}
				h.sendToUser(challenge.ToUser, &expireMsg)
			}
			delete(h.challenges, challengeID)
		}
	}

	delete(h.users, user.ID)
	h.broadcastUserList()
}

func (h *Hub) handleClientMessage(client *Client, msg *Message) {
	switch msg.Type {
	case "challenge":
		h.handleChallenge(client.user, msg)
	case "accept_challenge":
		h.handleAcceptChallenge(client.user, msg)
	case "decline_challenge":
		h.handleDeclineChallenge(client.user, msg)
	case "submit_bid":
		h.handleSubmitBid(client.user, msg)
	case "rematch":
		h.handleRematch(client.user, msg)
	case "resign":
		h.handleResign(client.user, msg)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// Challenge handlers

func (h *Hub) handleChallenge(from *User, msg *Message) {
	to, exists := h.users[msg.TargetUserID]
	if !exists {
		log.Printf("Target user not found: %s", msg.TargetUserID)
		return
	}

	if to.InGame {
		h.sendError(from, "User is already in a game")
		return
	}

	// Check for existing pending challenges from this user to the target
	for _, c := range h.challenges {
		if c.FromUser.ID == from.ID && c.ToUser.ID == to.ID {
			h.sendError(from, "You already have a pending challenge to this user")
			return
		}
	}

	challengeID := uuid.New().String()
	challenge := &Challenge{
		ID:        challengeID,
		FromUser:  from,
		ToUser:    to,
		Timestamp: time.Now(),
	}
	h.challenges[challengeID] = challenge

	// Send challenge notification to target user
	challengeMsg := Message{
		Type:         "challenge_received",
		ChallengeID:  challengeID,
		FromUserID:   from.ID,
		FromUsername: from.Username,
	}
	h.sendToUser(to, &challengeMsg)

	log.Printf("Challenge created: %s -> %s", from.Username, to.Username)
}

func (h *Hub) handleAcceptChallenge(user *User, msg *Message) {
	challenge, exists := h.challenges[msg.ChallengeID]
	if !exists {
		log.Printf("Challenge not found: %s", msg.ChallengeID)
		return
	}

	if challenge.ToUser.ID != user.ID {
		log.Printf("User %s tried to accept challenge not meant for them", user.Username)
		return
	}

	// Create new game
	gameID := uuid.New().String()
	game := &Game{
		ID:             gameID,
		Player1:        challenge.FromUser,
		Player2:        challenge.ToUser,
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
	h.games[gameID] = game

	// Mark users as in game
	challenge.FromUser.InGame = true
	challenge.FromUser.GameID = gameID
	challenge.ToUser.InGame = true
	challenge.ToUser.GameID = gameID

	// Send game start to both players
	p1Msg := Message{
		Type:             "game_start",
		GameID:           gameID,
		OpponentID:       challenge.ToUser.ID,
		OpponentUsername: challenge.ToUser.Username,
		YourPlayer:       1,
	}
	h.sendToUser(challenge.FromUser, &p1Msg)

	p2Msg := Message{
		Type:             "game_start",
		GameID:           gameID,
		OpponentID:       challenge.FromUser.ID,
		OpponentUsername: challenge.FromUser.Username,
		YourPlayer:       2,
	}
	h.sendToUser(challenge.ToUser, &p2Msg)

	// Send initial waiting_for_bids state to both
	h.sendWaitingForBids(game)

	// Clean up challenge
	delete(h.challenges, msg.ChallengeID)

	// Broadcast updated user list
	h.broadcastUserList()

	log.Printf("Game started: %s vs %s (Game ID: %s)", challenge.FromUser.Username, challenge.ToUser.Username, gameID)
}

func (h *Hub) handleDeclineChallenge(user *User, msg *Message) {
	challenge, exists := h.challenges[msg.ChallengeID]
	if !exists {
		return
	}

	if challenge.ToUser.ID != user.ID {
		return
	}

	// Notify challenger
	declineMsg := Message{
		Type:        "challenge_declined",
		ChallengeID: msg.ChallengeID,
	}
	h.sendToUser(challenge.FromUser, &declineMsg)

	delete(h.challenges, msg.ChallengeID)
	log.Printf("Challenge declined: %s declined %s", user.Username, challenge.FromUser.Username)
}

func (h *Hub) checkExpiredChallenges() {
	now := time.Now()
	for challengeID, challenge := range h.challenges {
		if now.Sub(challenge.Timestamp) > CHALLENGE_EXPIRY*time.Second {
			// Notify the sender that their challenge expired
			expireMsg := Message{
				Type:        "challenge_expired",
				ChallengeID: challengeID,
				Username:    challenge.ToUser.Username,
			}
			h.sendToUser(challenge.FromUser, &expireMsg)

			delete(h.challenges, challengeID)
			log.Printf("Challenge expired: %s -> %s", challenge.FromUser.Username, challenge.ToUser.Username)
		}
	}
}

// Game logic

func (h *Hub) handleSubmitBid(user *User, msg *Message) {
	game, exists := h.games[msg.GameID]
	if !exists {
		return
	}

	// Determine player number
	var playerNum int
	if game.Player1.ID == user.ID {
		playerNum = 1
	} else if game.Player2.ID == user.ID {
		playerNum = 2
	} else {
		return
	}

	// Validate bid
	if msg.Bid < 0 {
		h.sendError(user, "Bid must be non-negative")
		return
	}

	// Get current balance
	var balance int
	if playerNum == 1 {
		balance = game.Player1Balance
	} else {
		balance = game.Player2Balance
	}

	if msg.Bid > balance {
		h.sendError(user, "Bid exceeds your balance")
		return
	}

	// Store bid
	if playerNum == 1 {
		bid := msg.Bid
		game.Player1Bid = &bid
	} else {
		bid := msg.Bid
		game.Player2Bid = &bid
	}

	log.Printf("Bid submitted in game %s: Player %d bid %d", game.ID, playerNum, msg.Bid)

	// Check if both bids are submitted
	if game.Player1Bid != nil && game.Player2Bid != nil {
		game.Status = "RESOLVING"
		h.resolveRound(game)
	}
}

func (h *Hub) resolveRound(game *Game) {
	p1Bid := *game.Player1Bid
	p2Bid := *game.Player2Bid

	// Deduction (both lose their bid regardless of outcome)
	game.Player1Balance -= p1Bid
	game.Player2Balance -= p2Bid

	// Movement determination
	var result string
	var p1NewPos = game.Player1Pos
	var p2NewPos = game.Player2Pos

	if p1Bid > p2Bid {
		p1NewPos++
		result = "P1_WINS_ROUND"
	} else if p2Bid > p1Bid {
		p2NewPos++
		result = "P2_WINS_ROUND"
	} else {
		result = "DRAW"
	}

	// Update positions
	game.Player1Pos = p1NewPos
	game.Player2Pos = p2NewPos

	// Record history
	history := RoundHistory{
		Turn:     game.CurrentRound,
		P1Bid:    p1Bid,
		P2Bid:    p2Bid,
		P1NewPos: p1NewPos,
		P2NewPos: p2NewPos,
		Result:   result,
	}
	game.History = append(game.History, history)

	// Send round result to both players
	resultMsg := Message{
		Type:        "round_result",
		GameID:      game.ID,
		Turn:        game.CurrentRound,
		P1Bid:       p1Bid,
		P2Bid:       p2Bid,
		P1Position:  p1NewPos,
		P2Position:  p2NewPos,
		P1Balance:   game.Player1Balance,
		P2Balance:   game.Player2Balance,
		Result:      result,
	}
	h.sendToUser(game.Player1, &resultMsg)
	h.sendToUser(game.Player2, &resultMsg)

	log.Printf("Round %d result: P1 bid %d, P2 bid %d, Result: %s, Positions: P1=%d, P2=%d",
		game.CurrentRound, p1Bid, p2Bid, result, p1NewPos, p2NewPos)

	// Check win condition
	winner, reason := h.checkWinCondition(game)
	if winner > 0 {
		game.GameOver = true
		game.Winner = winner
		game.EndTime = time.Now()
		game.Status = "GAME_OVER"

		endMsg := Message{
			Type:   "game_end",
			GameID: game.ID,
			Winner: winner,
			Reason: reason,
		}
		h.sendToUser(game.Player1, &endMsg)
		h.sendToUser(game.Player2, &endMsg)

		// Mark players as not in game
		game.Player1.InGame = false
		game.Player1.GameID = ""
		game.Player2.InGame = false
		game.Player2.GameID = ""

		// Broadcast updated user list
		h.broadcastUserList()

		// Remove game after a delay
		go func() {
			time.Sleep(10 * time.Second)
			delete(h.games, game.ID)
		}()

		log.Printf("Game %s ended: Winner=%d, Reason=%s", game.ID, winner, reason)
	} else {
		// Continue to next round
		game.CurrentRound++
		game.Player1Bid = nil
		game.Player2Bid = nil
		game.Status = "WAITING_FOR_BIDS"

		// Send waiting for bids state
		h.sendWaitingForBids(game)
	}
}

func (h *Hub) checkWinCondition(game *Game) (int, string) {
	// Check if either player reached MAX_STEPS
	if game.Player1Pos >= MAX_STEPS {
		return 1, "Reached final step"
	}
	if game.Player2Pos >= MAX_STEPS {
		return 2, "Reached final step"
	}

	// Check for bankruptcy stalemate
	if game.Player1Balance == 0 && game.Player2Balance == 0 {
		if game.Player1Pos > game.Player2Pos {
			return 1, "Bankruptcy stalemate - higher position wins"
		} else if game.Player2Pos > game.Player1Pos {
			return 2, "Bankruptcy stalemate - higher position wins"
		} else {
			return 3, "Bankruptcy stalemate - draw"
		}
	}

	// Check if both players are at position 0 with 0 balance (edge case)
	if game.Player1Pos == 0 && game.Player2Pos == 0 && game.Player1Balance == 0 && game.Player2Balance == 0 {
		return 3, "No moves possible - draw"
	}

	return 0, ""
}

func (h *Hub) sendWaitingForBids(game *Game) {
	msg := Message{
		Type:        "waiting_for_bids",
		GameID:      game.ID,
		Turn:        game.CurrentRound,
		P1Balance:   game.Player1Balance,
		P2Balance:   game.Player2Balance,
		P1Position:  game.Player1Pos,
		P2Position:  game.Player2Pos,
	}
	log.Printf("Sending waiting_for_bids to both players for game %s", game.ID)
	h.sendToUser(game.Player1, &msg)
	h.sendToUser(game.Player2, &msg)
}

func (h *Hub) handleRematch(user *User, msg *Message) {
	game, exists := h.games[msg.GameID]
	if !exists {
		return
	}

	var opponent *User
	if game.Player1.ID == user.ID {
		opponent = game.Player2
	} else if game.Player2.ID == user.ID {
		opponent = game.Player1
	} else {
		return
	}

	// Send rematch request to opponent
	rematchMsg := Message{
		Type:       "rematch_received",
		GameID:     msg.GameID,
		FromUserID: user.ID,
	}
	h.sendToUser(opponent, &rematchMsg)
}

func (h *Hub) handleResign(user *User, msg *Message) {
	game, exists := h.games[msg.GameID]
	if !exists {
		return
	}

	if game.GameOver {
		return
	}

	var opponent *User
	var winner int
	if game.Player1.ID == user.ID {
		opponent = game.Player2
		winner = 2
	} else if game.Player2.ID == user.ID {
		opponent = game.Player1
		winner = 1
	} else {
		return
	}

	// End game with opponent as winner
	game.GameOver = true
	game.Winner = winner
	game.EndTime = time.Now()
	game.Status = "GAME_OVER"

	endMsg := Message{
		Type:   "game_end",
		GameID: game.ID,
		Winner: winner,
		Reason: "Opponent resigned",
	}
	h.sendToUser(opponent, &endMsg)
	h.sendToUser(user, &endMsg)

	// Mark players as not in game
	game.Player1.InGame = false
	game.Player1.GameID = ""
	game.Player2.InGame = false
	game.Player2.GameID = ""

	// Broadcast updated user list
	h.broadcastUserList()

	// Remove game after a delay
	go func() {
		time.Sleep(10 * time.Second)
		delete(h.games, game.ID)
	}()
}

// Utility methods

func (h *Hub) sendToClient(client *Client, msg *Message) {
	data, _ := json.Marshal(msg)
	client.send <- data
}

func (h *Hub) sendToUser(user *User, msg *Message) {
	if user != nil && user.Client != nil {
		h.sendToClient(user.Client, msg)
	}
}

func (h *Hub) sendError(user *User, errorMsg string) {
	msg := Message{
		Type:     "error",
		Username: errorMsg,
	}
	h.sendToUser(user, &msg)
}

func (h *Hub) broadcastUserList() {
	users := make([]UserInfo, 0, len(h.users))
	for _, user := range h.users {
		users = append(users, UserInfo{
			UserID:   user.ID,
			Username: user.Username,
			InGame:   user.InGame,
		})
	}

	msg := Message{
		Type:  "users_update",
		Users: users,
	}

	for _, user := range h.users {
		h.sendToUser(user, &msg)
	}
}
