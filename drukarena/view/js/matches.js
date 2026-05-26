// matches.js — DrukArena Match Management

function scheduleMatch() {
  const tournament_id = parseInt(document.getElementById('match-tournament').value);
  const team1_id      = parseInt(document.getElementById('match-team1').value);
  const team2_id      = parseInt(document.getElementById('match-team2').value);
  const match_date    = document.getElementById('match-date').value;
  const match_time    = document.getElementById('match-time').value;

  if (!tournament_id || !team1_id || !team2_id || !match_date) {
    showError('match-error', 'Tournament, both teams and date are required.');
    return;
  }

  if (team1_id === team2_id) {
    showError('match-error', 'A team cannot play against itself.');
    return;
  }

  fetch('/api/match', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ tournament_id, team1_id, team2_id, match_date, match_time, status: 'scheduled' })
  })
  .then(r => r.json())
  .then(data => {
    if (data.error) { showError('match-error', data.error); return; }
    showSuccess('match-success', 'Match scheduled! ID: ' + data.match_id);
    loadAdminMatches();
  })
  .catch(() => showError('match-error', 'Network error.'));
}

function updateMatchScore(matchId) {
  const team1Score = prompt('Score for Team 1:');
  const team2Score = prompt('Score for Team 2:');
  const winner     = prompt('Winner team ID (or leave blank):');
  const status     = prompt('Match status (scheduled/live/completed):', 'completed');

  if (team1Score === null) return;

  fetch('/api/match/' + matchId, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({
      score_team1: team1Score,
      score_team2: team2Score,
      winner_team_id: parseInt(winner) || 0,
      status: status || 'completed'
    })
  })
  .then(r => r.json())
  .then(data => {
    if (data.error) { alert('Error: ' + data.error); return; }
    alert('Match updated!');
    loadAdminMatches();
  })
  .catch(() => alert('Network error.'));
}

function deleteMatch(id) {
  if (!confirm('Delete this match?')) return;

  fetch('/api/match/' + id, {
    method: 'DELETE',
    credentials: 'include'
  })
  .then(r => r.json())
  .then(data => {
    if (data.error) { alert('Error: ' + data.error); return; }
    loadAdminMatches();
  })
  .catch(() => alert('Network error.'));
}

function loadMatchTournaments() {
  fetch('/api/tournaments')
    .then(r => r.json())
    .then(data => {
      const sel = document.getElementById('match-tournament');
      if (!sel) return;
      sel.innerHTML = '<option value="">Select tournament...</option>';
      (data || []).forEach(t => {
        const opt = document.createElement('option');
        opt.value = t.tournament_id;
        opt.textContent = t.game_title + ' — ' + t.title;
        sel.appendChild(opt);
      });
    });
}

function loadMatchTeams(tournamentId) {
  if (!tournamentId) return;

  fetch('/api/tournament/' + tournamentId + '/teams')
    .then(r => r.json())
    .then(teams => {
      ['match-team1', 'match-team2'].forEach(selId => {
        const sel = document.getElementById(selId);
        if (!sel) return;
        sel.innerHTML = '<option value="">Select team...</option>';
        (teams || []).forEach(t => {
          const opt = document.createElement('option');
          opt.value = t.team_id;
          opt.textContent = t.team_name;
          sel.appendChild(opt);
        });
      });
    });
}
