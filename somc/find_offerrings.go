package somc

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/privatix/dappctrl/data"
)

const findOfferingsMethod = "getOfferings"

type findOfferingsParams struct {
	Hashes []data.HexString `json:"hashes"`
}

type findOfferingsResult []struct {
	Hash data.HexString    `json:"hash"`
	Data data.Base64String `json:"data"`
}

// OfferingData is a simple container for offering JSON.
type OfferingData struct {
	Hash     data.HexString
	Offering []byte
}

// FindOfferings requests SOMC to find offerings by their hashes.
func (c *Conn) FindOfferings(hashes []data.HexString) ([]OfferingData, error) {
	logger := c.logger.Add("method", "FindOfferings")

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
		hstr := data.HexFromBytes(hash.Bytes())
		if hstr != v.Hash {
			logger.Add("hashes", hashes,
				"res", res).Error("hash mismatch")
			return nil, ErrInternal
		}

		ret = append(ret, OfferingData{hstr, bytes})
	}

	return ret, nil
}
