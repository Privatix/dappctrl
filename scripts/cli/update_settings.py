import json
import sys

from dappctrl_rpc import *

settings = json.loads((sys.argv[1]))

token = get_token(default_password)

update_settings(token, settings)
