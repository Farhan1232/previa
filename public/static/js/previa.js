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
  //
  // The dots also scrub. Press and hold one, and the active dot grows; drag
  // left or right and the picture follows the finger, the active dot moving
  // with it; let go and the dot settles back. Built on Pointer Events, so one
  // code path serves mouse, touch and pen.
  //
  // Distance per picture. Comfortably above the threshold below, so a drag
  // that has been recognised always lands on a neighbouring image first.
  var DOT_DRAG_STEP = 22;
  // Movement before a press counts as a drag rather than a click.
  var DOT_DRAG_SLOP = 6;

  window.previaCardGallery = function (count) {
    return {
      i: 0,
      total: count || 1,

      // Gesture state. `holding` is a press, `dragging` a press that has been
      // confirmed horizontal; only the second one takes the pointer captive
      // and starts suppressing the browser's own handling.
      holding: false,
      dragging: false,
      dragged: false, // a real drag just finished — the click it produces is not a choice
      pointerId: null,
      startX: 0,
      startY: 0,
      startIndex: 0,

      next: function () {
        this.i = (this.i + 1) % this.total;
      },
      prev: function () {
        this.i = (this.i - 1 + this.total) % this.total;
      },
      // Jump straight to one image, from a dot.
      go: function (n) {
        if (n >= 0 && n < this.total) this.i = n;
      },

      dotDown: function (e) {
        // Left button / touch / pen only: a right-click must still open the
        // context menu rather than starting a gesture.
        if (e.button > 0) return;
        this.pointerId = e.pointerId;
        this.startX = e.clientX;
        this.startY = e.clientY;
        this.startIndex = this.i;
        this.holding = true;
        this.dragging = false;
        this.dragged = false;
      },

      dotMove: function (e) {
        if (this.pointerId === null || e.pointerId !== this.pointerId) return;
        var dx = e.clientX - this.startX;
        var dy = e.clientY - this.startY;

        if (!this.dragging) {
          // Not a drag until it has travelled far enough *and* travelled
          // further across than down. The second test is what leaves a
          // vertical swipe to the page: until it passes, nothing is captured
          // and no default is prevented, so the browser scrolls as usual.
          if (Math.abs(dx) < DOT_DRAG_SLOP || Math.abs(dx) <= Math.abs(dy)) return;
          this.dragging = true;
          this.dragged = true;
          try {
            e.currentTarget.setPointerCapture(this.pointerId);
          } catch (err) {
            /* capture is an optimisation — the gesture still works without it */
          }
        }

        // Now that the gesture is ours, stop the browser turning it into a
        // scroll, a text selection or a drag-and-drop of the image.
        if (e.cancelable) e.preventDefault();

        var n = this.startIndex + Math.round(dx / DOT_DRAG_STEP);
        if (n < 0) n = 0;
        if (n > this.total - 1) n = this.total - 1;
        this.i = n;
      },

      // pointerup, pointercancel and lostpointercapture all land here, so the
      // dot cannot be left enlarged by a gesture that ended somewhere odd —
      // dragged off the window, interrupted by a system menu, or stolen.
      dotUp: function (e) {
        if (this.pointerId === null) return;

        // A touch pointer is implicitly captured by whatever it landed on —
        // one of the dots. Taking the capture for the strip therefore makes
        // that dot fire lostpointercapture, and the event bubbles here. It
        // means the drag has just begun, not that it has ended, so only a loss
        // reported by the strip itself counts.
        if (e && e.type === 'lostpointercapture' && e.target !== e.currentTarget) return;
        if (e && e.pointerId === this.pointerId && e.currentTarget.hasPointerCapture) {
          try {
            if (e.currentTarget.hasPointerCapture(this.pointerId)) {
              e.currentTarget.releasePointerCapture(this.pointerId);
            }
          } catch (err) {}
        }
        this.pointerId = null;
        this.holding = false;
        this.dragging = false;
      },

      // A dot's click. After a real drag the click is the tail of the gesture,
      // not a choice of picture, so it is swallowed — and because these dots
      // sit over the card's link, swallowing it is also what stops a drag from
      // opening the property.
      dotClick: function (e, n) {
        if (this.dragged) {
          this.dragged = false;
          e.preventDefault();
          e.stopPropagation();
          return;
        }
        this.go(n);
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

  // -- Location autocomplete -------------------------------------------------
  // One Google-Maps-style Location box, shared by the homepage search, the
  // filter sidebar, the map filter and the add-listing form.
  //
  // The options are server-rendered buttons; this filters them, drives keyboard
  // selection, and fires an optional expression after a pick so a caller can
  // move a map or fill its own fields. Replacing the mock with Places means
  // replacing where the options come from — nothing here changes shape.
  window.previaLocation = function () {
    return {
      open: false,
      q: '',
      empty: false,
      active: -1,
      hasValue: false,
      options: [],

      init: function () {
        var self = this;
        var list = this.$refs.list;
        this.options = list
          ? Array.prototype.map.call(list.querySelectorAll('[data-label]'), function (el) {
              return { el: el, hay: (el.getAttribute('data-label') || '').toLowerCase() };
            })
          : [];
        this.q = this.$refs.input ? this.$refs.input.value : '';
        this.hasValue = !!this.q;

        // Fit the placeholder to the field it actually got.
        //
        // A placeholder cannot ellipsis, so one that is a few pixels too long
        // is simply sliced — which is what happened in the filter sidebar:
        // "Country, city or address" fitted the width by 25px in one browser
        // and overflowed it in another. Rather than guessing a string short
        // enough for every font, zoom level and container, the field measures
        // the candidates and takes the longest that fits.
        this.fitPlaceholder();
        window.addEventListener('resize', function () {
          clearTimeout(self.fitTimer);
          self.fitTimer = setTimeout(function () { self.fitPlaceholder(); }, 120);
        });
        // Webfonts change the metrics after first paint, so measure again once
        // they have settled.
        if (document.fonts && document.fonts.ready) {
          document.fonts.ready.then(function () { self.fitPlaceholder(); });
        }
      },

      fitPlaceholder: function () {
        var input = this.$refs.input;
        if (!input) return;

        // The caller's own wording first, then progressively shorter fallbacks.
        var full = input.getAttribute('data-placeholder-full') || input.placeholder;
        var options = [full, 'Country, city or address', 'City or address', 'Location'];

        var cs = window.getComputedStyle(input);
        var space = input.clientWidth
          - parseFloat(cs.paddingLeft) - parseFloat(cs.paddingRight);
        if (!(space > 0)) return;

        // Canvas text metrics rather than a probe element: no layout, no
        // reflow, and it accounts for the real font once it has loaded.
        var canvas = previaLocation._canvas ||
          (previaLocation._canvas = document.createElement('canvas'));
        var ctx = canvas.getContext('2d');
        ctx.font = cs.font || (cs.fontWeight + ' ' + cs.fontSize + ' ' + cs.fontFamily);

        for (var i = 0; i < options.length; i++) {
          // A couple of pixels of slack so a rounding difference cannot clip it.
          if (ctx.measureText(options[i]).width <= space - 2) {
            input.placeholder = options[i];
            return;
          }
        }
        input.placeholder = options[options.length - 1];
      },

      // Matches a word prefix anywhere in the label, so "Tallinn" finds
      // "Kalamaja, Tallinn, Estonia" as well as "Tallinn, Estonia", while
      // "all" does not drag in every label containing those letters.
      matches: function (hay, needle) {
        if (hay.indexOf(needle) === 0) return true;
        var parts = hay.split(/[\s,\-]+/);
        for (var i = 0; i < parts.length; i++) {
          if (parts[i].indexOf(needle) === 0) return true;
        }
        return false;
      },

      filter: function () {
        this.q = this.$refs.input ? this.$refs.input.value : '';
        this.hasValue = !!this.q;
        var needle = this.q.trim().toLowerCase();
        var shown = 0;
        for (var i = 0; i < this.options.length; i++) {
          // Cap the visible list: with no query at all, showing every address
          // in the catalogue is noise, so only the broadest entries lead.
          var hit = needle ? this.matches(this.options[i].hay, needle) : shown < 8;
          this.options[i].el.hidden = !hit;
          if (hit) shown++;
        }
        this.empty = needle !== '' && shown === 0;
        this.open = true;
        this.active = -1;
      },

      visible: function () {
        return this.options.filter(function (o) { return !o.el.hidden; });
      },

      move: function (step) {
        if (!this.open) this.filter();
        var vis = this.visible();
        if (!vis.length) return;
        this.active += step;
        if (this.active < 0) this.active = vis.length - 1;
        if (this.active >= vis.length) this.active = 0;
        // Track the index within the full option list, which is what the
        // aria-activedescendant ids and the :class bindings are keyed on.
        this.active = this.options.indexOf(vis[this.active]);
        vis = this.options[this.active];
        if (vis && vis.el.scrollIntoView) vis.el.scrollIntoView({ block: 'nearest' });
      },

      // Enter takes the highlighted row if there is one. With nothing
      // highlighted it lets the keypress through so the form submits, which is
      // what someone who typed a place and hit Enter expects.
      pick: function (ev) {
        if (!this.open || this.active < 0) return;
        ev.preventDefault();
        this.choose(this.options[this.active].el);
      },

      choose: function (el) {
        var label = el.getAttribute('data-label') || '';
        if (this.$refs.input) {
          this.$refs.input.value = label;
          // Let HTMX and any listening form see a real change event.
          this.$refs.input.dispatchEvent(new Event('change', { bubbles: true }));
        }
        this.q = label;
        this.hasValue = true;
        this.close();
        // The whole structured place, so a listener can move a map and fill
        // address fields without looking anything up again.
        this.$dispatch('previa-location-picked', {
          label: label,
          kind: el.getAttribute('data-kind') || '',
          lat: parseFloat(el.getAttribute('data-lat')) || 0,
          lng: parseFloat(el.getAttribute('data-lng')) || 0,
          countryCode: el.getAttribute('data-country') || '',
          city: el.getAttribute('data-city') || '',
          district: el.getAttribute('data-district') || '',
          address: el.getAttribute('data-address') || '',
        });
      },

      clear: function () {
        if (this.$refs.input) {
          this.$refs.input.value = '';
          this.$refs.input.dispatchEvent(new Event('change', { bubbles: true }));
          this.$refs.input.focus();
        }
        this.q = '';
        this.hasValue = false;
        this.empty = false;
        this.filter();
      },

      close: function () {
        this.open = false;
        this.active = -1;
      },
    };
  };

  // -- Property-type multi-select --------------------------------------------
  //
  // Property type is a multiple choice: House, Modular house and Panelized
  // house can all be on at once and results are the union of them.
  //
  // The checkboxes are native and are the source of truth — this only enforces
  // the relationship between "Any type" and the rest, which HTML has no way to
  // express on its own:
  //
  //   * Any type on   -> no specific type is on (no filter at all)
  //   * a type goes on -> Any type goes off
  //   * the last type goes off -> Any type comes back
  //
  // "Any type" carries no `name`, so whatever it does here it is never
  // submitted as a category.
  window.previaTypeFilter = function () {
    return {
      init: function () {
        // Reconcile once on load, so a URL like ?property_type=house arrives
        // with Any type already off even though the server rendered both.
        this.reconcile();
      },

      // $root is the component's own element. $el would be whichever element
      // the expression fired on — the checkbox, not the grid — which returned
      // null for every lookup and left Any type permanently ticked.
      grid: function () {
        return this.$root;
      },

      anyBox: function () {
        var g = this.grid();
        return g ? g.querySelector('input[data-any]') : null;
      },

      typeBoxes: function () {
        var g = this.grid();
        return g
          ? Array.prototype.slice.call(g.querySelectorAll('input[name="property_type"]'))
          : [];
      },

      // Any type clears every specific choice.
      pickAny: function () {
        var any = this.anyBox();
        if (!any) return;
        if (any.checked) {
          this.typeBoxes().forEach(function (b) { b.checked = false; });
        } else {
          // Unticking Any type on its own would leave nothing selected at all,
          // which is the same filter with a worse-looking control — so it
          // stays on until a specific type takes over.
          any.checked = true;
        }
        this.announce();
      },

      // A specific type was toggled.
      pickType: function () {
        this.reconcile();
        this.announce();
      },

      reconcile: function () {
        var any = this.anyBox();
        if (!any) return;
        var on = this.typeBoxes().filter(function (b) { return b.checked; });
        any.checked = on.length === 0;
      },

      // Ticking Any type clears the others by script, which fires no change
      // event of its own — so HTMX would never hear that the filter changed.
      // One synthetic event on the grid tells it, and the homepage picker
      // listens to the same event to update its summary.
      announce: function () {
        var g = this.grid();
        if (g) g.dispatchEvent(new Event('change', { bubbles: true }));
      },
    };
  };

  // -- Homepage property-type picker -----------------------------------------
  //
  // Wraps the shared checkbox grid in a dropdown, because the hero has room
  // for one field-sized control rather than a twelve-tile grid. The trigger
  // summarises the selection; the checkboxes inside are the real inputs and
  // submit with the form.
  window.previaTypePicker = function () {
    return {
      open: false,
      summary: 'Any type',

      init: function () {
        var self = this;
        this.refresh();
        // The grid inside manages its own Any-type bookkeeping; this only
        // needs to re-read the result afterwards.
        this.$el.addEventListener('change', function () { self.refresh(); });
      },

      refresh: function () {
        var on = Array.prototype.slice
          .call(this.$el.querySelectorAll('input[name="property_type"]'))
          .filter(function (b) { return b.checked; });

        if (on.length === 0) {
          this.summary = 'Any type';
        } else if (on.length === 1) {
          // The label lives in the tile beside the input.
          var label = on[0].closest('.type-check').querySelector('.type-check__box span');
          this.summary = label ? label.textContent.trim() : '1 type';
        } else {
          this.summary = on.length + ' types';
        }
      },
    };
  };

  // -- Market picker ---------------------------------------------------------
  // A dropdown over the full country list, so it needs a filter.
  //
  // The countries are server-rendered links; this only hides and shows them.
  // That keeps the control usable with JavaScript off and means the filter
  // never has to agree with a second copy of the list held in the client.
  window.previaMarketPicker = function () {
    return {
      open: false,
      q: '',
      empty: false,
      items: [],   // { el, hay } for every country row
      groups: [],  // the "With listings" / "All countries" captions
      cursor: -1,  // index into the visible rows, for arrow-key navigation

      init: function () {
        var list = this.$refs.list;
        if (!list) return;
        this.items = Array.prototype.map.call(
          list.querySelectorAll('[data-market-name]'),
          function (el, i) {
            var name = (el.getAttribute('data-market-name') || '').toLowerCase();
            return {
              el: el,
              order: i,   // the curated order, restored when the query clears
              name: name,
              code: (el.getAttribute('data-market-code') || '').toLowerCase(),
              // Split on spaces and punctuation so "Herzegovina" finds Bosnia
              // and "Kingdom" finds the United Kingdom.
              words: name.split(/[\s\-'’(),.]+/).filter(Boolean),
            };
          }
        );
        this.groups = Array.prototype.slice.call(list.querySelectorAll('[data-market-group]'));
      },

      toggle: function () {
        this.open = !this.open;
        if (this.open) {
          // Focus the field so typing filters immediately. $nextTick because
          // x-show has not revealed the dropdown yet at this point.
          var self = this;
          this.$nextTick(function () { if (self.$refs.q) self.$refs.q.focus(); });
        }
      },

      close: function () {
        if (!this.open) return;
        this.open = false;
        this.clear();
        // Return focus to the trigger rather than leaving it on a hidden field.
        if (this.$refs.trigger) this.$refs.trigger.focus();
      },

      // How well a country answers the query. Lower is better; -1 is no match.
      //
      // The ranking exists because filtering alone put Estonia above Spain for
      // "ES" — Estonia is a seeded market so it came first in the list, and its
      // name starts with those letters. But ES *is* Spain's country code, and
      // that is the stronger signal, so an exact code match now outranks
      // everything else.
      //
      //   0  the query is exactly the country code   ES -> Spain
      //   1  the query is exactly the country name   Spain -> Spain
      //   2  the code starts with the query
      //   3  the name starts with the query          Ger -> Germany
      //   4  a word inside the name starts with it   Kingdom -> United Kingdom
      //   5  the name merely contains it             ustri -> Austria
      score: function (item, needle) {
        if (item.code === needle) return 0;
        if (item.name === needle) return 1;
        if (item.code.indexOf(needle) === 0) return 2;
        if (item.name.indexOf(needle) === 0) return 3;
        for (var w = 0; w < item.words.length; w++) {
          if (item.words[w].indexOf(needle) === 0) return 4;
        }
        if (item.name.indexOf(needle) !== -1) return 5;
        return -1;
      },

      filter: function () {
        var needle = this.q.trim().toLowerCase();
        var list = this.$refs.list;
        var i;

        if (!needle) {
          // No query: every country, back in its curated order — the seeded
          // markets first, then the rest of the world alphabetically.
          var restore = this.items.slice().sort(function (a, b) { return a.order - b.order; });
          for (i = 0; i < restore.length; i++) {
            restore[i].el.hidden = false;
            if (list) list.appendChild(restore[i].el);
          }
          // Put the group captions back where they belong.
          for (var g0 = 0; g0 < this.groups.length; g0++) {
            this.groups[g0].hidden = false;
          }
          this.regroup();
          this.empty = false;
          this.cursor = -1;
          if (list) list.scrollTop = 0;
          return;
        }

        var hits = [];
        for (i = 0; i < this.items.length; i++) {
          var s = this.score(this.items[i], needle);
          this.items[i].el.hidden = s < 0;
          if (s >= 0) hits.push({ item: this.items[i], score: s });
        }

        // Best first, and within a score the curated order, so results never
        // jump around between two keystrokes that score the same.
        hits.sort(function (a, b) {
          return a.score - b.score || a.item.order - b.item.order;
        });
        for (i = 0; i < hits.length; i++) {
          if (list) list.appendChild(hits[i].item.el);
        }

        // The group captions only make sense while the whole list is showing;
        // once a query is narrowing things they just fragment the results.
        for (var g = 0; g < this.groups.length; g++) {
          this.groups[g].hidden = true;
        }

        this.empty = hits.length === 0;
        this.cursor = -1;
        if (list) list.scrollTop = 0;
      },

      // Move the captions back above the runs they head, after a filter has
      // reordered the rows underneath them.
      regroup: function () {
        var list = this.$refs.list;
        if (!list || this.groups.length < 2) return;
        var rows = list.querySelectorAll('[data-market-name]');
        // The first caption leads the list; the second leads the first row
        // that is not one of the seeded markets.
        list.insertBefore(this.groups[0], rows[0]);
        for (var i = 0; i < rows.length; i++) {
          if (rows[i].getAttribute('data-market-seeded') !== 'yes') {
            list.insertBefore(this.groups[1], rows[i]);
            return;
          }
        }
      },

      clear: function () {
        this.q = '';
        this.filter();
        if (this.open && this.$refs.q) this.$refs.q.focus();
      },

      visible: function () {
        return this.items.filter(function (it) { return !it.el.hidden; });
      },

      move: function (step) {
        var vis = this.visible();
        if (!vis.length) return;
        this.cursor += step;
        if (this.cursor < 0) this.cursor = vis.length - 1;
        if (this.cursor >= vis.length) this.cursor = 0;
        vis[this.cursor].el.focus();
      },

      // Enter picks the highlighted row, or the only remaining match when the
      // query has narrowed the list to one — the common case after typing.
      choose: function () {
        var vis = this.visible();
        if (!vis.length) return;
        var pick = this.cursor >= 0 && this.cursor < vis.length ? vis[this.cursor] : vis[0];
        pick.el.click();
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

        // Arriving from the homepage's "Advanced filters" button. The whole
        // point of that button is to land with the panel open, so it overrides
        // the remembered collapsed preference for this visit — without writing
        // it back, so the preference survives for the next ordinary search.
        if (new URLSearchParams(location.search).get('filters') === 'open') {
          this.collapsed = false;
          // Below 1024px the panel is an off-canvas drawer rather than a
          // sidebar, so "open" has to mean the drawer there.
          if (window.matchMedia('(max-width: 1023px)').matches) this.drawer = true;
        }
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
  // -- Add-listing form ------------------------------------------------------
  //
  // One continuous form with a waypoint rail beside it. This owns three things:
  // autosave, which section is active, and moving between sections.
  //
  // Activity is tracked with an IntersectionObserver rather than a scroll
  // handler: no work per frame, and — because clicking a waypoint mutes the
  // observer until its smooth scroll settles — no loop where the scroll updates
  // the highlight which re-triggers the scroll.
  window.previaListingForm = function (total, anchor) {
    return {
      state: 'idle', // idle | saving | saved
      total: total,
      timer: null,
      sections: [],
      activeKey: '',
      activeLabel: '',
      activeIndex: 0,
      progressPct: 0,
      observer: null,
      muted: false,     // true while a click-driven scroll is running
      muteTimer: null,

      init: function () {
        var self = this;

        this.sections = Array.prototype.map.call(
          document.querySelectorAll('.listing-section'),
          function (el) {
            var head = el.querySelector('h2');
            return {
              key: el.id.replace(/^ls-/, ''),
              el: el,
              label: head ? head.textContent.trim() : el.id,
            };
          }
        );
        if (!this.sections.length) return;
        this.setActive(0);

        // Scroll-spy. rootMargin biases the "current" section towards the top
        // third of the viewport, which is where a reader's attention sits —
        // without it the last section can never become active on a short page.
        this.observer = new IntersectionObserver(
          function (entries) {
            if (self.muted) return;
            var best = null;
            entries.forEach(function (entry) {
              if (!entry.isIntersecting) return;
              if (!best || entry.intersectionRatio > best.intersectionRatio) best = entry;
            });
            if (!best) return;
            var i = self.sections.findIndex(function (sec) { return sec.el === best.target; });
            if (i >= 0) self.setActive(i);
          },
          { rootMargin: '-12% 0px -60% 0px', threshold: [0, 0.15, 0.5, 1] }
        );
        this.sections.forEach(function (sec) { self.observer.observe(sec.el); });

        // A deep link ("?step=6", or a #hash) opens on that section.
        var target = anchor || (location.hash || '').replace(/^#ls-/, '');
        if (target) {
          this.$nextTick(function () { self.goTo(target, true); });
        }

        // Restore anything typed previously.
        try {
          var raw = localStorage.getItem('previa-draft');
          if (raw) this.draft = JSON.parse(raw);
        } catch (e) {
          this.draft = {};
        }
      },

      destroy: function () {
        if (this.observer) this.observer.disconnect();
      },

      setActive: function (i) {
        var sec = this.sections[i];
        if (!sec) return;
        this.activeIndex = i;
        this.activeKey = sec.key;
        this.activeLabel = sec.label;
        this.progressPct = Math.round(((i + 1) / this.total) * 100);
      },

      // Marker state for a waypoint: the section's own state wins (an error
      // stays an error), otherwise everything above the active one is done.
      stateOf: function (key, base) {
        if (base === 'error') return 'error';
        if (key === this.activeKey) return 'current';
        var i = this.sections.findIndex(function (s) { return s.key === key; });
        return i >= 0 && i < this.activeIndex ? 'done' : 'todo';
      },

      goTo: function (key, instant) {
        var self = this;
        var i = this.sections.findIndex(function (s) { return s.key === key; });
        if (i < 0) return;

        // Update the rail immediately so the click feels answered, then mute
        // the observer while the smooth scroll runs past intervening sections.
        //
        // The mute has to outlast the scroll. A fixed timer cannot know how
        // long that is — a jump from the first section to the last is far
        // slower than to the next one — so this waits for `scrollend` and
        // keeps a generous timer only as a fallback for browsers without it.
        // Getting this wrong left the rail highlighting whichever section the
        // scroll happened to be passing when the timer expired.
        this.setActive(i);
        this.muted = true;
        clearTimeout(this.muteTimer);

        var settle = function () {
          clearTimeout(self.muteTimer);
          window.removeEventListener('scrollend', settle);
          self.setActive(i);   // re-assert: the scroll may have overshot
          self.muted = false;
        };
        window.addEventListener('scrollend', settle, { once: true });
        this.muteTimer = setTimeout(settle, 1400);

        var reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
        this.sections[i].el.scrollIntoView({
          behavior: instant || reduce ? 'auto' : 'smooth',
          block: 'start',
        });

        // Move focus to the section so keyboard and screen-reader users land
        // where the scroll went, not where they were.
        var head = this.sections[i].el.querySelector('h2');
        if (head) {
          head.setAttribute('tabindex', '-1');
          head.focus({ preventScroll: true });
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
          existing.fields = data;
          existing.lastSection = this.activeKey;
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

  // -- Add-listing location --------------------------------------------------
  //
  // Holds the structured place behind the Location section. Two inputs feed it:
  // a pick from the Location search box, and a click on the map. Both land here
  // and fill the same read-only fields, so the two can never disagree.
  //
  // Map clicks are resolved by /mock/reverse-geocode, which reads the seeded
  // catalogue. Swapping in the Google Geocoding API is a change to that
  // endpoint's body — the response shape and everything below stay as they are.
  window.previaListingLocation = function () {
    return {
      place: {
        country: '', countryCode: '', state: '', city: '',
        district: '', address: '', lat: '', lng: '',
      },
      publicAddress: '',

      init: function () {
        var self = this;

        // From the search box.
        window.addEventListener('previa-location-picked', function (e) {
          var d = e.detail || {};
          self.apply({
            Country: '', CountryCode: d.countryCode, City: d.city,
            District: d.district, Address: d.address, Lat: d.lat, Lng: d.lng,
          }, d.label);
          // Resolve the rest (country name, postcode) from the same endpoint
          // the map uses, so both routes end up with identical fields.
          if (d.lat || d.lng) self.lookup(d.lat, d.lng, d.label);
        });

        // From a click on the map.
        window.addEventListener('previa-map-click', function (e) {
          var d = e.detail || {};
          self.lookup(d.lat, d.lng);
        });
      },

      lookup: function (lat, lng, keepLabel) {
        var self = this;
        fetch('/mock/reverse-geocode?lat=' + encodeURIComponent(lat) + '&lng=' + encodeURIComponent(lng))
          .then(function (r) { return r.ok ? r.json() : null; })
          .then(function (p) { if (p) self.apply(p, keepLabel); })
          .catch(function () { /* offline: the fields simply keep their values */ });
      },

      apply: function (p, label) {
        this.place = {
          country: p.Country || this.place.country || '',
          countryCode: p.CountryCode || '',
          state: p.State || '',
          city: p.City || '',
          district: p.District || '',
          address: p.Address || '',
          lat: p.Lat ? Number(p.Lat).toFixed(6) : '',
          lng: p.Lng ? Number(p.Lng).toFixed(6) : '',
        };
        // Seed the public display address the first time only — after that it
        // is the seller's text and must not be overwritten by a map nudge.
        if (!this.publicAddress) {
          this.publicAddress = label || [p.District, p.City, p.Country]
            .filter(Boolean).join(', ');
        }
      },
    };
  };

  // -- Add-listing media -----------------------------------------------------
  //
  // Photos and videos for a listing: local previews, ordering, and the cover.
  //
  // Nothing is uploaded in this milestone. Files are read as object URLs and
  // held in memory, which is enough to design and test the ordering, the cover
  // marker, the validation message and the removal confirmation. The second
  // milestone replaces addFiles() with a real upload and keeps the rest.
  window.previaMediaUploader = function () {
    var MAX_VIDEO_BYTES = 15 * 1024 * 1024; // the client's 15 MB video ceiling

    return {
      items: [],
      error: '',
      dragover: false,
      draggingId: null,
      overIndex: -1,
      nextId: 1,

      init: function () {
        // Two sample photographs, so the ordering controls have something to
        // act on before anything is chosen.
        var samples = Array.prototype.slice.call(
          document.querySelectorAll('[data-sample-image]')
        );
        var self = this;
        samples.forEach(function (el) {
          self.items.push({
            id: self.nextId++,
            url: el.getAttribute('data-sample-image'),
            name: el.getAttribute('data-sample-name') || 'Photo',
            kind: 'image',
          });
        });
      },

      dropFiles: function (e) {
        this.dragover = false;
        if (e.dataTransfer && e.dataTransfer.files) this.addFiles(e.dataTransfer.files);
      },

      addFiles: function (fileList) {
        this.error = '';
        var files = Array.prototype.slice.call(fileList || []);
        var rejected = [];

        for (var i = 0; i < files.length; i++) {
          var f = files[i];
          var isVideo = f.type.indexOf('video/') === 0;
          var isImage = f.type.indexOf('image/') === 0;

          if (!isVideo && !isImage) {
            rejected.push(f.name + ' is not an image or a video');
            continue;
          }
          if (isVideo && f.size > MAX_VIDEO_BYTES) {
            // Rounded up, and to two places when one would read as exactly the
            // limit: a file one byte over 15 MB reporting "15.0 MB — must be
            // 15 MB or less" looks like the validator is wrong.
            var mb = f.size / 1024 / 1024;
            var shown = mb.toFixed(1) === '15.0' ? (Math.ceil(mb * 100) / 100).toFixed(2)
                                                 : mb.toFixed(1);
            rejected.push(
              f.name + ' is ' + shown + ' MB — videos must be 15 MB or less'
            );
            continue;
          }

          this.items.push({
            id: this.nextId++,
            url: URL.createObjectURL(f),
            name: f.name,
            kind: isVideo ? 'video' : 'image',
          });
        }

        if (rejected.length) this.error = rejected.join('. ') + '.';
      },

      remove: function (id) {
        var item = this.items.find(function (it) { return it.id === id; });
        if (!item) return;
        if (!window.confirm('Remove ' + item.name + ' from this listing?')) return;
        if (item.url.indexOf('blob:') === 0) URL.revokeObjectURL(item.url);
        this.items = this.items.filter(function (it) { return it.id !== id; });
      },

      // Keyboard-accessible reordering — the alternative to dragging.
      move: function (i, step) {
        var j = i + step;
        if (j < 0 || j >= this.items.length) return;
        var copy = this.items.slice();
        var moved = copy.splice(i, 1)[0];
        copy.splice(j, 0, moved);
        this.items = copy;
      },

      startDrag: function (id) { this.draggingId = id; },

      dragOver: function (i) {
        if (this.draggingId === null) return;
        var from = this.items.findIndex(function (it) { return it.id === this.draggingId; }, this);
        if (from < 0 || from === i) return;
        var copy = this.items.slice();
        var moved = copy.splice(from, 1)[0];
        copy.splice(i, 0, moved);
        this.items = copy;
      },

      endDrag: function () { this.draggingId = null; },
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

      // How the results beside the map are laid out: 'list' shows each
      // property with its location and facts, 'grid' packs two to a row with
      // fewer details. Switching is purely a class on the results container —
      // no request, so the filters, the market and the map viewport are
      // untouched, and the open popup stays open.
      listMode: 'list',

      initListMode: function () {
        try {
          var saved = localStorage.getItem('previa-map-list-mode');
          if (saved === 'grid' || saved === 'list') this.listMode = saved;
        } catch (e) {}
      },

      setListMode: function (mode) {
        if (mode !== 'grid' && mode !== 'list') return;
        this.listMode = mode;
        try {
          localStorage.setItem('previa-map-list-mode', mode);
        } catch (e) {}
        // The panel's width does not change, but the browser may have been
        // mid-layout; let Leaflet re-measure so tiles never come back short.
        var self = this;
        setTimeout(function () {
          if (self.provider && self.provider.invalidate) self.provider.invalidate();
        }, 60);
      },

      init: function () {
        this.initListMode();
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

    // Popup markup.
    //
    // The photograph keeps its size; everything below it was tightened at the
    // client's request — price, title and location now read as one block and
    // the facts sit on a single line, so the popup covers much less map.
    //
    // The image carries the same pager as a property card: green dots with a
    // small green arrow either side. Leaflet builds popups from an HTML string
    // rather than from the Alpine-managed DOM, so the paging is wired up by
    // hand in bindPopupCarousel() once the popup opens.
    function popupHtml(p) {
      var shots = p.images || [];
      var slides = shots
        .map(function (src, i) {
          return (
            '<img class="map-popup__slide" src="' + src + '" alt="" width="300" height="188"' +
            (i === 0 ? '' : ' loading="lazy"') + '>'
          );
        })
        .join('');

      var pager = '';
      if (shots.length > 1) {
        var dots = shots
          .map(function (_, i) {
            return (
              '<button type="button" class="map-popup__dot' + (i === 0 ? ' is-on' : '') +
              '" data-go="' + i + '" aria-label="Show photo ' + (i + 1) + '"></button>'
            );
          })
          .join('');
        pager =
          '<div class="map-popup__pager">' +
          '<button type="button" class="map-popup__arrow" data-step="-1" aria-label="Previous photo">' +
          CHEVRON_LEFT + '</button>' +
          '<span class="map-popup__dots">' + dots + '</span>' +
          '<button type="button" class="map-popup__arrow" data-step="1" aria-label="Next photo">' +
          CHEVRON_RIGHT + '</button>' +
          '</div>';
      }

      var media = shots.length
        ? '<div class="map-popup__media">' +
          '<div class="map-popup__track">' + slides + '</div>' +
          '<a class="map-popup__media-link" href="' + p.url + '" aria-hidden="true" tabindex="-1"></a>' +
          pager +
          '</div>'
        : '';

      var rooms = p.rooms ? '<span>' + p.rooms + ' rooms</span>' : '';
      return (
        '<div class="map-popup__card">' +
        media +
        '<div class="map-popup__body">' +
        '<p class="map-popup__price">' + escapeHtml(p.full) + '</p>' +
        '<a class="map-popup__title" href="' + p.url + '">' + escapeHtml(p.title) + '</a>' +
        '<p class="map-popup__meta">' + escapeHtml(p.city) + '</p>' +
        '<p class="map-popup__facts">' +
        '<span>' + escapeHtml(p.type) + '</span>' + rooms +
        '<span>' + escapeHtml(p.area) + '</span></p>' +
        '<a class="btn btn--primary btn--sm btn--block" href="' + p.url + '">View property</a>' +
        '</div></div>'
      );
    }

    var CHEVRON_LEFT =
      '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" ' +
      'stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="m15 18-6-6 6-6"/></svg>';
    var CHEVRON_RIGHT =
      '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" ' +
      'stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="m9 18 6-6-6-6"/></svg>';

    // Paging inside an open popup.
    //
    // Delegated from the map container rather than bound per popup: Leaflet
    // creates and destroys popup DOM as markers open and close, so a listener
    // attached to the popup itself has to be re-attached on every open and
    // leaks if that is ever missed. One listener here handles every popup for
    // the life of the map.
    //
    // The current index is read back off the track's own transform, so no
    // state has to be kept in step with DOM Leaflet may have thrown away.
    function popupPaging(container) {
      container.addEventListener('click', function (e) {
        var control = e.target.closest
          ? e.target.closest('.map-popup__arrow, .map-popup__dot')
          : null;
        if (!control) return;

        var media = control.closest('.map-popup__media');
        var track = media && media.querySelector('.map-popup__track');
        if (!track) return;

        e.preventDefault();
        e.stopPropagation();

        var total = track.children.length;
        var match = /translate3d\(-([\d.]+)%/.exec(track.style.transform || '');
        var i = match ? Math.round(parseFloat(match[1]) / 100) : 0;

        var step = control.getAttribute('data-step');
        i = step !== null
          ? i + Number(step)
          : Number(control.getAttribute('data-go'));
        i = ((i % total) + total) % total;

        track.style.transform = 'translate3d(-' + i * 100 + '%,0,0)';
        var dots = media.querySelectorAll('.map-popup__dot');
        for (var d = 0; d < dots.length; d++) {
          dots[d].classList.toggle('is-on', d === i);
        }
      });
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

    // One delegated listener covers every popup this map will ever open.
    popupPaging(map.getContainer());

    // A click on the map surface itself (not on a marker or a popup) is a
    // location choice. The add-listing form listens for this and resolves the
    // point to an address; the search map has no listener, so it is inert there.
    map.on('click', function (e) {
      window.dispatchEvent(new CustomEvent('previa-map-click', {
        detail: { lat: e.latlng.lat, lng: e.latlng.lng },
      }));
    });

    // Move the single draft pin to a place chosen in the Location search box.
    window.addEventListener('listing-map-goto', function (e) {
      var d = e.detail || {};
      if (!d.lat && !d.lng) return;
      var zoom = d.kind === 'country' ? 6 : d.kind === 'city' ? 12 : 15;
      map.setView([d.lat, d.lng], zoom);
      for (var id in markers) {
        if (Object.prototype.hasOwnProperty.call(markers, id)) {
          markers[id].marker.setLatLng([d.lat, d.lng]);
        }
      }
    });

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
