// news.js powers the homepage, news page, and admin news sections.

let allNewsItems = [];

function escapeHTML(value) {
  return String(value || '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

function formatNewsDate(item) {
  return item.published_date || item.published_at || item.created_at || 'Latest Update';
}

function getNewsExcerpt(content, maxLength) {
  const text = String(content || '').trim();
  if (text.length <= maxLength) return text;
  return text.slice(0, maxLength).trim() + '...';
}

function renderNewsGrid(items) {
  const grid = document.getElementById('news-grid');
  if (!grid) return;

  const count = document.getElementById('news-count');
  if (count) count.textContent = items.length + (items.length === 1 ? ' article' : ' articles');

  if (!items.length) {
    grid.innerHTML =
      '<div class="empty-state" style="grid-column:1/-1"><div class="empty-state-icon">📰</div><div class="empty-state-title">No News Yet</div><div class="empty-state-desc">Fresh updates will land here soon.</div></div>';
    return;
  }

  const isNewsPage = document.body.contains(document.getElementById('news-search'));
  const visibleItems = isNewsPage ? items : items.slice(0, 4);

  grid.innerHTML = visibleItems.map(function (item, index) {
    const image = item.image_url
      ? '<img src="' + escapeHTML(item.image_url) + '" alt="' + escapeHTML(item.title) + '" loading="lazy">'
      : '<div class="news-card-img-placeholder">📰</div>';

    return (
      '<article class="news-card" tabindex="0" onclick="openNewsModal(' + index + ')" onkeydown="handleNewsCardKey(event, ' + index + ')">' +
        '<div class="news-card-img">' +
          '<span class="year-badge">' + escapeHTML(formatNewsDate(item)) + '</span>' +
          image +
        '</div>' +
        '<div class="news-card-body">' +
          '<h3 class="news-card-title">' + escapeHTML(item.title) + '</h3>' +
          '<p class="news-card-excerpt">' + escapeHTML(getNewsExcerpt(item.content, isNewsPage ? 180 : 110)) + '</p>' +
          '<div class="news-card-date">' + escapeHTML(formatNewsDate(item)) + '</div>' +
          '<button class="btn-read-more" type="button">Read Article</button>' +
        '</div>' +
      '</article>'
    );
  }).join('');
}

function getFilteredNews() {
  const search = document.getElementById('news-search');
  const sort = document.getElementById('news-sort');
  const query = search ? search.value.trim().toLowerCase() : '';
  const sortMode = sort ? sort.value : 'newest';

  let items = allNewsItems.filter(function (item) {
    const haystack = (item.title + ' ' + item.content).toLowerCase();
    return !query || haystack.includes(query);
  });

  items = items.slice().sort(function (a, b) {
    if (sortMode === 'title') return String(a.title || '').localeCompare(String(b.title || ''));

    const aTime = Date.parse(formatNewsDate(a)) || 0;
    const bTime = Date.parse(formatNewsDate(b)) || 0;
    return sortMode === 'oldest' ? aTime - bTime : bTime - aTime;
  });

  return items;
}

function refreshNewsGrid() {
  renderNewsGrid(getFilteredNews());
}

function openNewsModal(index) {
  const items = getFilteredNews();
  const item = items[index];
  const modal = document.getElementById('news-modal');
  const content = document.getElementById('news-modal-content');
  if (!item || !modal || !content) return;

  const image = item.image_url
    ? '<img class="news-modal-img" src="' + escapeHTML(item.image_url) + '" alt="' + escapeHTML(item.title) + '">'
    : '';

  content.innerHTML =
    image +
    '<div class="news-card-date" style="margin-bottom:12px;">' + escapeHTML(formatNewsDate(item)) + '</div>' +
    '<h2 class="modal-title">' + escapeHTML(item.title) + '</h2>' +
    '<p class="news-modal-content">' + escapeHTML(item.content).replace(/\n/g, '<br>') + '</p>';

  modal.classList.add('open');
}

function closeNewsModal() {
  const modal = document.getElementById('news-modal');
  if (modal) modal.classList.remove('open');
}

function handleNewsCardKey(event, index) {
  if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault();
    openNewsModal(index);
  }
}

function loadNews() {
  if (window.isStaticPreview) {
    allNewsItems = [];
    refreshNewsGrid();
    return Promise.resolve(allNewsItems);
  }

  return fetch('/api/news')
    .then(function (response) { return response.json(); })
    .then(function (data) {
      allNewsItems = Array.isArray(data) ? data : [];
      refreshNewsGrid();
      return allNewsItems;
    })
    .catch(function () {
      allNewsItems = [];
      refreshNewsGrid();
      return [];
    });
}

function loadAdminNews() {
  const tbody = document.getElementById('admin-news-tbody');
  if (!tbody) return Promise.resolve([]);

  if (window.isStaticPreview) {
    tbody.innerHTML = '<tr><td colspan="3" style="text-align:center;padding:32px;color:var(--text-secondary);">No news articles yet.</td></tr>';
    return Promise.resolve([]);
  }

  return fetch('/api/news')
    .then(function (response) { return response.json(); })
    .then(function (data) {
      const items = Array.isArray(data) ? data : [];
      if (!items.length) {
        tbody.innerHTML = '<tr><td colspan="3" style="text-align:center;padding:32px;color:var(--text-secondary);">No news articles yet.</td></tr>';
        return items;
      }

      tbody.innerHTML = items.map(function (item) {
        return (
          '<tr>' +
            '<td>' + escapeHTML(item.title) + '</td>' +
            '<td>' + escapeHTML(formatNewsDate(item)) + '</td>' +
            '<td><button class="btn btn-danger btn-sm" onclick="deleteNews(' + item.news_id + ')">Delete</button></td>' +
          '</tr>'
        );
      }).join('');

      return items;
    })
    .catch(function () {
      tbody.innerHTML = '<tr><td colspan="3" style="text-align:center;padding:32px;color:var(--text-secondary);">Could not load news.</td></tr>';
      return [];
    });
}

function addNews() {
  if (window.isStaticPreview) {
    showSuccess('news-success', 'Preview mode: news publishing is disabled.');
    return;
  }

  const title = document.getElementById('news-title').value.trim();
  const content = document.getElementById('news-content').value.trim();
  const image_url = document.getElementById('news-image').value.trim();

  if (!title || !content) {
    showError('news-error', 'Title and content are required.');
    return;
  }

  fetch('/api/news', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ title, content, image_url }),
  })
    .then(function (response) { return response.json(); })
    .then(function (data) {
      if (data.error) {
        showError('news-error', data.error);
        return;
      }
      document.getElementById('news-title').value = '';
      document.getElementById('news-content').value = '';
      document.getElementById('news-image').value = '';
      showSuccess('news-success', 'News article published.');
      loadAdminNews();
    })
    .catch(function () {
      showError('news-error', 'Network error. Try again.');
    });
}

function deleteNews(id) {
  if (window.isStaticPreview) return;

  fetch('/api/news/' + id, {
    method: 'DELETE',
    credentials: 'include',
  })
    .then(function (response) { return response.json(); })
    .then(function () { loadAdminNews(); })
    .catch(function () {});
}

window.addEventListener('DOMContentLoaded', function () {
  if (document.getElementById('news-grid')) {
    loadNews();
  }

  const search = document.getElementById('news-search');
  const sort = document.getElementById('news-sort');
  if (search) search.addEventListener('input', refreshNewsGrid);
  if (sort) sort.addEventListener('change', refreshNewsGrid);

  const modal = document.getElementById('news-modal');
  if (modal) {
    modal.addEventListener('click', function (event) {
      if (event.target === modal) closeNewsModal();
    });
  }
});
