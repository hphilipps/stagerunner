package main

import (
	"context"
	"net/http"
	"time"

	"github.com/hphilipps/stagerunner/domain"
	myhttp "github.com/hphilipps/stagerunner/http"
	"github.com/hphilipps/stagerunner/store"
	"github.com/urfave/cli/v2"
)

// serverCommand is the cli command for starting the API server
var serverCommand = &cli.Command{
	Name:  "server",
	Usage: "Start the API server",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "addr",
			Value:   defaultServerAddr,
			Usage:   "Server address to listen on",
			EnvVars: []string{"STAGERUNNER_ADDR"},
		},
		&cli.IntFlag{
			Name:    "workers",
			Value:   2,
			Usage:   "Number of workers to run pipeline runs",
			EnvVars: []string{"STAGERUNNER_WORKERS"},
		},
		&cli.IntFlag{
			Name:    "queue-size",
			Aliases: []string{"qs"},
			Value:   10,
			Usage:   "Size of the pipeline run queue",
			EnvVars: []string{"STAGERUNNER_QUEUE_SIZE"},
		},
		&cli.IntFlag{
			Name:    "per-pipeline-queue",
			Aliases: []string{"pp"},
			Value:   3,
			Usage:   "Maximum number of queued runs per pipeline",
			EnvVars: []string{"STAGERUNNER_PER_PIPELINE_QUEUE"},
		},
		&cli.IntFlag{
			Name:    "executor-delay",
			Aliases: []string{"delay"},
			Value:   5,
			Usage:   "Delay in seconds between pipeline run stages",
			EnvVars: []string{"STAGERUNNER_EXECUTOR_DELAY"},
		},
		&cli.Float64Flag{
			Name:    "fail-probability",
			Aliases: []string{"fp"},
			Value:   0.0,
			Usage:   "Probability of a pipeline run stage failing",
			EnvVars: []string{"STAGERUNNER_FAIL_PROBABILITY"},
		},
	},
	Action: runServer,
}

func runServer(c *cli.Context) error {
	store := store.NewMemoryStore()
	executor := domain.NewExecutor(
		store,
		c.Int("workers"),
		c.Int("queue-size"),
		c.Int("per-pipeline-queue"),
		c.Float64("fail-probability"),
		time.Duration(c.Int("executor-delay"))*time.Second,
	)
	api := myhttp.NewAPI(store, executor)
	router := api.SetupRouter()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start workers and process pipeline runs
	go executor.Start(ctx)
	return http.ListenAndServe(c.String("addr"), router)
}
