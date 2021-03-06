package ui_test

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/rpc"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/client/somc"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

type gasPriceSuggestor struct {
	v uint64
}

func (s gasPriceSuggestor) SuggestGasPrice(_ context.Context) (*big.Int, error) {
	return new(big.Int).SetUint64(s.v), nil
}

var (
	conf struct {
		DB   *data.DBConfig
		Log  *log.WriterConfig
		Job  *job.Config
		Proc *proc.Config
	}
	logger log.Logger

	db                    *reform.DB
	handler               *ui.Handler
	client                *rpc.Client
	testSOMCClient        *somc.TestClient
	testToken             *dumbToken
	testGasPriceSuggestor gasPriceSuggestor
)

type dumbToken struct {
	v string
}

// Contract.
var _ ui.TokenMakeChecker = new(dumbToken)

func (t *dumbToken) Make() (string, error) {
	t.v = fmt.Sprint(rand.Int())
	return t.v, nil
}

func (t *dumbToken) Check(s string) bool {
	return s == t.v
}

type fixture struct {
	*data.TestFixture
	hash *data.Setting
	salt *data.Setting
}

func newTest(t *testing.T, method string) (*fixture, func(error, error)) {
	testToken.Make()

	fxt := fixture{TestFixture: data.NewTestFixture(t, db)}
	fxt.Offering.Agent = data.NewTestAccount(data.TestPassword).EthAddr
	fxt.Offering.Status = data.OfferRegistered
	data.SaveToTestDB(t, db, fxt.Offering)

	hash, err := data.HashPassword(
		data.TestPassword, fmt.Sprint(data.TestSalt))
	util.TestExpectResult(t, "HashPassword", nil, err)
	fxt.hash = &data.Setting{
		Key:   data.SettingPasswordHash,
		Name:  "hash",
		Value: string(hash),
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

func subscribe(client *rpc.Client, channel interface{}, method string,
	args ...interface{}) (*rpc.ClientSubscription, error) {
	return client.Subscribe(context.Background(),
		"ui", channel, append([]interface{}{method}, args...)...)
}

func setTestJobQueueToExpectJobAdd(t *testing.T, j *data.Job) job.QueueMock {
	m := job.QueueMock(func(method int, tx *reform.TX,
		j2 *data.Job, relatedIDs []string, subID string,
		subFunc job.SubFunc) error {
		switch method {
		case job.MockAdd:
			*j = *j2
			t.Log("HERE", j)
		default:
			t.Fatal("unexpected queue call")
		}
		return nil
	})
	handler.SetMockQueue(m)
	return m
}

func TestMain(m *testing.M) {
	var err error

	conf.DB = data.NewDBConfig()
	conf.Log = log.NewWriterConfig()
	conf.Proc = proc.NewConfig()

	args := &util.TestArgs{
		Conf: &conf,
	}
	util.ReadTestArgs(args)

	logger, err = log.NewTestLogger(conf.Log, args.Verbose)
	if err != nil {
		panic(err)
	}

	db = data.NewTestDB(conf.DB)
	defer data.CloseDB(db)

	server := rpc.NewServer()
	pwdStorage := data.NewPWDStorage(data.TestToPrivateKey)
	testSOMCClient = somc.NewTestClient()
	testToken = &dumbToken{}
	handler = ui.NewHandler(logger, db, nil, pwdStorage,
		data.TestEncryptedKey, data.RoleAgent, nil,
		somc.NewTestClientBuilder(testSOMCClient), testToken, &testGasPriceSuggestor)
	if err := server.RegisterName("ui", handler); err != nil {
		panic(err)
	}

	client = rpc.DialInProc(server)
	defer client.Close()

	os.Exit(m.Run())
}
