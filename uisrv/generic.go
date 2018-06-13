package uisrv

import (
	"fmt"
	"net/http"
	"strings"

	reform "gopkg.in/reform.v1"
)

// queryParam is a description of a query param.
type queryParam struct {
	Name  string // in request.
	Field string // column name in db.
	Op    string // comparison operator ({Field} {Op} {Value}, default "=")
}

// getConf is a config for generic get handler.
type getConf struct {
	Params       []queryParam
	View         reform.View
	FilteringSQL string
}

func (s *Server) formatConditions(r *http.Request, conf *getConf) (conds []string, args []interface{}) {
	placei := 1

	for _, param := range conf.Params {
		op := "="
		if param.Op != "" {
			op = param.Op
		}

		val := r.FormValue(param.Name)
		if val == "" {
			continue
		}

		var ph string
		if op == "in" {
			subvals := strings.Split(val, ",")
			for _, subval := range subvals {
				args = append(args, subval)
			}

			phs := s.db.Placeholders(placei, len(subvals))
			placei += len(subvals)
			ph = "(" + strings.Join(phs, ",") + ")"
		} else {
			args = append(args, val)
			ph = s.db.Placeholder(placei)
			placei++
		}

		cond := fmt.Sprintf("%s %s %s", param.Field, op, ph)
		conds = append(conds, cond)
	}

	return conds, args

}

// handleGetResources select and returns records.
func (s *Server) handleGetResources(w http.ResponseWriter,
	r *http.Request, conf *getConf) {

	conds, args := s.formatConditions(r, conf)
	if conf.FilteringSQL != "" {
		conds = append(conds, conf.FilteringSQL)
	}

	var tail string
	if len(conds) > 0 {
		tail = "WHERE " + strings.Join(conds, " AND ")
	}

	records, err := s.db.SelectAllFrom(conf.View, tail, args...)
	if err != nil {
		s.logger.Warn("failed to select: %v", err)
		s.replyUnexpectedErr(w)
		return
	}

	if records == nil {
		records = []reform.Struct{}
	}

	s.reply(w, records)
}
