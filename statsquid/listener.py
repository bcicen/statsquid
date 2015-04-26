import logging,msgpack
from datetime import datetime,timedelta
from redis import StrictRedis
from util import output
from container import Container
from stat import Stat

log = logging.getLogger('statsquid')

class StatListener(object):
    """
    StatListener subscribes to a redis channel and listens for messages from collectors,
    processing and storing them back in redis for persistence.
    params:
     - redis_host(str): redis host to connect to. default 127.0.0.1
     - redis_port(int): port to connect to redis host on. default 6379
    """
    def __init__(self,redis_host='127.0.0.1',redis_port=6379,log_interval=60):
        self.containers = {}
        self.log_interval = log_interval
        self.last_log = datetime.now()

        self.redis = StrictRedis(host=redis_host,port=redis_port,db=0)
        self.sub = self.redis.pubsub(ignore_subscribe_messages=True)
        self.sub.subscribe('stats')

        self.run_forever()

    def run_forever(self):
        stat_count = 0
        output('listener started')
        for msg in self.sub.listen():
            self._process_msg(msgpack.unpackb(msg['data']))
            stat_count += 1
            if self._is_log_interval():
                output('processed %s stats in last %ss' % \
                        (stat_count,self.log_interval))
                stat_count = 0

    def _is_log_interval(self):
        diff_seconds = int((datetime.now() - self.last_log).total_seconds()) 
        if diff_seconds >= self.log_interval:
            self.last_log = datetime.now()
            return True
        return False

    def _process_msg(self,statdict):
        """
        message handler
        """
        stat = Stat(statdict)
        cid = stat.container_id
        if cid not in self.containers:
            #create a new container object to track stats if we haven't
            #seen this container before
            self.containers[cid] = Container(cid,self.redis)
        self.containers[cid].append_stat(stat)
