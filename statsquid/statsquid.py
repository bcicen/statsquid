import os,sys,logging
from argparse import ArgumentParser
from . import __version__
from top import StatSquidTop
from listener import StatListener
from collector import StatCollector

log = logging.getLogger('statsquid')

def main():
    commands = [ 'agent', 'master', 'top' ]
    envvars = { 'STATSQUID_COMMAND' : 'command',
                'STATSQUID_REDIS_HOST' : 'redis_host',
                'STATSQUID_REDIS_PORT' : 'redis_port',
                'STATSQUID_DOCKER_HOST' : 'docker_host' }

    parser = ArgumentParser(description='statsquid %s' % __version__)
    parser.add_argument('--docker-host',
                        dest='docker_host',
                        help='docker host to connect to (default: unix://var/run/docker.sock)',
                        default='unix://var/run/docker.sock')
    parser.add_argument('--redis-host',
                        dest='redis_host',
                        help='redis host to connect to (default: 127.0.0.1)',
                        default='127.0.0.1')
    parser.add_argument('--redis-port',
                        dest='redis_port',
                        help='redis port to connect on (default: 6379)',
                        default='6379')
    parser.add_argument('--command',
                        dest='command',
                        help='statsquid mode (%s)' % ','.join(commands),
                        default=None)

    args = parser.parse_args()

    #override args with env vars
    [ args.__setattr__(v,os.getenv(k)) \
            for k,v in envvars.iteritems() if os.getenv(k) ]

    if not args.command:
        print('No command provided')
        exit(1)
    if args.command not in commands:
        print('Unknown command %s' % args.command)
        exit(1)

    if args.command == 'top':
        StatSquidTop(redis_host=args.redis_host,redis_port=args.redis_port)

    if args.command == 'master':
        s = StatListener(redis_host=args.redis_host,
                         redis_port=args.redis_port)

    if args.command == 'agent':
        s = StatCollector(args.docker_host,
                          redis_host=args.redis_host,
                          redis_port=args.redis_port)

if __name__ == '__main__':
    main()
