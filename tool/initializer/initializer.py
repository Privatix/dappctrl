#!/usr/bin/python
# -*- coding: utf-8 -*-

"""
    Initializer on pure Python 2.7
    mode:
    python initializer.py  -h                        get help information
    python initializer.py                            start full install
    python initializer.py --build                    create cmd for dapp
    python initializer.py --vpn start                start vpn servise
    python initializer.py --comm stop                stop common servise
    python initializer.py --test                     start in test mode
    python initializer.py --no-gui                   install without GUI
"""

import sys
import logging
import socket
import signal
from contextlib import closing
from re import search, sub, findall
from codecs import open
from os import remove, mkdir, path
from json import load, dump
from time import time, sleep
from urllib import URLopener
from urllib2 import urlopen
from os.path import isfile, isdir
from argparse import ArgumentParser
from platform import linux_distribution
from subprocess import Popen, PIPE, STDOUT
from shutil import copyfile

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
"""

log_conf = dict(
    filename='initializer.log',
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
            'ExecStop=/sbin/sysctl': False,
            'ExecStopPost=/sbin/sysctl': False,

            'ExecStop=/sbin/iptables': 'ExecStop=/sbin/iptables -t nat -A POSTROUTING -s {} -o {} -j MASQUERADE\n',
            'ExecStartPre=/sbin/iptables': 'ExecStartPre=/sbin/iptables -t nat -A POSTROUTING -s {} -o {} -j MASQUERADE\n',
            'ExecStopPost=/sbin/iptables': 'ExecStopPost=/sbin/iptables -t nat -A POSTROUTING -s {} -o {} -j MASQUERADE\n',
        }

    ),

    build={
        'cmd': '/opt/privatix/initializer/dappinst -dappvpnconftpl=\'{}\' -dappvpnconf={} -connstr=\"{}\"',
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
        'conf_link': 'https://raw.githubusercontent.com/Privatix/dappctrl/release/0.6.0/dappctrl.config.json',
        'templ': 'https://raw.githubusercontent.com/Privatix/dappctrl/release/0.6.0/svc/dappvpn/dappvpn.config.json',
        'dappctrl_conf_local': '/var/lib/container/common/opt/privatix/config/dappctrl.config.local.json',
        'dappctrl_search_field': 'PayAddress',
    },
    final={'dapp_port': [], 'vpn_port': 443},
    gui={
        'npm_tmp_file': 'tmp_nodesource',
        'npm_url': 'https://deb.nodesource.com/setup_9.x',
        'npm_tmp_file_call': 'sudo -E bash ',
        'npm_inst': [
            'sudo apt-get install -y nodejs',
            # 'sudo apt-get install -y npm',
            # 'sudo npm install dappctrlgui'
        ],
    },
    test={
        'path': 'test_data.sql',
        'sql': 'https://raw.githubusercontent.com/Privatix/dappctrl/develop/data/test_data.sql',
        'cmd': 'psql -d dappctrl -h 127.0.0.1 -P 5433 -f {}'
    },
    addr='10.217.3.0',
    mask=['/24', '255.255.255.0'],
    mark_final='/var/run/installer.pid',
)


class CMD:
    recursion = 0

    def _reletive_path(self, name):
        dirname = path.dirname(__file__)
        return path.join(dirname, name)

    def _rolback(self, sysctl, code):

        # Rolback net.ipv4.ip_forward
        if not sysctl:
            logging.info('Rolback ip_forward')
            cmd = '/sbin/sysctl -w net.ipv4.ip_forward=0'
            self._sys_call(cmd)
        sys.exit(code)

    def _file_rw(self, p, w=False, data=None, log=None, json_r=False):
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
            else:
                f = open(p, 'r')
                if json_r:
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
                self._sys_call('systemctl stop {}'.format(self.f_com), sysctl)
            else:
                logging.info('Run common service')
                self._sys_call('systemctl daemon-reload', sysctl)
                sleep(2)
                self._sys_call('systemctl enable {}'.format(self.f_com), sysctl)
            sleep(2)
            self._sys_call('systemctl start {}'.format(self.f_com), sysctl)
        else:
            if restart:
                logging.info('Restart vpn service')
                self._sys_call('systemctl stop {}'.format(self.f_vpn), sysctl)
            else:
                logging.info('Run vpn service')
                self._sys_call('systemctl enable {}'.format(self.f_vpn), sysctl)
            sleep(2)
            self._sys_call('systemctl start {}'.format(self.f_vpn), sysctl)

    def _sys_call(self, cmd, sysctl=False, rolback=True, s_exit=4):
        resp = Popen(cmd, shell=True, stdout=PIPE,
                     stderr=STDOUT).communicate()
        logging.debug('Sys call cmd: {}. Stdout: {}'.format(cmd, resp))
        if resp[1]:
            logging.error(resp[1])
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

    def _ping_port(self, port):
        with closing(
            socket.socket(socket.AF_INET, socket.SOCK_STREAM)) as sock:
            if sock.connect_ex(('0.0.0.0', int(port))) == 0:
                logging.debug("Port is busy")
                return True
            else:
                logging.debug("Port is free")
                return False

    def __checker_port(self, p):
        ts = time()
        tw = 180
        while True:
            if self._ping_port(p):
                return True
            if time() - ts > tw:
                return False
            sleep(2)

    def __wait_up(self, sysctl):
        logging.debug('Wait run services')
        use_port = main_conf['final']
        dupp_conf = self._get_url(main_conf['build']['conf_link'])
        for k, v in dupp_conf.iteritems():
            if isinstance(v, dict) and v.get('Addr'):
                use_port['dapp_port'].append(int(v['Addr'].split(':')[-1]))

        if not self.__checker_port(use_port['vpn_port'],'VPN'):
            logging.info('Restart VPN')
            self.run_service(sysctl=sysctl,comm=False,restart=True)
            if not self.__checker_port(use_port['vpn_port']):
                logging.error('VPN is not ready')
                exit(13)

        for port in dupp_conf:
            if not self.__checker_port(port):
                logging.info('Restart Common')
                self.run_service(sysctl=sysctl, comm=True, restart=True)
                if not self.__checker_port(use_port['vpn_port'], 'VPN'):
                    logging.error('Common is not ready')
                    exit(14)

    def _finalizer(self, rw=None, sysctl=False):
        f_path = main_conf['mark_final']
        if not isfile(f_path):
            self._file_rw(p=f_path, w=True, log='First start')
            return True
        if rw:
            self.__wait_up(sysctl)
            self._file_rw(p=f_path, w=True, data='1')
            return True
        mark = self._file_rw(p=f_path)
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

        json_db = self._get_url(conf['conf_link'])
        db_conf = json_db.get('DB')
        if db_conf:
            conf['db_conf'].update(db_conf['Conn'])

        # templ = str(self._get_url(conf['templ'])).replace('\'', '"')
        templ = self._get_url(link=conf['templ'], to_json=False).replace(
            '\n', '')

        conf['db_conf'] = (sub("'|{|}", "", str(conf['db_conf']))).replace(
            ': ', '=').replace(',', '')

        conf['cmd'] = conf['cmd'].format(templ, conf['dappvpnconf_path'],
                                         conf['db_conf'])
        logging.debug('Build cmd: {}'.format(conf['cmd']))
        self._file_rw(
            p=self._reletive_path(conf['cmd_path']),
            w=True,
            data=conf['cmd'],
            log='Create file with dapp cmd')


class Params(CMD):
    """ This class provide check sysctl and iptables """

    def __init__(self):
        self.f_vpn = main_conf['iptables']['unit_vpn']
        self.f_com = main_conf['iptables']['unit_com']
        self.p_dest = main_conf['iptables']['path_unit']
        self.p_dwld = main_conf['iptables']['path_download']
        self.params = main_conf['iptables']['unit_field']

    def service(self, srv, status):
        if status not in ['start', 'stop', 'restart']:
            logging.error('{} status must be in '
                          '[\'start\',\'stop\',\'restart\']'.format(srv))
            return False
        cmd = 'systemctl {} {}'.format(status, self.f_vpn)
        return bool(self._sys_call(cmd, rolback=False))

    def get_npm(self, sysctl):
        npm_path = self._reletive_path(main_conf['gui']['npm_tmp_file'])
        self._file_rw(
            p=npm_path,
            w=True,
            data=urlopen(main_conf['gui']['npm_url']),
            log='Download nodesource'
        )

        cmd = main_conf['gui']['npm_tmp_file_call'] + npm_path
        self._sys_call(cmd=cmd, sysctl=sysctl, s_exit=11)

        cmds = main_conf['gui']['npm_inst']
        for cmd in cmds:
            self._sys_call(cmd, sysctl=sysctl, s_exit=11)

    def check_port(self):
        port = main_conf['iptables']['openvpn_port']
        port = findall('\d\d\d', port)[0]

        if self._ping_port(port=port):
            while True:
                logging.info("Port: {} is busy."
                             "Select a different port.".format(port))
                port = raw_input('>')
                if not self._ping_port(port=port):
                    break
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
            logging.info('Please enter one of your '
                         'available tun interfaces: {}'.format(i))

            new_tun = raw_input('>')
            if new_tun not in i:
                logging.info('Wrong. Interface must be one of: {}'.format(i))
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
                         'available interfaces: {}'.format(i))

            new_intrfs = raw_input('>')
            if new_intrfs not in i:
                logging.info('Wrong. Interface must be one of: {}'.format(i))
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
                                 'right address.Example: 255.255.255.255')
                    addr = check_addr(p)
                break
            return addr

        addr = main_conf['addr']
        for i in arr:
            if main_conf['addr'] + main_conf['mask'][0] in i:
                logging.info('Addres {} is busy,'
                             'please enter free address.'.format(addr))

                pattern = r'^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$'
                # pattern = r'^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\/(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$'
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
            tmp_data = self._file_rw(p=self.p_dwld + self.f_vpn)
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
            self._file_rw(p=self.p_dwld + self.f_vpn, w=True, data=tmp_data)
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
            tmp_data = self._file_rw(
                p=conf_file,
                log='Read openvpn server.conf'
            )

            # replace all search fields
            for row in tmp_data:

                for field in [f for f in search_fields]:
                    if field.format(def_ip, def_mask) in row:
                        indx = tmp_data.index(row)
                        tmp_data[indx] = field.format(new_ip, def_mask)

                if search_tun.format('tun') in row:
                    logging.debug(
                        'Rewrite tun interface on: {}'.format(new_tun))
                    indx = tmp_data.index(row)
                    tmp_data[indx] = search_tun.format(new_tun)

                if search_port in row:
                    logging.debug('Rewrite port on: {}'.format(new_port))
                    indx = tmp_data.index(row)
                    tmp_data[indx] = 'port {}'.format(new_port)


            # rewrite server.conf file
            self._file_rw(
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
            raw = self._file_rw(p=main_conf['build']['db_log'],
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
            sleep(4)

    def _clear_db_log(self):
        self._file_rw(p=main_conf['build']['db_log'],
                      w=True,
                      log='Clear DB log')

    def _run_dapp_cmd(self, sysctl):
        cmd = self._file_rw(
            p=self._reletive_path(main_conf['build']['cmd_path']),
            log='Read dapp cmd')

        if cmd:
            self._sys_call(cmd=cmd[0], sysctl=sysctl)
            sleep(1)
        else:
            self._rolback(sysctl, 10)

    def _test_mode(self, sysctl):
        data = urlopen(url=main_conf['test']['sql']).read()
        self._file_rw(p=main_conf['test']['path'], w=True, data=data,
                      log='Create file with test sql data.')
        cmd = main_conf['test']['cmd'].format(main_conf['test']['path'])

        self._sys_call(cmd=cmd, sysctl=sysctl, s_exit=12)
        raw_tmpl = self._get_url(main_conf['build']['templ'])
        self._file_rw(p=main_conf['build']['dappvpnconf_path'], w=True,
                      data=raw_tmpl, log='Create file with test sql data.')

    def ip_dappctrl(self):
        """Change ip addr in dappctrl.config.local.json"""
        search_field = main_conf['build']['dappctrl_search_field']
        my_ip = urlopen(url='http://icanhazip.com').read().replace('\n', '')
        path = main_conf['build']['dappctrl_conf_local']

        data = self._file_rw(p=path, json_r=True,
                             log='Read dappctrl.config.local.json.')
        raw = data[search_field].split(':')

        raw[1] = '//{}'.format(my_ip)
        data[search_field] = ':'.join(raw)

        self._file_rw(p=path, w=True, json_r=True, data=data,
                      log='Rewrite dappctrl.config.local.json.')


class Rdata(CMD):
    def __init__(self):
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
                logging.info('Start download {}.'.format(f))
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


class Checker(Params, Rdata):
    def __init__(self):
        Rdata.__init__(self)
        Params.__init__(self)
        self.task = dict(ubuntu=self._upgr_ub_pack,
                         debian=self._upgr_deb_pack
                         )

    def init_os(self, args):
        if self._finalizer():
            dist_name, ver, name_ver = linux_distribution()
            upgr_pack = self.task.get(dist_name.lower(), False)
            if not upgr_pack:
                logging.error('You system is {}.'
                              'She is not supported yet'.format(dist_name))
            upgr_pack(ver)
            ip, intfs, tun, port, sysctl = self.revise_params()
            self.download(sysctl, 6)
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
                logging.info('Install GUI.')
                self.get_npm(sysctl)
            self._finalizer(True, sysctl)


if __name__ == '__main__':
    parser = ArgumentParser(description=' *** Installer *** ')
    parser.add_argument("--build", nargs='?', default=True,
                        help='')
    parser.add_argument('--vpn', type=str, default=True,
                        help='vpn status [start,stop,restart]')
    parser.add_argument('--comm', type=str, default=True,
                        help='comm status [start,stop,restart]')

    parser.add_argument("--test", nargs='?', default=True,
                        help='')

    parser.add_argument("--no-gui", nargs='?', default=True,
                        help='')

    args = vars(parser.parse_args())
    if not args['build']:
        logging.info('Build mode.')
        CMD().build_cmd()
    elif not args['vpn']:
        logging.info('Vpn mode.')
        Params().service('vpn', args['vpn'])
    elif not args['comm']:
        logging.info('Comm mode.')
        Params().service('comm', args['comm'])
    else:
        logging.info('Begin init.')
        Checker().init_os(args)
        logging.info('All done.')
