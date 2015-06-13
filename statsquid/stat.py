import logging
from datetime import datetime,timedelta

log = logging.getLogger('statsquid')

class AttrDict(dict):
    __setattr__ = dict.__setitem__

    def __getattr__(self, k):
        v = self.__getitem__(k)
        return v if not isinstance(v, dict) else AttrDict(v)

    def __str__(self):
        return str(dict(self.__dict__))

    def __repr__(self):
        return "AttrDict({})".format(repr(self.__dict__))

    def __getstate__(self):
        return self.__dict__

class Stat(AttrDict):
    """
    Stat object, created from json received from stat collector
    """
    def __init__(self,statdict):
        super(Stat, self).__init__(statdict)
        self.id        = self.container_id.split('/')[-1]
        self.name      = self.container_name.split('/')[-1]
        self.timestamp = self._readtime(self.read)

    def _readtime(self,timestamp):
        #TODO: use time.strptime
        d,t = timestamp.split('T')
        year,month,day = d.split('-')

        if '-' in t:
            t,tz = t.split('-')
            tz = tz
        elif '+' in t:
            t,tz = t.split('-')
            tz = '-' + tz
        else:
            tz = None

        hour,minute,second = t.split(':')
        second,microsecond = second.split('.')

        ts = datetime(int(year),
                      int(month),
                      int(day),
                      int(hour),
                      int(minute),
                      int(second),
                      int(microsecond[0:6]))
        if tz:
            ts = ts + timedelta(hours=int(tz.strip(':00')))

        return ts
