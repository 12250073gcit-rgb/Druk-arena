// Shared DrukArena navigation.

(function () {
  const staticPreview = window.location.protocol === 'file:' || window.location.pathname.startsWith('/view/');
  const pagePath = staticPreview ? window.location.pathname.split('/').pop() : window.location.pathname;

  const links = [
    { href: '/', preview: 'index.html', key: 'home', icon: '⌂', label: 'Home' },
    { href: '/tournaments', preview: 'tournaments.html', key: 'tournaments', icon: '⚔', label: 'Competitions' },
    { href: '/teams', preview: 'teams.html', key: 'teams', icon: '👥', label: 'Teams' },
    { href: '/matches', preview: 'matches.html', key: 'matches', icon: '📊', label: 'Matches' },
    { href: '/news', preview: 'news.html', key: 'news', icon: '📰', label: 'News' },
    { href: '/gallery', preview: 'gallery.html', key: 'gallery', icon: '▣', label: 'Gallery' },
    { href: '/profile', preview: 'profile.html', key: 'profile', icon: '◎', label: 'Profile' },
  ];

  function hrefFor(link) {
    return staticPreview ? link.preview : link.href;
  }

  function isActive(link) {
    if (staticPreview) {
      if (link.key === 'home') return pagePath === '' || pagePath === 'index.html';
      return pagePath === link.preview;
    }

    if (link.key === 'home') return window.location.pathname === '/';
    if (link.key === 'tournaments') return window.location.pathname.startsWith('/tournament') || window.location.pathname === '/create-tournament';
    return window.location.pathname === link.href;
  }

  function renderNav() {
    document.querySelectorAll('.mobile-nav, .nav-backdrop').forEach(function (node) {
      node.remove();
    });

    const oldSidebar = document.querySelector('.sidebar');
    const sidebar = oldSidebar || document.createElement('nav');
    sidebar.className = 'sidebar';
    sidebar.setAttribute('aria-label', 'Primary navigation');

    sidebar.innerHTML =
      '<a href="' + hrefFor(links[0]) + '" class="sidebar-logo" aria-label="DrukArena home">' +
        '<img src="' + (staticPreview ? 'logo.png' : '/logo.png') + '" alt="DrukArena">' +
        '<span class="sidebar-brand">Druk<br>Arena</span>' +
      '</a>' +
      '<div class="nav-links">' +
        links.map(function (link) {
          return '<a href="' + hrefFor(link) + '" class="nav-item' + (isActive(link) ? ' active' : '') + '">' +
            '<span class="nav-icon">' + link.icon + '</span><span class="nav-label">' + link.label + '</span>' +
          '</a>';
        }).join('') +
      '</div>' +
      '<div class="nav-divider"></div>' +
      '<a href="' + (staticPreview ? 'admin-dashboard.html' : '/admin') + '" id="admin-nav-item" class="nav-item admin-link" style="display:none;">' +
        '<span class="nav-icon">⚙</span><span class="nav-label">Admin</span>' +
      '</a>' +
      '<a href="' + (staticPreview ? 'login.html' : '/login') + '" class="nav-item logout-btn" id="auth-btn">' +
        '<span class="nav-icon">👤</span><span class="nav-label">Login</span>' +
      '</a>';

    if (!oldSidebar) document.body.prepend(sidebar);

    const mobile = document.createElement('div');
    mobile.className = 'mobile-nav';
    mobile.id = 'mobile-nav';
    mobile.innerHTML =
      '<button class="mobile-menu-btn" type="button" aria-label="Open navigation" aria-expanded="false"><span class="nav-icon">☰</span></button>' +
      '<a href="' + hrefFor(links[0]) + '" class="mobile-brand">' +
        '<img src="' + (staticPreview ? 'logo.png' : '/logo.png') + '" alt="DrukArena">' +
        '<span>DRUKARENA</span>' +
      '</a>' +
      '<a href="' + (staticPreview ? 'login.html' : '/login') + '" class="btn btn-primary btn-sm" id="mobile-auth-btn">Login</a>';

    const backdrop = document.createElement('button');
    backdrop.className = 'nav-backdrop';
    backdrop.type = 'button';
    backdrop.setAttribute('aria-label', 'Close navigation');

    document.body.prepend(backdrop);
    document.body.prepend(mobile);

    document.body.classList.remove('nav-collapsed');
    localStorage.removeItem('drukarena-nav-collapsed');

    function closeMobileNav() {
      document.body.classList.remove('nav-open');
      mobile.querySelector('.mobile-menu-btn').setAttribute('aria-expanded', 'false');
    }

    mobile.querySelector('.mobile-menu-btn').addEventListener('click', function () {
      const next = !document.body.classList.contains('nav-open');
      document.body.classList.toggle('nav-open', next);
      this.setAttribute('aria-expanded', String(next));
    });
    backdrop.addEventListener('click', closeMobileNav);
    sidebar.querySelectorAll('a').forEach(function (link) {
      link.addEventListener('click', closeMobileNav);
    });
    window.addEventListener('resize', function () {
      if (window.innerWidth > 540) closeMobileNav();
    });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', renderNav);
  } else {
    renderNav();
  }
})();
