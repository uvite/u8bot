import {Nats} from 'k6/x/nats';
import ta from 'k6/x/ta';
import {sleep} from 'k6'

const natsConfig = {
	servers: ['nats://54.160.229.90:80'],
	unsafe: true,
};

const publisher = new Nats(natsConfig);
const subscriber = new Nats(natsConfig);

class Funk {
    constructor() {
        this.name = null;     // resp.status

    };
    setName(name){
        this.name=name
    }
    getName(){
        return this.name
    }
	run(){
		let message = {
			symbol: "ETHUSDT",
			side: "Short",
			price: 1,
			quantity: 1
		}

		console.log("close",close.tail(3).reverse())

		//
		publisher.publish('Genv.Close.Short', JSON.stringify(message));
		console.log("buy", message)
		console.log("===========\n")
	}



}

// export function setup() {
// 	const res = http.get('https://httpbin.test.k6.io/get');
// 	return { data: res.json() };
// }
export default function (data) {


	// console.log(JSON.stringify(data));
	// console.log("funck");
	subscriber.subscribe('Genv.*.*', (msg) => {
		console.log("order", msg.topic, msg.data)
	});

	let funk=new Funk()
	while (true) {

		let sma=ta.sma(close,14)



		console.log("sma",sma.tail(3).reverse() )
		funk.run()

		//publisher.publish('Genv.Open.Long', JSON.stringify(message));
		sleep(Math.random() * 10);

	}

}

