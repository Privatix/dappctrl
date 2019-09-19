package main

import (
	"flag"
	"fmt"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/util"
)

func main() {
	fconfig := flag.String("config", "dappctrl.config.json", "dappctrl configuration file")

	flag.Parse()

	conf := struct {
		DB *data.DBConfig
	}{
		DB: data.NewDBConfig(),
	}
	if err := util.ReadJSONFile(*fconfig, &conf); err != nil {
		panic(fmt.Sprintf("failed to read configuration: %s", err))
	}

	db, err := data.NewDB(conf.DB)
	if err != nil {
		panic(fmt.Sprintf("failod to make db client: %v", err))
	}

	query := `UPDATE jobs
	            SET try_count=0, status=$1
			    WHERE type=$2 AND status=$3`
	if _, err := db.Exec(query, data.JobActive,
		data.JobClientPreUncooperativeClose, data.JobFailed); err != nil {
		panic(fmt.Sprintf("failed to re-activate uncoop close jobs: %v", err))
	}

	jobConf := struct {
		Job *job.Config
	}{
		Job: job.NewConfig(),
	}
	if err := util.ReadJSONFile(*fconfig, &jobConf); err != nil {
		panic(fmt.Sprintf("failed to read configuration: %s", err))
	}
	tmp := jobConf.Job.Types[data.JobClientPreUncooperativeClose]
	tmp.FirstStartDelay = 750000000
	jobConf.Job.Types[data.JobClientPreUncooperativeClose] = tmp
	allConf := make(map[string]interface{})
	if err := util.ReadJSONFile(*fconfig, &allConf); err != nil {
		panic(fmt.Sprintf("failed to read configuration: %s", err))
	}
	allConf["Job"] = jobConf.Job
	if err := util.WriteJSONFile(*fconfig, "", "	", allConf); err != nil {
		panic(fmt.Sprintf("failed to write file: %v", err))
	}
}
