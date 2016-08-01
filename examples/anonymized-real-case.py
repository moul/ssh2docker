#!/usr/bin/env python2
# coding: utf-8

# this example is used in production, it is depending on private libraries
# to communicate with internal APIs, but it can help you build your own
# production company-specific hook.

import sys
import json
import pprint
pp = pprint.PrettyPrinter(indent=4)

sys.path.insert(0, '/opt/python-provisioning')
from tools.verbose_logging import logging
from localcloud import compute
from api import web_hosting


# version of remote container
ALPINE_VERSION = '3.3'


def nope(msg):
    return {'allowed': False, 'message': msg}


def archify(arch):
    return {
        'arm': 'armhf',
        'x86_64': 'amd64',
    }[arch]


def auth(hosting_id, *keys):

    if len(keys) < 1:
        return nope('no ssh key')

    granted = False
    web_hosting_int = web_hosting.WebHosting(hosting_id)
    for key in keys:

        try:
            if web_hosting_int.is_valid_ssh_key(key) == True:
                granted = True
                break
        except Exception as e:
            logging.error(e)
            return nope('http error')

    if not granted:
        return nope('access denied')

    compute_int = compute.Compute()
    try:
        server = compute_int.get_server_by_webid(hosting_id)
        logging.debug(pp.pformat(server))
    except Exception as e:
        logging.error(e)
        return nope('error while trying to resolve server')

    return {

        'allowed': True,

        'remote-user': hosting_id,

        'image-name': 'local_web/alpine:{}-{}'.format(archify(server['arch']), ALPINE_VERSION),

        'docker-run-args': [
            '--name', 'ssh2docker_{}'.format(hosting_id),
            '--hostname', server['name'],
            '--rm',
            '-it',
            '-v', '/storage/users/{}/ftp:/ftp:rw'.format(hosting_id),
            '-v', '/storage/users/{}/backups:/ftp/backups:ro'.format(hosting_id),
            '-v', '/storage/users/{}/logs:/ftp/logs:ro'.format(hosting_id),
            '-v', '/storage/users/{}/websites:/ftp/websites:rw'.format(hosting_id),
            '-m', '256m',
            '--cpu-shares', '512',  # default = 1024, so ssh2docker gets half quota
            '-u', 'webuser',
        ],

        'env': {
            'DOCKER_HOST': 'tcp://{}.local:2376'.format(server['id']),
            'DOCKER_TLS_VERIFY': '1',
            'DOCKER_CERT_PATH': '/opt/docker-tls/{}/.docker/'.format(server['hostname']),
        },

        'command': ['/bin/sh', '-i', '-l'],

    }


print(json.dumps(auth(*sys.argv[1:])))
