# Quo Vadis - Multiplayer Browser Game Implementation Plan

## Overview

Implementation of a 2-player strategy browser game "Quo Vadis" based on "All-pay auction" game theory. The game will feature:
- **Goal:** Be the first player to reach the final step on the bridge
- **Core Mechanic:** Simultaneous blind bidding
- **Key Constraint:** Bids are deducted from players' balances regardless of outcome
- **Multiplayer:** 1x1 over internet using WebSockets with challenge/accept system

## Architecture

### Copy from Virusgame

The backend architecture is directly copied from virusgame with minimal modifications:

```
backend/
├── main.go          # HTTP server, WebSocket endpoint, static files
├── hub.go           # Central event loop, game logic, message handlers
├── client.go        # WebSocket client, read/write pumps
├── types.go         # Data structures (Game, User, Challenge, Message)
└── names.go         # Random username generator
```

### Key Architectural Patterns

1. **Hub Pattern:** Single-threaded event loop handles all state mutations
2. **WebSocket Communication:** JSON messages between client and server
3. **Challenge System:** Users see online players and can send/accept challenges
4. **No Bots:** Simplified to human-only 1v1 matches

## Game State Structure

```go
type GameState struct {
    Turn           int              // Current turn number
    Status         string           // "WAITING_FOR_BIDS", "RESOLVING", "GAME_OVER"
    Player1        *PlayerState     // Human player
    Player2        *PlayerState     // Opponent
    History        []RoundHistory   // Log of previous rounds
    Winner         int              // 0 = none, 1 = player1, 2 = player2
    StartTime      time.Time
}

type PlayerState struct {
    ID        string  // User ID
    Username  string  // Display name
    Position  int     // 0 to MAX_STEPS (3)
    Balance   int     // INITIAL_BUDGET (20)
    CurrentBid *int   // nil until bid submitted
    IsConnected bool
}

type RoundHistory struct {
    Turn        int
    P1Bid       int
    P2Bid       int
    P1NewPos    int
    P2NewPos    int
    Result      string // "P1_WINS", "P2_WINS", "DRAW"
}
```

## Game Constants

```go
const (
    MAX_STEPS        = 3    // Target position to win
    INITIAL_BUDGET   = 20   // Starting points/stones
    MIN_BID          = 0
    MAX_BID          = 0    // 0 means no limit (can bid up to current balance)
    GAME_TIMEOUT     = 120  // seconds for move timeout
    CHALLENGE_EXPIRY = 60   // seconds
)
```

## Message Protocol

### Client → Server Messages

| Type | Purpose | Fields |
|------|---------|--------|
| `challenge` | Challenge another user | `targetUserId` |
| `accept_challenge` | Accept a challenge | `challengeId` |
| `decline_challenge` | Decline a challenge | `challengeId` |
| `submit_bid` | Submit bid for current round | `gameId`, `bid` (int) |
| `rematch` | Request rematch after game | `gameId` |
| `resign` | Resign from game | `gameId` |

### Server → Client Messages

| Type | Purpose | Fields |
|------|---------|--------|
| `welcome` | Initial connection | `userId`, `username` |
| `users_update` | Online users list | `users: [{userId, username, inGame}]` |
| `challenge_received` | Incoming challenge | `challengeId`, `fromUserId`, `fromUsername` |
| `challenge_declined` | Challenge declined | `challengeId` |
| `challenge_expired` | Challenge timed out | `challengeId`, `username` |
| `game_start` | Game begins | `gameId`, `opponentId`, `opponentUsername`, `yourPlayer` |
| `waiting_for_bids` | Bidding phase | `gameId`, `turn`, `p1Balance`, `p2Balance` |
| `bids_submitted` | Both bids in (internal notification) | `gameId` |
| `round_result` | Round resolution | `gameId`, `turn`, `p1Bid`, `p2Bid`, `p1NewPos`, `p2NewPos`, `result` |
| `game_end` | Game over | `gameId`, `winner`, `reason` |
| `opponent_disconnected` | Opponent left | `gameId` |
| `error` | Error message | `message` |

## Game Flow

```
┌─────────────┐
│  Connected  │ User connects, gets random username
└──────┬──────┘
       │ users_update
       ▼
┌─────────────┐
│   Lobby     │ See online users, send challenges
└──────┬──────┘
       │ accept_challenge
       ▼
┌─────────────┐
│  Game Prep  │ Both players sent game_start
└──────┬──────┘
       │ start bidding
       ▼
┌─────────────┐     ┌─────────────┐
│  Bidding    │◀───▶│  Both bids  │
│  Phase      │     │  submitted  │
└──────┬──────┘     └──────┬──────┘
       │ round_result      │
       ▼                   │
┌─────────────┐            │
│ Check Win   │            │
│ Condition   │            │
└──────┬──────┘            │
       │ game continues    │
       │                   │
       │ game_over         │
       ▼                   │
┌─────────────┐            │
│  Game Over  │◀───────────┘
│  (winner)   │
└──────┬──────┘
       │ rematch / leave
       ▼
┌─────────────┐
│  Back to    │
│  Lobby      │
└─────────────┘
```

## Bidding Logic (All-Pay Auction)

### Phase A: Bidding
1. Both players independently submit a bid (0 ≤ bid ≤ current_balance)
2. Bids are hidden until both are submitted

### Phase B: Resolution
```go
// Deduction (CRITICAL: Both lose their bid regardless of outcome)
p1.balance -= p1.bid
p2.balance -= p2.bid

// Movement
if p1.bid > p2.bid:
    p1.position += 1
    result = "P1_WINS_ROUND"
elif p2.bid > p1.bid:
    p2.position += 1
    result = "P2_WINS_ROUND"
else:  // tie
    // No movement
    result = "DRAW"
```

### Phase C: Win Condition
```go
if p1.position >= MAX_STEPS:
    winner = 1
elif p2.position >= MAX_STEPS:
    winner = 2
elif p1.balance == 0 && p2.balance == 0:
    // Bankruptcy stalemate - player with higher position wins
    if p1.position > p2.position:
        winner = 1
    elif p2.position > p1.position:
        winner = 2
    else:
        winner = 0  // Draw
```

## Frontend Components

### UI Structure (index.html)
```html
<div class="main-container">
    <div id="sidebar">
        <!-- Connection status -->
        <div id="connection-status"></div>
        <div id="welcome-message"></div>

        <!-- Game Settings -->
        <div class="sidebar-section">
            <h3>Game Settings</h3>
            <button id="new-game-button">New Local Game (PvP)</button>
        </div>

        <!-- Online Players -->
        <div class="sidebar-section users-section">
            <h3>Online Players</h3>
            <div id="users-list"></div>
        </div>
    </div>

    <div id="main-content">
        <div id="game-container">
            <h1>Quo Vadis</h1>

            <!-- Bridge/Track Visualization -->
            <div id="bridge">
                <!-- 4 slots: Start, Step 1, Step 2, Finish -->
                <div class="bridge-slot" data-pos="0">Start</div>
                <div class="bridge-slot" data-pos="1">Step 1</div>
                <div class="bridge-slot" data-pos="2">Step 2</div>
                <div class="bridge-slot" data-pos="3">Finish</div>
            </div>

            <!-- Player Avatars -->
            <div id="players-display">
                <div id="p1-display" class="player-display">
                    <div class="player-avatar"></div>
                    <div class="player-info">
                        <span class="player-name">Player</span>
                        <span class="player-balance">Balance: 20</span>
                    </div>
                </div>
                <div id="p2-display" class="player-display">
                    <div class="player-avatar"></div>
                    <div class="player-info">
                        <span class="player-name">Opponent</span>
                        <span class="player-balance">Balance: 20</span>
                    </div>
                </div>
            </div>

            <!-- Bidding Controls -->
            <div id="bidding-controls" style="display: none;">
                <div class="bid-input-group">
                    <label>Your Bid:</label>
                    <input type="number" id="bid-input" min="0" max="20">
                    <button id="submit-bid-button">Submit Bid</button>
                </div>
                <div id="bidding-status">Waiting for opponent...</div>
            </div>

            <!-- Game Log -->
            <div id="game-log"></div>

            <!-- Game Controls -->
            <div id="game-controls">
                <button id="resign-button" style="display: none;">Resign</button>
                <button id="rematch-button" style="display: none;">Request Rematch</button>
            </div>
        </div>
    </div>
</div>
```

### Frontend Classes (script.js)

1. **GameState Management**
   - Track current game state, player positions, balances
   - Handle state transitions (bidding → resolving → game_over)

2. **Bridge Renderer**
   - Render 4 slots with player avatars
   - Animate movement between positions
   - Show reveal animation after both bids submitted

3. **Bidding UI**
   - Input validation (0 ≤ bid ≤ balance)
   - Submit button with loading state
   - Show opponent's bid after reveal

4. **Game History**
   - Log each round with bids and results
   - Allow reviewing past rounds

### WebSocket Client (multiplayer.js)

```javascript
class MultiplayerClient {
    connect() { /* ... */ }
    handleMessage(msg) {
        switch(msg.type) {
            case 'welcome': this.handleWelcome(msg); break;
            case 'users_update': this.handleUsersUpdate(msg); break;
            case 'challenge_received': this.handleChallengeReceived(msg); break;
            case 'game_start': this.handleGameStart(msg); break;
            case 'waiting_for_bids': this.handleWaitingForBids(msg); break;
            case 'round_result': this.handleRoundResult(msg); break;
            case 'game_end': this.handleGameEnd(msg); break;
        }
    }
    challengeUser(userId) { /* ... */ }
    acceptChallenge(challengeId) { /* ... */ }
    submitBid(gameId, bid) { /* ... */ }
}
```

## CSS Styling (style.css)

- **Bridge/Track:** Horizontal or vertical track with 4 positions
- **Player Avatars:** Distinct colors for P1 and P2
- **Bidding Reveal:** Animation showing both bids simultaneously
- **Game Log:** Scrollable history of rounds
- **Responsive:** Work on desktop and mobile

## File Structure

```
quevadis/
├── backend/
│   ├── main.go          # HTTP server, WebSocket endpoint
│   ├── hub.go           # Central event loop and game logic
│   ├── client.go        # WebSocket client management
│   ├── types.go         # Data structures and message types
│   ├── names.go         # Random username generator
│   ├── go.mod
│   └── go.sum
├── index.html           # Main HTML structure
├── style.css            # All styling
├── script.js            # Frontend game logic
├── multiplayer.js       # WebSocket client
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## Implementation Steps

### Step 1: Backend Foundation
1. Copy backend files from virusgame
2. Update types.go with Quo Vadis-specific structures
3. Update hub.go message handlers for new game logic
4. Implement bidding and resolution logic

### Step 2: Frontend Structure
1. Create index.html with bridge UI
2. Copy multiplayer.js and adapt for Quo Vadis
3. Create script.js with game state management

### Step 3: Game Logic Integration
1. Connect frontend to WebSocket
2. Implement challenge/accept flow
3. Implement bidding UI and submission
4. Implement round result display with reveal animation
5. Implement win condition checks

### Step 4: Polish
1. Add CSS styling for bridge and players
2. Add game log/history
3. Add rematch functionality
4. Test with multiple browser windows

## Testing

1. **Local Testing:** Open multiple browser windows, challenge between them
2. **Edge Cases:**
   - Both players bid 0 (no movement, both lose 0)
   - One player bids all-in
   - Bankruptcy stalemate (both at 0 balance)
   - Disconnection handling
   - Challenge expiration

## Docker Configuration

Copy docker-compose.yml and Dockerfile from virusgame with minimal changes:
- Update service name to `quevadis`
- Update exposed port if needed
- Update static files path

## Estimated Complexity

- **Backend:** ~60% copied from virusgame, ~40% new game logic
- **Frontend:** ~30% copied from virusgame (multiplayer.js), ~70% new UI/game logic
- **Total New Code:** ~500-700 lines (backend + frontend)
