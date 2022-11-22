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


       let hma=ta.hma(close,14)



       console.log("close",close.tail(3).reverse())
 
       console.log("hma",hma.tail(3).reverse())

         sleep(Math.random() * 10);

    }

}

