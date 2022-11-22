const natsConfig = {
    servers: ['nats://54.160.229.90:80'],
    unsafe: true,
};

const publisher = new Nats(natsConfig);
const subscriber = new Nats(natsConfig);

// class Strategy {
//     constructor() {
//         this.name = null;     // resp.status
//         this.sma=ta.sma(close,14)
//     };
//     setName(name){
//         this.name=name
//     }
//     getName(){
//         return this.name
//     }
//     run(){
//
//
//     }
//
//
//
// }

// export function setup() {
// 	const res = http.get('https://httpbin.test.k6.io/get');
// 	return { data: res.json() };
// }
export default function (data) {


    let jma = ta.jma(close, 7, 50, 1)

    let dwma = ta.dwma(close, 10)

   while (true) {




       console.log("close", close.tail(3).reverse())

       console.log("jma", jma.tail(3).reverse())
       console.log("dwma", dwma.tail(3).reverse())
       console.log(strategy.long)
       // console.log(strategy)
       strategy.entry("buy","Long",{qty:0.4,limit:0.5,comment:"开多"})
       if (jma.crossOver(dwma)) {
           let message = {
               symbol: "ETHUSDT",
               side: "BUY",
               price: close.last(),
               quantity: 1
           }
           console.log("buy", message)

       }

       if (jma.crossUnder(dwma)) {
           let message = {
               symbol: "ETHUSDT",
               side: "SELL",
               price: close.last(),
               quantity: 1
           }
           console.log("sell", message)

       }

       console.log("=========\n")

         sleep(Math.random() * 30);

    }

}

