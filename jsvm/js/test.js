import http from 'k6/http';

import {sleep} from 'k6'
import ta from 'k6/x/ta'

exports.options = { setupTimeout: "10s", teardownTimeout: "10s" };

export function setup() {

	console.log("setup234")
	 const res = http.get('https://httpbin.test.k6.io/get');
	//
	// console.log(res.json())
	return { data: res.json()};
}

// export function teardown(data) {
// 	console.log(JSON.stringify(data));
// 	console.log("tear");
//
// }

export default function ( data) {
	console.log(data)


    while (true){

		let sma=ta.sma(close,14)
	console.log("funck",close.last());
	console.log("sma:",sma.last());
	console.log(strategy.getSMA().last())

	sleep(10)
	}
}
