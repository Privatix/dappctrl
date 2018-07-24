#!/usr/bin/python

import logging
from sys import exit as sys_exit
from os import remove,path
from os import environ,system
from os.path import isfile
from urllib import URLopener
from platform import linux_distribution
from subprocess import Popen, PIPE, STDOUT

log_path = '/var/log/preparation.log'
pack_name = 'dapp-privatix.deb'
dwnld_url = 'http://art.privatix.net/{}'.format(pack_name)
dst_path = '{}'.format(pack_name)
inst_pack = 'sudo dpkg -i {} || sudo apt-get install -f -y'.format(dst_path)
create_sudoers = 'su - root -c\'touch {0} && echo "{1}" >> {0}\''
user_file = '/etc/sudoers.d/{}'

check_sudo = 'dpkg -l | grep sudo'
install_sudo = 'su -c \'apt install sudo\''

def logger():
    logging.getLogger().setLevel('DEBUG')
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
    logging.info(' - Begin - ')


def sys_call(cmd):
    resp = Popen(cmd, shell=True, stdout=PIPE, stderr=STDOUT).communicate()

    logging.debug('Sys call cmd: {}. Stdout: {}'.format(cmd, resp))
    if resp[1]:
        logging.error('Trouble when call: {}. Result: {}'.format(cmd, resp[1]))
        return False
    return resp[0]


def ubn():
    user = environ.get('SUDO_USER')
    logging.debug('SUDO_USER: {}'.format(user))
    f_path = user_file.format(user)

    if user:
        if not isfile(f_path):

            line = '{} ALL=(ALL:ALL) NOPASSWD:ALL'.format(user)
            logging.debug('Add line: {} to file: {}'.format(line, f_path))

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


def deb():

    sudo = sys_call(check_sudo)
    if not sudo:
        logging.info('Install sudo.\n')
        system(install_sudo)

    user = sys_call('whoami').replace('\n', '')
    f_path = user_file.format(user)

    if not isfile(f_path):
        line = '{} ALL=(ALL:ALL) NOPASSWD:ALL'.format(user)
        logging.debug('Add line: {} to file: {}'.format(line, f_path))

        raw = create_sudoers.format(f_path, line)
        logging.debug('CMD: {}'.format(raw))
        logging.info('Create sudoers file.')
        if system(raw):
            sys_exit(5)


def check_dist():
    dist_name, ver, name_ver = linux_distribution()
    task = dict(ubuntu=ubn,
                debian=deb
                )
    dist_task = task.get(dist_name.lower(), False)
    if dist_task:
        return dist_task
    else:
        logging.info('You OS is not support yet.')
        sys_exit(4)


def dwnld_pack():

    obj = URLopener()
    logging.info('Download {}.\n'
                 'Please wait, '
                 'this may take a few minutes!'.format(pack_name))
    obj.retrieve(dwnld_url, dst_path)
    logging.info('Download done. Install pack.')

    if system(inst_pack):
        sys_exit(2)

    logging.info('The installation was successful done.\n'
                 'Run: sudo /opt/privatix/initializer/initializer.py')
    sys_exit(0)


def main():

        check_dist()()
        dwnld_pack()


if __name__ == "__main__":
    logger()
    main()

