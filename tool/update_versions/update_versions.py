import os
import re
import subprocess

import sys

_git_branch_name_command = 'git rev-parse --abbrev-ref HEAD'
_release_prefix = 'release/'

_prod_data_sql_path = 'data/prod_data.sql'
_prod_data_sql_pattern = r"(system.version.db'[,\n\t\s]+)'(\d+\.\d+\.\d+)"

_dappinst_path = 'tool/dappinst/main.go'
_dappinst_pattern = r'(appVersion\s+=\s+)"(\d+\.\d+\.\d+)'


def check_arguments():
    if len(sys.argv) < 2:
        print('usage: update_versions.py <path_dappctrl_folder>\n\n'
              'Example:\n'
              'python update_versions.py /go/src/github.com/privatix/dappctrl\n\n')
        exit(1)


def take_repo_folder():
    dappctrl_folder_path = sys.argv[1]

    if not os.path.exists(dappctrl_folder_path):
        print('\n\nFailed. Folder does not exists: {}'.format(dappctrl_folder_path))
        exit(1)
    print('Repository folder: {}'.format(dappctrl_folder_path))
    return dappctrl_folder_path


def take_release_version():
    current_branch_name = subprocess.check_output(_git_branch_name_command).decode("utf-8").strip()
    print('Current branch: {}'.format(current_branch_name))
    if not current_branch_name.startswith(_release_prefix):
        print('\n\nFailed. Please checkout to release branch')

    release_version = current_branch_name[len(_release_prefix)::]
    print('Release version: {}'.format(release_version))

    return release_version


def replace_in_file(file, pattern, replacement):
    with open(file, 'r') as f:
        content = f.read()

    content = re.sub(pattern, replacement, content)
    with open(file, 'w') as f:
        f.write(content)


def actualize_prod_data(dappctrl_folder_path, release_version):
    file_path = os.path.join(dappctrl_folder_path, _prod_data_sql_path)

    replace_in_file(file_path, _prod_data_sql_pattern, r"\1'{}".format(release_version))


def actualize_dappinst(dappctrl_folder_path, release_version):
    file_path = os.path.join(dappctrl_folder_path, _dappinst_path)

    replace_in_file(file_path, _dappinst_pattern, r'\1"{}'.format(release_version))


check_arguments()
_dappctrl_folder_path = take_repo_folder()
os.chdir(_dappctrl_folder_path)
_release_version = take_release_version()
actualize_prod_data(_dappctrl_folder_path, _release_version)
actualize_dappinst(_dappctrl_folder_path, _release_version)
