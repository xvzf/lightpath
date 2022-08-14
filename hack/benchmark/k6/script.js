import http from 'k6/http';
import {check, sleep} from 'k6';


export const options = {
  // discardResponseBodies: true,
  scenarios: {
    rampup: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '120s', target: 500 },
        { duration: '60s', target: 0 },
      ],
      gracefulRampDown: '0s',
    },
  },
}

export default function() {
  let res = http.post(`http://${__ENV.ISOTOPE_ENTRYPOINT}/`);
  check(res, { 'success': (r) => r.status === 200 });
  sleep(0.05); // sleep for 5ms
}
