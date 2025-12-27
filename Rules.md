# Game Specification: Quo Vadis (All-Pay Auction Mechanic)

## 1. Game Overview
A two-player strategy browser game based on the "All-pay auction" game theory. 
- **Goal:** Be the first player to reach the final step on the bridge.
- **Core Mechanic:** Simultaneous blind bidding.
- **Key Constraint:** Bids are deducted from players' balances regardless of the outcome (win, lose, or draw).

## 2. Configuration & Constants
- `MAX_STEPS`: 3 (The target position to win. Positions range from 0 to 3).
- `INITIAL_BUDGET`: 20 (Each player starts with 20 points/stones).
- `PLAYER_1_ID`: "Player" (Human).
- `PLAYER_2_ID`: "Opponent" (CPU or 2nd Human).

## 3. Game State Data Structure
The application state must track:
```json
{
  "turn": 1,
  "status": "WAITING_FOR_BIDS", // or "RESOLVING", "GAME_OVER"
  "players": {
    "p1": { "position": 0, "balance": 20, "currentBid": null },
    "p2": { "position": 0, "balance": 20, "currentBid": null }
  },
  "history": [] // Logs of previous turns
}
4. Game Loop Logic
Phase A: Bidding (Input)
Both players select a bid amount (b1 and b2).

Validation: Bid must be an integer, 0 <= bid <= current_balance.

Inputs are hidden (blind) until both have submitted.

Phase B: Resolution (The Core Logic)
Once both bids are submitted:

Deduction:

p1.balance = p1.balance - b1

p2.balance = p2.balance - b2 (CRITICAL: Points are lost even if the player loses the round)

Movement Determination:

IF b1 > b2:

Player 1 moves forward (p1.position += 1).

Player 2 stays.

Result: "Player 1 Wins Round".

IF b2 > b1:

Player 2 moves forward (p2.position += 1).

Player 1 stays.

Result: "Player 2 Wins Round".

IF b1 == b2:

No one moves.

Result: "Draw/Deadlock".

Phase C: Win Condition Check
After movement:

IF p1.position >= MAX_STEPS: Player 1 Wins Game.

IF p2.position >= MAX_STEPS: Player 2 Wins Game.

Bankruptcy Stalemate (Edge Case): - IF both players have balance == 0 AND neither reached MAX_STEPS:

The player with the higher position wins.

If positions are equal, it is a Draw.

5. UI/UX Requirements
Visuals: - A simple track/bridge with 4 slots (Start, Step 1, Step 2, Finish).

Avatars for P1 and P2 showing their current step.

HUD:

Display remaining Balance clearly for both.

Input field for Bid (number).

Feedback:

Show a "Reveal" animation or log after bids are made.

Example Log: "Player bid 5, CPU bid 3. Player moves! Both spent their points."

6. AI Opponent Logic (Basic)
If playing against CPU, the bot should:

Randomly bid between 0 and min(cpu_balance, 5).

Occasionally go "All-in" if close to winning.