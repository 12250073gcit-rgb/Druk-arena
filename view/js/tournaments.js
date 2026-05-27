// tournaments.js — DrukArena Tournament Management
// Following CSC103 v1.3 fetch patterns with then/catch

// ── Build Tournament Card HTML ─────────────────────────────

function buildTournamentCard(t, showActions = false) {
  const statusClass = t.status === 'ongoing' ? 'ongoing' : t.status === 'completed' ? 'completed' : '';
  const isAdmin = window._currentUser && window._currentUser.role === 'admin';
  const isClosed = t.status === 'ongoing' || t.status === 'completed' || (t.max_teams && t.participant_count >= t.max_teams);
  const joinLabel = t.max_teams && t.participant_count >= t.max_teams ? 'FULL' : 'JOIN';

  return `
    <div class="tournament-card ${statusClass}">
      <span class="tournament-status-badge ${statusClass}">${(t.status || 'upcoming').toUpperCase()}</span>
      ${t.approval_status && t.approval_status !== 'approved' ? `<span class="tournament-status-badge completed">${t.approval_status.toUpperCase()}</span>` : ''}
      <div class="tournament-game">${t.game_title || 'TBA'}</div>
      <div style="color:var(--text-secondary); font-size:13px; margin-bottom:16px; font-weight:600;">${t.title}</div>

      <div class="tournament-info-row">
        <span class="tournament-info-label">Start – End</span>
        <span class="tournament-info-value">${t.start_date || 'TBA'} – ${t.end_date || 'TBA'}</span>
      </div>
      <div class="tournament-info-row">
        <span class="tournament-info-label">Prize Pool</span>
        <span class="tournament-info-value" style="color:var(--neon-gold)">${t.prize_pool || 'TBA'}</span>
      </div>
      <div class="tournament-info-row">
        <span class="tournament-info-label">Players</span>
        <span class="tournament-info-value">${t.participant_count || 0}/${t.max_teams || 16}</span>
      </div>
      ${t.venue ? `
      <div class="tournament-info-row">
        <span class="tournament-info-label">Venue</span>
        <span class="tournament-info-value">${t.venue}</span>
      </div>` : ''}

      <div class="tournament-card-actions">
        <a href="/tournament/${t.tournament_id}" class="btn btn-secondary btn-sm">VISIT GAME PAGE</a>
        ${!isAdmin ? `<button class="btn btn-primary btn-sm" ${isClosed ? 'disabled' : ''} onclick="joinTournament(${t.tournament_id}, event)">${isClosed ? joinLabel : 'JOIN'}</button>` : ''}
        ${isAdmin && showActions ? `
          <button class="btn btn-danger btn-sm" onclick="deleteTournament(${t.tournament_id}, event)">DEL</button>
        ` : ''}
      </div>
    </div>
  `;
}

// ── Create Tournament ──────────────────────────────────────

function createTournament() {
  const title      = document.getElementById('title').value.trim();
  const game_title = document.getElementById('game_title').value;
  const description= document.getElementById('description').value.trim();
  const start_date = document.getElementById('start_date').value;
  const end_date   = document.getElementById('end_date').value;
  const venue      = document.getElementById('venue').value.trim();
  const max_teams  = parseInt(document.getElementById('max_teams').value) || 16;
  const prize_pool = document.getElementById('prize_pool').value.trim();
  const status     = document.getElementById('status').value;
  const image_url  = document.getElementById('image_url').value.trim();

  if (!title || !game_title || !start_date || !end_date) {
    showError('error-msg', 'Title, game, start date and end date are required.');
    return;
  }
  if (new Date(end_date) <= new Date(start_date)) {
    showError('error-msg', 'End date/time must be after start date/time.');
    return;
  }
  if (max_teams < 1) {
    showError('error-msg', 'Maximum players must be at least 1.');
    return;
  }

  fetch('/api/tournament', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ title, game_title, description, start_date, end_date, venue, max_teams, prize_pool, status, image_url })
  })
  .then(r => r.json())
  .then(data => {
    if (data.error) {
      showError('error-msg', data.error);
      return;
    }
    showSuccess('success-msg', 'Tournament created! Redirecting...');
    setTimeout(() => window.location.href = '/tournaments', 1000);
  })
  .catch(() => showError('error-msg', 'Network error. Try again.'));
}

function joinTournament(id, event) {
  if (event) event.stopPropagation();
  fetch('/api/tournament/' + id + '/join', {
    method: 'POST',
    credentials: 'include',
  })
    .then(r => r.json())
    .then(data => {
      if (data.error) {
        alert(data.error);
        return;
      }
      alert('Joined tournament.');
      if (typeof loadTournamentDetail === 'function') loadTournamentDetail(id);
      if (typeof loadTournamentParticipants === 'function') loadTournamentParticipants(id);
      if (typeof loadTournaments === 'function') loadTournaments();
    })
    .catch(() => alert('Network error.'));
}

// ── Update Tournament ──────────────────────────────────────

function updateTournament(id) {
  const title      = document.getElementById('edit-title').value.trim();
  const game_title = document.getElementById('edit-game').value.trim();
  const status     = document.getElementById('edit-status').value;

  fetch('/api/tournament/' + id, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ title, game_title, status,
      start_date: document.getElementById('edit-start').value,
      end_date:   document.getElementById('edit-end').value,
      venue:      document.getElementById('edit-venue').value,
      max_teams:  parseInt(document.getElementById('edit-max').value),
      prize_pool: document.getElementById('edit-prize').value,
      description:document.getElementById('edit-desc').value,
    })
  })
  .then(r => r.json())
  .then(data => {
    if (data.error) { alert('Error: ' + data.error); return; }
    alert('Tournament updated!');
    location.reload();
  })
  .catch(() => alert('Network error.'));
}

// ── Delete Tournament ──────────────────────────────────────

function deleteTournament(id, event) {
  if (event) event.stopPropagation();
  if (!confirm('Are you sure you want to DELETE this tournament?')) return;

  fetch('/api/tournament/' + id, {
    method: 'DELETE',
    credentials: 'include'
  })
  .then(r => r.json())
  .then(data => {
    if (data.error) { alert('Error: ' + data.error); return; }
    alert('Tournament deleted.');
    location.reload();
  })
  .catch(() => alert('Network error.'));
}
