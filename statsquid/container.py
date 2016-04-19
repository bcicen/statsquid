import logging
import json

from statsquid.util import unix_time
from statsquid import key_prefix

log = logging.getLogger('statsquid')

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
    def __init__(self, container_id, redis):
        self.id = container_id
        self.key = key_prefix + ':' + self.id
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
        tx_bytes, rx_bytes = self._get_aggr_net(stat)
        self.current['net_tx_bytes_total'] = tx_bytes
        self.current['net_rx_bytes_total'] = rx_bytes

        read_io, write_io = self._get_rw_io(stat)
        self.current['io_read_bytes_total'] = read_io
        self.current['io_write_bytes_total'] = write_io

        self.redis.hincrby(self.key, 'stats_read', amount=1)
        self.redis.hmset(self.key, self.current)

        self.stats.append(stat)

    def flush(self):
        del self.stats[:-1] # remove all but most recent stat
        log.debug('flush performed for container %s' % self.id)

    def delete(self):
        self.redis.delete(self.key)

    def _get_aggr_net(self, stat):
        tx, rx = float(), float()
        for iface, net in stat.networks.items():
            tx += net['tx_bytes']
            rx += net['rx_bytes']
        return tx, rx

    def _get_rw_io(self, stat):
        r,w = 0,0
        for s in stat.blkio_stats.io_service_bytes_recursive:
            if s['op'] == 'Read':
                r = s['value']
            if s['op'] == 'Write':
                w = s['value']
        return (r,w)

    @staticmethod
    def _calculate_cpu(newstat, oldstat):
        """
        Calculate the cpu usage in percentage from two stats.
        """
        cpu_tick = 100

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
