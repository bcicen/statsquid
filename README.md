# StatSquid

<p align="center">
  <img src="https://raw.githubusercontent.com/bcicen/statsquid/master/statsquid.png" alt="Statsquid"/>
</p>

Statsquid aggregates containers stats from multiple docker hosts, providing real-time monitoring and (planned) flushing of stats to a pluggable backend.

# Quickstart

To get started running statsquid on a single host, use the included docker-compose file:

```bash
git clone https://github.com/bcicen/statsquid.git
cd statsquid/
docker-compose up
```

And run the statsquid top utility to view the streamed stats in real-time:
```bash
docker run -ti --link statsquid_redis_1:redis bcicen/statsquid top --redis-host redis
```

![top][top]

# Components

![diagram][diagram]

## Agent

A single statsquid agent can be started on every Docker host you wish to collect stats from by running:
```bash
docker run -td -v /var/run/docker.sock:/var/run/docker.sock bcicen/statsquid agent --redis-host redis.mydomain.com
```

## Master

A statquid master connects to the common redis instance and listens for new stats, storing them for persistence
```bash
docker run -td bcicen/statsquid master --redis-host redis.mydomain.com
```

## Top

Statsquid comes with a curses-based top utility that can be used to view the aggregated stats in real time.
```bash
docker run -ti bcicen/statsquid top --redis-host redis.mydomain.com
```

Statsquid top supports filtering by host and container name/id, sorting by field, and cumulative vs incremental views. Hit 'h' for a full list of features.

# Options

Statsquid supports the following options:
```
  --docker-host DOCKER_HOST
                        docker host to connect to (default:
                        /var/run/docker.sock) (agent only)
  --redis-host REDIS_HOST
                        redis host to connect to (default: 127.0.0.1)
  --redis-port REDIS_PORT
                        redis port to connect on (default: 6379)
```
Likewise, any of the below environmental variables will supersede its equivalent command line option:
```
DOCKER_HOST
STATSQUID_REDIS_HOST
STATSQUID_REDIS_PORT
```

# Improvements

Statsquid is still an early stage project so there's quite a few things that can be improved:
- Add master-agent communication to stop or start following stats for specific containers
- Improve roll-up of metrics for arbitrary time period averaging during flushing

[diagram]: http://i.imgur.com/3scUxxZl.png
[top]: http://i.imgur.com/5OfeDhV.gif
