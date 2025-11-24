import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  scenarios: {
    write_load: {
      executor: 'ramping-arrival-rate',
      startRate: 100,
      timeUnit: '1s',
      preAllocatedVUs: 50,
      maxVUs: 200,
      stages: [
        { target: 200, duration: '10s' },
        { target: 300, duration: '10s' },
        { target: 400, duration: '10s' },
        { target: 500, duration: '20s' }, // 최종 목표
      ],
    },
  },
};

export default function () {
  const payload = JSON.stringify({
    title: `load-test-title-${Math.random()}`,
    artist: `artist-${Math.random()}`,
    price: Math.floor(Math.random() * 100),
  });

  const headers = { 'Content-Type': 'application/json' };

  let res = http.post('http://localhost:8080/albums', payload, { headers });

  check(res, {
    'status is 201': (r) => r.status === 201,
  });

  sleep(0.1);
}
