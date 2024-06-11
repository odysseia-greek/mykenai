# Dareios

Dareios is a powerful and versatile load testing repository named after the renowned Persian Great King. 
This repository is specifically designed to facilitate load testing for odysseia-greek.


## Baselines

The first set of baselines was created using the `homeros.js` script with [k6](https://k6.io).

There are 4 main stages of testing in terms of tracing which is what the performance is test is all about:

- 0%
- 10%
- 100%

Requests are in ms.

The tests were run against the homelab raspie cluster with the following nodes and elastic pods running:

```shell
NAME               STATUS   ROLES                       AGE   VERSION
k3s-s-athenai      Ready    control-plane,etcd,master   80d   v1.28.6+k3s2
k3s-s-sparta       Ready    control-plane,etcd,master   80d   v1.28.6+k3s2
k3s-s-syrakousai   Ready    control-plane,etcd,master   80d   v1.28.6+k3s2
k3s-w-argos        Ready    <none>                      80d   v1.28.6+k3s2
k3s-w-korinth      Ready    <none>                      80d   v1.28.6+k3s2
k3s-w-thebai       Ready    <none>                      80d   v1.28.6+k3s2

NAME               CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
k3s-s-athenai      332m         8%     5268Mi          65%
k3s-s-sparta       703m         17%    4638Mi          59%
k3s-s-syrakousai   438m         10%    3798Mi          48%
k3s-w-argos        222m         5%     1148Mi          30%
k3s-w-korinth      328m         8%     1115Mi          29%
k3s-w-thebai       302m         7%     2447Mi          64%
```

```shell
aristoteles-es-hot-0                 1/1     Running      0             18h
aristoteles-es-hot-1                 1/1     Running      0             17h
aristoteles-es-masters-0             1/1     Running      0             18h
aristoteles-es-warm-0                1/1     Running      0             18h
```

### Alexandros

Manual test from a local client:

searchInText: true
no tracing: 70-90ms
tracing: 2.5s

searchInText: false
no tracing: 40-60ms
tracing: 500ms

STREAM with only alexandros changed:

searchInText: false
no tracing: 40-60ms
tracing: 80-120ms

STREAM with changes:

searchInText: false
no tracing: 40-80ms
tracing: 50-80ms

#### 0% tracing 10vu

```json
        "http_req_duration": {
            "p(95)": 77.67189999999998,
            "avg": 35.98916733067732,
            "min": 14.629,
            "med": 28.756999999999998,
            "max": 292.949,
            "p(90)": 59.042
        },
```

#### 0% tracing 500vu scaling

```json
        "http_req_duration": {
            "p(95)": 8849.523599999993,
            "avg": 1990.5245805125999,
            "min": 14.962,
            "med": 974.635,
            "max": 10221.437,
            "p(90)": 5439.589600000003
        },
        "http_reqs": {
            "count": 37963,
            "rate": 124.78491341554738
        },
```

#### 10% tracing 500vu scaling

```json
        "http_req_duration": {
            "avg": 3521.7346648594594,
            "min": 14.264,
            "med": 1395.5715,
            "max": 46151.885,
            "p(90)": 10008.3913,
            "p(95)": 15772.655149999999
        },
        "http_reqs": {
            "count": 23978,
            "rate": 76.6654316470349
        },
```

#### 100% tracing 500vu scaling

```json
        "http_req_duration": {
            "avg": 27211.865072268403,
            "min": 16.03,
            "med": 26770.074,
            "max": 60001.266,
            "p(90)": 49787.1406,
            "p(95)": 59996.1909
        }
        "http_reqs": {
            "rate": 10.566600929600884,
            "count": 3487
        },
```


### Dionysios

#### 0% tracing 10vu

```json
        "http_req_duration": {
            "p(90)": 65.38800000000002,
            "p(95)": 202.17224999999996,
            "avg": 42.57051673228355,
            "min": 5.884,
            "med": 10.2635,
            "max": 1359.841
        }
```

#### 0% tracing 500vu scaling

```json
        "http_req_duration": {
            "avg": 1432.8984697981061,
            "min": 2.811,
            "med": 104.993,
            "max": 30083.16,
            "p(90)": 5243.9218,
            "p(95)": 10008.440400000001
        },
        "http_reqs": {
            "count": 48987,
            "rate": 157.72923479667844
        },
```

#### 100% tracing 500vu scaling

Cannot run test since homeros fails to create this many traces

### Herodotos

#### 0% tracing 10vu

```json
        "http_req_duration": {
            "med": 21.163,
            "max": 344.465,
            "p(90)": 26.8478,
            "p(95)": 30.014099999999996,
            "avg": 21.41193767976987,
            "min": 4.905
        }
```

#### 0% tracing 500vu scaling

```json
        "http_req_duration": {
            "p(95)": 1671.4953999999998,
            "avg": 502.16463322344873,
            "min": 3.036,
            "med": 321.426,
            "max": 8816.33,
            "p(90)": 1263.43
        },
        "http_reqs": {
            "count": 91309,
            "rate": 302.4111634294888
        },
```

#### 100% tracing 500vu scaling



### Sokrates

#### 0% tracing 10vu

```json
        "http_req_duration": {
            "max": 289.189,
            "p(90)": 59.73440000000001,
            "p(95)": 70.27624999999992,
            "avg": 30.439050766283525,
            "min": 12.861,
            "med": 22.239
        },
```

#### 0% tracing 500vu scaling

```json
        "http_req_duration": {
            "p(95)": 3850.974799999999,
            "avg": 1409.253346442747,
            "min": 3.635,
            "med": 1030.979,
            "max": 30049.829,
            "p(90)": 2975.2360000000003
        },
        "http_reqs": {
            "count": 49125,
            "rate": 162.14230232266496
        },
```

#### 100% tracing 500vu scaling


