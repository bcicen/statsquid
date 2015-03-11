import logging

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
     params:
       - stat(obj): A dockerstats.stat object
    """
    #TODO: add rollup and averaging over time for metrics
    def __init__(self,container_id):
        self.id = container_id
        self.name = ""
        self.stats_read = 0

    def append_stat(self,stat):
        if not self.name:
            self.name = stat.container_name

        #cpu
        self.cpu = float(stat.statdict['cpu_stats']['system_cpu_usage'])

        #memory
        self.memory_usage = self._format_bytes(stat.statdict['memory_stats']['usage'])

        #network
        netin = stat.statdict['network']['rx_bytes']
        netout = stat.statdict['network']['tx_bytes']
        self.net_in = self._format_bytes(netin)
        self.net_out = self._format_bytes(netout)

        self.stats_read += 1
        
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
