
import crypto from 'k6/crypto';

import {sleep} from 'k6';
import {Nats} from 'k6/x/nats';

const natsConfig = {
    servers: ['nats://54.160.229.90:80'],
    unsafe: true,
};

const publisher = new Nats(natsConfig);
const subscriber = new Nats(natsConfig);

export default function () {
    subscriber.subscribe('Message.Debug', (msg) => {
        console.log(msg.topic)
        console.log("--------------")
        console.log(msg.data)

    });

    console.log(crypto.sha256('hello world!', 'hex'));
    const hasher = crypto.createHash('sha256');
    hasher.update('hello ');
    hasher.update('world!');
    console.log(hasher.digest('hex'));
}
export function init(){

    console.log("init 4444")
}

export function onbar(kline){
    console.log(kline)
    publisher.publish('topic', kline.open);
}

