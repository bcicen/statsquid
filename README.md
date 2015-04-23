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
docker run -ti --link statsquid_redis_1:redis -e STATSQUID_REDIS_HOST="redis" bcicen/statsquid top
```

# Components

## Agent

A single statsquid agent can be started on every Docker host you wish to collect stats from by running:
```bash
docker run -d -e STATSQUID_REDIS_HOST="redis.domain.com" -v /var/run/docker.sock:/var/run/docker.sock bcicen/statsquid agent
```

## Master

A statquid master connects to the common redis instance and listens for new stats, storing them for persistence
```bash
docker run -d -e STATSQUID_REDIS_HOST="redis.domain.com" bcicen/statsquid master
```

## Top

Statsquid comes with a curses-based top utility that can be used to view the aggregated stats in real time.
```bash
docker run -ti -e STATSQUID_REDIS_HOST="redis.domain.com" bcicen/statsquid top
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
```
Likewise, any of the below environmental variables will supersede its equivalent command line option:
```
STATSQUID_REDIS_HOST
STATSQUID_REDIS_PORT
STATSQUID_DOCKER_HOST
```

# Improvements

Statsquid is still an early stage project so there's quite a few things on wishlist:
- Adding stat shipping plugins to the master server(statsd,librato,etc.)
- Add master-agent communication to stop or start following stats for specific containers
- Improve roll-up of metrics for arbitrary time period averaging
- Add cumulative and interval delta views to top component 
