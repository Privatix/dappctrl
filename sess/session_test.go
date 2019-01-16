package sess_test

import (
	"testing"
	"time"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/sess"
	"github.com/privatix/dappctrl/util"
)

func TestAuthClient(t *testing.T) {
	fxt := newTestFixture(t)
	defer fxt.Close()

	err := handler.AuthClient("bad-channel", data.TestPassword,
		fxt.Channel.ID, data.TestPassword)
	util.TestExpectResult(t, "AuthClient", sess.ErrAccessDenied, err)

	err = handler.AuthClient(fxt.Product.ID, "bad-password",
		fxt.Channel.ID, data.TestPassword)
	util.TestExpectResult(t, "AuthClient", sess.ErrAccessDenied, err)

	err = handler.AuthClient(fxt.Product.ID, data.TestPassword,
		"bad-channel", data.TestPassword)
	util.TestExpectResult(t, "AuthClient", sess.ErrChannelNotFound, err)

	err = handler.AuthClient(fxt.Product.ID, data.TestPassword,
		fxt.Channel.ID, "bad-password")
	util.TestExpectResult(t, "AuthClient", sess.ErrBadClientPassword, err)

	err = handler.AuthClient(fxt.Product.ID, data.TestPassword,
		fxt.Channel.ID, data.TestPassword)
	util.TestExpectResult(t, "AuthClient", nil, err)
}

func TestStartSession(t *testing.T) {
	fxt := newTestFixture(t)
	defer fxt.Close()

	const ip, port = "1.2.3.4", 1234

	_, err := handler.StartSession("bad-channel", data.TestPassword,
		fxt.Channel.ID, ip, port)
	util.TestExpectResult(t, "StartSession", sess.ErrAccessDenied, err)

	_, err = handler.StartSession(fxt.Product.ID, "bad-password",
		fxt.Channel.ID, ip, port)
	util.TestExpectResult(t, "StartSession", sess.ErrAccessDenied, err)

	_, err = handler.StartSession(fxt.Product.ID, data.TestPassword,
		"bad-channel", ip, port)
	util.TestExpectResult(t, "StartSession", sess.ErrChannelNotFound, err)

	before := time.Now()
	offer, err := handler.StartSession(
		fxt.Product.ID, data.TestPassword, fxt.Channel.ID, ip, port)
	util.TestExpectResult(fxt.T, "StartSession", nil, err)
	after := time.Now()

	if offer == nil || offer.ID != fxt.Channel.Offering {
		fxt.T.Fatal("bad start result offering")
	}

	var sess data.Session
	if err := db.FindOneTo(&sess, "channel", fxt.Channel.ID); err != nil {
		fxt.T.Fatalf("cannot find new session: %s", err)
	}
	defer db.Delete(&sess)

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

	if sess.ClientIP == nil || *sess.ClientIP != ip {
		db.Delete(&sess)
		fxt.T.Fatalf("wrong session client IP")
	}

	if sess.ClientPort == nil || *sess.ClientPort != port {
		db.Delete(&sess)
		fxt.T.Fatalf("wrong session client port")
	}
}

func testUpdateSession(
	fxt *data.TestFixture, sess *data.Session, stopSession bool) {
	const units = 12345

	sess.UnitsUsed = units
	data.SaveToTestDB(fxt.T, db, sess)

	for i, v := range []string{
		data.ProductUsageTotal, data.ProductUsageIncremental} {
		fxt.Product.UsageRepType = v
		data.SaveToTestDB(fxt.T, db, fxt.Product)

		before := time.Now()
		err := handler.UpdateSession(fxt.Product.ID, data.TestPassword,
			fxt.Channel.ID, units, stopSession)
		util.TestExpectResult(fxt.T, "UpdateSession", nil, err)

		after := time.Now()
		data.ReloadFromTestDB(fxt.T, db, sess)

		if sess.LastUsageTime.Before(before) ||
			sess.LastUsageTime.After(after) ||
			sess.UnitsUsed != (uint64((i+1)*units)/1024/1024) {
			fxt.T.Fatalf("wrong session data after update")
		}

		if stopSession {
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

func TestUpdateSession(t *testing.T) {
	fxt := newTestFixture(t)
	defer fxt.Close()

	err := handler.UpdateSession("bad-channel", data.TestPassword,
		fxt.Channel.ID, 0, false)
	util.TestExpectResult(t, "UpdateSession", sess.ErrAccessDenied, err)

	err = handler.UpdateSession(fxt.Product.ID, "bad-password",
		fxt.Channel.ID, 0, false)
	util.TestExpectResult(t, "UpdateSession", sess.ErrAccessDenied, err)

	err = handler.UpdateSession(fxt.Product.ID, data.TestPassword,
		"bad-channel", 0, false)
	util.TestExpectResult(t, "UpdateSession", sess.ErrChannelNotFound, err)

	_, err = handler.StartSession(fxt.Product.ID, data.TestPassword,
		fxt.Channel.ID, "1.2.3.4", 1234)
	util.TestExpectResult(t, "Start", nil, err)

	var sess data.Session
	if err := db.FindOneTo(&sess, "channel", fxt.Channel.ID); err != nil {
		fxt.T.Fatalf("cannot find new session: %s", err)
	}
	defer db.Delete(&sess)

	testUpdateSession(fxt, &sess, false)
	testUpdateSession(fxt, &sess, true)
}
