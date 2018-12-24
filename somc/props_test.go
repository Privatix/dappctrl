package somc_test

import (
	"fmt"
	"testing"

	"github.com/privatix/dappctrl/somc"

	"github.com/privatix/dappctrl/data"
)

type propsTestCase struct {
	usingTor    bool
	usingDirect bool
	somcType    uint8
	somcData    string
	err         error
}

func TestProps(t *testing.T) {
	usingTorSetting, usingDirectSetting := allTransportActiveSettings()
	data.InsertToTestDB(t, db, usingTorSetting, usingDirectSetting)
	defer data.DeleteFromTestDB(t, db, usingTorSetting, usingDirectSetting)

	props := somc.NewProps(&somc.TorAgentConfig{
		Hostname: "myaddress.onion",
	}, &somc.DirectAgentConfig{
		Addr: "myaddress.com",
	}, db)

	tests := []propsTestCase{
		{
			usingTor:    true,
			usingDirect: false,
			somcType:    1,
			somcData:    "bXlhZGRyZXNzLm9uaW9u",
			err:         nil,
		}, // TOR is active.
		{
			usingTor:    false,
			usingDirect: true,
			somcType:    2,
			somcData:    "bXlhZGRyZXNzLmNvbQ==",
			err:         nil,
		}, // Direct is active.
		{
			usingTor:    true,
			usingDirect: true,
			somcType:    3,
			somcData:    "bXlhZGRyZXNzLm9uaW9u&bXlhZGRyZXNzLmNvbQ==",
			err:         nil,
		}, // Both TOR and Direct are active.
		{
			usingTor:    false,
			usingDirect: false,
			somcType:    0,
			somcData:    "",
			err:         somc.ErrNoActiveTransport,
		}, // No active transport.
	}
	checkPropsTestCases(t, props, tests)

	props = somc.NewProps(&somc.TorAgentConfig{
		Hostname: "",
	}, &somc.DirectAgentConfig{
		Addr: "",
	}, db)
	invalidConfigTests := []propsTestCase{
		{
			usingTor:    true,
			usingDirect: false,
			somcType:    0,
			somcData:    "",
			err:         somc.ErrNoTorHostname,
		}, // TOR is active.
		{
			usingTor:    false,
			usingDirect: true,
			somcType:    0,
			somcData:    "",
			err:         somc.ErrNoDirectAddr,
		}, // Direct is active.
	}
	checkPropsTestCases(t, props, invalidConfigTests)
}

func checkPropsTestCases(t *testing.T, props *somc.Props, tests []propsTestCase) {
	usingTorSetting, usingDirectSetting := allTransportActiveSettings()
	for _, test := range tests {
		usingTorSetting.Value = fmt.Sprint(test.usingTor)
		usingDirectSetting.Value = fmt.Sprint(test.usingDirect)
		data.SaveToTestDB(t, db, usingTorSetting, usingDirectSetting)

		somcType, somcData, err := props.Get()
		if test.somcType != somcType || test.somcData != somcData || test.err != err {
			t.Fatalf("wanted somcType: %v somcData: %v err: %v,"+
				" got somcType: %v somcData: %v err: %v", test.somcType,
				test.somcData, test.err, somcType, somcData, err)
		}
	}
}
