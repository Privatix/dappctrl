package somc

import (
	"encoding/json"

	"github.com/privatix/dappctrl/data"

	"github.com/ethereum/go-ethereum/crypto"
)

const publishOfferingMethod = "newOffering"

type publishOfferingParams struct {
	Hash string `json:"hash"`
	Data string `json:"data"`
}

// PublishOffering publishes a given offering JSON in SOMC.
func (c *Conn) PublishOffering(o []byte) error {
	hash := crypto.Keccak256(o)
	params := publishOfferingParams{
		Hash: data.FromBytes(hash),
		Data: data.FromBytes(o),
	}

	data, err := json.Marshal(&params)
	if err != nil {
		return err
	}

	return c.request(publishOfferingMethod, data).err
}
