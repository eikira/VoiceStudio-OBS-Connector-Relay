'use strict';
// Update check and "open folder" tests. They copy the program into a temp folder (settings.txt lives next
// to it), point the update check at a fake GitHub server and put a fake file manager first in PATH.
const test = require('node:test');
const assert = require('node:assert');
const http = require('http');
const fs = require('fs');
const os = require('os');
const path = require('path');
const { spawn } = require('child_process');
const BIN = process.env.RELAY_BIN || path.join(__dirname, '..', 'relay-test');
const SKIP = process.platform === 'win32';
const OPENER = process.platform === 'darwin' ? 'open' : 'xdg-open';
const wait = (ms) => new Promise((r) => setTimeout(r, ms));

function fakeGitHub(tag, status) {
  const hits = [];
  const s = http.createServer((req, res) => {
    hits.push({ url: req.url, ua: req.headers['user-agent'] });
    res.writeHead(status || 200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({ tag_name: tag, html_url: 'https://evil.example/x' }));
  });
  return new Promise((resolve) => s.listen(0, '127.0.0.1', () => resolve({ s, hits, url: 'http://127.0.0.1:' + s.address().port + '/latest' })));
}

function sandbox() {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'relay-upd-'));
  const app = path.join(dir, 'app');
  fs.mkdirSync(app);
  const bin = path.join(app, 'relay');
  fs.copyFileSync(BIN, bin);
  fs.chmodSync(bin, 0o755);
  const fakeBin = path.join(dir, 'fakebin');
  fs.mkdirSync(fakeBin);
  const log = path.join(dir, 'opened.txt');
  fs.writeFileSync(path.join(fakeBin, OPENER), '#!/bin/sh\necho "$1" >> "' + log + '"\n', { mode: 0o755 });
  const env = Object.assign({}, process.env, { PATH: fakeBin + path.delimiter + process.env.PATH, HOME: dir, XDG_CONFIG_HOME: path.join(dir, 'cfg') });
  return { dir, app, bin, log, env };
}

// Async on purpose: the fake GitHub server lives in this process and must keep answering.
function cli(sb, args, extraEnv) {
  return new Promise((resolve) => {
    const p = spawn(sb.bin, args, { env: Object.assign({}, sb.env, extraEnv || {}) });
    let stdout = '', stderr = '';
    p.stdout.on('data', (c) => (stdout += c));
    p.stderr.on('data', (c) => (stderr += c));
    const t = setTimeout(() => p.kill(), 15000);
    p.on('close', (status) => { clearTimeout(t); resolve({ status, stdout, stderr }); });
  });
}

function startRelay(sb, extraEnv) {
  const port = 20000 + Math.floor(Math.random() * 20000);
  const proc = spawn(sb.bin, ['--port', String(port)], { env: Object.assign({}, sb.env, extraEnv || {}), stdio: 'ignore' });
  return { proc, base: 'http://127.0.0.1:' + port, port };
}

function req(base, method, p, headers) {
  return new Promise((resolve, reject) => {
    const r = http.request(base + p, { method, headers: headers || {} }, (res) => {
      let d = '';
      res.on('data', (c) => (d += c));
      res.on('end', () => resolve({ status: res.statusCode, body: d }));
    });
    r.on('error', reject);
    r.end();
  });
}

async function up(rel) {
  for (let i = 0; i < 50; i++) { try { await req(rel.base, 'GET', '/'); return; } catch (e) { await wait(100); } }
  throw new Error('relay did not start');
}

test('--check-update: newer, same, older, error, garbage', { skip: SKIP }, async () => {
  const sb = sandbox();
  const cases = [
    ['v9.9.9', 200, 0, /update is available: 9\.9\.9/],
    ['0.0.7', 200, 0, /newest version \(0\.0\.7\)/],
    ['0.0.1', 200, 0, /newest version/],
    ['v9.9.9', 500, 1, /Could not check for updates/],
    ['latest', 200, 1, /Could not check for updates/]
  ];
  for (const [tag, status, code, re] of cases) {
    const g = await fakeGitHub(tag, status);
    const r = await cli(sb, ['--check-update'], { RELAY_UPDATE_URL: g.url });
    g.s.close();
    assert.strictEqual(r.status, code, tag + ' ' + status + ': ' + r.stdout + r.stderr);
    assert.match(r.stdout, re);
    if (status === 200 && tag === 'v9.9.9') {
      // the download page is always the fixed GitHub page, never what the server sent
      assert.match(r.stdout, /https:\/\/github\.com\/eikira\/VoiceStudio-OBS-Connector-Relay\/releases\/latest/);
      assert.doesNotMatch(r.stdout, /evil\.example/);
      assert.match(g.hits[0].ua, /^VoiceStudio-OBS-Connector-Relay\/0\.0\.7/);
    }
  }
});

test('--set-updates writes the setting and rejects other values', { skip: SKIP }, async () => {
  const sb = sandbox();
  assert.strictEqual((await cli(sb, ['--set-updates', 'off'])).status, 0);
  assert.match(fs.readFileSync(path.join(sb.app, 'settings.txt'), 'utf8'), /UPDATECHECK=off/);
  assert.strictEqual((await cli(sb, ['--set-updates', 'on'])).status, 0);
  assert.match(fs.readFileSync(path.join(sb.app, 'settings.txt'), 'utf8'), /UPDATECHECK=on/);
  assert.strictEqual((await cli(sb, ['--set-updates', 'maybe'])).status, 2);
});

test('running relay: checks GitHub when on, never when off; /info reports it', { skip: SKIP }, async () => {
  const gOn = await fakeGitHub('v9.9.9');
  const gOff = await fakeGitHub('v9.9.9');
  const sbOn = sandbox();
  const sbOff = sandbox();
  await cli(sbOff, ['--set-updates', 'off']);
  const a = startRelay(sbOn, { RELAY_UPDATE_URL: gOn.url });
  const b = startRelay(sbOff, { RELAY_UPDATE_URL: gOff.url });
  try {
    await up(a); await up(b);
    const before = JSON.parse((await req(a.base, 'GET', '/info')).body);
    assert.strictEqual(before.version, '0.0.7');
    assert.strictEqual(before.updateCheck, true);
    assert.strictEqual(before.updateAvailable, false); // not asked yet
    await wait(9500);
    const infoOn = JSON.parse((await req(a.base, 'GET', '/info')).body);
    assert.strictEqual(infoOn.updateAvailable, true);
    assert.strictEqual(infoOn.latest, '9.9.9');
    assert.strictEqual(infoOn.downloadUrl, 'https://github.com/eikira/VoiceStudio-OBS-Connector-Relay/releases/latest');
    assert.strictEqual(gOn.hits.length, 1);
    const infoOff = JSON.parse((await req(b.base, 'GET', '/info')).body);
    assert.strictEqual(infoOff.updateCheck, false);
    assert.strictEqual(infoOff.updateAvailable, false);
    assert.strictEqual(gOff.hits.length, 0);
  } finally { a.proc.kill(); b.proc.kill(); gOn.s.close(); gOff.s.close(); }
});

test('a failing GitHub does not disturb the relay', { skip: SKIP }, async () => {
  const sb = sandbox();
  const rel = startRelay(sb, { RELAY_UPDATE_URL: 'http://127.0.0.1:9/nothing' });
  try {
    await up(rel);
    await wait(9500);
    const info = JSON.parse((await req(rel.base, 'GET', '/info')).body);
    assert.strictEqual(info.updateAvailable, false);
    assert.strictEqual((await req(rel.base, 'GET', '/')).status, 200);
  } finally { rel.proc.kill(); }
});

test('/open-folder: refused without the header, opens the program folder with it', { skip: SKIP }, async () => {
  const sb = sandbox();
  const rel = startRelay(sb, { RELAY_UPDATE_URL: 'http://127.0.0.1:9/x' });
  try {
    await up(rel);
    assert.strictEqual((await req(rel.base, 'POST', '/open-folder')).status, 403);
    assert.strictEqual((await req(rel.base, 'GET', '/open-folder', { 'X-Relay-Open': '1' })).status, 404);
    assert.ok(!fs.existsSync(sb.log));
    const ok = await req(rel.base, 'POST', '/open-folder', { 'X-Relay-Open': '1' });
    assert.strictEqual(ok.status, 200);
    assert.strictEqual(JSON.parse(ok.body).path, fs.realpathSync(sb.app));
    await wait(300);
    assert.strictEqual(fs.readFileSync(sb.log, 'utf8').trim(), fs.realpathSync(sb.app));
  } finally { rel.proc.kill(); }
});

test('--open-folder opens the program folder', { skip: SKIP }, async () => {
  const sb = sandbox();
  const r = await cli(sb, ['--open-folder']);
  assert.strictEqual(r.status, 0, r.stdout + r.stderr);
  await wait(300);
  assert.strictEqual(fs.readFileSync(sb.log, 'utf8').trim(), fs.realpathSync(sb.app));
});
