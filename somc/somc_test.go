package somc_test

import (
	"os"
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
	reform "gopkg.in/reform.v1"
)

var (
	db *reform.DB
)

func allTransportActiveSettings() (*data.Setting, *data.Setting) {
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
	return usingTorSetting, usingDirectSetting
}

func TestMain(m *testing.M) {
	var (
		conf struct {
			DB *data.DBConfig
		}
	)
	conf.DB = data.NewDBConfig()
	args := &util.TestArgs{
		Conf: &conf,
	}
	util.ReadTestArgs(args)
	db = data.NewTestDB(conf.DB)
	defer data.CloseDB(db)

	os.Exit(m.Run())
}
