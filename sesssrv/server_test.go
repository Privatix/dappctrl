// +build !nosesssrvtest

package sesssrv

import (
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/srv"
)

type testConfig struct {
	ServerStartupDelay uint // In milliseconds.
	Product            *testProduct
}

type testProduct struct {
	ValidFormatConfig map[string]string
	EmptyConfig       string
}

func newTestConfig() *testConfig {
	return &testConfig{
		ServerStartupDelay: 10,
		Product:            &testProduct{},
	}
}

var (
	conf struct {
		DB                *data.DBConfig
		Log               *util.LogConfig
		SessionServer     *Config
		SessionServerTest *testConfig
	}

	db         *reform.DB
	server     *Server
	allMethods = []string{PathAuth, PathStart, PathStop, PathUpdate,
		PathProductConfig}
)

func TestBadMethod(t *testing.T) {
	client := &http.Client{}
	for _, v := range allMethods {
		req, err := srv.NewHTTPRequest(conf.SessionServer.Config,
			http.MethodPut, v, &srv.Request{Args: nil})
		if err != nil {
			t.Fatalf("failed to create request: %s", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to send request: %s", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Fatalf("unexpected status for bad method: %d",
				resp.StatusCode)
		}
	}
}

func TestBadProductAuth(t *testing.T) {
	fxt := data.NewTestFixture(t, db)
	defer fxt.Close()

	for _, v := range allMethods {
		err := Post(conf.SessionServer.Config,
			"bad-product", "bad-password", v, nil, nil)
		util.TestExpectResult(t, "Post", srv.ErrAccessDenied, err)

		err = Post(conf.SessionServer.Config,
			fxt.Product.ID, "bad-password", v, nil, nil)
		util.TestExpectResult(t, "Post", srv.ErrAccessDenied, err)
	}
}

func TestBadClientIdent(t *testing.T) {
	fxt := data.NewTestFixture(t, db)
	defer fxt.Close()

	var args AuthArgs
	for _, v := range []string{PathAuth, PathStart, PathStop, PathUpdate} {
		args.ClientID = "bad-channel"
		fxt.Channel.ServiceStatus = data.ServiceActive
		data.SaveToTestDB(t, db, fxt.Channel)

		err := Post(conf.SessionServer.Config,
			fxt.Product.ID, data.TestPassword, v, args, nil)
		util.TestExpectResult(t, "Post", ErrChannelNotFound, err)

		args.ClientID = fxt.Channel.ID
		fxt.Channel.ServiceStatus = data.ServicePending
		data.SaveToTestDB(t, db, fxt.Channel)

		err = Post(conf.SessionServer.Config,
			fxt.Product.ID, data.TestPassword, PathAuth, args, nil)
		util.TestExpectResult(t, "Post", ErrNonActiveChannel, err)
	}
}

func TestBadProductConfig(t *testing.T) {
	fxt := data.NewTestFixture(t, db)
	defer fxt.Close()

	args := ProductArgs{}

	err := Post(conf.SessionServer.Config,
		fxt.Product.ID, data.TestPassword, PathProductConfig,
		args, nil)
	util.TestExpectResult(t, "Post", ErrProductConfNotValid, err)
}

func TestProductConfigAlreadyUploaded(t *testing.T) {
	fxt := data.NewTestFixture(t, db)
	defer fxt.Close()

	args := productConfigNormalFlow(fxt)

	err := Post(conf.SessionServer.Config,
		fxt.Product.ID, data.TestPassword, PathProductConfig,
		args, nil)
	util.TestExpectResult(t, "Post", ErrProductConfAlreadyUploaded, err)
}

func TestBadAuth(t *testing.T) {
	fxt := data.NewTestFixture(t, db)
	defer fxt.Close()

	args := AuthArgs{ClientID: fxt.Channel.ID, Password: "bad-password"}

	args.ClientID = fxt.Channel.ID
	err := Post(conf.SessionServer.Config,
		fxt.Product.ID, data.TestPassword, PathAuth, args, nil)
	util.TestExpectResult(t, "Post", ErrBadAuthPassword, err)
}

func TestBadUpdate(t *testing.T) {
	fxt := data.NewTestFixture(t, db)
	defer fxt.Close()

	args := UpdateArgs{ClientID: fxt.Channel.ID}
	err := Post(conf.SessionServer.Config,
		fxt.Product.ID, data.TestPassword, PathUpdate, args, nil)
	util.TestExpectResult(t, "Post", ErrSessionNotFound, err)
}

func testAuthNormalFlow(fxt *data.TestFixture) {
	args := AuthArgs{ClientID: fxt.Channel.ID, Password: data.TestPassword}
	err := Post(conf.SessionServer.Config,
		fxt.Product.ID, data.TestPassword, PathAuth, args, nil)
	util.TestExpectResult(fxt.T, "Post", nil, err)
}

func productConfigNormalFlow(fxt *data.TestFixture) ProductArgs {
	args := ProductArgs{
		Config: conf.SessionServerTest.Product.ValidFormatConfig}
	err := Post(conf.SessionServer.Config,
		fxt.Product.ID, data.TestPassword, PathProductConfig,
		args, nil)
	util.TestExpectResult(fxt.T, "Post", nil, err)
	return args
}

func testStartNormalFlow(fxt *data.TestFixture) *data.Session {
	const clientIP, clientPort = "1.2.3.4", 12345
	args2 := StartArgs{
		ClientID:   fxt.Channel.ID,
		ClientIP:   clientIP,
		ClientPort: clientPort,
	}

	before := time.Now()
	err := Post(conf.SessionServer.Config,
		fxt.Product.ID, data.TestPassword, PathStart, args2, nil)
	util.TestExpectResult(fxt.T, "Post", nil, err)
	after := time.Now()

	var sess data.Session
	if err := db.FindOneTo(&sess, "channel", fxt.Channel.ID); err != nil {
		fxt.T.Fatalf("cannot find new session: %s", err)
	}

	if sess.Started.Before(before) || sess.Started.After(after) {
		db.Delete(&sess)
		fxt.T.Fatalf("wrong session started time")
	}

	if sess.LastUsageTime.Before(before) ||
		sess.LastUsageTime.After(after) ||
		sess.Started != sess.LastUsageTime {
		db.Delete(&sess)
		fxt.T.Fatalf("wrong session last usage time")
	}

	if sess.ClientIP == nil || *sess.ClientIP != clientIP {
		db.Delete(&sess)
		fxt.T.Fatalf("wrong session client IP")
	}

	if sess.ClientPort == nil || *sess.ClientPort != clientPort {
		db.Delete(&sess)
		fxt.T.Fatalf("wrong session client port")
	}

	return &sess
}

func testUpdateStopNormalFlow(fxt *data.TestFixture, sess *data.Session, stop bool) {
	const units = 12345

	path := PathUpdate
	if stop {
		path = PathStop
	}

	sess.UnitsUsed = units
	data.SaveToTestDB(fxt.T, db, sess)

	for i, v := range []string{
		data.ProductUsageTotal, data.ProductUsageIncremental} {
		fxt.Product.UsageRepType = v
		data.SaveToTestDB(fxt.T, db, fxt.Product)

		before := time.Now()
		args := UpdateArgs{ClientID: fxt.Channel.ID, Units: units}
		err := Post(conf.SessionServer.Config, fxt.Product.ID,
			data.TestPassword, path, args, nil)
		util.TestExpectResult(fxt.T, "Post", nil, err)

		after := time.Now()
		data.ReloadFromTestDB(fxt.T, db, sess)

		if sess.LastUsageTime.Before(before) ||
			sess.LastUsageTime.After(after) ||
			sess.UnitsUsed != uint64((i+1)*units) {
			fxt.T.Fatalf("wrong session data after update")
		}

		if stop {
			if sess.Stopped == nil ||
				sess.Stopped.Before(before) ||
				sess.Stopped.After(after) {
				fxt.T.Fatalf("wrong session stopped time")
			}
		} else {
			if sess.Stopped != nil {
				fxt.T.Fatalf("non-nil session stopped time")
			}
		}

		sess.Stopped = nil
		data.SaveToTestDB(fxt.T, db, sess)
	}
}

func TestNormalFlow(t *testing.T) {
	fxt := data.NewTestFixture(t, db)
	defer fxt.Close()

	testAuthNormalFlow(fxt)

	productConfigNormalFlow(fxt)

	sess := testStartNormalFlow(fxt)
	defer db.Delete(sess)

	testUpdateStopNormalFlow(fxt, sess, false)
	testUpdateStopNormalFlow(fxt, sess, true)
}

func TestMain(m *testing.M) {
	conf.DB = data.NewDBConfig()
	conf.Log = util.NewLogConfig()
	conf.SessionServer = NewConfig()
	conf.SessionServerTest = newTestConfig()
	util.ReadTestConfig(&conf)

	logger := util.NewTestLogger(conf.Log)

	db = data.NewTestDB(conf.DB, logger)
	defer data.CloseDB(db)

	server = NewServer(conf.SessionServer, logger, db)
	defer server.Close()
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("failed to serve session requests: %s", err)
		}
	}()

	time.Sleep(time.Duration(conf.SessionServerTest.ServerStartupDelay) *
		time.Millisecond)

	os.Exit(m.Run())
}
