====================================
Description of initializer arguments
====================================
**Initializer write on pure Python 2.7**

*Mode example:*


* python initializer.py  -h
    * Get help information
* python initializer.py
    * Start full install, with GUI
* python initializer.py --build
    * Create file wiwh cmds for generate dappvpn.config.json
* python initializer.py --vpn [start/stop/restart/status]
    * Control vpn servise. Example python initializer.py --vpn start
* python initializer.py --comm [start/stop/restart/status]
    * Control common servise. Example python initializer.py --comm start
* python initializer.py --mass [start/stop/restart/status]
    * Control common + vpn servise. Example python initializer.py --mass start
* python initializer.py --no-gui
    * Install without GUI
* python initializer.py --update-back
    * Update all contaiter without GUI.All containers are stopped first, then removed, and installed in a new way.
* python initializer.py --update-mass
    * Update all contaiter with GUI.Stop all runs containers,then removed, and installed in a new way.
* python initializer.py --update-gui
    * Update only GUI,delete gui files and reinstall
* python initializer.py --link
    * Use another link for download.If not use, default link as will be the same as in the application configuration in main_conf[link_download]
* python initializer.py --branch
    * Use another branch than 'develop' for download.
    * Template https://raw.githubusercontent.com/Privatix/dappctrl/{ branch }/
* python initializer.py --cli
    * Run initializer in auto offer mode.In which will create account,transfer PRIX and ETH to address of account,Transfer some PRIX from PTC balance to PSC balance,Create offering and Publish offering.
* python initializer.py --cli --file [path/to/file.json]
    * Run initializer in auto offer mode,but with unique offer parameters from offer.json
    * Path and file name, arbitrary. File format - JSON.
    * Run example: python initializer.py --cli --file /home/offer.json.
    * Example of all default parameters and offer file structure:

    .. code-block:: json

        {
            "serviceName": "my service",
            "description": "my service description",
            "country": "UA",
            "supply": 3,
            "unitName": "MB",
            "autoPopUp": True,
            "unitType": "units",
            "billingType": "postpaid",
            "setupPrice": 0,
            "unitPrice": 10000,
            "minUnits": 10000,
            "maxUnit": 30000,
            "billingInterval": 1,
            "maxBillingUnitLag": 100,
            "maxSuspendTime": 1800,
            "maxInactiveTimeSec": 1800,
            "additionalParams":
                {
                "minDownloadMbits":100,
                "minUploadMbits":80
                },

        }

    * If you like, you can set all your parameters, or selectively. If you specify only specific parameters, the rest will be the default. You can create a file and specify in it for example:

    .. code-block:: json

        {
            "serviceName": "Test Name",
            "country": "UK",
        }

* python initializer.py --clean
    * Stop all run containers and clear all dirs where store containers, gui and initializer.pid file.

