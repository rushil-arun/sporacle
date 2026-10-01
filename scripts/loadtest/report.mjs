#!/usr/bin/env node
// Builds a self-contained HTML report (inline SVG charts, no external deps or
// network calls) from the raw k6 NDJSON + summary JSON produced by run_all.sh.
//
// Usage: node report.mjs <run-dir>   (e.g. scripts/loadtest/reports/20260930-120000)

import fs from 'node:fs';
import path from 'node:path';

const runDir = process.argv[2];
if (!runDir) {
  console.error('usage: node report.mjs <run-dir>');
  process.exit(1);
}

const TESTS = [
  { name: 'create_game', title: 'Create Game', latencyMetric: 'create_game_latency', unit: 'ms' },
  { name: 'join_game', title: 'Join Game (lobby handshake)', latencyMetric: 'join_handshake_latency', unit: 'ms' },
  { name: 'play_game', title: 'Play Game (claim round trip)', latencyMetric: 'play_claim_round_trip', unit: 'ms' },
];

const BUCKET_MS = 5000;
const COLORS = { p50: '#4C78A8', p90: '#F58518', p99: '#E45756', vus: '#72B7B2', err: '#B279A2' };

function percentile(sorted, p) {
  if (sorted.length === 0) return 0;
  const idx = Math.min(sorted.length - 1, Math.round((p / 100) * (sorted.length - 1)));
  return sorted[idx];
}

function minOf(arr) {
  let m = Infinity;
  for (const v of arr) if (v < m) m = v;
  return m;
}

function maxOf(arr) {
  let m = -Infinity;
  for (const v of arr) if (v > m) m = v;
  return m;
}

function readNdjson(file) {
  if (!fs.existsSync(file)) return [];
  const lines = fs.readFileSync(file, 'utf8').split('\n').filter(Boolean);
  const points = [];
  for (const line of lines) {
    let rec;
    try {
      rec = JSON.parse(line);
    } catch {
      continue;
    }
    if (rec.type === 'Point') points.push(rec);
  }
  return points;
}

function buildSeries(points) {
  if (points.length === 0) return null;
  const t0 = minOf(points.map((p) => new Date(p.data.time).getTime()));

  const byMetricBucket = {}; // metric -> bucketIndex -> values[]
  for (const p of points) {
    const metric = p.metric;
    const t = new Date(p.data.time).getTime() - t0;
    const bucket = Math.floor(t / BUCKET_MS);
    byMetricBucket[metric] = byMetricBucket[metric] || {};
    byMetricBucket[metric][bucket] = byMetricBucket[metric][bucket] || [];
    byMetricBucket[metric][bucket].push(p.data.value);
  }

  const maxBucket = Math.max(
    0,
    ...Object.values(byMetricBucket).flatMap((b) => Object.keys(b).map(Number))
  );

  function seriesFor(metric, reducer) {
    const buckets = byMetricBucket[metric] || {};
    const out = [];
    for (let i = 0; i <= maxBucket; i++) {
      const vals = buckets[i];
      out.push(vals ? reducer(vals) : null);
    }
    return out;
  }

  return { t0, maxBucket, byMetricBucket, seriesFor };
}

function statsFor(series, metric) {
  return {
    p50: series.seriesFor(metric, (vals) => percentile([...vals].sort((a, b) => a - b), 50)),
    p90: series.seriesFor(metric, (vals) => percentile([...vals].sort((a, b) => a - b), 90)),
    p99: series.seriesFor(metric, (vals) => percentile([...vals].sort((a, b) => a - b), 99)),
  };
}

function svgLineChart({ width = 720, height = 260, series, colors, yLabel, xLabelEvery = 4 }) {
  const padL = 50, padR = 20, padT = 16, padB = 30;
  const w = width - padL - padR, h = height - padT - padB;
  const n = Math.max(...series.map((s) => s.data.length));
  if (n === 0) return '<svg></svg>';

  const allVals = series.flatMap((s) => s.data.filter((v) => v !== null && v !== undefined));
  const maxY = Math.max(1, ...allVals) * 1.1;

  const x = (i) => padL + (i / Math.max(1, n - 1)) * w;
  const y = (v) => padT + h - (v / maxY) * h;

  function pathFor(data) {
    let d = '';
    let started = false;
    data.forEach((v, i) => {
      if (v === null || v === undefined) {
        started = false;
        return;
      }
      d += (started ? ' L ' : 'M ') + x(i).toFixed(1) + ' ' + y(v).toFixed(1);
      started = true;
    });
    return d;
  }

  const gridLines = [];
  const steps = 4;
  for (let i = 0; i <= steps; i++) {
    const val = (maxY / steps) * i;
    const yy = y(val);
    gridLines.push(
      `<line x1="${padL}" y1="${yy.toFixed(1)}" x2="${padL + w}" y2="${yy.toFixed(1)}" stroke="#e2e2e2" stroke-width="1"/>` +
      `<text x="${padL - 8}" y="${(yy + 4).toFixed(1)}" font-size="10" text-anchor="end" fill="#666">${val.toFixed(0)}</text>`
    );
  }

  const xLabels = [];
  for (let i = 0; i < n; i += xLabelEvery) {
    xLabels.push(
      `<text x="${x(i).toFixed(1)}" y="${padT + h + 18}" font-size="10" text-anchor="middle" fill="#666">${(i * BUCKET_MS) / 1000}s</text>`
    );
  }

  const paths = series
    .map((s) => `<path d="${pathFor(s.data)}" fill="none" stroke="${s.color}" stroke-width="2"/>`)
    .join('');

  const legend = series
    .map(
      (s, i) =>
        `<circle cx="${padL + i * 90}" cy="${height - 4}" r="4" fill="${s.color}"/>` +
        `<text x="${padL + i * 90 + 10}" y="${height}" font-size="11" fill="#333">${s.label}</text>`
    )
    .join('');

  return `<svg width="${width}" height="${height + 14}" viewBox="0 0 ${width} ${height + 14}" xmlns="http://www.w3.org/2000/svg">
    <text x="${padL}" y="12" font-size="11" fill="#333">${yLabel}</text>
    ${gridLines.join('')}
    ${paths}
    ${xLabels.join('')}
    ${legend}
  </svg>`;
}

function loadSummary(file) {
  if (!fs.existsSync(file)) return null;
  try {
    return JSON.parse(fs.readFileSync(file, 'utf8'));
  } catch {
    return null;
  }
}

function metricRow(label, m) {
  if (!m) return `<tr><td>${label}</td><td colspan="5">n/a</td></tr>`;
  // This k6 build flattens trend stats directly onto the metric object
  // (avg/med/p(90)/p(99)/max), rather than nesting under `.values`.
  const fmt = (x) => (x === undefined ? '-' : Number(x).toFixed(2));
  return `<tr><td>${label}</td><td>${fmt(m.avg)}</td><td>${fmt(m.med)}</td><td>${fmt(m['p(90)'])}</td><td>${fmt(m['p(99)'])}</td><td>${fmt(m.max)}</td></tr>`;
}

function buildTestSection(test) {
  const rawFile = path.join(runDir, 'raw', `${test.name}.ndjson`);
  const summaryFile = path.join(runDir, 'summary', `${test.name}.json`);
  const points = readNdjson(rawFile);
  const summary = loadSummary(summaryFile);

  if (points.length === 0) {
    return `<section><h2>${test.title}</h2><p><em>No data (scenario not run).</em></p></section>`;
  }

  const series = buildSeries(points);
  const latStats = statsFor(series, test.latencyMetric);
  const vus = series.seriesFor('vus', (vals) => Math.max(...vals));
  const errRate = series.seriesFor(
    'http_req_failed',
    (vals) => (vals.reduce((a, b) => a + b, 0) / vals.length) * 100
  );
  const httpReqs = series.seriesFor('http_reqs', (vals) => vals.length);
  const throughput = httpReqs.map((c) => (c === null ? null : c / (BUCKET_MS / 1000)));

  const latencyChart = svgLineChart({
    series: [
      { data: latStats.p50, color: COLORS.p50, label: 'p50' },
      { data: latStats.p90, color: COLORS.p90, label: 'p90' },
      { data: latStats.p99, color: COLORS.p99, label: 'p99' },
    ],
    yLabel: `${test.title} latency (${test.unit})`,
  });

  const loadChart = svgLineChart({
    series: [
      { data: vus, color: COLORS.vus, label: 'concurrent VUs' },
      { data: errRate, color: COLORS.err, label: 'error rate %' },
    ],
    yLabel: 'Concurrency & error rate',
  });

  const m = summary && summary.metrics ? summary.metrics : {};
  const summaryTable = `
    <table class="stats">
      <thead><tr><th>Metric</th><th>avg</th><th>p50</th><th>p90</th><th>p99</th><th>max</th></tr></thead>
      <tbody>
        ${metricRow(test.latencyMetric, m[test.latencyMetric])}
        ${metricRow('http_req_duration', m['http_req_duration'])}
      </tbody>
    </table>
    <p class="meta">
      total iterations: ${m.iterations?.count ?? '-'} |
      http reqs: ${m.http_reqs?.count ?? '-'} |
      error rate: ${m.http_req_failed?.value !== undefined ? (m.http_req_failed.value * 100).toFixed(2) + '%' : '-'} |
      peak VUs: ${m.vus_max?.value ?? '-'}
    </p>`;

  return `<section>
    <h2>${test.title}</h2>
    ${summaryTable}
    <div class="charts">
      <div>${latencyChart}</div>
      <div>${loadChart}</div>
    </div>
  </section>`;
}

const sections = TESTS.map(buildTestSection).join('\n');

const html = `<!doctype html>
<html>
<head>
<meta charset="utf-8">
<title>Sporacle load test report — ${path.basename(runDir)}</title>
<style>
  :root { color-scheme: light dark; }
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; margin: 2rem; background: #fff; color: #1a1a1a; }
  h1 { font-size: 1.4rem; }
  h2 { font-size: 1.1rem; margin-top: 2.5rem; border-bottom: 1px solid #ddd; padding-bottom: 4px; }
  table.stats { border-collapse: collapse; margin: 0.75rem 0; font-size: 0.85rem; }
  table.stats th, table.stats td { border: 1px solid #ddd; padding: 4px 10px; text-align: right; }
  table.stats th:first-child, table.stats td:first-child { text-align: left; }
  .meta { font-size: 0.8rem; color: #555; }
  .charts { display: flex; flex-wrap: wrap; gap: 24px; }
  .charts svg { max-width: 100%; height: auto; }
  @media (prefers-color-scheme: dark) {
    body { background: #16181d; color: #e6e6e6; }
    table.stats th, table.stats td { border-color: #333; }
    .meta { color: #aaa; }
  }
</style>
</head>
<body>
  <h1>Sporacle load test report</h1>
  <p class="meta">Run: ${path.basename(runDir)}</p>
  ${sections}
</body>
</html>`;

fs.writeFileSync(path.join(runDir, 'report.html'), html);
console.log(`Wrote ${path.join(runDir, 'report.html')}`);

// Regenerate the top-level index of all runs.
const reportsRoot = path.dirname(runDir);
const runs = fs
  .readdirSync(reportsRoot)
  .filter((d) => fs.existsSync(path.join(reportsRoot, d, 'report.html')))
  .sort()
  .reverse();
const indexHtml = `<!doctype html>
<html><head><meta charset="utf-8"><title>Sporacle load test runs</title>
<style>body{font-family:sans-serif;margin:2rem}li{margin:4px 0}</style></head>
<body><h1>Sporacle load test runs</h1><ul>
${runs.map((r) => `<li><a href="${r}/report.html">${r}</a></li>`).join('\n')}
</ul></body></html>`;
fs.writeFileSync(path.join(reportsRoot, 'index.html'), indexHtml);
