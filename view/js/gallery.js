function escapeGalleryText(value) {
  return String(value || '').replace(/[&<>"']/g, function (ch) {
    return ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[ch];
  });
}

let galleryCurrentUser = null;
let galleryRulesInitialized = false;

function galleryRulesKey(user) {
  return 'drukarena-gallery-rules-agreed-' + user.user_id;
}

function formatGalleryDate(value) {
  if (!value) return '';
  const normalized = String(value).replace(' ', 'T');
  const date = new Date(normalized);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString([], {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
}

function openGalleryRules() {
  const modal = document.getElementById('gallery-rules-modal');
  if (modal) modal.classList.add('open');
}

function closeGalleryRules() {
  const modal = document.getElementById('gallery-rules-modal');
  if (modal) modal.classList.remove('open');
}

function acceptGalleryRules() {
  window.galleryRulesAccepted = true;
  document.body.dataset.galleryRulesAccepted = 'true';
  if (galleryCurrentUser) {
    localStorage.setItem(galleryRulesKey(galleryCurrentUser), 'true');
  }
  closeGalleryRules();
}
window.acceptGalleryRules = acceptGalleryRules;

function setupGalleryRules(user) {
  galleryCurrentUser = user;
  if (user) {
    if (localStorage.getItem(galleryRulesKey(user)) === 'true') {
      window.galleryRulesAccepted = true;
      document.body.dataset.galleryRulesAccepted = 'true';
    } else if (window.galleryRulesAccepted) {
      localStorage.setItem(galleryRulesKey(user), 'true');
      document.body.dataset.galleryRulesAccepted = 'true';
    }
  }
  const agreeBtn = document.getElementById('gallery-rules-agree');
  if (agreeBtn && !galleryRulesInitialized) {
    agreeBtn.addEventListener('click', acceptGalleryRules);
    galleryRulesInitialized = true;
  }

  if (window.galleryRulesAccepted) {
    closeGalleryRules();
  } else {
    openGalleryRules();
  }
}

function renderGallery(items) {
  const grid = document.getElementById('gallery-grid');
  const count = document.getElementById('gallery-count');
  if (count) count.textContent = items.length + (items.length === 1 ? ' photo' : ' photos');
  if (!items.length) {
    grid.innerHTML = '<div class="empty-state" style="grid-column:1/-1"><div class="empty-state-icon">▣</div><div class="empty-state-title">No Photos Yet</div><div class="empty-state-desc">Community match photos will appear here.</div></div>';
    return;
  }
  grid.innerHTML = items.map(function (item) {
    const uploader = item.uploader_name || item.username || 'Guest';
    const uploadedAt = formatGalleryDate(item.created_at);
    return '<article class="gallery-card">' +
      '<img src="' + escapeGalleryText(item.image_url) + '" alt="' + escapeGalleryText(item.title) + '">' +
      '<div class="gallery-card-body">' +
        '<div class="gallery-card-title">' + escapeGalleryText(item.title) + '</div>' +
        '<div class="gallery-card-meta">Uploaded by ' + escapeGalleryText(uploader) + '</div>' +
        (uploadedAt ? '<div class="gallery-card-time">' + escapeGalleryText(uploadedAt) + '</div>' : '') +
      '</div>' +
    '</article>';
  }).join('');
}

function loadGallery() {
  if (window.isStaticPreview) {
    renderGallery([]);
    return;
  }
  fetch('/api/gallery')
    .then(function (response) { return response.json(); })
    .then(function (data) { renderGallery(Array.isArray(data) ? data : []); })
    .catch(function () { renderGallery([]); });
}

function setupGalleryUpload() {
  const form = document.getElementById('gallery-form');
  if (!form) return;
  if (document.body.dataset.galleryRulesAccepted !== 'true') {
    window.galleryRulesAccepted = false;
    document.body.dataset.galleryRulesAccepted = 'false';
  }
  openGalleryRules();
  const agreeBtn = document.getElementById('gallery-rules-agree');
  if (agreeBtn && !galleryRulesInitialized) {
    agreeBtn.addEventListener('click', acceptGalleryRules);
    galleryRulesInitialized = true;
  }
  checkAuth().then(function (user) {
    setupGalleryRules(user);
  });
  form.addEventListener('submit', function (event) {
    event.preventDefault();
    if (window.isStaticPreview) {
      showSuccess('gallery-success', 'Preview mode: upload disabled.');
      return;
    }
    if (document.body.dataset.galleryRulesAccepted !== 'true') {
      openGalleryRules();
      showError('gallery-error', 'Please agree to the photo rules before uploading.');
      return;
    }
    const image = document.getElementById('gallery-image');
    if (!image.files || !image.files.length) {
      showError('gallery-error', 'Please choose a photo to upload.');
      return;
    }
    const body = new FormData(form);
    body.set('agree_terms', 'true');
    fetch('/api/gallery', { method: 'POST', body: body, credentials: 'include' })
      .then(function (response) { return response.json(); })
      .then(function (data) {
        if (data.error) {
          showError('gallery-error', data.error);
          return;
        }
        form.reset();
        showSuccess('gallery-success', 'Photo uploaded to the gallery.');
        loadGallery();
      })
      .catch(function () { showError('gallery-error', 'Could not upload photo.'); });
  });
}

window.addEventListener('DOMContentLoaded', function () {
  setupGalleryUpload();
  loadGallery();
});
