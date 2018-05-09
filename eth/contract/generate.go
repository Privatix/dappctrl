package contract

//go:generate abigen --abi abis/psc.abi --pkg contract --type PrivatixServiceContract --out psc.go
//go:generate abigen --abi abis/ptc.abi --pkg contract --type PrivatixTokenContract --out ptc.go
//go:generate abigen --abi abis/sale.abi --pkg contract --type Sale --out sale.go
