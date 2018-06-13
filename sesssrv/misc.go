package sesssrv

import (
	"net/http"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/srv"
)

func (s *Server) authProduct(username, password string) bool {
	var prod data.Product
	if s.db.FindByPrimaryKeyTo(&prod, username) != nil ||
		data.ValidatePassword(prod.Password, password, string(prod.Salt)) != nil {
		return false
	}
	return true
}

func (s *Server) findProduct(
	w http.ResponseWriter, productID string) (*data.Product, bool) {
	var prod data.Product
	if err := s.db.FindByPrimaryKeyTo(&prod, productID); err != nil {
		s.Logger().Error("failed to find product: %s", err)
		s.RespondError(w, srv.ErrInternalServerError)
		return nil, false
	}
	return &prod, true
}

func (s *Server) updateProduct(
	w http.ResponseWriter, prod *data.Product) bool {
	if err := s.db.Update(prod); err != nil {
		s.Logger().Error("failed to update product: %v", err)
		s.RespondError(w, srv.ErrInternalServerError)
		return false
	}
	return true
}

func (s *Server) identClient(w http.ResponseWriter,
	productID, clientID string) (*data.Channel, bool) {
	prod, ok := s.findProduct(w, productID)
	if !ok {
		return nil, false
	}

	var ch data.Channel
	if prod.ClientIdent == data.ClientIdentByChannelID {
		if err := s.db.FindByPrimaryKeyTo(&ch, clientID); err != nil {
			s.Logger().Warn("failed to find channel: %s", err)
			s.RespondError(w, ErrChannelNotFound)
			return nil, false
		}
	} else {
		panic("unsupported client identification type: " +
			prod.ClientIdent)
	}

	if ch.ServiceStatus != data.ChannelActive {
		s.Logger().Warn("non-active channel: %s", ch.ID)
		s.RespondError(w, ErrNonActiveChannel)
		return nil, false
	}

	return &ch, true
}

func (s *Server) findCurrentSession(
	w http.ResponseWriter, channel string) (*data.Session, bool) {
	var sess data.Session
	if err := s.db.SelectOneTo(&sess, `
		WHERE channel = $1 AND stopped IS NULL
		ORDER BY started DESC
		LIMIT 1`, channel); err != nil {
		msg := "failed to find current session for channel %s: %s"
		s.Logger().Warn(msg, channel, err)
		s.RespondError(w, ErrSessionNotFound)
		return nil, false
	}

	return &sess, true
}
