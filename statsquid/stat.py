import json,logging,jsontree
from datetime import datetime,timedelta

log = logging.getLogger('statsquid')

class Stat(object):
    """
    Stat object, created from json received from stat collector
    """
    def __init__(self,stat_json):
        self.__dict__ = jsontree.loads(stat_json)
        self.timestamp = self._readtime(self.read)
        self.name = self.container_name.split('/')[-1]
        self.id = self.container_id.split('/')[-1]

    def _readtime(self,timestamp):
        #TODO: use time.strptime
        d,t = timestamp.split('T')
        year,month,day = d.split('-')
        if '-' in t:
            t,tz = t.split('-')
            tz = tz
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
        ts = ts + timedelta(hours=int(tz.strip(':00')))
        return ts
