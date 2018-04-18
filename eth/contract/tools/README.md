# Instructions for processing ABI files

* Each ABI file must converted to the contract implementation in the next way:
    * `./tools/abigen --abi abis/psc.abi --pkg contract --type PrivatixServiceContract --out psc.go`
    * `./tools/abigen --abi abis/ptc.abi --pkg contract --type PrivatixTokenContract --out ptc.go`
    * `./tools/abigen --abi abis/sale.abi --pkg contract --type Sale --out sale.go`

* By default, current implementation provides abigen build for linux platform only. 
For other platforms it must be built separately. 

* Note about go-generate usage: 
It seems that, `go generate` can't be used for ABIs processing 
because of complexity of building abigen for different platforms in automated way: 
the main motivation to not do this - is that abigen uses various dependencies on various platforms: 
on mac it uses `brew`, that can't be installed/used without sudo privileges. 
Also, mac-developers might not wan't to collect rarely used tools as `abigen` in OS package manager.
On the linux platform it uses go-dep, that is also one more external dependency, that is not really required.
