#!/usr/bin/env python

import os,sys,logging,signal
from argparse import ArgumentParser
#from . import __version__
from listener import StatListener
from collector import StatCollector
from top import StatSquidTop

__version__ = 'alpha'
log = logging.getLogger('statsquid')

class StatSquid(object):
    """
    StatSquid 
    params:
     - role(str): Role of this statsquid instance. Either master or agent.
     - options(dict): dictionary of options to start instance with
    """
    #TODO: improve graceful exiting, fix signal catching
    def __init__(self,role,options):
        self.role = role
        signal.signal(signal.SIGTERM, self.sig_handler)
        print('Starting statsquid %s' % role)
        if self.role == 'master':
            self.instance = self.start_master(options)
        if self.role == 'agent':
            self.instance = self.start_agent(options)
        
    def start_master(self,opts):
        return StatListener(redis_host=opts['redis_host'],
                            redis_port=opts['redis_port'])

    def start_agent(self,opts):
        return StatCollector(opts['docker_host'],
                            redis_host=opts['redis_host'],
                            redis_port=opts['redis_port'])

    def sig_handler(self,signal,frame):
        print('signal caught, exiting')
        self.instance.stop()
        sys.exit(0)

def main():
    commands = [ 'agent', 'master', 'top' ]

    parser = ArgumentParser(description='statsquid %s' % __version__)
    parser.add_argument('--docker-host',
                        dest='docker_host',
                        help='docker host to connect to (default: tcp://127.0.0.1:4243)',
                        default='tcp://127.0.0.1:4243')
    parser.add_argument('--docker-port',
                        dest='docker_port',
                        help='docker port to connect on (default: 4243)',
                        default=4243)
    parser.add_argument('--redis-host',
                        dest='redis_host',
                        help='redis host to connect to (default: 127.0.0.1)',
                        default='127.0.0.1')
    parser.add_argument('--redis-port',
                        dest='redis_port',
                        help='redis port to connect on (default: 6379)',
                        default='6379')
    parser.add_argument('command',
                        help='Mode to run as or command to run (%s)' % \
                                ','.join(commands))

    args = parser.parse_args()

    if args.command not in commands:
        log.error('Unknown command %s' % args.command)
        exit(1)

    if args.command == 'top':
        StatSquidTop(redis_host=args.redis_host,redis_port=args.redis_port)
    else:
        s = StatSquid(args.command,args.__dict__)


if __name__ == '__main__':
    main()
