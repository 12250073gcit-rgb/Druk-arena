// auth.js — DrukArena Authentication
// Following CSC103 v1.3 patterns: fetch, then/catch, cookie-based session

// ── Helpers ──────────────────────────────────────────────

function showError(id, msg) {
  const el = document.getElementById(id);
  if (!el) { alert('Error: ' + msg); return; }
  el.textContent = msg;
  el.classList.add('show');
  setTimeout(() => el.classList.remove('show'), 5000);
}

function showSuccess(id, msg) {
  const el = document.getElementById(id);
  if (!el) { alert(msg); return; }
  el.textContent = msg;
  el.classList.add('show');
  setTimeout(() => el.classList.remove('show'), 4000);
}

function isValidEmail(email) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

function validatePassword(password) {
  if (password.length < 8) return 'Password must be at least 8 characters.';
  if (!/[A-Za-z]/.test(password) || !/[0-9]/.test(password)) {
    return 'Password must include at least one letter and one number.';
  }
  return '';
}

function togglePassword(inputId, btn) {
  const input = document.getElementById(inputId);
  if (!input) return;
  const show = input.type === 'password';
  input.type = show ? 'text' : 'password';
  updatePasswordToggle(btn, show);
}

function passwordIcon(isVisible) {
  if (isVisible) {
    return '<svg class="password-eye" viewBox="0 0 24 24" aria-hidden="true"><path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12z"/><circle cx="12" cy="12" r="3"/></svg>';
  }
  return '<svg class="password-eye" viewBox="0 0 24 24" aria-hidden="true"><path d="M3 3l18 18"/><path d="M10.6 10.6A2 2 0 0 0 13.4 13.4"/><path d="M9.9 5.2A10.8 10.8 0 0 1 12 5c5 0 8.5 4.1 10 7a16.5 16.5 0 0 1-3.1 4.1"/><path d="M6.6 6.6A16.5 16.5 0 0 0 2 12c1.5 2.9 5 7 10 7 1.5 0 2.9-.4 4.1-1"/></svg>';
}

function updatePasswordToggle(btn, isVisible) {
  if (!btn) return;
  btn.innerHTML = passwordIcon(isVisible);
  btn.setAttribute('aria-label', isVisible ? 'Hide password' : 'Show password');
  btn.setAttribute('aria-pressed', isVisible ? 'true' : 'false');
  btn.title = isVisible ? 'Hide password' : 'Show password';
}

function initPasswordToggles() {
  document.querySelectorAll('[data-password-toggle]').forEach(btn => {
    const inputId = btn.getAttribute('data-target');
    updatePasswordToggle(btn, false);
    btn.addEventListener('click', () => togglePassword(inputId, btn));
  });
}

// ── Check session ─────────────────────────────────────────

function checkAuth() {
  return fetch('/verify', { credentials: 'include' })
    .then(r => {
      if (!r.ok) return null;
      return r.json();
    })
    .then(data => {
      if (!data || data.error) {
        updateNavForGuest();
        return null;
      }
      updateNavForUser(data);
      return data;
    })
    .catch(() => { updateNavForGuest(); return null; });
}

function updateNavForUser(user) {
  const authBtn = document.getElementById('auth-btn');
  if (authBtn) {
    authBtn.innerHTML = '<span class="nav-icon">⏻</span><span class="nav-label">Logout</span>';
    authBtn.href = '#';
    authBtn.onclick = logout;
  }
  const mobileAuthBtn = document.getElementById('mobile-auth-btn');
  if (mobileAuthBtn) {
    mobileAuthBtn.textContent = 'Logout';
    mobileAuthBtn.href = '#';
    mobileAuthBtn.onclick = logout;
  }
  if (user.role === 'admin') {
    const adminItem = document.getElementById('admin-nav-item');
    if (adminItem) adminItem.style.display = 'flex';
  }
  // Store in session for page-level checks
  window._currentUser = user;
}

function updateNavForGuest() {
  const loginHref = window.previewRoute ? window.previewRoute('/login') : '/login';
  const authBtn = document.getElementById('auth-btn');
  if (authBtn) {
    authBtn.innerHTML = '<span class="nav-icon">👤</span><span class="nav-label">Login</span>';
    authBtn.href = loginHref;
    authBtn.onclick = null;
  }
  const mobileAuthBtn = document.getElementById('mobile-auth-btn');
  if (mobileAuthBtn) {
    mobileAuthBtn.textContent = 'Login';
    mobileAuthBtn.href = loginHref;
    mobileAuthBtn.onclick = null;
  }
  window._currentUser = null;
}

// ── Login ─────────────────────────────────────────────────

function login() {
  const email    = document.getElementById('email').value.trim();
  const password = document.getElementById('password').value;

  if (!email || !password) {
    showError('error-msg', 'Please fill in all fields.');
    return;
  }
  if (!isValidEmail(email)) {
    showError('error-msg', 'Enter a valid email address.');
    return;
  }

  document.getElementById('login-btn').disabled = true;
  document.getElementById('login-btn').textContent = 'Logging in...';

  fetch('/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ email, password })
  })
  .then(r => r.json())
  .then(data => {
    document.getElementById('login-btn').disabled = false;
    document.getElementById('login-btn').textContent = 'ENTER THE ARENA';

    if (data.error) {
      showError('error-msg', data.error);
      return;
    }
    showSuccess('success-msg', 'Login successful! Redirecting...');
    setTimeout(() => {
      window.location.href = data.role === 'admin' ? '/admin' : '/';
    }, 800);
  })
  .catch(err => {
    document.getElementById('login-btn').disabled = false;
    document.getElementById('login-btn').textContent = 'ENTER THE ARENA';
    showError('error-msg', 'Network error. Try again.');
  });
}

// ── Signup ────────────────────────────────────────────────

function signup() {
  const username        = document.getElementById('username').value.trim();
  const email           = document.getElementById('email').value.trim();
  const password        = document.getElementById('password').value;
  const confirmPassword = document.getElementById('confirm-password').value;

  if (!username || !email || !password || !confirmPassword) {
    showError('error-msg', 'Please fill in all fields.');
    return;
  }
  if (username.length < 3) {
    showError('error-msg', 'Username must be at least 3 characters.');
    return;
  }
  if (!isValidEmail(email)) {
    showError('error-msg', 'Enter a valid email address.');
    return;
  }

  if (password !== confirmPassword) {
    showError('error-msg', 'Passwords do not match.');
    return;
  }

  const passwordError = validatePassword(password);
  if (passwordError) {
    showError('error-msg', passwordError);
    return;
  }

  fetch('/signup', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ username, email, password })
  })
  .then(r => r.json())
  .then(data => {
    if (data.error) {
      showError('error-msg', data.error);
      return;
    }
    showSuccess('success-msg', 'Account created! Entering DrukArena...');
    setTimeout(() => window.location.href = data.role === 'admin' ? '/admin' : '/', 800);
  })
  .catch(() => showError('error-msg', 'Network error. Try again.'));
}

// ── Logout ────────────────────────────────────────────────

function logout() {
  fetch('/logout', {
    method: 'POST',
    credentials: 'include'
  })
  .then(() => {
    window._currentUser = null;
    window.location.href = '/';
  })
  .catch(() => window.location.href = '/');
}

// ── Auto-check session on page load ───────────────────────

window.addEventListener('DOMContentLoaded', () => {
  initPasswordToggles();
  // Only run on pages that have a sidebar auth button
  if (document.getElementById('auth-btn') || document.getElementById('admin-nav-item')) {
    checkAuth();
  }
});

// Support Enter key on login/signup forms
document.addEventListener('keydown', (e) => {
  if (e.key !== 'Enter') return;
  if (document.getElementById('login-btn') && !document.getElementById('username')) {
    login();
  }
});
