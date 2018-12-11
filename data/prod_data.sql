--templateid = efc61769-96c8-4c0d-b50a-e4d11fc30523
BEGIN TRANSACTION;

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('eth.min.confirmations',
        '1',
        2,
        'have value (stored as string) that is null or integer and' ||
        ' represents how many ethereum blocks should be mined after' ||
        ' block where transaction of interest exists. As there is non' ||
        ' zero probability of attack where some last blocks can be' ||
        ' generated by attacker and will be than ignored by ethereum' ||
        ' network (uncle blocks) after attack detection. dappctrl' ||
        ' give ability to user to specify how many latest blocks' ||
        ' are considered non reliable. These last blocks' ||
        ' will not be used to fetch events or transactions.',
        'ethereum confirmation blocks');

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('eth.event.freshblocks',
        '1000',
        2,
        'defines number of latest ethereum blocks to retrieve.' ||
        ' If eth.event.freshblocks is null or zero then all events' ||
        ' will be downloaded.',
        'last events blocks');

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('eth.event.blocklimit',
        '500',
        2,
        'maximum number of ethereum blocks that is used to scan' ||
        ' for new events. It is used as pagination mechanism while' ||
        ' querying ethereum JSON RPC. If eth.event.blocklimit is null' ||
        ' or zero then no pagination is used, which is not recommended.',
        'maximum events blocks');

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('error.sendremote',
        'true',
        2,
        'Allow error reporting to send logs to Privatix.',
        'error reporting');

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('eth.default.gasprice',
        '20000000000',
        2,
        'Default GAS price for transactions.',
        'default gas price');

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('eth.max.deposit',
        '30000000000',
        2,
        'We temporarily limit total token deposits in a channel to 300' ||
        ' PRIX. This is just for the bug bounty release, as a safety measure.',
        'maximum deposit');

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('eth.event.lastProcessedBlock',
        '0',
        1,
        'Last block number in blockchain stores last proccessed block.',
        'last processed block');

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('system.version.db',
        '0.16.0',
        1,
        'Version of database.',
        'db version');

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('offering.autopopup',
        'true',
        2,
        'Allow offerings to pop up automatically.',
        'offering autopopup');

INSERT INTO settings (key, value, permissions, description, name)
VALUES ('somc.transport.tor',
        'false',
        2,
        'Whether to use Tor as service offering messaging protocol or not',
        'use Tor');

INSERT INTO settings (key, value, permissions, description, name)
VALUES('psc.periods.challenge',
       '5000',
       1,
       'Number of blocks to be mined to finish uncooperative channel close',
       'Challenge period');

INSERT INTO settings (key, value, permissions, description, name)
VALUES('psc.periods.popup',
       '500',
       1,
       'Number of blocks to be mined to repeatedly pop up an offering',
       'Popup period');

INSERT INTO settings (key, value, permissions, description, name)
VALUES('psc.periods.remove',
       '100',
       1,
       'Number of blocks to be mined from last offering update in blockchain to remove offering',
       'Remove period');

END TRANSACTION;
