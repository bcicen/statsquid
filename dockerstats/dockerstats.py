import json,yaml,logging
from datetime import datetime
from redis import StrictRedis
from collector import StatCollector
from container  import Container

log = logging.getLogger('dockerstats')

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

        ts = datetime(int(year),
                      int(month),
                      int(day),
                      int(hour),
                      int(minute),
                      int(second),
                      int(microsecond[0:6]))
        return (ts,tz)

class DockerStats(object):
    def __init__(self,config_file='config.yaml'):
        self.config = self._load_config(config_file)
        
        r_host = self.config['redis'].split(':')[0]
        r_port = self.config['redis'].split(':')[1]
        redis = StrictRedis(host=r_host,port=r_port,db=0)
        self.sub = redis.pubsub(ignore_subscribe_messages=True)
        self.sub.subscribe("stats")
        self.containers = {}
        self.start_collectors(redis)
        self.run_forever()

    def run_forever(self):
        while True:
            msg = self.sub.get_message()
            if msg:
                self._process_msg(msg['data'])

    def _process_msg(self,msg):
        stat = Stat(msg)
        cid = stat.container_id
        if cid not in self.containers:
            self.containers[cid] = Container(cid)
        self.containers[cid].append_stat(stat)

    def start_collectors(self,redis):
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
