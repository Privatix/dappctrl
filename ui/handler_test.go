package ui

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/rpc"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

type fixture struct {
	*data.TestFixture
	hash *data.Setting
	salt *data.Setting
}

func newFixture(t *testing.T) *fixture {
	fxt := fixture{TestFixture: data.NewTestFixture(t, db)}

	fxt.Offering.Agent = data.NewTestAccount(data.TestPassword).EthAddr
	fxt.Offering.OfferStatus = data.OfferRegister
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

	return &fxt
}

func (f *fixture) close() {
	data.DeleteFromTestDB(f.T, db, f.salt)
	data.DeleteFromTestDB(f.T, db, f.hash)
	f.TestFixture.Close()
}

func subscribe(client *rpc.Client, channel interface{}, method string,
	args ...interface{}) (*rpc.ClientSubscription, error) {
	return client.Subscribe(context.Background(),
		"ui", channel, append([]interface{}{method}, args...)...)
}

var db *reform.DB
var handler *Handler
var client *rpc.Client

func TestMain(m *testing.M) {
	var conf struct {
		DB      *data.DBConfig
		FileLog *log.FileConfig
		Job     *job.Config
		UI      *Config
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
	handler = NewHandler(conf.UI, logger, db, nil)
	if err := server.RegisterName("ui", handler); err != nil {
		panic(err)
	}

	client = rpc.DialInProc(server)
	defer client.Close()

	os.Exit(m.Run())
}
