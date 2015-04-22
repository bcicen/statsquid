import logging,json
from util import unix_time

log = logging.getLogger('statsquid')

cpu_tick = 100

class Container(object):
    """
    Container object holds a collection of stats for a specific container_id,
    rolling up the data at regular intervals.
    params:
     - container_id(str): Docker container id
     - redis(obj): Instance of a redis client object
     - flush_interval: Optional. Number of stats to keep in memory before flush
    methods:
     - append_stat: Appends a new stat, recalculating averages
     params:
       - stat(obj): A statsquid.stat object
    """
    def __init__(self,container_id,redis,flush_interval=60):
        self.id = container_id
        self.redis = redis
        self.flush_interval = flush_interval

        #setup initial fields in redis 
        self._set('id',container_id)
        self._set('stats_read',0)

        self.stats = []

    def append_stat(self,stat):
        if not self._get('name'):
            self._set('name', stat.container_name)

        if len(self.stats) > 0:
            cpu_percent = self._calculate_cpu_percentage(stat,self.stats[-1])
            self._set('cpu',round(cpu_percent,2))
            rx,tx = self._calculate_net_delta(stat,self.stats[-1])
            self._set('net_rx', rx)
            self._set('net_tx', tx)
        self._set('mem',float(stat.statdict['memory_stats']['usage']))
        self._set('net_tx_bytes_total',float(stat.statdict['network']['tx_bytes']))
        self._set('net_rx_bytes_total',float(stat.statdict['network']['rx_bytes']))
        self._set('source',stat.statdict['source'])
        #TODO: add io read/write metrics
        #self._set('io_read_bytes',float(stat.statdict['io_service_bytes_recursive']['rx_bytes'])
        #self._set('io_write_bytes',float(stat.statdict['io_service_bytes_recursive']['rx_bytes'])

        #TODO: utilize redis increment
        self._set('stats_read', int(self._get('stats_read')) + 1)
        self._set('last_read', unix_time(stat.timestamp))
        self.stats.append(stat)

        if len(self.stats) > self.flush_interval: 
            self._flush()

    def _get(self,attribute):
        r = self.redis
        return r.hget(self.id, attribute)

    def _set(self,attribute,value):
        r = self.redis
        r.hset(self.id, attribute, value) 

    def _flush(self):
        del self.stats[:-1] # remove all but most recent stat
        log.debug('flush performed for container %s' % self.id)

    def _calculate_net_delta(self,newstat,oldstat):
        time_delta = newstat.timestamp - oldstat.timestamp
        rx_delta = newstat.statdict['network']['rx_bytes'] - oldstat.statdict['network']['rx_bytes']
        tx_delta = newstat.statdict['network']['tx_bytes'] - oldstat.statdict['network']['tx_bytes']
        if time_delta.total_seconds() > 1:
            rx_delta = rx_delta / time_delta.total_seconds()
            tx_delta = tx_delta / time_delta.total_seconds()

        return (rx_delta,tx_delta) 
        
    def _calculate_cpu_percentage(self,newstat,oldstat):
        """
        Calculate the cpu usage in percentage form from two stats.
        """
        time_delta = newstat.timestamp - oldstat.timestamp
        sys_delta = newstat.system_cpu - oldstat.system_cpu
        container_delta = newstat.container_cpu - oldstat.container_cpu
        if time_delta.total_seconds() > 1:
            sys_delta = sys_delta / time_delta.total_seconds()
            container_delta = container_delta / time_delta.total_seconds()

        return (container_delta / sys_delta) * newstat.cpu_count * cpu_tick
