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
    python initializer.py --update-back                    update all contaiter without GUI
    python initializer.py --update-mass                    update all contaiter with GUI
    python initializer.py --update-gui                     update only GUI
    python initializer.py --link                           use another link for download.if not use, def link in main_conf[link_download]
    python initializer.py --branch                         use another branch than 'develop' for download. template https://raw.githubusercontent.com/Privatix/dappctrl/{ branch }/
"""

import sys
import logging
import socket
from signal import SIGINT, signal, pause
from contextlib import closing
from re import search, sub, findall, compile, match, IGNORECASE
from codecs import open
# from threading import Thread
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
    22 - Problem with read dappctrl.config.json
    23 - Problem with read dappvpn.config.json
    24 - Attempt to install gui on a system without gui
    25 - Problem with R/W dappctrlgui/settings.json
    26 - Problem with download dappctrlgui.tar.xz
    27 - The dappctrlgui package is not installed correctly.Missing file settings.json
    28 - Absent .dapp_cmd file

"""

logging.getLogger().setLevel('DEBUG')
form_console = logging.Formatter(
    '%(message)s',
    datefmt='%m/%d %H:%M:%S')

form_file = logging.Formatter(
    '%(levelname)7s [%(lineno)3s] %(message)s',
    datefmt='%m/%d %H:%M:%S')

fh = logging.FileHandler('/var/log/initializer.log')  # file debug
fh.setLevel('DEBUG')
fh.setFormatter(form_file)
logging.getLogger().addHandler(fh)

ch = logging.StreamHandler()  # console debug
ch.setLevel('INFO')
ch.setFormatter(form_console)
logging.getLogger().addHandler(ch)



main_conf = dict(
    branch='develop',
    link_download='http://art.privatix.net/',
    dappctrl_dev_conf_json='https://raw.githubusercontent.com/Privatix/dappctrl/{}/dappctrl-dev.config.json',

    back=dict(
        file_download=[
            'vpn.tar.xz',
            'common.tar.xz',
            'systemd-nspawn@vpn.service',
            'systemd-nspawn@common.service'],
        path_download='/var/lib/container/',
        path_vpn='vpn/',
        path_com='common/',
        path_unit='/lib/systemd/system/',
        openvpn_conf='etc/openvpn/config/server.conf',
        openvpn_fields=[
            'server {} {}',
            'push "route {} {}"'
        ],
        openvpn_tun='dev {}',
        openvpn_port=['port 443', 'management'],

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
        'conf_link': 'https://raw.githubusercontent.com/Privatix/dappctrl/{}/dappctrl.config.json',
        'templ': 'https://raw.githubusercontent.com/Privatix/dappctrl/{}/svc/dappvpn/dappvpn.config.json',
        'dappctrl_conf_local': '/var/lib/container/common/opt/privatix/config/dappctrl.config.local.json',
        'dappctrl_id_raw': 'https://raw.githubusercontent.com/Privatix/dappctrl/{}/data/prod_data.sql',
        'field_name_id': '--templateid = ',
        'dappctrl_id': None,
    },

    gui={
        'gui_arch': 'dappctrlgui.tar.xz',
        'gui_path': '/opt/privatix/gui/',
        'link_dev_gui': 'dappctrlgui/',
        'icon_name': 'privatix-dappgui.desktop',
        'icon_sh': 'privatix-dappgui.sh',
        'icon_dir': '{}{}/Desktop/',
        'icon_tmpl_f_sh': '{}{}/{}',
        'icon_tmpl': {
            'Section': 'Desktop Entry',
            'Comment': 'First Internet Broadband Marketplace powered by P2P VPN Network on Blockchain',
            'Terminal': 'false',
            'Name': 'Privatix Dapp',
            'Exec': 'sh -c "sudo /opt/privatix/initializer/initializer.py --mass start && sudo npm start --prefix /opt/privatix/gui/{}"',
            'Type': 'Application',
            'Icon': '/opt/privatix/gui/{}icon_64.png',
        },

        'icon_prod': 'node_modules/dappctrlgui/',
        'dappctrlgui': '/opt/privatix/gui/node_modules/dappctrlgui/settings.json',

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

    dnsmasq={
        'conf': '/etc/NetworkManager/NetworkManager.conf',
        'section': ['main', 'dns', 'dnsmasq'],
        'disable': 'sudo sed -i \'s/^dns=dnsmasq/#&/\' /etc/NetworkManager/NetworkManager.conf && '
                   'sudo service network-manager restart',
    },

    addr='10.217.3.0',
    mask=['/24', '255.255.255.0'],
    mark_final='/var/run/installer.pid',
    wait_mess='{}.Please wait until completed.\n It may take about 5-10 minutes.\n Do not turn it off.',
    ports=dict(vpn_port=[], comm_port=[], mangmt=dict(vpn=None, com=None)),
    tmp_var=None
)


class Init:
    recursion = 0
    target = None  # may will be back,gui,both
    sysctl = False
    waiting = True

    def __init__(self):
        self.url_dwnld = main_conf['link_download']
        self.p_dapctrl_dev_conf = main_conf['dappctrl_dev_conf_json'].format(main_conf['branch'])

        self.f_vpn = main_conf['back']['unit_vpn']
        self.f_com = main_conf['back']['unit_com']
        self.p_dest = main_conf['back']['path_unit']
        self.p_dwld = main_conf['back']['path_download']
        self.params = main_conf['back']['unit_field']
        self.path_vpn = main_conf['back']['path_vpn']
        self.path_com = main_conf['back']['path_com']
        self.ovpn_port = main_conf['back']['openvpn_port']
        self.ovpn_conf = main_conf['back']['openvpn_conf']
        self.ovpn_fields = main_conf['back']['openvpn_fields']
        self.ovpn_tun = main_conf['back']['openvpn_tun']
        self.f_dwnld = main_conf['back']['file_download']

        self.dupp_conf_url = main_conf['build']['conf_link'].format(
            main_conf['branch'])
        self.dupp_vpn_templ = main_conf['build']['templ'].format(
            main_conf['branch'])
        self.dupp_raw_id = main_conf['build']['dappctrl_id_raw'].format(
            main_conf['branch'])
        self.p_dap_conf = main_conf['build'][
            'dappctrl_conf_local']  # take ip and port for ping [3000,8000,9000]
        self.use_ports = main_conf['ports']  # store all need ports

        self.dappvpnconf = main_conf['build']['dappvpnconf_path']
        self.dappconconf = main_conf['build']['dappconconf_path']
        self.wait_mess = main_conf['wait_mess']

        self.gui_installer = main_conf['gui']['gui_inst']
        self.gui_path = main_conf['gui']['gui_path']
        self.gui_version = main_conf['gui']['version']
        self.gui_icon_name = main_conf['gui']['icon_name']
        self.gui_icon_sh = main_conf['gui']['icon_sh']
        self.gui_icon_path = main_conf['gui']['icon_dir']
        self.gui_icon_path_sh = main_conf['gui']['icon_tmpl_f_sh']
        self.gui_icon_tmpl = main_conf['gui']['icon_tmpl']
        self.gui_icon_prod = main_conf['gui']['icon_prod']
        self.gui_icon_chown = main_conf['gui']['chown']
        self.gui_npm_tmp_f = main_conf['gui']['npm_tmp_f']
        self.gui_npm_url = main_conf['gui']['npm_url']
        self.gui_npm_node = main_conf['gui']['npm_node']
        self.gui_arch = main_conf['gui']['gui_arch']
        self.gui_npm_cmd_call = main_conf['gui']['npm_tmp_file_call']
        self.gui_dev_link = main_conf['gui']['link_dev_gui']

        self.dappctrlgui = main_conf['gui']['dappctrlgui']

        self.dns_conf = main_conf['dnsmasq']['conf']
        self.dns_sect = main_conf['dnsmasq']['section']
        self.dns_disable = main_conf['dnsmasq']['disable']

        self.tmp_var = main_conf['tmp_var']
        self.fin_file = main_conf['mark_final']

    def re_init(self):
        self.__init__()


class CMD(Init):
    def __init__(self):
        Init.__init__(self)

    def _reletive_path(self, name):
        dirname = path.dirname(__file__)
        logging.debug('Dir name: {}'.format(dirname))
        return path.join(dirname, name)

    def signal_handler(self, sign, frm):
        logging.info('You pressed Ctrl+C!')
        self._rolback(code=18)
        pause()

    def _clear_dir(self, p):
        logging.debug('Clear dir: {}'.format(p))
        cmd = main_conf['del_dirs'].format(p)
        self._sys_call(cmd)

    def long_waiting(self):
        logging.debug('Long waiting: {}'.format(self.waiting))
        while self.waiting:
            print '*',
            sleep(0.1)
        logging.info('\n')
        sleep(0.01)
        self.waiting = True

    def _rolback(self, code):
        # Rolback net.ipv4.ip_forward and clear store by target
        logging.debug('Rolback target: {}, sysctl: {}'.format(self.target,
                                                              self.sysctl))
        if not self.sysctl:
            logging.debug('Rolback ip_forward')
            cmd = '/sbin/sysctl -w net.ipv4.ip_forward=0'
            self._sys_call(cmd)

        if self.target == 'back':
            self.clear_contr(pass_check=True)

        elif self.target == 'gui':
            self._clear_dir(main_conf['gui']['gui_path'])

        elif self.target == 'both':
            self.clear_contr(pass_check=True)
            self._clear_dir(main_conf['gui']['gui_path'])
        else:
            logging.debug('Absent `target` for cleaning!')

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
            logging.info('\n\n   --- Attention! ---\n'
                         ' During installation a failure occurred'
                         ' or you pressed Ctrl+C\n'
                         ' All installed will be removed and returned to'
                         ' the initial state.\n Wait for the end!\n '
                         ' And try again.\n   ------------------\n')
        self.service('vpn', 'stop', self.use_ports['vpn_port'])
        self.service('comm', 'stop', self.use_ports['comm_port'])
        sleep(3)

        if pass_check or not self.service('vpn', 'status',
                                          self.use_ports['vpn_port'],
                                          True) and \
                not self.service('comm', 'status',
                                 self.use_ports['comm_port'], True):
            self._clear_dir(self.p_dwld)
            # rmtree(p_dowld + main_conf['back']['path_vpn'],
            #        ignore_errors=True)
            # rmtree(p_dowld + main_conf['back']['path_com'],
            #        ignore_errors=True)
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

    def run_service(self, comm=False, restart=False):

        if comm:
            if restart:
                logging.info('Restart common service')
                self._sys_call('systemctl stop {}'.format(self.f_com),
                               self.sysctl)
            else:
                logging.info('Run common service')
                self._sys_call('systemctl daemon-reload', self.sysctl)
                sleep(2)
                self._sys_call('systemctl enable {}'.format(self.f_com),
                               self.sysctl)
            sleep(2)
            self._sys_call('systemctl start {}'.format(self.f_com),
                           self.sysctl)
        else:
            if restart:
                logging.info('Restart vpn service')
                self._sys_call('systemctl stop {}'.format(self.f_vpn),
                               self.sysctl)
            else:
                logging.info('Run vpn service')
                self._sys_call('systemctl enable {}'.format(self.f_vpn),
                               self.sysctl)
            sleep(2)
            self._sys_call('systemctl start {}'.format(self.f_vpn),
                           self.sysctl)

    def _sys_call(self, cmd, rolback=True, s_exit=4):
        resp = Popen(cmd, shell=True, stdout=PIPE,
                     stderr=STDOUT).communicate()
        logging.debug('Sys call cmd: {}. Stdout: {}'.format(cmd, resp))
        if resp[1]:
            logging.debug(resp[1])
            if rolback:
                self._rolback(s_exit)
            else:
                return False

        elif 'The following packages have unmet dependencies:' in resp[0]:
            if rolback:
                self._rolback(s_exit)
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

    def __disable_dns(self):
        logging.debug('Disable dnsmasq')
        if isfile(self.dns_conf):
            logging.debug('dnsmasq conf exist')
            cfg = ConfigParser()
            cfg.read(self.dns_conf)
            if cfg.has_option(self.dns_sect[0], self.dns_sect[1]) and \
                            cfg.get(self.dns_sect[0], self.dns_sect[1]) == \
                            self.dns_sect[2]:
                logging.debug('Section {}={} found.'.format(self.dns_sect[1],
                                                            self.dns_sect[
                                                                2]))

                logging.debug('Disable dnsmasq !')
                self._sys_call(self.dns_disable, rolback=False)
            else:
                logging.debug(
                    'dnsmasq conf has not {}'.format(self.dns_sect[0:2]))
        else:
            logging.debug('dnsmasq conf not exist')

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
        self.__disable_dns()

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

    def __all_use_ports(self, d):
        for k, v in d.iteritems():
            if v is None:
                continue
            elif isinstance(v, dict):
                self.__all_use_ports(v)
            elif isinstance(v, list):
                self.tmp_var += map(int, v)
            else:
                self.tmp_var.append(int(v))

    def check_port(self, port, auto=False):

        if self._ping_port(port=port):
            mark = False
            while True:

                if auto and mark:
                    port = int(port)
                    port += 1
                else:
                    logging.info("\nPort: {} is busy or wrong.\n"
                                 "Select a different port,in range 1 - 65535.".format(
                        port))
                    port = raw_input('>')
                try:
                    self.tmp_var = []
                    self.__all_use_ports(self.use_ports)
                    if int(port) in range(65535)[1:] and not self._ping_port(
                            port=port) and int(port) not in self.tmp_var:
                        break
                except BaseException as bexpm:
                    logging.error('Check port: {}'.format(bexpm))

                finally:
                    mark = True

        return port

    def __wait_up(self):
        logging.info(self.wait_mess.format('Run services'))

        logging.debug('Check ports: {}'.format(self.use_ports))
        if not self._checker_port(port=self.use_ports['vpn_port'],
                                  verb=True):
            logging.info('Restart VPN')
            self.run_service(comm=False, restart=True)
            if not self._checker_port(port=self.use_ports['vpn_port'],
                                      verb=True):
                logging.error('VPN is not ready')
                exit(13)

        if not self._checker_port(port=self.use_ports['comm_port'],
                                  verb=True):
            logging.info('Restart Common')
            self.run_service(comm=True, restart=True)
            if not self._checker_port(port=self.use_ports['comm_port'],
                                      verb=True):
                logging.error('Common is not ready')
                exit(14)

    def _finalizer(self, rw=None, pass_check=False):
        logging.debug('Finalizer')
        if pass_check:
            return True

        if not isfile(self.fin_file):
            self.file_rw(p=self.fin_file, w=True, log='First start')
            return True

        if rw:
            if not args['no_wait']:
                self.__wait_up()
            self.file_rw(p=self.fin_file, w=True, data=self.use_ports,
                         log='Finalizer.Write port info', json_r=True)
            return True

        mark = self.file_rw(p=self.fin_file)
        logging.debug('Start marker: {}'.format(mark))
        if not mark:
            logging.info('First start')
            return True

        logging.info('Second start.'
                     'This is protection against restarting the program.'
                     'If you need to re-run the script, '
                     'you need to delete the file {}'.format(self.fin_file))
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
        json_db = self._get_url(self.dupp_conf_url)
        db_conf = json_db.get('DB')
        logging.debug('DB params: {}'.format(db_conf))
        if db_conf:
            conf['db_conf'].update(db_conf['Conn'])

        # Get dappctrl_id from prod_data.sql
        if not conf['dappctrl_id']:
            raw_id = self._get_url(link=self.dupp_raw_id,
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
        templ = self._get_url(link=self.dupp_vpn_templ,
                              to_json=False).replace(
            '\n', '')

        conf['db_conf'] = (sub("'|{|}", "", str(conf['db_conf']))).replace(
            ': ', '=').replace(',', '')

        conf['cmd'] = conf['cmd'].format(templ,
                                         self.dappvpnconf,
                                         self.dappconconf,
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

        port = self.ovpn_port[0]
        port = findall('\d+', port)[0]

        port = self.check_port(port)
        self.use_ports['vpn_port'] = port
        logging.debug('Addr,interface,tun: {}'.format((addr, infs, tun)))
        return addr, infs, tun, port

    def check_tun(self):
        def check_tun(i):
            max_tun_index = max([int(x.replace('tun', '')) for x in i])

            logging.info('You have the following interfaces {}. '
                         'Please enter another tun interface.'
                         'For example tun{}.\n'.format(i, max_tun_index + 1))

            new_tun = raw_input('>')
            if new_tun in i or ''.join(findall('[^\d+]', new_tun)) != 'tun':
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

    def _rw_unit_file(self, ip, intfs, code):
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
                            if self.sysctl:
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
            self._rolback(code)

    def revise_params(self):
        self.sysctl = self.__sysctl()
        ip, intfs, tun, port = self.__iptables()
        return ip, intfs, tun, port

    def _rw_openvpn_conf(self, new_ip, new_tun, new_port, code):
        # rewrite in /var/lib/container/vpn/etc/openvpn/config/server.conf
        # two fields: server,push "route",  if ip =! default addr.
        conf_file = "{}{}{}".format(self.p_dwld,
                                    self.path_vpn,
                                    self.ovpn_conf)
        def_ip = main_conf['addr']
        def_mask = main_conf['mask'][1]
        try:
            # read a list of lines into data
            tmp_data = self.file_rw(
                p=conf_file,
                log='Read openvpn server.conf'
            )

            # replace all search fields
            for row in tmp_data:
                for field in self.ovpn_fields:
                    if field.format(def_ip, def_mask) in row:
                        indx = tmp_data.index(row)
                        tmp_data[indx] = field.format(new_ip,
                                                      def_mask) + '\n'

                if self.ovpn_tun.format('tun') in row:
                    logging.debug(
                        'Rewrite tun interface on: {}'.format(new_tun))
                    indx = tmp_data.index(row)
                    tmp_data[indx] = self.ovpn_tun.format(new_tun) + '\n'

                elif self.ovpn_port[0] in row:
                    logging.debug('Rewrite port on: {}'.format(new_port))
                    indx = tmp_data.index(row)
                    tmp_data[indx] = 'port {}\n'.format(new_port)

                elif self.ovpn_port[1] in row:
                    # management 127.0.0.1 7505
                    indx = tmp_data.index(row)
                    delim = ' '
                    raw_row = row.split(delim)
                    port = int(raw_row[-1])
                    logging.debug('Raw port: {}'.format(port))

                    self.use_ports['mangmt']['vpn'] = self.check_port(port,
                                                                      True)

                    self.use_ports['mangmt']['com'] = self.check_port(
                        int(self.use_ports['mangmt']['vpn']) + 1, True)

                    raw_row[-1] = '{}\n'.format(
                        self.use_ports['mangmt']['vpn'])
                    tmp_data[indx] = delim.join(raw_row)
            logging.debug('--server.conf')
            logging.debug(tmp_data)

            # rewrite server.conf file
            if not self.file_rw(
                    p=conf_file,
                    w=True,
                    data=tmp_data,
                    log='Rewrite server.conf'
            ):
                self._rolback(7)

            del tmp_data

            logging.debug('server.conf done')
        except BaseException as f_rw:
            logging.error('R/W server.conf: {}'.format(f_rw))
            self._rolback(code)

    def _check_dapp_conf(self):
        for servs, port in self.use_ports['mangmt'].iteritems():

            logging.debug('Dapp {} conf. Port: {}'.format(servs, port))
            if servs == 'vpn':
                p = self.dappvpnconf

            elif servs == 'com':
                p = self.dappconconf

            raw_data = self.file_rw(p=p,
                                    log='Check dapp {} conf'.format(servs),
                                    json_r=True)
            if not raw_data:
                self._rolback(23)
            # "localhost:7505"
            logging.debug('dapp {} conf: {}'.format(servs, raw_data))
            delim = ':'
            raw_tmp = raw_data['Monitor']['Addr'].split(delim)
            raw_tmp[-1] = str(port)
            raw_data['Monitor']['Addr'] = delim.join(raw_tmp)
            logging.debug(
                'Monitor Addr: {}.'.format(raw_data['Monitor']['Addr']))

            if hasattr(self, 'sessServPort'):
                delim = ':'
                raw_tmp = raw_data['Server']['Addr'].split(delim)
                raw_tmp[-1] = str(self.sessServPort)
                raw_data['Server']['Addr'] = delim.join(raw_tmp)
                logging.debug(
                    'Server Addr: {}.'.format(raw_data['Server']['Addr']))

            self.file_rw(p=p,
                         log='Rewrite {} conf'.format(servs),
                         data=raw_data,
                         w=True,
                         json_r=True)

    def _check_db_run(self, code):
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
                self._rolback(code)
            sleep(5)

    def _clear_db_log(self):
        self.file_rw(p=main_conf['build']['db_log'],
                     w=True,
                     log='Clear DB log')

    def _run_dapp_cmd(self):
        # generate two conf in path:
        #  /var/lib/container/vpn/opt/privatix/config/dappvpn.config.json
        #  /var/lib/container/common/opt/privatix/config/dappvpn.config.json

        cmds = self.file_rw(
            p=self._reletive_path(main_conf['build']['cmd_path']),
            log='Read dapp cmd')
        logging.debug('Dupp cmds: {}'.format(cmds))

        if cmds:
            for cmd in cmds:
                self._sys_call(cmd=cmd)
                sleep(1)
        else:
            logging.error('Have not {} file for further execution. '
                          'It is necessary to run the initializer '
                          'in build mode.'.format(
                main_conf['build']['cmd_path']))
            self._rolback(10)

    def _test_mode(self):
        data = urlopen(url=main_conf['test']['sql']).read()
        self.file_rw(p=main_conf['test']['path'], w=True, data=data,
                     log='Create file with test sql data.')
        cmd = main_conf['test']['cmd'].format(main_conf['test']['path'])

        self._sys_call(cmd=cmd, s_exit=12)
        raw_tmpl = self._get_url(self.dupp_vpn_templ)
        self.file_rw(p=self.dappvpnconf, w=True,
                     data=raw_tmpl, log='Create file with test sql data.')

    def conf_dappctrl_dev_json(self, db_conn):
        ''' replace the config dappctrl_dev_conf_json
        with dappctrl.config.local.json '''
        json_conf = self._get_url(self.p_dapctrl_dev_conf)
        if json_conf and json_conf.get['DB']:
            json_conf['DB'] = db_conn
            return json_conf
        return False


    def ip_port_dapp(self):
        """Check ip addr, free ports and replace it in
        dappctrl.config.local.json"""
        logging.debug('Check IP, Port for dappctrl.')
        search_keys = ['AgentServer', 'PayAddress', 'PayServer',
                       'SessionServer']
        pay_port = dict(old=None, new=None)

        # Read dappctrl.config.local.json
        data = self.file_rw(p=self.p_dap_conf, json_r=True,
                            log='Read dappctrl conf')
        if not data:
            self._rolback(22)

        if args['link']:
            db_conn = data.get('DB')  # save db params from local
            res = self.conf_dappctrl_dev_json(db_conn)
            if res: data = res

        # Check and change self ip and port for PayAddress
        my_ip = urlopen(url='http://icanhazip.com').read().replace('\n', '')
        logging.debug('Find IP: {}. Write it.'.format(my_ip))

        raw = data['PayAddress'].split(':')

        raw[1] = '//{}'.format(my_ip)

        delim = '/'
        rout = raw[-1].split(delim)
        pay_port['old'] = rout[0]
        pay_port['new'] = self.check_port(pay_port['old'])
        rout[0] = pay_port['new']
        raw[-1] = delim.join(rout)

        data['PayAddress'] = ':'.join(raw)

        # chenge role agent or client
        data['Role'] = self.dappctrl_role

        # Search ports in conf and store it to main_conf['ports']
        for k, v in data.iteritems():
            if isinstance(v, dict) and v.get('Addr'):
                delim = ':'
                raw_row = v['Addr'].split(delim)
                port = raw_row[-1]
                logging.debug('Find port: {}. Check it. {}'.format(port, k))

                # if int(port) == int(pay_port['old']):
                if k == 'PayServer':
                    raw_row[-1] = pay_port['new']
                    if not self.dappctrl_role == 'client':
                        self.use_ports['comm_port'].append(pay_port['new'])

                else:
                    port = self.check_port(port)
                    raw_row[-1] = port
                    self.use_ports['comm_port'].append(port)
                    if k == 'AgentServer':
                        self.apiEndpoint = port
                        self.use_ports['apiEndpoint'] = port
                    if k == 'SessionServer':
                        self.sessServPort = port
                        if self.dappctrl_role == 'client':
                            self.use_ports['common'].remove(port)

                data[k]['Addr'] = delim.join(raw_row)

        # Rewrite dappctrl.config.local.json
        self.file_rw(p=self.p_dap_conf, w=True, json_r=True, data=data,
                     log='Rewrite conf')


class Rdata(CMD):
    def __init__(self):
        CMD.__init__(self)

        self.p_unpck = dict(vpn=self.path_vpn, common=self.path_com)

    def download(self, code):
        try:
            logging.info('Begin download files.')
            dev_url = ''
            if not isdir(self.p_dwld):
                mkdir(self.p_dwld)

            obj = URLopener()
            if hasattr(self, 'back_route'):
                dev_url = self.back_route + '/'
                logging.debug('Back dev rout: "{}"'.format(self.back_route))

            for f in self.f_dwnld:
                logging.info(
                    self.wait_mess.format('Start download {}'.format(f)))

                # st = Thread(target=self.long_waiting)
                # st.daemon = True
                # st.start()
                logging.debug('url_dwnld:{}, dev_url:{} ,f: {}'.format(self.url_dwnld,dev_url,f))
                dwnld_url = self.url_dwnld + '/' + dev_url + f
                dwnld_url = dwnld_url.replace('///', '/')
                logging.debug(' - dwnld url: "{}"'.format(dwnld_url))
                obj.retrieve(dwnld_url, self.p_dwld + f)
                self.waiting = False
                sleep(0.1)
                logging.info('Download {} done.'.format(f))
            return True

        except BaseException as down:
            logging.error('Download: {}.'.format(down))
            self._rolback(code)

    def unpacking(self):
        logging.info('Begin unpacking download files.')
        try:
            for f in self.f_dwnld:
                if '.tar.xz' == f[-7:]:
                    logging.info('Unpacking {}.'.format(f))
                    # self.long_waiting()

                    for k, v in self.p_unpck.items():
                        if k in f:
                            if not isdir(self.p_dwld + v):
                                mkdir(self.p_dwld + v)
                            cmd = 'tar xpf {} -C {} --numeric-owner'.format(
                                self.p_dwld + f, self.p_dwld + v)
                            self._sys_call(cmd)
                            logging.info('Unpacking {} done.'.format(f))
                            # self.waiting = False

        except BaseException as p_unpck:
            logging.error('Unpack: {}.'.format(p_unpck))

    def clean(self):
        logging.info('Delete downloaded files.')

        for f in self.f_dwnld:
            logging.info('Delete {}'.format(f))
            remove(self.p_dwld + f)


class GUI(CMD):
    def __init__(self):
        CMD.__init__(self)

    def _prepare_icon(self):
        if environ.get('SUDO_USER'):
            logging.debug('SUDO_USER')
            if self.__check_desctop_dir('/home/', environ['SUDO_USER']):
                self.gui_icon = self.gui_icon_path.format(
                    '/home/', environ['SUDO_USER']) + self.gui_icon_name

                self.chown_cmd = self.gui_icon_chown.format(
                    environ['SUDO_USER'], self.gui_icon
                )
                self.__create_icon()
            else:
                self.gui_icon = self.gui_icon_path_sh.format(
                    '/home/', environ['SUDO_USER'], self.gui_icon_sh)

                self.chown_cmd = self.gui_icon_chown.format(
                    environ['SUDO_USER'], self.gui_icon
                )
                self.__create_icon_sh()

        else:
            logging.debug('HOME')
            if self.__check_desctop_dir('', environ['HOME']):
                self.gui_icon = self.gui_icon_path.format(
                    '', environ['HOME']) + self.gui_icon_name

                self.chown_cmd = self.gui_icon_chown.format(
                    environ['USER'], self.gui_icon
                )
                self.__create_icon()

            else:
                self.gui_icon = self.gui_icon_path_sh.format(
                    '', environ['HOME'], self.gui_icon_sh)

                self.chown_cmd = self.gui_icon_chown.format(
                    environ['USER'], self.gui_icon
                )
                self.__create_icon_sh()

        logging.debug('Gui icon: {}'.format(self.gui_icon))

    def __check_desctop_dir(self, p, u):
        if not isdir(self.gui_icon_path.format(p, u)):
            logging.debug(
                '{} not exist'.format(self.gui_icon_path.format(p, u)))
            return False
        logging.debug('{} exist'.format(self.gui_icon_path.format(p, u)))
        return True

    def __create_icon_sh(self):
        logging.debug('Create file: {}'.format(self.gui_icon))

        logging.info('The directory needed to create the startup '
                     'icon file was not found.\n'
                     'After the installation is complete, to run the program\n'
                     'you will need to run the file "sudo {}".\n'
                     'Press enter to continue.'.format(self.gui_icon))

        raw_input('')
        with open(self.gui_icon, 'w') as icon:
            cmd = self.gui_icon_tmpl['Exec']

            icon.writelines(cmd)

        self.__icon_rights()

    def __icon_rights(self):
        logging.debug('Create {} file done'.format(self.gui_icon))

        chmod(self.gui_icon,
              stat(self.gui_icon).st_mode | S_IXUSR | S_IXGRP | S_IXOTH)
        logging.debug('Chmod file done')
        self._sys_call(self.chown_cmd)
        logging.debug('Chown file done')

    def __create_icon(self):
        config = ConfigParser()
        config.optionxform = str
        section = self.gui_icon_tmpl['Section']
        logging.debug('Create file: {}'.format(self.gui_icon))

        with open(self.gui_icon, 'w') as icon:
            config.add_section(section)
            [config.set(section, k, v) for k, v in
             self.gui_icon_tmpl.items()]
            config.write(icon)

        self.__icon_rights()

    def __get_gui(self):
        if hasattr(self, 'gui_route'):
            try:

                dev_url = self.url_dwnld + self.gui_dev_link + self.gui_route + '/' + self.gui_arch
                self._sys_call(self.gui_installer[0], s_exit=11)
                logging.debug('Gui dev rout: "{}"'.format(dev_url))
                obj = URLopener()
                obj.retrieve(dev_url, self.gui_path + self.gui_arch)
                logging.info('Download {} done.'.format(self.gui_arch))
                logging.info('Begin unpacking download file.')

                cmd = 'tar xpf {} -C {} --numeric-owner'.format(
                    self.gui_path + self.gui_arch, self.gui_path)
                self._sys_call(cmd)
                logging.info('Unpacking {} done.'.format(self.gui_arch))

                cmd = 'cd / && sudo npm install --prefix /opt/privatix/gui && sudo chown -R root:root /opt/privatix/gui'
                self._sys_call(cmd)
                self.dappctrlgui = '/opt/privatix/gui/settings.json'

                self.gui_icon_tmpl['Exec'] = self.gui_icon_tmpl['Exec'].format('')
                self.gui_icon_tmpl['Icon'] = self.gui_icon_tmpl['Icon'].format('')

            except BaseException as down:
                logging.error('Download {}.'.format(down))
                self._rolback(26)

        else:
            self.gui_icon_tmpl['Exec'] = self.gui_icon_tmpl['Exec'].format(self.gui_icon_prod)

            self.gui_icon_tmpl['Icon'] = self.gui_icon_tmpl['Icon'].format(self.gui_icon_prod)
            for cmd in self.gui_installer:
                self._sys_call(cmd, s_exit=11)

        if not isfile(self.dappctrlgui):
            logging.info(
                'The dappctrlgui package is not installed correctly')
            self._rolback(27)
        self._prepare_icon()
        self.__rewrite_config()

    def __rewrite_config(self):
        """
        /opt/privatix/gui/node_modules/dappctrlgui/settings.json
        example data structure:
        {
            "firstStart": false,
            "accountCreated": true,
            "apiEndpoint": "http://localhost:3000",
            "gas": {
                "acceptOffering": 100000,
                "createOffering": 100000,
                "transfer": 100000
            },
            "network": "rinkeby"
        }
        """
        try:
            raw_data = self.file_rw(p=self.dappctrlgui,
                                    log='Read settings.json',
                                    json_r=True)
            delim = ':'
            raw_link = raw_data['apiEndpoint'].split(delim)
            raw_link[-1] = self.apiEndpoint
            raw_data['apiEndpoint'] = delim.join(raw_link)

            if not self.file_rw(p=self.dappctrlgui,
                                w=True,
                                data=raw_data,
                                json_r=True,
                                log='Rewrite settings.json'):
                raise BaseException(
                    '{} was not found.'.format(self.dappctrlgui))
        except BaseException as rwconf:
            logging.error('R\W settings.json: {}'.format(rwconf))
            self._rolback(25)

    def __get_npm(self):
        # install npm and nodejs
        logging.debug('Get NPM for GUI.')
        npm_path = self._reletive_path(self.gui_npm_tmp_f)
        self.file_rw(
            p=npm_path,
            w=True,
            data=urlopen(self.gui_npm_url),
            log='Download nodesource'
        )

        cmd = self.gui_npm_cmd_call + npm_path
        self._sys_call(cmd=cmd, s_exit=11)

        cmd = self.gui_npm_node
        self._sys_call(cmd=cmd, s_exit=11)

    def __get_pack_ver(self):
        res = False
        for k, v in self.gui_version.items():
            logging.info('Check {} version.'.format(k))
            cmd = main_conf['search_pack'].format(k)
            raw = self._sys_call(cmd=cmd)
            if raw:
                res = True
                cmd = '{} -v'.format(k)
                raw = self._sys_call(cmd=cmd)
                ver = '.'.join(findall('\d+', raw)[0:2])

                if StrictVersion(ver) < StrictVersion(v[0]):
                    self.gui_version[k][1] = True
                    self.gui_version[k][2] = ver

            else:
                logging.info('{} not installed yet.'.format(k))
        return res

    def __check_version(self):
        self.__get_pack_ver()

        logging.debug('Check dependencies.')

        if any([x[1] for x in self.gui_version.values()]):
            logging.info('\n\nYou have installed obsolete packages.\n'
                         'To continue the installation, '
                         'you need to update the following packages:')
            for k, v in self.gui_version.items():
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
                    self._rolback(15)
                    break

            for k, v in self.gui_version.items():
                if v[1]:
                    logging.info('Preparing for deletion '
                                 '{} {}'.format(k, v[2]))
                    cmd = main_conf['del_pack'].format(k)
                    self._sys_call(cmd=cmd)

            if self.__get_pack_ver():
                logging.info('The problem with deleting one of the listed '
                             'packages. Try to delete in manual mode '
                             'and repeat the process again.')
                self._rolback(16)

    def install_gui(self):
        logging.debug('Install GUI.')
        self.__check_version()
        self.__get_npm()
        self.__get_gui()

    def _clear_gui(self):
        logging.info('Clear GUI.')
        p = self.gui_path
        self._clear_dir(p)
        # rmtree(p, ignore_errors=True)
        # mkdir(p)

    def update_gui(self):
        logging.info('Update GUI.')
        self.apiEndpoint = self.use_ports.get('apiEndpoint')
        if not self.use_ports.get('apiEndpoint'):
            logging.info('You can not upgrade GUI before '
                         'you not complete the installation.')
            sys.exit()

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
            if not isfile(self._reletive_path(main_conf['build']['cmd_path'])):
                logging.info('There is no .dapp_cmd file for further work.\n'
                             'To create it, you must run '
                             './initializer.py --build')
                logging.debug(self._reletive_path(main_conf['build']['cmd_path']))
                sys.exit(28)
            dist_name, ver, name_ver = linux_distribution()
            upgr_pack = self.task.get(dist_name.lower(), False)
            if not upgr_pack:
                logging.error('You system is {}.'
                              'She is not supported yet'.format(dist_name))
                sys.exit(19)
            upgr_pack(ver)
            ip, intfs, tun, port = self.revise_params()
            self.download(6)
            try:
                self.unpacking()
                self._rw_openvpn_conf(ip, tun, port, 7)
                self._rw_unit_file(ip, intfs, 5)
                self.clean()
                self._clear_db_log()
                self.ip_port_dapp()
                self.run_service(comm=True)
                self._check_db_run(9)
                if not args['test']:
                    logging.info('Test mode.')
                    self._test_mode()
                else:
                    logging.info('Full mode.')
                    self._run_dapp_cmd()
                    self._check_dapp_conf()

                self.run_service()
                if not args['no_gui']:
                    logging.info('GUI mode.')
                    check.target = 'both'
                    if pass_check:
                        self.update_gui()
                    else:
                        self.install_gui()

                self._finalizer(rw=True)
            except BaseException as mexpt:
                logging.error('Main trouble: {}'.format(mexpt))
                self._rolback(17)

    def prompt(self, mess, choise = ('n', 'y')):
        logging.info(mess)

        answ = raw_input('>')

        while True:
            if answ.lower() not in choise:
                logging.info('Invalid choice. Select {}.'.format(choise))
                answ = raw_input('> ')
                continue
            if answ.lower() == choise[1]:
                return True
            return False

    def check_graph(self):
        if not isdir(self.gui_icon_path) and not args['no_gui']:
            mess = 'You chosen a full installation with a GUI,\n' \
                   'but did not find a GUI on your computer.\n' \
                   'Y - I understand. Continue the installation but without GUI\n' \
                   'N - Stop the installation.'

            if self.prompt(mess=mess):
                args['no_gui'] = True
            else:
                sys.exit(24)
        logging.debug('Path exist: {}'.format(self.gui_icon_path))

    def __select_sources(self):

        def back():
            logging.info('Enter the Back build number for downloading')
            back_build = raw_input('>')
            mess = "You enter '{}'\n" \
                   "Please enter N and try again or " \
                   "Y if everything is correct".format(back_build)

            if not self.prompt(mess):
                back()
            self.back_route = back_build

        def gui():
            logging.info('Enter the GUI build number for downloading')
            gui_build = raw_input('>')
            mess = "You enter '{}'\n" \
                   "Please enter N and try again or " \
                   "Y if everything is correct".format(gui_build)

            if not self.prompt(mess):
                gui()

            self.gui_route = gui_build

        def back_gui():
            back()
            gui()

        if args['update_back'] or args['no_gui']:
            back()
        elif args['update_gui']:
            gui()
        else:
            back_gui()
        # elif args['update_mass']:
        #     back_gui()
        # else:
        #     logging.info('You have entered a new address.\n'
        #                  'Choose what you want to apply it for.\n'
        #                  'Make your choice:\n'
        #                  '1 - Back\n'
        #                  '2 - GUI\n'
        #                  '3 - Back and GUI\n')
        #     choise_task = {1: back, 2: gui, 3: back_gui}
        #     while True:
        #         choise_code = raw_input('>')
        #
        #         if choise_code.isdigit() and int(choise_code) in choise_task:
        #             choise_task[int(choise_code)]()
        #             break
        #         else:
        #             logging.info('Wrong choice. Make a choice between: '
        #                          '{}'.format(choise_task.keys()))

    def validate_url(self, url):
        regex_url = compile(
            r'^(?:http|https)s?://'  # http:// or https://
            r'(?:(?:[A-Z0-9](?:[A-Z0-9-]{0,61}[A-Z0-9])?\.)+(?:[A-Z]{2,6}\.?|[A-Z0-9-]{2,}\.?)|'  # domain...
            r'localhost|'  # localhost...
            r'\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})'  # ...or ip
            r'(?::\d+)?'  # optional port
            r'(?:/?|[/?]\S+)$', IGNORECASE)

        while True:
            if match(regex_url, url):
                logging.info('The address: {} is correct.'.format(url))
                check.url_dwnld = url + '/'
                self.__select_sources()
                break
            else:
                logging.info('\nThe address: {} was entered incorrectly.\n'
                             'Please enter it according to the example:\n'
                             'http://www.example.com/'.format(url))
                url = raw_input('>')

    def check_sudo(self):
        mess = 'Make sure that the user from which you are installing\n' \
               'is added to the sudo file with parameters\n' \
               'ALL = (ALL: ALL) NOPASSWD: ALL\n' \
               'or this user himself is a local root\n\n' \
               'Y - I understand. Everything is fine\n' \
               'N - The user does not root, stop the installation.'

        if not self.prompt(mess=mess):
            sys.exit(21)

        self.check_role()
            # self.check_graph()

    def check_role(self):
        mess = 'Please select your role.\n Enter digits 1 or 2.\n' \
               '1 - You role are agent\n' \
               '2 - You role are client\n' \

        if self.prompt(mess=mess,choise=('1','2')):
            #choise 2
            self.dappctrl_role = 'client'
        else:
            #choise 1
            self.dappctrl_role = 'agent'

if __name__ == '__main__':
    parser = ArgumentParser(description=' *** Installer *** ')
    parser.add_argument("--build", action='store_true', default=False,
                        help='Create .dapp_cmd file.')

    parser.add_argument("--update-back", action='store_true', default=False,
                        help='Update containers and rerun initializer.')

    parser.add_argument("--update-gui", action='store_true', default=False,
                        help='Update GUI.')

    parser.add_argument("--update-mass", action='store_true', default=False,
                        help='Update All.')

    parser.add_argument('--vpn', type=str, default=False,
                        help='[start,stop,restart,status]')

    parser.add_argument('--comm', type=str, default=False,
                        help='[start,stop,restart,status]')

    parser.add_argument('--mass', type=str, default=False,
                        help='[start,stop,restart,status]')

    parser.add_argument("--test", nargs='?', default=True,
                        help='')

    parser.add_argument("--no-gui", action='store_true', default=False,
                        help='Full install without GUI.')

    parser.add_argument("--no-wait", action='store_true',
                        default=False,
                        help='Installation without checking ports and waiting for their open.')

    parser.add_argument("--clean", action='store_true', default=False,
                        help='Cleaning after the initialization process. Removing GUI, downloaded files, initialization pid file, stopping containers.')

    parser.add_argument("--link", type=str, default=False, nargs='?',
                        help='Enter link for download. default "http://art.privatix.net/"')

    parser.add_argument("--branch", type=str, default=False, nargs='?',
                        help='Enter different branch for download. default "develop"')

    args = vars(parser.parse_args())

    check = Checker()

    if isfile(check.fin_file):
        raw = check.file_rw(p=check.fin_file,
                            json_r=True,
                            log='Search port in finalizer.pid')
        if raw:
            check.use_ports.update(raw)

    logging.debug('Input args: {}'.format(args))
    logging.debug('Inside finalizer.pid: {}'.format(check.use_ports))

    signal(SIGINT, check.signal_handler)

    if args['link']:
        logging.info('You chose was to change link from: {}   to: {}'.format(
            main_conf['link_download'], args['link']))

        check.validate_url(args['link'])

    if args['branch']:
        logging.debug('Change branch from: {}, to: {}'.format(
            main_conf['branch'], args['branch']))

        main_conf['branch'] = args['branch']
        check.re_init()

    if args['build']:
        logging.info('Build mode.')
        check.build_cmd()

    elif args['clean']:
        logging.info('Clean mode.')
        check.clear_contr()
        check._clear_dir(check.p_dwld)
        if isfile(check.fin_file):
            remove(check.fin_file)

    elif args['vpn']:
        logging.debug('Vpn mode.')
        sys.stdout.write(
            str(check.service('vpn', args['vpn'],
                              check.use_ports['vpn_port'])))

    elif args['comm']:
        logging.debug('Comm mode.')
        sys.stdout.write(
            str(check.service('comm', args['comm'],
                              check.use_ports['comm_port'])))

    elif args['mass']:
        logging.debug('Mass mode.')
        comm_stat = check.service('comm', args['mass'],
                                  check.use_ports['comm_port'])
        vpn_stat = check.service('vpn', args['mass'],
                                 check.use_ports['vpn_port'])
        sys.stdout.write(str(bool(all((comm_stat, vpn_stat)))))

    elif args['update_back']:
        logging.info('Update containers mode.')
        check.check_sudo()
        check.target = 'back'
        if check.clear_contr():
            check.use_ports = dict(vpn_port=[], comm_port=[],
                                   mangmt=dict(vpn=None, com=None))
            args['no_gui'] = True
            check.init_os(args, True)
        else:
            logging.info('Problem with clear all old file.')

    elif args['update_gui']:
        logging.info('Update GUI mode.')
        check.target = 'gui'
        check.check_sudo()
        check.update_gui()

    elif args['update_mass']:
        check.target = 'both'
        logging.info('Update All mode.')
        check.check_sudo()
        if check.clear_contr():
            check.use_ports = dict(vpn_port=[], comm_port=[],
                                   mangmt=dict(vpn=None, com=None))
            check.init_os(args, True)
        else:
            logging.info('Problem with clear all old file.')

    else:
        logging.info('Begin init.')
        check.target = 'back'
        check.check_sudo()
        check.init_os(args)
        logging.info('All done.')
