//
// class Funk {
//     constructor() {
//         this.name = null;     // resp.status
//
//     };
//     setName(name){
//         this.name=name
//     }
//     getName(){
//         return this.name
//     }
//
//
//
// }
//
// let describe = function (testname) {
//      return testname
// };


export function jma(_src, _length, _phase, _power) {

	let phaseRatio = _phase < -100 ? 0.5 : _phase > 100 ? 2.5 : _phase / 100 + 1.5
	let beta = 0.45 * (_length - 1) / (0.45 * (_length - 1) + 2)
	let alpha = math.pow(beta, _power)
	jma = 0.0
	let e0 = 0.0
	e0 = (1 - alpha) * _src + alpha * nz(e0[1])
	let e1 = 0.0
	e1 = (_src - e0) * (1 - beta) + beta * nz(e1[1])
	let e2 = 0.0
	e2 = (e0 + phaseRatio * e1 - nz(jma[1])) * math.pow(1 - alpha, 2) + math.pow(alpha, 2) * nz(e2[1])
	jma = e2 + nz(jma[1])

	return jma

}
