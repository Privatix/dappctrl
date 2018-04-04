package uisrv

import (
	"database/sql"
	"fmt"
	"net/http"

	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
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
	return func(w http.ResponseWriter, r *http.Request) {
		_, givenPassword, ok := r.BasicAuth()
		if !ok {
			s.replyErr(w, http.StatusUnauthorized, &serverError{
				Message: "Wrong password",
			})
			return
		}

		if !s.correctPassword(w, givenPassword) {
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
	payload := &passwordPayload{}
	if !s.parsePasswordPayload(w, r, payload) || !s.setPasswordAllowed(w) {
		return
	}

	tx, ok := s.begin(w)
	if !ok {
		return
	}

	if !s.setPassword(w, payload.Password, tx) {
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) parsePasswordPayload(w http.ResponseWriter,
	r *http.Request, payload *passwordPayload) bool {
	return s.parsePayload(w, r, payload) && s.validPasswordString(w, payload.Password)
}

type newPasswordPayload struct {
	Current string `json:"current"`
	New     string `json:"new"`
}

func (s *Server) handleUpdatePassword(w http.ResponseWriter, r *http.Request) {
	payload := &newPasswordPayload{}
	if !s.parseNewPasswordPayload(w, r, payload) || !s.correctPassword(w, payload.Current) {
		return
	}

	tx, ok := s.begin(w)
	if !ok {
		return
	}

	if !s.deleteTx(w, &data.Setting{Key: saltKey}, tx) ||
		!s.deleteTx(w, &data.Setting{Key: passwordKey}, tx) {
		return
	}

	s.setPassword(w, payload.New, tx)
}

func (s *Server) parseNewPasswordPayload(w http.ResponseWriter,
	r *http.Request, payload *newPasswordPayload) bool {
	return s.parsePayload(w, r, payload) && s.validPasswordString(w, payload.New)
}

func (s *Server) validPasswordString(w http.ResponseWriter, password string) bool {
	if len(password) < passwordMinLen || len(password) > passwordMaxLen {
		msg := fmt.Sprintf(
			"password must be at least %d and at most %d long",
			passwordMinLen, passwordMaxLen)
		s.replyErr(w, http.StatusBadRequest, &serverError{Message: msg})
		return false
	}
	return true
}

func (s *Server) correctPassword(w http.ResponseWriter, pwd string) bool {
	password := s.findPasswordSetting(w, passwordKey)
	salt := s.findPasswordSetting(w, saltKey)
	if password == nil || salt == nil {
		return false
	}

	if data.ValidatePassword(password.Value, pwd, salt.Value) != nil {
		s.replyErr(w, http.StatusUnauthorized, &serverError{
			Message: "Wrong password",
		})
		return false
	}

	return true
}

func (s *Server) findPasswordSetting(w http.ResponseWriter, key string) *data.Setting {
	rec := &data.Setting{}
	if err := s.db.FindByPrimaryKeyTo(rec, key); err != nil {
		s.logger.Warn("failed to retrieve %s: %v", key, err)
		s.replyErr(w, http.StatusUnauthorized, &serverError{
			Code:    1,
			Message: "Wrong password",
		})
		return nil
	}
	return rec
}

func (s *Server) setPasswordAllowed(w http.ResponseWriter) bool {
	if _, err := s.db.FindByPrimaryKeyFrom(data.SettingTable, passwordKey); err != sql.ErrNoRows {
		s.replyErr(w, http.StatusUnauthorized, &serverError{
			Code:    0,
			Message: "Password exists, access denied",
		})
		return false
	}

	accounts, err := s.db.SelectAllFrom(data.AccountTable, "")
	if err != nil {
		s.logger.Error("failed to select account: %v", err)
		s.replyUnexpectedErr(w)
		return false
	}
	if len(accounts) > 0 {
		s.replyErr(w, http.StatusUnauthorized, &serverError{
			Code:    1,
			Message: "No password exists, while some accounts found in the system. Please, reinstall the application",
		})
		return false
	}

	return true
}

func (s *Server) setPassword(w http.ResponseWriter, password string, tx *reform.TX) bool {
	salt := util.NewUUID()
	passwordSetting := &data.Setting{Key: saltKey, Value: salt, Name: "Password"}
	if !s.insertTx(w, passwordSetting, tx) {
		return false
	}

	hashed, err := data.HashPassword(password, salt)
	if err != nil {
		s.logger.Error("failed to hash password: %v", err)
		s.replyUnexpectedErr(w)
		return false
	}

	saltSetting := &data.Setting{Key: passwordKey, Value: string(hashed), Name: "Salt"}
	if !s.insertTx(w, saltSetting, tx) || !s.commit(w, tx) {
		return false
	}

	return true
}
