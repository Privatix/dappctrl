package somc

import (
	"github.com/privatix/dappctrl/data"
	reform "gopkg.in/reform.v1"
)

// Props is agent offerings' somc props maker.
type Props struct {
	directConf *DirectAgentConfig
	db         *reform.DB
	torConf    *TorAgentConfig
}

// NewProps creates new builder.
func NewProps(torConf *TorAgentConfig, directConf *DirectAgentConfig, db *reform.DB) *Props {
	return &Props{directConf, db, torConf}
}

// Get computes and returns somc properties.
func (p *Props) Get() (uint8, string, error) {
	transports := []struct {
		code   uint8
		active func() (bool, error)
		data   func() (data.Base64String, error)
	}{
		{torCode, p.torIsActive, p.torData},
		{directCode, p.directIsActive, p.directData},
	}

	somcType := uint8(0)
	dataParts := []string{}
	for _, transport := range transports {
		ok, err := transport.active()
		if err != nil {
			return 0, "", err
		}

		if ok {
			d, err := transport.data()
			if err != nil {
				return 0, "", err
			}
			dataParts = append(dataParts, string(d))

			somcType += transport.code
		}
	}

	if somcType == 0 {
		return 0, "", ErrNoActiveTransport
	}

	return somcType, combineURLBase64Strings(dataParts), nil
}

func (p *Props) torIsActive() (bool, error) {
	return data.ReadBoolSetting(p.db.Querier, data.SettingSOMCTOR)
}

func (p *Props) directIsActive() (bool, error) {
	return data.ReadBoolSetting(p.db.Querier, data.SettingSOMCDirect)
}

func (p *Props) torData() (data.Base64String, error) {
	hostname := p.torConf.Hostname
	if hostname == "" {
		return "", ErrNoTorHostname
	}
	return data.FromBytes([]byte(hostname)), nil
}

func (p *Props) directData() (data.Base64String, error) {
	addr := p.directConf.Addr
	if addr == "" {
		return "", ErrNoDirectAddr
	}
	return data.FromBytes([]byte(addr)), nil
}
