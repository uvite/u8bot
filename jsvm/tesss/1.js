

import ta from "k6/ta";
// export function setup() {
// 	const res = http.get('https://httpbin.test.k6.io/get');
// 	console.log(res)
// 	return { data: res.json() };
// }
//
// export function teardown(data) {
// 	console.log(JSON.stringify(data));
// }

export default function (data) {
	// someHelper();
	// const res = http.get('https://httpbin.test.k6.io/get');
	// console.log(res)
	// console.log(JSON.stringify(data));


	console.log("=========")
	let bb=ta.boll("SMA",close,20,2,2)

	console.log("=========")

	console.log("close",close.tail(3).reverse())


	console.log("u",bb.u.tail(3))
	console.log("m",bb.m.tail(3))
	console.log("l",bb.l.tail(3))


	console.log("=========\n")

}

