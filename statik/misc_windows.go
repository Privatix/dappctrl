package statik

//go:generate .\pkgList\gen.bat
//go:generate go run ..\tool\copy_dbscripts\copy.go
//go:generate cmd /C IF EXIST "statik.go" (del /F /Q statik.go)
//go:generate statik -f -src=. -dest=..
