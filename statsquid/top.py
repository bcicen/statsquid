import os,sys,signal,curses
from time import time
from util import format_bytes
from redis import StrictRedis

class StatSquidTop(object):
    def __init__(self,redis_host='127.0.0.1',redis_port=6379):
        self.redis = StrictRedis()
        self.poll()

    def sig_handler(self, signal, frame):
        curses.endwin()
        sys.exit(0)

    #TODO: add cleanup of non-reporting/exited containers and clean adding of new ones
    def poll(self):
        stats = {}
        s = curses.initscr()
        s.timeout(1000)
        s.border(0)
        while True:
            signal.signal(signal.SIGINT, self.sig_handler)
            s.clear()

            #first build a dictionary with all containers
            for cid in self.redis.keys():
                stats[cid] = self.redis.hgetall(cid)
            
            #now display it 
            s.addstr(2, 2, "NAME")
            s.addstr(2, 25, "ID")
            s.addstr(2, 41, "CPU")
            s.addstr(2, 48, "MEM")
            s.addstr(2, 64, "NET TX")
            s.addstr(2, 80, "NET RX")
            s.addstr(2, 96, "HOST")
            cid_line = 4
            for cid in stats:
                s.addstr(cid_line, 2,  stats[cid]['name'][:20])
                s.addstr(cid_line, 25, stats[cid]['id'][:12])
                s.addstr(cid_line, 41, stats[cid]['cpu'])
                s.addstr(cid_line, 48, format_bytes(stats[cid]['mem']))
                s.addstr(cid_line, 64, format_bytes(stats[cid]['net_tx_bytes']))
                s.addstr(cid_line, 80, format_bytes(stats[cid]['net_rx_bytes']))
                s.addstr(cid_line, 96, stats[cid]['source'])
                cid_line += 1
            s.refresh()
            x = s.getch()
            if x == ord('q'):
                break
        curses.endwin()
