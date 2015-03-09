import json,yaml,logging,thread
from datetime import datetime
from docker import Client
from redis import StrictRedis

logging.basicConfig(level=logging.INFO)
log = logging.getLogger('dockerstats')

cpu_tick = 100

class Container(object):
    """
    Container object holds a collection of stats for a specific container_id,
    rolling up the data at regular intervals.
    params:
     - container_id(str): Docker container id
    methods:
     - append_stat: Appends a new stat, recalculating averages
    """
    def __init__(self,container_id):
        self.id = container_id
        self.name = ""
        self.stats_read = 0

    def append_stat(self,stat):
        self.stats_read += 1
        if not self.name:
            self.name = stat.container_name
        #cpu
        #convert to str and back to float
        cpu = str(stat['cpu_stats']['system_cpu_usage']).split('e')[0]
        result['cpu'] = '%.2f' % round(float(cpu))
        #memory
        self.memory_usage = self._format_byte(stat['memory_stats']['usage'])
        #network
        netin = stat['network']['rx_bytes']
        netout = stat['network']['tx_bytes']
        self.net_in = self._format_bytes(netin)
        self.net_out = self._format_bytes(netout)
        
        return result

    def _format_bytes(self,b):
        if b < 1000:
            return '%i' % b + 'b'
        elif 1000 <= b < 1000000:
            return '%.1f' % float(b/1000) + 'kb'
        elif 1000000 <= b < 1000000000:
            return '%.1f' % float(b/1000000) + 'mb'
        elif 1000000000 <= b < 1000000000000:
            return '%.1f' % float(b/1000000000) + 'gb'
        elif 1000000000000 <= b:
            return '%.1f' % float(b/1000000000000) + 'tb'

class Stat(object):
    """
    Stat object, created from json received from stat collector
    """
    def __init__(self,statjson):
        self.raw = statjson
        self.statdict = json.loads(self.raw)
        self.timezone,self.timestamp = self._readtime(self.statdict['read'])
        self.container_name = self.statdict['container_name'].split('/')[-1]
        self.container_id = self.statdict['container_id'].split('/')[-1]


    def _readtime(self,timestamp):
        d,t = timestamp.split('T')
        year,month,day = d.split('-')
        if '-' in t:
            t,tz = t.split('-')
            tz = '+' + tz
        if '+' in t:
            t,tz = t.split('-')
            tz = '-' + tz
        hour,minute,second = t.split(':')
        second,microsecond = second.split('.')

        timestamp = datetime(int(year),
                             int(month),
                             int(day),
                             int(hour),
                             int(minute),
                             int(second),
                             int(microsecond[0:6]))
        return (timestamp,tz)


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

class DockerStats(object):
    def __init__(self,config_file='config.yaml'):
        self.config = self._load_config(config_file)
        
        r_host = self.config['redis'].split(':')[0]
        r_port = self.config['redis'].split(':')[1]
        redis = StrictRedis(host=r_host,port=r_port,db=0)
        self.sub = redis.pubsub()
        self.sub.subscribe("stats")
        self.containers = {}

    def run_forever(self):
        while True:
            stat = Stat(self.sub.get_message()['data'])
            if stat.container_name not in self.containers:
                self.containers['stat.container_name'] = Container(stat.container_name)

    def start_collectors(self):
        self.collectors = []
        for host in self.config['hosts']:
            self.collectors.append(StatCollector(host,redis))
            log.info('started collector for host %s' % host)

    def reload_collectors(self):
        for collector in self.collectors:
            log.debug('performing reload for %s collector' % collector.host)
            collector.reload()

    def _load_config(self,config_file):
        of = open(config_file, 'r')
        return yaml.load(of.read())
