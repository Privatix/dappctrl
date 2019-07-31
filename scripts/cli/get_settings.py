from dappctrl_rpc import *

token = get_token(default_password)

settings = get_settings(token)
for setting in settings:
    print("\n{}: {}\n\tPermissions: {}".format(
        setting,
        settings[setting]["value"],
        settings[setting]["permissions"],
    ))
