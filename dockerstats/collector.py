import json,logging,thread
from docker import Client

log = logging.getLogger('dockerstats')

class StatCollector(object):
    """
    Collects stats from all containers on a single Docker host, appending
    container name and id fields and publishing to redis
    params:
     - host(str): full base_url of a Docker host to connect to.
                  (e.g. 'tcp://127.0.0.1:4243')
     - redis(obj): redis client connection object
    """
    def __init__(self,host,redis):
        self.host    = host
        self.client  = Client(base_url=host)
        self.redis   = redis
        self.threads = []
        self.reload()

    def start(self,container_id,container_name):
        log.info('stat collector started for container %s' % container_id)
        stats = self.client.stats(container_id)
        for stat in stats:
            #append additional information to the returned stat
            s = json.loads(stat)
            s['container_name'] = container_id
            s['container_id'] = container_id
            s['host'] = self.host
            self.redis.publish("stats",json.dumps(s))

    def reload(self):
        for t in self.threads:
            t.exit()
        self.threads = []
        for cid,cname in self._get_containers().items():
            t = thread.start_new_thread(self.start,(cid,cname))
            self.threads.append(t)

    def _get_containers(self):
        containers = self.client.containers()
        return { c['Id'] : c['Names'][0].strip('/') for c in containers }
