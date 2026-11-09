/*
 * OBS overlay: receives speech commands from the local relay and plays them
 * one at a time. Audio is requested from VoiceStudio directly by this page.
 *
 * URL parameters:
 *   id        (required) your unique ID from the extension settings
 *   relay     relay base URL (default: the address this page was loaded from)
 *   endpoint  override the VoiceStudio address, e.g. http://192.168.1.20:3900
 *   volume    0 to 1 (default 1)
 *   maxqueue  drop new items when this many are waiting (default 30)
 *   timeout   seconds before a stuck item is skipped (default 60)
 *   debug     show a status/log panel
 */
(function () {
  'use strict';

  var q = new URLSearchParams(location.search);
  var id = q.get('id') || '';
  var relay = (q.get('relay') || location.origin).replace(/\/+$/, '');
  var debug = q.has('debug');
  var endpointOverride = (q.get('endpoint') || '').replace(/\/+$/, '');
  var volume = num(q.get('volume'), 0, 1, 1);
  var maxQueue = Math.max(1, Math.floor(num(q.get('maxqueue'), 1, 500, 30)));
  var timeoutMs = Math.max(5, num(q.get('timeout'), 5, 600, 60)) * 1000;

  var el = {
    banner: document.getElementById('banner'),
    status: document.getElementById('status'),
    stats: document.getElementById('stats'),
    log: document.getElementById('log')
  };
  if (debug) document.body.classList.add('debug');

  var queue = [];
  var busy = false;
  var blocked = false;
  var played = 0;
  var errors = 0;
  var es = null;
  var useProxy = false; // set once a direct request to VoiceStudio fails but the relay can reach it

  function num(value, min, max, fallback) {
    var n = parseFloat(value);
    if (!isFinite(n)) return fallback;
    return Math.min(max, Math.max(min, n));
  }

  function setStatus(text, bad) {
    el.status.textContent = text;
    el.status.className = bad ? 'bad' : '';
  }

  function updateStats() {
    el.stats.textContent = 'Queued: ' + queue.length + '  Played: ' + played + '  Errors: ' + errors;
  }

  function log(text, bad) {
    if (!debug) return;
    var li = document.createElement('li');
    li.textContent = new Date().toLocaleTimeString() + '  ' + text;
    if (bad) li.className = 'bad';
    el.log.insertBefore(li, el.log.firstChild);
    while (el.log.children.length > 40) el.log.removeChild(el.log.lastChild);
  }

  function speechUrl(item) {
    if (!endpointOverride) return item.endpoint;
    return /\/speech$/.test(endpointOverride) ? endpointOverride : endpointOverride + '/v1/audio/speech';
  }

  function play(audio) {
    return new Promise(function (resolve, reject) {
      var timer = null;
      function finish(err) {
        clearTimeout(timer);
        audio.removeEventListener('ended', onEnded);
        audio.removeEventListener('error', onError);
        audio.removeEventListener('loadedmetadata', onMeta);
        if (err) reject(err); else resolve();
      }
      function arm(ms) {
        clearTimeout(timer);
        timer = setTimeout(function () {
          try { audio.pause(); } catch (e) { /* ignore */ }
          finish(new Error('Playback stalled'));
        }, ms);
      }
      function onEnded() { finish(); }
      function onError() { finish(new Error('Audio could not be decoded')); }
      function onMeta() {
        if (isFinite(audio.duration)) arm(audio.duration * 1000 + 15000);
      }
      audio.addEventListener('ended', onEnded);
      audio.addEventListener('error', onError);
      audio.addEventListener('loadedmetadata', onMeta);
      arm(timeoutMs);
      var p = audio.play();
      if (p && p.catch) p.catch(function (e) { finish(e); });
    });
  }

  // Ask VoiceStudio directly. If the browser blocks or fails that request, ask the relay to do it.
  function requestSpeech(target, body, signal) {
    function viaRelay() {
      return fetch(relay + '/tts?id=' + encodeURIComponent(id), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ endpoint: target, body: body }),
        signal: signal
      });
    }
    if (useProxy) return viaRelay();
    function fallback(reason) {
      return viaRelay().then(function (res) {
        if (res.ok) {
          useProxy = true;
          log('Direct request to VoiceStudio was refused; now going through the relay');
        }
        return res;
      }, function () {
        throw new Error('Could not reach VoiceStudio or the relay (' + reason + ')');
      });
    }
    return fetch(target, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
      signal: signal
    }).then(function (res) {
      // 403: VoiceStudio refuses requests that come from a web page. The relay's request is accepted.
      return res.status === 403 ? fallback('HTTP 403') : res;
    }, function (err) {
      if (err && err.name === 'AbortError') throw err;
      return fallback(err && err.message ? err.message : 'network error');
    });
  }

  function speak(item) {
    var ctrl = new AbortController();
    var timer = setTimeout(function () { ctrl.abort(); }, timeoutMs);
    var body = { model: item.model, input: item.text, voice: item.voice, response_format: item.format };
    if (typeof item.speed === 'number') body.speed = item.speed;
    var url = null;
    return requestSpeech(speechUrl(item), body, ctrl.signal).then(function (res) {
      if (!res.ok) {
        return res.text().then(function (t) {
          var msg = t.slice(0, 160);
          var relayError = false;
          try {
            var j = JSON.parse(t);
            if (j && j.ok === false && j.error) { msg = j.error; relayError = true; }
          } catch (e) { /* plain text from VoiceStudio */ }
          throw new Error(relayError ? msg : 'VoiceStudio HTTP ' + res.status + ' ' + msg);
        });
      }
      return res.blob();
    }).then(function (blob) {
      clearTimeout(timer);
      if (!blob.size) throw new Error('VoiceStudio returned empty audio');
      url = URL.createObjectURL(blob);
      var audio = new Audio();
      audio.volume = volume;
      audio.src = url;
      return play(audio).then(function () {
        audio.removeAttribute('src');
      });
    }).then(function () {
      if (url) URL.revokeObjectURL(url);
    }, function (err) {
      clearTimeout(timer);
      if (url) URL.revokeObjectURL(url);
      if (err && err.name === 'AbortError') throw new Error('VoiceStudio did not answer in time');
      throw err;
    });
  }

  function pump() {
    if (busy || blocked) return;
    if (!queue.length) { updateStats(); return; }
    busy = true;
    var item = queue.shift();
    updateStats();
    log('Speaking !' + (item.command || '?') + ' (' + item.voice + ') from ' + (item.user || 'viewer'));
    speak(item).then(function () {
      played++;
    }, function (err) {
      if (err && err.name === 'NotAllowedError') {
        // Browser blocked autoplay (normal Chrome tab; OBS does not do this).
        queue.unshift(item);
        blocked = true;
        el.banner.style.display = 'flex';
        log('Audio blocked until you click the page', true);
        return;
      }
      errors++;
      log('ERROR: ' + (err && err.message ? err.message : err), true);
    }).then(function () {
      busy = false;
      updateStats();
      pump();
    });
  }

  el.banner.addEventListener('click', function () {
    el.banner.style.display = 'none';
    blocked = false;
    pump();
  });

  function enqueue(item) {
    if (!item || typeof item.text !== 'string' || typeof item.voice !== 'string' || typeof item.endpoint !== 'string') return;
    if (queue.length >= maxQueue) {
      log('Queue full, dropped a message', true);
      return;
    }
    queue.push(item);
    updateStats();
    pump();
  }

  function connect() {
    if (!id) {
      setStatus('Missing ?id= in the URL. Copy the full URL from the extension settings.', true);
      el.banner.style.display = 'none';
      document.body.classList.add('debug');
      return;
    }
    if (es) es.close();
    es = new EventSource(relay + '/events?id=' + encodeURIComponent(id));
    es.onopen = function () {
      setStatus('Connected to relay');
      log('Connected to relay');
    };
    es.onmessage = function (e) {
      var item;
      try { item = JSON.parse(e.data); } catch (err) { return; }
      enqueue(item);
    };
    es.onerror = function () {
      setStatus('Relay not reachable, retrying...', true);
      if (es.readyState === 2) { // closed by browser: reconnect ourselves
        setTimeout(connect, 3000);
      }
    };
  }

  updateStats();
  connect();
})();
