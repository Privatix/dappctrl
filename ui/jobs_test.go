package ui_test

import (
	"testing"
	"time"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/util"
)

func TestGetJobs(t *testing.T) {
	// Token validation.
	if _, err := handler.GetJobs("wrong-token", "", "", "", nil, 0, 0); err != ui.ErrAccessDenied {
		t.Fatalf("wanted: %v, got: %v", ui.ErrAccessDenied, err)
	}
	job := data.NewTestJob(data.JobAccountUpdateBalances, data.JobUser, data.JobOffering)
	job.RelatedID = util.NewUUID()
	job.Status = data.JobActive
	data.InsertToTestDB(t, db, job)
	defer data.DeleteFromTestDB(t, db, job)
	// Filter by status.
	if ret, err := handler.GetJobs(testToken.v, "", "", "", []string{data.JobCanceled}, 0, 0); err != nil {
		t.Fatal(err)
	} else if len(ret.Items) != 0 {
		t.Fatalf("filtering by status, wanted 0 jobs, got: %d", len(ret.Items))
	}
	if ret, err := handler.GetJobs(testToken.v, "", "", "", []string{job.Status, data.JobCanceled}, 0, 0); err != nil {
		t.Fatal(err)
	} else if len(ret.Items) != 1 {
		t.Fatalf("filtering by status, wanted 1 job, got %d", len(ret.Items))
	}
	// Filter by job type.
	if ret, err := handler.GetJobs(testToken.v, data.JobAgentAfterChannelCreate, "", "", nil, 0, 0); err != nil {
		t.Fatal(err)
	} else if len(ret.Items) != 0 {
		t.Fatalf("filtering by type, wanted 0 jobs, got: %d", len(ret.Items))
	}
	if ret, err := handler.GetJobs(testToken.v, job.Type, "", "", nil, 0, 0); err != nil {
		t.Fatal(err)
	} else if len(ret.Items) != 1 {
		t.Fatalf("filtering by type, wanted 1 job, got %d", len(ret.Items))
	}
	// Filter by date range.
	if ret, err := handler.GetJobs(testToken.v, "", dateArg(time.Now().Add(time.Minute)),
		dateArg(time.Now().Add(2*time.Minute)), nil, 0, 0); err != nil {
		t.Fatal(err)
	} else if len(ret.Items) != 0 {
		t.Fatalf("filtering by date range, wanted 0 jobs, got: %d", len(ret.Items))
	}
	if ret, err := handler.GetJobs(testToken.v, "", dateArg(time.Now().Add(-time.Minute)),
		dateArg(time.Now().Add(time.Minute)), nil, 0, 0); err != nil {
		t.Fatal(err)
	} else if len(ret.Items) != 1 {
		t.Fatalf("filtering by date range, wanted 1 job, got %d", len(ret.Items))
	}
	if ret, err := handler.GetJobs(testToken.v, "", dateArg(time.Now().Add(-time.Minute)),
		dateArg(time.Now().Add(-2*time.Minute)), nil, 0, 0); err != nil {
		t.Fatal(err)
	} else if len(ret.Items) != 0 {
		t.Fatalf("filtering by date range, wanted 0 jobs, got %d", len(ret.Items))
	}
	if ret, err := handler.GetJobs(testToken.v, "", "",
		dateArg(time.Now().Add(-time.Minute)), nil, 0, 0); err != nil {
		t.Fatal(err)
	} else if len(ret.Items) != 0 {
		t.Fatalf("filtering by date range, wanted 0 jobs, got %d", len(ret.Items))
	}
	// Pagination, accept offset, limit. Return items and total number of items.
	job2 := *job
	job2.ID = util.NewUUID()
	data.InsertToTestDB(t, db, &job2)
	defer data.DeleteFromTestDB(t, db, &job2)
	if ret, err := handler.GetJobs(testToken.v, "", "", "", nil, 0, 0); err != nil {
		t.Fatal(err)
	} else if ret.TotalItems != 2 {
		t.Fatalf("wanted total items 2, got: %d", ret.TotalItems)
	}
	if ret, err := handler.GetJobs(testToken.v, "", "", "", nil, 2, 0); err != nil {
		t.Fatal(err)
	} else if ret.TotalItems != 2 {
		t.Fatalf("wanted total items 2, got: %d", ret.TotalItems)
	} else if len(ret.Items) != 0 {
		t.Fatalf("wanted 0 items, got: %d", len(ret.Items))
	}
	if ret, err := handler.GetJobs(testToken.v, "", "", "", nil, 0, 1); err != nil {
		t.Fatal(err)
	} else if ret.TotalItems != 2 {
		t.Fatalf("wanted total items 2, got: %d", ret.TotalItems)
	} else if len(ret.Items) != 1 {
		t.Fatalf("wanted 1 items, got: %d", len(ret.Items))
	}
}

func TestReactivateJob(t *testing.T) {
	// Token validation.
	if err := handler.ReactivateJob("wrong-token", ""); err != ui.ErrAccessDenied {
		t.Fatalf("wanted: %v, got: %v", ui.ErrAccessDenied, err)
	}
	// Doesn't exist.
	if err := handler.ReactivateJob(testToken.v, ""); err != ui.ErrJobNotFound {
		t.Fatalf("wanted: %v, got: %v", ui.ErrJobNotFound, err)
	}
	if err := handler.ReactivateJob(testToken.v, util.NewUUID()); err != ui.ErrJobNotFound {
		t.Fatalf("wanted: %v, got: %v", ui.ErrJobNotFound, err)
	}
	// Can't reactivate successful job.
	job := data.NewTestJob(data.JobAccountUpdateBalances, data.JobUser, data.JobOffering)
	job.RelatedID = util.NewUUID()
	job.Status = data.JobDone
	data.InsertToTestDB(t, db, job)
	data.ReloadFromTestDB(t, db, job)
	defer data.DeleteFromTestDB(t, db, job)
	if err := handler.ReactivateJob(testToken.v, job.ID); err != ui.ErrSuccessJobNonReactivatable {
		t.Fatalf("wanted: %v, got: %v", ui.ErrSuccessJobNonReactivatable, err)
	}
	// Can't reactivate if already active.
	job.Status = data.JobActive
	data.SaveToTestDB(t, db, job)
	if err := handler.ReactivateJob(testToken.v, job.ID); err != ui.ErrAlreadyActiveJob {
		t.Fatalf("wanted: %v, got: %v", ui.ErrAlreadyActiveJob, err)
	}
	// Reactivate job.
	job.Status = data.JobFailed
	job.TryCount = 10
	data.SaveToTestDB(t, db, job)
	if err := handler.ReactivateJob(testToken.v, job.ID); err != nil {
		t.Fatal(err)
	}
	data.ReloadFromTestDB(t, db, job)
	if job.Status != data.JobActive {
		t.Fatalf("wanted job status: %v, got: %v", data.JobActive, job.Status)
	}
	if job.TryCount != 0 {
		t.Fatalf("wanted try count 0, got: %d", job.TryCount)
	}
}
