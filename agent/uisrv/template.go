package uisrv

import (
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

// handleTempaltes calls appropriate handler by scanning incoming request.
func (s *Server) handleTempaltes(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		s.handleTemplateCreate(w, r)
		return
	}
	if r.Method == "GET" {
		s.handleGetTemplates(w, r)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleTemplateCreate creates new template.
func (s *Server) handleTemplateCreate(w http.ResponseWriter, r *http.Request) {
	tpl := &data.Template{}
	if !s.parseTemplatePayload(w, r, tpl) {
		return
	}
	tpl.ID = util.NewUUID()
	tpl.Hash = data.FromBytes(crypto.Keccak256(tpl.Raw))
	if err := s.db.Insert(tpl); err != nil {
		s.logger.Warn("failed to insert template: %v", err)
		s.replyUnexpectedErr(w)
		return
	}
	s.replyEntityCreated(w, tpl.ID)
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
