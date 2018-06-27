#!/usr/bin/python
# -*- coding: utf-8 -*-

"""
    Initializer on pure Python 2.7
    mode:
    python initializer.py  -h                              get help information
    python initializer.py                                  start full install
    python initializer.py --build                          create cmd for dapp
    python initializer.py --vpn start/stop/restart/status  control vpn servise
    python initializer.py --comm start/stop/restart/status control common servise
    python initializer.py --mass start/stop/restart/status control common + vpn servise
    python initializer.py --test                           start in test mode
    python initializer.py --no-gui                         install without GUI
    python initializer.py --update                         update all contaiter without GUI
    python initializer.py --mass-update                    update all contaiter with GUI
    python initializer.py --update-gui                     update only GUI
"""

import sys
import logging
import socket
from signal import SIGINT, signal, pause
from contextlib import closing
from re import search, sub, findall
from codecs import open
from shutil import copyfile, rmtree
from json import load, dump
from time import time, sleep
from urllib import URLopener
from urllib2 import urlopen
from os.path import isfile, isdir
from argparse import ArgumentParser
from ConfigParser import ConfigParser
from platform import linux_distribution
from subprocess import Popen, PIPE, STDOUT
from distutils.version import StrictVersion
from stat import S_IEXEC, S_IXUSR, S_IXGRP, S_IXOTH
from os import remove, mkdir, path, environ, stat, chmod

"""
Exit code:
    1 - Problem with get or upgrade systemd ver
    2 - If version of Ubuntu lower than 16
    3 - If sysctl net.ipv4.ip_forward = 0 after sysctl -w net.ipv4.ip_forward=1
    4 - Problem when call system command from subprocess
    5 - Problem with operation R/W unit file 
    6 - Problem with operation download file 
    7 - Problem with operation R/W server.conf 
    8 - Default DB conf is empty, and no section 'DB' in dappctrl-test.config.json
    9 - Check the run of the database is negative
    10 - Problem with read dapp cmd from file
    11 - Problem NPM
    12 - Problem with run psql
    13 - Problem with ready Vpn 
    14 - Problem with ready Common 
    15 - The version npm is not satisfied. The user self reinstall
    16 - Problem with deleting one of the GUI pack
    17 - Exception in code, see logging
    18 - Exit Ctrl+C
    19 - OS not supported
    20 - In build mode,None in dappctrl_id
    21 - User from which was installing not root

"""

log_conf = dict(
    filename='/var/log/initializer.log',
    datefmt='%m/%d %H:%M:%S',
    format='%(levelname)7s [%(lineno)3s] %(message)s')
log_conf.update(level='DEBUG')
logging.basicConfig(**log_conf)
logging.getLogger().addHandler(logging.StreamHandler())

main_conf = dict(
    iptables=dict(
        link_download='http://art.privatix.net/',
        file_download=[
            'vpn.tar.xz',
            'common.tar.xz',
            'systemd-nspawn@vpn.service',
            'systemd-nspawn@common.service'],
        path_download='/var/lib/container/',
        path_vpn='vpn/',
        path_com='common/',
        path_unit='/lib/systemd/system/',
        openvpn_conf='/etc/openvpn/config/server.conf',
        openvpn_fields=[
            'server {} {}',
            'push "route {} {}"'
        ],
        openvpn_tun='dev {}',
        openvpn_port='port 443',

        unit_vpn='systemd-nspawn@vpn.service',
        unit_com='systemd-nspawn@common.service',
        unit_field={
            'ExecStopPost=/sbin/sysctl': False,

            'ExecStartPre=/sbin/iptables': 'ExecStartPre=/sbin/iptables -t nat -A POSTROUTING -s {} -o {} -j MASQUERADE\n',
            'ExecStopPost=/sbin/iptables': 'ExecStopPost=/sbin/iptables -t nat -D POSTROUTING -s {} -o {} -j MASQUERADE\n',
        }

    ),

    build={
        'cmd': '/opt/privatix/initializer/dappinst -dappvpnconftpl=\'{0}\' -dappvpnconf={1} -connstr=\"{3}\" -template={4} -agent=true\n'
               '/opt/privatix/initializer/dappinst -dappvpnconftpl=\'{0}\' -dappvpnconf={2} -connstr=\"{3}\" -template={4} -agent=false',
        'cmd_path': '.dapp_cmd',

        'db_conf': {
            "dbname": "dappctrl",
            "sslmode": "disable",
            "user": "postgres",
            "host": "localhost",
            "port": "5433"
        },
        'db_log': '/var/lib/container/common/var/log/postgresql/postgresql-10-main.log',
        'db_stat': 'database system is ready to accept connections',

        'dappvpnconf_path': '/var/lib/container/vpn/opt/privatix/config/dappvpn.config.json',
        'dappconconf_path': '/var/lib/container/common/opt/privatix/config/dappvpn.config.json',
        'conf_link': 'https://raw.githubusercontent.com/Privatix/dappctrl/release/0.6.0/dappctrl.config.json',
        'templ': 'https://raw.githubusercontent.com/Privatix/dappctrl/release/0.6.0/svc/dappvpn/dappvpn.config.json',
        'dappctrl_conf_local': '/var/lib/container/common/opt/privatix/config/dappctrl.config.local.json',
        'dappctrl_search_field': 'PayAddress',
        'dappctrl_id_raw': 'https://raw.githubusercontent.com/Privatix/dappctrl/develop/data/prod_data.sql',
        'field_name_id': '--templateid = ',
        'dappctrl_id': None,
    },

    final={'dapp_port': [], 'vpn_port': 443},

    gui={
        'gui_path': '/opt/privatix/gui/',

        'icon_tmpl_f': '{}{}/Desktop/privatix-dappgui.desktop',
        'icon_tmpl': {
            'Section': 'Desktop Entry',
            'Comment': 'First Internet Broadband Marketplace powered by P2P VPN Network on Blockchain',
            'Terminal': 'false',
            'Name': 'Privatix Dapp',
            'Exec': 'sh -c "sudo /opt/privatix/initializer/initializer.py --mass start && sudo npm start --prefix /opt/privatix/gui/node_modules/dappctrlgui/"',
            'Type': 'Application',
            'Icon': '/opt/privatix/gui/node_modules/dappctrlgui/icon_64.png',
        },

        'npm_tmp_f': 'tmp_nodesource',
        'npm_url': 'https://deb.nodesource.com/setup_9.x',
        'npm_tmp_file_call': 'sudo -E bash ',
        'npm_node': 'sudo apt-get install -y nodejs',

        'gui_inst': [
            'chown -R $USER:$(id -gn $USER) /opt/privatix/gui/',
            'sudo su - $USER -c \'cd /opt/privatix/gui && sudo npm install dappctrlgui\''
        ],
        'chown': 'sudo chown -R {0}:$(id -gn {0}) {1}',
        'version': {
            'npm': ['5.6', None, '0'],
            'nodejs': ['9.0', None, '0'],
        },

    },
    del_pack='sudo apt purge {} -y',
    del_dirs='sudo rm -rf {}*',
    search_pack='sudo dpkg -l | grep {}',

    test={
        'path': 'test_data.sql',
        'sql': 'https://raw.githubusercontent.com/Privatix/dappctrl/develop/data/test_data.sql',
        'cmd': 'psql -d dappctrl -h 127.0.0.1 -P 5433 -f {}'
    },

    addr='10.217.3.0',
    mask=['/24', '255.255.255.0'],
    mark_final='/var/run/installer.pid',
    ports=dict(vpn_port=None, dapp_port=None)
)


class CMD:
    recursion = 0

    def __init__(self):
        self.f_vpn = main_conf['iptables']['unit_vpn']
        self.f_com = main_conf['iptables']['unit_com']
        self.p_dest = main_conf['iptables']['path_unit']
        self.p_dwld = main_conf['iptables']['path_download']
        self.params = main_conf['iptables']['unit_field']
        self.path_vpn = main_conf['iptables']['path_vpn']
        self.path_com = main_conf['iptables']['path_com']

    def _reletive_path(self, name):
        dirname = path.dirname(__file__)
        return path.join(dirname, name)

    def signal_handler(self, sign, frm):
        logging.info('You pressed Ctrl+C!')
        self._rolback(sysctl=False, code=18)
        pause()

    def _rolback(self, sysctl, code):
        # Rolback net.ipv4.ip_forward
        if not sysctl:
            logging.debug('Rolback ip_forward')
            cmd = '/sbin/sysctl -w net.ipv4.ip_forward=0'
            self._sys_call(cmd)

        self.clear_contr(pass_check=True)
        sys.exit(code)

    def service(self, srv, status, port, reverse=False):
        logging.debug(
            'Service:{}, port:{}, status:{}, reverse:{}'.format(srv, port,
                                                                status,
                                                                reverse))
        tmpl = ['systemctl {} {} && sleep 0.5']
        rmpl_rest = ['systemctl stop {1} && sleep 0.5',
                     'systemctl start {1} && sleep 0.5']
        rmpl_stat = ['systemctl is-active {1}']

        scroll = {'start': tmpl, 'stop': tmpl,
                  'restart': rmpl_rest, 'status': rmpl_stat}
        unit_serv = {'vpn': self.f_vpn, 'comm': self.f_com}

        if status not in scroll.keys():
            logging.error('Status {} not suitable for service {}. '
                          'Status must be one from {}'.format(
                status, srv, scroll.keys())
            )
            return None

        raw_res = list()
        for cmd in scroll[status]:
            cmd = cmd.format(status, unit_serv[srv])
            res = self._sys_call(cmd, rolback=False)

            if not port:
                continue

            if status == 'status':
                if res == 'active\n':
                    if reverse:
                        return self._checker_port(port, 'stop')
                    else:
                        return self._checker_port(port)
                else:
                    return False

            if status == 'restart':
                check_stat = 'start' if 'start' in cmd else 'stop'
            else:
                check_stat = status

            if 'failed' in res or not self._checker_port(port, check_stat):
                return False
            raw_res.append(True)

        if not port:
            return None
        return all(raw_res)

    def clear_contr(self, pass_check=False):
        # Stop container.Check it if pass_check True.Clear conteiner path
        if pass_check:
            logging.info('\n\n    Attention! \n'
                         ' During installation, a failure occurred. \n'
                         ' All installed will be removed and returned to '
                         'the initial state.\n Wait for the end.\n '
                         'And try again.')
        self.service('vpn', 'stop', main_conf['ports']['vpn_port'])
        self.service('comm', 'stop', main_conf['ports']['dapp_port'])
        sleep(3)

        p_dowld = main_conf['iptables']['path_download']
        logging.debug('Crear {}*'.format(p_dowld))

        if pass_check or not self.service('vpn', 'status',
                                          main_conf['ports']['vpn_port'],
                                          True) and \
                not self.service('comm', 'status',
                                 main_conf['ports']['dapp_port'], True):
            rmtree(p_dowld + main_conf['iptables']['path_vpn'],
                   ignore_errors=True)
            rmtree(p_dowld + main_conf['iptables']['path_com'],
                   ignore_errors=True)
            return True
        return False

    def file_rw(self, p, w=False, data=None, log=None, json_r=False):
        try:
            if log:
                logging.debug('{}. Path: {}'.format(log, p))

            if w:
                f = open(p, 'w')
                if data:
                    if json_r:
                        dump(data, f, indent=4)
                    else:
                        f.writelines(data)
                f.close()
                return True
            else:
                f = open(p, 'r')
                if json_r:
                    if f:
                        data = load(f)
                else:
                    data = f.readlines()
                f.close()
                return data
        except BaseException as rwexpt:
            logging.error('R/W File: {}'.format(rwexpt))
            return False

    def run_service(self, sysctl=False, comm=False, restart=False):

        if comm:
            if restart:
                logging.info('Restart common service')
                self._sys_call('systemctl stop {}'.format(self.f_com),
                               sysctl)
            else:
                logging.info('Run common service')
                self._sys_call('systemctl daemon-reload', sysctl)
                sleep(2)
                self._sys_call('systemctl enable {}'.format(self.f_com),
                               sysctl)
            sleep(2)
            self._sys_call('systemctl start {}'.format(self.f_com), sysctl)
        else:
            if restart:
                logging.info('Restart vpn service')
                self._sys_call('systemctl stop {}'.format(self.f_vpn),
                               sysctl)
            else:
                logging.info('Run vpn service')
                self._sys_call('systemctl enable {}'.format(self.f_vpn),
                               sysctl)
            sleep(2)
            self._sys_call('systemctl start {}'.format(self.f_vpn), sysctl)

    def _sys_call(self, cmd, sysctl=False, rolback=True, s_exit=4):
        resp = Popen(cmd, shell=True, stdout=PIPE,
                     stderr=STDOUT).communicate()
        logging.debug('Sys call cmd: {}. Stdout: {}'.format(cmd, resp))
        if resp[1]:
            logging.debug(resp[1])
            if rolback:
                self._rolback(sysctl, s_exit)
            else:
                return False

        if 'The following packages have unmet dependencies:' in resp[0]:
            if rolback:
                self._rolback(sysctl, s_exit)
            exit(s_exit)

        return resp[0]

    def _upgr_deb_pack(self, v):
        logging.info('Debian: {}'.format(v))

        cmd = 'echo deb http://http.debian.net/debian jessie-backports main ' \
              '> /etc/apt/sources.list.d/jessie-backports.list'
        logging.debug('Add jessie-backports.list')
        self._sys_call(cmd)
        self._sys_call(cmd='apt-get install lshw -y')

        logging.info('Update')
        self._sys_call('apt-get update')
        self.__upgr_sysd(
            cmd='apt-get -t jessie-backports install systemd -y')

        logging.debug('Install systemd-container')
        self._sys_call('apt-get install systemd-container -y')

    def _upgr_ub_pack(self, v):
        logging.info('Ubuntu: {}'.format(v))

        if int(v.split('.')[0]) < 16:
            logging.error('Your version of Ubuntu is lower than 16. '
                          'It is not supported by the program')
            sys.exit(2)

        logging.info('Update')
        self._sys_call('apt-get update')
        logging.debug('Install systemd-container')
        self._sys_call('apt-get install systemd-container -y')

    def __upgr_sysd(self, cmd):
        try:
            raw = self._sys_call('systemd --version')

            ver = raw.split('\n')[0].split(' ')[1]
            logging.debug('systemd --version: {}'.format(ver))

            if int(ver) < 229:
                logging.info('Upgrade systemd')

                raw = self._sys_call(cmd)

                if self.recursion < 1:
                    self.recursion += 1

                    logging.info('Install systemd')
                    logging.debug(self.__upgr_sysd(cmd))
                else:
                    raise BaseException(raw)
                logging.info('Upgrade systemd done')

            logging.info('Systemd version: {}'.format(ver))
            self.recursion = 0

        except BaseException as sysexp:
            logging.error('Get/upgrade systemd ver: {}'.format(sysexp))
            sys.exit(1)

    def _ping_port(self, port, verb=False):
        with closing(
                socket.socket(socket.AF_INET, socket.SOCK_STREAM)) as sock:
            if sock.connect_ex(('0.0.0.0', int(port))) == 0:
                if verb:
                    logging.info("Port {} is open".format(port))
                else:
                    logging.debug("Port {} is open".format(port))
                return True
            else:
                if verb:
                    logging.info("Port {} is not available".format(port))
                else:
                    logging.debug("Port {} is not available".format(port))
                return False

    def _cycle_ask(self, p, status, verb=False):
        logging.debug('Ask port: {}, status: {}'.format(p, status))
        ts = time()
        tw = 350

        if status == 'stop':
            logging.debug('Stop mode')
            while True:
                if not self._ping_port(p, verb):
                    return True
                if time() - ts > tw:
                    return False
                sleep(2)
        else:
            logging.debug('Start mode')
            while True:
                if self._ping_port(p, verb):
                    return True
                if time() - ts > tw:
                    return False
                sleep(2)

    def _checker_port(self, port, status='start', verb=False):
        logging.debug('Checker: {}'.format(status))
        if not port:
            return None
        if isinstance(port, (list, set)):
            resp = list()
            for p in port:
                resp.append(self._cycle_ask(p, status, verb))
            return True if all(resp) else False
        else:
            return self._cycle_ask(port, status, verb)

    def __wait_up(self, sysctl):
        logging.info('Wait run services. This may take 15 minutes. '
                     'Do not turn off !')
        use_port = main_conf['final']
        dupp_conf = self._get_url(main_conf['build']['conf_link'])
        for k, v in dupp_conf.iteritems():
            if isinstance(v, dict) and v.get('Addr'):
                use_port['dapp_port'].append(int(v['Addr'].split(':')[-1]))

        logging.debug('Check ports: {}'.format(use_port))
        if not self._checker_port(port=use_port['vpn_port'], verb=True):
            logging.info('Restart VPN')
            self.run_service(sysctl=sysctl, comm=False, restart=True)
            if not self._checker_port(port=use_port['vpn_port'], verb=True):
                logging.error('VPN is not ready')
                exit(13)

        if not self._checker_port(port=use_port['dapp_port'], verb=True):
            logging.info('Restart Common')
            self.run_service(sysctl=sysctl, comm=True, restart=True)
            if not self._checker_port(port=use_port['dapp_port'], verb=True):
                logging.error('Common is not ready')
                exit(14)

    def _finalizer(self, rw=None, sysctl=False, pass_check=False):
        logging.debug('Finalizer')
        if pass_check:
            return True

        f_path = main_conf['mark_final']
        if not isfile(f_path):
            self.file_rw(p=f_path, w=True, log='First start')
            return True

        if rw:
            self.__wait_up(sysctl)
            dest = 'opt/privatix/config/dappvpn.config.json'
            from_path = '{}{}{}'.format(self.p_dwld, self.path_vpn, dest)
            to_path = '{}{}{}'.format(self.p_dwld, self.path_com, dest)
            copyfile(from_path, to_path)

            self.file_rw(p=f_path, w=True, data=main_conf['final'],
                         log='Finalizer.Write port info', json_r=True)
            return True

        mark = self.file_rw(p=f_path)
        logging.debug('Start marker: {}'.format(mark))
        if not mark:
            logging.info('First start')
            return True

        logging.info('Second start.'
                     'This is protection against restarting the program.'
                     'If you need to re-run the script, '
                     'you need to delete the file {}'.format(f_path))
        return False

    def _byteify(self, data, ignore_dicts=False):
        if isinstance(data, unicode):
            return data.encode('utf-8')
        if isinstance(data, list):
            return [self._byteify(item, ignore_dicts=True) for item in data]
        if isinstance(data, dict) and not ignore_dicts:
            return {
                self._byteify(key, ignore_dicts=True): self._byteify(value,
                                                                     ignore_dicts=True)
                for key, value in data.iteritems()
            }
        return data

    def __json_load_byteified(self, file_handle):
        return self._byteify(
            load(file_handle, object_hook=self._byteify),
            ignore_dicts=True
        )

    def _get_url(self, link, to_json=True):
        resp = urlopen(url=link)
        if to_json:
            return self.__json_load_byteified(resp)
        else:
            return resp.read()

    def build_cmd(self):
        conf = main_conf['build']

        # Get DB params
        json_db = self._get_url(conf['conf_link'])
        db_conf = json_db.get('DB')
        logging.debug('DB params: {}'.format(db_conf))
        if db_conf:
            conf['db_conf'].update(db_conf['Conn'])

        # Get dappctrl_id from prod_data.sql
        if not conf['dappctrl_id']:
            raw_id = self._get_url(link=conf['dappctrl_id_raw'],
                                   to_json=False).split('\n')

            for i in raw_id:
                if conf['field_name_id'] in i:
                    conf['dappctrl_id'] = i.split(conf['field_name_id'])[1]
                    logging.debug('dapp_id: {}'.format(conf['dappctrl_id']))
                    break
            else:
                logging.error(
                    'dappctrl_id not exist: {}'.format(conf['dappctrl_id']))
                sys.exit(20)

        # Get dappvpn.config.json
        templ = self._get_url(link=conf['templ'], to_json=False).replace(
            '\n', '')

        conf['db_conf'] = (sub("'|{|}", "", str(conf['db_conf']))).replace(
            ': ', '=').replace(',', '')

        conf['cmd'] = conf['cmd'].format(templ,
                                         conf['dappvpnconf_path'],
                                         conf['dappconconf_path'],
                                         conf['db_conf'],
                                         conf['dappctrl_id']
                                         )

        logging.debug('Build cmd: {}'.format(conf['cmd']))
        self.file_rw(
            p=self._reletive_path(conf['cmd_path']),
            w=True,
            data=conf['cmd'],
            log='Create file with dapp cmd')


class Params(CMD):
    """ This class provide check sysctl and iptables """

    def check_port(self):
        port = main_conf['iptables']['openvpn_port']
        port = findall('\d\d\d', port)[0]

        if self._ping_port(port=port):
            while True:
                logging.info("Port: {} is busy or wrong."
                             "Select a different port,in range 1 - 65535.".format(
                    port))
                port = raw_input('>')
                try:
                    if int(port) in range(65535)[1:] and not self._ping_port(
                            port=port):
                        break
                except BaseException:
                    pass
        main_conf['final']['vpn_port'] = port
        return port

    def __iptables(self):
        logging.debug('Check iptables')

        cmd = '/sbin/iptables -t nat -L'
        chain = 'Chain POSTROUTING'
        raw = self._sys_call(cmd)
        arr = raw.split('\n\n')
        chain_arr = []
        for i in arr:
            if chain in i:
                chain_arr = i.split('\n')
                break
        del arr

        addr = self.addres(chain_arr)
        infs = self.interfase()
        tun = self.check_tun()
        port = self.check_port()
        logging.debug('Addr,interface,tun: {}'.format((addr, infs, tun)))
        return addr, infs, tun, port

    def check_tun(self):
        def check_tun(i):
            max_tun_index = max([int(x.replace('tun', '')) for x in i])

            logging.info('You have the following interfaces {}. '
                         'Please enter another tun interface.'
                         'For example tun{}.\n'.format(i, max_tun_index + 1))

            new_tun = raw_input('>')
            if new_tun in i or ''.join(findall('[^\d+]',new_tun)) != 'tun':
                logging.info(
                    'Wrong. The interface must called tun\n'
                    'and should be different from: {}\n'.format(
                        i))
                new_tun = check_tun(i)
            return new_tun

        cmd = 'ip link show'
        raw = self._sys_call(cmd)
        tuns = findall("tun\d", raw)
        tun = 'tun1'
        if tuns:
            tun = check_tun(tuns)
        return tun

    def interfase(self):
        def check_interfs(i):
            logging.info('Please enter one of your '
                         'available external interfaces: {}\n'.format(i))

            new_intrfs = raw_input('>')
            if new_intrfs not in i:
                logging.info(
                    'Wrong. Interface must be one of: {}\n'.format(i))
                new_intrfs = check_interfs(i)
            return new_intrfs

        arr_intrfs = []
        cmd = 'LANG=POSIX lshw -C network'
        raw = self._sys_call(cmd)
        arr = raw.split('logical name: ')
        arr.pop(0)
        for i in arr:
            arr_intrfs.append(i.split('\n')[0])
        del arr
        if len(arr_intrfs) > 1:
            intrfs = check_interfs(arr_intrfs)
        else:
            intrfs = arr_intrfs[0]

        return intrfs

    def addres(self, arr):
        def check_addr(p):
            while True:
                addr = raw_input('>')
                match = search(p, addr)
                if not match:
                    logging.info('You addres is wrong,please enter '
                                 'right address.Last octet is always 0.Example: 255.255.255.0\n')
                    addr = check_addr(p)
                break
            return addr

        addr = main_conf['addr']
        for i in arr:
            if main_conf['addr'] + main_conf['mask'][0] in i:
                logging.info(
                    'Addres {} is busy or wrong, please enter new address '
                    'without changing the 4th octet.'
                    'Example: xxx.xxx.xxx.0\n'.format(addr))

                pattern = r'^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}0$'
                addr = check_addr(pattern)
                break
        return addr

    def __sysctl(self):
        """ Return True if ip_forward=1 by default,
        and False if installed by script """
        cmd = '/sbin/sysctl net.ipv4.ip_forward'
        res = self._sys_call(cmd).decode()
        param = int(res.split(' = ')[1])

        if not param:
            if self.recursion < 1:
                logging.debug('Change net.ipv4.ip_forward from 0 to 1')

                cmd = '/sbin/sysctl -w net.ipv4.ip_forward=1'
                self._sys_call(cmd)
                sleep(0.5)
                self.recursion += 1
                self.__sysctl()
                return False
            else:
                logging.error('sysctl net.ipv4.ip_forward didnt change to 1')
                sys.exit(3)
        return True

    def _rw_unit_file(self, ip, intfs, sysctl, code):
        logging.debug('Preparation unit file: {},{}'.format(ip, intfs))
        addr = ip + main_conf['mask'][0]
        try:
            # read a list of lines into data
            tmp_data = self.file_rw(p=self.p_dwld + self.f_vpn)
            logging.debug('Read {}'.format(self.f_vpn))
            # replace all search fields
            for row in tmp_data:

                for param in self.params.keys():
                    if param in row:
                        indx = tmp_data.index(row)

                        if self.params[param]:
                            tmp_data[indx] = self.params[param].format(addr,
                                                                       intfs)
                        else:
                            if sysctl:
                                tmp_data[indx] = ''

            # rewrite unit file
            logging.debug('Rewrite {}'.format(self.f_vpn))
            self.file_rw(p=self.p_dwld + self.f_vpn, w=True, data=tmp_data)
            del tmp_data

            # move unit files
            logging.debug('Move units.')
            copyfile(self.p_dwld + self.f_vpn, self.p_dest + self.f_vpn)
            copyfile(self.p_dwld + self.f_com, self.p_dest + self.f_com)
        except BaseException as f_rw:
            logging.error('R/W unit file: {}'.format(f_rw))
            self._rolback(sysctl, code)

    def revise_params(self):
        sysctl = self.__sysctl()
        ip, intfs, tun, port = self.__iptables()
        return ip, intfs, tun, port, sysctl

    def _rw_openvpn_conf(self, new_ip, new_tun, new_port, sysctl, code):
        # rewrite in /var/lib/container/vpn/etc/openvpn/config/server.conf
        # two fields: server,push "route",  if ip =! default addr.
        conf_file = "{}{}{}".format(main_conf['iptables']['path_download'],
                                    main_conf['iptables']['path_vpn'],
                                    main_conf['iptables']['openvpn_conf'])
        def_ip = main_conf['addr']
        def_mask = main_conf['mask'][1]
        search_fields = main_conf['iptables']['openvpn_fields']
        search_tun = main_conf['iptables']['openvpn_tun']
        search_port = main_conf['iptables']['openvpn_port']
        try:
            # read a list of lines into data
            tmp_data = self.file_rw(
                p=conf_file,
                log='Read openvpn server.conf'
            )

            # replace all search fields
            for row in tmp_data:

                for field in search_fields:
                    if field.format(def_ip, def_mask) in row:
                        indx = tmp_data.index(row)
                        tmp_data[indx] = field.format(new_ip, def_mask) + '\n'

                if search_tun.format('tun') in row:
                    logging.debug(
                        'Rewrite tun interface on: {}'.format(new_tun))
                    indx = tmp_data.index(row)
                    tmp_data[indx] = search_tun.format(new_tun) + '\n'

                if search_port in row:
                    logging.debug('Rewrite port on: {}'.format(new_port))
                    indx = tmp_data.index(row)
                    tmp_data[indx] = 'port {}\n'.format(new_port)

            # rewrite server.conf file
            self.file_rw(
                p=conf_file,
                w=True,
                data=tmp_data,
                log='Rewrite server.conf'
            )

            del tmp_data

            logging.debug('server.conf done')
        except BaseException as f_rw:
            logging.error('R/W server.conf: {}'.format(f_rw))
            self._rolback(sysctl, code)

    def _check_db_run(self, sysctl, code):
        # wait 't_wait' sec until the DB starts, if not started, exit.

        t_start = time()
        t_wait = 300
        mark = True
        logging.info('Waiting for the launch of the DB.')
        while mark:
            logging.debug('Wait.')
            raw = self.file_rw(p=main_conf['build']['db_log'],
                               log='Read DB log')
            for i in raw:
                if main_conf['build']['db_stat'] in i:
                    logging.info('DB was run.')
                    mark = False
                    break
            if time() - t_start > t_wait:
                logging.error(
                    'DB after {} sec does not run.'.format(t_wait))
                self._rolback(sysctl, code)
            sleep(5)

    def _clear_db_log(self):
        self.file_rw(p=main_conf['build']['db_log'],
                     w=True,
                     log='Clear DB log')

    def _run_dapp_cmd(self, sysctl):
        cmds = self.file_rw(
            p=self._reletive_path(main_conf['build']['cmd_path']),
            log='Read dapp cmd')

        if cmds:
            cmds = cmds[0].split('\n')
            for cmd in cmds:
                self._sys_call(cmd=cmd, sysctl=sysctl)
                sleep(1)
        else:
            logging.error('Have not {} file for further execution. '
                          'It is necessary to run the initializer '
                          'in build mode.'.format(
                main_conf['build']['cmd_path']))
            self._rolback(sysctl, 10)

    def _test_mode(self, sysctl):
        data = urlopen(url=main_conf['test']['sql']).read()
        self.file_rw(p=main_conf['test']['path'], w=True, data=data,
                     log='Create file with test sql data.')
        cmd = main_conf['test']['cmd'].format(main_conf['test']['path'])

        self._sys_call(cmd=cmd, sysctl=sysctl, s_exit=12)
        raw_tmpl = self._get_url(main_conf['build']['templ'])
        self.file_rw(p=main_conf['build']['dappvpnconf_path'], w=True,
                     data=raw_tmpl, log='Create file with test sql data.')

    def ip_dappctrl(self):
        """Change ip addr in dappctrl.config.local.json"""
        search_field = main_conf['build']['dappctrl_search_field']
        my_ip = urlopen(url='http://icanhazip.com').read().replace('\n', '')
        p_dap_conf = main_conf['build']['dappctrl_conf_local']

        data = self.file_rw(p=p_dap_conf, json_r=True,
                            log='Read dappctrl.config.local.json.')
        raw = data[search_field].split(':')

        raw[1] = '//{}'.format(my_ip)
        data[search_field] = ':'.join(raw)

        self.file_rw(p=p_dap_conf, w=True, json_r=True, data=data,
                     log='Rewrite dappctrl.config.local.json.')


class Rdata(CMD):
    def __init__(self):
        CMD.__init__(self)
        self.url = main_conf['iptables']['link_download']
        self.files = main_conf['iptables']['file_download']
        self.p_dwld = main_conf['iptables']['path_download']
        self.p_dest_vpn = main_conf['iptables']['path_vpn']
        self.p_dest_com = main_conf['iptables']['path_com']
        self.p_unpck = dict(vpn=self.p_dest_vpn, common=self.p_dest_com)

    def download(self, sysctl, code):
        try:
            logging.info('Begin download files.')

            if not isdir(self.p_dwld):
                mkdir(self.p_dwld)

            obj = URLopener()
            for f in self.files:
                logging.info('Start download {}.\nWait.This may take some '
                             'long time. Do not turn off !'.format(f))
                obj.retrieve(self.url + f, self.p_dwld + f)
                logging.info('Download {} done.'.format(f))
            return True

        except BaseException as down:
            logging.error('Download {}.'.format(down))
            self._rolback(sysctl, code)

    def unpacking(self, sysctl):
        logging.info('Begin unpacking download files.')
        try:
            for f in self.files:
                if '.tar.xz' == f[-7:]:
                    logging.info('Unpacking {}.'.format(f))
                    for k, v in self.p_unpck.items():
                        if k in f:
                            if not isdir(self.p_dwld + v):
                                mkdir(self.p_dwld + v)
                            cmd = 'tar xpf {} -C {} --numeric-owner'.format(
                                self.p_dwld + f, self.p_dwld + v)
                            self._sys_call(cmd, sysctl)
                            logging.info('Unpacking {} done.'.format(f))
        except BaseException as p_unpck:
            logging.error('Unpack: {}.'.format(p_unpck))

    def clean(self):
        logging.info('Delete downloaded files.')

        for f in self.files:
            logging.info('Delete {}'.format(f))
            remove(self.p_dwld + f)


class GUI(CMD):
    def __init__(self):
        CMD.__init__(self)
        self.gui = main_conf['gui']

        if environ.get('SUDO_USER'):
            self.__icon_file = self.gui['icon_tmpl_f'].format(
                '/home/', environ['SUDO_USER'])

            self.__chown_cmd = self.gui['chown'].format(
                environ['SUDO_USER'], self.__icon_file
            )
        else:
            self.__icon_file = self.gui['icon_tmpl_f'].format(
                '', environ['HOME']
            )
            self.__chown_cmd = self.gui['chown'].format(
                environ['USER'], self.__icon_file
            )

    def __create_icon(self):
        config = ConfigParser()
        config.optionxform = str
        tmpl = self.gui['icon_tmpl']
        section = tmpl['Section']

        logging.debug('Create icon file: {}'.format(self.__icon_file))

        with open(self.__icon_file, 'w') as icon:
            config.add_section(section)
            [config.set(section, k, v) for k, v in tmpl.items()]
            config.write(icon)

        logging.debug('Create icon file done')
        chmod(self.__icon_file,
              stat(self.__icon_file).st_mode | S_IXUSR | S_IXGRP | S_IXOTH)

        logging.debug('Chmod icon file done')
        self._sys_call(self.__chown_cmd)
        logging.debug('Chown icon file done')

    def __get_gui(self, sysctl=False):
        cmds = self.gui['gui_inst']
        for cmd in cmds:
            self._sys_call(cmd, sysctl=sysctl, s_exit=11)
        self.__create_icon()

    def __get_npm(self, sysctl):
        # install npm and nodejs
        logging.debug('Get NPM for GUI.')
        npm_path = self._reletive_path(self.gui['npm_tmp_f'])
        self.file_rw(
            p=npm_path,
            w=True,
            data=urlopen(self.gui['npm_url']),
            log='Download nodesource'
        )

        cmd = self.gui['npm_tmp_file_call'] + npm_path
        self._sys_call(cmd=cmd, sysctl=sysctl, s_exit=11)

        cmd = self.gui['npm_node']
        self._sys_call(cmd=cmd, sysctl=sysctl, s_exit=11)

    def __get_pack_ver(self, sysctl):
        res = False
        for k, v in self.gui['version'].items():
            logging.info('Check {} version.'.format(k))
            cmd = main_conf['search_pack'].format(k)
            raw = self._sys_call(cmd=cmd, sysctl=sysctl)
            if raw:
                res = True
                cmd = '{} -v'.format(k)
                raw = self._sys_call(cmd=cmd, sysctl=sysctl)
                ver = '.'.join(findall('\d+', raw)[0:2])

                if StrictVersion(ver) < StrictVersion(v[0]):
                    self.gui['version'][k][1] = True
                    self.gui['version'][k][2] = ver

            else:
                logging.info('{} not installed yet.'.format(k))
        return res

    def __check_version(self, sysctl):
        self.__get_pack_ver(sysctl)

        logging.debug('Check dependencies.')

        if any([x[1] for x in self.gui['version'].values()]):
            logging.info('\n\nYou have installed obsolete packages.\n'
                         'To continue the installation, '
                         'you need to update the following packages:')
            for k, v in self.gui['version'].items():
                if v[1]:
                    logging.info(
                        ' - {} {}. Min requirements: {}'.format(k, v[2],
                                                                v[0]))

            answ = raw_input('\n\nDo you want to re-install the packages '
                             'yourself or in automatic mode?\n '
                             'Y - automatically, N - by yourself.\n'
                             'If you choose N, the installation is interrupted, '
                             'and you will need to run it again.\n'
                             '\n> ')

            while True:
                if answ.lower() not in ['n', 'y']:
                    logging.info('Invalid choice. Select y or n.')
                    answ = raw_input('> ')
                    continue
                if answ.lower() == 'y':
                    logging.info('You have selected automatic mode')
                    break
                else:
                    logging.info('You have chosen manual mode')
                    self._rolback(sysctl, 15)
                    break

            for k, v in self.gui['version'].items():
                if v[1]:
                    logging.info('Preparing for deletion '
                                 '{} {}'.format(k, v[2]))
                    cmd = main_conf['del_pack'].format(k)
                    self._sys_call(cmd=cmd, sysctl=sysctl)

            if self.__get_pack_ver(sysctl):
                logging.info('The problem with deleting one of the listed '
                             'packages. Try to delete in manual mode '
                             'and repeat the process again.')
                self._rolback(sysctl, 16)

    def install_gui(self, sysctl):
        logging.debug('Install GUI.')
        self.__check_version(sysctl)
        self.__get_npm(sysctl)
        self.__get_gui(sysctl)

    def _clear_gui(self):
        logging.info('Clear GUI.')
        p = self.gui['gui_path']
        logging.debug('Clear: {}'.format(p))
        rmtree(p, ignore_errors=True)
        mkdir(p)

    def update_gui(self):
        logging.info('Update GUI.')
        self._clear_gui()
        self.__get_gui()


class Nspawn():
    pass


class LXC():
    pass


class Checker(Params, Rdata, GUI):
    def __init__(self):
        GUI.__init__(self)
        Rdata.__init__(self)
        Params.__init__(self)
        self.task = dict(ubuntu=self._upgr_ub_pack,
                         debian=self._upgr_deb_pack
                         )

    def init_os(self, args, pass_check=False):
        if self._finalizer(pass_check=pass_check):
            dist_name, ver, name_ver = linux_distribution()
            upgr_pack = self.task.get(dist_name.lower(), False)
            if not upgr_pack:
                logging.error('You system is {}.'
                              'She is not supported yet'.format(dist_name))
                sys.exit(19)
            upgr_pack(ver)
            ip, intfs, tun, port, sysctl = self.revise_params()
            self.download(sysctl, 6)
            try:
                self.unpacking(sysctl)
                self._rw_openvpn_conf(ip, tun, port, sysctl, 7)
                self._rw_unit_file(ip, intfs, sysctl, 5)
                self.clean()
                self._clear_db_log()
                self.ip_dappctrl()
                self.run_service(sysctl, comm=True)
                self._check_db_run(sysctl, 9)
                if not args['test']:
                    logging.info('Test mode.')
                    self._test_mode(sysctl)
                else:
                    logging.info('Full mode.')
                    self._run_dapp_cmd(sysctl)

                self.run_service(sysctl)
                if args['no_gui']:
                    logging.info('GUI mode.')
                    if pass_check:
                        self.update_gui()
                    else:
                        self.install_gui(sysctl)

                self._finalizer(rw=True, sysctl=sysctl)
            except BaseException as mexpt:
                logging.error('Main trouble: {}'.format(mexpt))
                self._rolback(sysctl, 17)

    def sudo_prompt(self):
        mess = 'Make sure that the user from which you are installing\n' \
               'is added to the sudo file with parameters\n' \
               'ALL = (ALL: ALL) NOPASSWD: ALL\n' \
               'or this user himself is a local root\n\n' \
               'Y - I understand. Everything is fine\n' \
               'N - The user does not root, stop the installation.'
        logging.info(mess)

        answ = raw_input('>')

        while True:
            if answ.lower() not in ['n', 'y']:
                logging.info('Invalid choice. Select Y or N.')
                answ = raw_input('> ')
                continue
            if answ.lower() == 'n':
                sys.exit(21)
            else:
                break


if __name__ == '__main__':
    parser = ArgumentParser(description=' *** Installer *** ')
    parser.add_argument("--build", nargs='?', default=True,
                        help='')

    parser.add_argument("--update", nargs='?', default=True,
                        help='Update containers and rerun initializer')

    parser.add_argument("--update-gui", nargs='?', default=True,
                        help='Update GUI')

    parser.add_argument("--mass-update", nargs='?', default=True,
                        help='Update All')

    parser.add_argument('--vpn', type=str, default=False,
                        help='[start,stop,restart,status]')

    parser.add_argument('--comm', type=str, default=False,
                        help='[start,stop,restart,status]')

    parser.add_argument('--mass', type=str, default=False,
                        help='[start,stop,restart,status]')

    parser.add_argument("--test", nargs='?', default=True,
                        help='')

    parser.add_argument("--no-gui", nargs='?', default=True,
                        help='Full install without GUI')
    args = vars(parser.parse_args())

    check = Checker()

    if isfile(main_conf['mark_final']):
        raw = check.file_rw(p=main_conf['mark_final'],
                            json_r=True,
                            log='Search port in finalizer.pid')
        if raw:
            main_conf['ports'].update(raw)

    logging.debug('Input args: {}'.format(args))
    logging.debug('Inside finalizer.pid: {}'.format(main_conf['ports']))

    signal(SIGINT, check.signal_handler)

    if not args['build']:
        logging.info('Build mode.')
        check.build_cmd()

    elif args['vpn']:
        logging.debug('Vpn mode.')
        sys.stdout.write(
            str(check.service('vpn', args['vpn'],
                              main_conf['ports']['vpn_port'])))

    elif args['comm']:
        logging.debug('Comm mode.')
        sys.stdout.write(
            str(check.service('comm', args['comm'],
                              main_conf['ports']['dapp_port'])))

    elif args['mass']:
        logging.debug('Mass mode.')
        comm_stat = check.service('comm', args['mass'],
                                  main_conf['ports']['dapp_port'])
        vpn_stat = check.service('vpn', args['mass'],
                                 main_conf['ports']['vpn_port'])
        sys.stdout.write(str(bool(all((comm_stat, vpn_stat)))))

    elif not args['update']:
        logging.info('Update containers mode.')
        check.sudo_prompt()
        if check.clear_contr():
            args['no_gui'] = False
            check.init_os(args, True)
        else:
            logging.info('Problem with clear all old file.')

    elif not args['update_gui']:
        logging.info('Update GUI mode.')
        check.sudo_prompt()
        check.update_gui()

    elif not args['mass_update']:
        logging.info('Update All mode.')
        check.sudo_prompt()
        if check.clear_contr():
            check.init_os(args, True)
        else:
            logging.info('Problem with clear all old file.')

    else:
        logging.info('Begin init.')
        check.sudo_prompt()
        check.init_os(args)
        logging.info('All done.')
