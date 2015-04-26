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
        self._set('id',self.id)
        self._set('stats_read',0)

        self.stats = []

    def append_stat(self,stat):
        if not self._get('name'):
            self._set('name', stat.name)

        if len(self.stats) > 0:
            last_stat = self.stats[-1]

            cpu_percent = self._calculate_cpu_percentage(stat,last_stat)
            self._set('cpu',round(cpu_percent,2))

            rx,tx = self._calculate_net_delta(stat,last_stat)
            self._set('net_rx', rx)
            self._set('net_tx', tx)

            write,read = self._calculate_io_delta(stat,last_stat)
            self._set('io_write', write)
            self._set('io_read', read)

        self._set('mem',float(stat.memory_stats.usage))
        self._set('net_tx_bytes_total',float(stat.network.tx_bytes))
        self._set('net_rx_bytes_total',float(stat.network.rx_bytes))
        self._set('source',stat.source)
        read_io,write_io = self._get_rw_io(stat)
        self._set('io_read_total',read_io)
        self._set('io_write_total',write_io)

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

    def _get_rw_io(self,stat):
        r,w = 0,0
        for s in stat.blkio_stats.io_service_bytes_recursive:
            if s['op'] == 'Read':
                r = s['value']
            if s['op'] == 'Write':
                w = s['value']
        return (r,w)

    def _calculate_io_delta(self,newstat,oldstat):
        time_delta = newstat.timestamp - oldstat.timestamp

        oldstat_read,oldstat_write = self._get_rw_io(oldstat)
        newstat_read,newstat_write = self._get_rw_io(newstat)

        write_delta = newstat_write - oldstat_write
        read_delta = newstat_read - oldstat_read
        if time_delta.total_seconds() > 1:
            write_delta = int(write_delta / time_delta.total_seconds())
            read_delta = int(read_delta / time_delta.total_seconds())

        return (write_delta,read_delta) 

    def _calculate_net_delta(self,newstat,oldstat):
        time_delta = newstat.timestamp - oldstat.timestamp
        rx_delta = newstat.network.rx_bytes - oldstat.network.rx_bytes
        tx_delta = newstat.network.tx_bytes - oldstat.network.tx_bytes
        if time_delta.total_seconds() > 1:
            rx_delta = rx_delta / time_delta.total_seconds()
            tx_delta = tx_delta / time_delta.total_seconds()

        return (rx_delta,tx_delta) 
        
    def _calculate_cpu_percentage(self,newstat,oldstat):
        """
        Calculate the cpu usage in percentage from two stats.
        """
        time_delta = newstat.timestamp - oldstat.timestamp
        sys_delta = newstat.cpu_stats.system_cpu_usage - \
                        oldstat.cpu_stats.system_cpu_usage
        container_delta = newstat.cpu_stats.cpu_usage.total_usage - \
                            oldstat.cpu_stats.cpu_usage.total_usage

        if time_delta.total_seconds() > 1:
            sys_delta = sys_delta / time_delta.total_seconds()
            container_delta = container_delta / time_delta.total_seconds()

        return (container_delta / sys_delta) * newstat.ncpu * cpu_tick
