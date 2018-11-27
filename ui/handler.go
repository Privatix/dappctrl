package ui

import (
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/util/log"
)

// Handler is an UI RPC handler.
type Handler struct {
	logger         log.Logger
	db             *reform.DB
	queue          job.Queue
	pwdStorage     data.PWDGetSetter
	encryptKeyFunc data.EncryptedKeyFunc
	decryptKeyFunc data.ToPrivateKeyFunc
	userRole       string
	processor      *proc.Processor
}

// NewHandler creates a new handler.
func NewHandler(logger log.Logger, db *reform.DB,
	queue job.Queue, pwdStorage data.PWDGetSetter,
	encryptKeyFunc data.EncryptedKeyFunc,
	decryptKeyFunc data.ToPrivateKeyFunc, userRole string,
	processor *proc.Processor) *Handler {
	logger = logger.Add("type", "uisrv.Handler")
	return &Handler{
		logger:         logger,
		db:             db,
		queue:          queue,
		pwdStorage:     pwdStorage,
		encryptKeyFunc: encryptKeyFunc,
		decryptKeyFunc: decryptKeyFunc,
		userRole:       userRole,
		processor:      processor,
	}
}
