import http from 'k6/http';
import { check } from 'k6';

export const options = {
    scenarios: {
        accept_order_rps_5000: {
            executor: 'constant-arrival-rate',

            rate: 5000,
            timeUnit: '1s',

            duration: '3m',
            preAllocatedVUs: 600,
            maxVUs: 2000,
        },
    },

    thresholds: {
        http_req_failed:   ['rate<0.01'],
        http_req_duration: ['p(95)<400'],
    },
};

const BASE =  'http://localhost:8080';

function rand(max) {
    return Math.floor(Math.random() * max) + 1;
}

function futureDate(days) {
    const d = new Date();
    d.setDate(d.getDate() + days);
    return d.toISOString();
}

export default function () {
    const payload = JSON.stringify({
        order_id:  rand(1e12),
        user_id:   rand(1e6),
        expires_at: futureDate(7),
        package:   'PACKAGE_TYPE_BOX',
        weight:    1.5,
        price:     100.0,
    });

    const res = http.post(`${BASE}/v1/orders/accept`, payload, {
        headers: { 'Content-Type': 'application/json' },
        tags:    { name: 'AcceptOrder' },
    });

    check(res, {
        '200 OK':            r => r.status === 200,
        'order_id returned': r => r.json('order_id') !== undefined,
    });
}