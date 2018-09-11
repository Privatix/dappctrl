import os
import re
import subprocess

_dappctrl_folder_path = os.path.expanduser("~") + '/go/src/github.com/privatix/dappctrl'

_git_branch_name_command = ['git', 'rev-parse', '--abbrev-ref', 'HEAD']
_git_add_all_command = ['git', 'add', '-A']
_git_commit_command = ['git', 'commit', '-m']

_release_prefix = 'release/'

_prod_data_sql_path = 'data/prod_data.sql'
_prod_data_sql_pattern = r"(system.version.db'[,\n\t\s]+)'(\d+\.\d+\.\d+)"

_dappinst_path = 'tool/dappinst/main.go'
_dappinst_pattern = r'(appVersion\s+=\s+)"(\d+\.\d+\.\d+)'


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


def commit_all(release_version):
    commit_message = 'change version to {}'.format(release_version)

    subprocess.call(_git_add_all_command)
    subprocess.call(_git_commit_command+[commit_message])


os.chdir(_dappctrl_folder_path)
_release_version = take_release_version()
actualize_prod_data(_dappctrl_folder_path, _release_version)
actualize_dappinst(_dappctrl_folder_path, _release_version)
commit_all(_release_version)
