import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    scenarios: {
        // baseline_10_vu: {
        //     executor: 'constant-vus',
        //     exec: 'loadTest',
        //     duration: '1m',
        //     gracefulStop: '5s',
        //     vus: 1,
        //     tags: {baseline: 'low'},
        // },
        baseline_500_vu: {
            executor: 'ramping-vus',
            exec: 'loadTest',
            startVUs: 10,
            stages: [
                { duration: '10s', target: 20 },
                { duration: '20s', target: 100 },
                { duration: '30s', target: 200 },
                { duration: '1m', target: 300 },
                { duration: '1m', target: 400 },
                { duration: '2m', target: 500 },
            ],
            tags: {baseline: 'high'},
        },
    },
};

const alexandrosQueryTemplate = `query dictionary {
    dictionary(word: $word, language: $language, mode: $mode, searchInText: $searchInText) {
        hits {
            hit {
                english
                greek
            }
            foundInText {
                rhemai {
                    author
                    greek
                    translations
                }
            }
        }
    }
}`;

const alexandrosVariables = [
    { word: "αγα", language: "greek", mode: "partial", searchInText: false },
    { word: "σκω", language: "greek", mode: "partial", searchInText: false },
    { word: "αθ", language: "greek", mode: "partial", searchInText: false },
    { word: "θήναι", language: "greek", mode: "partial", searchInText: false },
    { word: "Ἀθῆ", language: "greek", mode: "partial", searchInText: false },
    { word: "ἀγάπ", language: "greek", mode: "fuzzy", searchInText: false },
    { word: "ἀγά", language: "greek", mode: "fuzzy", searchInText: false },
    { word: "ἀγάπη", language: "greek", mode: "exact", searchInText: false },
    { word: "house", language: "english", mode: "extended", searchInText: false },
    { word: "ezel", language: "dutch", mode: "extended", searchInText: false },
    { word: "aanwijzen", language: "dutch", mode: "exact", searchInText: false },
    { word: "house", language: "english", mode: "exact", searchInText: false },
    { word: "πινύσκω", language: "greek", mode: "exact", searchInText: false },
    { word: "Παναθήναια", language: "greek", mode: "exact", searchInText: false },
    { word: "ἀγάπη", language: "greek", mode: "exact", searchInText: false },
    { word: "Ἀθηναῖος", language: "greek", mode: "exact", searchInText: true },
    { word: "λόγος", language: "greek", mode: "exact", searchInText: true },
    { word: "Λακεδαιμόνιος", language: "greek", mode: "exact", searchInText: true },
    { word: "πεμπω", language: "greek", mode: "exact", searchInText: true },
    { word: "υπαρχω", language: "greek", mode: "exact", searchInText: true },
];


const herodotosQueries = [
    `query authors{\n\tauthors{\n\t\tname\n\t\tbooks {\n\t\t\tbook\n\t\t}\n\t}\n}\n`,
    `query sentence{\n\tsentence(author: "aristophanes", book: "1"){\n\t\tid\n\t\tgreek\n\t\tauthor\n\t\tbook\n\t}\n}\n`,
    `query sentence{\n\tsentence(author: "aeschylus", book: "1"){\n\t\tid\n\t\tgreek\n\t\tauthor\n\t\tbook\n\t}\n}\n`,
    `query sentence{\n\tsentence(author: "ploutarchos", book: "1"){\n\t\tid\n\t\tgreek\n\t\tauthor\n\t\tbook\n\t}\n}\n`,
    `query sentence{\n\tsentence(author: "ploutarchos", book: "2"){\n\t\tid\n\t\tgreek\n\t\tauthor\n\t\tbook\n\t}\n}\n`,
]

const sokratesQueries = [
    `query options{\n\toptions(quizType: "media") {\n\t\taggregates{\n\t\t\tname\n\t\t\thighestSet\n\t\t}\n\t}\n}\n`,
    `query options{\n\toptions(quizType: "dialogue") {\n\t\taggregates{\n\t\t\tname\n\t\t\thighestSet\n\t\t}\n\t}\n}\n`,
    `query options{\n\toptions(quizType: "authorbased") {\n\t\taggregates{\n\t\t\tname\n\t\t\thighestSet\n\t\t}\n\t}\n}\n`,
    `query quiz{\n\tquiz(theme: "Aristophanes - Frogs", set: "12", quizType: "authorbased") {\n\t\t... on QuizResponse {\n\t\t\tquizItem\n\t\t\toptions{\n\t\t\t\toption\n\t\t\t}\n\t\t}\n\t}\n}\n`,
    `query quiz{\n\tquiz(theme: "Herodotos - Clio", set: "135", quizType: "authorbased") {\n\t\t... on QuizResponse {\n\t\t\tquizItem\n\t\t\toptions{\n\t\t\t\toption\n\t\t\t}\n\t\t}\n\t}\n}\n`,
    `query quiz{\n\tquiz(theme: "Plato - Apology", set: "1", quizType: "authorbased") {\n\t\t... on QuizResponse {\n\t\t\tquizItem\n\t\t\toptions{\n\t\t\t\toption\n\t\t\t}\n\t\t}\n\t}\n}\n`,
    `query quiz{\n\tquiz(set: "1", theme: "Basic", quizType: "media") {\n\t\t... on QuizResponse {\n\t\t\tquizItem\n\t\t\toptions{\n\t\t\t\toption\n\t\t\t\timageUrl\n\t\t\t}\n\t\t}\n\t}\n}\n`,
    `query quiz{\n\tquiz(set: "2", theme: "Daily Life", quizType: "media") {\n\t\t... on QuizResponse {\n\t\t\tquizItem\n\t\t\toptions{\n\t\t\t\toption\n\t\t\t\timageUrl\n\t\t\t}\n\t\t}\n\t}\n}\n`,
    `query quiz{\n\tquiz(theme: "Plato - Euthyphro", set: "1", quizType: "dialogue") {\n\t\t... on QuizResponse {\n\t\t\tquizItem\n\t\t\toptions{\n\t\t\t\toption\n\t\t\t}\n\t\t}\n\t\t... on DialogueQuiz {\n\t\t\tquizType\n\t\t\tdialogue{\n\t\t\t\tintroduction\n\t\t\t\tspeakers {\n\t\t\t\t\tshorthand\n\t\t\t\t\ttranslation\n\t\t\t\t}\n\t\t\t}\n\t\t\tcontent{\n\t\t\t\ttranslation\n\t\t\t\tgreek\n\t\t\t\tplace\n\t\t\t\tspeaker\n\t\t\t}\n\t\t}\n\t}\n}\n`,
    `query quiz{\n\tquiz(theme: "Euripides - Medea", set: "1", quizType: "dialogue") {\n\t\t... on QuizResponse {\n\t\t\tquizItem\n\t\t\toptions{\n\t\t\t\toption\n\t\t\t}\n\t\t}\n\t\t... on DialogueQuiz {\n\t\t\tquizType\n\t\t\tdialogue{\n\t\t\t\tintroduction\n\t\t\t\tspeakers {\n\t\t\t\t\tshorthand\n\t\t\t\t\ttranslation\n\t\t\t\t}\n\t\t\t}\n\t\t\tcontent{\n\t\t\t\ttranslation\n\t\t\t\tgreek\n\t\t\t\tplace\n\t\t\t\tspeaker\n\t\t\t}\n\t\t}\n\t}\n}\n`,
    `query answer {\n\tanswer(\n\t\ttheme: "Aristophanes - Frogs"\n\t\tset: "12"\n\t\tquizType: "authorbased"\n\t\tquizWord: "κανών"\n\t\tanswer: "any straight rod"\n\t\tcomprehensive: true\n\t) {\n\t\t... on ComprehensiveResponse {\n\t\t\tcorrect\n\t\t\tquizWord\n\t\t\tsimilarWords {\n\t\t\t\tgreek\n\t\t\t\tenglish\n\t\t\t}\n\t\t\tfoundInText {\n\t\t\t\trhemai {\n\t\t\t\t\tauthor\n\t\t\t\t\tgreek\n\t\t\t\t\ttranslations\n\t\t\t\t}\n\t\t\t}\n\t\t\tprogress {\n\t\t\t\taverageAccuracy\n\t\t\t\ttimesCorrect\n\t\t\t\ttimesIncorrect\n\t\t\t}\n\t\t}\n\t}\n}\n`,
    `query answer {\n\tanswer(\n\t\ttheme: "Aristophanes - Frogs"\n\t\tset: "12"\n\t\tquizType: "authorbased"\n\t\tquizWord: "κανών"\n\t\tanswer: "any straight rod"\n\t\tcomprehensive: false\n\t) {\n\t\t... on ComprehensiveResponse {\n\t\t\tcorrect\n\t\t\tquizWord\n\t\t\tsimilarWords {\n\t\t\t\tgreek\n\t\t\t\tenglish\n\t\t\t}\n\t\t\tfoundInText {\n\t\t\t\trhemai {\n\t\t\t\t\tauthor\n\t\t\t\t\tgreek\n\t\t\t\t\ttranslations\n\t\t\t\t}\n\t\t\t}\n\t\t\tprogress {\n\t\t\t\taverageAccuracy\n\t\t\t\ttimesCorrect\n\t\t\t\ttimesIncorrect\n\t\t\t}\n\t\t}\n\t}\n}\n`,
    `query answer {\n\tanswer(\n\t\ttheme: "Plato - Euthyphro"\n\t\tset: "1"\n\t\tquizType: "dialogue"\n\t\tdialogue: [\n  {\n    greek: "ἀλλʼ ἴσως οὐδὲν ἔσται, ὦ Σώκρατες, πρᾶγμα, ἀλλὰ σύ τε κατὰ νοῦν ἀγωνιῇ τὴν δίκην, οἶμαι δὲ καὶ ἐμὲ τὴν ἐμήν.",\n    place: 1,\n    speaker: "ΕΥΘ",\n    translation: "Well, Socrates, perhaps it won’t amount to much, and you will bring your case to a satisfactory ending, as I think I shall mine."\n  },\n  {\n    greek: "ὁ σός, ὦ βέλτιστε;",\n    place: 2,\n    speaker: "ΣΩ",\n    translation: "Your father, my dear man?"\n  },\n  {\n    greek: "διώκω.",\n    place: 3,\n    speaker: "ΕΥΘ",\n    translation: "Prosecuting."\n  },\n  {\n    greek: "ἔστιν δὲ τί τὸ ἔγκλημα καὶ τίνος ἡ δίκη;",\n    place: 12,\n    speaker: "ΣΩ",\n    translation: "But what is the charge, and what is the suit about?"\n  },\n  {\n    greek: "ὃν διώκων αὖ δοκῶ μαίνεσθαι.",\n    place: 5,\n    speaker: "ΕΥΘ",\n    translation: "Such a man that they think I am insane because I am prosecuting him."\n  },\n  {\n    greek: "τίνα;",\n    place: 4,\n    speaker: "ΣΩ",\n    translation: "Whom?"\n  },\n  {\n    greek: "πολλοῦ γε δεῖ πέτεσθαι, ὅς γε τυγχάνει ὢν εὖ μάλα πρεσβύτης.",\n    place: 7,\n    speaker: "ΕΥΘ",\n    translation: "No flying for him at his ripe old age."\n  },\n  {\n    greek: "τί δέ; πετόμενόν τινα διώκεις;",\n    place: 6,\n    speaker: "ΣΩ",\n    translation: "Why? Are you prosecuting one who has wings to fly away with?"\n  },\n  {\n    greek: "ὁ ἐμὸς πατήρ.",\n    place: 9,\n    speaker: "ΕΥΘ",\n    translation: "My father."\n  },\n  {\n    greek: "τίς οὗτος;",\n    place: 8,\n    speaker: "ΣΩ",\n    translation: "Who is he?"\n  },\n  {\n    greek: "πάνυ μὲν οὖν.",\n    place: 11,\n    speaker: "ΕΥΘ",\n    translation: "Certainly."\n  },\n  {\n    greek: "ἔστιν δὲ δὴ τῶν οἰκείων τις ὁ τεθνεὼς ὑπὸ τοῦ σοῦ πατρός; ἢ δῆλα δή; οὐ γὰρ ἄν που ὑπέρ γε ἀλλοτρίου ἐπεξῇσθα φόνου αὐτῷ.",\n    place: 16,\n    speaker: "ΣΩ",\n    translation: "Is the one who was killed by your father a relative? But of course he was; for you would not bring a charge of murder against him on a stranger’s account."\n  },\n  {\n    greek: "φόνου, ὦ Σώκρατες.",\n    place: 13,\n    speaker: "ΕΥΘ",\n    translation: "Murder, Socrates."\n  },\n  {\n    greek: "Ἡράκλεις. ἦ που, ὦ Εὐθύφρων, ἀγνοεῖται ὑπὸ τῶν πολλῶν ὅπῃ ποτὲ ὀρθῶς ἔχει· οὐ γὰρ οἶμαί γε τοῦ ἐπιτυχόντος ὀρθῶς αὐτὸ πρᾶξαι ἀλλὰ πόρρω που ἤδη σοφίας ἐλαύνοντος.",\n    place: 14,\n    speaker: "ΣΩ",\n    translation: "Heracles! Surely, Euthyphro, most people do not know where the right lies; for I fancy it is not everyone who can rightly do what you are doing, but only one who is already very far advanced in wisdom."\n  },\n  {\n    greek: "πόρρω μέντοι νὴ Δία, ὦ Σώκρατες.",\n    place: 15,\n    speaker: "ΕΥΘ",\n    translation: "Very far, indeed, Socrates, by Zeus."\n  },\n  {\n    greek: "ἔστιν δὲ δὴ σοί, ὦ Εὐθύφρων, τίς ἡ δίκη; φεύγεις αὐτὴν ἢ διώκεις;",\n    place: 2,\n    speaker: "ΣΩ",\n    translation: "What is your case, Euthyphro? Are you defending or prosecuting?"\n  }\n]\n\t) {\n\t\t... on DialogueAnswer {\n\t\t\tpercentage\n\t\t\tinput{\n\t\t\t\tplace\n\t\t\t}\n\t\t\twronglyPlaced{\n\t\t\t\tgreek\n\t\t\t\ttranslation\n\t\t\t\tspeaker\n\t\t\t\tplace\n\t\t\t}\n\t\t}\n\t}\n}\n`,
    `query answer{\n\tanswer(\n\t\ttheme: "Daily Life"\n\t\tset: "1"\n\t\tquizType: "media"\n\t\tquizWord: "πόλις"\n\t\tanswer: "word"\n\t\tcomprehensive: true\n\t) {\n\t\t\t\t... on ComprehensiveResponse {\n\t\tcorrect\n\t\tquizWord\n\t\tsimilarWords{\n\t\t\tgreek\n\tenglish\n\t\t}\n\t\tfoundInText{\n\t\t\trhemai{\n\t\t\t\tauthor\n\t\t\t\tgreek\n\t\t\t\ttranslations\n\t\t\t}\n\t\t}\n\t}\n\t}\n}\n`,
    `query answer{\n\tanswer(\n\t\ttheme: "Daily Life"\n\t\tset: "1"\n\t\tquizType: "media"\n\t\tquizWord: "πόλις"\n\t\tanswer: "word"\n\t\tcomprehensive: false\n\t) {\n\t\t\t\t... on ComprehensiveResponse {\n\t\tcorrect\n\t\tquizWord\n\t\tsimilarWords{\n\t\t\tgreek\n\tenglish\n\t\t}\n\t\tfoundInText{\n\t\t\trhemai{\n\t\t\t\tauthor\n\t\t\t\tgreek\n\t\t\t\ttranslations\n\t\t\t}\n\t\t}\n\t}\n\t}\n}\n`,

]
const dionysiosQueries =  [
    `query grammar{\n\tgrammar(word: "πέμπει") {\n\t\ttranslation\n\t\tword\n\t\trule\n\t\trootWord\n\t}\n}\n`,
    `query grammar{\n\tgrammar(word: "ἔβαλλε") {\n\t\ttranslation\n\t\tword\n\t\trule\n\t\trootWord\n\t}\n}\n`,
    `query grammar{\n\tgrammar(word: "φέροντος") {\n\t\ttranslation\n\t\tword\n\t\trule\n\t\trootWord\n\t}\n}\n`,
    `query grammar{\n\tgrammar(word: "ἀληθῆ") {\n\t\ttranslation\n\t\tword\n\t\trule\n\t\trootWord\n\t}\n}\n`,
    `query grammar{\n\tgrammar(word: "πάντων") {\n\t\ttranslation\n\t\tword\n\t\trule\n\t\trootWord\n\t}\n}\n`,
    `query grammar{\n\tgrammar(word: "ἐκεῖνον") {\n\t\ttranslation\n\t\tword\n\t\trule\n\t\trootWord\n\t}\n}\n`,
    `query grammar{\n\tgrammar(word: "Πελοποννησίους") {\n\t\ttranslation\n\t\tword\n\t\trule\n\t\trootWord\n\t}\n}\n`,
    `query grammar{\n\tgrammar(word: "κατέλιπον") {\n\t\ttranslation\n\t\tword\n\t\trule\n\t\trootWord\n\t}\n}\n`,
    `query grammar{\n\tgrammar(word: "στρατιὰ") {\n\t\ttranslation\n\t\tword\n\t\trule\n\t\trootWord\n\t}\n}\n`,
    `query grammar{\n\tgrammar(word: "Ἀθηναίων") {\n\t\ttranslation\n\t\tword\n\t\trule\n\t\trootWord\n\t}\n}\n`,
    `query grammar{\n\tgrammar(word: "ἀνήγετο") {\n\t\ttranslation\n\t\tword\n\t\trule\n\t\trootWord\n\t}\n}\n`,
    `query grammar{\n\tgrammar(word: "ἐν") {\n\t\ttranslation\n\t\tword\n\t\trule\n\t\trootWord\n\t}\n}\n`,
    `query grammar{\n\tgrammar(word: "ἦν") {\n\t\ttranslation\n\t\tword\n\t\trule\n\t\trootWord\n\t}\n}\n`,
    `query grammar{\n\tgrammar(word: "καὶ") {\n\t\ttranslation\n\t\tword\n\t\trule\n\t\trootWord\n\t}\n}\n`,
    `query grammar{\n\tgrammar(word: "θεόν") {\n\t\ttranslation\n\t\tword\n\t\trule\n\t\trootWord\n\t}\n}\n`,
    `query grammar{\n\tgrammar(word: "θεὸς") {\n\t\ttranslation\n\t\tword\n\t\trule\n\t\trootWord\n\t}\n}\n`,
    `query grammar{\n\tgrammar(word: "λόγων") {\n\t\ttranslation\n\t\tword\n\t\trule\n\t\trootWord\n\t}\n}\n`,
    `query grammar{\n\tgrammar(word: "γὰρ") {\n\t\ttranslation\n\t\tword\n\t\trule\n\t\trootWord\n\t}\n}\n`,
]

const API_QUERIES = {
    herodotos: herodotosQueries,
    sokrates: sokratesQueries,
    dionysios: dionysiosQueries
};

function generateAlexandrosQuery() {
    const variables = alexandrosVariables[Math.floor(Math.random() * alexandrosVariables.length)];
    let query = alexandrosQueryTemplate;

    // Manually replace the placeholders in the query template
    query = query.replace('$word', JSON.stringify(variables.word));
    query = query.replace('$language', JSON.stringify(variables.language));
    query = query.replace('$mode', JSON.stringify(variables.mode));
    query = query.replace('$searchInText', JSON.stringify(variables.searchInText));

    return JSON.stringify({ query, variables });
}


function generateRandomApiQuery() {
    const apiKeys = Object.keys(API_QUERIES);
    const randomApi = apiKeys[Math.floor(Math.random() * apiKeys.length)];
    const queries = API_QUERIES[randomApi];
    const randomQuery = queries[Math.floor(Math.random() * queries.length)];
    return JSON.stringify({ query: randomQuery });
}

function randomApiQuery(apiName) {
    if (apiName) {
        if (apiName === 'alexandros') {
            return generateAlexandrosQuery();
        } else if (API_QUERIES[apiName]) {
            const queries = API_QUERIES[apiName];
            const randomQuery = queries[Math.floor(Math.random() * queries.length)];
            return JSON.stringify({ query: randomQuery });
        } else {
            return null;
        }
    } else {
        if (Math.random() < 0.75) {
            return generateAlexandrosQuery();
        } else {
            return generateRandomApiQuery();
        }
    }
}

function callHomeros(body) {
    const headers = {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
    };

    const res = http.post(`http://k3s-odysseia.greek/graphql`, body, { headers: headers });

    check(res, { 'status was 200': (r) => r.status == 200 });

    const randomSleep = Math.random() * 900 + 100; // Random sleep between 100ms and 1000ms
    sleep(randomSleep / 1000); // sleep function expects seconds, so divide by 1000
}

export function loadTestWithAllApis() {
    const requestBody = randomApiQuery();
    callHomeros(requestBody);
}

export function loadTest() {
    const apiName = __ENV.API_NAME;
    const requestBody = randomApiQuery(apiName);
    callHomeros(requestBody);
}