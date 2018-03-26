package somc

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
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

	data, err := json.Marshal(&params)
	if err != nil {
		return nil, err
	}

	repl := c.request(findOfferingsMethod, data)
	if repl.err != nil {
		return nil, repl.err
	}

	var res findOfferingsResult
	if err := json.Unmarshal(repl.data, &res); err != nil {
		return nil, err
	}

	var ret []OfferingData
	for _, v := range res {
		data, err := base64.URLEncoding.DecodeString(v.Data)
		if err != nil {
			return nil, err
		}

		hash := crypto.Keccak256Hash(data)
		hstr := base64.URLEncoding.EncodeToString(hash.Bytes())
		if hstr != v.Hash {
			return nil, fmt.Errorf(
				"SOMC hash mismatch: %s != %s", hstr, v.Hash)
		}

		ret = append(ret, OfferingData{hstr, data})
	}

	return ret, nil
}
