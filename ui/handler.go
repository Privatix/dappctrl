package ui

import (
	"context"
	"math/big"

	"github.com/privatix/dappctrl/client/somc"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/util/log"
)

// Suggestor suggests best gas price for this moment.
type Suggestor interface {
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
}

// Handler is an UI RPC handler.
type Handler struct {
	logger            log.Logger
	db                *reform.DB
	queue             job.Queue
	pwdStorage        data.PWDGetSetter
	encryptKeyFunc    data.EncryptedKeyFunc
	userRole          string
	processor         *proc.Processor
	somcClientBuilder somc.ClientBuilderInterface
	token             TokenMakeChecker
	suggestor         Suggestor
}

// NewHandler creates a new handler.
func NewHandler(logger log.Logger, db *reform.DB,
	queue job.Queue, pwdStorage data.PWDGetSetter,
	encryptKeyFunc data.EncryptedKeyFunc, userRole string,
	processor *proc.Processor,
	somcClientBuilder somc.ClientBuilderInterface,
	token TokenMakeChecker, suggestor Suggestor) *Handler {
	logger = logger.Add("type", "ui.Handler")
	return &Handler{
		logger:            logger,
		db:                db,
		queue:             queue,
		pwdStorage:        pwdStorage,
		encryptKeyFunc:    encryptKeyFunc,
		userRole:          userRole,
		processor:         processor,
		somcClientBuilder: somcClientBuilder,
		token:             token,
		suggestor:         suggestor,
	}
}
