import os,sys,signal,curses
from curses.textpad import Textbox,rectangle
from datetime import datetime
from util import format_bytes,unix_time
from redis import StrictRedis

class StatSquidTop(object):
    def __init__(self,redis_host='127.0.0.1',redis_port=6379,filter=None):
        self.filter = filter
        self.redis = StrictRedis(host=redis_host,port=redis_port)
        self.keys = [
                'name',
                'id',
                'cpu',
                'mem',
                'net_rx',
                'net_tx',
                'io_read',
                'io_write',
                'source'
                ]
        self.poll()

    def sig_handler(self, signal, frame):
        curses.endwin()
        sys.exit(0)

    def poll(self):
        s = curses.initscr()
        curses.noecho()
        s.timeout(1000)
        s.border(0)
        while True:
            h,w = s.getmaxyx()
            signal.signal(signal.SIGINT, self.sig_handler)
            s.clear()

            #first build a dictionary with all containers
            stats = {}
            now = datetime.utcnow()
            now_seconds = unix_time(now)
            for cid in self.redis.keys():
                cidstats = self.redis.hgetall(cid)
                #only include containers with all req keys
                if not False in [cidstats.has_key(k) for k in self.keys]:
                    #and only if update within last 5s
                    if now_seconds - int(cidstats['last_read']) < 10:
                        stats[cid] = cidstats

            #TODO: add filtering for name, host, id based on "host:<str>" filter
            if self.filter:
                stats = { k:v for k,v in stats.iteritems() \
                          if self.filter in stats[k]['name'] }

            #first line
            s.addstr(1, 2, 'statsquid top -')
            s.addstr(1, 18, now.strftime('%H:%M:%S'))
            s.addstr(1, 28, ('%s containers' % len(stats)))
            if self.filter:
                s.addstr(1, 42, ('filter: %s' % self.filter))

            #second line
            s.addstr(3, 2, "NAME", curses.A_BOLD)
            s.addstr(3, 25, "ID", curses.A_BOLD)
            s.addstr(3, 41, "CPU", curses.A_BOLD)
            s.addstr(3, 48, "MEM", curses.A_BOLD)
            s.addstr(3, 58, "NET TX", curses.A_BOLD)
            s.addstr(3, 68, "NET RX", curses.A_BOLD)
            s.addstr(3, 78, "READ IO", curses.A_BOLD)
            s.addstr(3, 88, "WRITE IO", curses.A_BOLD)
            s.addstr(3, 98, "HOST", curses.A_BOLD)

            #remainder of lines
            line = 5
            maxlines = h - 2
            for cid in stats:
                s.addstr(line, 2,  stats[cid]['name'][:20])
                s.addstr(line, 25, stats[cid]['id'][:12])
                s.addstr(line, 41, stats[cid]['cpu'])
                s.addstr(line, 48, format_bytes(stats[cid]['mem']))
                s.addstr(line, 58, format_bytes(stats[cid]['net_tx']))
                s.addstr(line, 68, format_bytes(stats[cid]['net_rx']))
                s.addstr(line, 78, format_bytes(stats[cid]['io_read']))
                s.addstr(line, 88, format_bytes(stats[cid]['io_write']))
                s.addstr(line, 98, stats[cid]['source'])
                if line >= maxlines:
                    break
                line += 1
            s.refresh()
            x = s.getch()
            if x == ord('q'):
                break
            if x == ord('f'):
                startx = w / 2 - 20 # I have no idea why this offset of 20 is needed

                s.addstr(10, startx, "String to filter for:")

                editwin = curses.newwin(1,30, 12,(startx+1))
                rectangle(s, 11,startx, 13,(startx+31))
                s.refresh()

                box = Textbox(editwin)
                box.edit()

                self.filter = str(box.gather()).strip(' ')

        curses.endwin()
