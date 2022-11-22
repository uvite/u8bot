import ta from "k6/ta";

import {Nats} from 'k6/x/nats';

const natsConfig = {
	servers: ['nats://54.160.229.90:80'],
	unsafe: true,
};

const publisher = new Nats(natsConfig);
const subscriber = new Nats(natsConfig);

export default function (data) {



		let jma = ta.jma(close, 7, 50, 1)

		let dwma = ta.dwma(close, 10)

		console.log("close", close.tail(3).reverse())
		//-----
		// let message= {
		// 	symbol:"ETHUSDT",
		// 	side:"BUY",
		// 	price:close.last(),
		// 	quantity:1
		// }
		// console.log("buy",message)
		// publisher.publish('Order.Close.Long',JSON.stringify( message));
		//------
		console.log("jma", jma.tail(3).reverse())
		console.log("dwma", dwma.tail(3).reverse())
		if (jma.crossOver(dwma)) {
			let message = {
				symbol: "ETHUSDT",
				side: "BUY",
				price: close.last(),
				quantity: 1
			}
			console.log("buy", message)
			publisher.publish('Order.Close.Short', JSON.stringify(message));

			publisher.publish('Order.Open.Long', JSON.stringify(message));
		}

		if (jma.crossUnder(dwma)) {
			let message = {
				symbol: "ETHUSDT",
				side: "SELL",
				price: close.last(),
				quantity: 1
			}
			console.log("sell", message)
			publisher.publish('Order.Close.Long', JSON.stringify(message));
			publisher.publish('Order.Open.Short', JSON.stringify(message));
		}

		console.log("=========\n")

}

