import logging
import json

from statsquid.util import unix_time

log = logging.getLogger('statsquid')

cpu_tick = 100

class Container(object):
    """
    Container object holds a collection of stats for a specific container_id
    params:
     - container_id(str): Docker container id
     - redis(obj): Instance of a redis client object
    methods:
     - append_stat: Appends a new stat, recalculating averages
     params:
       - stat(obj): A statsquid.stat object
    """
    def __init__(self,container_id,redis):
        self.id = container_id
        self.redis = redis

        self.current = { 'id':self.id }

        self.stats = []

    def append_stat(self,stat):
        self.last_read = unix_time(stat.timestamp)

        self.current['name'] = stat.name
        self.current['source'] = stat.source
        self.current['last_read'] = self.last_read

        if len(self.stats) > 0:
            last_stat = self.stats[-1]
            self.current['cpu'] = self._calculate_cpu(stat,last_stat)

        self.current['mem'] = float(stat.memory_stats.usage)
        self.current['net_tx_bytes_total'] = float(stat.network.tx_bytes)
        self.current['net_rx_bytes_total'] = float(stat.network.rx_bytes)

        read_io,write_io = self._get_rw_io(stat)
        self.current['io_read_bytes_total'] = read_io
        self.current['io_write_bytes_total'] = write_io

        self.redis.hincrby(self.id,'stats_read', amount=1)
        self.redis.hmset(self.id,self.current)

        self.stats.append(stat)

    def flush(self):
        del self.stats[:-1] # remove all but most recent stat
        log.debug('flush performed for container %s' % self.id)

    def delete(self):
        self.redis.delete(self.id)

    def _get_rw_io(self,stat):
        r,w = 0,0
        for s in stat.blkio_stats.io_service_bytes_recursive:
            if s['op'] == 'Read':
                r = s['value']
            if s['op'] == 'Write':
                w = s['value']
        return (r,w)
        
    def _calculate_cpu(self,newstat,oldstat):
        """
        Calculate the cpu usage in percentage from two stats.
        """
        time_delta = newstat.timestamp - oldstat.timestamp
        sys_delta = newstat.cpu_stats.system_cpu_usage - \
                        oldstat.cpu_stats.system_cpu_usage
        container_delta = newstat.cpu_stats.cpu_usage.total_usage - \
                            oldstat.cpu_stats.cpu_usage.total_usage

        #for odd cases where system_cpu_usage hasn't changed
        if not sys_delta:
            return 0

        if time_delta.total_seconds() > 1:
            sys_delta = sys_delta / time_delta.total_seconds()
            container_delta = container_delta / time_delta.total_seconds()

        cpu_percent = (container_delta / sys_delta) * newstat.ncpu * cpu_tick
        
        return round(cpu_percent,2)
