package uisrv

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
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
	Paginated    bool
}

// paginatedReply is a format of paginated reply.
type paginatedReply struct {
	Items   []reform.Struct `json:"items"`
	Current int             `json:"current"`
	Pages   int             `json:"pages"`
}

func (s *Server) newPaginatedReply(conf *getConf, tail string,
	args []interface{}, page, perPage int) (*paginatedReply, error) {
	ret := &paginatedReply{
		Current: page,
	}

	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`,
		conf.View.Name(), tail)
	row := s.db.QueryRow(query, args...)
	var count int
	if err := row.Scan(&count); err != nil {
		return nil, err
	}

	ret.Pages = int(math.Ceil(
		float64(count) / float64(perPage)))

	return ret, nil
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

	var paginatedItems *paginatedReply

	if conf.Paginated {
		page, err := strconv.Atoi(r.FormValue("page"))
		if err != nil || page == 0 {
			s.logger.Warn("invalid param: %v", err)
			s.replyInvalidRequest(w)
			return
		}

		limit, err := strconv.Atoi(r.FormValue("perPage"))
		if err != nil || limit == 0 {
			s.logger.Warn("invalid param: %v", err)
			s.replyInvalidRequest(w)
			return
		}

		paginatedItems, err = s.newPaginatedReply(
			conf, tail, args, page, limit)
		if err != nil {
			s.logger.Error("failed to get resources: %v", err)
			s.replyUnexpectedErr(w)
			return
		}

		tail += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, (page-1)*limit)
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

	if paginatedItems != nil {
		paginatedItems.Items = records
		s.reply(w, paginatedItems)
		return
	}

	s.reply(w, records)
}
