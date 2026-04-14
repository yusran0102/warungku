/* ─────────────────────────────────────────────────────────────────────────
   Warung-Ku  — app.js
   Shared utilities: modal manager, confirm dialog, toast, mobile nav.
   ───────────────────────────────────────────────────────────────────────── */

/* ═══════════════════════════════════════════════════════════════════════
   MODAL MANAGER
   ═══════════════════════════════════════════════════════════════════════
   openModal(id)  — shows overlay with enter animation
   closeModal(id) — plays exit animation then hides

   Supports two styles of modal element:
   A) New: <div id="X" class="wk-modal-overlay" style="display:none">
           Uses CSS class-based animation (.wk-modal-open).
   B) Legacy: any element with style="display:none" (raw display toggle).
              Still works; no animation on legacy ones until migrated.
   ═══════════════════════════════════════════════════════════════════════ */
function openModal(id) {
  var el = document.getElementById(id);
  if (!el) return;
  el.style.display = 'flex';
  // double rAF ensures transition fires after display:flex paint
  requestAnimationFrame(function () {
    requestAnimationFrame(function () {
      el.classList.add('wk-modal-open');
    });
  });
  document.body.style.overflow = 'hidden';
}

function closeModal(id) {
  var el = document.getElementById(id);
  if (!el) return;
  if (el.classList.contains('wk-modal-overlay')) {
    el.classList.remove('wk-modal-open');
    // wait for CSS transition before hiding
    var done = false;
    var onEnd = function () { if (done) return; done = true; el.style.display = 'none'; };
    el.addEventListener('transitionend', onEnd, { once: true });
    // safety fallback if transition doesn't fire (e.g. prefers-reduced-motion)
    setTimeout(onEnd, 280);
  } else {
    el.style.display = 'none';
  }
  document.body.style.overflow = '';
}

// ── Event delegation: [data-modal-open] and [data-modal-close] buttons ──
document.addEventListener('click', function (e) {
  var opener = e.target.closest('[data-modal-open]');
  if (opener) { openModal(opener.dataset.modalOpen); return; }

  var closer = e.target.closest('[data-modal-close]');
  if (closer) {
    var targetId = closer.dataset.modalClose;
    if (!targetId) {
      // fall back: close the nearest modal ancestor
      var overlay = closer.closest('.wk-modal-overlay');
      if (overlay) targetId = overlay.id;
    }
    if (targetId) closeModal(targetId);
    return;
  }

  // Click directly on the dark backdrop (not modal box children)
  if (e.target.classList.contains('wk-modal-overlay')) {
    closeModal(e.target.id);
  }
});

// ── Escape key closes topmost open modal ────────────────────────────────
document.addEventListener('keydown', function (e) {
  if (e.key !== 'Escape') return;
  var open = document.querySelectorAll('.wk-modal-overlay.wk-modal-open');
  if (open.length) closeModal(open[open.length - 1].id);
});


/* ═══════════════════════════════════════════════════════════════════════
   CUSTOM CONFIRM DIALOG
   ═══════════════════════════════════════════════════════════════════════
   Replaces native browser confirm() with a polished dialog.

   Usage (JS):
     wkConfirm('Hapus transaksi ini?').then(function(ok) {
       if (ok) form.submit();
     });

   Usage (HTML, replaces onsubmit="return confirm(...)"):
     Add  data-confirm="Pesan konfirmasi"  to the <form> tag.
     Remove the onsubmit handler — this file intercepts it automatically.
   ═══════════════════════════════════════════════════════════════════════ */
(function () {
  var _resolve = null;

  /* Build the DOM once on first call */
  function buildDialog () {
    if (document.getElementById('wk-confirm-overlay')) return;

    var html = [
      '<div id="wk-confirm-overlay">',
        '<div id="wk-confirm-box">',
          '<div id="wk-confirm-icon">',
            '<svg width="22" height="22" fill="none" stroke="currentColor" viewBox="0 0 24 24">',
              '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"',
              ' d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858',
              'L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>',
            '</svg>',
          '</div>',
          '<p id="wk-confirm-title"></p>',
          '<p id="wk-confirm-msg"></p>',
          '<div id="wk-confirm-actions">',
            '<button id="wk-confirm-cancel" class="btn btn-secondary">Batal</button>',
            '<button id="wk-confirm-ok" class="btn btn-danger">Hapus</button>',
          '</div>',
        '</div>',
      '</div>',
    ].join('');

    var wrapper = document.createElement('div');
    wrapper.innerHTML = html;
    document.body.appendChild(wrapper.firstChild);

    document.getElementById('wk-confirm-cancel').addEventListener('click', function () { settle(false); });
    document.getElementById('wk-confirm-ok').addEventListener('click', function () { settle(true); });
    document.getElementById('wk-confirm-overlay').addEventListener('click', function (e) {
      if (e.target.id === 'wk-confirm-overlay') settle(false);
    });
    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape') settle(false);
    });
  }

  function openDialog (msg, title, okLabel) {
    buildDialog();
    document.getElementById('wk-confirm-title').textContent = title  || 'Hapus Data?';
    document.getElementById('wk-confirm-msg').textContent   = msg    || 'Tindakan ini tidak bisa dibatalkan.';
    document.getElementById('wk-confirm-ok').textContent   = okLabel || 'Hapus';

    var overlay = document.getElementById('wk-confirm-overlay');
    overlay.style.display = 'flex';
    requestAnimationFrame(function () {
      requestAnimationFrame(function () { overlay.classList.add('open'); });
    });
  }

  function settle (result) {
    var overlay = document.getElementById('wk-confirm-overlay');
    if (!overlay || !overlay.classList.contains('open')) return;
    overlay.classList.remove('open');
    setTimeout(function () { overlay.style.display = 'none'; }, 220);
    if (_resolve) { var cb = _resolve; _resolve = null; cb(result); }
  }

  /* Public API */
  window.wkConfirm = function (msg, title, okLabel) {
    return new Promise(function (resolve) {
      _resolve = resolve;
      openDialog(msg, title, okLabel);
    });
  };

  /* Auto-intercept forms with data-confirm attribute.
     This replaces onsubmit="return confirm(...)" — just swap the attribute:
       Before: onsubmit="return confirm('Hapus ini?')"
       After:  data-confirm="Hapus ini?"               (remove onsubmit)    */
  document.addEventListener('submit', function (e) {
    var msg = e.target.dataset.confirm;
    if (!msg) return;
    e.preventDefault();
    var form = e.target;
    wkConfirm(msg).then(function (ok) {
      if (ok) {
        // Remove attribute so recursive submit doesn't re-trigger
        delete form.dataset.confirm;
        form.submit();
      }
    });
  }, true /* capture: fires before any inline onsubmit */);
}());


/* ═══════════════════════════════════════════════════════════════════════
   TOAST NOTIFICATIONS
   ═══════════════════════════════════════════════════════════════════════
   showToast(message, type, duration)
     type:     'success' | 'error' | 'warning' | 'info'  (default: 'success')
     duration: ms before auto-dismiss                     (default: 3000)

   Example:
     showToast('Produk berhasil disimpan!', 'success');
     showToast('Terjadi kesalahan.', 'error');
   ═══════════════════════════════════════════════════════════════════════ */
(function () {
  var ICONS = {
    success: '<svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7"/></svg>',
    error:   '<svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12"/></svg>',
    warning: '<svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 9v4m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"/></svg>',
    info:    '<svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>',
  };

  function getContainer () {
    var c = document.getElementById('wk-toast-container');
    if (!c) {
      c = document.createElement('div');
      c.id = 'wk-toast-container';
      document.body.appendChild(c);
    }
    return c;
  }

  window.showToast = function (message, type, duration) {
    type     = type     || 'success';
    duration = duration || 3000;

    var t = document.createElement('div');
    t.className = 'wk-toast toast-' + type;
    t.innerHTML =
      '<span class="wk-toast-icon">' + (ICONS[type] || ICONS.info) + '</span>' +
      '<span>' + message + '</span>';

    getContainer().appendChild(t);

    requestAnimationFrame(function () {
      requestAnimationFrame(function () { t.classList.add('show'); });
    });

    setTimeout(function () {
      t.classList.remove('show');
      t.classList.add('hide');
      setTimeout(function () { t.remove(); }, 320);
    }, duration);
  };
}());


/* ═══════════════════════════════════════════════════════════════════════
   MOBILE BOTTOM NAV POPUPS
   ═══════════════════════════════════════════════════════════════════════ */
(function () {
  var _open = null;

  window.toggleMobPopup = function (id) {
    if (_open === id) { closeMobPopups(); return; }
    document.querySelectorAll('.mob-popup').forEach(function (p) { p.classList.remove('open'); });
    _open = id;
    var popup = document.getElementById('mob-popup-' + id);
    if (popup) popup.classList.add('open');
    var bd = document.getElementById('mob-backdrop');
    if (bd) bd.classList.remove('hidden');
  };

  window.closeMobPopups = function () {
    document.querySelectorAll('.mob-popup').forEach(function (p) { p.classList.remove('open'); });
    var bd = document.getElementById('mob-backdrop');
    if (bd) bd.classList.add('hidden');
    _open = null;
  };

  document.addEventListener('DOMContentLoaded', function () {
    var bd = document.getElementById('mob-backdrop');
    if (!bd) return;
    bd.addEventListener('click', closeMobPopups);
    bd.addEventListener('touchend', function (e) { e.preventDefault(); closeMobPopups(); });
  });
}());
