package contract

//go:generate ./tools/abigen --abi abis/psc.abi --pkg contract --type PrivatixServiceContract --out psc.go
//go:generate ./tools/abigen --abi abis/ptc.abi --pkg contract --type PrivatixTokenContract --out ptc.go
//go:generate ./tools/abigen --abi abis/sale.abi --pkg contract --type Sale --out sale.go
