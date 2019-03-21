[![Build Status](https://travis-ci.org/dtgorski/eventsourced.svg?branch=master)](https://travis-ci.org/dtgorski/eventsourced)
[![Coverage Status](https://coveralls.io/repos/github/dtgorski/eventsourced/badge.svg?branch=master)](https://coveralls.io/github/dtgorski/eventsourced?branch=master)

# eventsourced
Asynchronous server message push (received via ServerSentEvents, SSE) using RabbitMQ and discrete client queues.

```
 ┌───────────┐              ┌─     optional    ─┐                      ┌─────────────┐
 │  client   ├── queue A ─> │   ┌───────────┐   │   ╔══════════════╗   │ AMQP broker │
 │ [browser] │<─── SSE ──── │ <─┤           ├─> │ <─╢              ║<─>│ [RabbitMQ]  │
 └───────────┘              │   │ rev proxy │   │   ║ eventsourced ║   └──────┬──────┘
 ┌───────────┐              │   │  [nginx]  │   │   ║    [this]    ║   ┌──────┴──────┐
 │  client   │<─── SSE ──── │ <─┤           ├─> │ <─╢              ║<─>│ AMQP broker │
 │ [browser] ├── queue B ─> │   └───────────┘   │   ╚══════════════╝   │ [RabbitMQ]  │
 └───────────┘              └─                 ─┘           .          └──────┬──────┘
       .                                                    .                 .
       .                                                    .                 .
       .                                                    .                 .
```

## Synopsis
This server will consume text messages from a broker via AMQP (tested with RabbitMQ) and deliver them to a requesting HTTP client.

## Installation
Pull (or [download](https://github.com/dtgorski/eventsourced/releases)) the `stable` branch to get the most recent stable version. Do not use the `master` branch in production as it may contain junk.

```bash
# pull the repo
git clone https://github.com/dtgorski/eventsourced.git

# inherently use the 'stable' branch!
cd eventsourced && git checkout stable

# create custom configuration file
cp config.yml-dist config.yml
```

Review and adjust the settings in the newly created `config.yml` configuration file and continue with ...

```
make docker
docker run -d -p 2069:2069 eventsourced
```

The final docker image is based on an [Alpine Linux](https://alpinelinux.org/) and will be < 10 MB in size.

## Configuration
The configuration file is `config.yml`. The `eventsourced` server will look up in following locations in the given order:

 * `/etc/eventsourced/config.yml`
 * `<PATH_OF_EVENTSOURCED_BINARY>/config.yml`

Entries from subsequent configuration sources will overwrite existing values. When a configuration file is not provided at all, `eventsourced` will fall back to its defaults - and this is not what you want. Use `./eventsourced -c` to dump the current set up, which may look like follows:

### `server`
```yaml
server:
  address: 0.0.0.0:2069
```
The `server.address` entry denotes the TCP address of the listening `eventsourced` server. As the server must not run as root, the listening port should be >= 1024.

### `broker`
```yaml
broker:
  node:
    - amqp://guest:guest@10.0.0.11:5672/ # channelMax 2047
    - amqp://guest:guest@10.0.0.11:5672/ # channelMax 2047
    - amqp://guest:guest@10.0.0.55:5672/ # channelMax 2047
```

List of URLs that address the message broker nodes.

It is legal to repeat the same broker URL multiple times as each node connection can not exceed 2047 distinct communication channels (or clients).

### `queue`
```yaml
queue:
  pattern: ${query:id}
  expire:  1800
```
These are the queue name pattern and the expiration of the queue in seconds. The queue name is generated from the clients request to the `eventsourced` endpoint as denoted by the `queue.pattern`:

 * `${query:id}` - will extract the `id` from the query part of the request.
 * `${cookie:sid}` - will extract the `sid` (here session ID) from the request Cookie header.
 * `queue-${query:id}-${cookie:sid}-name` - will do all above and concatenate the result.

When a queue with the requested name does not yet exist in the broker queue pool, it will be created. When a queue exceed its `queue.expire` limit without having a consumer connected, it will be dropped from the broker queue pool.

Simultaneous access to a particular queue will lead to a HTTP status 503 (Server Unavailable) response for all clients except for the first one.

### `header`
```yaml
header:
  cors:
    Access-Control-Allow-Origin:  '*'
    Access-Control-Allow-Methods: GET, OPTIONS
    Access-Control-Allow-Headers: Content-Type, Accept, Cache-Control
  sse:
    Content-Type:      text/event-stream; charset=utf-8
    Cache-Control:     no-cache
    Transfer-Encoding: identity
    X-Accel-Buffering: no
```
HTTP response headers for the [CORS](https://en.wikipedia.org/wiki/Cross-origin_resource_sharing) mechanism and [SSE](https://en.wikipedia.org/wiki/Server-sent_events) requests.

## Runtime metrics
A running `eventsourced` server exposes the [expvar](https://golang.org/pkg/expvar/) runtime monitoring information under `/debug/vars`.

## License
[MIT](https://opensource.org/licenses/MIT) - © dtg [at] lengo [dot] org
