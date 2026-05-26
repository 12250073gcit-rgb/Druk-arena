// teams.js — DrukArena Team Management

function createTeam() {
  const team_name     = document.getElementById('team-name').value.trim();
  const tournament_id = parseInt(document.getElementById('team-tournament').value);

  if (!team_name || !tournament_id) {
    showError('team-error', 'Team name and tournament are required.');
    return;
  }

  fetch('/api/team', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ team_name, tournament_id })
  })
  .then(r => r.json())
  .then(data => {
    if (data.error) {
      showError('team-error', data.error);
      return;
    }
    showSuccess('team-success', 'Team created! Team ID: ' + data.team_id);
    document.getElementById('team-name').value = '';
    setTimeout(() => location.reload(), 1500);
  })
  .catch(() => showError('team-error', 'Network error. Try again.'));
}

function deleteTeam(id) {
  if (!confirm('Are you sure you want to DELETE this team?')) return;

  fetch('/api/team/' + id, {
    method: 'DELETE',
    credentials: 'include'
  })
  .then(r => r.json())
  .then(data => {
    if (data.error) { alert('Error: ' + data.error); return; }
    alert('Team deleted.');
    location.reload();
  })
  .catch(() => alert('Network error.'));
}

function addTeamMember(teamId, userId) {
  fetch('/api/team/' + teamId + '/member', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ team_id: teamId, user_id: userId, role: 'player' })
  })
  .then(r => r.json())
  .then(data => {
    if (data.error) { alert('Error: ' + data.error); return; }
    alert('Member added!');
  })
  .catch(() => alert('Network error.'));
}

function removeMember(memberID, teamID) {
  if (!confirm('Remove this member?')) return;

  fetch('/api/team/' + teamID + '/member/' + memberID, {
    method: 'DELETE',
    credentials: 'include'
  })
  .then(r => r.json())
  .then(data => {
    if (data.error) { alert('Error: ' + data.error); return; }
    location.reload();
  })
  .catch(() => alert('Network error.'));
}
