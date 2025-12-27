// Multiplayer WebSocket client for Quo Vadis

class MultiplayerClient {
    constructor() {
        this.ws = null;
        this.userId = null;
        this.username = null;
        this.gameId = null;
        this.yourPlayer = null;
        this.opponentId = null;
        this.opponentUsername = null;
        this.onlineUsers = [];
        this.pendingChallenges = new Map();
        this.connected = false;
    }

    connect() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;

        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
            console.log('Connected to multiplayer server');
            this.connected = true;
            this.updateConnectionStatus(true);
        };

        this.ws.onmessage = (event) => {
            try {
                const msg = JSON.parse(event.data);
                this.handleMessage(msg);
            } catch (error) {
                console.error('Error parsing message:', error);
            }
        };

        this.ws.onclose = () => {
            console.log('Disconnected from multiplayer server');
            this.connected = false;
            this.updateConnectionStatus(false);
            // Attempt to reconnect after 3 seconds
            setTimeout(() => this.connect(), 3000);
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }

    disconnect() {
        if (this.ws) {
            this.ws.close();
        }
    }

    send(message) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(message));
        }
    }

    handleMessage(msg) {
        console.log('Received message:', msg);

        switch (msg.type) {
            case 'welcome':
                this.handleWelcome(msg);
                break;
            case 'users_update':
                this.handleUsersUpdate(msg);
                break;
            case 'challenge_received':
                this.handleChallengeReceived(msg);
                break;
            case 'challenge_declined':
                this.handleChallengeDeclined(msg);
                break;
            case 'challenge_expired':
                this.handleChallengeExpired(msg);
                break;
            case 'game_start':
                this.handleGameStart(msg);
                break;
            case 'waiting_for_bids':
                this.handleWaitingForBids(msg);
                break;
            case 'round_result':
                this.handleRoundResult(msg);
                break;
            case 'game_end':
                this.handleGameEnd(msg);
                break;
            case 'opponent_disconnected':
                this.handleOpponentDisconnected(msg);
                break;
            case 'rematch_received':
                this.handleRematchReceived(msg);
                break;
            case 'error':
                this.handleError(msg);
                break;
        }
    }

    handleWelcome(msg) {
        this.userId = msg.userId;
        this.username = msg.username;
        console.log(`Welcome! You are ${this.username} (${this.userId})`);
        document.getElementById('welcome-message').textContent = `You are: ${this.username}`;
    }

    handleUsersUpdate(msg) {
        this.onlineUsers = msg.users.filter(u => u.userId !== this.userId);
        this.updateUsersList();
    }

    handleChallengeReceived(msg) {
        this.pendingChallenges.set(msg.challengeId, {
            fromUserId: msg.fromUserId,
            fromUsername: msg.fromUsername,
        });
        this.showChallengeNotification(msg);
    }

    handleChallengeDeclined(msg) {
        showNotification(`${this.username} declined your challenge`, 'info');
    }

    handleChallengeExpired(msg) {
        // Find notification by challenge ID and remove it
        const notification = document.querySelector(`.challenge-notification[data-challenge-id="${msg.challengeId}"]`);
        if (notification) {
            notification.remove();
        }

        if (msg.username) {
            showNotification(`Challenge to ${msg.username} expired`, 'info');
        }

        this.pendingChallenges.delete(msg.challengeId);
    }

    handleGameStart(msg) {
        this.gameId = msg.gameId;
        this.yourPlayer = msg.yourPlayer;
        this.opponentId = msg.opponentId;
        this.opponentUsername = msg.opponentUsername;

        // Update game state
        gameState.gameId = msg.gameId;
        gameState.yourPlayer = msg.yourPlayer;
        gameState.opponentId = msg.opponentId;
        gameState.opponentUsername = msg.opponentUsername;
        gameState.gameOver = false;
        gameState.turn = 1;
        gameState.p1Balance = 20;
        gameState.p2Balance = 20;
        gameState.p1Position = 0;
        gameState.p2Position = 0;
        gameState.waitingForBid = true;
        gameState.yourBidSubmitted = false;

        // Update UI
        const p1NameEl = document.getElementById('p1-name');
        const p2NameEl = document.getElementById('p2-name');

        if (msg.yourPlayer === 1) {
            p1NameEl.textContent = this.username;
            p2NameEl.textContent = msg.opponentUsername;
        } else {
            p1NameEl.textContent = msg.opponentUsername;
            p2NameEl.textContent = this.username;
        }

        // Show resign button
        document.getElementById('resign-button').style.display = 'inline-block';
        document.getElementById('rematch-button').style.display = 'none';

        // Clear log and add entry
        document.getElementById('log-entries').innerHTML = '';
        addLogEntry(`Game started! You are ${msg.yourPlayer === 1 ? 'X' : 'O'}`);

        // Update UI
        updateUI();

        showNotification(`Game started against ${msg.opponentUsername}!`, 'success');
    }

    handleWaitingForBids(msg) {
        gameState.turn = msg.turn;
        gameState.p1Balance = msg.p1Balance;
        gameState.p2Balance = msg.p2Balance;
        gameState.p1Position = msg.p1Position;
        gameState.p2Position = msg.p2Position;
        gameState.currentPlayer = msg.currentPlayer || 1;
        gameState.yourBidSubmitted = false;
        gameState.waitingForBid = true;

        updateUI();
    }

    handleRoundResult(msg) {
        // Update game state
        gameState.turn = msg.turn;
        gameState.p1Balance = msg.p1Balance;
        gameState.p2Balance = msg.p2Balance;
        gameState.p1Position = msg.p1Position;
        gameState.p2Position = msg.p2Position;

        // Determine result text
        let resultText;
        if (msg.result === 'P1_WINS_ROUND') {
            resultText = gameState.yourPlayer === 1 ? 'You win the round!' : 'Opponent wins the round!';
        } else if (msg.result === 'P2_WINS_ROUND') {
            resultText = gameState.yourPlayer === 2 ? 'You win the round!' : 'Opponent wins the round!';
        } else {
            resultText = 'Draw - no movement!';
        }

        // Log the round
        const p1Bid = msg.p1Bid;
        const p2Bid = msg.p2Bid;
        addLogEntry(`Round ${msg.turn}: X bid ${p1Bid}, O bid ${p2Bid}. ${resultText}`);

        // Show reveal animation
        showBidReveal(msg.turn, msg.p1Bid, msg.p2Bid, resultText);
    }

    handleGameEnd(msg) {
        gameState.gameOver = true;
        gameState.gameId = null;

        document.getElementById('resign-button').style.display = 'none';
        document.getElementById('rematch-button').style.display = 'inline-block';

        let winnerText;
        if (msg.winner === gameState.yourPlayer) {
            winnerText = 'You win!';
            addLogEntry(msg.reason, 'win');
        } else if (msg.winner === 3) {
            winnerText = 'Draw!';
            addLogEntry(msg.reason, 'draw');
        } else {
            winnerText = 'You lose!';
            addLogEntry(msg.reason, 'lose');
        }

        document.getElementById('status').textContent = `Game Over! ${winnerText}`;
        document.getElementById('status').classList.remove('your-turn');

        showNotification(`Game ended: ${winnerText}`, msg.winner === gameState.yourPlayer ? 'success' : 'info');
    }

    handleOpponentDisconnected(msg) {
        showNotification('Opponent disconnected', 'error');
        this.endMultiplayerGame();
    }

    handleRematchReceived(msg) {
        showNotification(`${this.opponentUsername} wants a rematch!`, 'info');
        // Auto-accept rematch for now
        this.acceptRematch(msg.gameId);
    }

    handleError(msg) {
        showNotification(msg.username || 'An error occurred', 'error');
    }

    // Challenge methods
    challengeUser(userId) {
        this.send({
            type: 'challenge',
            targetUserId: userId,
        });
        showNotification('Challenge sent!', 'info');
    }

    acceptChallenge(challengeId) {
        // Remove notification
        const notification = document.querySelector(`.notification.challenge[data-challenge-id="${challengeId}"]`);
        if (notification) {
            notification.remove();
        }

        this.send({
            type: 'accept_challenge',
            challengeId: challengeId,
        });
        this.pendingChallenges.delete(challengeId);
    }

    declineChallenge(challengeId) {
        // Remove notification
        const notification = document.querySelector(`.notification.challenge[data-challenge-id="${challengeId}"]`);
        if (notification) {
            notification.remove();
        }

        this.send({
            type: 'decline_challenge',
            challengeId: challengeId,
        });
        this.pendingChallenges.delete(challengeId);
    }

    // Game methods
    submitBid(gameId, bid) {
        this.send({
            type: 'submit_bid',
            gameId: gameId,
            bid: bid,
        });
        gameState.yourBidSubmitted = true;
        updateUI();
    }

    resign(gameId) {
        this.send({
            type: 'resign',
            gameId: gameId,
        });
    }

    rematch(gameId) {
        this.send({
            type: 'rematch',
            gameId: gameId,
        });
        showNotification('Rematch requested!', 'info');
    }

    acceptRematch(gameId) {
        // In a real implementation, this would send a message to accept the rematch
        // For now, just start a new game
        this.send({
            type: 'accept_challenge',
            challengeId: 'rematch-' + gameId,
        });
    }

    // UI update methods
    updateConnectionStatus(connected) {
        const statusEl = document.getElementById('connection-status');
        if (connected) {
            statusEl.textContent = 'Connected';
            statusEl.classList.remove('disconnected');
            statusEl.classList.add('connected');
        } else {
            statusEl.textContent = 'Disconnected - Reconnecting...';
            statusEl.classList.remove('connected');
            statusEl.classList.add('disconnected');
        }
    }

    updateUsersList() {
        const usersList = document.getElementById('users-list');

        if (this.onlineUsers.length === 0) {
            usersList.innerHTML = '<div class="no-users">No other players online</div>';
            return;
        }

        usersList.innerHTML = this.onlineUsers.map(user => `
            <div class="user-item ${user.inGame ? 'in-game' : ''}" data-user-id="${user.userId}">
                <span class="user-name">${user.username}</span>
                ${!user.inGame ? `<button class="challenge-btn" onclick="mpClient.challengeUser('${user.userId}')">Challenge</button>` : '<span style="color: #888; font-size: 12px;">In Game</span>'}
            </div>
        `).join('');
    }

    showChallengeNotification(msg) {
        const container = document.getElementById('notifications');
        const notification = document.createElement('div');
        notification.className = 'notification challenge';
        notification.dataset.challengeId = msg.challengeId;
        notification.innerHTML = `
            <div><strong>${msg.fromUsername}</strong> challenges you!</div>
            <div class="challenge-actions">
                <button class="accept-btn" onclick="mpClient.acceptChallenge('${msg.challengeId}')">Accept</button>
                <button class="decline-btn" onclick="mpClient.declineChallenge('${msg.challengeId}')">Decline</button>
            </div>
        `;
        container.appendChild(notification);
    }

    endMultiplayerGame() {
        this.gameId = null;
        this.yourPlayer = null;
        this.opponentId = null;
        this.opponentUsername = null;

        gameState.gameId = null;
        gameState.yourPlayer = null;
        gameState.gameOver = true;

        document.getElementById('resign-button').style.display = 'none';
        document.getElementById('rematch-button').style.display = 'none';
    }
}

// Initialize multiplayer client
let mpClient;

document.addEventListener('DOMContentLoaded', () => {
    mpClient = new MultiplayerClient();
});
