package uisrv

import (
	"database/sql"
	"fmt"
	"net/http"

	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

const (
	passwordKey = "system.password"
	saltKey     = "system.salt"

	passwordMinLen = 8
	passwordMaxLen = 24
)

// basicAuthMiddleware implements HTTP Basic Authentication check.
// If no password stored replies with 401 and serverError.Code=1.
// On wrong password replies with 401.
func basicAuthMiddleware(s *Server, h http.HandlerFunc) http.HandlerFunc {
	logger := s.logger.Add("method", "basicAuthMiddleware")

	return func(w http.ResponseWriter, r *http.Request) {
		_, givenPassword, ok := r.BasicAuth()
		if !ok {
			s.replyErr(logger, w, http.StatusUnauthorized, &serverError{
				Message: "Wrong password",
			})
			return
		}

		if !s.correctPassword(logger, w, givenPassword) {
			return
		}

		// Make password available through storage.
		s.pwdStorage.Set(givenPassword)

		h(w, r)
	}
}

type passwordPayload struct {
	Password string `json:"password"`
}

func (s *Server) handleAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.handleSetPassword(w, r)
		return
	}
	if r.Method == http.MethodPut {
		basicAuthMiddleware(s, s.handleUpdatePassword)(w, r)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (s *Server) handleSetPassword(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.Add("method", "handleSetPassword")

	payload := &passwordPayload{}
	if !s.parsePasswordPayload(logger, w, r, payload) ||
		!s.setPasswordAllowed(w) {
		return
	}

	tx, ok := s.begin(logger, w)
	if !ok {
		return
	}

	if !s.setPassword(w, payload.Password, tx) {
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) parsePasswordPayload(logger log.Logger, w http.ResponseWriter,
	r *http.Request, payload *passwordPayload) bool {
	return s.parsePayload(logger, w, r, payload) &&
		s.validPasswordString(logger, w, payload.Password)
}

type newPasswordPayload struct {
	Current string `json:"current"`
	New     string `json:"new"`
}

func (s *Server) handleUpdatePassword(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.Add("method", "handleUpdatePassword")

	payload := &newPasswordPayload{}
	if !s.parseNewPasswordPayload(logger, w, r, payload) ||
		!s.correctPassword(logger, w, payload.Current) {
		return
	}

	tx, ok := s.begin(logger, w)
	if !ok {
		return
	}

	if !s.deleteTx(logger, w, &data.Setting{Key: saltKey}, tx) ||
		!s.deleteTx(logger, w, &data.Setting{Key: passwordKey}, tx) {
		return
	}

	s.setPassword(w, payload.New, tx)
}

func (s *Server) parseNewPasswordPayload(logger log.Logger,
	w http.ResponseWriter, r *http.Request, payload *newPasswordPayload) bool {
	return s.parsePayload(logger, w, r, payload) &&
		s.validPasswordString(logger, w, payload.New)
}

func (s *Server) validPasswordString(logger log.Logger,
	w http.ResponseWriter, password string) bool {
	if len(password) < passwordMinLen || len(password) > passwordMaxLen {
		msg := fmt.Sprintf(
			"password must be at least %d and at most %d long",
			passwordMinLen, passwordMaxLen)
		s.replyErr(logger,
			w, http.StatusBadRequest, &serverError{Message: msg})
		return false
	}
	return true
}

func (s *Server) correctPassword(logger log.Logger, w http.ResponseWriter, pwd string) bool {
	password := s.findPasswordSetting(w, passwordKey)
	salt := s.findPasswordSetting(w, saltKey)
	if password == nil || salt == nil {
		return false
	}

	if data.ValidatePassword(password.Value, pwd, salt.Value) != nil {
		s.replyErr(logger, w, http.StatusUnauthorized, &serverError{
			Message: "Wrong password",
		})
		return false
	}

	return true
}

func (s *Server) findPasswordSetting(w http.ResponseWriter, key string) *data.Setting {
	logger := s.logger.Add("method", "findPasswordSetting")

	rec := &data.Setting{}
	if err := s.db.FindByPrimaryKeyTo(rec, key); err != nil {
		logger.Warn(fmt.Sprintf("failed to retrieve %s: %v", key, err))
		s.replyErr(logger, w, http.StatusUnauthorized, &serverError{
			Code:    1,
			Message: "Wrong password",
		})
		return nil
	}
	return rec
}

func (s *Server) setPasswordAllowed(w http.ResponseWriter) bool {
	logger := s.logger.Add("method", "setPasswordAllowed")

	if _, err := s.db.FindByPrimaryKeyFrom(data.SettingTable, passwordKey); err != sql.ErrNoRows {
		s.replyErr(logger, w, http.StatusUnauthorized, &serverError{
			Code:    0,
			Message: "Password exists, access denied",
		})
		return false
	}

	accounts, err := s.db.SelectAllFrom(data.AccountTable, "")
	if err != nil {
		logger.Error(fmt.Sprintf("failed to select account: %v", err))
		s.replyUnexpectedErr(logger, w)
		return false
	}
	if len(accounts) > 0 {
		s.replyErr(logger, w, http.StatusUnauthorized, &serverError{
			Code:    1,
			Message: "No password exists, while some accounts found in the system. Please, reinstall the application",
		})
		return false
	}

	return true
}

func (s *Server) setPassword(w http.ResponseWriter, password string, tx *reform.TX) bool {
	logger := s.logger.Add("method", "setPassword")

	salt := util.NewUUID()
	passwordSetting := &data.Setting{Key: saltKey, Value: salt,
		Permissions: data.AccessDenied, Name: "Password"}
	if !s.insertTx(logger, w, passwordSetting, tx) {
		return false
	}

	hashed, err := data.HashPassword(password, salt)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to hash password: %v", err))
		s.replyUnexpectedErr(logger, w)
		return false
	}

	saltSetting := &data.Setting{Key: passwordKey, Value: string(hashed),
		Permissions: data.AccessDenied, Name: "Salt"}
	if !s.insertTx(logger, w, saltSetting, tx) || !s.commit(logger, w, tx) {
		return false
	}

	return true
}
