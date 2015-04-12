import json,logging,threading
from docker import Client
from redis import StrictRedis
from time import sleep

logging.basicConfig(level=logging.DEBUG)
log = logging.getLogger('dockerstats')

class StatCollector(object):
    """
    Collects stats from all containers on a single Docker host, appending
    container name and id fields and publishing to redis
    params:
     - docker_host(str): full base_url of a Docker host to connect to.
                  (e.g. 'tcp://127.0.0.1:4243')
     - redis_host(str): redis host to connect to. default 127.0.0.1
     - redis_port(int): port to connect to redis host on. default 6379
    """
    def __init__(self,docker_host,redis_host='127.0.0.1',redis_port=6379):
        self.docker  = Client(base_url=docker_host)
        self.source  = self.docker.info()['Name']
        self.redis   = StrictRedis(host=redis_host,port=redis_port,db=0)
        self.stopped = False
        self.threads = []
        log.info('starting collector on source %s' % self.source)
        self.reload()

    def _collector(self,cid,cname):
        """
        Collector instance collects stats via Docker API streaming web socket,
        appending container name and source, and publishing to redis
        params:
         - cid(str): ID of container to collect stats from
         - cname(str): Name of container
        """
        log.info('starting collector for container %s' % cid)
        stats = self.docker.stats(cid)
        for stat in stats:
            #append additional information to the returned stat
            s = json.loads(stat)
            s['container_name'] = cname
            s['container_id'] = cid
            s['source'] = self.source
            self.redis.publish("stats",json.dumps(s))
            if self.stopped:
                log.info('collector stopped for container %s' % cid)
                return

    #TODO: dynamically add and return containers based on docker events
    def reload(self):
        self.stop()

        for cid,cname in self._get_containers().items():
            log.debug('creating collector for container %s' % cid)
            self.threads.append(
                    threading.Thread(
                        target=self._collector,
                        name=cid,
                        args=(cid,cname)
                        )
                    )

        [ t.start() for t in self.threads ] 

    def stop(self):
        self.stopped = True
        for idx,t in enumerate(self.threads):
            while t.is_alive():
                sleep(.2)
        self.threads = []
        self.stopped = False

    def _get_containers(self):
        containers = self.docker.containers()
        return { c['Id'] : c['Names'][0].strip('/') for c in containers }
