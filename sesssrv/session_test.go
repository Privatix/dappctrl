// +build !nosesssrvtest

package sesssrv

import (
	"testing"
	"time"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

func TestBadClientIdent(t *testing.T) {
	fxt := newTestFixtures(t)
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

func TestBadAuth(t *testing.T) {
	fxt := newTestFixtures(t)
	defer fxt.Close()

	args := AuthArgs{ClientID: fxt.Channel.ID, Password: "bad-password"}

	args.ClientID = fxt.Channel.ID
	err := Post(conf.SessionServer.Config,
		fxt.Product.ID, data.TestPassword, PathAuth, args, nil)
	util.TestExpectResult(t, "Post", ErrBadAuthPassword, err)
}

func TestBadUpdate(t *testing.T) {
	fxt := newTestFixtures(t)
	defer fxt.Close()

	args := UpdateArgs{ClientID: fxt.Channel.ID}
	err := Post(conf.SessionServer.Config,
		fxt.Product.ID, data.TestPassword, PathUpdate, args, nil)
	util.TestExpectResult(t, "Post", ErrSessionNotFound, err)
}

func TestNormalSessionFlow(t *testing.T) {
	fxt := newTestFixtures(t)
	defer fxt.Close()

	testAuthNormalFlow(fxt)

	sess := testStartNormalFlow(fxt)
	defer db.Delete(sess)

	testUpdateStopNormalFlow(fxt, sess, false)
	testUpdateStopNormalFlow(fxt, sess, true)
}

func testAuthNormalFlow(fxt *data.TestFixture) {
	args := AuthArgs{ClientID: fxt.Channel.ID, Password: data.TestPassword}
	err := Post(conf.SessionServer.Config,
		fxt.Product.ID, data.TestPassword, PathAuth, args, nil)
	util.TestExpectResult(fxt.T, "Post", nil, err)
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
