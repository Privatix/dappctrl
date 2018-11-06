package statik

//go:generate chmod +x ./pkgList/gen.sh
//go:generate ./pkgList/gen.sh
//go:generate go run ../tool/copy_dbscripts/copy.go
//go:generate rm -f statik.go
//go:generate statik -f -src=. -dest=..
