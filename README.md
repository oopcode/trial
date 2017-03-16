# Trial application

## 1. General description

`Trial` is a sample project which contains an implementation of a simple microservice. The microservice reads JSON messages fron [NSQ](http://nsq.io/), tries to parse them into `common.AppMessage` structs and then sends them to [Aerospike](http://aerospike.com/).

`AppMessage` is a simple type with just two fields:

```
type AppMsg struct {
	ID        int64  `json:"id"`
	Timestamp string `json:"timestamp"`
}
```

Aerospike keys are `AppMsg.ID`s, rows have just one column: `"timestamp"`.

Another twist is that NSQ consumer is modified to guarantee that at most `nsq_consumer_max_read` messaged are received each `nsq_consumer_delta` seconds; this is not achievable via standard NSQ client config and demonstrates how we can modify standard behavior in a somewhat tricky way (see [source code](https://github.com/oopcode/trial/blob/master/microservice/app.go)).

## 2. Configuration

The application expects to find a valid JSON configuration file located at `/opt/trial/config.json`. Default configuration looks as follows:

```
{
    "nsq_topic_name": "trial_topic",
    "nsq_host_port": "127.0.0.1:4150",
    "nsq_consumer_delta": 5,
    "nsq_consumer_max_read": 10,
    "as_host": "127.0.0.1",
    "as_port": 3000,
    "as_namespace": "test",
    "as_set": "trial_set"
}
```

* `nsq_topic_name` is the topic name for NSQ. If no such topic exists, it will be created;
* `nsq_host_port` specifies NSQ listen address;
* `nsq_consumer_delta` specifies time (in seconds) between reads from NSQ;
* `nsq_consumer_max_read` specifies maximum number of messages that can be retrieved from NSQ by a single read;
* `as_host` specifies Aerospike listen host;
* `as_port` specifies Aerospike listen port;
* `as_namespace` is the namespace to be used with Aerospike; is no such namespace exists, writes will **fail with an error**;
* `as_set`: is the set to be used with Aerospike; is no such set exists, it will be created.

## 3. Running the code

Running the code is simple (sudo is required to write PID):

```
reset && go build -o trial && sudo ./trial

```

Logs are stored at `/var/log/trial.log`.

## 4. Daemon

Move `./trial.init` from project root to `/etc/init.d/` and it will be available as a service. You will also have to move the binary to `/usr/local/bin/`:

```
go build -o trial && sudo mv trial /usr/local/bin/
```

Specific handlers are set for `SIGINT, SIGTERM`.

## 5. Additional

The application starts with a simple producer sending messages to NSQ. You can easily disable producer in `main.go`: just comment it out.