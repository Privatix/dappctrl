import json
import sys
from datetime import datetime, timedelta

from dappctrl_rpc import *

minutes = int(sys.argv[1]) if (len(sys.argv)) >= 2 else 30

print("Show errors for the last {} minutes".format(minutes))

current_date = datetime.now()
past_date = current_date - timedelta(minutes=minutes)

token = get_token(default_password)

errors = get_logs(token, ["error"], "", str(past_date.isoformat()), str(current_date.isoformat()), 0, 100)
if errors is None:
    exit(0)

for error in errors:
    print("-" * 80)
    print("\n{}:\n\tMessage: {}\n\n\tContext: {}\n\n\tTime: {}".format(
        error["level"],
        error["message"],
        json.dumps(error["context"]),
        error["time"]))
