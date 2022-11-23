# xk6-nats

This is a [k6](https://go.k6.io/k6) extension using the [xk6](https://github.com/k6io/xk6) system, that allows to use NATS protocol.

| :exclamation: This is a proof of concept, isn't supported by the k6 team, and may break in the future. USE AT YOUR OWN RISK! |
|------|

## Build

To build a `k6` binary with this extension, first ensure you have the prerequisites:

- [Go toolchain](https://go101.org/article/go-toolchain.html)
- Git

1. Install `xk6` framework for extending `k6`:
```shell
go install go.k6.io/xk6/cmd/xk6@latest
```

2. Build the binary:
```shell
xk6 build --with github.com/ydarias/xk6-nats@latest
```

3. Run a test
```shell
./k6 run folder/test.js
```

## API

### Nats

A Nats instance represents the connection with the NATS server, and it is created with `new Nats(configuration)`, where configuration attributes are:

| Attribute | Description |
| --- | --- |
| servers | (mandatory) is the list of servers where NATS is available (e.g. `[nats://localhost:4222]`) |
| unsafe | (optional) allows running with self-signed certificates when doing tests against a testing environment, it is a boolean value (default value is `false`) |
| token | (optional) is the value of the token used to connect to the NATS server |

#### Available functions

| Function | Description |
| --- | --- |
| publish(topic, message) | publish a new message using the topic (string) and the given payload that is a string representation that later is serialized as a byte array |
| subscribe(topic, handler) | subscribes to the publication of a message using the topic (string) and a handler that is a function like `(msg) => void` |
| request | sends a request to the topic (string) and the given payload as string representation, and returns a message |

#### Return values

A message return value has the following attributes:

| Attribute | Description | 
| --- | --- |
| data | the payload in string format |
| topic | the topic where the message was published |

**Some examples at the section below**.

## Testing

NATS supports the classical pub/sub pattern, but also it implements a request-reply pattern, this extension provides support for both.

![xk6-nats operations diagram](assets/xk6-nats-operations.png)

### Pub/sub test

```javascript
import {check, sleep} from 'k6';
import {Nats} from 'k6/x/nats';

const natsConfig = {
    servers: ['nats://localhost:4222'],
    unsafe: true,
};

const publisher = new Nats(natsConfig);
const subscriber = new Nats(natsConfig);

export default function () {
    subscriber.subscribe('topic', (msg) => {
        check(msg, {
            'Is expected message': (m) => m.data === 'the message',
            'Is expected topic': (m) => m.topic === 'topic',
        })
    });

    sleep(1)

    publisher.publish('topic', 'the message');

    sleep(1)
}

export function teardown() {
    publisher.close();
    subscriber.close();
}
```

Because K6 doesn't provide an event loop we need to use the `sleep` function to wait for async operations to complete.

### Request-reply test

```javascript
import { Nats } from 'k6/x/nats';
import { check, sleep } from 'k6';

const natsClient = new Nats({
  servers: ['nats://localhost:4222'],
});

export default function () {
    const payload = {
        foo: 'bar',
    };

    const res = natsClient.request('my.subject', JSON.stringify(payload));

    check(res, {
        'payload pushed': (r) => r.status === 'success',
    });
}

export function teardown() {
    natsClient.close();
}
```
