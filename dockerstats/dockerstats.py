import json,yaml,logging,thread
from docker import Client
from redis import StrictRedis

logging.basicConfig(level=logging.INFO)
log = logging.getLogger('dockerstats')

stat_types = [ 'all',
               'blkio_stats',
               'container',
               'network',
               'read',
               'host',
               'memory_stats',
               'cpu_stats' ]

class StatCollector(object):
    def __init__(self,host,redis):
        self.host    = host
        self.client  = Client(base_url=host)
        self.redis   = redis
        self.threads = []
        self.reload()

    def start(self,c):
        log.info('starting collecting of stats for container %s' % c)
        stats = self.client.stats(c)
        for stat in stats:
            s = json.loads(stat)
            s['container'] = c
            s['host'] = self.host
            self.redis.publish("stats",json.dumps(s))

    def reload(self):
        for t in self.threads:
            t.exit()
        self.threads = []
        for c in self._get_containers():
            t = thread.start_new_thread(self.start,(c,))
            self.threads.append(t)

    def _get_containers(self):
        return [ c['Names'][0].strip('/') for c in self.client.containers() ]

class DockerStats(object):
    def __init__(self,config_file='config.yaml'):
        self.config = self._load_config(config_file)

        r_host = self.config['redis'].split(':')[0]
        r_port = self.config['redis'].split(':')[1]
        redis = StrictRedis(host=r_host,port=r_port,db=0)
        self.sub = redis.pubsub()
        self.sub.subscribe("stats")

        self.collectors = []
        for host in self.config['hosts']:
            self.collectors.append(StatCollector(host,redis))
            log.info('started collector for host %s' % host)

    def get_stat(self,stat_type='all'):
        if stat_type not in stat_types:
            raise NameError('%s is not a recognized stat type' % stat_type)
        stat = self.sub.get_message()['data']
        if stat_type == 'all':
            return stat
        else:
            s = json.loads(stat)
            return json.dumps(s[stat_type])

    def get_basic_stat(self):
        stat = json.loads(self.sub.get_message()['data'])
        #container name
        name = stat['container'].split('/')[-1]
        result = { 'container': name }
        #cpu
        #convert to str and back to float
        print(stat['cpu_stats']['system_cpu_usage'])
        cpu = str(stat['cpu_stats']['system_cpu_usage']).split('e')[0]
        print cpu
        result['cpu'] = '%.2f' % round(float(cpu))
        print result['cpu']
        #memory
        mem = stat['memory_stats']['usage']
        result['memory_stats'] = self._format_bytes(mem)
        #network
        netin = stat['network']['rx_bytes']
        netout = stat['network']['tx_bytes']
        result['net_in'] = self._format_bytes(netin)
        result['net_out'] = self._format_bytes(netout)
        
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

    def _load_config(self,config_file):
        of = open(config_file, 'r')
        return yaml.load(of.read())
