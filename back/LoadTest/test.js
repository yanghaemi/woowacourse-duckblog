import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    scenarios: {
    constant_request_rate: {
      executor: 'constant-arrival-rate',
      rate: 5000,        // 5000 req/s로 도전!
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 100,
      maxVUs: 200,
    },
  },
};

export default function () {
    // GET /albums
    let res = http.get('http://localhost:8080/albums');
    check(res, {
        'status is 200': (r) => r.status === 200,
    });

}