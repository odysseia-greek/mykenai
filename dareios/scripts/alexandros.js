import http from 'k6/http';
import { check, sleep } from 'k6';

const searchTerms = [
    'search?word=αγα&lang=greek&mode=partial',
    'search?word=σκω&lang=greek&mode=partial',
    'search?word=αθ&lang=greek&mode=partial',
    'search?word=θήναι&lang=greek&mode=partial',
    'search?word=Ἀθῆ&lang=greek&mode=partial',
    'search?word=ἀγάπ&lang=greek&mode=fuzzy',
    'search?word=ἀγά&lang=greek&mode=fuzzy',
    'search?word=ἀγάπη&lang=greek&mode=exact',
    'search?word=house&lang=english&mode=extended',
    'search?word=ezel&lang=dutch&mode=extended',
    'search?word=aanwijzen&lang=dutch&mode=exact',
    'search?word=house&lang=english&mode=exact',
    'search?word=πινύσκω&lang=greek&mode=exact',
    'search?word=Παναθήναια&lang=greek&mode=exact',
    'search?word=Ἀθηναῖος&language=greek&mode=exact&searchInText=true',
    'search?word=λόγος&language=greek&mode=exact&searchInText=true',
    'search?word=Λακεδαιμόνιος&language=greek&mode=exact&searchInText=true',
]; // Add more terms as needed

export const options = {
    stages: [
        { duration: '30s', target: 20 },
        { duration: '30s', target: 100 },
        { duration: '30s', target: 200 },
        { duration: '30s', target: 300 },
        { duration: '60s', target: 500 },
        { duration: '30s', target: 250 },
        { duration: '30s', target: 100 },
        { duration: '30s', target: 50 },
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
