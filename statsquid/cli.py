import os
import sys
import logging
from argparse import ArgumentParser

from statsquid.version import version
from statsquid.agent import Agent
from statsquid.top import StatSquidTop
from statsquid.listener import StatListener

log = logging.getLogger('statsquid')

def main():
    envvars = { 'STATSQUID_REDIS' : 'redis',
                'DOCKER_HOST'     : 'docker_host' }

    common_parser = ArgumentParser(add_help=False)
    common_parser.add_argument('--redis',
                        dest='redis',
                        help='redis host to connect to (127.0.0.1:6379)',
                        default='127.0.0.1:6379')

    parser = ArgumentParser(description='statsquid %s' % (version))
    subparsers = parser.add_subparsers(description='statsquid subcommands',
                                       dest='subcommand')

    #master
    parser_master = subparsers.add_parser('master',parents=[common_parser])

    #agent
    parser_agent = subparsers.add_parser('agent',parents=[common_parser])
    parser_agent.add_argument('--docker-host',
        dest='docker_host',
        help='docker host to connect to (unix://var/run/docker.sock)',
        default='unix://var/run/docker.sock')

    #top
    parser_top = subparsers.add_parser('top',parents=[common_parser])

    args = parser.parse_args()
    #override command line with env vars
    [ args.__setattr__(v,os.getenv(k)) for k,v \
            in list(envvars.items()) if os.getenv(k) ]

    if ':' in args.redis:
        redis_host,redis_port = args.redis.split(':')
    else:
        redis_host = args.redis
        redis_port = 6379

    if args.subcommand == 'top':
        StatSquidTop(redis_host=redis_host,redis_port=redis_port)

    if args.subcommand == 'master':
        s = StatListener(redis_host=redis_host,redis_port=redis_port)

    if args.subcommand == 'agent':
        s = Agent(args.docker_host,
                  redis_host=redis_host,
                  redis_port=redis_port)

if __name__ == '__main__':
    main()
