function loadProfile() {
  if (window.isStaticPreview) {
    renderProfile({ display_name: 'Preview Player', username: 'preview', favorite_game: 'Valorant', bio: '', matches_played: 12, wins: 7 });
    return;
  }

  checkAuth().then(function (user) {
    if (!user) {
      window.location.href = window.previewRoute ? window.previewRoute('/login') : '/login';
      return;
    }
    fetch('/api/profile', { credentials: 'include' })
      .then(function (response) { return response.json(); })
      .then(renderProfile)
      .catch(function () { showError('profile-error', 'Could not load profile.'); });
  });
}

function renderProfile(profile) {
  const name = profile.display_name || profile.username || 'Player';
  document.getElementById('profile-name').textContent = name;
  document.getElementById('profile-username').textContent = '@' + (profile.username || 'player');
  document.getElementById('profile-avatar').textContent = name.charAt(0).toUpperCase();
  document.getElementById('profile-matches').textContent = profile.matches_played || 0;
  document.getElementById('profile-wins').textContent = profile.wins || 0;
  document.getElementById('profile-display-name').value = profile.display_name || '';
  document.getElementById('profile-favorite-game').value = profile.favorite_game || '';
  document.getElementById('profile-bio').value = profile.bio || '';
}

function saveProfile() {
  fetch('/api/profile', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({
      display_name: document.getElementById('profile-display-name').value.trim(),
      favorite_game: document.getElementById('profile-favorite-game').value.trim(),
      bio: document.getElementById('profile-bio').value.trim(),
    }),
  })
    .then(function (response) { return response.json(); })
    .then(function (data) {
      if (data.error) {
        showError('profile-error', data.error);
        return;
      }
      showSuccess('profile-success', 'Profile saved.');
      loadProfile();
    })
    .catch(function () { showError('profile-error', 'Could not save profile.'); });
}

window.addEventListener('DOMContentLoaded', loadProfile);
