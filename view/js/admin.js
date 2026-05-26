// admin.js provides the missing admin dashboard data loaders.

function adminEscape(value) {
  return String(value || '').replace(/[&<>"']/g, function (ch) {
    return ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[ch];
  });
}

function adminFormatDate(value) {
  if (!value) return '—';
  const date = new Date(String(value).replace(' ', 'T'));
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString([], {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
}

function setAdminOverview(stats) {
  const tournaments = document.getElementById('stat-tournaments');
  const teams = document.getElementById('stat-teams');
  const users = document.getElementById('stat-users');

  if (tournaments) tournaments.textContent = stats.tournaments;
  if (teams) teams.textContent = stats.teams;
  if (users) users.textContent = stats.users;
}

function loadAdminData() {
  if (window.isStaticPreview) {
    setAdminOverview({ tournaments: 3, teams: 12, users: 48 });
    loadAdminTournaments();
    loadAdminNews();
    loadAdminMatches();
    loadAdminUsers();
    loadAdminGallery();
    return Promise.resolve();
  }

  return Promise.all([
    fetch('/api/admin/stats').then(function (response) { return response.json(); }).catch(function () { return {}; }),
    loadAdminTournaments(),
    loadAdminNews(),
    loadAdminMatches(),
    loadAdminUsers(),
    loadAdminGallery(),
  ]).then(function (results) {
    const stats = results[0] || {};
    setAdminOverview({
      tournaments: stats.total_tournaments || stats.tournaments || 0,
      teams: stats.total_teams || stats.teams || 0,
      users: stats.total_users || stats.users || 0,
    });
  });
}

function loadAdminTournaments() {
  const tbody = document.getElementById('admin-tournaments-tbody');
  if (!tbody) return Promise.resolve([]);

  if (window.isStaticPreview) {
    tbody.innerHTML =
      '<tr><td>Campus Championship</td><td>Valorant</td><td>upcoming</td><td>2026-07-06</td><td>Nu. 50,000</td><td><button class="btn btn-secondary btn-sm" disabled>Preview Only</button></td></tr>';
    return Promise.resolve([]);
  }

  return fetch('/api/tournaments?scope=admin', { credentials: 'include' })
    .then(function (response) { return response.json(); })
    .then(function (data) {
      const items = Array.isArray(data) ? data : [];
      if (!items.length) {
        tbody.innerHTML = '<tr><td colspan="6" style="text-align:center;padding:32px;color:var(--text-secondary);">No tournaments found.</td></tr>';
        return items;
      }

      tbody.innerHTML = items.map(function (item) {
        return (
          '<tr>' +
            '<td>' + item.title + '</td>' +
            '<td>' + (item.game_title || '—') + '</td>' +
            '<td>' + (item.status || 'upcoming') + ' / ' + (item.approval_status || 'approved') + '</td>' +
            '<td>' + (item.start_date || '—') + '</td>' +
            '<td>' + (item.prize_pool || '—') + '</td>' +
            '<td>' +
              (item.approval_status !== 'approved' ? '<button class="btn btn-primary btn-sm" onclick="approveTournament(' + item.tournament_id + ')">Approve</button> ' : '') +
              (item.approval_status === 'pending' ? '<button class="btn btn-secondary btn-sm" onclick="rejectTournament(' + item.tournament_id + ')">Reject</button> ' : '') +
              '<button class="btn btn-danger btn-sm" onclick="deleteTournament(' + item.tournament_id + ', event)">Delete</button>' +
            '</td>' +
          '</tr>'
        );
      }).join('');

      return items;
    })
    .catch(function () {
      tbody.innerHTML = '<tr><td colspan="6" style="text-align:center;padding:32px;color:var(--text-secondary);">Could not load tournaments.</td></tr>';
      return [];
    });
}

function approveTournament(id) {
  fetch('/api/tournament/' + id + '/approve', { method: 'POST', credentials: 'include' })
    .then(function () { loadAdminTournaments(); });
}

function rejectTournament(id) {
  fetch('/api/tournament/' + id + '/reject', { method: 'POST', credentials: 'include' })
    .then(function () { loadAdminTournaments(); });
}

function loadAdminMatches() {
  const tbody = document.getElementById('admin-matches-tbody');
  if (!tbody) return Promise.resolve([]);

  if (window.isStaticPreview) {
    tbody.innerHTML =
      '<tr><td>Thunder Dragons</td><td>vs</td><td>Royal Gaming</td><td>2026-07-08</td><td>scheduled</td><td>—</td><td><button class="btn btn-secondary btn-sm" disabled>Preview Only</button></td></tr>';
    return Promise.resolve([]);
  }

  return fetch('/api/matches')
    .then(function (response) { return response.json(); })
    .then(function (data) {
      const items = Array.isArray(data) ? data : [];
      if (!items.length) {
        tbody.innerHTML = '<tr><td colspan="7" style="text-align:center;padding:32px;color:var(--text-secondary);">No matches scheduled yet.</td></tr>';
        return items;
      }

      tbody.innerHTML = items.map(function (item) {
        return (
          '<tr>' +
            '<td>' + (item.team1_name || ('Team #' + item.team1_id)) + '</td>' +
            '<td>vs</td>' +
            '<td>' + (item.team2_name || ('Team #' + item.team2_id)) + '</td>' +
            '<td>' + (item.match_date || '—') + '</td>' +
            '<td>' + (item.status || 'scheduled') + '</td>' +
            '<td>' + ((item.score_team1 && item.score_team2) ? (item.score_team1 + ' - ' + item.score_team2) : '—') + '</td>' +
            '<td><button class="btn btn-secondary btn-sm" onclick="updateMatchScore(' + item.match_id + ')">Manual Score</button> <button class="btn btn-danger btn-sm" onclick="deleteMatch(' + item.match_id + ')">Delete</button></td>' +
          '</tr>'
        );
      }).join('');

      return items;
    })
    .catch(function () {
      tbody.innerHTML = '<tr><td colspan="7" style="text-align:center;padding:32px;color:var(--text-secondary);">Could not load matches.</td></tr>';
      return [];
    });
}

function loadAdminGallery() {
  const tbody = document.getElementById('admin-gallery-tbody');
  if (!tbody) return Promise.resolve([]);

  if (window.isStaticPreview) {
    tbody.innerHTML = '<tr><td>Preview photo</td><td>Guest-2048</td><td>guest@example.test</td><td>Static Preview</td><td>Preview</td><td><button class="btn btn-secondary btn-sm" disabled>Preview Only</button></td></tr>';
    return Promise.resolve([]);
  }

  return fetch('/api/admin/gallery', { credentials: 'include' })
    .then(function (response) { return response.json(); })
    .then(function (data) {
      const items = Array.isArray(data) ? data : [];
      if (!items.length) {
        tbody.innerHTML = '<tr><td colspan="6" style="text-align:center;padding:32px;color:var(--text-secondary);">No gallery uploads yet.</td></tr>';
        return items;
      }
      tbody.innerHTML = items.map(function (item) {
        const email = item.uploader_email || '';
        return '<tr>' +
          '<td>' + adminEscape(item.title) + '</td>' +
          '<td>' + adminEscape(item.uploader_name || item.username || 'Guest') + '</td>' +
          '<td>' + adminEscape(email || '—') + '</td>' +
          '<td>' + adminEscape(adminFormatDate(item.created_at)) + '</td>' +
          '<td><a class="form-link" href="' + adminEscape(item.image_url) + '" target="_blank">View</a></td>' +
          '<td>' +
            '<button class="btn btn-danger btn-sm" onclick="deleteGalleryPhoto(' + item.gallery_id + ')">Delete</button> ' +
            (email ? '<button class="btn btn-secondary btn-sm" onclick="blockGalleryUploader(' + item.gallery_id + ')">Block Email</button>' : '') +
          '</td>' +
        '</tr>';
      }).join('');
      return items;
    })
    .catch(function () {
      tbody.innerHTML = '<tr><td colspan="6" style="text-align:center;padding:32px;color:var(--text-secondary);">Could not load gallery uploads.</td></tr>';
      return [];
    });
}

function approveGallery(id) {
  fetch('/api/admin/gallery/' + id + '/approve', { method: 'POST', credentials: 'include' })
    .then(function () { loadAdminGallery(); });
}

function rejectGallery(id) {
  fetch('/api/admin/gallery/' + id + '/reject', { method: 'POST', credentials: 'include' })
    .then(function () { loadAdminGallery(); });
}

function deleteGalleryPhoto(id) {
  if (!confirm('Delete this gallery photo?')) return;
  fetch('/api/admin/gallery/' + id, { method: 'DELETE', credentials: 'include' })
    .then(function (response) { return response.json(); })
    .then(function (data) {
      if (data.error) {
        alert(data.error);
        return;
      }
      loadAdminGallery();
    })
    .catch(function () { alert('Could not delete gallery photo.'); });
}

function blockGalleryUploader(id) {
  if (!confirm('Block this uploader email from future photo uploads?')) return;
  fetch('/api/admin/gallery/' + id + '/block-uploader', { method: 'POST', credentials: 'include' })
    .then(function (response) { return response.json(); })
    .then(function (data) {
      if (data.error) {
        alert(data.error);
        return;
      }
      alert('Uploader email blocked.');
      loadAdminGallery();
    })
    .catch(function () { alert('Could not block uploader email.'); });
}

function loadAdminUsers() {
  const tbody = document.getElementById('admin-users-tbody');
  if (!tbody) return Promise.resolve([]);

  if (window.isStaticPreview) {
    tbody.innerHTML =
      '<tr><td>1</td><td>preview-admin</td><td>preview@drukarena.test</td><td>admin</td><td>Static Preview</td><td>—</td></tr>';
    return Promise.resolve([]);
  }

  return fetch('/api/admin/users', { credentials: 'include' })
    .then(function (response) { return response.json(); })
    .then(function (data) {
      const items = Array.isArray(data) ? data : [];
      if (!items.length) {
        tbody.innerHTML = '<tr><td colspan="6" style="text-align:center;padding:32px;color:var(--text-secondary);">No users found.</td></tr>';
        return items;
      }

      tbody.innerHTML = items.map(function (item) {
        return (
          '<tr>' +
            '<td>' + item.user_id + '</td>' +
            '<td>' + item.username + '</td>' +
            '<td>' + item.email + '</td>' +
            '<td>' + item.role + '</td>' +
            '<td>' + (item.created_at || '—') + '</td>' +
            '<td>' + (item.role === 'admin' ? '—' : '<button class="btn btn-danger btn-sm" onclick="removeUser(' + item.user_id + ')">Remove</button>') + '</td>' +
          '</tr>'
        );
      }).join('');

      return items;
    })
    .catch(function () {
      tbody.innerHTML = '<tr><td colspan="6" style="text-align:center;padding:32px;color:var(--text-secondary);">Could not load users.</td></tr>';
      return [];
    });
}

function removeUser(id) {
  if (!confirm('Remove this user from DrukArena?')) return;
  fetch('/api/admin/users/' + id, { method: 'DELETE', credentials: 'include' })
    .then(function (response) { return response.json(); })
    .then(function (data) {
      if (data.error) {
        alert(data.error);
        return;
      }
      loadAdminUsers();
      loadAdminData();
    })
    .catch(function () { alert('Could not remove user.'); });
}
