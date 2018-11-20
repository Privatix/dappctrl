package sesssrv

import (
	"net/http"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/log"
	"github.com/privatix/dappctrl/util/srv"
)

func (s *Server) authProduct(username, password string) bool {
	var prod data.Product
	if s.db.FindByPrimaryKeyTo(&prod, username) != nil ||
		data.ValidatePassword(prod.Password,
			password, string(prod.Salt)) != nil {
		return false
	}
	return true
}

func (s *Server) findProduct(logger log.Logger,
	w http.ResponseWriter, productID string) (*data.Product, bool) {
	var prod data.Product
	if err := s.db.FindByPrimaryKeyTo(&prod, productID); err != nil {
		logger.Add("productId", productID).Error(err.Error())
		s.RespondError(logger, w, srv.ErrInternalServerError)
		return nil, false
	}
	return &prod, true
}

func (s *Server) findChannel(logger log.Logger,
	w http.ResponseWriter, channelID string) (*data.Channel, bool) {
	var ch data.Channel
	if err := s.db.FindByPrimaryKeyTo(&ch, channelID); err != nil {
		logger.Add("channelId", channelID).Error(err.Error())
		s.RespondError(logger, w, ErrChannelNotFound)
		return nil, false
	}
	return &ch, true
}

func (s *Server) findEndpoint(logger log.Logger,
	w http.ResponseWriter, channelID string) (*data.Endpoint, bool) {
	var ept data.Endpoint
	if err := data.FindOneTo(s.db.Querier, &ept, "channel",
		channelID); err != nil {
		logger.Error(err.Error())
		s.RespondError(logger, w, ErrEndpointNotFound)
		return nil, false
	}
	return &ept, true
}

func (s *Server) updateProduct(logger log.Logger,
	w http.ResponseWriter, prod *data.Product) bool {
	if err := s.db.Update(prod); err != nil {
		logger.Error(err.Error())
		s.RespondError(logger, w, srv.ErrInternalServerError)
		return false
	}
	return true
}

func (s *Server) identClient(logger log.Logger,
	w http.ResponseWriter, productID, clientID string) (*data.Channel, bool) {
	prod, ok := s.findProduct(logger, w, productID)
	if !ok {
		return nil, false
	}

	var ch *data.Channel
	if prod.ClientIdent == data.ClientIdentByChannelID {
		ch, ok = s.findChannel(logger, w, clientID)
		if !ok {
			return nil, false
		}
	} else {
		logger.Fatal("unsupported client identification type")
	}

	return ch, true
}

func (s *Server) findCurrentSession(
	logger log.Logger, w http.ResponseWriter,
	channel string, noErrorIfNotFound bool) (*data.Session, bool) {
	var sess data.Session
	if err := s.db.SelectOneTo(&sess, `
		WHERE channel = $1 AND stopped IS NULL
		ORDER BY started DESC
		LIMIT 1`, channel); err != nil {
		logger.Warn(err.Error())
		if noErrorIfNotFound {
			s.RespondResult(logger, w, nil)
		} else {
			s.RespondError(logger, w, ErrSessionNotFound)
		}
		return nil, false
	}

	return &sess, true
}
