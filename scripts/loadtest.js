import http from "k6/http";
import { check, sleep } from "k6";
import { Rate, Trend } from "k6/metrics";

// ── Custom metrics ────────────────────────────────────────────────────────────
const errorRate    = new Rate("error_rate");
const getPostTrend = new Trend("get_post_duration", true);

// ── Config ────────────────────────────────────────────────────────────────────
const BASE_URL = __ENV.BASE_URL || "http://localhost";

// Seed IDs to cycle through — adjust max to match how many posts are in your DB
const MAX_POST_ID = __ENV.MAX_POST_ID || 10;

export const options = {
  scenarios: {
    moderate_traffic: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "1m",  target: 50  },  // warm up
        { duration: "2m",  target: 1000 },  // ramp to moderate load
        { duration: "10m", target: 300 },
        { duration: "99h", target: 500 },  // hold forever — stop with Ctrl+C
      ],
    },
  },
  thresholds: {
    error_rate:        ["rate<0.01"],   // fail if errors exceed 1%
    get_post_duration: ["p(95)<500"],   // fail if p95 latency exceeds 500ms
  },
};

// ── Main VU loop ──────────────────────────────────────────────────────────────
export default function () {
  // Pick a random post ID
  const id  = Math.floor(Math.random() * MAX_POST_ID) + 1;
  const res = http.get(`${BASE_URL}/posts/${id}`);

  getPostTrend.add(res.timings.duration);

  const ok = check(res, {
    "GET /posts/:id → 200": (r) => r.status === 200,
    "response time < 500ms": (r) => r.timings.duration < 500,
  });

  errorRate.add(!ok);

  // Realistic think time between requests (1-3 seconds)
  sleep(0.1 + Math.random() * 0.4);
}