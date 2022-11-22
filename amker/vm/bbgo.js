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

    while (true) {
    let jma = ta.jma(close, 7, 50, 1)

    let dwma = ta.dwma(close, 10)


    console.log("close", close.tail(3).reverse())

    console.log("jma", jma.tail(3).reverse())
    console.log("dwma", dwma.tail(3).reverse())

    // if (jma.crossOver(dwma)) {
    //     console.log("buy")
    //     strategy.close("sell", "Short")
    //     strategy.entry("buy", "Long", {qty: 0.1,   comment: "开多"})
    // }
    //
    // if (jma.crossUnder(dwma)) {
    //     console.log("short")
    //     strategy.close("buy", "Long")
    //     strategy.entry("sell", "Short", {qty: 0.1,   comment: "开空"})
    //
    //
    // }
    //

    if (postion.isLong()==true){
        strategy.closeFuck("sell")


    }

   strategy.entry("buy", "Short", {qty: 0.1,   comment: "开空"})
    if (postion.isShort()==true){
        strategy.closeFuck("sell")

        strategy.entry("buy", "Long", {qty: 0.1,   comment: "开多"})

    }

    // strategy.close("sell", "Short")
    // strategy.entry("buy", "Long", {qty: 0.1,   comment: "开多"})
    console.log(postion.getBase(),postion.isLong())
    console.log(postion.isShort())
    console.log("=========\n")
        sleep(Math.random() * 30);

    }

}

