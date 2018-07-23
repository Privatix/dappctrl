#!/usr/bin/python

import logging
from sys import exit as sys_exit
from os import remove
from os import environ
from os.path import isfile
from urllib import URLopener
from subprocess import Popen, PIPE, STDOUT

user = environ.get('SUDO_USER')
log_path = 'pre.log'
pack_name = 'dapp-privatix.deb'
dwnld_url = 'http://art.privatix.net/{}'.format(pack_name)
dst_path = '{}'.format(pack_name)
inst_pack = 'sudo dpkg -i {} || sudo apt-get install -f -y'.format(dst_path)


def logger():
    logging.getLogger().setLevel('DEBUG')  # console debug
    form_console = logging.Formatter(
        '%(message)s',
        datefmt='%m/%d %H:%M:%S')

    form_file = logging.Formatter(
        '%(levelname)7s [%(lineno)3s] %(message)s',
        datefmt='%m/%d %H:%M:%S')

    fh = logging.FileHandler(log_path)  # file debug
    fh.setLevel('DEBUG')
    fh.setFormatter(form_file)
    logging.getLogger().addHandler(fh)

    ch = logging.StreamHandler()  # console debug
    ch.setLevel('INFO')
    ch.setFormatter(form_console)
    logging.getLogger().addHandler(ch)

    logging.debug('SUDO_USER: {}'.format(user))


def main():
    if user:
        line = '{} ALL=(ALL:ALL) NOPASSWD:ALL'.format(user)
        user_file = '/etc/sudoers.d/{}'.format(user)
        logging.debug('Add line: {} to file: {}'.format(line, user_file))

        f = open(user_file, "wb")
        try:
            f.writelines(line)
            f.close()
        except BaseException as fexpt:
            logging.error('Trouble: {}'.format(fexpt))
            if isfile(user_file):
                remove(user_file)
            sys_exit(1)

        finally:
            f.close()

        obj = URLopener()
        logging.info('Download {}.\n'
                     'Please wait, '
                     'this may take a few minutes!'.format(pack_name))
        obj.retrieve(dwnld_url, dst_path)
        logging.info('Download done. Install pack')
        resp = Popen(inst_pack, shell=True, stdout=PIPE,
                     stderr=STDOUT).communicate()

        logging.debug('Resp: {}'.format(resp))
        logging.info('The installation was successful done.\n'
                     'Run: sudo /opt/privatix/initializer/initializer.py')
        sys_exit(0)

    else:

        logging.error('Trouble. Run the script from sudo!')
        sys_exit(2)


if __name__ == "__main__":
    logger()
    main()

