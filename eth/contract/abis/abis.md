# Instructions for ABI files processing.

* ABI files are used as basis for contracts interfaces generation.
* This files are generated on each run of Ethereum test environment, 
and can be located in `<dapp-smart-contract-dir>/psc/{psc, token, sale}.abi`
(please, see `https://github.com/Privatix/dapp-smart-contract` for the details)

* Each ABI file should be copied here and processed via abigen. 
(Please, see instructions located in `eth/contract/tools/`).
 
