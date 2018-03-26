package somc

import (
	"encoding/base64"
	"encoding/json"

	"github.com/ethereum/go-ethereum/crypto"
)

const publishOfferingMethod = "newOffering"

type publishOfferingParams struct {
	Hash string `json:"hash"`
	Data string `json:"data"`
}

// PublishOffering publishes a given offering JSON in SOMC.
func (c *Conn) PublishOffering(o []byte) error {
	hash := crypto.Keccak256Hash(o)
	params := publishOfferingParams{
		Hash: base64.URLEncoding.EncodeToString(hash.Bytes()),
		Data: base64.URLEncoding.EncodeToString(o),
	}

	data, err := json.Marshal(&params)
	if err != nil {
		return err
	}

	return c.request(publishOfferingMethod, data).err
}
