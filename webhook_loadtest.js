import http from 'k6/http';
import { check } from 'k6';

export const options = {
  scenarios: {
    ingestion_test: {
      executor: 'constant-vus',
      vus: 50,
      duration: '60s',
    },
  },
};

export default function () {
  const url = 'http://localhost:8000/events';

  const payload = JSON.stringify({
    event_type: "user.created",   // IMPORTANT: match your schema
    data: {
      id: "usr_" + __VU + "_" + __ITER,
      email: "loadtest@example.com"
    }
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
    timeout: '10s',
  };

  const res = http.post(url, payload, params);

  check(res, {
    'accepted (202)': (r) => r.status === 202,
  });
}
