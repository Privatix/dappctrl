// +build !novpnsrvtest

package srv

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
	vpnutil "github.com/privatix/dappctrl/vpn/util"
)

var conf struct {
	DB        *data.DBConfig
	Log       *util.LogConfig
	TestData  *data.TestData
	VPNServer *Config
}

var db *reform.DB
var srv *Server

func postRaw(t *testing.T, path string, data []byte, rep interface{}) {
	var proto = "http"
	if conf.VPNServer.TLS != nil {
		proto += "s"
	}

	resp, err := http.Post(proto+"://"+conf.VPNServer.Addr+path,
		"application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to post request: %s", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(rep); err != nil {
		t.Fatalf("failed to decode reply: %s", err)
	}
}

func post(t *testing.T, path string, req interface{}, rep interface{}) {
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to encode request: %s", err)
	}

	postRaw(t, path, data, rep)
}

func TestNonParsableRequest(t *testing.T) {
	for _, v := range []string{PathAuth, PathStart, PathStop} {
		var repl AuthReply
		postRaw(t, v, []byte("{"), &repl)

		if repl.Error != ErrFailedToParseRequest {
			msg := "unexpected error for non-parsable request: %s"
			t.Fatalf(msg, repl.Error)
		}
	}
}

const (
	serverIP   = "1.2.3.4"
	clientIP   = "4.3.2.1"
	clientPort = 1234
	uploaded   = 2345
	downloaded = 5432
)

func TestMalformedRequest(t *testing.T) {
	for _, v := range []StartRequest{
		{ClientIP: clientIP, ClientPort: clientPort},
		{ServerIP: serverIP, ClientPort: clientPort},
		{ServerIP: serverIP, ClientIP: clientIP},
	} {
		var repl AuthReply
		post(t, PathStart, &v, &repl)

		if repl.Error != ErrMalformedRequest {
			t.Fatalf("unexpected error for malformed request: %s",
				repl.Error)
		}
	}
}

func newStartRequest(ch string) StartRequest {
	return StartRequest{
		Channel:    ch,
		ServerIP:   serverIP,
		ClientIP:   clientIP,
		ClientPort: clientPort,
	}
}

func TestObjectNotFound(t *testing.T) {
	type pathReq struct {
		path string
		req  interface{}
	}

	for _, v := range []string{"", "foo", util.NewUUID()} {
		for _, v2 := range []pathReq{
			pathReq{PathAuth, AuthRequest{Channel: v}},
			pathReq{PathStart, newStartRequest(v)},
			pathReq{PathStop, StopRequest{Channel: v}},
		} {
			var repl errorReply
			post(t, v2.path, &v2.req, &repl)

			if repl.Error != ErrObjectNotFound {
				t.Fatalf("unexpected error for non-existing "+
					"channel for %s: %s", v, repl.Error)
			}
		}
	}
}

func TestAccessDenied(t *testing.T) {
	for _, v := range []string{"", "foo"} {
		req := AuthRequest{Channel: conf.TestData.Channel, Password: v}
		var repl AuthReply
		post(t, PathAuth, &req, &repl)

		if repl.Error != ErrAccessDenied {
			t.Fatalf("unexpected error for wrong password: %s",
				repl.Error)
		}
	}
}

func testRegularAuth(t *testing.T) {
	req := AuthRequest{
		Channel:  conf.TestData.Channel,
		Password: conf.TestData.Password,
	}
	var repl AuthReply
	post(t, PathAuth, &req, &repl)
	if len(repl.Error) != 0 {
		t.Fatalf("unexpected authentication error: %s", repl.Error)
	}
}

func testRegularStart(t *testing.T) string {
	before := time.Now()
	var repl StartReply
	post(t, PathStart, newStartRequest(conf.TestData.Channel), &repl)
	if len(repl.Error) != 0 {
		t.Fatalf("unexpected error while starting session: %s",
			repl.Error)
	}

	var sess data.Session
	sid, _ := vpnutil.FindCurrentSession(db, conf.TestData.Channel)
	if err := db.FindByPrimaryKeyTo(&sess, sid); err != nil {
		t.Fatalf("failed to find session: %s", err)
	}

	if sess.Started.Before(before) || sess.Started.After(time.Now()) {
		t.Fatalf("wrong session start time")
	}

	var vsess data.VPNSession
	if err := db.FindByPrimaryKeyTo(&vsess, sid); err != nil {
		t.Fatalf("failed to find vpn session: %s", err)
	}

	if vsess.ServerIP != serverIP || vsess.ClientIP != clientIP ||
		vsess.ClientPort != clientPort {
		t.Fatalf("wrong vpn session client/server info")
	}

	return sid
}

func testRegularStop(t *testing.T, sid string) {
	req := StopRequest{
		Channel:    conf.TestData.Channel,
		Uploaded:   uploaded,
		Downloaded: downloaded,
	}
	before := time.Now()
	var repl StopReply
	post(t, PathStop, req, &repl)
	if len(repl.Error) != 0 {
		t.Fatalf("unexpected error while stopping session: %s",
			repl.Error)
	}

	var sess data.Session
	if err := db.FindByPrimaryKeyTo(&sess, sid); err != nil {
		t.Fatalf("failed to find session: %s", err)
	}

	if sess.Stopped.Before(before) || sess.Stopped.After(time.Now()) {
		t.Fatalf("wrong session stop time")
	}

	var vsess data.VPNSession
	if err := db.FindByPrimaryKeyTo(&vsess, sid); err != nil {
		t.Fatalf("failed to find vpn session: %s", err)
	}

	if vsess.Uploaded != uploaded || vsess.Downloaded != downloaded {
		t.Fatalf("wrong vpn session uploaded/downloaded values")
	}
}

func TestNonOpenChannel(t *testing.T) {
	vpnutil.SetChannelState(t, db,
		conf.TestData.Channel, data.ChannelClosing)
	defer vpnutil.SetChannelState(t, db,
		conf.TestData.Channel, data.ChannelOpen)

	var repl StartReply
	post(t, PathStart, newStartRequest(conf.TestData.Channel), &repl)
	if repl.Error != ErrNonOpenChannel {
		t.Fatalf("unexpected error while starting session "+
			"for non-open channel: %s", repl.Error)
	}
}

func TestRegularFlow(t *testing.T) {
	testRegularAuth(t)
	sid := testRegularStart(t)
	testRegularStop(t, sid)
}

func TestMain(m *testing.M) {
	conf.DB = data.NewDBConfig()
	conf.Log = util.NewLogConfig()
	conf.TestData = &data.TestData{}
	conf.VPNServer = NewConfig()
	util.ReadTestConfig(&conf)

	logger := util.NewTestLogger(conf.Log)

	db = data.NewTestDB(conf.DB, logger)
	defer data.CloseDB(db)

	srv = NewServer(conf.VPNServer, logger, db)
	defer srv.Close()
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("failed to serve vpn session requests: %s\n",
				err)
		}
	}()

	os.Exit(m.Run())
}
