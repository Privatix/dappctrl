package somc_test

import (
	"fmt"
	"testing"

	"github.com/privatix/dappctrl/agent/somcsrv"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"

	"github.com/privatix/dappctrl/somc"
)

func TestClientBuilder(t *testing.T) {
	usingTorSetting := &data.Setting{
		Key:   data.SettingSOMCTOR,
		Value: "true",
		Name:  data.SettingSOMCTOR,
	}
	usingDirectSetting := &data.Setting{
		Key:   data.SettingSOMCDirect,
		Value: "true",
		Name:  data.SettingSOMCDirect,
	}

	data.InsertToTestDB(t, db, usingTorSetting, usingDirectSetting)
	defer data.DeleteFromTestDB(t, db, usingTorSetting, usingDirectSetting)

	torConf := &somc.TorClientConfig{
		Socks: 9999,
	}
	builder := somc.NewClientBuilder(torConf, db)

	tests := []struct {
		somcType    uint8
		somcData    string
		usingTor    bool
		usingDirect bool
		ok          bool
	}{
		{
			somcType:    1,
			somcData:    "bXlhZGRyZXNzLm9uaW9u",
			usingTor:    true,
			usingDirect: true,
			ok:          true,
		}, // myaddress.onion
		{
			somcType:    2,
			somcData:    "bXlhZGRyZXNzLmNvbQ==",
			usingTor:    true,
			usingDirect: true,
			ok:          true,
		}, // myaddress.com
		{
			somcType:    3,
			somcData:    "bXlhZGRyZXNzLm9uaW9u&bXlhZGRyZXNzLmNvbQ==",
			usingTor:    true,
			usingDirect: true,
			ok:          true,
		}, // TOR and Direct.
		{
			somcType:    1,
			somcData:    "bXlhZGRyZXNzLm9uaW9u",
			usingTor:    false,
			usingDirect: true,
			ok:          false,
		}, // Not using TOR.
		{
			somcType:    2,
			somcData:    "bXlhZGRyZXNzLmNvbQ==",
			usingTor:    true,
			usingDirect: false,
			ok:          false,
		}, // Not using Direct.
		{
			somcType:    3,
			somcData:    "bXlhZGRyZXNzLm9uaW9u&bXlhZGRyZXNzLmNvbQ==",
			usingTor:    false,
			usingDirect: true,
			ok:          true,
		}, // Not using TOR.
		{
			somcType:    3,
			somcData:    "bXlhZGRyZXNzLm9uaW9u&bXlhZGRyZXNzLmNvbQ==",
			usingTor:    true,
			usingDirect: false,
			ok:          true,
		}, // Not using Direct.
		{
			somcType:    3,
			somcData:    "bXlhZGRyZXNzLm9uaW9u&bXlhZGRyZXNzLmNvbQ==",
			usingTor:    false,
			usingDirect: false,
			ok:          false,
		}, // Not using TOR and Direct.
	}

	for i, test := range tests {
		usingTorSetting.Value = fmt.Sprint(test.usingTor)
		usingDirectSetting.Value = fmt.Sprint(test.usingDirect)
		data.SaveToTestDB(t, db, usingTorSetting, usingDirectSetting)

		var expErr error
		if !test.ok {
			expErr = somc.ErrUnknownSOMCType
		}

		client, err := builder.NewClient(test.somcType, test.somcData)

		util.TestExpectResult(t, "NewClient", expErr, err)
		if _, ok := client.(*somcsrv.Client); ok != test.ok {
			t.Fatalf("expected client created: %v, got: %v. Test #%d", test.ok, ok, i)
		}
	}
}
