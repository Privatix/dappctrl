package data

import "crypto/ecdsa"

// keyStorage stores private keys in memory.
type keyStorage struct {
	keys        map[HexString]*ecdsa.PrivateKey
	decryptFunc ToPrivateKeyFunc
}

func (s *keyStorage) getKey(acc *Account, password string) (*ecdsa.PrivateKey, error) {
	if v := s.keys[acc.EthAddr]; v != nil {
		return v, nil
	}
	v, err := s.decryptFunc(acc.PrivateKey, password)
	if err != nil {
		return nil, err
	}
	s.keys[acc.EthAddr] = v
	return v, nil
}

// PWDStorage stores password in memory.
type PWDStorage struct {
	ks       keyStorage
	password string
}

// NewPWDStorage returns fresh instance of PWDStorage.
func NewPWDStorage(decryptFunc ToPrivateKeyFunc) *PWDStorage {
	return &PWDStorage{
		ks: keyStorage{
			keys:        make(map[HexString]*ecdsa.PrivateKey),
			decryptFunc: decryptFunc,
		},
	}
}

// Get returns stored password.
func (s *PWDStorage) Get() string {
	return s.password
}

// Set sets stored password.
func (s *PWDStorage) Set(pwd string) {
	s.password = pwd
}

// GetKey returns private key of account.
func (s *PWDStorage) GetKey(acc *Account) (*ecdsa.PrivateKey, error) {
	return s.ks.getKey(acc, s.password)
}

// StaticPWDStorage returns static static password, can't be rewritten.
type StaticPWDStorage struct {
	ks       keyStorage
	password string
}

// NewStatisPWDStorage returns fresh instance of StaticPWDStorage.
func NewStatisPWDStorage(p string, decryptFunc ToPrivateKeyFunc) *StaticPWDStorage {
	return &StaticPWDStorage{
		ks: keyStorage{
			keys:        make(map[HexString]*ecdsa.PrivateKey),
			decryptFunc: decryptFunc,
		},
		password: p,
	}
}

// Get returns stored static password.
func (s *StaticPWDStorage) Get() string {
	return s.password
}

// Set does nothing.
func (s *StaticPWDStorage) Set(_ string) {}

// GetKey returns private key of account.
func (s *StaticPWDStorage) GetKey(acc *Account) (*ecdsa.PrivateKey, error) {
	return s.ks.getKey(acc, s.password)
}

// PWDGetter can retrieve stored password.
type PWDGetter interface {
	Get() string
	GetKey(*Account) (*ecdsa.PrivateKey, error)
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
