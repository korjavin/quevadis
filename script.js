// Game Constants
const MAX_STEPS = 3;
const INITIAL_BUDGET = 20;

// Game State
let gameState = {
    gameId: null,
    yourPlayer: null, // 1 or 2
    opponentUsername: null,
    p1Balance: INITIAL_BUDGET,
    p2Balance: INITIAL_BUDGET,
    p1Position: 0,
    p2Position: 0,
    turn: 1,
    gameOver: false,
    yourBidSubmitted: false
};

// DOM Elements
let statusDisplay;
let bidInput;
let submitBidButton;
let biddingControls;
let resignButton;
let rematchButton;
let logEntries;
let p1Name, p2Name;
let p1Balance, p2Balance;
let p1Position, p2Position;

// Initialize game
function initGame() {
    // Get DOM elements
    statusDisplay = document.getElementById('status');
    bidInput = document.getElementById('bid-input');
    submitBidButton = document.getElementById('submit-bid-button');
    biddingControls = document.getElementById('bidding-controls');
    resignButton = document.getElementById('resign-button');
    rematchButton = document.getElementById('rematch-button');
    logEntries = document.getElementById('log-entries');
    p1Name = document.getElementById('p1-name');
    p2Name = document.getElementById('p2-name');
    p1Balance = document.getElementById('p1-balance');
    p2Balance = document.getElementById('p2-balance');
    p1Position = document.getElementById('p1-position');
    p2Position = document.getElementById('p2-position');

    // Event listeners
    submitBidButton.addEventListener('click', submitBid);
    resignButton.addEventListener('click', resign);
    rematchButton.addEventListener('click', requestRematch);

    // Reset game state
    resetGameState();
}

// Reset game state
function resetGameState() {
    gameState = {
        gameId: null,
        yourPlayer: null,
        opponentUsername: null,
        p1Balance: INITIAL_BUDGET,
        p2Balance: INITIAL_BUDGET,
        p1Position: 0,
        p2Position: 0,
        turn: 1,
        gameOver: false,
        yourBidSubmitted: false
    };

    // Reset UI
    biddingControls.style.display = 'none';
    resignButton.style.display = 'none';
    rematchButton.style.display = 'none';
    logEntries.innerHTML = '';
    updateBridgeMarkers();
    updatePlayerDisplays();
    updateStatus();
}

// Update UI based on game state
function updateUI() {
    updatePlayerDisplays();
    updateBridgeMarkers();
    updateStatus();
    updateBiddingControls();
}

// Update player displays
function updatePlayerDisplays() {
    p1Balance.textContent = gameState.p1Balance;
    p2Balance.textContent = gameState.p2Balance;
    p1Position.textContent = gameState.p1Position;
    p2Position.textContent = gameState.p2Position;
}

// Update bridge markers
function updateBridgeMarkers() {
    // Hide all markers first
    document.querySelectorAll('.player-marker').forEach(marker => {
        marker.classList.remove('active');
        marker.textContent = '';
    });

    // Show markers at current positions
    const p1Marker = document.getElementById(`p1-pos-${gameState.p1Position}`);
    const p2Marker = document.getElementById(`p2-pos-${gameState.p2Position}`);

    if (p1Marker) {
        p1Marker.classList.add('active');
        p1Marker.textContent = 'X';
    }
    if (p2Marker) {
        p2Marker.classList.add('active');
        p2Marker.textContent = 'O';
    }
}

// Update status display
function updateStatus() {
    if (gameState.gameOver) {
        statusDisplay.textContent = 'Game Over!';
        statusDisplay.classList.remove('your-turn');
        return;
    }

    if (!gameState.gameId) {
        statusDisplay.textContent = 'Connect to start playing!';
        return;
    }

    // Both players bid simultaneously in all-pay auction
    if (gameState.yourBidSubmitted) {
        statusDisplay.textContent = 'Bid submitted! Waiting for opponent...';
        statusDisplay.classList.remove('your-turn');
    } else {
        statusDisplay.textContent = 'Your turn! Place your bid.';
        statusDisplay.classList.add('your-turn');
    }
}

// Update bidding controls
function updateBiddingControls() {
    if (gameState.gameOver || !gameState.gameId) {
        biddingControls.style.display = 'none';
        return;
    }

    // Both players bid simultaneously - show controls if you haven't bid yet
    if (!gameState.yourBidSubmitted) {
        biddingControls.style.display = 'block';
        document.getElementById('current-turn').textContent = `Round ${gameState.turn}`;

        // Set max bid to current balance
        const balance = gameState.yourPlayer === 1 ? gameState.p1Balance : gameState.p2Balance;
        bidInput.max = balance;
        bidInput.value = Math.min(1, balance);

        submitBidButton.disabled = false;
        document.getElementById('bidding-status').textContent = 'Enter your bid and click Submit.';
    } else {
        biddingControls.style.display = 'block';
        submitBidButton.disabled = true;
        document.getElementById('bidding-status').textContent = 'Bid submitted! Waiting for opponent...';
    }
}

// Submit bid
function submitBid() {
    const bid = parseInt(bidInput.value) || 0;
    const balance = gameState.yourPlayer === 1 ? gameState.p1Balance : gameState.p2Balance;

    if (bid < 0 || bid > balance) {
        showNotification('Invalid bid! Bid must be between 0 and your balance.', 'error');
        return;
    }

    gameState.yourBidSubmitted = true;

    // Send bid to server
    if (typeof mpClient !== 'undefined') {
        mpClient.submitBid(gameState.gameId, bid);
    }
}

// Show bid reveal animation
function showBidReveal(turn, p1Bid, p2Bid, result) {
    const revealEl = document.getElementById('bid-reveal');
    document.getElementById('reveal-round').textContent = turn;
    document.getElementById('reveal-p1-bid').textContent = p1Bid;
    document.getElementById('reveal-p2-bid').textContent = p2Bid;
    document.getElementById('reveal-result').textContent = result;

    revealEl.style.display = 'flex';

    setTimeout(() => {
        revealEl.style.display = 'none';
        updateUI();
    }, 2000);
}

// End game
function endGame(winner, reason) {
    gameState.gameOver = true;

    let winnerText;
    if (winner === gameState.yourPlayer) {
        winnerText = 'You win!';
        addLogEntry(reason, 'win');
    } else if (winner === 3) {
        winnerText = 'Draw!';
        addLogEntry(reason, 'draw');
    } else {
        winnerText = 'You lose!';
        addLogEntry(reason, 'lose');
    }

    statusDisplay.textContent = `Game Over! ${winnerText}`;
    statusDisplay.classList.remove('your-turn');

    // Show controls
    resignButton.style.display = 'none';
    rematchButton.style.display = 'inline-block';
}

// Resign
function resign() {
    if (typeof mpClient !== 'undefined') {
        mpClient.resign(gameState.gameId);
    }
}

// Request rematch
function requestRematch() {
    if (typeof mpClient !== 'undefined' && gameState.gameId) {
        mpClient.rematch(gameState.gameId);
    }
}

// Add log entry
function addLogEntry(text, type = '') {
    const entry = document.createElement('div');
    entry.className = `log-entry ${type}`;
    entry.textContent = text;
    logEntries.insertBefore(entry, logEntries.firstChild);
}

// Show notification
function showNotification(message, type = 'info') {
    const container = document.getElementById('notifications');
    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    notification.textContent = message;
    container.appendChild(notification);

    setTimeout(() => {
        notification.style.animation = 'slideIn 0.3s ease reverse';
        setTimeout(() => notification.remove(), 300);
    }, 5000);
}

// Make functions available globally for multiplayer.js
window.initGame = initGame;
window.submitBid = submitBid;
window.resign = resign;
window.requestRematch = requestRematch;
window.showNotification = showNotification;
