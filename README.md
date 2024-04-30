# waitingroom
I. Waiting room solution

  ![Alt text](./img/s.png?raw=true "Summary")

II. Loadtest result

```
    loadtest conntext:
      redis: 4 cores
      RAM: 7.76Gi
      server local machine: Apple M2, 16Gi 
      max CCU: 2k
      http_reqs: 613 req/s
      tool: k6

     ✗ Internal server error
      ↳  0% — ✓ 0 / ✗ 220566
     ✗ Unexpected resp
      ↳  0% — ✓ 0 / ✗ 220566
     ✓ Request success

     checks.........................: 33.33% ✓ 220566     ✗ 441132
     data_received..................: 39 MB  108 kB/s
     data_sent......................: 44 MB  123 kB/s
     http_req_blocked...............: avg=6.94µs  min=0s    med=2µs    max=2.4ms   p(90)=6µs    p(95)=8µs
     http_req_connecting............: avg=3.52µs  min=0s    med=0s     max=2.28ms  p(90)=0s     p(95)=0s
     http_req_duration..............: avg=2.09ms  min=283µs med=1.46ms max=84.88ms p(90)=4.01ms p(95)=5.43ms
       { expected_response:true }...: avg=2.09ms  min=283µs med=1.46ms max=84.88ms p(90)=4.01ms p(95)=5.43ms
     http_req_failed................: 0.00%  ✓ 0          ✗ 220566
     http_req_receiving.............: avg=19.85µs min=3µs   med=12µs   max=1.71ms  p(90)=41µs   p(95)=56µs
     http_req_sending...............: avg=12.47µs min=1µs   med=7µs    max=1.25ms  p(90)=26µs   p(95)=36µs
     http_req_tls_handshaking.......: avg=0s      min=0s    med=0s     max=0s      p(90)=0s     p(95)=0s
     http_req_waiting...............: avg=2.06ms  min=273µs med=1.43ms max=84.87ms p(90)=3.95ms p(95)=5.36ms
     http_reqs......................: 220566 612.669053/s
     iteration_duration.............: avg=15.93s  min=6s    med=15.02s max=55.26s  p(90)=25.03s p(95)=25.04s
     iterations.....................: 13671  37.974115/s
     vus............................: 10     min=3        max=2000
     vus_max........................: 2000   min=2000     max=2000