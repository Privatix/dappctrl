# Environment setup instructions
This doc contains instructions for environment setup for eth. client library.

### Requirements
* Truffle environment;
* Test API setup and running;

### Truffle env. setup
Provided steps must be performed only once.
No need to repeat it on each tests cycle,
but it is necessary to repeat `npm update` on each repository change,
because there might be new dependencies added.

1. Clone https://github.com/Privatix/dapp-smart-contract
1. Go to the `dapp-smart-contract-private/psc`
and run `npm install` or `npm update` (no `-g` param is needed)
1. ensure node.js version is 9+ (`node -v`)

### GO test env. setup
1. Ensure GOROOT and GOPATH are configured correctly.
1. Check `dappctrl-test.config.json` (in projects root directory),
section `"EthereumLib"`.
If necessary, edit credentials for remote geth node access and tests API.
By default, test API should be available on port `5000`
and geth node should be available on port `8545`.

### Tests running
1. Ensure test API is up and running: `TARGET=dev npm run dev`;
1. Wait some period of time (5-10 sec);
1. `go test pxctrl/eth/lib`

### Troubleshooting
* *Can't fetch private key. It seems that test environment is broken*

Please, check if test API is run and running.
Use commands `npm run stop` and `TARGET=dev npm run dev` to restart test environment.