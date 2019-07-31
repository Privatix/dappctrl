import json

from dappctrl_rpc import *

token = get_token(default_password)

all_channel_statuses = ['pending', 'active', 'wait_coop', 'closed_coop', 'wait_challenge', 'in_challenge',
                        'wait_uncoop',
                        'closed_uncoop']

for service_status in ['pending', 'activating', 'active', 'suspending',
                       'suspended', 'terminating', 'terminated']:
    print("-" * 80)
    print("\n{}:\n".format(service_status))
    channels = get_agent_channels(token, all_channel_statuses, [service_status], 0, 100)
    print("\n\tChannels: {}".format(json.dumps(channels, indent=8)))

print("-" * 80)
income = get_total_income(token)
print("\nTotal income: {} PRIX".format(prix(income)))
