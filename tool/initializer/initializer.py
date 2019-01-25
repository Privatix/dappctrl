#!/usr/bin/python
# -*- coding: utf-8 -*-

"""
        Initializer on pure Python 2.7

        mode:
    python initializer.py                                  start full installation
    python initializer.py  -h                              get help information
    python initializer.py --build                          create cmd for dapp
    python initializer.py --vpn start/stop/restart/status  control vpn servise
    python initializer.py --comm start/stop/restart/status control common servise
    python initializer.py --mass start/stop/restart/status control common + vpn servise
    python initializer.py --test                           start in test mode
    python initializer.py --no-gui                         install without GUI
    python initializer.py --update-back                    update all contaiter without GUI
    python initializer.py --update-mass                    update all contaiter with GUI
    python initializer.py --update-gui                     update only GUI
    python initializer.py --update-bin                     download and update binary files in containers
    python initializer.py --link                           use another link for download.if not use, def link in main_conf[link_download]
    python initializer.py --branch                         use another branch than 'develop' for download. template https://raw.githubusercontent.com/Privatix/dappctrl/{ branch }/
    python initializer.py --cli                            auto offer mode
    python initializer.py --cli --file [path/to/file.json] auto offer mode with offer data from offer.json
    python initializer.py --cli --republish                republish offer mode
    python initializer.py --cli --file [] --republish      republish offer mode with offer data from offer.json
    python initializer.py --clean                          stop and delete all dirs containers and gui
    python initializer.py --no-wait                        installation without checking ports and waiting for their open
    python initializer.py --D                              switch on debug mode, and show it on console on installation process

"""

import sys
import logging
from codecs import open
from shutil import copyfile
from urllib import URLopener
from uuid import uuid1
from time import time, sleep
from threading import Thread
from contextlib import closing
from SocketServer import TCPServer
from urllib2 import urlopen, Request
from argparse import ArgumentParser
from ConfigParser import ConfigParser
from platform import linux_distribution
from random import randint, SystemRandom
from signal import SIGINT, signal, pause
from os.path import isfile, isdir, exists
from json import load, dump, loads, dumps
from stat import S_IXUSR, S_IXGRP, S_IXOTH
from subprocess import Popen, PIPE, STDOUT
from distutils.version import StrictVersion
from socket import socket, AF_INET, SOCK_STREAM
from SimpleHTTPServer import SimpleHTTPRequestHandler
from string import ascii_uppercase, ascii_lowercase, digits
from re import search, sub, findall, compile, match, IGNORECASE
from os import remove, mkdir, path, environ, stat, chmod, listdir, symlink


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
    29 - trouble when try install LXC
    30 - trouble when try check LXC contaiters 
    31 - found installed containers 
    32 - Problem with operation R/W LXC contaiter config or run sh or interfaces conf
    33 - Problem with check or validate offer conf file
    34 - Try run app with update mode in first time
    35 - Not start service on 8000 port
    36 - Problem with download binary for update
    37 - Problem with replase new binary for update

    not save port 8000 from Sess and 9000 from PayServer if role is client
"""

main_conf = dict(
    i_version='0.2.9',
    bind_port=False,
    bind_ports=[5555],
    log_path='/var/log/initializer.log',
    branch='develop',
    link_download='https://github.com/Privatix/privatix/releases/download/{}',
    mask=['/24', '255.255.255.0'],
    mark_final='/var/run/installer.pid',
    wait_mess='{}.Please wait until completed.\n It may take about 5-10 minutes.\n Do not turn it off.',
    tmp_var=None,
    del_pack='sudo apt purge {} -y',
    del_dirs='sudo rm -rf {}*',
    search_pack='sudo dpkg -l | grep {}',
    nspawn=False,
    openvpn_conf='etc/openvpn/config/server.conf',

    dappctrl_dev_conf_json='https://raw.githubusercontent.com/Privatix/dappctrl/{}/dappctrl-dev.config.json',
    dappctrl_conf_json='opt/privatix/config/dappctrl.config.local.json',
    dappvpn_conf_json='opt/privatix/config/dappvpn.config.json',

    back_nspwn=dict(
        addr='10.217.3.0',
        ports=dict(vpn=[], common=[],
                   mangmt=dict(vpn=None, common=None)),

        file_download_git=[
            'systemd_containers_ubuntu_x64_{}.tar.xz',
            'systemd_units_ubuntu_x64_{}.tar.xz'],

        file_download=[
            'vpn.tar.xz',
            'common.tar.xz',
            'systemd-nspawn@vpn.service',
            'systemd-nspawn@common.service'],
        path_container='/var/lib/container/',
        path_vpn='vpn/',
        path_com='common/',
        path_unit='/lib/systemd/system/',
        path_symlink_unit='/lib/systemd/system/multi-user.target.wants/',
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
        },
        db_log='/var/lib/container/common/var/log/postgresql/postgresql-10-main.log',
        db_stat='database system is ready to accept connections',
        tor=dict(
            socks_port=9099,
            config='etc/tor/torrcprod',
            hostname='var/lib/tor/hidden_service/hostname'
        )

    ),
    back_lxc=dict(
        addr='10.0.4.0',
        common_octet='52',
        vpn_octet='51',
        ports=dict(vpn=[], common=[],
                   mangmt=dict(vpn=None, common=None)),

        install=[
            'apt-add-repository ppa:ubuntu-lxc/stable -y',
            'apt-get update',
            'apt-get install lxc lxc-templates cgroup-bin bridge-utils debootstrap -y'
        ],
        path_container='/var/lib/lxc/',
        exist_contrs='lxc-ls -f',
        exist_contrs_ip=dict(),
        openvpn_port=['port 443', ''],

        bridge_cmd='ip addr show',
        bridge_conf='/etc/default/lxc-net',
        deff_lxc_cont_path='/var/lib/lxc/',
        lxc_cont_conf_name='config',
        lxc_cont_interfs='/rootfs/etc/network/interfaces',
        lxc_cont_fs_file='/rootfs/home/ubuntu/go/bin/dappctrl',
        kind_of_cont=dict(common='/rootfs/etc/postgresql/',
                          vpn='/rootfs/etc/openvpn/'),
        run_sh_path='/etc/init.d/',
        chmod_run_sh='sudo chmod +x {}',
        update_cont_conf='lxc-update-config -c {}',
        search_name='lxcbr',
        name_in_main_conf={
            'LXC_BRIDGE=': None,
            'LXC_ADDR=': None,
            'LXC_NETWORK=': None,
            'USE_LXC_BRIDGE=': None,
        },
        name_in_contnr_conf=[
            'lxc.network.ipv4',
            'lxc.uts.name',
            'hwaddr'
        ],

        file_download_git=[
            'systemd_containers_ubuntu_x64_{}.tar.xz',
            'systemd_units_ubuntu_x64_{}.tar.xz'],

        file_download=[
            'dapp-common',
            'dapp-vpn',
            'lxc-common.tar.xz',
            'lxc-vpn.tar.xz'],
        path_vpn='vpn/',
        path_com='common/',

        db_log='/var/lib/lxc/{}rootfs/var/log/postgresql/postgresql-10-main.log',
        db_stat='database system is ready to accept connections',
        db_conf_path='/var/lib/lxc/{}rootfs/etc/postgresql/10/main/',
    ),

    build={

        'cmd': '/opt/privatix/initializer/installer -rootdir=\"{0}\" -connstr=\"{1}\" -setauth\n'
               'ls {0} | grep .config.json | wc -l\n'
               'sudo cp {0}/{2} {3}\n'
               'sudo cp {0}/{4} {5}\n',

        'cmd_path': '.dapp_cmd',

        'db_conf': {
            "dbname": "dappctrl",
            "sslmode": "disable",
            "user": "postgres",
            "host": "localhost",
            "port": "5433"
        },

        'dappvpnconf_path': '/var/lib/container/vpn/opt/privatix/config/dappvpn.config.json',
        'dappconconf_path': '/var/lib/container/common/opt/privatix/config/dappvpn.config.json',
        'conf_link': 'https://raw.githubusercontent.com/Privatix/dappctrl/{}/dappctrl.config.json',
        'templ': 'https://raw.githubusercontent.com/Privatix/dappctrl/{}/svc/dappvpn/dappvpn.config.json',
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
)

logging.getLogger().setLevel('DEBUG')
form_console = logging.Formatter(
    '%(message)s',
    datefmt='%m/%d %H:%M:%S')

form_file = logging.Formatter(
    '%(levelname)7s [%(lineno)3s] %(message)s',
    datefmt='%m/%d %H:%M:%S')

fh = logging.FileHandler(main_conf['log_path'])  # file debug
fh.setLevel('DEBUG')
fh.setFormatter(form_file)
logging.getLogger().addHandler(fh)

ch = logging.StreamHandler()  # console debug
ch.setLevel('INFO')
ch.setFormatter(form_console)
logging.getLogger().addHandler(ch)
logging.debug('\n\n\n--- Begin ---')


class Init:
    recursion = 0
    target = None  # may will be back,gui,both
    sysctl = False
    waiting = True
    in_args = None  # arguments with which the script was launched.
    dappctrl_role = None  # the role: agent|client.

    def __init__(self):
        self.old_vers = None  # True if use LXC ond False if Nspawn
        self.i_version = main_conf['i_version']
        self.url_dwnld = main_conf['link_download']
        self.uid_dict = dict(userid=str(uuid1()))
        self.bind_port = main_conf['bind_port']
        self.bind_ports = main_conf['bind_ports']
        self.tmp_var = main_conf['tmp_var']
        self.fin_file = main_conf['mark_final']
        self.wait_mess = main_conf['wait_mess']
        self.p_dapctrl_conf = main_conf[
            'dappctrl_conf_json']  # ping [3000,8000,9000]

        self.p_dapvpn_conf = main_conf['dappvpn_conf_json']
        self.ovpn_conf = main_conf['openvpn_conf']

        test = main_conf['test']
        self.test_path = test['path']
        self.test_sql = test['sql']
        self.test_cmd = test['cmd']

        gui = main_conf['gui']
        self.gui_arch = gui['gui_arch']
        self.gui_path = gui['gui_path']
        self.gui_version = gui['version']
        self.gui_icon_sh = gui['icon_sh']
        self.dappctrlgui = gui['dappctrlgui']
        self.gui_npm_url = gui['npm_url']
        self.gui_dev_link = gui['link_dev_gui']
        self.gui_npm_node = gui['npm_node']
        self.gui_icon_name = gui['icon_name']
        self.gui_icon_path = gui['icon_dir']
        self.gui_icon_tmpl = gui['icon_tmpl']
        self.gui_icon_prod = gui['icon_prod']
        self.gui_installer = gui['gui_inst']
        self.gui_npm_tmp_f = gui['npm_tmp_f']
        self.gui_icon_chown = gui['chown']
        self.gui_icon_path_sh = gui['icon_tmpl_f_sh']
        self.gui_npm_cmd_call = gui['npm_tmp_file_call']

        dnsmasq = main_conf['dnsmasq']
        self.dns_conf = dnsmasq['conf']
        self.dns_sect = dnsmasq['section']
        self.dns_disable = dnsmasq['disable']
        self.init_branch()

    def init_branch(self):
        logging.debug('Initial vars')
        bld = main_conf['build']
        self.db_conf = bld['db_conf']
        self.build_cmd = bld['cmd']
        self.dupp_raw_id = bld['dappctrl_id_raw'].format(main_conf['branch'])
        self.dappctrl_id = bld['dappctrl_id']

        self.field_name_id = bld['field_name_id']
        self.dupp_conf_url = bld['conf_link'].format(main_conf['branch'])
        self.dupp_vpn_templ = bld['templ'].format(main_conf['branch'])
        self.build_cmd_path = bld['cmd_path']

        self.p_dapctrl_dev_conf = main_conf['dappctrl_dev_conf_json'].format(
            main_conf['branch'])

    def __init_back(self, back):
        self.addr = back['addr']
        self.p_contr = back['path_container']
        self.f_dwnld_git = back['file_download_git']
        self.f_dwnld = back['file_download']
        self.path_vpn = back['path_vpn']
        self.path_com = back['path_com']
        self.ovpn_port = back['openvpn_port']

        self.use_ports = back['ports']  # store all ports for monitor

        self.db_log = back['db_log']
        self.db_stat = back['db_stat']
        self.p_unpck = dict(
            vpn=[self.path_vpn, '0.0.0.0'],
            common=[self.path_com, '0.0.0.0']
        )
        tor = back['tor']
        self.tor_socks_port = tor['socks_port']
        self.tor_config = tor['config']
        self.tor_hostname_config = tor['hostname']

    def _init_nspwn(self):
        back = main_conf['back_nspwn']
        self.__init_back(back)
        self.ovpn_tun = back['openvpn_tun']
        self.ovpn_fields = back['openvpn_fields']
        self.unit_dest = back['path_unit']
        self.unit_symlink = back['path_symlink_unit']
        self.unit_f_com = back['unit_com']
        self.unit_f_vpn = back['unit_vpn']
        self.unit_params = back['unit_field']

    def _init_lxc(self):
        back = main_conf['back_lxc']
        self.__init_back(back)
        self.db_conf_path = back['db_conf_path']
        self.lxc_install = back['install']
        self.exist_contrs = back['exist_contrs']
        self.lxc_contrs = back['exist_contrs_ip']
        self.bridge_cmd = back['bridge_cmd']
        self.bridge_conf = back['bridge_conf']
        self.deff_lxc_cont_path = back['deff_lxc_cont_path']
        self.lxc_cont_conf_name = back['lxc_cont_conf_name']
        self.lxc_cont_interfs = back['lxc_cont_interfs']
        self.lxc_cont_fs_file = back['lxc_cont_fs_file']
        self.kind_of_cont = back['kind_of_cont']
        self.run_sh_path = back['run_sh_path']
        self.chmod_run_sh = back['chmod_run_sh']
        self.update_cont_conf = back['update_cont_conf']
        self.search_name = back['search_name']
        self.name_in_main_conf = back['name_in_main_conf']
        self.name_in_contnr_conf = back['name_in_contnr_conf']
        self.p_unpck['vpn'] = [self.path_vpn, back['vpn_octet']]
        self.p_unpck['common'] = [self.path_com, back['common_octet']]
        self.def_comm_addr = self.addr.split('.')
        self.def_comm_addr[-1] = back['common_octet']
        self.def_comm_addr = '.'.join(self.def_comm_addr)

    @staticmethod
    def long_waiting():
        symb = ['|', '/', '-', '\\']

        while Init.waiting:
            for i in symb:
                sys.stdout.write("\r[%s]" % (i))
                sys.stdout.flush()
                if not Init.waiting:
                    break
                sleep(0.3)
                if not Init.waiting:
                    break

        sys.stdout.write("\r")
        sys.stdout.write("")
        sys.stdout.flush()
        Init.waiting = True

    @staticmethod
    def wait_decor(func):
        def wrap(obj, args=None):
            logging.debug('Wait decor args: {}.'.format(args))
            st = Thread(target=Init.long_waiting)
            st.daemon = True
            st.start()
            if args:
                res = func(obj, args)
            else:
                res = func(obj)
            Init.waiting = False
            sleep(0.5)
            return res

        return wrap


class CommonCMD(Init):
    def __init__(self):
        Init.__init__(self)

    def get_latest_tag(self):
        logging.info('Get latest tag in repo.')

        owner = 'Privatix'
        repo = 'privatix'
        url_api = 'https://api.github.com/repos/{}/{}/releases/latest'.format(
            owner, repo)

        resp = self._get_url(link=url_api)

        if resp and resp.get('tag_name'):
            tag_name = resp['tag_name']
            logging.info(' * Latest release: {}'.format(tag_name))
            self.latest_tag = tag_name

            self.url_dwnld = self.url_dwnld.format(tag_name)

            for i, f in enumerate(self.f_dwnld_git):
                self.f_dwnld_git[i] = f.format(tag_name)

            self.bin_arch = 'privatix_ubuntu_x64_{}_binary.tar.xz'.format(tag_name)

        else:
            raise BaseException('GitHub not responding')

    def _sysctl(self):
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
                self._sysctl()
                return False
            else:
                logging.error('sysctl net.ipv4.ip_forward didnt change to 1')
                sys.exit(3)
        return True

    def _reletive_path(self, name):
        dirname = path.dirname(path.abspath(__file__))
        logging.debug('Reletive path: {}'.format(dirname))
        return path.join(dirname, name)

    def signal_handler(self, sign, frm):
        logging.info('You pressed Ctrl+C!')
        self._rolback(code=18)
        pause()

    def _clear_dir(self, p):
        logging.debug('Clear dir: {}'.format(p))
        cmd = main_conf['del_dirs'].format(p)
        self._sys_call(cmd)

    def _rolback(self, code):
        # Rolback net.ipv4.ip_forward and clear store by target
        logging.debug('Rolback target: {}, sysctl: {}'.format(self.target,
                                                              self.sysctl))
        if not self.old_vers and not self.sysctl:
            logging.debug('Rolback ip_forward')
            cmd = '/sbin/sysctl -w net.ipv4.ip_forward=0'
            self._sys_call(cmd)

        # if self.target == 'back':
        #     self.clear_contr(True)
        #
        # elif self.target == 'gui':
        #     self._clear_dir(self.gui_path)
        #
        # elif self.target == 'both':
        #     self.clear_contr(True)
        #     self._clear_dir(self.gui_path)
        # else:
        #     logging.debug('Absent `target` for cleaning!')

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
        unit_serv = {'vpn': self.unit_f_vpn, 'comm': self.unit_f_com}

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
                        return self._checker_port(port=port, status='stop')
                    else:
                        return self._checker_port(port=port)
                else:
                    return False

            if status == 'restart':
                check_stat = 'start' if 'start' in cmd else 'stop'
            else:
                check_stat = status

            if 'failed' in res or not self._checker_port(port=port,
                                                         status=check_stat):
                return False
            raw_res.append(True)

        if not port:
            return None
        return all(raw_res)

    @Init.wait_decor
    def clear_contr(self, pass_check=False):
        # Stop container.Check it if pass_check True.Clear conteiner path
        if pass_check:
            logging.info('\n\n   --- Attention! ---\n'
                         ' During installation a failure occurred'
                         ' or you pressed Ctrl+C\n'
                         ' All installed will be removed and returned to'
                         ' the initial state.\n Wait for the end!\n '
                         ' And try again.\n   ------------------\n')
        if not self.old_vers:
            # Nspawn mode
            self.service('vpn', 'stop', self.use_ports['vpn'])
            self.service('comm', 'stop', self.use_ports['common'])
            sleep(3)

            if pass_check or not self.service('vpn', 'status',
                                              self.use_ports['vpn'],
                                              True) and \
                    not self.service('comm', 'status',
                                     self.use_ports['common'], True):
                self._clear_dir(self.p_contr)
                return True
        else:
            # LXC mode
            check._check_contrs_by_path(True)
            logging.debug('Default container: {}'.format(self.p_unpck))
            for name, data in self.p_unpck.items():
                if len(data) > 2:
                    logging.debug('We may clear:'.format(data[2]))
                    cmd = 'lxc-stop -n {}'.format(data[2])
                    self._sys_call(cmd, rolback=False)
                    sleep(0.1)
                    res = self._sys_call(cmd, rolback=False)
                    if "is not running" in res:
                        logging.debug(
                            'Container {} is stoped'.format(data[2]))
                    else:
                        cmd = 'lxc-stop -W -n {}'.format(data[2])
                        self._sys_call(cmd, rolback=False)

                    cmd = 'lxc-destroy -f -n {}'.format(data[2])
                    self._sys_call(cmd, rolback=False)
                else:
                    logging.debug('Nothing to clean')

            check.lxc_contrs = dict()

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

    def run_service(self, comm=False, restart=False, nspwn=None):
        nspwn = False if self.old_vers else True

        if comm:
            if nspwn:
                if restart:
                    logging.info('Restart common service')
                    self._sys_call(
                        'systemctl stop {}'.format(self.unit_f_com))
                else:
                    logging.info('Run common service')
                    self._sys_call('systemctl daemon-reload')
                    sleep(2)
                    self._sys_call(
                        'systemctl enable {}'.format(self.unit_f_com))
                sleep(2)
                self._sys_call('systemctl start {}'.format(self.unit_f_com))
            else:
                if restart:
                    logging.info('Restart common service')
                    self._sys_call('service dapp-common stop')
                self._sys_call('service dapp-common start')

        else:
            if nspwn:
                if restart:
                    logging.info('Restart vpn service')
                    self._sys_call(
                        'systemctl stop {}'.format(self.unit_f_vpn))
                else:
                    logging.info('Run vpn service')
                    self._sys_call(
                        'systemctl enable {}'.format(self.unit_f_vpn))
                sleep(2)
                self._sys_call('systemctl start {}'.format(self.unit_f_vpn))
            else:
                if restart:
                    logging.info('Restart vpn service')
                    self._sys_call('service dapp-vpn stop')

                self._sys_call('service dapp-vpn start')

    def _sys_call(self, cmd, rolback=True, s_exit=4, code_ex=False):
        logging.debug('Sys call cmd: {}.'.format(cmd))
        obj = Popen(cmd, shell=True, stdout=PIPE, stderr=STDOUT)

        resp = obj.communicate()
        if code_ex:
            exit_code = obj.returncode
            logging.debug('Code: {}.\nResp: {}'.format(exit_code, resp))
            return (exit_code, resp)

        if resp[1]:
            logging.error('Response: {}'.format(resp))

            if rolback:
                self._rolback(s_exit)
            else:
                return False

        elif 'The following packages have unmet dependencies:' in resp[0]:
            if rolback:
                self._rolback(s_exit)
            exit(s_exit)

        return resp[0]

    def _disable_dns(self):
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

    def _ping_port(self, port, host='0.0.0.0', verb=False):
        '''open -> True  close -> False'''

        with closing(socket(AF_INET, SOCK_STREAM)) as sock:
            if sock.connect_ex((host, int(port))) == 0:
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

    def _cycle_ask(self, h, p, status, verb=False):
        logging.debug('Ask port: {}, status: {}'.format(p, status))
        ts = time()
        tw = 350

        if status == 'stop':
            logging.debug('Stop mode')
            while True:
                if not self._ping_port(port=p, host=h, verb=verb):
                    return True
                if time() - ts > tw:
                    return False
                sleep(2)
        else:
            logging.debug('Start mode')
            while True:
                if self._ping_port(port=p, host=h, verb=verb):
                    return True
                if time() - ts > tw:
                    return False
                sleep(2)

    def _checker_port(self, port, host='0.0.0.0', status='start',
                      verb=False):
        logging.debug('Checker: {}'.format(status))
        if not port:
            return None
        if isinstance(port, (list, set)):
            resp = list()
            for p in port:
                resp.append(self._cycle_ask(host, p, status, verb))
            return True if all(resp) else False
        else:
            return self._cycle_ask(host, port, status, verb)

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
        """_ping_port: open -> True  close -> False"""
        logging.debug('Check port. In: {}'.format(port))

        if self._ping_port(port=port):
            mark = False
            while True:

                if auto and mark:
                    port = int(port)+1
                else:
                    logging.info("\nPort: {} is busy or wrong.\n"
                                 "Select a different port,"
                                 "in range 1 - 65535.".format(port))
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
        logging.debug('Check port. Out: {}'.format(port))
        return port

    def __up_ports(self):
        logging.debug('Bind ports mode')

        def run_server(p):
            try:
                httpd = TCPServer(('localhost', p), SimpleHTTPRequestHandler)
                logging.debug(" ^ UP PORT: {}".format(p))
                httpd.serve_forever()
            except BaseException as thrExpt:
                logging.error(" ^ UP PORT: {}".format(thrExpt))

        for p in self.bind_ports:
            t = Thread(target=run_server, args=(p,))
            t.daemon = True
            t.start()

    @Init.wait_decor
    def __wait_up(self):
        logging.info(self.wait_mess.format('Run services'))

        logging.debug('Wait when up all ports: {}'.format(self.use_ports))
        if self.bind_port:
            self.__up_ports()
        # check common
        if not self._checker_port(
                host=self.p_unpck['common'][1],
                port=self.use_ports['common'],
                verb=True):
            logging.info('Restart Common')
            self.run_service(comm=True, restart=True)
            if not self._checker_port(
                    host=self.p_unpck['common'][1],
                    port=self.use_ports['common'],
                    verb=True):
                logging.error('Common is not ready')
                exit(14)

        # check vpn
        if not self._checker_port(
                host=self.p_unpck['vpn'][1],
                port=self.use_ports['vpn'],
                verb=True):
            logging.info('Restart VPN')
            self.run_service(comm=False, restart=True)
            if not self._checker_port(
                    host=self.p_unpck['vpn'][1],
                    port=self.use_ports['vpn'],
                    verb=True):
                logging.error('VPN is not ready')
                exit(13)

        if not self.old_vers and self.in_args['cli']:
            self._sym_lynk()

    def _sym_lynk(self):
        logging.debug('Symlink')
        for unit in (self.unit_f_com, self.unit_f_vpn):
            p = self.unit_symlink + unit
            logging.debug('Symlink: {}'.format(p))
            symlink(self.unit_dest + unit, p)
        logging.debug('Symlink done')

    def _finalizer(self, rw=None, pass_check=False):
        logging.debug(
            'Finalizer. rw: {}, pass_check: {}'.format(rw, pass_check))
        if pass_check:
            logging.debug('Pass check PID file')
            return True

        if not isfile(self.fin_file):
            self.file_rw(p=self.fin_file, w=True, log='First start')
            return True
        logging.debug('No wait args: {}'.format(self.in_args['no_wait']))
        if rw:
            if not self.in_args['no_wait']:
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
                     'If you need to re-install, you need to perform first\n'
                     'initializer.py --clean and then initializer.py.\n'
                     'Or you can perform initializer.py with one of three \n'
                     'update modes:\n'
                     '--update-mass  and update back and gui\n'
                     '--update-bin   and update only binary\n'
                     '--update-back  and update only back\n'
                     '--update-gui   and update only gui\n'
                     )
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

    def build_file(self):

        # Get DB params
        json_db = self._get_url(self.dupp_conf_url)
        db_conf = json_db.get('DB')
        logging.debug('DB params: {}'.format(db_conf))
        if db_conf:
            self.db_conf.update(db_conf['Conn'])

        # # Get dappctrl_id from prod_data.sql
        # if not self.dappctrl_id:
        #     raw_id = self._get_url(link=self.dupp_raw_id,
        #                            to_json=False).split('\n')
        #
        #     for i in raw_id:
        #         if self.field_name_id in i:
        #             self.dappctrl_id = i.split(self.field_name_id)[1]
        #             logging.debug('dapp_id: {}'.format(self.dappctrl_id))
        #             break
        #     else:
        #         logging.error(
        #             'dappctrl_id not exist: {}'.format(self.dappctrl_id))
        #         sys.exit(20)

        # Get dappvpn.config.json
        # templ = self._get_url(link=self.dupp_vpn_templ,
        #                       to_json=False).replace(
        #     '\n', '')

        self.db_conf = (sub("'|{|}", "", str(self.db_conf))).replace(
            ': ', '=').replace(',', '')
        p_vpn = self.p_contr + self.path_vpn + self.p_dapvpn_conf
        p_com = self.p_contr + self.path_com + self.p_dapvpn_conf
        root_dir = '/opt/privatix/initializer/files/example'
        conf_agent = 'dappvpn.agent.config.json'  # vpn config
        conf_client = 'dappvpn.client.config.json'  # common config

        logging.debug('Path configs: \n\t{}\n\t{}'.format(p_com, p_vpn))

        self.build_cmd = self.build_cmd.format(
            root_dir,
            self.db_conf,
            conf_agent,
            p_vpn,
            conf_client,
            p_com
        )

        logging.debug('Build cmd: {}'.format(self.build_cmd))
        self.file_rw(
            p=self._reletive_path(self.build_cmd_path),
            w=True,
            data=self.build_cmd,
            log='Create file with dapp cmd')

    def conf_dappctrl_dev_json(self, db_conn):
        ''' replace the config dappctrl_dev_conf_json
        with dappctrl.config.local.json '''
        logging.debug('Get dappctrl_dev_conf_json from repo')
        json_conf = self._get_url(self.p_dapctrl_dev_conf)
        if json_conf and json_conf.get('DB'):
            json_conf['DB'] = db_conn
            return json_conf
        return False

    def __exclude_port(self, tmp_store):
        logging.debug('Exclude port mode.')
        # only if agent is client !
        by_key = ['PayServer', 'SOMCServer']  # 9000,5555
        for i in by_key:
            if tmp_store.get(i):
                logging.debug('Exclude: {}'.format(i))
                del tmp_store[i]

        return tmp_store

    def pswd_from_conf(self):
        # get password stored in dappctrl.config.local.json
        logging.debug('Search pswd')

        p = self.p_contr + self.path_com
        if old_vers:
            p += 'rootfs/'
        p += self.p_dapctrl_conf

        # Read dappctrl.config.local.json
        data = self.file_rw(p=p, json_r=True, log='Read dappctrl conf')
        if not data:
            return False, 'Config dappctrl.config.local.json absend.\n' \
                          'You must complete the installation.'

        self.pswd = data.get('StaticPassword')
        if self.pswd:
            return True, 'Password was found.'
        return False, 'The password does not exist.\n' \
                      'It is necessary to perform auto-offering first, --cli mode.'

    def conf_dappctrl_json(self):
        """
        Check ip addr, free ports and replace it in
        common dappctrl.config.local.json
        in get_pswd=True,get password stored in dappctrl.config.local.json
        """
        logging.debug('Check IP, Port in common dappctrl.local.json')

        pay_port = dict(old=None, new=None)
        p = self.p_contr + self.path_com
        if self.old_vers:
            p += 'rootfs/'
        p += self.p_dapctrl_conf

        # Read dappctrl.config.local.json
        data = self.file_rw(p=p, json_r=True, log='Read dappctrl conf')
        if not data:
            self._rolback(22)

        if self.old_vers:
            # LXC mode
            r = data.get('DB').get('Conn')
            if r:
                r.update({"host": self.p_unpck['common'][1]})

        if self.in_args['link']:
            db_conn = data.get('DB')  # save db params from local
            res = self.conf_dappctrl_dev_json(db_conn)
            if res:
                data = res

        # Check and change self ip and port for PayAddress
        my_ip = urlopen(url='http://icanhazip.com').read().replace('\n', '')
        logging.debug('Found IP: {}. Write it.'.format(my_ip))

        raw = data['PayAddress'].split(':')

        raw[1] = '//{}'.format(my_ip)

        delim = '/'
        rout = raw[-1].split(delim)
        pay_port['old'] = rout[0]
        pay_port['new'] = self.check_port(pay_port['old'])
        rout[0] = pay_port['new']
        raw[-1] = delim.join(rout)

        data['PayAddress'] = ':'.join(raw)

        # change role: agent, client
        data['Role'] = self.dappctrl_role
        if self.pswd:
            data['StaticPassword'] = self.pswd

        # Search ports in conf and store it to main_conf['ports']
        tmp_store = dict()
        for k, v in data.iteritems():
            if isinstance(v, dict) and v.get('Addr'):
                delim = ':'
                raw_row = v['Addr'].split(delim)
                port = raw_row[-1]
                logging.debug('Key: {} port: {}, Check it.'.format(k, port))

                if k == 'PayServer':
                    # default Addr is 0.0.0.0:9000
                    # ping only when role agent
                    raw_row[-1] = pay_port['new']
                    # if self.dappctrl_role == 'agent':
                    tmp_store[k] = pay_port['new']

                else:
                    if self.old_vers and raw_row[0] == self.def_comm_addr:
                        # LXC mode
                        raw_row[0] = self.p_unpck['common'][1]

                    port = self.check_port(port)
                    raw_row[-1] = port

                    tmp_store[k] = port
                    if k == 'UI':
                        # default Addr is localhost:8888
                        self.wsEndpoint = port
                        self.use_ports['wsEndpoint'] = port

                    elif k == 'Sess':
                        # default Addr is localhost:8000
                        self.sessServPort = port

                data[k]['Addr'] = delim.join(raw_row)

        # add uid key in conf
        logging.debug('Add userid on dappctrl.config.local.json')
        if data.get('Report'):
            data['Report'].update(self.uid_dict)
        else:
            data['Report'] = self.uid_dict

        # Rewrite dappctrl.config.local.json
        self.file_rw(p=p, w=True, json_r=True, data=data,
                     log='Rewrite conf')

        if self.dappctrl_role == 'client':
            tmp_store = self.__exclude_port(tmp_store)

        self.use_ports['common'] = [v for k, v in tmp_store.items()]


class Tor(CommonCMD):
    def __init__(self):
        CommonCMD.__init__(self)

    def check_tor_port(self):
        logging.info('Check Tor config')

        self.tor_socks_port = int(self.check_port(port=self.tor_socks_port,
                                              auto=True))

        full_comm_p = self.p_contr + self.path_com
        data = self.file_rw(p=full_comm_p + self.p_dapctrl_conf,
                            json_r=True,
                            log='Read dappctrl conf')

        somc_serv = data.get('SOMCServer')
        if somc_serv:
            somc_serv_port = somc_serv['Addr'].split(':')[1]
            serv_port = '80 127.0.0.1:{}\n'.format(somc_serv_port)
            logging.debug('Tor HiddenServicePort: {}'.format(serv_port))
            data = self.file_rw(p=full_comm_p + self.tor_config,
                                log='Read tor conf.')
            if not data:
                raise BaseException('Tor config are absent!')

            search_line = {
                'SocksPort': '{}\n'.format(self.tor_socks_port),
                'HiddenServicePort': serv_port
            }

            for row in data:
                for k, v in search_line.items():
                    if k in row:
                        indx = data.index(row)
                        data[indx] = '{} {}'.format(k, v)

            self.file_rw(p=full_comm_p + self.tor_config,
                         w=True,
                         data=data,
                         log='Write tor conf.')

        else:
            mess = 'On dappctrl.config.json absent SOMCServer field'
            logging.error(mess)
            raise BaseException(mess)

    def get_onion_key(self):
        logging.debug('Get onion key')
        hostname_config = self.p_contr + self.path_com + self.tor_hostname_config
        onion_key = self.file_rw(
            p=hostname_config,
            log='Read hostname conf')
        logging.debug('Onion key: {}'.format(onion_key))

        data = self.file_rw(
            p=self.p_contr + self.path_com + self.p_dapctrl_conf,
            json_r=True,
            log='Read dappctrl conf')

        data.update(dict(
            TorHostname=onion_key[0].replace('\n', '')
        ))

        self.file_rw(p=self.p_contr + self.path_com + self.p_dapctrl_conf,
                     w=True,
                     json_r=True,
                     data=data,
                     log='Write add TorHostname to dappctrl conf.')

    def set_socks_list(self):
        logging.debug('Add TorSocksListener. Port: {}'.format(self.tor_socks_port))
        data = self.file_rw(p=self.p_contr + self.path_com + self.p_dapctrl_conf,
                            json_r=True,
                            log='Read dappctrl conf')

        data['TorSocksListener'] = self.tor_socks_port

        self.file_rw(p=self.p_contr + self.path_com + self.p_dapctrl_conf,
                     json_r=True,
                     w=True,
                     data=data,
                     log='Write dappctrl conf')


class DB(Tor):
    """This class provides a check
       if the database is started from its logs"""

    def __init__(self):
        Tor.__init__(self)

    @Init.wait_decor
    def _check_db_run(self, code):
        # wait 't_wait' sec until the DB starts, if not started, exit.

        t_start = time()
        t_wait = 300
        mark = True
        logging.info('Waiting for the launch of the DB.')
        while mark:
            logging.debug('Wait.')
            raw = self.file_rw(p=self.db_log,
                               log='Read DB log')
            for i in raw:
                logging.debug('DB : {}'.format(i))

                if self.db_stat in i:
                    logging.info('DB was run.')
                    mark = False
                    break
            if time() - t_start > t_wait:
                logging.error(
                    'DB after {} sec does not run.'.format(t_wait))
                logging.debug('Data base log: \n  {}'.format(raw))
                self._rolback(code)
            sleep(5)

    def _clear_db_log(self):

        self.file_rw(p=self.db_log,
                     w=True,
                     log='Clear DB log')


class Params(DB):
    """ This class provide check
    sysctl, iptables, port, ip"""

    def __init__(self):
        DB.__init__(self)

    def _iptables(self):
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
        self.use_ports['vpn'] = port
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
                if not search(p, addr):
                    logging.info('You addres is wrong,please enter '
                                 'right address.Last octet is always 0.Example: 255.255.255.0\n')
                    addr = check_addr(p)
                break
            return addr

        for i in arr:
            if self.addr + main_conf['mask'][0] in i:
                logging.info(
                    'Addres {} is busy or wrong, please enter new address '
                    'without changing the 4th octet.'
                    'Example: xxx.xxx.xxx.0\n'.format(self.addr))

                pattern = r'^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}0$'
                self.addr = check_addr(pattern)
                break
        return self.addr

    def _rw_unit_file(self, ip, intfs, code):
        logging.debug('Preparation unit file: {},{}'.format(ip, intfs))
        addr = ip + main_conf['mask'][0]
        try:
            # read a list of lines into data
            tmp_data = self.file_rw(p=self.p_contr + self.unit_f_vpn)
            logging.debug('Read {}'.format(self.unit_f_vpn))
            # replace all search fields
            for row in tmp_data:

                for param in self.unit_params.keys():
                    if param in row:
                        indx = tmp_data.index(row)

                        if self.unit_params[param]:
                            tmp_data[indx] = self.unit_params[param].format(
                                addr,
                                intfs)
                        else:
                            if self.sysctl:
                                tmp_data[indx] = ''

            # rewrite unit file
            logging.debug('Rewrite {}'.format(self.unit_f_vpn))
            self.file_rw(p=self.p_contr + self.unit_f_vpn, w=True,
                         data=tmp_data)
            del tmp_data

            # move unit files
            logging.debug('Move units.')
            copyfile(self.p_contr + self.unit_f_vpn,
                     self.unit_dest + self.unit_f_vpn)
            copyfile(self.p_contr + self.unit_f_com,
                     self.unit_dest + self.unit_f_com)
        except BaseException as f_rw:
            logging.error('R/W unit file: {}'.format(f_rw))
            self._rolback(code)

    def _check_dapp_conf(self):
        for servs, port in self.use_ports['mangmt'].iteritems():

            logging.debug('Dapp {} conf. Port: {}'.format(servs, port))
            if servs == 'vpn':
                p = self.p_contr + self.path_vpn + self.p_dapvpn_conf

            elif servs == 'common':
                p = self.p_contr + self.path_com + self.p_dapvpn_conf

            raw_data = self.file_rw(p=p,
                                    log='Check dapp {} conf'.format(servs),
                                    json_r=True)
            if not raw_data:
                self._rolback(23)

            if servs == 'common':
                raw_data = self._add_updown_path(raw_data)

            # "localhost:7505"
            logging.debug('dapp {} conf: {}'.format(servs, raw_data))
            delim = ':'
            raw_tmp = raw_data['Monitor']['Addr'].split(delim)
            raw_tmp[-1] = str(port)
            raw_data['Monitor']['Addr'] = delim.join(raw_tmp)
            logging.debug(
                'Monitor Addr: {}.'.format(raw_data['Monitor']['Addr']))

            if hasattr(self, 'sessServPort'):
                # "Endpoint": "ws://localhost:8000/ws"
                delim = ':'
                raw_tmp = raw_data['Sess']['Endpoint'].split(delim)
                raw_tmp[-1] = '{}/ws'.format(self.sessServPort)
                raw_data['Sess']['Endpoint'] = delim.join(raw_tmp)
                logging.debug(
                    'Sess Endpoint: {}.'.format(
                        raw_data['Sess']['Endpoint']))

            self.file_rw(p=p,
                         log='Rewrite {} conf'.format(servs),
                         data=raw_data,
                         w=True,
                         json_r=True)

    def _add_updown_path(self, raw_data):
        # Edit dappvpn.config.json and add up/down script path
        # up/down script path - /etc/openvpn/update-resolv-conf
        up_down = '/etc/openvpn/update-resolv-conf'
        oVpn = raw_data.get('OpenVPN')
        if oVpn:
            oVpn['UpScript'], oVpn['DownScript'] = up_down, up_down
        return raw_data



    def _run_dapp_cmd(self):
        # generate two conf in path:
        #  /var/lib/container/vpn/opt/privatix/config/dappvpn.config.json
        #  /var/lib/container/common/opt/privatix/config/dappvpn.config.json

        cmds = self.file_rw(
            p=self._reletive_path(self.build_cmd_path),
            log='Read dapp cmd')
        logging.debug('Dapp cmds: {}'.format(cmds))

        if cmds:
            for cmd in cmds:
                res = self._sys_call(cmd=cmd)
                if cmd.startswith('ls '):
                    # must be two configs files
                    logging.debug('Check quantity configs: {}'.format(res))
                    if not int(res) == 2:
                        logging.error('After generation, '
                                      'the necessary configuration files '
                                      'are missing.')
                        self._rolback(10)

                sleep(0.2)
        else:
            logging.error('Have not {} file for further execution. '
                          'It is necessary to run the initializer '
                          'in build mode.'.format(self.build_cmd_path))
            self._rolback(10)

    def _test_mode(self):
        data = urlopen(url=self.test_sql).read()
        self.file_rw(p=self.test_path, w=True, data=data,
                     log='Create file with test sql data.')
        cmd = self.test_cmd.format(self.test_path)

        self._sys_call(cmd=cmd, s_exit=12)
        raw_tmpl = self._get_url(self.dupp_vpn_templ)
        p = self.p_contr + self.path_vpn + self.p_dapvpn_conf

        self.file_rw(p=p, w=True,
                     data=raw_tmpl, log='Create file with test sql data.')


class UpdateBynary(CommonCMD):
    def __init__(self):
        CommonCMD.__init__(self)

    def init_vars(self):
        self.fold_route = 'binary/'
        self.vpn_bin = 'dappvpn'
        self.ctrl_bin = 'dappctrl'
        self.migrate_cmd = 'sudo {}common/root/go/bin/dappctrl db-migrate -conn {}'
        self.dump_tmp = '/var/lib/container_tmp'
        self.dump_path = [
            '{}common'.format(self.p_contr),
            '{}vpn/root/go/bin'.format(self.p_contr),
            '{}vpn/opt/privatix/config'.format(self.p_contr),
        ]
        self.dappvpn_route = ''
        self.dappctrl_route = ''

    @Init.wait_decor
    def __rollback_data(self):
        logging.info('Rollback containers data')
        for p in self.dump_path:
            p_src = self.dump_tmp + '/' + p.split('/')[-1]
            cmd = 'sudo cp -rf {} {}'.format(p_src, p)
            self._sys_call(cmd, rolback=False)

    @Init.wait_decor
    def update_binary(self):
        logging.debug('Stop containers')
        self.init_vars()
        self.service('vpn', 'stop', self.use_ports['vpn'])
        self.service('comm', 'stop', self.use_ports['common'])
        if self.__dump_data():
            self.get_latest_tag()
            self.__download_binary()
            self.__unpack_binary()
            self.__move_binary()

            if not self.__migrate_data():
                self.service('vpn', 'stop', self.use_ports['vpn'])
                self.service('comm', 'stop', self.use_ports['common'])
                self.__rollback_data()
                self.run_service(comm=True)
                self.run_service(comm=False)
            else:
                logging.debug('Clear tmp data')
                cmd = 'sudo rm -rf {}'.format(self.dump_tmp)
                self._sys_call(cmd, rolback=False)
        else:
            logging.debug('Run containers')
            self.run_service(comm=True)
            self.run_service(comm=False)

    @Init.wait_decor
    def __dump_data(self):
        # return 1
        logging.info('Prepare to dump data')
        cmd = 'sudo mkdir {0} && sudo chmod 777 {0}'.format(self.dump_tmp)
        self._sys_call(cmd, rolback=False)
        logging.info('Copying important files.\n'
                     'It may take some time. Do not interrupt the process.')
        for p in self.dump_path:
            cmd = 'sudo cp -rf {} {}'.format(p, self.dump_tmp)
            logging.debug('Dump: {}'.format(cmd))
            resp = self._sys_call(cmd=cmd, code_ex=True)
            if int(resp[0]):
                logging.error('Trouble when try dump data: {}'.format(resp[1]))
                return False
        return True

    def __read_dapp_cmd(self):
        data = self.file_rw(
            p=self._reletive_path(self.build_cmd_path),
            log='Read dapp cmd')
        logging.debug('Data dapp cmd: {}'.format(data))
        if data:
            for row in data[0].split(' -'):
                if 'connstr=' in row:
                    db_con = row.replace('connstr=', '')

                    db_port = [x for x in db_con.split(' ')
                               if 'port' in x]
                    db_port = findall('\d+', db_port[0])[0]
                    return db_con, db_port
        return False, False

    @Init.wait_decor
    def __migrate_data(self):
        logging.info('Migrate DB')
        logging.debug('Run containers')
        self.run_service(comm=True)
        self.run_service(comm=False)
        db_con, db_port = self.__read_dapp_cmd()

        logging.debug('Wait when run DB: {}'.format(db_port))

        if db_port and db_con and self._checker_port(
                port=db_port,
                verb=True):
            logging.debug('Wait 30 sec before DB init')
            sleep(30)
            cmd = self.migrate_cmd.format(self.p_contr, db_con)
            logging.debug('Migrate cmd: {}'.format(cmd))
            resp = self._sys_call(cmd=cmd, code_ex=True)
            if resp[0]:
                logging.info('Error when try migrate data: {}'.format(resp[1]))
                return False
            logging.info('Migrate done success')
            return True
        else:
            logging.info('Trouble when wait DB run.')
            return False

    @Init.wait_decor
    def __download_binary(self):
        logging.info('Download binary.')
        url = self.url_dwnld + '/' + self.bin_arch
        self._sys_call(
            cmd='wget -N {} -P {}'.format(url, self.p_contr))

        logging.info('Download binary arch done.')
        # url = self.url_dwnld + self.fold_route
        # dwnld_bin = {
        #     self.vpn_bin: url + self.dappvpn_route + self.vpn_bin,
        #     self.ctrl_bin: url + self.dappctrl_route + self.ctrl_bin
        # }
        #
        # logging.info('Begin download binary files.')
        # obj = URLopener()
        # try:
        #     for f, u in dwnld_bin.items():
        #         logging.info(
        #             self.wait_mess.format('Start download {}'.format(f)))
        #         logging.debug('Url binary downloading: {}'.format(u))
        #         obj.retrieve(u, f)
        #         sleep(0.1)
        #         logging.info('Download {} done.'.format(f))
        #
        # except BaseException as down:
        #     logging.error('Download: {}.'.format(down))
        #     sys.exit(36)

    @Init.wait_decor
    def __unpack_binary(self):
        logging.info('Begin unpacking binary.')

        cmd = 'tar xpf {} -C {} --numeric-owner'.format(
            self.p_contr + self.bin_arch, self.p_contr)
        self._sys_call(cmd)
        logging.info('Unpacking git {} done.'.format(self.bin_arch))


    @Init.wait_decor
    def __move_binary(self):
        logging.info('Start copy binary to destination')
        conts = dict(vpn=[self.vpn_bin], common=[self.ctrl_bin, self.vpn_bin])
        move_cmd = 'yes | sudo  cp -rf {} {}'
        try:
            for cont, fls in conts.items():
                for f in fls:
                    dest = self.p_contr + cont + '/root/go/bin/' + f
                    logging.info('Destination {}'.format(dest))
                    cmd = move_cmd.format(self.p_contr + f, dest)
                    logging.debug('Move cmd: {}'.format(cmd))
                    self._sys_call(cmd=cmd)
                    logging.debug('Binary chmod 777 for {}'.format(dest))
                    chmod(dest, 0777)

        except BaseException as down:
            logging.error('Move binary: {}.'.format(down))
            sys.exit(37)


class Rdata(UpdateBynary):
    """ Class for download, unpack, clear data """

    def __init__(self):
        UpdateBynary.__init__(self)

    @Init.wait_decor
    def download(self):
        try:
            logging.info(' - Begin download files.')
            dev_url = ''
            if not isdir(self.p_contr):
                logging.debug('Create dir: {}'.format(self.p_contr))
                mkdir(self.p_contr)

            obj = URLopener()
            if hasattr(self, 'back_route'):
                dev_url = self.back_route + '/'
                logging.debug('Back dev rout: "{}"'.format(self.back_route))

            for f in self.f_dwnld:
                logging.info(
                    self.wait_mess.format('Start download {}'.format(f)))

                logging.debug(
                    'url: {}, dev: {} ,file: {}'.format(self.url_dwnld,
                                                        dev_url,
                                                        f)
                )
                dwnld_url = self.url_dwnld + '/' + dev_url + f
                dwnld_url = dwnld_url.replace('///', '/')
                logging.debug(' - full url: "{}"'.format(dwnld_url))
                obj.retrieve(dwnld_url, self.p_contr + f)
                sleep(0.1)
                logging.info('Download {} done.'.format(f))

            logging.debug(' - Download all file ended.')

            return True

        except BaseException as down:
            logging.error(' - Download: {}.'.format(down))
            self._rolback(6)

    @Init.wait_decor
    def download_git(self):
        try:

            logging.info('Begin download from git repo')

            for f in self.f_dwnld_git:
                logging.info(
                    self.wait_mess.format('Start download {}'.format(f)))

                dwnld_url = self.url_dwnld + '/' + f
                dwnld_url = dwnld_url.replace('///', '/')
                logging.debug('Url git: {}'.format(dwnld_url))
                self._sys_call(cmd='wget -N {} -P {}'.format(dwnld_url, self.p_contr))
                sleep(0.1)
                logging.info('Download {} done.'.format(f))
            return True

        except BaseException as down:
            logging.error('Download: {}.'.format(down))
            self._rolback(6)

    @Init.wait_decor
    def unpacking(self):
        logging.info('Begin unpacking download files.')
        try:
            for f in self.f_dwnld:
                if '.tar.xz' == f[-7:]:
                    logging.info('Unpacking {}.'.format(f))

                    for k, v in self.p_unpck.items():
                        if k in f:
                            if not isdir(self.p_contr + v[0]):
                                mkdir(self.p_contr + v[0])
                            cmd = 'tar xpf {} -C {} --numeric-owner'.format(
                                self.p_contr + f, self.p_contr + v[0])
                            self._sys_call(cmd)
                            logging.info('Unpacking {} done.'.format(f))

        except BaseException as expt_unpck:
            logging.error('Unpack: {}.'.format(expt_unpck))

    @Init.wait_decor
    def unpacking_git(self):
        logging.info('Begin unpacking download files.')
        try:
            for f in self.f_dwnld_git:
                logging.info('Unpacking {}.'.format(f))

                cmd = 'tar xpf {} -C {} --numeric-owner'.format(
                    self.p_contr + f, self.p_contr)
                self._sys_call(cmd)
                logging.info('Unpacking git {} done.'.format(f))

        except BaseException as expt_unpck:
            logging.error('Unpack git: {}.'.format(expt_unpck))

    def clean(self):
        logging.info('Delete downloaded files.')
        arr_dwnld_fls = self.f_dwnld
        if self.link_from_def:
            arr_dwnld_fls += self.f_dwnld_git
        for f in arr_dwnld_fls:
            logging.info('Delete {}'.format(f))
            remove(self.p_contr + f)


class GUI(CommonCMD):
    def __init__(self):
        CommonCMD.__init__(self)

    def _prepare_icon(self):
        if environ.get('SUDO_USER'):
            logging.debug('SUDO_USER: {}'.format(environ['SUDO_USER']))
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

    @Init.wait_decor
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

                cmd = 'cd / && sudo npm install --prefix /opt/privatix/gui ' \
                      '&& sudo chown -R root:root /opt/privatix/gui'
                self._sys_call(cmd)
                self.dappctrlgui = '/opt/privatix/gui/settings.json'

                self.gui_icon_tmpl['Exec'] = self.gui_icon_tmpl[
                    'Exec'].format('')
                self.gui_icon_tmpl['Icon'] = self.gui_icon_tmpl[
                    'Icon'].format('')

            except BaseException as down:
                logging.error('Download {}.'.format(down))
                self._rolback(26)

        else:
            self.gui_icon_tmpl['Exec'] = self.gui_icon_tmpl['Exec'].format(
                self.gui_icon_prod)

            self.gui_icon_tmpl['Icon'] = self.gui_icon_tmpl['Icon'].format(
                self.gui_icon_prod)
            for cmd in self.gui_installer:
                self._sys_call(cmd, s_exit=11)

        if not isfile(self.dappctrlgui):
            logging.info(
                'The dappctrlgui package is not installed correctly')
            self._rolback(27)
        self._prepare_icon()
        self.__gui_config()

    def __gui_config(self):
        """
        RW GUI config
        /opt/privatix/gui/settings.json
        example data structure:
        {
            "firstStart": false,
            "accountCreated": true,
            "wsEndpoint": "ws://localhost:8888/ws",
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
            raw_link = raw_data['wsEndpoint'].split(delim)
            raw_link[-1] = '{}/ws'.format(self.wsEndpoint)
            raw_data['wsEndpoint'] = delim.join(raw_link)

            # add uid key in conf
            logging.debug('Add userid on settings.json')

            if raw_data.get('bugsnag'):
                raw_data['bugsnag'].update(self.uid_dict)
            else:
                raw_data['bugsnag'] = self.uid_dict

            # Rewrite settings.json
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

    @Init.wait_decor
    def __get_npm(self):
        # install npm and nodejs
        logging.debug('Get NPM for GUI.')
        npm_path = self._reletive_path(self.gui_npm_tmp_f)
        logging.debug('Npm path: {}'.format(npm_path))
        logging.debug('Npm url: {}'.format(self.gui_npm_url))
        if self.old_vers:
            logging.debug('Download node for lxc.')
            cmd = 'wget -O {} -q \'{}\''.format(npm_path, self.gui_npm_url)
            self._sys_call(cmd=cmd, s_exit=11)

        else:
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

    def update_gui(self):
        logging.info('Update GUI.')
        self.wsEndpoint = self.use_ports.get('wsEndpoint')
        if not self.use_ports.get('wsEndpoint'):
            logging.info('You can not upgrade GUI before '
                         'you not complete the installation.')
            sys.exit()

        self._clear_gui()
        self.__get_gui()


class Nspawn(Params):
    def __init__(self):
        self._init_nspwn()
        Params.__init__(self)

    def _rw_openvpn_conf(self, new_ip, new_tun, new_port, code):
        # rewrite in /var/lib/container/vpn/etc/openvpn/config/server.conf
        # two fields: server,push "route",  if ip =! default addr.
        logging.debug('Nspawn openvpn_conf')
        conf_file = "{}{}{}".format(self.p_contr,
                                    self.path_vpn,
                                    self.ovpn_conf)
        def_ip = self.addr
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

                    self.use_ports['mangmt']['common'] = self.check_port(
                        int(self.use_ports['mangmt']['vpn']) + 1, True)

                    raw_row[-1] = '{}\n'.format(
                        self.use_ports['mangmt']['vpn'])
                    tmp_data[indx] = delim.join(raw_row)
            logging.debug('--server.conf')

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

    def _nsp_ubu_pack(self):
        logging.debug('Update')
        self._sys_call('apt-get update')
        logging.debug('Install systemd-container')
        self._sys_call('apt-get install systemd-container -y')
        self._sys_call('apt-get install lshw -y')
        self._disable_dns()

    def _nsp_deb_pack(self):

        cmd = 'echo deb http://http.debian.net/debian jessie-backports main ' \
              '> /etc/apt/sources.list.d/jessie-backports.list'
        logging.debug('Add jessie-backports.list')
        self._sys_call(cmd)
        logging.debug('Update')
        self._sys_call('apt-get update')
        self._sys_call('apt-get install lshw -y')
        self.__upgr_sysd(
            cmd='apt-get -t jessie-backports install systemd -y')

        logging.debug('Install systemd-container')
        self._sys_call('apt-get install systemd-container -y')


class LXC(DB):
    def __init__(self):
        self._init_lxc()
        DB.__init__(self)

    def conf_dappvpn_json(self):
        """Check addr in vpn dappvpn.config.json"""
        logging.debug('Check addr in vpn dappvpn.config.json')
        search_keys = ['Monitor', 'Sess']
        delim = ":"
        for cont_path in (self.path_com, self.path_vpn):

            p = self.p_contr + cont_path + 'rootfs/' + self.p_dapvpn_conf

            # Read dappctrl.config.local.json
            data = self.file_rw(p=p, json_r=True, log='Read dappvpn conf')
            if not data:
                self._rolback(22)

            serv_addr = data.get('Sess').get('Endpoint')
            # "Endpoint": "ws://localhost:8000/ws"
            if serv_addr:
                raw = serv_addr.split(delim)
                raw[0] = self.p_unpck['common'][1]
                data['Sess']['Endpoint'] = delim.join(raw)
            else:
                logging.error('Field Sess not exist')

            monit_addr = data.get('Monitor').get('Addr')
            if monit_addr:
                raw = monit_addr.split(delim)
                raw[0] = self.p_unpck['vpn'][1]
                data['Monitor']['Addr'] = delim.join(raw)
            else:
                logging.error('Field Monitor not exist')

            # Rewrite dappvpn.config.json
            self.file_rw(p=p, w=True, json_r=True, data=data,
                         log='Rewrite conf')

    def _check_cont_addr(self):
        # Check if 0,0,ххх,51 & 10,0,ххх,52
        # If not, increment 4-th octet and check again
        def increment_octet(octet, all_ip):
            if octet in all_ip:
                octet = randint(5, 254)
                increment_octet(octet, all_ip)
            return octet

        found_contrs_ip = [v.get('lxc.network.ipv4', '...').split('.')
                           for k, v in self.lxc_contrs.items()]

        found_contrs_ip = [int(ip[3])
                           for ip in found_contrs_ip
                           if ip and ip[3].isdigit()]

        for name, addr in self.p_unpck.items():
            octet = increment_octet(int(addr[1]), found_contrs_ip)
            self.p_unpck[name][1] = str(octet)

    def _rw_openvpn_conf(self, code):
        # rewrite in /var/lib/lxc/vpn/rootfs/etc/openvpn/config/server.conf
        # management field
        logging.debug('Lxc openvpn_conf')
        conf_file = "{}{}{}{}".format(self.p_contr,
                                      self.path_vpn,
                                      'rootfs/',
                                      self.ovpn_conf)
        try:
            # read a list of lines into data
            tmp_data = self.file_rw(
                p=conf_file,
                log='Read openvpn server.conf'
            )

            # replace all search fields
            for row in tmp_data:

                if 'management' in row:
                    logging.debug(
                        'Rewrite management: {}'.format(row))
                    indx = tmp_data.index(row)
                    row = row.split(' ')
                    row[1] = self.p_unpck['vpn'][1]

                    tmp_data[indx] = ' '.join(row)

            logging.debug('--server.conf')

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

    @Init.wait_decor
    def __install_lxc(self):
        logging.debug('Install lxc')

        for cmd in self.lxc_install:
            if not self._sys_call(cmd=cmd, rolback=False):
                logging.error('Error when try: {}.'.format(cmd))
                sys.exit(29)

    def __change_mac(self, macs):
        mac = "00:16:3e:%02x:%02x:%02x" % (
            randint(0, 255),
            randint(0, 255),
            randint(0, 255)
        )

        if mac in macs:
            self.__change_mac(macs)
        return mac

    def __check_mac(self, mac):
        hwaddrs = []
        for cont, data in self.lxc_contrs.items():
            hwaddrs.append(data.get('hwaddr'))

        if mac in hwaddrs:
            mac = self.__change_mac(hwaddrs)
        return mac

    def _rw_container_run_sh(self):
        logging.debug('LXC cont name: {}'.format(self.p_unpck))

        for f_name in self.f_dwnld_git:
            try:
                if not '.tar.xz' == f_name[-7:]:
                    logging.info('Rewrite {} run file.'.format(f_name))
                    for target_name, cont_name in self.p_unpck.items():
                        if target_name in f_name:
                            conf_file = self.deff_lxc_cont_path + f_name
                            # Read run sh file
                            tmp_data = self.file_rw(
                                p=conf_file,
                                log='Read LXC {} run file'.format(f_name)
                            )
                            for row in tmp_data:
                                indx = tmp_data.index(row)
                                if 'CONTAINER_NAME=' in row:
                                    tmp_data[
                                        indx] = 'CONTAINER_NAME={}\n'.format(
                                        cont_name[0])
                                elif 'VPN_PORT=' in row:
                                    tmp_data[indx] = 'VPN_PORT={}\n'.format(
                                        self.use_ports['vpn'])

                            # rewrite run sh file
                            if not self.file_rw(
                                    p=conf_file,
                                    w=True,
                                    data=tmp_data,
                                    log='Rewrite LXC {} run file'.format(
                                        f_name)
                            ):
                                self._rolback(7)

                            del tmp_data
                            dest_run_sh_path = self.run_sh_path + f_name
                            logging.debug(
                                'LXC {} run file done'.format(f_name))
                            copyfile(conf_file, dest_run_sh_path)
                            logging.debug(
                                'LXC {} run file copy done'.format(f_name))
                            cmd = self.chmod_run_sh.format(dest_run_sh_path)
                            self._sys_call(cmd)
                            logging.debug(
                                'LXC {} run file chown done'.format(
                                    dest_run_sh_path))

            except BaseException as f_rw:
                logging.error('R/W LXC run sh: {}'.format(f_rw))
                self._rolback(32)

    @Init.wait_decor
    def _rw_container_intrfs(self):
        logging.debug('LXC containers: {}'.format(self.name_in_main_conf))
        for target_name, cont_name in self.p_unpck.items():
            try:
                conf_file = self.deff_lxc_cont_path + cont_name[
                    0] + self.lxc_cont_interfs
                # Read conf file
                tmp_data = self.file_rw(
                    p=conf_file,
                    log='Read LXC {} interfaces'.format(cont_name[0])
                )
                for row in tmp_data:
                    indx = tmp_data.index(row)

                    if 'address' in row:
                        tmp_data[indx] = 'address {}\n'.format(cont_name[1])
                    elif 'gateway' in row:
                        tmp_data[indx] = 'gateway {}\n'.format(
                            self.name_in_main_conf['LXC_ADDR='])
                    elif 'network' in row:
                        newrk = \
                            self.name_in_main_conf['LXC_NETWORK='].split(
                                '/')[0]
                        tmp_data[indx] = 'network {}\n'.format(newrk)

                # rewrite conf file
                if not self.file_rw(
                        p=conf_file,
                        w=True,
                        data=tmp_data,
                        log='Rewrite LXC {} interfaces'.format(cont_name)
                ):
                    self._rolback(7)

                del tmp_data

                logging.debug('LXC {} interfaces done'.format(cont_name))
                self._sys_call(self.update_cont_conf.format(conf_file))

            except BaseException as f_rw:
                logging.error('R/W LXC interfaces : {}'.format(f_rw))
                self._rolback(32)

    def _composit_addr(self, last_octet):
        addr = self.name_in_main_conf['LXC_ADDR='].split('.')
        addr[3] = last_octet
        addr = '.'.join(addr)
        return addr

    def _rw_psql_conf(self):
        logging.debug('Begin check DB configs')
        self.db_conf_path = self.db_conf_path.format(self.path_com)

        db_configs = {
            'pg_hba.conf':
                dict(
                    l_from=self.addr + "/24",
                    l_to=self.name_in_main_conf['LXC_NETWORK=']
                ),
            'postgresql.conf':
                dict(
                    l_from=self.def_comm_addr,
                    l_to=self.p_unpck['common'][1]
                )
        }
        logging.debug('Containers: {}'.format(self.p_unpck))
        logging.debug('Name in lxc conf: {}'.format(self.name_in_main_conf))
        logging.debug('DB configs: {}'.format(db_configs))
        for p_conf, fields in db_configs.items():
            p = self.db_conf_path + p_conf
            raw_data = self.file_rw(p=p, log='Read {} conf'.format(p_conf))
            for row in raw_data:
                if fields['l_from'] in row:
                    indx = raw_data.index(row)
                    raw_data[indx] = row.replace(fields['l_from'],
                                                 fields['l_to'])
            self.file_rw(p=p, log='Write db conf', w=True, data=raw_data)

    def _rw_container_conf(self):
        for target_name, cont_name in self.p_unpck.items():
            logging.debug('Rewrite {} conf'.format(cont_name[0]))
            try:
                conf_file = self.deff_lxc_cont_path + cont_name[
                    0] + self.lxc_cont_conf_name
                # Read conf file
                tmp_data = self.file_rw(
                    p=conf_file,
                    log='Read LXC {} config'.format(cont_name[0])
                )
                ipv4_indx = False
                for row in tmp_data:
                    indx = tmp_data.index(row)

                    if 'lxc.rootfs.path' in row:
                        tmp_data[
                            indx] = 'lxc.rootfs.path = dir:{}{}rootfs\n'.format(
                            self.deff_lxc_cont_path, cont_name[0])
                    elif 'lxc.uts.name' in row:
                        cont_name[0] = cont_name[0][0:-1]
                        tmp_data[indx] = 'lxc.uts.name = {}\n'.format(
                            cont_name[0])
                    elif 'lxc.net.0.hwaddr' in row:
                        hwaddr = row.split('=')[1].strip()
                        hwaddr = self.__check_mac(hwaddr)
                        tmp_data[indx] = 'lxc.net.0.hwaddr = {}\n'.format(
                            hwaddr)
                    elif 'lxc.net.0.ipv4.gateway' in row:
                        tmp_data[
                            indx] = 'lxc.net.0.ipv4.gateway = {}\n'.format(
                            self.name_in_main_conf['LXC_ADDR='])
                    elif 'lxc.net.0.ipv4.address' in row:
                        ipv4_indx = indx

                addr = self._composit_addr(str(cont_name[1]))
                self.p_unpck[target_name][1] = addr

                raw_ip_line = 'lxc.net.0.ipv4.address = {}/24\n'.format(addr)

                if ipv4_indx:
                    tmp_data[ipv4_indx] = raw_ip_line
                else:
                    tmp_data.append(raw_ip_line)

                # rewrite conf file
                if not self.file_rw(
                        p=conf_file,
                        w=True,
                        data=tmp_data,
                        log='Rewrite LXC {} config'.format(cont_name)
                ):
                    self._rolback(7)

                del tmp_data

                logging.debug('LXC {} config done'.format(cont_name))
                self._sys_call(self.update_cont_conf.format(conf_file))

            except BaseException as f_rw:
                logging.error('R/W LXC config : {}'.format(f_rw))
                self._rolback(32)

    def __check_contrs_by_cmd(self):
        logging.debug('Check containers by cmd')
        beg_line = 'NAME'
        pattern = r'^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$'

        raw = self._sys_call(cmd=self.exist_contrs, rolback=False)
        if raw:
            raw = raw.split('\n')
            marker = False
            for i in range(len(raw)):
                if raw[i].startswith(beg_line):
                    marker = i
                    break
            if len(raw) - 1 == marker:
                logging.debug('Containers are absent')
            else:
                for line in raw[marker + 1:]:
                    if line:
                        raw_line = line.split(' ')
                        for i in raw_line:
                            if i and i[-1] == ',':
                                i = i[:-1]
                            if i and match(pattern, i):
                                logging.debug(
                                    'Found new container: {}'.format(
                                        raw_line[0]))
                                if not self.lxc_contrs.get(raw_line[0]):
                                    self.lxc_contrs[raw_line[0]] = {
                                        'lxc.network.ipv4': i}
        else:
            logging.debug('Containers are absent')

    def _check_contrs_by_path(self, update_data=False):
        list_dir = listdir(self.deff_lxc_cont_path)
        folders = []
        for f in list_dir:
            if isdir(self.deff_lxc_cont_path + f):
                config_path = '{}{}'.format(
                    self.deff_lxc_cont_path,
                    f,
                )
                folders.append(config_path)
        # todo add folder where may be stored user containers
        # folders.append('')


        logging.debug('Check by path: {}'.format(folders))
        # read config file in container.Get data from it
        for folder in folders:
            tmp_data_name = None
            conf_path = '{}/{}'.format(folder, self.lxc_cont_conf_name)
            logging.debug('Conf in path: {}'.format(conf_path))
            if isfile(conf_path):
                res = self.file_rw(p=conf_path, log='Read container conf')
                tmp_data = {}
                for line in res:
                    for v in self.name_in_contnr_conf:
                        if v in line:
                            clear_line = line.split('=')[1].strip()
                            if v == 'lxc.network.ipv4':
                                clear_line = clear_line.split('/')[0]
                            tmp_data[v] = clear_line

                tmp_data_name = tmp_data['lxc.uts.name']
                del tmp_data['lxc.uts.name']
                if tmp_data_name in self.lxc_contrs:
                    self.lxc_contrs[tmp_data_name].update(tmp_data)
                else:
                    self.lxc_contrs[tmp_data_name] = tmp_data

            # exist file /rootfs/home/ubuntu/go/bin/dappctrl
            # search for the dappctrl file, by which we determine that
            # this is our container
            logging.debug(
                'Check container: {}'.format(folder + self.lxc_cont_fs_file))
            if isfile(folder + self.lxc_cont_fs_file) and tmp_data_name:
                logging.debug('Our container: {}'.format(tmp_data_name))
                # check what kind of container it is, vpn or common
                for c_name, c_path in self.kind_of_cont.items():
                    logging.debug('Check exist: {}'.format(folder + c_path))
                    if exists(folder + c_path):
                        logging.debug(
                            'Path: {} are exist'.format(folder + c_path))
                        if update_data:
                            self.p_unpck[c_name].append(
                                folder.split('/')[-1])
                            logging.debug(
                                'Update: {}'.format(self.p_unpck[c_name]))
                        else:
                            logging.info(
                                '\nWe found in your mashine installed our,\n '
                                '`Privatix` container: {} ,'
                                'was called in lxc as: {}.\nPlease remove it, '
                                'and repeat instalation.\n'
                                'Or re-run initializer in update mode!'.format(
                                    c_name, tmp_data_name)
                            )
                            sys.exit(31)
            else:
                logging.debug('Our container is absend')

    def __read_lxc_conf(self):
        # Check if exist /etc/default/lxc-net and 'LXC_BRIDGE' in it
        if isfile(self.bridge_conf):
            # conf exist, check it and search key from name_in_main_conf
            raw = self.file_rw(p=self.bridge_conf, log='Check bridge')
            if raw:
                for row in raw:
                    for k in self.name_in_main_conf:
                        if search('^' + k, row):
                            self.name_in_main_conf[k] = sub('"|{|}|\n', '',
                                                            row.split(k)[1])
                            logging.info(
                                'Found the bridge in the config: '
                                '{}{}'.format(k, self.name_in_main_conf[k]))

    def __check_lxc_exist(self):
        # Check if exist lxcbr bridge
        raw = self._sys_call(self.bridge_cmd)
        raw_arr = compile("\n\d:").split(raw)
        for row in raw_arr:
            if self.search_name in row:
                self.name_in_main_conf['LXC_BRIDGE='] = row.split(':', 1)[0]

        self.__read_lxc_conf()
        if self.name_in_main_conf['LXC_BRIDGE=']:
            logging.info('LXC already installed on computer')
            self._lxc_exist = True
            return True
        else:
            logging.info('LXC is not installed on computer.Installing it.')
            self._lxc_exist = False
            return False

    def __rename_cont_path(self):
        similar_contr = set(self.lxc_contrs.keys()) & set(
            self.kind_of_cont.keys())
        logging.debug('Rename path. Similar: {}'.format(similar_contr))
        for name in similar_contr:
            if name in self.path_vpn:
                self.path_vpn = 'dapp' + self.path_vpn
                self.p_unpck['vpn'][0] = self.path_vpn
            elif name in self.path_com:
                self.path_com = 'dapp' + self.path_com
                self.p_unpck['common'][0] = self.path_com

        self.db_log = self.db_log.format(self.path_com)
        logging.debug('Rename container path: {}'.format(self.p_unpck))

    def __check_wget(self):
        cmd = main_conf['search_pack'].format('wget')
        raw = self._sys_call(cmd=cmd)
        if not raw:
            logging.info('Install wget.')
            self._sys_call(cmd='sudo apt install wget')

    def _lxc_ubu_pack(self):
        self.__check_wget()
        if self.__check_lxc_exist():
            # lxc installed
            self.__check_contrs_by_cmd()
            self._check_contrs_by_path()
            logging.debug('Found LXC conteiners: {}'.format(self.lxc_contrs))
            self.__rename_cont_path()
            self._check_cont_addr()

        else:
            # lxc not installed
            self.__install_lxc()
            # update to new config params
            self.__read_lxc_conf()


class AutoOffer:
    def __init__(self):
        self.id = 1
        self.url = 'http://localhost:8888/http'
        self.pswdSymbol = 12
        self.acc_name = 'TestAcc'
        self.botUrl = 'http://89.38.96.53:3000/getprix'
        self.botAuth = 'dXNlcjpoRmZWRWRVMkNva0Y='
        self.offerData = None
        self.pswd = None
        self.token = None
        self.ethAddr = None
        self.prixHash = None
        self.agent_id = None  # id of account to be created.
        self.product_id = None  # id of product on config /var/lib/container/vpn/opt/privatix/config/dappvpn.config.json  Product.
        self.ethHash = None
        self.ptcBalance = None
        self.pscBalance = None
        self.ethBalance = None
        self.offer_id = None
        self.gasPrice = 2000000000
        self.waitBot = 1
        self.waitblockchain = 90
        self.vpnConf = '/var/lib/container/vpn/opt/privatix/config/dappvpn.config.json'

    def _getAgentOffer(self, mark):
        logging.info('Get Offerings. Mark: {}'.format(mark))
        # Get Offerings For Agent
        data = {
            'method': 'ui_getAgentOfferings',
            'params': [
                self.token,
                self.product_id,
                'registered',
                0,
                1
            ],
            'id': self.id,
        }
        timeWait = 25 * 60
        timeStar = time()
        while time() - timeStar < timeWait:
            res = self.__urlOpen(data=data, key='result')
            if res[0]:
                items = res[1].get('items')
                logging.debug("items: {}".format(items))
                if items and isinstance(items, (list, set, tuple)):
                    # status = items[0].get('status')
                    offerStatus = items[0].get('offerStatus')
                    logging.debug('offerStatus: {}'.format(offerStatus))
                    if offerStatus == 'registered':
                        logging.debug('Offerings for agent exist.')
                        return True, 'All done'
            logging.debug('Wait')
            sleep(60)
        logging.info('Does not exist offerings for agent.')
        if not mark:
            logging.info('Try again.')
            return self._statusOffer(mark=True)
        return False, res[1]

    def __getProductId(self):
        logging.info('Get Product Id')
        if isfile(self.vpnConf):
            try:
                f = open(self.vpnConf)
                raw_data = loads(f.read())
                logging.debug('Read vpn conf: {}'.format(raw_data))
                self.product_id = raw_data['Sess']['Product']
                logging.debug('Product id: {}'.format(self.product_id))
                return True, 'Product id was found'

            except BaseException as readexpt:
                logging.error('Read vpn conf: {}'.format(readexpt))
                return False, readexpt

        return False, 'there is no {} to determine Product Id'.format(
            self.vpnConf)

    def __checkOfferData(self):
        logging.debug('Check offer data')
        params = {
            "product": self.product_id,
            "template": "efc61769-96c8-4c0d-b50a-e4d11fc30523",
            "agent": self.agent_id,
            "serviceName": "my service",
            "description": "my service description",
            "country": self.__getCountryName(),
            "supply": 30,
            "unitName": "MB",
            "autoPopUp": True,
            "unitType": "units",
            "billingType": "postpaid",
            "setupPrice": 0,
            "unitPrice": 1000,
            "minUnits": 10000,
            "maxUnit": 30000,
            "billingInterval": 1,
            "maxBillingUnitLag": 100,
            "maxSuspendTime": 1800,
            "maxInactiveTimeSec": 1800,
            "freeUnits": 0,
            "additionalParams": {"minDownloadMbits": 100,
                                 "minUploadMbits": 80},
        }

        if self.offerData:
            logging.debug('From file: {}'.format(self.offerData))
            res = self.offerData.get('country')
            if res:
                if not params['country'].lower() == res.lower():
                    logging.info('You country name {} from config : {}\n'
                                 'does not match with country calculated '
                                 'by your IP : {}\n.'.format(
                        res, self.in_args['file'], params['country']))
                    logging.info('Choose which country to use 1 or 2:\n'
                                 '1 - {}\n'
                                 '2 - {}'.format(res, params['country']))
                    choise_task = {1: res, 2: params['country']}

                    while True:
                        choise_code = raw_input('>')

                        if choise_code.isdigit() and int(
                                choise_code) in choise_task:
                            self.offerData['country'] = choise_task[
                                int(choise_code)]
                            break
                        else:
                            logging.info(
                                'Wrong choice. Make a choice between: '
                                '{}'.format(choise_task.keys()))

            params.update(self.offerData)

        else:
            logging.info(
                'You file offer is empty.Install with default params')

        return params

    def validateJson(self, path):
        logging.info('Checking JSON')
        try:
            with open(path) as f:
                try:
                    self.offerData = load(f)
                    logging.debug('Json is valid: {}'.format(self.offerData))
                    return True, 'Json is valid'
                except ValueError as e:
                    logging.error('Read file: {}'.format(e))
                    return False, 'This is not json format.' \
                                  'Perhaps you are using a single quote \', ' \
                                  'instead of a double quote ".Check your structure.'
        except BaseException as oexpt:
            logging.error('Open file: {}'.format(oexpt))
            return False, 'Trouble when try open file: {}. Error: {}'.format(
                path, oexpt)

    @Init.wait_decor
    def republishOffer(self):
        logging.debug('Republish')
        res = self.__getProductId()
        if res[0]:
            res = self._getAcc()
            logging.debug('Get Acc: {}'.format(res))
            if res[0]:

                self.ethAddr = res[1][0]['ethAddr']
                self.ptcBalance = res[1][0]['ptcBalance']
                self.pscBalance = res[1][0]['pscBalance']
                self.ethBalance = res[1][0]['ethBalance']
                self.agent_id = res[1][0]['id']

                logging.debug('ethAddr: {}'.format(self.ethAddr))
                logging.debug('agent_id: {}'.format(self.agent_id))
                logging.debug('ethBalance: {}'.format(self.ethBalance))
                logging.debug('pscBalance: {}'.format(self.pscBalance))
                logging.debug('ptcBalance: {}'.format(self.ptcBalance))
                res = self._askBot()
                if not res[0]:
                    return res

                self._wait_blockchain(target='ptc', republ=True)
                res = self._transfer()
                if not res[0]:
                    return res
                self._wait_blockchain(target='psc', republ=True)
                res = self._createOffer()

                logging.debug('ethAddr: {}'.format(self.ethAddr))
                logging.debug('agent_id: {}'.format(self.agent_id))
                logging.debug('ethBalance: {}'.format(self.ethBalance))
                logging.debug('pscBalance: {}'.format(self.pscBalance))
                logging.debug('ptcBalance: {}'.format(self.ptcBalance))
                logging.debug('product_id: {}'.format(self.product_id))
                logging.debug('ethHash: {}'.format(self.ethHash))
                logging.debug('offer_id: {}'.format(self.offer_id))
                logging.debug('prixHash: {}'.format(self.prixHash))
                logging.debug('gasPrice: {}'.format(self.gasPrice))
                if not res[0]:
                    return res
                return self._statusOffer()
        return res

    @Init.wait_decor
    def offerRun(self):
        res = self.__getProductId()
        if not res[0]:
            return res
        res = self._setPswd()
        if res[0]:
            logging.debug('Eth addr: {}'.format(self.ethAddr))
            logging.debug('Generate acc id: {}'.format(self.agent_id))

            res = self._askBot()
            if not res[0]:
                return res

            self._wait_blockchain(target='ptc')
            res = self._transfer()
            if not res[0]:
                return res
            self._wait_blockchain(target='psc')
            res = self._createOffer()
            logging.debug('product_id: {}'.format(self.product_id))
            logging.debug('offer_id: {}'.format(self.offer_id))
            logging.debug('ethBalance: {}'.format(self.ethBalance))
            logging.debug('pscBalance: {}'.format(self.pscBalance))
            logging.debug('ptcBalance: {}'.format(self.ptcBalance))
            logging.debug('ethHash: {}'.format(self.ethHash))
            logging.debug('agent_id: {}'.format(self.agent_id))
            logging.debug('prixHash: {}'.format(self.prixHash))
            logging.debug('ethAddr: {}'.format(self.ethAddr))
            logging.debug('gasPrice: {}'.format(self.gasPrice))
            if not res[0]:
                return res
            return self._statusOffer()

        else:
            return res

    def __getCountryName(self):
        country = None
        try:
            ip = urlopen('http://icanhazip.com').read()
            raw_data = urlopen(
                'http://ipinfo.io/{}'.format(ip)).read()
            country = loads(raw_data)['country']
        except BaseException as cntr:
            logging.debug('Error when try get cuntry: {}'.format(cntr))
            country = raw_input(
                prompt='Please enter your country name, abbreviated. For example US.')
        finally:
            return country

    def _statusOffer(self, mark=False):
        logging.info('Offering status')
        data = {
            'method': 'ui_changeOfferingStatus',
            'params': [
                self.token,
                self.offer_id,
                'publish',
                self.gasPrice,

            ],
            'id': self.id,
        }
        res = self.__urlOpen(data=data)
        if res[0]:
            return self._getAgentOffer(mark)
        else:
            return False, res[1]

    def _createOffer(self):
        logging.info('Offering create')
        data = {
            'method': 'ui_createOffering',
            'params': [
                self.token,
                self.__checkOfferData()
            ],
            'id': self.id,
        }

        res = self.__urlOpen(data=data, key='result')
        if res[0]:
            self.offer_id = res[1]
            return True, res[1]
        else:
            return False, res[1]

    def _wait_blockchain(self, target, republ=False):
        logging.info('Wait blockchain.Target is: {}.'.format(target))
        waitCounter = 0
        while True:
            sleep(self.waitblockchain)
            res = self._getEth()
            logging.debug('Wait {} min'.format(waitCounter))
            waitCounter += 1
            if res[0]:
                if target == 'ptc' and int(res[1].get('ptcBalance', '0')):
                    if republ and int(self.ptcBalance) >= int(
                            res[1]['ptcBalance']):
                        continue
                    self.ptcBalance = res[1]['ptcBalance']
                    self.ethBalance = res[1]['ethBalance']
                    break
                elif target == 'psc' and int(res[1].get('pscBalance', '0')):
                    if republ and int(self.pscBalance) >= int(
                            res[1]['pscBalance']):
                        continue
                    self.pscBalance = res[1]['pscBalance']
                    self.ptcBalance = res[1]['ptcBalance']
                    self.ethBalance = res[1]['ethBalance']
                    break

    def _transfer(self):
        # Transfer some PRIX from PTC balance to PSC balance
        logging.info('Transfer PRIX')
        data = {
            'method': 'ui_transferTokens',
            'params': [
                self.token,
                self.agent_id,
                'psc',
                self.ptcBalance,
                self.gasPrice
            ],
            'id': self.id,
        }

        res = self.__urlOpen(data)
        if res[0]:
            return True, 'Ok'
        else:
            return False, res[1]

    def __urlOpen(self, data, key=None, url=None, auth=None):
        try:
            url = self.url if not url else url
            logging.debug('Request: {}'.format(data))
            request = Request(url)
            request.add_header('Content-Type', 'application/json')
            if auth:
                request.add_header('Authorization', "Basic {}".format(auth))
            response = urlopen(request, dumps(data))
            response = response.read()
            logging.debug('Response: {0}'.format(response))
            try:
                response = loads(response)
            except BaseException as jsonExpt:
                logging.error(jsonExpt)
                return False, jsonExpt

            if response.get('error', False):
                logging.error(
                    "Error on response: {}".format(response['error']))

                return False, response['error']
            if key:
                logging.debug('Get key: {}'.format(key))
                response = response.get(key, False)
                if not response:
                    logging.error('Key {} not exist in response'.format(key))
                    return False, 'Key {} not exist in response'.format(key)

            logging.debug('Response OK: {}'.format(response))
            return True, response

        except BaseException as urlexpt:
            logging.error('Url Exept: {}'.format(urlexpt))
            return False, urlexpt

    def _setPswd(self):
        # 1.Set password for UI API access
        logging.info('Set password')

        data = {
            'method': 'ui_setPassword',
            'params': [self.pswd],
            'id': self.id,
        }

        res = self.__urlOpen(data)
        if res[0]:
            return self._getTok()
        else:
            return False, res[1]

    def _getTok(self):
        # Given paswd and returns new access token.
        logging.info('Get token')

        data = {
            'method': 'ui_getToken',
            'params': [self.pswd],
            'id': self.id,
        }
        res = self.__urlOpen(data=data, key='result')
        if res[0]:
            self.token = res[1]
            logging.debug('Token: {}'.format(self.token))
            return self._createAcc()
        else:
            return False, res[1]

    def _getAcc(self):
        # Create account
        '''curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getAccounts", "params": ["qwert"], "id": 67}' http://localhost:8888/http

        {"jsonrpc":"2.0",
        "id":67,
        "result":[{"id":"25b84988-fde6-4882-91bc-ab1dfb86cbdb","ethAddr":"35218b6fc288e093d55295e2c3ce7304d216be64","isDefault":true,"inUse":true,"name":"TestAcc","ptcBalance":0,"pscBalance":700000000,"ethBalance":49654270000000000,"lastBalanceCheck":"2018-11-29T14:24:41.730652+01:00"}]}

        '''
        logging.info('Get account')
        data = {
            'method': 'ui_getAccounts',
            'params': [self.token],
            'id': self.id,
        }
        res = self.__urlOpen(data=data, key='result')
        return res

    def _createAcc(self):
        # Create account
        logging.info('Create account')
        data = {
            'method': 'ui_generateAccount',
            'params': [
                self.token,
                {
                    'name': self.acc_name,
                    'isDefault': True,
                    'inUse': True,
                }
            ],
            'id': self.id,
        }
        res = self.__urlOpen(data=data, key='result')
        if res[0]:
            if isinstance(res[1], unicode):
                self.agent_id = res[1].encode()
            else:
                self.agent_id = res[1]

            return self._getEth()
        else:
            return False, res[1]

    def _getEth(self):
        # Get ethereum address of newly created account
        logging.debug('Get ethereum address')
        data = {
            'method': 'ui_getObject',
            'params': [
                self.token,
                'account',
                self.agent_id,
            ],
            'id': self.id,
        }
        res = self.__urlOpen(data=data, key='result')
        if res[0]:
            self.ethAddr = res[1]['ethAddr']
            return res
        else:
            return False, res[1]

    def _askBot(self):
        # Ask Privatix bot to transfer PRIX and ETH to address of account in
        logging.info('Ask Privatix Bot')
        stop_mark = 0

        data = {
            'address': '0x{}'.format(self.ethAddr),
        }
        while True:
            res = self.__urlOpen(data=data, url=self.botUrl,
                                 auth=self.botAuth)
            if res[0]:
                if res[1].get('code') and res[1]['code'] == 200:
                    self.prixHash = res[1].get('prixHash')
                    self.ethHash = res[1].get('ethHash')
                    if self.prixHash and self.ethHash:
                        return True, 'OK'
                    logging.debug(
                        'prixHash:{}, ethHash:{}'.format(self.prixHash,
                                                         self.ethHash))
                else:
                    logging.error('Bot error: {}'.format(res[1]))
                stop_mark += 1
                sleep(self.waitBot)
            else:
                logging.error('Bot not answer: {}'.format(res[1]))
                return False, res[1]

            if stop_mark > 5:
                logging.error('Error when try ask to bot')
                return False, res[1]


def checker_fabric(inherit_class, old, v, dist):
    class Checker(Rdata, GUI, AutoOffer, inherit_class):
        def __init__(self, old_v, ver, dist_name):
            GUI.__init__(self)
            AutoOffer.__init__(self)
            Rdata.__init__(self)
            inherit_class.__init__(self)
            self.old_vers = old_v
            self.ver = ver
            self.dist_name = dist_name
            self.firstinst = None
            self.pswd = None
            self.link_from_def = True

        def __ubuntu(self):
            logging.debug('Ubuntu: {}'.format(self.ver))
            v = int(self.ver.split('.')[0])
            if v >= 16:
                logging.debug('--- Nspawn ---')
                self._nsp_ubu_pack()

            elif v >= 14:
                logging.debug('--- LXC ---')
                self._lxc_ubu_pack()

            else:
                logging.error('Your version of Ubuntu is not suitable. '
                              'It is not supported by the program')
                sys.exit(2)

        def __debian(self):
            logging.debug('Debian: {}'.format(self.ver))
            self._nsp_deb_pack()

        @Init.wait_decor
        def __check_os(self):
            self.task = dict(ubuntu=self.__ubuntu,
                             debian=self.__debian
                             )
            task_os = self.task.get(self.dist_name.lower(), False)
            if not task_os:
                logging.error('You system is {}.'
                              'She is not supported yet'.format(
                    self.dist_name))
                sys.exit(19)
            task_os()

            self.sysctl = self._sysctl() if not self.old_vers else True

        def init_os(self, update=False):

            if self._finalizer(pass_check=update):
                if not isfile(self._reletive_path(self.build_cmd_path)):
                    logging.info(
                        'There is no .dapp_cmd file for further work.\n'
                        'To create it, you must run '
                        './initializer.py --build')
                    logging.debug(self._reletive_path(self.build_cmd_path))
                    sys.exit(28)

                self.__check_os()

                if self.link_from_def:
                    # use default link for download from git by latest tag
                    self.get_latest_tag()

                if not self.old_vers:
                    ip, intfs, tun, port = self._iptables()

                    self.download_git() if self.link_from_def else self.download()
                    try:
                        if self.link_from_def:
                            self.unpacking_git()
                        self.unpacking()
                        self._rw_openvpn_conf(ip, tun, port, 7)
                        self._rw_unit_file(ip, intfs, 5)
                        self.clean()
                        self._clear_db_log()
                        self.conf_dappctrl_json()
                        self.check_tor_port()
                        self.run_service(comm=True)
                        self._check_db_run(9)

                        if self.in_args['test']:
                            logging.info('Test mode.')
                            self._test_mode()
                        else:
                            logging.info('Full mode.')
                            self._run_dapp_cmd()
                            self._check_dapp_conf()
                        if self.dappctrl_role == 'agent':
                            self.get_onion_key()
                        else:
                            self.set_socks_list()
                        self.run_service(comm=True, restart=True)
                        self.run_service()
                        if not self.in_args['no_gui']:
                            self.target = 'both'
                            logging.info('GUI mode.')
                            check.target = 'both'
                            if update:
                                self.update_gui()
                            else:
                                self.install_gui()
                        elif self.in_args['cli']:
                            logging.debug('Cli mode.Wait when up 8000 port')
                            if not self._checker_port(
                                    port='8000',
                                    verb=True):
                                logging.info(
                                    'Sorry, but for unknown reasons,\n'
                                    'the required service to continue work is not responding.\n'
                                    'Try again.')
                                self._rolback(35)
                            res = self.offerRun()
                            if not res[0]:
                                logging.error(
                                    'Auto offer: {}'.format(res[1]))
                                raise BaseException(res[1])
                            mess = '    Congratulations, you posted your offer!\n' \
                                   '    It will be published once an hour.\n' \
                                   '    Your ethereum address: 0x{}\n' \
                                   '    Your pasword : {}\n' \
                                   '    Please press enter to finalize the application.'.format(
                                self.ethAddr, self.pswd)
                            raw_input(mess)
                        self._finalizer(rw=True)
                    except BaseException as mexpt:
                        logging.error('Main trouble: {}'.format(mexpt))
                        self._rolback(17)
                else:
                    port = findall('\d+', self.ovpn_port[0])[0]
                    self.use_ports['vpn'] = self.check_port(port)
                    self.download_git() if self.link_from_def else self.download()
                    try:
                        if self.link_from_def:
                            self.unpacking_git()
                        self.unpacking()
                        self._rw_container_conf()
                        self._rw_openvpn_conf(7)
                        self._rw_psql_conf()
                        self._rw_container_intrfs()
                        self._rw_container_run_sh()
                        self.clean()
                        self._clear_db_log()
                        self.conf_dappctrl_json()
                        self.conf_dappvpn_json()

                        self.run_service(comm=True)
                        self._check_db_run(9)

                        self.run_service(comm=False)
                        logging.debug('Containers: {}'.format(self.p_unpck))
                        logging.debug('Ports: {}'.format(self.use_ports))
                        if not self.in_args['no_gui']:
                            logging.info('GUI mode.')
                            self.target = 'both'
                            if update:
                                self.update_gui()
                            else:
                                self.install_gui()
                        elif self.in_args['cli']:
                            logging.debug('Cli mode.Wait when up 8000 port')
                            if not self._checker_port(
                                    port='8000',
                                    verb=True):
                                logging.info(
                                    'Sorry, but for unknown reasons,\n'
                                    'the required service to continue work is not responding.\n'
                                    'Try again.')
                                self._rolback(35)
                            res = self.offerRun()
                            if not res[0]:
                                logging.error(
                                    'Auto offer: {}'.format(res[1]))
                                raise BaseException(res[1])
                            mess = '    Congratulations, you posted your offer!\n' \
                                   '    It will be published once an hour.\n' \
                                   '    Your ethereum address: 0x{}\n' \
                                   '    Please press enter to finalize the application.'.format(
                                self.ethAddr)
                            raw_input(mess)
                        self._finalizer(rw=True)
                    except BaseException as mexpt:
                        logging.error('Main trouble: {}'.format(mexpt))
                        self._rolback(17)

        def prompt(self, mess, choise=('N', 'Y')):
            logging.info(mess)

            answ = raw_input('>')

            while True:
                if answ.upper() not in choise:
                    logging.info('Invalid choice. Select {}.'.format(choise))
                    answ = raw_input('> ')
                    continue

                logging.debug('Choise {}'.format(choise[1]))
                if answ.upper() == choise[1]:
                    return True
                return False

        def check_graph(self):
            if not isdir(self.gui_icon_path) and not self.in_args['no_gui']:
                mess = 'You chosen a full installation with a GUI,\n' \
                       'but did not find a GUI on your computer.\n' \
                       'Y - I understand. Continue the installation but without GUI\n' \
                       'N - Stop the installation.'

                if self.prompt(mess=mess):
                    self.in_args['no_gui'] = True
                else:
                    sys.exit(24)
            logging.debug('Path exist: {}'.format(self.gui_icon_path))

        def del_pid(self):
            if isfile(self.fin_file):
                logging.debug('Delete PID file')
                remove(self.fin_file)

        def check_inst(self):
            if self.firstinst:
                logging.info('You selected update mode,\n'
                             'but no trace of complete installation was detected.\n'
                             'You should first go through the full installation process.')
                exit(34)

        def __select_sources(self):

            def dappvpn():
                logging.info(
                    'Enter the dappvpn build number for downloading')
                vpnBin = raw_input('>')
                mess = "You enter '{}'\n" \
                       "Please enter N and try again or " \
                       "Y if everything is correct".format(vpnBin)

                if not self.prompt(mess):
                    vpnBin = dappvpn()
                return vpnBin

            def dappctrl():
                logging.info(
                    'Enter the dappctrl build number for downloading')
                ctrlBin = raw_input('>')
                mess = "You enter '{}'\n" \
                       "Please enter N and try again or " \
                       "Y if everything is correct".format(ctrlBin)

                if not self.prompt(mess):
                    ctrlBin = dappctrl()
                return ctrlBin + '/'

            def back():
                logging.info('Enter the Back build number for downloading')
                back_build = raw_input('>')
                mess = "You enter '{}'\n" \
                       "Please enter N and try again or " \
                       "Y if everything is correct".format(back_build)

                if not self.prompt(mess):
                    back_build = back()
                return back_build

            def gui():
                logging.info('Enter the GUI build number for downloading')
                gui_build = raw_input('>')
                mess = "You enter '{}'\n" \
                       "Please enter N and try again or " \
                       "Y if everything is correct".format(gui_build)

                if not self.prompt(mess):
                    gui_build = gui()
                return gui_build

            def back_gui():
                self.back_route = back()
                self.gui_route = gui()

            if self.in_args['update_bin']:
                self.dappvpn_route = dappvpn() + '/'
                self.dappctrl_route = dappctrl() + '/'
            elif self.in_args['update_back'] or self.in_args['no_gui'] or self.in_args['cli']:
                self.back_route = back()
            elif self.in_args['update_gui']:
                self.gui_route = gui()
            else:
                back_gui()
                # elif self.in_args['update_mass']:
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

        def __select_binary_sources(self):

            def dappvpn():
                logging.info(
                    'Enter the dappvpn build number for downloading')
                vpnBin = raw_input('>')
                mess = "You enter '{}'\n" \
                       "Please enter N and try again or " \
                       "Y if everything is correct".format(vpnBin)

                if not self.prompt(mess):
                    vpnBin = dappvpn()
                return vpnBin

            def dappctrl():
                logging.info(
                    'Enter the dappctrl build number for downloading')
                ctrlBin = raw_input('>')
                mess = "You enter '{}'\n" \
                       "Please enter N and try again or " \
                       "Y if everything is correct".format(ctrlBin)

                if not self.prompt(mess):
                    ctrlBin = dappctrl()
                return ctrlBin

            self.dappvpnBin = dappvpn()
            self.dappctrlBin = dappctrl()

        def validate_url(self, url, binary=False):
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
                    # if binary:
                    #     self.__select_binary_sources()
                    # else:
                    self.__select_sources()
                    break
                else:
                    logging.info(
                        '\nThe address: {} was entered incorrectly.\n'
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

        def check_role(self):
            mess = 'Please select your role.\n Enter digits 1 or 2.\n' \
                   '1 - You role are agent\n' \
                   '2 - You role are client\n'

            if self.prompt(mess=mess, choise=('1', '2')):
                # choise 2
                self.dappctrl_role = 'client'
            else:
                # choise 1
                self.dappctrl_role = 'agent'

        def input_args(self):
            ''' Get input args and parse it '''
            parser = ArgumentParser(description=' *** Initializer v.{} *** '.format(self.i_version))

            parser.add_argument("--build", action='store_true',
                                default=False,
                                help='Create .dapp_cmd file.')

            parser.add_argument("--cli", action='store_true',
                                default=False,
                                help='Install app without GUI and automate publishing offer.')

            parser.add_argument("--republish", action='store_true',
                                default=False,
                                help='Republish offer. Use only with --cli.')

            parser.add_argument("--file", type=str, default=False, nargs='?',
                                help='Add full path to file with offering configs on JSON format. Use only with --cli. Example: /opt/privatix/offer.json. cat offer.json --> {"country": "US"}')

            parser.add_argument("--update-back", action='store_true',
                                default=False,
                                help='Update containers and rerun initializer.')

            parser.add_argument("--update-bin", action='store_true',
                                default=False,
                                help='Download and update binary files in containers.')

            parser.add_argument("--update-gui", action='store_true',
                                default=False,
                                help='Update GUI.')

            parser.add_argument("--update-mass", action='store_true',
                                default=False,
                                help='Update All.')

            parser.add_argument("--no-gui", action='store_true',
                                default=False,
                                help='Full install without GUI.')

            parser.add_argument("--no-wait", action='store_true',
                                default=False,
                                help='Installation without checking ports and waiting for their open.')

            parser.add_argument("--clean", action='store_true',
                                default=False,
                                help='Cleaning after the initialization process. Removing GUI, downloaded files, initialization pid file, stopping containers.')

            parser.add_argument("--link", type=str, default=False, nargs='?',
                                help='Enter link for download. default "http://art.privatix.net/"')

            parser.add_argument("--D", action='store_true', default=False,
                                help='Switch on debug mode')

            parser.add_argument("--branch", type=str, default=False,
                                nargs='?',
                                help='Enter different branch for download. default "develop"')

            if not self.old_vers:
                parser.add_argument('--vpn', type=str, default=False,
                                    help='[start,stop,restart,status]')

                parser.add_argument('--comm', type=str, default=False,
                                    help='[start,stop,restart,status]')

                parser.add_argument('--mass', type=str, default=False,
                                    help='[start,stop,restart,status]')

                parser.add_argument("--test", action='store_true',
                                    default=False,
                                    help='Test mode')

            return vars(parser.parse_args())

        def main_cycle(self):
            if self.in_args['D']:
                ch.setLevel('DEBUG')
                ch.setFormatter(form_console)
                logging.getLogger().addHandler(ch)
                logging.debug('Debug mode enabled')

            if self.in_args['link']:
                logging.info(
                    'You chose was to change link from: {}   to: {}'.format(
                        main_conf['link_download'], self.in_args['link']))

                self.validate_url(self.in_args['link'])

            if self.in_args['branch']:
                logging.debug('Change branch from: {}, to: {}'.format(
                    main_conf['branch'], self.in_args['branch']))

                main_conf['branch'] = self.in_args['branch']
                self.init_branch()

            if self.in_args['build']:
                logging.info('Build mode.')
                self.build_file()

            elif self.in_args['cli']:
                logging.info('Auto offering.')
                self.in_args['no_gui'] = True
                if self.in_args['file']:
                    logging.debug('Check existence offer file: {}'.format(
                        self.in_args['file']))
                    res = self.validateJson(self.in_args['file'])
                    if not res[0]:
                        logging.error(res[1])
                        exit(33)

                if self.in_args['republish']:
                    logging.info('Republish offering.')
                    resp = self.pswd_from_conf()
                    logging.debug('Response: {}'.format(resp))
                    if resp[0]:
                        logging.debug('Republish mode available')
                        res = self.republishOffer()
                        if res[0]:
                            logging.info(
                                'Republish done successfully.\n'
                                'ethAddr: {}\n'
                                'offer id: {}\n'
                                'agent id: {}\n'
                                'eth Balance: {}\n'
                                'psc Balance: {}\n'
                                'ptc Balance: {}\n'.format(
                                    self.ethAddr,
                                    self.offer_id,
                                    self.agent_id,
                                    self.ethBalance,
                                    self.pscBalance,
                                    self.ptcBalance,
                                ))

                    else:
                        logging.error(
                            'Repablish trouble: {}'.format(resp[1]))

                else:
                    self.target = 'back'
                    self.pswd = ''.join(SystemRandom().choice(
                        ascii_uppercase + ascii_lowercase + digits
                    ) for _ in range(self.pswdSymbol))
                    self.check_sudo()
                    self.dappctrl_role = 'agent'
                    self.init_os()

            elif self.in_args['clean']:
                logging.info('Clean mode.')
                self.clear_contr()
                self.del_pid()
                self._clear_dir(self.gui_path)
                try:
                    if not self.old_vers:
                        for unit in (self.unit_f_com, self.unit_f_vpn):
                            p = self.unit_symlink + unit
                            remove(p)
                            logging.debug('Remove {} done'.format(p))
                except BaseException as symlinkExpt:
                    logging.debug('Symlink expt: {}'.format(symlinkExpt))

            elif not self.old_vers and self.in_args['vpn']:
                logging.debug('Vpn mode.')
                sys.stdout.write(
                    str(self.service('vpn', self.in_args['vpn'],
                                     self.use_ports['vpn'])))

            elif not self.old_vers and self.in_args['comm']:
                logging.debug('Comm mode.')
                sys.stdout.write(
                    str(self.service('comm', self.in_args['comm'],
                                     self.use_ports['common'])))

            elif not self.old_vers and self.in_args['mass']:
                logging.debug('Mass mode.')
                comm_stat = self.service('comm', self.in_args['mass'],
                                         self.use_ports['common'])
                vpn_stat = self.service('vpn', self.in_args['mass'],
                                        self.use_ports['vpn'])
                sys.stdout.write(str(bool(all((comm_stat, vpn_stat)))))

            elif self.in_args['update_back']:
                logging.info('Update containers mode.')
                self.check_inst()
                self.check_sudo()
                self.check_role()
                self.target = 'back'
                if self.clear_contr():
                    self.use_ports = dict(vpn=[],
                                          common=[],
                                          mangmt=dict(
                                              vpn=None,
                                              common=None)
                                          )
                    self.in_args['no_gui'] = True
                    self.init_os(True)
                else:
                    logging.info('Problem with clear all old file.')

            elif self.in_args['update_gui']:
                self.check_inst()
                logging.info('Update GUI mode.')
                self.target = 'gui'
                self.check_sudo()
                self.update_gui()

            elif self.in_args['update_mass']:
                self.check_inst()
                self.target = 'both'
                logging.info('Update All mode.')
                self.check_sudo()
                self.check_role()
                if self.clear_contr():
                    self.use_ports = dict(vpn=[], common=[],
                                          mangmt=dict(vpn=None,
                                                      common=None))
                    self.init_os(True)
                else:
                    logging.info('Problem with clear all old file.')

            elif self.in_args['update_bin']:
                logging.info('Update binary mode.')
                self.check_sudo()
                self.update_binary()

            else:
                logging.info('Begin init.')
                self.target = 'back'
                self.check_sudo()
                self.check_role()
                self.init_os()
                logging.info('All done.')

    return Checker(old, v, dist)


if __name__ == '__main__':

    dist_name, ver, name_ver = linux_distribution()
    old_vers = True if dist_name.lower() == 'ubuntu' and int(
        ver.split('.')[0]) < 16 else False
    if old_vers:
        check = checker_fabric(LXC, old_vers, ver, dist_name)
    else:
        check = checker_fabric(Nspawn, old_vers, ver, dist_name)

    check.in_args = check.input_args()
    check.link_from_def = False if check.in_args.get('link') else True

    if isfile(check.fin_file):
        logging.debug('Pid file exist.')
        raw = check.file_rw(p=check.fin_file,
                            json_r=True,
                            log='Search port in finalizer.pid')
        if raw:
            check.use_ports.update(raw)
    else:
        logging.debug('Pid file not exist.This is first run initializer')
        check.firstinst = True

    logging.debug('Input args: {}'.format(check.in_args))
    logging.debug('Inside finalizer.pid: {}'.format(check.use_ports))

    signal(SIGINT, check.signal_handler)
    try:
        check.main_cycle()
    except BaseException as mainexpt:
        logging.error('Main trouble: {}'.format(mainexpt))
        check._rolback(17)
