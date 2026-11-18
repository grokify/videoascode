// Transcript review page. The model comes embedded (a read-only snapshot) or
// from /api/model (a live review served by `vac html review`).
(function () {
  'use strict';

  var $ = function (id) { return document.getElementById(id); };
  var embedded = JSON.parse($('model').textContent);
  var live = embedded === null;
  var model = null;
  var cur = 0;
  var lang = '';
  var compare = '';
  var filter = 'all';
  var frameSrc = '';
  var reloads = 0;
  var prefsKey = 'vac-review:' + location.pathname;
  // Browser speech for previews (the @plexusone/speech standalone build).
  var Speech = window.PlexusSpeech;
  var canSpeak = !!(Speech && Speech.isSynthesisSupported());
  var voices = [];
  var voiceByLang = {};
  var rate = 1;
  // The preview in progress: { token, slide, continuous, handle }.
  var playing = null;
  var playingSeg = -1;
  var playingWord = null; // { segment, start, end } from onWord

  function loadPrefs() {
    try { return JSON.parse(localStorage.getItem(prefsKey)) || {}; } catch (e) { return {}; }
  }
  function savePrefs() {
    try {
      localStorage.setItem(prefsKey, JSON.stringify({
        lang: lang, compare: compare, filter: filter, rate: rate, voices: voiceByLang
      }));
    } catch (e) { /* storage unavailable: preferences last for this visit only */ }
  }

  function esc(s) {
    return String(s).replace(/[&<>"']/g, function (c) {
      return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c];
    });
  }
  function reEsc(s) { return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'); }

  // Highlights pronunciation terms in text; each mark says how it is spoken.
  function withTerms(text, terms) {
    if (!terms || !terms.length) return esc(text);
    var re = new RegExp('(' + terms.map(function (t) { return reEsc(t.written); }).join('|') + ')', 'g');
    var spoken = {};
    terms.forEach(function (t) { spoken[t.written] = t.spoken; });
    return text.split(re).map(function (part, i) {
      return i % 2 ? '<mark class="term" title="Said as: ' + esc(spoken[part]) + '">' + esc(part) + '</mark>' : esc(part);
    }).join('');
  }

  function narr(slide, l) { return (slide && slide.langs[l]) || null; }

  function needsReview(n) { return !n || n.state !== 'approved'; }
  function matches(slide) {
    var n = narr(slide, lang);
    switch (filter) {
      case 'todo': return needsReview(n);
      case 'changes': return !!n && n.state === 'changes';
      case 'flagged': return !!n && (n.checks || []).some(function (c) { return c.level !== 'info'; });
      default: return true;
    }
  }

  // ---- Deck frame ----------------------------------------------------------

  function deckURL(slide, bust) {
    var q = model.frame.query || '';
    if (bust) q += (q ? '&' : '') + '_vacreview=' + bust;
    return model.frame.url + (q ? '?' + q : '') + slide.fragment;
  }

  function showSlideInDeck(force) {
    var slide = model.slides[cur];
    var frame = $('deck');
    if (force || !frameSrc || model.frame.nav === 'reload') {
      frameSrc = deckURL(slide, force || model.frame.nav === 'reload' ? ++reloads : 0);
      frame.src = frameSrc;
      return;
    }
    // Fragment-only navigation: the deck moves on hashchange, without a reload.
    var next = new URL(deckURL(slide, reloads), document.baseURI).href;
    try {
      frame.contentWindow.location.replace(next);
    } catch (e) {
      frame.src = next;
    }
    frameSrc = next;
  }

  function fitStage() {
    var stage = $('stage');
    var w = stage.clientWidth;
    var fw = model.frame.width;
    var fh = model.frame.height;
    stage.style.height = Math.round(w * fh / fw) + 'px';
    var f = $('deck');
    f.style.width = fw + 'px';
    f.style.height = fh + 'px';
    f.style.transform = 'scale(' + (w / fw) + ')';
  }

  // ---- Rendering -----------------------------------------------------------

  function renderHeader() {
    $('title').textContent = model.title || 'Transcript review';
    var bits = [model.slides.length + ' slides', model.lessons.length + ' lessons'];
    bits.push(model.voice ? 'voice: ' + model.voice : 'no voice set');
    if (model.readOnly) bits.push('read-only snapshot, ' + new Date(model.generated).toLocaleString());
    $('meta').textContent = bits.join(' · ');

    var opts = model.languages.map(function (l) {
      return '<option value="' + esc(l) + '">' + esc(l) + (l === model.defaultLanguage ? ' (source)' : '') + '</option>';
    }).join('');
    $('lang').innerHTML = opts;
    $('lang').value = lang;
    $('compare').innerHTML = '<option value="">None</option>' + opts;
    $('compare').value = compare;
    $('compare').disabled = model.languages.length < 2;
    $('filter').value = filter;

    var decks = model.decks || [];
    $('deck-pick').hidden = decks.length < 2;
    $('decks').innerHTML = decks.map(function (d) {
      return '<option value="' + esc(d.url) + '"' + (d.current ? ' selected' : '') + '>' + esc(d.name) + '</option>';
    }).join('');
  }

  function renderProgress() {
    var c = model.summary[lang] || { slides: 0, approved: 0, changes: 0, stale: 0, unreviewed: 0 };
    var pct = function (n) { return c.slides ? (100 * n / c.slides) + '%' : '0'; };
    $('progress').innerHTML =
      '<strong>' + c.approved + ' of ' + c.slides + '</strong> approved in ' + esc(lang) +
      (c.changes ? ' · ' + c.changes + ' changes' : '') +
      (c.stale ? ' · ' + c.stale + ' stale' : '') +
      '<div class="progress-bar" role="img" aria-label="' + c.approved + ' of ' + c.slides + ' approved">' +
      '<span style="width:' + pct(c.approved) + ';background:var(--ok)"></span>' +
      '<span style="width:' + pct(c.changes) + ';background:var(--err)"></span>' +
      '<span style="width:' + pct(c.stale) + ';background:var(--stale)"></span></div>';

    var banner = $('banner');
    if (c.slides && c.approved === c.slides) {
      banner.hidden = false;
      banner.className = 'banner';
      banner.style.background = 'var(--ok-bg)';
      banner.style.color = 'var(--ok)';
      banner.textContent = 'Every slide is approved in ' + lang + '. Ready for TTS.';
    } else if (!banner.classList.contains('error')) {
      banner.hidden = true;
    }
  }

  function renderOutline() {
    var html = '';
    model.lessons.forEach(function (l) {
      html += '<li class="lesson">' + l.number + '. ' + esc(l.title) + '</li>';
      for (var i = l.first; i <= l.last; i++) {
        var s = model.slides[i];
        var n = narr(s, lang);
        var worst = n && (n.checks || []).reduce(function (w, c) {
          return c.level === 'error' ? 'error' : (c.level === 'warn' && w !== 'error' ? 'warn' : w);
        }, '');
        html += '<li><button type="button" class="item' + (i === cur ? ' current' : '') + (matches(s) ? '' : ' dim') +
          '" data-i="' + i + '"><span class="num">' + (i + 1) + '</span>' +
          '<span class="dot ' + (n ? n.state : '') + '" title="' + esc(n ? n.state : 'no narration') + '"></span>' +
          '<span class="name">' + esc(s.title || s.id) + '</span>' +
          (worst ? '<span class="flag ' + worst + '" title="' + (worst === 'error' ? 'Error' : 'Warning') + '">' + (worst === 'error' ? '●' : '▲') + '</span>' : '') +
          '</button></li>';
      }
    });
    $('outline').innerHTML = html;
    var item = $('outline').querySelector('.current');
    if (item) item.scrollIntoView({ block: 'nearest' });
  }

  function renderColumn(slide, l) {
    var n = narr(slide, l);
    var h = '<div class="col"><h3>' + esc(l) + (l === model.defaultLanguage ? ' · source' : '') + '</h3>';
    if (!n || n.missing) {
      h += '<p class="text empty">No ' + esc(l) + ' narration for this slide.</p>';
    } else if (!n.text) {
      h += '<p class="text empty">No narration' + (n.pauseMs ? '; ' + n.pauseMs + ' ms of silence' : '') + '.</p>';
    } else {
      h += '<p class="text" lang="' + esc(l) + '">' + (n.segments || []).map(function (seg) {
        return '<span class="seg">' + withTerms(seg.text, n.terms) + '</span>' +
          (seg.pauseMs ? '<span class="pause" title="Pause">' + seg.pauseMs + ' ms</span>' : '');
      }).join('') + '</p>';
      h += '<div class="stats muted"><span>' + n.words + ' words</span><span>~' + Math.round(n.seconds) + ' s</span>' +
        (n.pauseMs ? '<span>' + n.pauseMs + ' ms pause</span>' : '') + '</div>';
      if (n.spoken !== n.text) {
        h += '<details class="spoken"><summary>What TTS receives</summary><p lang="' + esc(l) + '">' + esc(n.spoken) + '</p>' +
          '<ul class="terms">' + (n.terms || []).map(function (t) {
            return '<li><strong>' + esc(t.written) + '</strong> → ' + esc(t.spoken) + '</li>';
          }).join('') + '</ul></details>';
      }
    }
    if (n && n.checks && n.checks.length) {
      h += '<ul class="checks">' + n.checks.map(function (c) {
        return '<li class="' + c.level + '">' + esc(c.message) + '</li>';
      }).join('') + '</ul>';
    }
    return h + '</div>';
  }

  var stateLabel = {
    approved: 'Approved',
    changes: 'Changes requested',
    stale: 'Text changed since review',
    unreviewed: 'Not reviewed'
  };

  function renderDecision(slide) {
    var n = narr(slide, lang);
    if (!n) return '';
    var e = n.entry;
    var h = '<div class="decision"><span class="state ' + n.state + '">' + stateLabel[n.state] + ' · ' + esc(lang) + '</span>';
    if (model.readOnly) {
      if (e && e.note) h += '<p>' + esc(e.note) + '</p>';
    } else {
      h += '<textarea id="note" placeholder="Note: what should change, or anything to listen for">' + esc(e ? e.note || '' : '') + '</textarea>' +
        '<div class="actions">' +
        '<button type="button" class="approve" data-status="approved" title="Approve (A)">Approve</button>' +
        '<button type="button" class="changes" data-status="changes" title="Request changes (C, then Ctrl+Enter)">Request changes</button>' +
        (e ? '<button type="button" data-status="" title="Remove this decision">Clear</button>' : '') +
        '</div>';
    }
    if (e) {
      h += '<div class="who muted">' + (n.state === 'stale' ? 'Was ' + esc(e.status === 'approved' ? 'approved' : 'flagged for changes') + ' ' : 'Updated ') +
        new Date(e.updatedAt).toLocaleString() + (e.reviewer ? ' by ' + esc(e.reviewer) : '') + '</div>';
    }
    return h + '</div>';
  }

  function renderPanel() {
    var s = model.slides[cur];
    var lesson = model.lessons.filter(function (l) { return l.number === s.lesson; })[0];
    var two = compare && compare !== lang;
    $('panel').innerHTML =
      '<h2>Slide ' + (cur + 1) + ' · ' + esc(s.title || s.id) + '</h2>' +
      '<p class="sub muted">' + (lesson ? 'Lesson ' + lesson.number + ': ' + esc(lesson.title) + ' · ' : '') + '<code>' + esc(s.id) + '</code></p>' +
      '<div class="cols' + (two ? ' two' : '') + '">' + renderColumn(s, lang) + (two ? renderColumn(s, compare) : '') + '</div>' +
      renderDecision(s);
    markSegment();
    $('position').textContent = (cur + 1) + ' / ' + model.slides.length;
    $('prev').disabled = cur === 0;
    $('next').disabled = cur === model.slides.length - 1;
  }

  function render() {
    renderProgress();
    renderOutline();
    renderPanel();
  }

  // ---- Navigation and decisions -------------------------------------------

  // go moves to slide i. Moving by hand stops a preview; the preview itself
  // moves with keepPlaying.
  function go(i, keepPlaying) {
    if (i < 0 || i >= model.slides.length) return;
    if (!keepPlaying) stopPreview();
    cur = i;
    history.replaceState(null, '', '#' + encodeURIComponent(model.slides[cur].id));
    showSlideInDeck(0);
    render();
  }

  function nextToReview() {
    for (var k = 1; k <= model.slides.length; k++) {
      var i = (cur + k) % model.slides.length;
      if (needsReview(narr(model.slides[i], lang))) { go(i); return; }
    }
  }

  function showError(msg) {
    var b = $('banner');
    b.hidden = false;
    b.className = 'banner error';
    b.style.background = '';
    b.style.color = '';
    b.textContent = msg;
  }
  function clearError() {
    var b = $('banner');
    if (b.classList.contains('error')) { b.hidden = true; b.className = 'banner'; }
  }

  function decide(status) {
    if (model.readOnly) return;
    var s = model.slides[cur];
    var note = $('note') ? $('note').value : '';
    fetch('api/review', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ slideId: s.id, language: lang, status: status, note: note })
    }).then(function (r) {
      if (!r.ok) return r.text().then(function (t) { throw new Error(t); });
      return r.json();
    }).then(function (m) {
      clearError();
      setModel(m);
      if (status === 'approved') nextToReview(); else render();
    }).catch(function (err) { showError('Could not save: ' + err.message); });
  }

  function setModel(m) {
    model = m;
    if (model.languages.indexOf(lang) < 0) lang = model.languages[0];
    if (compare && model.languages.indexOf(compare) < 0) compare = '';
    if (cur >= model.slides.length) cur = model.slides.length - 1;
    renderHeader();
  }

  function refresh(reloadDeck) {
    if (!live) { if (reloadDeck) showSlideInDeck(++reloads); return; }
    fetch('api/model', { cache: 'no-store' }).then(function (r) {
      if (!r.ok) return r.text().then(function (t) { throw new Error(t); });
      return r.json();
    }).then(function (m) {
      clearError();
      setModel(m);
      if (reloadDeck) showSlideInDeck(++reloads);
      render();
    }).catch(function (err) { showError('Could not load the transcript: ' + err.message); });
  }

  function start(m) {
    var prefs = loadPrefs();
    lang = prefs.lang || '';
    compare = prefs.compare || '';
    filter = prefs.filter || 'all';
    rate = Number(prefs.rate) || 1;
    voiceByLang = prefs.voices || {};
    setModel(m);
    setupSpeech();
    var want = decodeURIComponent(location.hash.slice(1));
    var at = model.slides.findIndex(function (s) { return s.id === want; });
    cur = at < 0 ? 0 : at;
    fitStage();
    showSlideInDeck(0);
    render();
  }

  // ---- Browser speech preview ----------------------------------------------

  function setupSpeech() {
    if (!canSpeak) return;
    $('preview').hidden = false;
    document.querySelector('.speech-keys').hidden = false;
    $('rate').value = String(rate);
    renderVoices();
    Speech.getVoices().then(function (vs) { voices = vs; renderVoices(); });
  }

  function renderVoices() {
    var sel = $('voice');
    var list = voices.filter(function (v) { return Speech.langMatches(v.lang, lang); });
    if (!list.length) {
      sel.innerHTML = '<option value="">' + (voices.length ? 'No ' + esc(lang) + ' voice' : 'Default voice') + '</option>';
      sel.disabled = voices.length > 0;
      return;
    }
    sel.disabled = false;
    sel.innerHTML = list.map(function (v) {
      return '<option value="' + esc(v.voiceURI) + '">' + esc(v.name) + (v.localService ? '' : ' (online)') + '</option>';
    }).join('');
    var want = voiceByLang[lang];
    var best = Speech.pickVoice(list, lang);
    sel.value = list.some(function (v) { return v.voiceURI === want; }) ? want : (best || list[0]).voiceURI;
  }

  // Highlights the segment being spoken in the current language's column,
  // and the word being spoken when the voice reports words. Word offsets are
  // into the text the voice received, so while a word is marked the segment
  // shows that text; it goes back to the written text afterwards.
  function markSegment() {
    var segs = document.querySelectorAll('#panel .col:first-child .seg');
    var on = playing && playing.slide === cur ? playingSeg : -1;
    var w = on >= 0 && playingWord && playingWord.segment === on ? playingWord : null;
    var n = narr(model.slides[cur], lang);
    for (var k = 0; k < segs.length; k++) {
      var el = segs[k];
      el.classList.toggle('speaking', k === on);
      if (w && k === on) {
        var seg = n && n.segments ? n.segments[k] : null;
        var t = seg ? (seg.spoken || seg.text) : '';
        if (el.dataset.orig === undefined) el.dataset.orig = el.innerHTML;
        el.innerHTML = esc(t.slice(0, w.start)) + '<mark class="word">' + esc(t.slice(w.start, w.end)) +
          '</mark>' + esc(t.slice(w.end));
      } else if (el.dataset.orig !== undefined) {
        el.innerHTML = el.dataset.orig;
        delete el.dataset.orig;
      }
    }
  }

  function updatePlayButtons() {
    var one = playing && !playing.continuous;
    var many = playing && playing.continuous;
    $('play').textContent = one ? '■ Stop' : '▶ This slide';
    $('play').classList.toggle('on', !!one);
    $('play-on').textContent = many ? '■ Stop' : '▶ From here';
    $('play-on').classList.toggle('on', !!many);
  }

  function stopPreview() {
    if (!playing) return;
    var p = playing;
    playing = null;
    playingSeg = -1;
    playingWord = null;
    if (p.handle) p.handle.stop();
    updatePlayButtons();
    markSegment();
  }

  function preview(continuous) {
    if (!canSpeak || !model) return;
    var was = playing;
    stopPreview();
    // The button that started a preview also stops it.
    if (was && was.continuous === continuous) return;
    playSlide(cur, continuous);
  }

  function playSlide(i, continuous) {
    var n = narr(model.slides[i], lang);
    var segs = n && !n.missing ? (n.segments || []).map(function (seg) {
      return { text: seg.spoken || seg.text, pauseMs: seg.pauseMs || 0 };
    }) : [];
    var token = {};
    playing = { token: token, slide: i, continuous: continuous, handle: null };
    playingSeg = -1;
    playingWord = null;
    updatePlayButtons();
    var mine = function () { return playing && playing.token === token; };
    var h = Speech.speak(segs, {
      lang: lang,
      voice: $('voice').value || undefined,
      rate: rate,
      onSegment: function (k) { if (mine()) { playingSeg = k; playingWord = null; markSegment(); } },
      onWord: function (w) { if (mine()) { playingWord = w; markSegment(); } }
    });
    playing.handle = h;
    h.done.then(function (r) {
      if (!mine()) return;
      if (r.status === 'ended' && continuous && i + 1 < model.slides.length) {
        go(i + 1, true);
        playSlide(i + 1, true);
        return;
      }
      playing = null;
      playingSeg = -1;
      playingWord = null;
      updatePlayButtons();
      markSegment();
    }, function (err) {
      if (!mine()) return;
      playing = null;
      playingSeg = -1;
      playingWord = null;
      updatePlayButtons();
      markSegment();
      showError(err.code === 'not-allowed'
        ? 'The browser blocked speech. Click a preview button to start it.'
        : 'Browser speech failed: ' + (err.code || err.message));
    });
  }

  // ---- Events --------------------------------------------------------------

  $('outline').addEventListener('click', function (e) {
    var b = e.target.closest('.item');
    if (b) go(Number(b.dataset.i));
  });
  $('panel').addEventListener('click', function (e) {
    var b = e.target.closest('button[data-status]');
    if (b) decide(b.dataset.status);
  });
  $('prev').addEventListener('click', function () { go(cur - 1); });
  $('next').addEventListener('click', function () { go(cur + 1); });
  $('next-todo').addEventListener('click', nextToReview);
  $('reload').addEventListener('click', function () { refresh(true); });
  $('lang').addEventListener('change', function (e) {
    stopPreview(); lang = e.target.value; savePrefs(); if (canSpeak) renderVoices(); render();
  });
  $('play').addEventListener('click', function () { preview(false); });
  $('play-on').addEventListener('click', function () { preview(true); });
  $('voice').addEventListener('change', function (e) {
    if (e.target.value) voiceByLang[lang] = e.target.value;
    savePrefs();
  });
  $('rate').addEventListener('change', function (e) { rate = Number(e.target.value) || 1; savePrefs(); });
  $('compare').addEventListener('change', function (e) { compare = e.target.value; savePrefs(); render(); });
  $('filter').addEventListener('change', function (e) { filter = e.target.value; savePrefs(); render(); });
  $('decks').addEventListener('change', function (e) { location.href = e.target.value; });
  addEventListener('resize', function () { if (model) fitStage(); });
  addEventListener('focus', function () { if (model && live) refresh(false); });

  document.addEventListener('keydown', function (e) {
    if (!model) return;
    var typing = /^(TEXTAREA|INPUT|SELECT)$/.test(e.target.tagName);
    if (typing) {
      if (e.key === 'Escape') e.target.blur();
      if (e.target.id === 'note' && e.key === 'Enter' && (e.metaKey || e.ctrlKey)) { e.preventDefault(); decide('changes'); }
      return;
    }
    if (e.metaKey || e.ctrlKey || e.altKey) return;
    switch (e.key) {
      case 'ArrowRight': case 'j': go(cur + 1); break;
      case 'ArrowLeft': case 'k': go(cur - 1); break;
      case 'n': case 'N': nextToReview(); break;
      case 'a': case 'A': decide('approved'); break;
      case 'c': case 'C': if ($('note')) { e.preventDefault(); $('note').focus(); } break;
      case 'r': case 'R': refresh(true); break;
      case 'p': preview(false); break;
      case 'P': preview(true); break;
      case 'Escape': if (!playing) return; stopPreview(); break;
      default: return;
    }
    e.preventDefault();
  });

  if (live) {
    fetch('api/model', { cache: 'no-store' }).then(function (r) {
      if (!r.ok) return r.text().then(function (t) { throw new Error(t); });
      return r.json();
    }).then(start).catch(function (err) { showError('Could not load the transcript: ' + err.message); });
  } else {
    start(embedded);
  }
})();
