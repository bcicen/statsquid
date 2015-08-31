import logging
import msgpack
from datetime import datetime,timedelta
from redis import StrictRedis

from statsquid.stat import Stat
from statsquid.container import Container
from statsquid.util import output,unix_time

log = logging.getLogger('statsquid')

class StatListener(object):
    """
    StatListener subscribes to a redis channel and listens for messages from collectors,
    processing and storing them back in redis for persistence.
    params:
     - redis_host(str): redis host to connect to. default 127.0.0.1
     - redis_port(int): port to connect to redis host on. default 6379
    """
    def __init__(self,redis_host='127.0.0.1',redis_port=6379):
        self.containers = {}
        self.maint_interval = 10
        self.last_maint = datetime.now()

        self.redis = StrictRedis(host=redis_host,
                                 port=redis_port,
                                 decode_responses=True)
        self.sub = self.redis.pubsub(ignore_subscribe_messages=True)
        self.sub.subscribe('statsquid')

        self.run_forever()

    def run_forever(self):
        stat_count = 0
        output('listener started')
        for msg in self.sub.listen():
            self._process_msg(msg['data'])
            stat_count += 1
            if self._is_maint_interval():
                output('processed %s stats in last %ss' % \
                        (stat_count,self.maint_interval))
                stat_count = 0
                self._flush_all()

    def _is_maint_interval(self):
        diff_seconds = int((datetime.now() - self.last_maint).total_seconds()) 
        if diff_seconds >= self.maint_interval:
            self.last_maint = datetime.now()
            return True
        return False

    def _flush_all(self):
        now = unix_time(datetime.utcnow())
        containers = self.containers.copy()
        for cid,c in list(containers.items()):
            c.flush()
            if now - c.last_read > self.maint_interval:
                c.delete()
                del self.containers[cid]
                log.debug('cleared stale container %s' % cid)

    def _process_msg(self, data):
        """
        Message handler
        """
        stat = Stat(self._unpack(data))
        cid = stat.container_id
        if cid not in self.containers:
            #create a new container object to track stats if we haven't
            #seen this container before
            self.containers[cid] = Container(cid, self.redis)
        self.containers[cid].append_stat(stat)

    @staticmethod
    def _unpack(data):
        return msgpack.unpackb(data, encoding='utf-8')
