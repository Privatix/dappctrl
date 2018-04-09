package somc

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/privatix/dappctrl/data"
)

const findOfferingsMethod = "getOfferings"

type findOfferingsParams struct {
	Hashes []string `json:"hashes"`
}

type findOfferingsResult []struct {
	Hash string `json:"hash"`
	Data string `json:"data"`
}

// OfferingData is a simple container for offering JSON.
type OfferingData struct {
	Hash     string
	Offering []byte
}

// FindOfferings requests SOMC to find offerings by their hashes.
func (c *Conn) FindOfferings(hashes []string) ([]OfferingData, error) {
	params := findOfferingsParams{hashes}

	bytes, err := json.Marshal(&params)
	if err != nil {
		return nil, err
	}

	repl := c.request(findOfferingsMethod, bytes)
	if repl.err != nil {
		return nil, repl.err
	}

	var res findOfferingsResult
	if err := json.Unmarshal(repl.data, &res); err != nil {
		return nil, err
	}

	var ret []OfferingData
	for _, v := range res {
		bytes, err := data.ToBytes(v.Data)
		if err != nil {
			return nil, err
		}

		hash := crypto.Keccak256Hash(bytes)
		hstr := data.FromBytes(hash.Bytes())
		if hstr != v.Hash {
			return nil, fmt.Errorf(
				"SOMC hash mismatch: %s != %s", hstr, v.Hash)
		}

		ret = append(ret, OfferingData{hstr, bytes})
	}

	return ret, nil
}
