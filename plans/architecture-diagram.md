# Quo Vadis - Architecture Diagrams

## System Architecture

```mermaid
graph TB
    subgraph Client Browser
        UI[HTML/CSS UI]
        GC[Game Controller]
        MC[Multiplayer Client]
    end

    subgraph Go Backend
        HTTP[HTTP Server :8080]
        WS[WebSocket Handler]
        Hub[Hub Event Loop]
        GL[Game Logic]
    end

    UI --> GC
    GC --> MC
    MC -- JSON/WebSocket --> WS
    WS --> Hub
    Hub --> GL

    subgraph Data Structures
        Users[Users Map]
        Challenges[Challenges Map]
        Games[Games Map]
    end

    Hub --> Users
    Hub --> Challenges
    Hub --> Games
```

## Game State Machine

```mermaid
stateDiagram-v2
    [*] --> Connected: User connects

    Connected --> Lobby: Get welcome message
    Lobby --> Challenging: Send challenge
    Lobby --> Challenged: Receive challenge

    Challenging --> Lobby: Challenge declined/expired
    Challenged --> Lobby: Decline challenge
    Challenged --> GamePrep: Accept challenge

    GamePrep --> Bidding: Both players ready
    Bidding --> Bidding: Player submits bid
    Bidding --> Resolving: Both bids submitted

    Resolving --> RoundResult: Calculate outcome
    RoundResult --> CheckWin: Update positions/balances

    CheckWin --> GameOver: Winner determined
    CheckWin --> Bidding: Next round

    GameOver --> Lobby: Game ended
    Lobby --> [*]
```

## Bidding & Resolution Flow

```mermaid
sequenceDiagram
    participant P1 as Player 1
    participant H as Hub
    participant P2 as Player 2

    Note over P1,P2: Bidding Phase
    P1->>H: submit_bid (bid=5)
    P2->>H: submit_bid (bid=3)
    Note over H: Both bids received

    Note over H: Resolution Phase
    H->>H: p1.balance = 20-5 = 15
    H->>H: p2.balance = 20-3 = 17
    H->>H: p1.position = 1 (5 > 3)

    H->>P1: round_result (p1Bid=5, p2Bid=3, p1Wins)
    H->>P2: round_result (p1Bid=5, p2Bid=3, p1Wins)

    Note over P1,P2: Check Win Condition
    alt p1.position >= 3
        H->>P1: game_end (winner=1)
        H->>P2: game_end (winner=1)
    else Continue
        Note over P1,P2: Next Round Bidding
    end
```

## Frontend Component Structure

```mermaid
graph TD
    App[Main App Container]
    Sidebar[Sidebar]
    GameArea[Game Area]

    Sidebar --> Status[Connection Status]
    Sidebar --> Users[Online Users List]
    Sidebar --> Settings[Game Settings]

    GameArea --> Bridge[Bridge/Track]
    Bridge --> Slots[4 Slots: Start-Step1-Step2-Finish]

    GameArea --> Players[Player Displays]
    Players --> P1[P1 Avatar + Balance]
    Players --> P2[P2 Avatar + Balance]

    GameArea --> Bidding[Bidding Controls]
    Bidding --> Input[Bid Number Input]
    Bidding --> Submit[Submit Button]

    GameArea --> Log[Game Log]
    GameArea --> Controls[Game Controls]
    Controls --> Resign[Resign Button]
    Controls --> Rematch[Rematch Button]
```

## Data Flow During Game

```mermaid
flowchart LR
    subgraph Server State
        G[Game State]
        P1S[P1 State]
        P2S[P2 State]
    end

    subgraph Client State
        LC[Local Client State]
        Remote[Remote State from Server]
    end

    G -->|game_start| LC
    G -->|game_start| Remote

    LC -->|submit_bid| G
    Remote -->|submit_bid| G

    G -->|waiting_for_bids| LC
    G -->|waiting_for_bids| Remote

    G -->|round_result| LC
    G -->|round_result| Remote

    LC -->|render| UI[Bridge UI]
    Remote -->|render| UI
```

## Message Flow Summary

```mermaid
flowchart TB
    subgraph Client Messages
        C1[challenge]
        C2[accept_challenge]
        C3[submit_bid]
        C4[resign]
        C5[rematch]
    end

    subgraph Server Messages
        S1[users_update]
        S2[challenge_received]
        S3[game_start]
        S4[waiting_for_bids]
        S5[round_result]
        S6[game_end]
    end

    C1 --> S2
    C2 --> S3
    C3 --> S4 --> S5
    C3 --> S5
    C4 --> S6
```

## File Dependencies

```
backend/
├── main.go
│   └── imports: hub, client, names, storage
├── hub.go
│   └── imports: types, names
├── client.go
│   └── imports: hub, types
├── types.go
│   └── imports: encoding/json, time
└── names.go
    └── imports: math/rand, strings

Frontend/
├── index.html
│   └── loads: style.css, script.js, multiplayer.js
├── style.css
├── script.js
│   └── depends on: multiplayer.js
└── multiplayer.js
    └── depends on: script.js (game state)
```
