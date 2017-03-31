# Proxy

Proxy is a POC asynchronous proxy to legacy systems.

## Installation

```bash
$ github.com/mmatczuk/proxy/cmd/proxy
```

## Running

```bash
$ proxy 127.0.0.1:10001 127.0.0.1:10002 127.0.0.1:10003
```

`proxy` would start on `:80`, if you want to specify other address use `-http` flag.

## API by example

### Create new task

```bash
$ curl -XPOST -d'{
  "client_id": "f0a4fd40-44bf-4535-b807-632586645d6f",
  "info": "test",
  "mode": "sequential",
  "failonerror": true
}' localhost:8080/v1/task
"d74b0690-1619-11e7-8191-704d7b4a5d2f"
```

### Check task status

```bash
$ curl localhost:8080/v1/task/d74b0690-1619-11e7-8191-704d7b4a5d2f/status
[{"addr":"ala","status":"failure","message":"failed to send request: Post http://host: dial tcp: lookup host on 127.0.1.1:53: no such host"}]
```

### Kill task

```bash
$ curl localhost:8080/v1/task/d74b0690-1619-11e7-8191-704d7b4a5d2f/kill
       []

```
