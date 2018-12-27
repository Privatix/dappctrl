#!/usr/bin/python

"""
    Preparation before Initializer on pure Python 2.7
    mode:
    python initializer.py  -h                              get help information
    python preparation.py                                  start preparation
    python preparation.py --link                           use another link for download.If not use, def link in variable dwnld_url
"""

import logging
from json import load, dump, loads, dumps

from urllib import urlretrieve
from urllib2 import urlopen, Request
from os.path import isfile
from sys import exit as sys_exit
from argparse import ArgumentParser
from os import remove, environ, system
from platform import linux_distribution
from subprocess import Popen, PIPE, STDOUT
from re import compile, match, IGNORECASE


class Prepare:
    def __init__(self):
        self.log_path = '/var/log/preparation.log'
        self.pack_name = 'dapp-privatix'
        self.link = 'https://github.com/'
        self.dwnld_url = ''
        self.dst_path = '{}.deb'.format(self.pack_name)
        self.inst_pack_cmd = 'sudo apt update && sudo dpkg -i {} || sudo apt-get install -f -y'.format(
            self.dst_path)
        self.search_pack_cmd = 'sudo dpkg -l {} >/dev/null 2>&1'.format(
            self.pack_name)
        self.del_pack_cmd = 'sudo apt purge {} -y'.format(self.pack_name)
        self.create_sudoers = 'su - root -c\'touch {0} && echo "{1}" >> {0}\''
        self.user_file = '/etc/sudoers.d/{}'

        self.check_sudo = 'dpkg -l | grep sudo'
        self.install_sudo = 'su -c \'apt install sudo\''
        self.logger()
        self.get_latest_tag()

    def logger(self):
        logging.getLogger().setLevel('DEBUG')
        form_console = logging.Formatter(
            '%(message)s',
            datefmt='%m/%d %H:%M:%S')

        form_file = logging.Formatter(
            '%(levelname)7s [%(lineno)3s] %(message)s',
            datefmt='%m/%d %H:%M:%S')

        fh = logging.FileHandler(self.log_path)  # file debug
        fh.setLevel('DEBUG')
        fh.setFormatter(form_file)
        logging.getLogger().addHandler(fh)

        ch = logging.StreamHandler()  # console debug
        ch.setLevel('INFO')
        ch.setFormatter(form_console)
        logging.getLogger().addHandler(ch)
        logging.info(' - Begin - ')

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

    def get_latest_tag(self):
        '''https://github.com/Privatix/privatix/releases/download/test_draft/privatix_ubuntu_x64_0.18.0_cli.deb'''
        logging.info('Get latest tag in repo.')

        owner = 'Privatix'
        repo = 'privatix'
        url_api = 'https://api.github.com/repos/{}/{}/releases/latest'.format(
            owner, repo)

        resp = self._get_url(link=url_api)

        if resp and resp.get('tag_name'):
            tag_name = resp['tag_name']
            logging.info('Latest tag name: {}'.format(tag_name))
            # self.dwnld_url = '{}Privatix/privatix/releases/download/' \
            #                  '{}/privatix_ubuntu_x64_{}_cli.deb'.format(
            #     self.link,
            #     'test_draft',
            #     '0.18.0')

            # todo*
            self.dwnld_url = '{}Privatix/privatix/releases/download/' \
                             '{}/privatix_ubuntu_x64_{}_cli.deb'.format(
                self.link,
                tag_name,
                tag_name)

            logging.debug('Download url: {}'.format(self.dwnld_url))

        else:
            raise BaseException('GitHub not responding')

    def sys_call(self, cmd):
        resp = Popen(cmd, shell=True, stdout=PIPE,
                     stderr=STDOUT).communicate()

        logging.debug('Sys call cmd: {}. Stdout: {}'.format(cmd, resp))
        if resp[1]:
            logging.error(
                'Trouble when call: {}. Result: {}'.format(cmd, resp[1]))
            return False
        return resp[0]

    def ubn(self):
        user = environ.get('SUDO_USER')
        logging.debug('SUDO_USER: {}'.format(user))
        f_path = self.user_file.format(user)

        if user:
            if not isfile(f_path):

                line = '{} ALL=(ALL:ALL) NOPASSWD:ALL'.format(user)
                logging.debug(
                    'Add line: {} to file: {}'.format(line, f_path))

                f = open(f_path, "wb")
                try:
                    f.writelines(line)
                    f.close()
                except BaseException as fexpt:
                    logging.error('Trouble: {}'.format(fexpt))
                    if isfile(f_path):
                        remove(f_path)
                    sys_exit(1)

                finally:
                    f.close()
            else:
                logging.info('Sudoers file {} exist'.format(user))
        else:
            logging.error('Trouble. Run the script from sudo!')
            sys_exit(3)

    def deb(self):

        sudo = self.sys_call(self.check_sudo)
        if not sudo:
            logging.info('Install sudo.\n')
            system(self.install_sudo)

        user = self.sys_call('whoami').replace('\n', '')
        f_path = self.user_file.format(user)

        if not isfile(f_path):
            line = '{} ALL=(ALL:ALL) NOPASSWD:ALL'.format(user)
            logging.debug('Add line: {} to file: {}'.format(line, f_path))

            raw = self.create_sudoers.format(f_path, line)
            logging.debug('CMD: {}'.format(raw))
            logging.info('Create sudoers file.')
            if system(raw):
                sys_exit(5)

    def check_dist(self):
        dist_name, ver, name_ver = linux_distribution()
        task = dict(ubuntu=self.ubn,
                    debian=self.deb
                    )
        dist_task = task.get(dist_name.lower(), False)
        if dist_task:
            return dist_task
        else:
            logging.info('You OS is not support yet.')
            sys_exit(4)

    def dwnld_pack(self):

        logging.info('Download {}.\n'
                     'Please wait, '
                     'this may take a few minutes!'.format(self.pack_name))
        logging.info('Download from: {}'.format(self.dwnld_url))
        urlretrieve(self.dwnld_url, self.dst_path)
        logging.info('Download done. Install pack.')

        if system(self.inst_pack_cmd):
            sys_exit(2)

        logging.info('The installation was successful done.\n'
                     'Run: sudo /opt/privatix/initializer/initializer.py')
        sys_exit(0)

    def ask(self):
        logging.debug('Ask confirmation')
        """ 
        Disable reinstall confirmation request. 26.12.18.
        To activate the confirmation, you need to remove `return True` 
        and uncomment the cycle below.
        """
        return True
        # answ = raw_input('>')
        # while True:
        #     if answ.lower() not in ['n', 'y']:
        #         logging.info('Invalid choice. Select y or n.')
        #         answ = raw_input('> ')
        #         continue
        #     if answ.lower() == 'y':
        #         return True
        #     return False

    def prep_checks(self):
        if system(self.search_pack_cmd):
            logging.debug('First run')
            self.check_dist()()
        else:
            logging.info(
                'The package {} is already installed on your computer.\n'
                'Do you want to reinstall it?'.format(self.pack_name))
            if self.ask():
                logging.debug('Reinstall pack')
                if system(self.del_pack_cmd):
                    logging.error('An error occurred during the deletion.\n'
                                  'The process is interrupted.\n'
                                  'Try to remove the package manually and repeat the installation.')
                    sys_exit(7)
                else:
                    logging.info(
                        'The package {} deleted'.format(self.pack_name))

            else:
                logging.debug('Quit')
                sys_exit(6)

    def input_args(self):
        parser = ArgumentParser(description=' *** Preparation *** ')
        parser.add_argument("--link", type=str, default=False, nargs='?',
                            help='Enter link for download. default "http://art.privatix.net/"')
        return vars(parser.parse_args())

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
                self.link = url
                self.dwnld_url = '{}{}.deb'.format(self.link, self.pack_name)
                break
            else:
                logging.info(
                    '\nThe address: {} was entered incorrectly.\n'
                    'Please enter it according to the example:\n'
                    'http://www.example.com/'.format(url))
                url = raw_input('>')

    def run(self):
        in_args = self.input_args()
        if in_args['link']:
            logging.info(
                'You chose was to change link from: {}   to: {}'.format(
                    self.link, in_args['link']))

            self.validate_url(in_args['link'])
        self.prep_checks()
        self.dwnld_pack()


if __name__ == "__main__":
    pr = Prepare()
    pr.run()
