# protokaf

Kafka producer and consumer tool in protobuf format.

## Features
- Consume and produce messages using Protobuf protocol
- Trace messages with Jaeger
- Create custom templates for one or multiple messages and produce them to Kafka

## Install
```sh
go install github.com/SberMarket-Tech/protokaf@latest
```

## Configuration
Configuration file is optional, so you can skip this section.

In order for Protokaf to work, it needs to know how to reach your Kafka broker. First option is to provide `--broker` each time you invoke Protokaf. Another option is to use a configuration file. You can provide Protokaf with a configuration file with option `-F ..` on the command line. Or by default Protokaf will search its config files in `.protokaf.yaml` and `$HOME/.protokaf.yaml` respectively.

**Example of `.protokaf.yaml`**
```yaml
debug: true
broker: "<addr>:<port>"
kafka-auth-dsn: "SCRAM-SHA-512:<namespace>:<passwd>"
proto: "<dir>/<protofile>"
```

## Help
```sh
$ protokaf help
```

## List metadata
```sh
$ protokaf list [-t <topic>(,<topic>...)]
1 brokers:
 broker 1 "127.0.0.1:9093"
2 topics:
  topic "test-topic", partitions: 1
    partition 0, leader 1, replicas: [1] (offline: []), isrs: [1]
  topic "test", partitions: 1
    partition 0, leader 1, replicas: [1] (offline: []), isrs: [1]
```

## Produce
### Help
```sh
$ protokaf produce -h
```

### Examples
This proto file will be used in the examples below. 

`api/example.ptoto`
```protobuf
syntax = "proto3";

package example;

message HelloRequest {
  string name = 1;
  int32 age = 2;
}
```

**A simple produce message**
```sh
$ protokaf produce HelloRequest \
    --broker kafka:9092 \
    --proto internal/proto/testdata/example.proto \
    --topic test \
    --data '{"name": "Alice", "age": 11}'
```

**Produce message with headers**
```sh
$ protokaf produce HelloRequest \
    --broker kafka:9092 \
    --proto internal/proto/testdata/example.proto \
    --topic test \
    --header "priority=high" \
    --header "application=protokaf" \
    --data '{"name": "Alice", "age": 11}'
```

**Produce message with <a href="#template">template</a>**
```sh
$ protokaf produce HelloRequest \
    --broker kafka:9092 \
    --proto internal/proto/testdata/example.proto \
    --topic test \
    --data '{"name": {{randomFemaleName | quote}}, "age": {{randomNumber 10 20}}}' \
    --count 10 \
    --seed 42
```

**Produce message with Kafka auth**
```sh
$ protokaf produce HelloRequest \
    --broker kafka:9093 \
    --kafka-auth-dsn "SCRAM-SHA-512:login:passwd" \
    --proto internal/proto/testdata/example.proto \
    --topic test \
    --data '{"name": "Alice", "age": 11}'
```

**Read data from stdin or flag**

Read message `HelloRequest` from `stdin`, produce to `test` topic
```sh
$ echo '{"name": "Alice", "age": 11}' | protokaf produce HelloRequest -t test
```

Read message `HelloRequest` from `-d` value, produce to `test` topic
```sh
$ protokaf produce HelloRequest -t test -d '{"name": "Alice", "age": 11}'
```

### Template<a id="template"></a>
**Template options**
* `--seed <int>` You can set number greater then zero to produce the same pseudo-random sequence of messages
* `--count <int>` Useful for generating messages with random data
* `--concurrency <int>` Number of message senders to run concurrently for const concurrency producing

**Show all template functions**
```sh
$ protokaf produce --template-functions-print
```

## Consume
### Help
```sh
$ protokaf help consume
```

### Examples
```sh
$ protokaf consume HelloRequest \
    --broker kafka:9092 \
    --proto internal/proto/testdata/example.proto \
    --group mygroup \
    --topic test
```

**Read messages from Kafka `test` topic, use group `mygroup`, print to `stdout`**
```sh
$ protokaf consume HelloRequest -G mygroup -t test
```

**Read the last `10` messages from `test` topic, then exit**
```sh
$ protokaf consume HelloRequest -G mygroup -t test -c 10
```

**Read from offset `5` messages from `test` topic**
```sh
$ protokaf consume HelloRequest -G mygroup -t test -o 5
```

**Read from offset `5` messages from `test` topic and from offset `10` from `anothertest` topic**
```sh
$ protokaf consume HelloRequest -G mygroup -t test,anothertest -o test:5,anothertest:10
```

## Testing

### Prepare test environment
```sh
make docker-dev-up
make kafka-users
make install # optional (you can use 'go run . <args> <flags>')
```
