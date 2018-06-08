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

// StaticPWDStorage returns static static password, can't be rewritten.
type StaticPWDStorage string

// Get returns stored static password.
func (s *StaticPWDStorage) Get() string {
	return string(*s)
}

// Set does nothing.
func (s *StaticPWDStorage) Set(_ string) {}

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
