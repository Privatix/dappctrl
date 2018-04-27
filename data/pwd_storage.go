package data

// PWDStorage stores password in memory.
type PWDStorage string

// Get returns stored password.
func (p *PWDStorage) Get() string {
	return string(*p)
}

// Set sets stored password.
func (p *PWDStorage) Set(pwd string) {
	*p = PWDStorage(pwd)
}

// PWDGetter can retrieve stored password.
type PWDGetter interface {
	Get() string
}

// PWDSetter can set new password.
type PWDSetter interface {
	Set(string)
}

// PWDGetSetter can get and set password to storage.
type PWDGetSetter interface {
	PWDGetter
	PWDSetter
}
