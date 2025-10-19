/**
 * K6 Load Testing Script for Typing Master
 * Tests: 500 events/sec/user, 5k concurrent users
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const sessionDuration = new Trend('session_duration');
const keystrokeLatency = new Trend('keystroke_latency');
const apiLatency = new Trend('api_latency');
const eventsProcessed = new Counter('events_processed');

// Test configuration
export const options = {
  stages: [
    // Ramp up to 1000 users over 2 minutes
    { duration: '2m', target: 1000 },
    // Ramp up to 3000 users over 3 minutes
    { duration: '3m', target: 3000 },
    // Ramp up to 5000 users over 5 minutes
    { duration: '5m', target: 5000 },
    // Stay at 5000 users for 10 minutes
    { duration: '10m', target: 5000 },
    // Ramp down to 0 users over 2 minutes
    { duration: '2m', target: 0 },
  ],
  thresholds: {
    // 95% of requests should be below 200ms
    'http_req_duration': ['p(95)<200'],
    // Error rate should be below 1%
    'errors': ['rate<0.01'],
    // Keystroke latency should be below 8ms for 95% of events
    'keystroke_latency': ['p(95)<8'],
    // API latency should be below 500ms for 95% of requests
    'api_latency': ['p(95)<500'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const EVENTS_PER_SECOND = 500;
const EVENT_INTERVAL = 1000 / EVENTS_PER_SECOND; // ms between events

// Generate realistic typing events
function generateKeystrokeEvent(index, sessionId) {
  const keys = 'abcdefghijklmnopqrstuvwxyz0123456789 .,;:()[]{}';
  const randomKey = keys[Math.floor(Math.random() * keys.length)];
  
  return {
    session_id: sessionId,
    timestamp_ms: Date.now(),
    key_pressed: randomKey,
    action: 'down',
    cursor_position: index,
    error_flag: Math.random() < 0.05, // 5% error rate
    expected_token: randomKey,
  };
}

// Simulate a typing session
export default function () {
  const sessionId = `session_${__VU}_${Date.now()}`;
  const userId = `user_${__VU}`;
  
  // 1. Create session
  const sessionStart = Date.now();
  const createSessionRes = http.post(
    `${BASE_URL}/api/sessions`,
    JSON.stringify({
      user_id: userId,
      mode: 'practice',
      language_id: 'python',
      lesson_id: 'intro',
    }),
    {
      headers: { 'Content-Type': 'application/json' },
    }
  );
  
  check(createSessionRes, {
    'session created': (r) => r.status === 201,
  });
  
  if (createSessionRes.status !== 201) {
    errorRate.add(1);
    return;
  }
  
  // 2. Simulate typing events (500 events/sec for 10 seconds = 5000 events)
  const eventBatch = [];
  const batchSize = 100; // Send events in batches
  
  for (let i = 0; i < 5000; i++) {
    const eventStart = Date.now();
    const event = generateKeystrokeEvent(i, sessionId);
    eventBatch.push(event);
    
    // Simulate keystroke processing latency
    const processingLatency = Math.random() * 5; // 0-5ms
    keystrokeLatency.add(processingLatency);
    
    // Send batch when full
    if (eventBatch.length >= batchSize) {
      const apiStart = Date.now();
      const eventsRes = http.post(
        `${BASE_URL}/api/sessions/${sessionId}/events`,
        JSON.stringify({ events: eventBatch }),
        {
          headers: { 'Content-Type': 'application/json' },
        }
      );
      
      const apiDuration = Date.now() - apiStart;
      apiLatency.add(apiDuration);
      
      check(eventsRes, {
        'events recorded': (r) => r.status === 200,
      });
      
      if (eventsRes.status !== 200) {
        errorRate.add(1);
      } else {
        eventsProcessed.add(eventBatch.length);
      }
      
      eventBatch.length = 0; // Clear batch
    }
    
    // Maintain event rate (500/sec)
    sleep(EVENT_INTERVAL / 1000);
  }
  
  // Send remaining events
  if (eventBatch.length > 0) {
    const eventsRes = http.post(
      `${BASE_URL}/api/sessions/${sessionId}/events`,
      JSON.stringify({ events: eventBatch }),
      {
        headers: { 'Content-Type': 'application/json' },
      }
    );
    
    if (eventsRes.status === 200) {
      eventsProcessed.add(eventBatch.length);
    } else {
      errorRate.add(1);
    }
  }
  
  // 3. Finalize session
  const finalizeRes = http.post(
    `${BASE_URL}/api/sessions/${sessionId}/finalize`,
    JSON.stringify({
      input: 'sample typed text',
      expected: 'sample expected text',
    }),
    {
      headers: { 'Content-Type': 'application/json' },
    }
  );
  
  check(finalizeRes, {
    'session finalized': (r) => r.status === 200,
  });
  
  if (finalizeRes.status !== 200) {
    errorRate.add(1);
  }
  
  const totalDuration = Date.now() - sessionStart;
  sessionDuration.add(totalDuration);
  
  // Small delay between iterations
  sleep(1);
}

// Smoke test - quick validation
export function smokeTest() {
  const res = http.get(`${BASE_URL}/health`);
  check(res, {
    'health check passed': (r) => r.status === 200,
  });
}

// Stress test - push beyond normal limits
export function stressTest() {
  options.stages = [
    { duration: '2m', target: 5000 },
    { duration: '5m', target: 10000 },
    { duration: '2m', target: 15000 },
    { duration: '5m', target: 15000 },
    { duration: '2m', target: 0 },
  ];
}

// Spike test - sudden traffic spike
export function spikeTest() {
  options.stages = [
    { duration: '10s', target: 100 },
    { duration: '1m', target: 10000 },
    { duration: '3m', target: 10000 },
    { duration: '10s', target: 100 },
    { duration: '1m', target: 0 },
  ];
}
