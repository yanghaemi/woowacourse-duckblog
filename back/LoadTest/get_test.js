import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  scenarios: {
    ramping: {
      executor: 'ramping-arrival-rate',
      startRate: 1000,
      timeUnit: '1s',
      preAllocatedVUs: 50,
      maxVUs: 200,
      stages: [
        { target: 1000, duration: '30s' }, // 1000 (안정)
        { target: 1500, duration: '30s' }, // 1500 (도전)
        { target: 2000, duration: '30s' }, // 2000 (한계)
        { target: 2500, duration: '30s' }, // 2500 (무리)
      ],
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