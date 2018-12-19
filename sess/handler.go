package sess

import (
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/country"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/util/log"
)

// Handler is a session RPC handler.
type Handler struct {
	countryConf *country.Config
	db          *reform.DB
	logger      log.Logger
	queue       job.Queue
}

// NewHandler creates a new session handler.
func NewHandler(logger log.Logger, db *reform.DB,
	countryConf *country.Config, queue job.Queue) *Handler {
	logger = logger.Add("type", "sess.Handler")
	return &Handler{
		db:          db,
		logger:      logger,
		countryConf: countryConf,
		queue:       queue,
	}
}
