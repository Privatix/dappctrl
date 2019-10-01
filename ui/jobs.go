package ui

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/privatix/dappctrl/data"
)

// GetJobsResult is a jobs list after applying all filtering and paginations.
type GetJobsResult struct {
	Items      []data.Job `json:"items"`
	TotalItems int
}

// GetJobs returns list of jobs for given params.
func (h *Handler) GetJobs(tkn, jtype, dfrom, dto string, statuses []string, offset, limit uint) (*GetJobsResult, error) {
	logger := h.logger.Add("method", "GetJobs")

	if !h.token.Check(tkn) {
		logger.Warn("access denied")
		return nil, ErrAccessDenied
	}
	var conditions []string
	var args []interface{}
	if len(statuses) != 0 {
		for _, status := range statuses {
			args = append(args, status)
		}
		conditions = append(conditions, fmt.Sprintf("status in (%s)",
			strings.Join(h.db.Placeholders(1, len(statuses)), ",")))
	}
	addFilter := func(cond string, arg interface{}) {
		conditions = append(conditions, fmt.Sprintf("%s%s", cond, h.db.Placeholder(len(args)+1)))
		args = append(args, arg)
	}
	if jtype != "" {
		addFilter("type=", jtype)
	}
	if dfrom != "" {
		addFilter("created_at>=", dfrom)
	}
	if dto != "" {
		addFilter("created_at<", dto)
	}
	var qtail string
	if len(conditions) > 0 {
		qtail = "WHERE " + strings.Join(conditions, " AND ")
	}
	var count int
	err := h.db.QueryRow(`
		SELECT COUNT(*)
		  FROM jobs `+qtail, args...).Scan(&count)
	if err != nil {
		logger.Error(fmt.Sprintf("could not query total number of jobs: %v", err))
		return nil, ErrInternal
	}
	if offset != 0 {
		qtail += " OFFSET " + h.db.Placeholder(len(args)+1)
		args = append(args, offset)
	}
	if limit != 0 {
		qtail += " LIMIT " + h.db.Placeholder(len(args)+1)
		args = append(args, limit)
	}
	recs, err := h.db.SelectAllFrom(data.JobTable, qtail, args...)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	jobs := make([]data.Job, 0)
	for _, r := range recs {
		jobs = append(jobs, *r.(*data.Job))
	}
	return &GetJobsResult{Items: jobs, TotalItems: count}, nil
}

// ReactivateJob resets job to be run again.
func (h *Handler) ReactivateJob(tkn, id string) error {
	logger := h.logger.Add("method", "ReactivateJob", "id", id)
	if !h.token.Check(tkn) {
		logger.Warn("access denied")
		return ErrAccessDenied
	}
	if id == "" {
		logger.Warn("empty id param")
		return ErrJobNotFound
	}
	var j data.Job
	if err := h.db.FindByPrimaryKeyTo(&j, id); err != nil {
		if err == sql.ErrNoRows {
			logger.Warn("job not found")
			return ErrJobNotFound
		}
		logger.Error(err.Error())
		return ErrInternal
	}
	if j.Status == data.JobDone {
		logger.Warn("job is in done status, cannot be reactivated")
		return ErrSuccessJobNonReactivatable
	}
	if j.Status == data.JobActive {
		logger.Warn("attempt to reactivate active job")
		return ErrAlreadyActiveJob
	}
	j.Status = data.JobActive
	j.TryCount = 0
	if err := h.db.Save(&j); err != nil {
		logger.Warn(fmt.Sprintf("could not save job: %v", err))
		return ErrInternal
	}
	return nil
}
