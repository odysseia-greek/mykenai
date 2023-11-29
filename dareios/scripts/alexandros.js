import http from 'k6/http';
import { check, sleep } from 'k6';

const searchTerms = [
    'search?word=αγα&lang=greek&mode=fuzzy',
    'search?word=σκω&lang=greek&mode=fuzzy',
    'search?word=αθ&lang=greek&mode=fuzzy',
    'search?word=θήναι&lang=greek&mode=fuzzy',
    'search?word=ἀγάπη&lang=greek&mode=exact',
    'search?word=ezel&lang=dutch&mode=phrase',
    'search?word=aanwijzen&lang=dutch&mode=exact',
    'search?word=house&lang=english&mode=exact',
    'search?word=πινύσκω&lang=greek&mode=exact',
    'search?word=Παναθήναια&lang=greek&mode=exact'
]; // Add more terms as needed

export const options = {
    stages: [
        { duration: '30s', target: 20 },
        { duration: '30s', target: 100 },
        { duration: '60s', target: 500 },
        { duration: '30s', target: 1000 },
        { duration: '10s', target: 200 },
        { duration: '20s', target: 10 },
    ],
};

export default function () {
    const searchTerm = searchTerms[Math.floor(Math.random() * searchTerms.length)];
    const res = http.get(`http://k3s-odysseia.api.greek/alexandros/v1/${searchTerm}`);
    check(res, { 'status was 200': (r) => r.status == 200 });
    const randomSleep = Math.random() * 900 + 100; // Random sleep between 100ms and 1000ms
    sleep(randomSleep / 1000); // sleep function expects seconds, so divide by 1000
}
