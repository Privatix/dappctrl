package uisrv

import (
	"encoding/json"
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

// handleTempaltes calls appropriate handler by scanning incoming request.
func (s *Server) handleTempaltes(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.handleTemplateCreate(w, r)
		return
	}
	if r.Method == http.MethodGet {
		s.handleGetTemplates(w, r)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleTemplateCreate creates new template.
func (s *Server) handleTemplateCreate(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.Add("method", "handleTemplateCreate")

	tpl := &data.Template{}
	if !s.parseTemplatePayload(logger, w, r, tpl) {
		return
	}
	tpl.ID = util.NewUUID()
	tpl.Hash = data.HexFromBytes(crypto.Keccak256(tpl.Raw))
	if !s.insert(logger, w, tpl) {
		return
	}
	s.replyEntityCreated(logger, w, tpl.ID)
}

func (s *Server) parseTemplatePayload(logger log.Logger, w http.ResponseWriter,
	r *http.Request, tpl *data.Template) bool {

	v := make(map[string]interface{})
	if !s.parsePayload(logger, w, r, tpl) ||
		invalidTemplateKind(tpl.Kind) ||
		json.Unmarshal(tpl.Raw, &v) != nil {
		s.replyInvalidRequest(logger, w)
		return false
	}
	return true
}

func invalidTemplateKind(v string) bool {
	return v != data.TemplateOffer && v != data.TemplateAccess
}

// handleGetTemplates replies with all templates or template by id and/or type.
func (s *Server) handleGetTemplates(w http.ResponseWriter, r *http.Request) {
	s.handleGetResources(w, r, &getConf{
		Params: []queryParam{
			{Name: "type", Field: "kind"},
			{Name: "id", Field: "id"},
		},
		View: data.TemplateTable,
	})
}
