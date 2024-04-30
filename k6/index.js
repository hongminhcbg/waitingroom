import http from "k6/http";
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
import { check, sleep } from 'k6';


export const options = {
    scenarios: {
        contacts: {
            executor: 'ramping-vus',
            startVUs: 3,
            stages: [
                { duration: '20s', target: 10 },
                { duration: '60s', target: 100 },
                { duration: '180s', target: 2000 },
                { duration: '60s', target: 200 },
                { duration: '10s', target: 10 },
            ],
            gracefulRampDown: '1s',
        },
    },
};

export default function () {
    var userId = uuidv4()
    console.log("start respawn user", userId)

    var slotToken = ""
    while (true) {
        let data = {
            "user_id": userId,
        }

        let respCheck = http.post('http://localhost:8080/api/v1/slot-check',
            JSON.stringify(data),
            {
                headers: { 'Content-Type': 'application/json' }
            })
        check(respCheck, {
            'Internal server error': (r) => r.status !== 200,
            'Unexpected resp': (r) => r.json().message !== 'success',
            'request success': (r) => r.json().message==='success',
        })

        if (respCheck.status !== 200) {
            console.log("server error", respCheck.status)
            return
        }

        let respJson = respCheck.json()
        if (respJson.data.slot_token && respJson.data.slot_token.length > 0) {
            slotToken = respJson.data.slot_token
            console.log("get slot token success", slotToken)
            break
        }

        console.log("your rank is", respJson.data.rank)
        sleep(1)
    }

    sleep(5) // sleep 10 secs
    // release user
    let data = {
        "user_id": userId,
    }

    let respCheck = http.post('http://localhost:8080/api/v1/slot-release',
        JSON.stringify(data),
        {
            headers: { 'Content-Type': 'application/json' }
        })
    check(respCheck, {
        'Internal server error': (r) => r.status !== 200,
        'Unexpected resp': (r) => r.json().message !== 'success',
        'Request success': (r) => r.json().message==='success',
    })

    console.log("release user success")
}
