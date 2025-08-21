import http from 'k6/http';
import { sleep, check } from 'k6';
import encoding from 'k6/encoding';

export const options = {
  scenarios: {
    search_events_load_test: {
      executor: 'ramping-arrival-rate',
      startRate: 0,
      timeUnit: '1s',
      preAllocatedVUs: 50,
      maxVUs: 100,
      stages: [
        { duration: '2m', target: 50 },
        { duration: '5m', target: 100 },
        { duration: '1m', target: 0 },  
      ],
    },
  },
};

const searchPhrases = [
  'концерт',
  'Дмитрий Козлов',
  'азарт',
  'Чемпионат',
  'Красная шапочка',
  'Экспозиция "Звездные войны"'
];

const users = [
  { email: 'aysultan_talgat_1@fest.tix', password: '/8eC$AD>' },
  { email: 'ayaulym_bazarbaeva_3@quick.pass', password: 'LDb60_%]4' },
  { email: 'sultan_sultanov_2@show.go', password: '*IVSf?kh)xa' }
];

const months = [
  '2024-01', '2024-02', '2024-03', '2024-04', '2024-05', '2024-06',
  '2024-07', '2024-08', '2024-09', '2024-10', '2024-11', '2024-12'
];

function getRandomElement(array) {
  return array[Math.floor(Math.random() * array.length)];
}

function getRandomInt(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

function getRandomDate(month) {
  const year = month.split('-')[0];
  const monthNum = month.split('-')[1];
  const daysInMonth = new Date(year, monthNum, 0).getDate();
  const day = getRandomInt(1, daysInMonth);
  return `${year}-${monthNum}-${String(day).padStart(2, '0')}`;
}

function createBasicAuthHeader(user) {
  const credentials = `${user.email}:${user.password}`;
  const encodedCredentials = encoding.b64encode(credentials);
  return `Basic ${encodedCredentials}`;
}

export default function () {
  const baseUrl = __ENV.API_URL;
  const user = getRandomElement(users);
  
  const params = {
    headers: {
      'Authorization': createBasicAuthHeader(user),
      'Accept': 'application/json'
    },
    timeout: '30s'
  };

  const testScenarios = [
    // 1. Search with query phrase
    () => {
      const query = getRandomElement(searchPhrases);
      const page = getRandomInt(1, 50);
      const pageSize = getRandomInt(10, 20); // API max is 20
      return `${baseUrl}/api/events?query=${encodeURIComponent(query)}&page=${page}&pageSize=${pageSize}`;
    },
    
    // 2. Search with query and specific date
    () => {
      const query = getRandomElement(searchPhrases);
      const page = getRandomInt(1, 50);
      const pageSize = getRandomInt(10, 20);
      return `${baseUrl}/api/events?query=${encodeURIComponent(query)}&date=2024-12-25&page=${page}&pageSize=${pageSize}`;
    },
    
    // 3. Search with query and date range from 2024-12-25 to 2024-12-31
    () => {
      const query = getRandomElement(searchPhrases);
      const page = getRandomInt(1, 50);
      const pageSize = getRandomInt(10, 20);
      const endDate = getRandomInt(25, 31);
      return `${baseUrl}/api/events?query=${encodeURIComponent(query)}&date=2024-12-${endDate}&page=${page}&pageSize=${pageSize}`;
    },
    
    // 4. Search with query and random month date
    () => {
      const query = getRandomElement(searchPhrases);
      const randomMonth = getRandomElement(months);
      const randomDate = getRandomDate(randomMonth);
      const page = getRandomInt(1, 50);
      const pageSize = getRandomInt(10, 20);
      return `${baseUrl}/api/events?query=${encodeURIComponent(query)}&date=${randomDate}&page=${page}&pageSize=${pageSize}`;
    },
    
    // 5. Search without date (only query)
    () => {
      const query = getRandomElement(searchPhrases);
      const page = getRandomInt(1, 50);
      const pageSize = getRandomInt(10, 20);
      return `${baseUrl}/api/events?query=${encodeURIComponent(query)}&page=${page}&pageSize=${pageSize}`;
    }
  ];

  const scenario = getRandomElement(testScenarios);
  const url = scenario();

  const response = http.get(url, params);

  check(response, {
    'status is 200': (r) => r.status === 200,
    'response has events': (r) => {
      try {
        const data = JSON.parse(r.body);
        return Array.isArray(data);
      } catch {
        return false;
      }
    },
    'response time < 2000ms': (r) => r.timings.duration < 2000,
  });

  if (response.status !== 200) {
    console.error(`❌ Request failed: ${response.status} - ${url}`);
    console.error(`📝 Response body: ${response.body}`);
    console.error(`🔑 Auth header: ${params.headers.Authorization.substring(0, 20)}...`);
  } else {
    // Uncomment for debugging successful requests
    // console.log(`✅ Success: ${response.status} - ${url}`);
  }

  sleep(Math.random() * 2 + 0.5);
}

export function setup() {
  return {
    startTime: Date.now(),
    testVersion: 'v1.0.0',
    environment: 'search_events_load_test'
  };
}

export function teardown(data) {
  const duration = (Date.now() - data.startTime) / 1000;
  console.log(`✅ Search Events Load Test completed in ${duration}s`);
  console.log(`📈 Check metrics for search performance analysis`);
  console.log(`   Target: ${data.environment}`);
  console.log(`   Scenarios tested: query search, date filtering, pagination`);
}
