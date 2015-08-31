import sys
import json
import logging
import signal
import msgpack
from time import sleep
from docker import Client
from redis import StrictRedis
from multiprocessing import Process 

from statsquid.util import output

log = logging.getLogger('statsquid')

class Agent(object):
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
        self.docker     = Client(base_url=docker_host)
        self.source     = self.docker.info()['Name']
        self.ncpu       = self.docker.info()['NCPU']
        self.redis      = StrictRedis(host=redis_host,port=redis_port,db=0)
        self.children   = []
        self.stopped    = False

        log.info('Connected to Docker API at url %s' % docker_host)
        output('starting collector on source %s' % self.source)
        self.start()

    def start(self):
        signal.signal(signal.SIGINT, self._sig_handler)
        #start a collector for all existing containers
        for cid in [ c['Id'] for c in self.docker.containers() ]:
            self._add_collector(cid)

        #start event listener
        self._event_listener()

    def _sig_handler(self, signal, frame):
        self.stopped = True
        sys.exit(0)

    def _event_listener(self):
        """
        Listen for docker events and dynamically add or remove
        stat collectors based on start and die events
        """
        output('started event listener')
        for event in self.docker.events():
            event = json.loads(event.decode('utf-8'))
            if event['status'] == 'start':
                self._add_collector(event['id'])
            if event['status'] == 'die':
                self._remove_collector(event['id'])

    def _collector(self,cid,cname):
        """
        Collector instance collects stats via Docker API streaming web socket,
        appending container name and source, and publishing to redis
        params:
         - cid(str): ID of container to collect stats from
         - cname(str): Name of container
        """
        sleep(5) # sleep to allow container to fully start
        output('started collector for container %s' % cid)
        stats = self.docker.stats(cid, decode=True)
        for stat in stats:
            #append additional information to the returned stat
            stat['container_name'] = cname
            stat['container_id'] = cid
            stat['source'] = self.source
            stat['ncpu'] = self.ncpu
            self.redis.publish('statsquid', msgpack.packb(stat))
            if self.stopped:
                break
    
    #####
    # collector methods
    #####

    def _add_collector(self,cid):
        log.debug('creating collector for container %s' % cid)
        cname = self.docker.inspect_container(cid)['Name'].strip('/')

        p = Process(target=self._collector,name=cid,args=(cid,cname))
        p.start()

        self.children.append(p)

    def _remove_collector(self,cid):
        c = self._get_collector(cid)
        c.terminate()
        while c.is_alive():
            sleep(.2)
        output('collector stopped for container %s' % cid)
        self.children = [ c for c in self.children if c.name != cid ]

    def _get_collector(self,cid):
        return [ p for p in self.children if p.name == cid ][0]
