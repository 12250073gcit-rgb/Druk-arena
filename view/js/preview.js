// preview.js keeps root-routed pages navigable when opened from /view/*.html

(function () {
  const staticPreview =
    window.location.protocol === 'file:' ||
    window.location.pathname.startsWith('/view/');

  const routeMap = {
    '/': 'index.html',
    '/login': 'login.html',
    '/signup': 'signup.html',
    '/tournaments': 'tournaments.html',
    '/create-tournament': 'create-tournament.html',
    '/teams': 'teams.html',
    '/matches': 'matches.html',
    '/news': 'news.html',
    '/gallery': 'gallery.html',
    '/profile': 'profile.html',
    '/admin': 'admin-dashboard.html',
  };

  window.isStaticPreview = staticPreview;
  window.previewRoute = function previewRoute(path) {
    if (!staticPreview || !path) return path;
    if (routeMap[path]) return routeMap[path];
    if (path.startsWith('/tournament/')) return 'tournament-detail.html';
    return path;
  };

  function rewriteOnclickNavigation(node) {
    const onclick = node.getAttribute('onclick');
    if (!onclick || !onclick.includes('window.location.href=')) return;

    node.setAttribute(
      'onclick',
      onclick.replace(/window\.location\.href='([^']+)'/g, function (_, path) {
        return "window.location.href='" + window.previewRoute(path) + "'";
      })
    );
  }

  window.addEventListener('DOMContentLoaded', function () {
    if (!staticPreview) return;

    document.querySelectorAll('a[href^="/"]').forEach(function (anchor) {
      anchor.setAttribute('href', window.previewRoute(anchor.getAttribute('href')));
    });

    document.querySelectorAll('[onclick]').forEach(rewriteOnclickNavigation);
  });
})();
