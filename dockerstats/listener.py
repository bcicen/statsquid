import json,yaml,logging
from datetime import datetime
from redis import StrictRedis
from container  import Container

log = logging.getLogger('dockerstats')

class Stat(object):
    """
    Stat object, created from json received from stat collector
    """
    def __init__(self,statjson):
        self.raw = statjson
        self.statdict = json.loads(self.raw)
        self.timestamp,self.timezone = self._readtime(self.statdict['read'])
        self.container_name = self.statdict['container_name'].split('/')[-1]
        self.container_id = self.statdict['container_id'].split('/')[-1]

        self.container_cpu = self.statdict['cpu_stats']['cpu_usage']['total_usage']
        self.system_cpu = self.statdict['cpu_stats']['system_cpu_usage']
        self.cpu_count =  len(self.statdict['cpu_stats']['cpu_usage']['percpu_usage'])

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

class StatListener(object):
    """
    StatListener subscribes to a redis channel and listens for messages from collectors,
    processing and storing them back in redis for persistence.
    params:
     - redis_host(str): redis host to connect to. default 127.0.0.1
     - redis_port(int): port to connect to redis host on. default 6379
    """
    def __init__(self,redis_host='127.0.0.1',redis_port=6379):
        
        self.redis = StrictRedis(host=redis_host,port=redis_port,db=0)
        self.sub = self.redis.pubsub(ignore_subscribe_messages=True)
        self.sub.subscribe('stats')

        self.run_forever()

    def run_forever(self):
        while True:
            msg = self.sub.get_message()
            if msg:
                self._process_msg(msg['data'])

    def _process_msg(self,msg):
        """
        message handler
        """
        msg = msg['data']
        stat = Stat(msg)
        cid = stat.container_id
        if cid not in self.containers:
            #create a new container object to track stats if we haven't
            #seen this container before
            self.containers[cid] = Container(cid,self.redis)
        self.containers[cid].append_stat(stat)
