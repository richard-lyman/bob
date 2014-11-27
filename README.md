#### Acronym

Binary ~~Large~~ OBject store

#### Overview

A micro REST interface to Redis (that can't get smaller).

POST HTTP requests are a call to SET in Redis.

GET HTTP requests are a call to GET in Redis.

The Redis key used in calls to GET and SET is only the first URL Path Segment (see RFC 3986).

#### Building
 1. Install golang
 2. go get ./...
 3. go build

#### Options

Flag | Type | Default | Option | Explanation
---- | ---- | ------- | ------ | ----
lockVersions | bool | false | true | Setting lockVersions to true will use a Redis SET call with the NX flag. The default is a regular SET call in Redis.
hostPort | string | ":8080" | Any valid host:port | This is the host:port that bob will listen on.
redisHostPort | string | "127.0.0.1:6379" | The host:port for Redis |(Self explanatory)

#### Examples

 1. Run bob on port 8081 using lockVersions
```bash
./bob -lockVersions=true -hostPort ":8081"
```
 1. Store content
```bash
curl -v -X POST --data-binary "@locationOfSomeFile" :::8080/someKey
```
 1. Get content
```bash
curl -v :::8080/someKey
```
 1. Fail to modify content
```bash
curl -v -X POST --data-binary "@locationOfSomeFile" :::8080/someKey
```
