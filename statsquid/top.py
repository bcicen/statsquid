import os,sys,signal,curses
from datetime import datetime
from util import format_bytes,unix_time
from redis import StrictRedis

class StatSquidTop(object):
    def __init__(self,redis_host='127.0.0.1',redis_port=6379):
        self.redis = StrictRedis(host=redis_host,port=redis_port)
        self.keys = [
                'name',
                'id',
                'cpu',
                'mem',
                'net_rx',
                'net_tx',
                'source'
                ]
        self.poll()

    def sig_handler(self, signal, frame):
        curses.endwin()
        sys.exit(0)

    def poll(self):
        stats = {}
        s = curses.initscr()
        s.timeout(1000)
        s.border(0)
        while True:
            signal.signal(signal.SIGINT, self.sig_handler)
            s.clear()

            #first build a dictionary with all containers
            now = datetime.now()
            now_seconds = unix_time(now)
            for cid in self.redis.keys():
                cidstats = self.redis.hgetall(cid)
                #only include containers with all req keys
                if not False in [cidstats.has_key(k) for k in self.keys]:
                    #and only if update within last 5s
                    if now_seconds - int(stats[cid]['last_read']) < 5:
                        stats[cid] = cidstats

            #first line
            s.addstr(1, 2, 'statsquid top -')
            s.addstr(1, 18, now.strftime('%H:%m:%S'))
            s.addstr(1, 28, ('%s containers' % len(stats)))
            s.addstr(3, 2, "NAME")
            s.addstr(3, 25, "ID")
            s.addstr(3, 41, "CPU")
            s.addstr(3, 48, "MEM")
            s.addstr(3, 64, "NET TX")
            s.addstr(3, 80, "NET RX")
            s.addstr(3, 96, "HOST")
            line = 5
            for cid in stats:
                s.addstr(line, 2,  stats[cid]['name'][:20])
                s.addstr(line, 25, stats[cid]['id'][:12])
                s.addstr(line, 41, stats[cid]['cpu'])
                s.addstr(line, 48, format_bytes(stats[cid]['mem']))
                s.addstr(line, 64, format_bytes(stats[cid]['net_tx']))
                s.addstr(line, 80, format_bytes(stats[cid]['net_rx']))
                s.addstr(line, 96, stats[cid]['source'])
                line += 1
            s.refresh()
            x = s.getch()
            if x == ord('q'):
                break
        curses.endwin()
