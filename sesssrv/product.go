package sesssrv

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/privatix/dappctrl/util/srv"
)

// ProductArgs is a set of product arguments.
type ProductArgs struct {
	Config map[string]string `json:"config"`
}

func (s *Server) handleProductConfig(
	w http.ResponseWriter, r *http.Request, ctx *srv.Context) {
	var args ProductArgs
	if !s.ParseRequest(w, r, &args) {
		return
	}

	if len(args.Config) == 0 {
		s.RespondError(w, ErrProductConfNotValid)
		return
	}

	prod, ok := s.findProduct(w, ctx.Username)
	if !ok {
		return
	}

	if !prodConfigEmpty(prod.Config) {
		s.RespondError(w, ErrProductConfAlreadyUploaded)
		return
	}

	prodConf, err := json.Marshal(args.Config)
	if err != nil {
		s.RespondError(w, srv.ErrInternalServerError)
		return
	}

	prod.Config = prodConf

	if ok := s.updateProduct(w, prod); !ok {
		return

	}

	s.RespondResult(w, nil)
}

func prodConfigEmpty(data []byte) bool {
	return reflect.DeepEqual([]byte(`{}`), data)
}
