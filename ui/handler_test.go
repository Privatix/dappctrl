package ui_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/rpc"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

type fixture struct {
	*data.TestFixture
	hash *data.Setting
	salt *data.Setting
}

func newTest(t *testing.T, method string) (*fixture, func(error, error)) {
	fxt := fixture{TestFixture: data.NewTestFixture(t, db)}
	fxt.Offering.Agent = data.NewTestAccount(data.TestPassword).EthAddr
	fxt.Offering.OfferStatus = data.OfferRegistered
	fxt.Offering.Status = data.MsgChPublished
	data.SaveToTestDB(t, db, fxt.Offering)

	hash, err := data.HashPassword(
		data.TestPassword, fmt.Sprint(data.TestSalt))
	util.TestExpectResult(t, "HashPassword", nil, err)
	fxt.hash = &data.Setting{
		Key:   data.SettingPasswordHash,
		Name:  "hash",
		Value: hash,
	}
	fxt.salt = &data.Setting{
		Key:   data.SettingPasswordSalt,
		Name:  "salt",
		Value: fmt.Sprint(data.TestSalt),
	}
	data.SaveToTestDB(t, db, fxt.hash)
	data.SaveToTestDB(t, db, fxt.salt)

	return &fxt, func(expected, actual error) {
		util.TestExpectResult(t, method, expected, actual)
	}
}

func (f *fixture) close() {
	data.DeleteFromTestDB(f.T, db, f.salt)
	data.DeleteFromTestDB(f.T, db, f.hash)
	f.TestFixture.Close()
}

// insertDefaultGasPriceSetting inserts default gas price settings and returns
// clean up function.
func insertDefaultGasPriceSetting(t *testing.T, v uint64) func() {
	rec := &data.Setting{
		Key:   data.SettingDefaultGasPrice,
		Value: fmt.Sprint(v),
	}
	data.InsertToTestDB(t, db, rec)
	return func() { data.DeleteFromTestDB(t, db, rec) }
}

func subscribe(client *rpc.Client, channel interface{}, method string,
	args ...interface{}) (*rpc.ClientSubscription, error) {
	return client.Subscribe(context.Background(),
		"ui", channel, append([]interface{}{method}, args...)...)
}

var db *reform.DB
var handler *ui.Handler
var client *rpc.Client

func TestMain(m *testing.M) {
	var conf struct {
		DB      *data.DBConfig
		FileLog *log.FileConfig
		Job     *job.Config
		UI      *ui.Config
	}

	conf.DB = data.NewDBConfig()
	conf.FileLog = log.NewFileConfig()
	util.ReadTestConfig(&conf)

	db = data.NewTestDB(conf.DB)
	defer data.CloseDB(db)

	logger, err := log.NewStderrLogger(conf.FileLog)
	if err != nil {
		panic(err.Error())
	}

	server := rpc.NewServer()
	pwdStorage := new(data.PWDStorage)
	handler = ui.NewHandler(conf.UI, logger, db, nil, pwdStorage,
		data.TestEncryptedKey, data.TestToPrivateKey)
	if err := server.RegisterName("ui", handler); err != nil {
		panic(err)
	}

	client = rpc.DialInProc(server)
	defer client.Close()

	os.Exit(m.Run())
}
