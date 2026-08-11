/* ==========================================================================
   Previa — minimal client behaviour.

   Everything here is either an Alpine component factory or a small piece of
   glue that Alpine/HTMX/CSS genuinely cannot do:
     - theme persistence
     - focus/scroll bookkeeping around overlays
     - the Google Maps integration and its offline fallback
   No other client-side framework is loaded.
   ========================================================================== */

(function () {
  'use strict';

  // -- Theme -----------------------------------------------------------------
  // The initial value is applied by an inline script in <head> so there is no
  // flash; this store keeps it in sync afterwards.
  document.addEventListener('alpine:init', function () {
    Alpine.store('theme', {
      value: document.documentElement.getAttribute('data-theme') || 'light',
      toggle: function () {
        this.value = this.value === 'dark' ? 'light' : 'dark';
        document.documentElement.setAttribute('data-theme', this.value);
        try {
          localStorage.setItem('previa-theme', this.value);
        } catch (e) {
          /* private mode — the choice simply will not persist */
        }
      },
    });

    // Country preference. A manual choice always wins over geolocation and is
    // remembered, so the browser is never asked for permission twice.
    Alpine.store('country', {
      code: document.documentElement.getAttribute('data-country') || 'EE',
      setManual: function (code) {
        this.code = code;
        try {
          localStorage.setItem('previa-country', code);
          localStorage.setItem('previa-country-source', 'manual');
        } catch (e) {}
      },
      isManual: function () {
        try {
          return localStorage.getItem('previa-country-source') === 'manual';
        } catch (e) {
          return false;
        }
      },
    });
  });

  // -- Card image carousel ---------------------------------------------------
  // Steps through a card's photographs in place. Wraps in both directions so
  // the arrows never dead-end, and holds only an index — the images are all
  // in the DOM already.
  window.previaCardGallery = function (count) {
    return {
      i: 0,
      total: count || 1,
      next: function () {
        this.i = (this.i + 1) % this.total;
      },
      prev: function () {
        this.i = (this.i - 1 + this.total) % this.total;
      },
    };
  };

  // -- Header ----------------------------------------------------------------
  window.previaHeader = function () {
    return {
      scrolled: window.scrollY > 8,
      // True once the header has actually left the viewport. Kept separate
      // from `scrolled` (which trips at 8px, purely for the shadow) because
      // the floating menu button must not appear while the real header is
      // still on screen — two triggers for the same menu reads as a bug.
      pastHeader: false,
      menuOpen: false,
      lastFocus: null,

      init: function () {
        var self = this;
        var threshold = function () {
          var h = self.$el ? self.$el.offsetHeight : 0;
          return h > 0 ? h : 68;
        };
        var apply = function () {
          var y = window.scrollY;
          self.scrolled = y > 8;
          self.pastHeader = y > threshold();
        };
        apply();
        window.addEventListener('scroll', apply, { passive: true });
        window.addEventListener('resize', apply, { passive: true });
      },

      openMenu: function () {
        this.lastFocus = document.activeElement;
        this.menuOpen = true;
      },

      closeMenu: function () {
        this.menuOpen = false;
        // Return focus to whatever opened the drawer.
        if (this.lastFocus && this.lastFocus.focus) {
          var el = this.lastFocus;
          setTimeout(function () {
            el.focus();
          }, 0);
        }
      },
    };
  };

  // -- Generic dropdown ------------------------------------------------------
  window.previaMenu = function () {
    return {
      open: false,
      toggle: function () {
        this.open = !this.open;
      },
      close: function () {
        this.open = false;
      },
    };
  };

  // -- Search layout ---------------------------------------------------------
  // Controls the filter panel in both of its forms: a collapsible sidebar on
  // large screens and an off-canvas drawer below 1024px. The collapsed
  // preference persists; the drawer never does.
  window.previaSearch = function () {
    return {
      collapsed: false,
      drawer: false,

      init: function () {
        try {
          this.collapsed = localStorage.getItem('previa-filters-collapsed') === '1';
        } catch (e) {}
      },

      collapse: function () {
        this.collapsed = true;
        this.persist();
      },

      expand: function () {
        this.collapsed = false;
        this.persist();
      },

      persist: function () {
        try {
          localStorage.setItem('previa-filters-collapsed', this.collapsed ? '1' : '0');
        } catch (e) {}
      },

      openDrawer: function () {
        this.lastFocus = document.activeElement;
        this.drawer = true;
      },

      closeDrawer: function () {
        this.drawer = false;
        if (this.lastFocus && this.lastFocus.focus) {
          var el = this.lastFocus;
          setTimeout(function () {
            el.focus();
          }, 0);
        }
      },
    };
  };

  // -- Password field --------------------------------------------------------
  // Show/hide plus a strength read-out. The score is a simple, honest measure
  // of length and character variety — it is guidance for the user, never a
  // substitute for the server-side rules.
  window.previaPassword = function () {
    return {
      show: false,
      value: '',

      score: function () {
        var v = this.value;
        if (!v) return 0;
        var s = Math.min(v.length, 12) * 5; // up to 60 for length
        if (/[A-Z]/.test(v)) s += 12;
        if (/[a-z]/.test(v)) s += 8;
        if (/[0-9]/.test(v)) s += 10;
        if (/[^A-Za-z0-9]/.test(v)) s += 10;
        return Math.max(8, Math.min(100, s));
      },

      level: function () {
        var s = this.score();
        if (this.value.length < 8) return 'weak';
        if (s < 70) return 'fair';
        if (s < 90) return 'good';
        return 'strong';
      },

      label: function () {
        return {
          weak: 'Too short — use at least 8 characters',
          fair: 'Fair — add a capital letter or a number',
          good: 'Good password',
          strong: 'Strong password',
        }[this.level()];
      },
    };
  };

  // -- Add-listing wizard ----------------------------------------------------
  // Autosaves each step to localStorage as the user types, so a listing can be
  // abandoned and resumed. The real backend will persist drafts server-side;
  // the indicator states are identical either way.
  window.previaWizard = function (step, total) {
    return {
      state: 'idle', // idle | saving | saved
      step: step,
      total: total,
      timer: null,

      init: function () {
        // Restore anything typed on this step previously.
        try {
          var raw = localStorage.getItem('previa-draft');
          if (raw) this.draft = JSON.parse(raw);
        } catch (e) {
          this.draft = {};
        }
      },

      // Called on every input; debounced so we do not thrash storage.
      touch: function () {
        var self = this;
        this.state = 'saving';
        clearTimeout(this.timer);
        this.timer = setTimeout(function () {
          self.save();
        }, 600);
      },

      save: function () {
        var data = {};
        document.querySelectorAll('.wizard-card input, .wizard-card select, .wizard-card textarea')
          .forEach(function (el, i) {
            var key = el.id || el.name || 'f' + i;
            data[key] = el.type === 'checkbox' || el.type === 'radio' ? el.checked : el.value;
          });

        try {
          var existing = JSON.parse(localStorage.getItem('previa-draft') || '{}');
          existing['step' + this.step] = data;
          existing.lastStep = this.step;
          existing.updatedAt = new Date().toISOString();
          localStorage.setItem('previa-draft', JSON.stringify(existing));
        } catch (e) {
          /* storage unavailable — the indicator still reflects the attempt */
        }

        this.state = 'saved';
        var self = this;
        setTimeout(function () {
          if (self.state === 'saved') self.state = 'idle';
        }, 2500);
      },
    };
  };

  // -- Map -------------------------------------------------------------------
  //
  // previaMap() owns the map surface. Rendering goes through a small provider
  // adapter so the underlying library is an implementation detail: today it is
  // Leaflet with OpenStreetMap tiles, and the client's Google Maps build can
  // replace `provider` alone without touching markup, cards, filters or popups.
  //
  // The public surface used by templates is:
  //   points, ready, failed, panelOpen, mapVisible
  //   highlight(id), select(id), close(), retry()
  //   zoomIn(), zoomOut(), fitAll(), togglePanel(), toggleMap()
  window.previaMap = function (config) {
    return {
      points: config.points || [],
      ready: false,
      failed: false,
      active: null, // id under the cursor
      selected: null, // id whose popup is open
      panelOpen: true,
      mapVisible: false, // small screens: list or map
      provider: null,

      init: function () {
        var self = this;
        // Leaflet needs the container to have a size; on the split view the
        // pane is laid out by grid, so wait a frame before measuring.
        requestAnimationFrame(function () {
          self.build();
        });

        // Results are swapped by HTMX; rebuild markers from the new payload.
        window.addEventListener('previa-results-updated', function () {
          self.refreshFromDom();
        });
      },

      build: function () {
        var el = this.$refs.canvas;
        if (!el || !window.L) {
          this.failed = true;
          return;
        }
        try {
          this.provider = createLeafletProvider(el, config, this);
          this.ready = true;
          this.render();
        } catch (e) {
          this.failed = true;
        }
      },

      retry: function () {
        this.failed = false;
        this.ready = false;
        if (this.provider) {
          this.provider.destroy();
          this.provider = null;
        }
        this.build();
      },

      render: function () {
        if (!this.provider) return;
        this.provider.setPoints(this.points);
        this.provider.fitAll();
      },

      // After an HTMX swap the result cards carry the new ids; read them back
      // so the map matches the list without another round trip.
      refreshFromDom: function () {
        var ids = [].slice
          .call(document.querySelectorAll('[data-property-id]'))
          .map(function (n) {
            return n.getAttribute('data-property-id');
          });
        if (!ids.length || !this.provider) return;
        var keep = this.points.filter(function (p) {
          return ids.indexOf(p.id) !== -1;
        });
        if (keep.length) {
          this.provider.setPoints(keep);
          this.provider.fitAll();
        }
      },

      highlight: function (id) {
        this.active = id;
        if (this.provider) this.provider.highlight(id);
      },

      // Centre the map on one listing. Separate from select() because it
      // moves the viewport as well as opening the popup — select() is also
      // fired by marker clicks, where moving the map under the cursor would
      // be disorienting.
      locate: function (id) {
        this.selected = id;
        if (this.provider && this.provider.locate) this.provider.locate(id);
      },

      // Called from a marker click and from a result card.
      select: function (id) {
        this.selected = this.selected === id ? null : id;
        if (this.provider) this.provider.select(this.selected);
        if (this.selected) this.scrollCardIntoView(this.selected);
      },

      close: function () {
        this.selected = null;
        if (this.provider) this.provider.select(null);
      },

      scrollCardIntoView: function (id) {
        var card = document.querySelector('[data-property-id="' + id + '"]');
        if (card && card.scrollIntoView) {
          card.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
        }
      },

      zoomIn: function () {
        if (this.provider) this.provider.zoomIn();
      },
      zoomOut: function () {
        if (this.provider) this.provider.zoomOut();
      },
      fitAll: function () {
        if (this.provider) this.provider.fitAll();
      },

      togglePanel: function () {
        this.panelOpen = !this.panelOpen;
        this.invalidate();
      },

      toggleMap: function () {
        this.mapVisible = !this.mapVisible;
        this.invalidate();
      },

      // The pane changes size when a panel collapses; Leaflet must re-measure
      // or it renders grey bands where tiles were never requested.
      invalidate: function () {
        var self = this;
        setTimeout(function () {
          if (self.provider) self.provider.invalidate();
        }, 260);
      },
    };
  };

  // --- Leaflet provider -------------------------------------------------------
  // Everything Leaflet-specific lives here. A Google Maps provider would expose
  // the same six methods and nothing above would change.
  function createLeafletProvider(el, config, host) {
    var map = L.map(el, {
      center: [config.lat || 0, config.lng || 0],
      zoom: config.zoom || 11,
      zoomControl: false, // Previa supplies its own controls
      attributionControl: true,
      scrollWheelZoom: true,
    });

    var tiles = L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
      maxZoom: 19,
      // Required by the OpenStreetMap tile usage policy.
      attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
    });

    // If tiles cannot be fetched at all, surface the error state rather than
    // leaving a blank grey square.
    var errors = 0;
    tiles.on('tileerror', function () {
      errors++;
      if (errors > 6) host.failed = true;
    });
    tiles.addTo(map);

    var markers = {};
    var group = L.layerGroup().addTo(map);

    function priceMarker(p, isActive) {
      return L.divIcon({
        className: 'map-pin-wrap',
        html:
          '<span class="map-pin' +
          (p.featured ? ' is-featured' : '') +
          (isActive ? ' is-active' : '') +
          '">' +
          escapeHtml(p.price) +
          '</span>',
        iconSize: null,
        iconAnchor: [0, 0],
      });
    }

    function popupHtml(p) {
      var img = p.images && p.images.length
        ? '<img src="' + p.images[0] + '" alt="" width="300" height="188" loading="lazy">'
        : '';
      var rooms = p.rooms ? '<span>' + p.rooms + ' rooms</span>' : '';
      return (
        '<div class="map-popup__card">' +
        '<a class="map-popup__media" href="' + p.url + '">' + img + '</a>' +
        '<div class="map-popup__body">' +
        '<p class="map-popup__price">' + escapeHtml(p.full) + '</p>' +
        '<a class="map-popup__title" href="' + p.url + '">' + escapeHtml(p.title) + '</a>' +
        '<p class="map-popup__meta">' + escapeHtml(p.city) + '</p>' +
        '<p class="map-popup__facts">' +
        '<span>' + escapeHtml(p.type) + '</span>' + rooms +
        '<span>' + escapeHtml(p.area) + '</span></p>' +
        '<a class="btn btn--primary btn--sm btn--block" href="' + p.url + '">View Property</a>' +
        '</div></div>'
      );
    }

    function setPoints(points) {
      group.clearLayers();
      markers = {};
      points.forEach(function (p) {
        var m = L.marker([p.lat, p.lng], {
          icon: priceMarker(p, false),
          riseOnHover: true,
          keyboard: true,
          alt: p.title + ' — ' + p.full,
        });
        m.bindPopup(popupHtml(p), {
          className: 'map-popup-shell',
          maxWidth: 300,
          minWidth: 260,
          closeButton: true,
          autoPanPadding: [24, 24],
        });
        m.on('click', function () {
          host.select(p.id);
        });
        m.on('mouseover', function () {
          host.highlight(p.id);
        });
        m.on('mouseout', function () {
          host.highlight(null);
        });
        m.addTo(group);
        markers[p.id] = { marker: m, point: p };
      });
      host.points = points;
    }

    function highlight(id) {
      Object.keys(markers).forEach(function (key) {
        var entry = markers[key];
        var on = key === id;
        entry.marker.setIcon(priceMarker(entry.point, on));
        if (on) entry.marker.setZIndexOffset(1000);
        else entry.marker.setZIndexOffset(0);
      });
    }

    function select(id) {
      if (!id) {
        map.closePopup();
        return;
      }
      var entry = markers[id];
      if (!entry) return;
      entry.marker.openPopup();
    }

    function fitAll() {
      var latlngs = Object.keys(markers).map(function (k) {
        return markers[k].marker.getLatLng();
      });
      if (!latlngs.length) return;
      if (latlngs.length === 1) {
        map.setView(latlngs[0], Math.max(config.zoom || 13, 13));
        return;
      }
      map.fitBounds(L.latLngBounds(latlngs), { padding: [48, 48], maxZoom: 15 });
    }

    map.on('popupclose', function () {
      host.selected = null;
    });

    // Centre the map on one listing and open its popup. Used by the locate
    // button on each result card, which answers "where is this one?" without
    // making the visitor hunt for the pin.
    function locate(id) {
      var entry = markers[id];
      if (!entry) return;
      var target = entry.marker.getLatLng();
      var zoom = Math.max(map.getZoom(), 14);
      if (map.flyTo) map.flyTo(target, zoom, { duration: 0.6 });
      else map.setView(target, zoom);
      highlight(id);
      entry.marker.openPopup();
    }

    return {
      setPoints: setPoints,
      highlight: highlight,
      select: select,
      locate: locate,
      fitAll: fitAll,
      zoomIn: function () {
        map.zoomIn();
      },
      zoomOut: function () {
        map.zoomOut();
      },
      invalidate: function () {
        map.invalidateSize();
      },
      destroy: function () {
        map.remove();
      },
    };
  }

  function escapeHtml(s) {
    return String(s == null ? '' : s).replace(/[&<>"']/g, function (c) {
      return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c];
    });
  }

  // -- Toast -----------------------------------------------------------------
  // Used to confirm mock actions (saved search, restart simulation, and so on).
  window.previaToast = function (message, tone) {
    var stack = document.getElementById('toasts');
    if (!stack) return;

    var el = document.createElement('div');
    el.className = 'toast' + (tone ? ' toast--' + tone : '');
    el.setAttribute('role', 'status');

    var icon =
      tone === 'error'
        ? 'M12 8v5M12 16h.01'
        : tone === 'warning'
        ? 'M12 9v4M12 17h.01'
        : 'm8.5 12 2.5 2.5 4.5-5';

    el.innerHTML =
      '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" ' +
      'stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">' +
      '<circle cx="12" cy="12" r="9"/><path d="' + icon + '"/></svg>' +
      '<span></span>';
    el.querySelector('span').textContent = message;

    stack.appendChild(el);
    setTimeout(function () {
      el.style.opacity = '0';
      el.style.transform = 'translateY(6px)';
      setTimeout(function () {
        el.remove();
      }, 250);
    }, 3600);
  };

  // -- Geolocation -----------------------------------------------------------
  // Only ever runs on an explicit user action, never on page load.
  document.addEventListener('previa-locate', function () {
    if (!navigator.geolocation) {
      window.previaToast('Location is not available in this browser. Choose a country manually.', 'warning');
      return;
    }
    navigator.geolocation.getCurrentPosition(
      function (pos) {
        window.location.href =
          '/set-country?lat=' + pos.coords.latitude + '&lng=' + pos.coords.longitude;
      },
      function () {
        window.previaToast('Location permission denied. Choose a country from the list instead.', 'warning');
      },
      { timeout: 8000, maximumAge: 600000 }
    );
  });

  // -- HTMX glue -------------------------------------------------------------
  document.addEventListener('DOMContentLoaded', function () {
    if (!window.htmx) return;

    // Keep the address bar in step with filtered results so a search is
    // shareable and the back button behaves.
    document.body.addEventListener('htmx:pushedIntoHistory', function () {
      window.dispatchEvent(new CustomEvent('previa-results-updated'));
    });

    document.body.addEventListener('htmx:afterSwap', function (e) {
      if (e.detail.target && e.detail.target.id === 'results') {
        window.dispatchEvent(new CustomEvent('previa-results-updated'));
      }
    });

    // Alpine adds DOM at runtime (modals, x-if blocks, teleported drawers).
    // HTMX only wires up elements it has processed, so anything Alpine inserts
    // would have inert hx-* attributes without this. Process new subtrees as
    // they appear.
    var observer = new MutationObserver(function (mutations) {
      mutations.forEach(function (m) {
        m.addedNodes.forEach(function (node) {
          if (node.nodeType !== 1) return;
          var hasHx =
            node.hasAttribute('hx-post') ||
            node.hasAttribute('hx-get') ||
            node.querySelector('[hx-post],[hx-get],[hx-trigger]');
          if (hasHx) window.htmx.process(node);
        });
      });
    });
    observer.observe(document.body, { childList: true, subtree: true });

    // Surface a network failure instead of leaving the UI silently stale.
    document.body.addEventListener('htmx:sendError', function () {
      window.previaToast('Network error. Check your connection and try again.', 'error');
    });
    document.body.addEventListener('htmx:responseError', function (e) {
      if (e.detail.xhr && e.detail.xhr.status >= 500) {
        window.previaToast('Something went wrong on our side. Please try again.', 'error');
      }
    });
  });
})();
