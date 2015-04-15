# StatSquid

<p align="center">
  <img src="https://raw.githubusercontent.com/bcicen/statsquid/master/statsquid.png" alt="Statsquid"/>
</p>

Statsquid aggregates containers stats from multiple docker hosts

# Quickstart

To get started running statsquid on a single host, use the included docker-compose file:

```bash
git clone https://github.com/bcicen/statsquid.git
cd statsquid/
docker-compose up
```

To view the stats being collected:
```bash
docker run -ti --link statsquid_redis_1:redis -e STATSQUID_REDIS_HOST="redis" bcicen/statsquid --command top
```

# Components

## Agent

A single statsquid agent can be started on every Docker host you wish to collect stats from by running:
```bash
docker run -d -e STATSQUID_COMMAND="agent" -e STATSQUID_REDIS_HOST="redis.domain.com" -v /var/run/docker.sock:/var/run/docker.sock bcicen/statsquid
```

## Master

A statquid master connects to the common redis instance and listens for new stats, storing them for persistence
```bash
docker run -d -e STATSQUID_COMMAND="master" -e STATSQUID_REDIS_HOST="redis.domain.com" bcicen/statsquid
```

## Top

Statsquid comes with a curses-based top utility that can be used to view the aggregated stats in real time.
```bash
docker run -ti -e STATSQUID_COMMAND="top" -e STATSQUID_REDIS_HOST="redis.domain.com" bcicen/statsquid
```

# Options

Statsquid supports the following options:
```
  --docker-host DOCKER_HOST
                        docker host to connect to (default:
                        tcp://127.0.0.1:4243)
  --redis-host REDIS_HOST
                        redis host to connect to (default: 127.0.0.1)
  --redis-port REDIS_PORT
                        redis port to connect on (default: 6379)
  --command COMMAND     statsquid mode (agent,master,top)
```
Likewise, any of the below environmental variables will supersede its equivalent command line option:
```
STATSQUID_COMMAND
STATSQUID_REDIS_HOST
STATSQUID_REDIS_PORT
STATSQUID_DOCKER_HOST
```

# Improvements
