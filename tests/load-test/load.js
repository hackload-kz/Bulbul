import http from 'k6/http';
import { sleep, check } from 'k6';

export const options = {
  scenarios: {
    open_model_test: {
      executor: 'ramping-arrival-rate',
      startRate: 0,
      timeUnit: '1s',
      preAllocatedVUs: 2000,
      maxVUs: 4000,
      stages: [
        // { duration: '20s', target: 3 },
        { duration: '1m', target: 1000 },
        { duration: '4m', target: 4000 },
        { duration: '1m', target: 0 },  
      ],
    },
  },
};

export default function () {
  const url = `${__ENV.API_URL}/api/events?page=1&pageSize=20`;
  
  const params = {};

  if (__ENV.BASIC_AUTH && __ENV.BASIC_AUTH.length > 0) {
    params.headers = {
      'Authorization': `Basic ${__ENV.BASIC_AUTH}`,
    };
  }

  const response = http.get(url, params);

  check(response, {
    'status is 200': (r) => r.status === 200,
  });

  sleep(1);
}

export function setup() {
  return {
    startTime: Date.now(),
    testVersion: 'v2.1.0',
    environment: 'load_test'
  };
}

export function teardown(data) {
  const duration = (Date.now() - data.startTime) / 1000;
  console.log(`✅ Test completed in ${duration}s`);
  console.log(`📈 Check Prometheus for metrics with labels:`);
  console.log(`   scnr=tickets, tm=drim, environment=${data.environment}`);
}