package sess_test

import (
	"testing"
	"time"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/sess"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
	"gopkg.in/reform.v1"
)

func TestServiceReady(t *testing.T) {
	fxt := newTestFixture(t)
	defer fxt.Close()

	prod := fxt.Product.ID
	prodPass := data.TestPassword
	clientKey := fxt.Channel.ID

	t.Run("ErrAccessDenied", func(t *testing.T) {
		err := handler.ServiceReady("bad-product", prodPass, clientKey)
		util.TestExpectResult(t, "ServiceReady", sess.ErrAccessDenied, err)

		err = handler.ServiceReady(prod, "bad-password", clientKey)
		util.TestExpectResult(t, "ServiceReady", sess.ErrAccessDenied, err)
	})

	t.Run("ErrChannelNotFound", func(t *testing.T) {
		err := handler.ServiceReady(prod, prodPass, "bad-channel")
		util.TestExpectResult(t, "ServiceReady", sess.ErrChannelNotFound, err)
	})

	t.Run("OK", func(t *testing.T) {
		jobRecieved := make(chan *data.Job)
		queueMock := job.QueueMock(func(method int, tx *reform.TX,
			j *data.Job, _ []string, _ string, _ job.SubFunc) error {
			if method != job.MockAdd {
				t.Errorf("method=%v, want %v", method, job.MockAdd)
			}
			if got, exp := j.Type, data.JobCompleteServiceTransition; got != exp {
				t.Errorf("job.Type=%s, want %s", j.Type, exp)
			}
			close(jobRecieved)
			return nil
		})

		h := sess.NewHandler(log.NewMultiLogger(),
			db, newTestCountryConfig(), queueMock)
		fxt.Channel.ServiceStatus = data.ServiceActivating
		data.SaveToTestDB(t, fxt.DB, fxt.Channel)
		err := h.ServiceReady(prod, prodPass, clientKey)
		util.TestExpectResult(t, "ServiceReady", nil, err)

		select {
		case <-jobRecieved:
		case <-time.After(time.Second):
			t.Fatal("Timeout: job was not added")
		}
	})
}

func TestAuthClient(t *testing.T) {
	fxt := newTestFixture(t)
	defer fxt.Close()

	t.Run("ErrAccessDenied", func(t *testing.T) {
		err := handler.AuthClient("bad-product", data.TestPassword,
			fxt.Channel.ID, data.TestPassword)
		util.TestExpectResult(t, "AuthClient", sess.ErrAccessDenied, err)

		err = handler.AuthClient(fxt.Product.ID, "bad-password",
			fxt.Channel.ID, data.TestPassword)
		util.TestExpectResult(t, "AuthClient", sess.ErrAccessDenied, err)
	})

	t.Run("ErrChannelNotFound", func(t *testing.T) {
		err := handler.AuthClient(fxt.Product.ID, data.TestPassword,
			"bad-channel", data.TestPassword)
		util.TestExpectResult(t, "AuthClient", sess.ErrChannelNotFound, err)
	})

	t.Run("ErrBadClientPassword", func(t *testing.T) {
		err := handler.AuthClient(fxt.Product.ID, data.TestPassword,
			fxt.Channel.ID, "bad-password")
		util.TestExpectResult(t, "AuthClient", sess.ErrBadClientPassword, err)
	})

	t.Run("ErrNonActiveChannel", func(t *testing.T) {
		fxt.Channel.ServiceStatus = data.ServiceSuspended
		data.SaveToTestDB(t, fxt.DB, fxt.Channel)
		defer func() {
			fxt.Channel.ServiceStatus = data.ServiceActive
			data.SaveToTestDB(t, fxt.DB, fxt.Channel)
		}()
		err := handler.AuthClient(fxt.Product.ID, data.TestPassword,
			fxt.Channel.ID, data.TestPassword)
		util.TestExpectResult(t, "AuthClient", sess.ErrNonActiveChannel, err)
	})

	t.Run("OK", func(t *testing.T) {
		err := handler.AuthClient(fxt.Product.ID, data.TestPassword,
			fxt.Channel.ID, data.TestPassword)
		util.TestExpectResult(t, "AuthClient", nil, err)
	})
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

func TestUpdateSession(t *testing.T) {
	fxt := newTestFixture(t)
	defer fxt.Close()

	t.Run("ErrAccessDenied", func(t *testing.T) {
		err := handler.UpdateSession("bad-channel", data.TestPassword,
			fxt.Channel.ID, 0)
		util.TestExpectResult(t, "UpdateSession", sess.ErrAccessDenied, err)

		err = handler.UpdateSession(fxt.Product.ID, "bad-password",
			fxt.Channel.ID, 0)
		util.TestExpectResult(t, "UpdateSession", sess.ErrAccessDenied, err)
	})

	t.Run("ErrChannelNotFound", func(t *testing.T) {
		err := handler.UpdateSession(fxt.Product.ID, data.TestPassword,
			"bad-channel", 0)
		util.TestExpectResult(t, "UpdateSession", sess.ErrChannelNotFound, err)
	})

	t.Run("OK", func(t *testing.T) {
		_, err := handler.StartSession(fxt.Product.ID, data.TestPassword,
			fxt.Channel.ID, "1.2.3.4", 1234)
		util.TestExpectResult(t, "Start", nil, err)

		var sess data.Session
		if err := db.FindOneTo(&sess, "channel", fxt.Channel.ID); err != nil {
			fxt.T.Fatalf("cannot find new session: %s", err)
		}
		defer db.Delete(&sess)

		const units = 12345

		sess.UnitsUsed = units
		data.SaveToTestDB(t, fxt.DB, &sess)

		for i, v := range []string{data.ProductUsageTotal,
			data.ProductUsageIncremental} {
			fxt.Product.UsageRepType = v
			data.SaveToTestDB(fxt.T, fxt.DB, fxt.Product)

			before := time.Now()
			err := handler.UpdateSession(fxt.Product.ID, data.TestPassword,
				fxt.Channel.ID, units)
			util.TestExpectResult(fxt.T, "UpdateSession", nil, err)

			after := time.Now()
			data.ReloadFromTestDB(t, fxt.DB, &sess)

			if sess.LastUsageTime.Before(before) ||
				sess.LastUsageTime.After(after) ||
				sess.UnitsUsed != (uint64((i+1)*units)/1024/1024) {
				fxt.T.Fatalf("wrong session data after update")
			}
		}
	})
}

func TestStopSession(t *testing.T) {
	fxt := newTestFixture(t)
	defer fxt.Close()

	t.Run("ErrAccessDenied", func(t *testing.T) {
		err := handler.StopSession("bad-channel", data.TestPassword, fxt.Channel.ID)
		util.TestExpectResult(t, "StopSession", sess.ErrAccessDenied, err)

		err = handler.StopSession(fxt.Product.ID, "bad-password", fxt.Channel.ID)
		util.TestExpectResult(t, "StopSession", sess.ErrAccessDenied, err)
	})

	t.Run("ErrChannelNotFound", func(t *testing.T) {
		err := handler.StopSession(fxt.Product.ID, data.TestPassword, "bad-channel")
		util.TestExpectResult(t, "StopSession", sess.ErrChannelNotFound, err)
	})

	t.Run("OK", func(t *testing.T) {
		_, err := handler.StartSession(fxt.Product.ID, data.TestPassword,
			fxt.Channel.ID, "1.2.3.4", 1234)
		util.TestExpectResult(t, "Start", nil, err)

		var sess data.Session
		if err := db.FindOneTo(&sess, "channel", fxt.Channel.ID); err != nil {
			fxt.T.Fatalf("cannot find new session: %s", err)
		}
		defer db.Delete(&sess)

		beforeStop := time.Now()
		err = handler.StopSession(fxt.Product.ID, data.TestPassword, fxt.Channel.ID)
		util.TestExpectResult(t, "StopSession", nil, err)

		fxt.DB.Reload(&sess)

		if sess.Stopped == nil ||
			sess.Stopped.Before(beforeStop) ||
			sess.Stopped.After(time.Now()) {
			fxt.T.Fatalf("wrong session stopped time")
		}
	})
}
