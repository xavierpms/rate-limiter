import http from 'k6/http';

export default function () {
    let host = __ENV.RATELIMIT_HOST_TARGET
    let port = __ENV.RATELIMIT_PORT_TARGET
    let token_limit = __ENV.RATELIMIT_TOKEN_LIMIT_TARGET
    const URL = `http://${host}:${port}/hello`;
    const PARAMS = {
        headers: {
            'API_KEY': `Token${token_limit}`
        },
    };
    http.get(URL, PARAMS);
}

function numberOrDash(value, digits = 3) {
    if (typeof value !== 'number' || Number.isNaN(value)) {
        return '-';
    }
    return value.toFixed(digits);
}

function integerOrDash(value) {
    if (typeof value !== 'number' || Number.isNaN(value)) {
        return '-';
    }
    return String(Math.round(value));
}

function metricValue(metrics, metricName, field) {
    const metric = metrics[metricName];
    if (!metric || !metric.values) {
        return '-';
    }
    const value = metric.values[field];
    if (field === 'count' || field === 'passes' || field === 'fails') {
        return integerOrDash(value);
    }
    if (field === 'rate') {
        return numberOrDash(value, 4);
    }
    return numberOrDash(value, 3);
}

function metricRow(metrics, metricName, label) {
    return `<tr>
        <td>${label}</td>
        <td>${metricValue(metrics, metricName, 'count')}</td>
        <td>${metricValue(metrics, metricName, 'rate')}</td>
        <td>${metricValue(metrics, metricName, 'avg')}</td>
        <td>${metricValue(metrics, metricName, 'med')}</td>
        <td>${metricValue(metrics, metricName, 'p(95)')}</td>
        <td>${metricValue(metrics, metricName, 'max')}</td>
    </tr>`;
}

function rateStatus(rate) {
    if (typeof rate !== 'number' || Number.isNaN(rate)) {
        return { label: 'N/A', className: 'status-na' };
    }
    if (rate <= 0.05) {
        return { label: 'OK', className: 'status-ok' };
    }
    if (rate <= 0.2) {
        return { label: 'ATENÇÃO', className: 'status-warn' };
    }
    return { label: 'ALERTA', className: 'status-alert' };
}

function p95Status(p95) {
    if (typeof p95 !== 'number' || Number.isNaN(p95)) {
        return { label: 'N/A', className: 'status-na' };
    }
    if (p95 <= 50) {
        return { label: 'OK', className: 'status-ok' };
    }
    if (p95 <= 120) {
        return { label: 'ATENÇÃO', className: 'status-warn' };
    }
    return { label: 'ALERTA', className: 'status-alert' };
}

export function handleSummary(data) {
    const metrics = data.metrics || {};
    const host = __ENV.RATELIMIT_HOST_TARGET || 'unknown-host';
    const port = __ENV.RATELIMIT_PORT_TARGET || 'unknown-port';
    const tokenLimit = __ENV.RATELIMIT_TOKEN_LIMIT_TARGET || 'N/A';

    const totalRequests = metricValue(metrics, 'http_reqs', 'count');
    const failureRate = metricValue(metrics, 'http_req_failed', 'rate');
    const p95Duration = metricValue(metrics, 'http_req_duration', 'p(95)');
    const avgIteration = metricValue(metrics, 'iteration_duration', 'avg');

    const failureMetric = metrics.http_req_failed;
    const durationMetric = metrics.http_req_duration;
    const failureRateRaw = failureMetric && failureMetric.values ? failureMetric.values.rate : undefined;
    const p95DurationRaw = durationMetric && durationMetric.values ? durationMetric.values['p(95)'] : undefined;
    const failureBadge = rateStatus(failureRateRaw);
    const p95Badge = p95Status(p95DurationRaw);

    const html = `<!doctype html>
<html>
<head>
    <meta charset="utf-8" />
    <title>Summary - Token limiter</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 24px; color: #1f2937; }
        h1 { margin-bottom: 8px; }
        .subtitle { color: #4b5563; margin-bottom: 20px; }
        .cards { display: flex; flex-wrap: wrap; gap: 12px; margin-bottom: 20px; }
        .card { border: 1px solid #d1d5db; border-radius: 8px; padding: 10px 12px; min-width: 180px; }
        .card .label { font-size: 12px; color: #6b7280; }
        .card .value { font-size: 20px; font-weight: bold; margin-top: 4px; }
        .status { display: inline-block; margin-left: 8px; padding: 2px 8px; border-radius: 999px; font-size: 11px; font-weight: bold; }
        .status-ok { background: #dcfce7; color: #166534; }
        .status-warn { background: #fef3c7; color: #92400e; }
        .status-alert { background: #fee2e2; color: #991b1b; }
        .status-na { background: #e5e7eb; color: #374151; }
        .top3 { margin: 18px 0 14px 0; border: 1px solid #d1d5db; border-radius: 8px; padding: 10px 12px; }
        .top3 h2 { margin: 0 0 8px 0; font-size: 16px; }
        .top3 ul { margin: 0; padding-left: 18px; }
        .top3 li { margin: 4px 0; }
        table { border-collapse: collapse; width: 100%; margin-top: 10px; }
        th, td { border: 1px solid #d1d5db; padding: 8px; text-align: left; }
        th { background: #f3f4f6; }
    </style>
</head>
<body>
    <h1>Summary - Token limiter</h1>
    <div class="subtitle">Target: http://${host}:${port}/hello | Token: Token${tokenLimit}</div>

    <div class="cards">
        <div class="card"><div class="label">Total Requests</div><div class="value">${totalRequests}</div></div>
        <div class="card"><div class="label">Failure Rate</div><div class="value">${failureRate} <span class="status ${failureBadge.className}">${failureBadge.label}</span></div></div>
        <div class="card"><div class="label">P95 Duration (ms)</div><div class="value">${p95Duration} <span class="status ${p95Badge.className}">${p95Badge.label}</span></div></div>
        <div class="card"><div class="label">Avg Iteration (ms)</div><div class="value">${avgIteration}</div></div>
    </div>

    <section class="top3">
        <h2>Top 3 métricas-chave</h2>
        <ul>
            <li>Total Requests: <strong>${totalRequests}</strong></li>
            <li>Failure Rate: <strong>${failureRate}</strong> (${failureBadge.label})</li>
            <li>P95 Duration: <strong>${p95Duration} ms</strong> (${p95Badge.label})</li>
        </ul>
    </section>

    <table>
        <thead>
            <tr>
                <th>Metric</th>
                <th>Count</th>
                <th>Rate</th>
                <th>Avg</th>
                <th>Median</th>
                <th>P95</th>
                <th>Max</th>
            </tr>
        </thead>
        <tbody>
            ${metricRow(metrics, 'http_reqs', 'http_reqs')}
            ${metricRow(metrics, 'http_req_failed', 'http_req_failed')}
            ${metricRow(metrics, 'http_req_duration', 'http_req_duration (ms)')}
            ${metricRow(metrics, 'iteration_duration', 'iteration_duration (ms)')}
            ${metricRow(metrics, 'data_received', 'data_received (bytes)')}
            ${metricRow(metrics, 'data_sent', 'data_sent (bytes)')}
        </tbody>
    </table>
</body>
</html>`;

    return {
        "/home/k6/stress/summary-token.html": html,
    };
}
