
INSERT INTO settings (key, value, permissions, description, name)
VALUES ('system.gui',
        '{}',
        0,
        '',
        'GUI settings')
ON CONFLICT (key)
DO NOTHING;

-- TODO(furkhat) Remove updateDismissVersion when front-end migrates completely to use `system.gui`
INSERT INTO settings (key, value, permissions, description, name)
VALUES ('updateDismissVersion',
        '',
        2,
        '',
        'Update dismiss version')
ON CONFLICT (key)
DO NOTHING;

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('eth.min.confirmations',
        '1',
        2,
        'Have value (stored as string) that is null or integer and' ||
        ' represents how many ethereum blocks should be mined after' ||
        ' block where transaction of interest exists. As there is non' ||
        ' zero probability of attack where some last blocks can be' ||
        ' generated by attacker and will be than ignored by ethereum' ||
        ' network (uncle blocks) after attack detection. dappctrl' ||
        ' give ability to user to specify how many latest blocks' ||
        ' are considered non reliable. These last blocks' ||
        ' will not be used to fetch events or transactions.',
        'Ethereum confirmation blocks')
ON CONFLICT (key)
DO NOTHING;

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('eth.event.offeringsfreshblocks',
        '345600',
        2,
        'For clients, defines number of latest ethereum blocks to retrieve offerings from.' ||
        ' If eth.event.offeringsfreshblocks is null or zero then events starting' ||
        ' from the latest block will be downloaded.',
        'Offerings last events blocks')
ON CONFLICT (key)
DO NOTHING;

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('eth.event.blocklimit',
        '500',
        2,
        'Maximum number of ethereum blocks that is used to scan' ||
        ' for new events. It is used as pagination mechanism while' ||
        ' querying ethereum JSON RPC. If eth.event.blocklimit is null' ||
        ' or zero then no pagination is used, which is not recommended.',
        'Maximum events blocks')
ON CONFLICT (key)
DO NOTHING;

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('error.sendremote',
        'false',
        2,
        'Allow error reporting to send logs to Privatix.',
        'Error reporting')
ON CONFLICT (key)
DO NOTHING;

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('eth.max.deposit',
        '30000000000',
        2,
        'We temporarily limit total token deposits in a channel to 300' ||
        ' PRIX. This is just for the bug bounty release, as a safety measure.',
        'Maximum deposit')
ON CONFLICT (key)
DO NOTHING;

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('eth.event.clientMonitoringStartBlock',
        '0',
        0,
        'Block from which (Client) monitoring started started.',
        'Client monitoring start block')
ON CONFLICT (key)
DO NOTHING;

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('eth.event.lastProcessedBlock',
        '0',
        1,
        'Last proccessed blockchain block number.',
        'Last processed block')
ON CONFLICT (key)
DO NOTHING;

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('eth.event.lastBackSearchBlock',
        '0',
        1,
        'On client, the last block offerings searched from.',
        'Last back search block')
ON CONFLICT (key)
DO NOTHING;

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('offering.autopopup',
        'false',
        2,
        'Allow offerings to pop up automatically.',
        'Offering autopopup')
ON CONFLICT (key)
DO NOTHING;

INSERT INTO settings (key, value, permissions, description, name)
VALUES('psc.periods.challenge',
        '5000',
        1,
        'Number of blocks to be mined to finish uncooperative channel close',
        'Challenge period')
ON CONFLICT (key)
DO UPDATE SET value='5000';

INSERT INTO settings (key, value, permissions, description, name)
VALUES('psc.periods.popup',
        '500',
        1,
        'Number of blocks to be mined to repeatedly pop up an offering',
        'Popup period')
ON CONFLICT (key)
DO UPDATE SET value='500';

INSERT INTO settings (key, value, permissions, description, name)
VALUES('psc.periods.remove',
        '100',
        1,
        'Number of blocks to be mined from last offering update in blockchain to remove offering',
        'Remove period')
ON CONFLICT (key)
DO UPDATE SET value='100';

INSERT INTO settings (key, value, permissions, description, name)
VALUES('client.min.deposit',
        '0',
        2,
        'This value will override min. deposit proposed by Agent' ||
        ' in auto-increase mode, if greater than proposed. ',
        'Auto-increase min deposit')
ON CONFLICT (key)
DO NOTHING;

INSERT INTO settings (key, value, permissions, description, name)
VALUES('client.autoincrease.percent',
        '60',
        2,
        'This setting specifies when to increase deposit. Increase deposit,' ||
        ' when current usage is bigger or equal to this percent of total used units.',
        'Top up after using %')
ON CONFLICT (key)
DO NOTHING;

INSERT INTO settings (key, value, permissions, description, name)
VALUES('client.autoincrease.deposit',
        'true',
        2,
        'Enable or disable auto-increase mode. Only for client.',
        'Client auto-increase deposit')
ON CONFLICT (key)
DO NOTHING;

INSERT INTO settings (key, value, permissions, description, name)
VALUES('rating.ranking.steps',
        '30',
        2,
        'Number of iterations to compute rank in ratings calculation',
        'Ranking steps number')
ON CONFLICT (key)
DO NOTHING;
