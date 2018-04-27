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
}

func newTestConfig() *testConfig {
	return &testConfig{
		ServerStartupDelay: 10,
	}
}

var (
	conf struct {
		DB                *data.DBConfig
		Log               *util.LogConfig
		SessionServer     *Config
		SessionServerTest *testConfig
	}

	db     *reform.DB
	server *Server
)

type fixture struct {
	t        *testing.T
	product  *data.Product
	account  *data.Account
	template *data.Template
	offering *data.Offering
	channel  *data.Channel
}

func newFixture(t *testing.T) *fixture {
	prod := data.NewTestProduct()
	acc := data.NewTestAccount("")
	tmpl := data.NewTestTemplate(data.TemplateOffer)
	off := data.NewTestOffering(acc.EthAddr, prod.ID, tmpl.ID)
	ch := data.NewTestChannel(
		acc.EthAddr, acc.EthAddr, off.ID, 0, 0, data.ChannelActive)

	data.InsertToTestDB(t, db, prod, acc, tmpl, off, ch)

	return &fixture{
		t:        t,
		product:  prod,
		account:  acc,
		template: tmpl,
		offering: off,
		channel:  ch,
	}
}

func (f *fixture) close() {
	for _, v := range []reform.Record{
		f.channel, f.offering, f.template, f.account, f.product} {
		if err := db.Delete(v); err != nil {
			f.t.Fatalf("failed to delete %T: %s", v, err)
		}
	}
}

func TestBadMethod(t *testing.T) {
	client := &http.Client{}
	for _, v := range []string{PathAuth, PathStart, PathStop, PathUpdate} {
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
	fix := newFixture(t)
	defer fix.close()

	for _, v := range []string{PathAuth, PathStart, PathStop, PathUpdate} {
		err := Post(conf.SessionServer.Config,
			"bad-product", "bad-password", v, nil, nil)
		util.ExpectResult(t, "Post", srv.ErrAccessDenied, err)

		err = Post(conf.SessionServer.Config,
			fix.product.ID, "bad-password", v, nil, nil)
		util.ExpectResult(t, "Post", srv.ErrAccessDenied, err)
	}
}

func TestBadClientIdent(t *testing.T) {
	fix := newFixture(t)
	defer fix.close()

	var args AuthArgs
	for _, v := range []string{PathAuth, PathStart, PathStop, PathUpdate} {
		args.ClientID = "bad-channel"
		fix.channel.ServiceStatus = data.ServiceActive
		data.SaveToTestDB(t, db, fix.channel)

		err := Post(conf.SessionServer.Config,
			fix.product.ID, data.TestPassword, v, args, nil)
		util.ExpectResult(t, "Post", ErrChannelNotFound, err)

		args.ClientID = fix.channel.ID
		fix.channel.ServiceStatus = data.ServicePending
		data.SaveToTestDB(t, db, fix.channel)

		err = Post(conf.SessionServer.Config,
			fix.product.ID, data.TestPassword, PathAuth, args, nil)
		util.ExpectResult(t, "Post", ErrNonActiveChannel, err)
	}
}

func TestBadAuth(t *testing.T) {
	fix := newFixture(t)
	defer fix.close()

	args := AuthArgs{ClientID: fix.channel.ID, Password: "bad-password"}

	args.ClientID = fix.channel.ID
	err := Post(conf.SessionServer.Config,
		fix.product.ID, data.TestPassword, PathAuth, args, nil)
	util.ExpectResult(t, "Post", ErrBadAuthPassword, err)
}

func TestBadUpdate(t *testing.T) {
	fix := newFixture(t)
	defer fix.close()

	args := UpdateArgs{ClientID: fix.channel.ID}
	err := Post(conf.SessionServer.Config,
		fix.product.ID, data.TestPassword, PathUpdate, args, nil)
	util.ExpectResult(t, "Post", ErrSessionNotFound, err)
}

func testAuthNormalFlow(fix *fixture) {
	args := AuthArgs{ClientID: fix.channel.ID, Password: data.TestPassword}
	err := Post(conf.SessionServer.Config,
		fix.product.ID, data.TestPassword, PathAuth, args, nil)
	util.ExpectResult(fix.t, "Post", nil, err)
}

func testStartNormalFlow(fix *fixture) *data.Session {
	const clientIP, clientPort = "1.2.3.4", 12345
	args2 := StartArgs{
		ClientID:   fix.channel.ID,
		ClientIP:   clientIP,
		ClientPort: clientPort,
	}

	before := time.Now()
	err := Post(conf.SessionServer.Config,
		fix.product.ID, data.TestPassword, PathStart, args2, nil)
	util.ExpectResult(fix.t, "Post", nil, err)
	after := time.Now()

	var sess data.Session
	if err := db.FindOneTo(&sess, "channel", fix.channel.ID); err != nil {
		fix.t.Fatalf("cannot find new session: %s", err)
	}

	if sess.Started.Before(before) || sess.Started.After(after) {
		db.Delete(&sess)
		fix.t.Fatalf("wrong session started time")
	}

	if sess.LastUsageTime.Before(before) ||
		sess.LastUsageTime.After(after) ||
		sess.Started != sess.LastUsageTime {
		db.Delete(&sess)
		fix.t.Fatalf("wrong session last usage time")
	}

	if sess.ClientIP == nil || *sess.ClientIP != clientIP {
		db.Delete(&sess)
		fix.t.Fatalf("wrong session client IP")
	}

	if sess.ClientPort == nil || *sess.ClientPort != clientPort {
		db.Delete(&sess)
		fix.t.Fatalf("wrong session client port")
	}

	return &sess
}

func testUpdateStopNormalFlow(fix *fixture, sess *data.Session, stop bool) {
	const units = 12345

	path := PathUpdate
	if stop {
		path = PathStop
	}

	sess.UnitsUsed = units
	data.SaveToTestDB(fix.t, db, sess)

	for i, v := range []string{
		data.ProductUsageTotal, data.ProductUsageIncremental} {
		fix.product.UsageRepType = v
		data.SaveToTestDB(fix.t, db, fix.product)

		before := time.Now()
		args := UpdateArgs{ClientID: fix.channel.ID, Units: units}
		err := Post(conf.SessionServer.Config, fix.product.ID,
			data.TestPassword, path, args, nil)
		util.ExpectResult(fix.t, "Post", nil, err)

		after := time.Now()
		data.ReloadFromTestDB(fix.t, db, sess)

		if sess.LastUsageTime.Before(before) ||
			sess.LastUsageTime.After(after) ||
			sess.UnitsUsed != uint64((i+1)*units) {
			fix.t.Fatalf("wrong session data after update")
		}

		if stop {
			if sess.Stopped == nil ||
				sess.Stopped.Before(before) ||
				sess.Stopped.After(after) {
				fix.t.Fatalf("wrong session stopped time")
			}
		} else {
			if sess.Stopped != nil {
				fix.t.Fatalf("non-nil session stopped time")
			}
		}

		sess.Stopped = nil
		data.SaveToTestDB(fix.t, db, sess)
	}
}

func TestNormalFlow(t *testing.T) {
	fix := newFixture(t)
	defer fix.close()

	testAuthNormalFlow(fix)

	sess := testStartNormalFlow(fix)
	defer db.Delete(sess)

	testUpdateStopNormalFlow(fix, sess, false)
	testUpdateStopNormalFlow(fix, sess, true)
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
