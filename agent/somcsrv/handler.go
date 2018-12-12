package somcsrv

import (
	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/util/log"
)

// Handler is agents RPC handler.
type Handler struct {
	db     *reform.DB
	logger log.Logger
}

// NewHandler creates a new RPC handler.
func NewHandler(db *reform.DB, logger log.Logger) *Handler {
	return &Handler{db: db, logger: logger.Add("type", "tor-somc.Handler")}
}
