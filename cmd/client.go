package main

import (
	"context"
	"encoding/json"
	"fmt"

	myhttp "github.com/hphilipps/stagerunner/http"
	"github.com/urfave/cli/v2"
)

// clientCommand is the cli command for running client requests against the API
var clientCommand = &cli.Command{
	Name:  "client",
	Usage: "Run client commands against the API",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "url",
			Value:   defaultAPIURL,
			Usage:   "API server URL",
			EnvVars: []string{"STAGERUNNER_API_URL"},
		},
		&cli.StringFlag{
			Name:    "token",
			Usage:   "Authorization token",
			EnvVars: []string{"STAGERUNNER_TOKEN"},
		},
	},
	Subcommands: []*cli.Command{
		{
			Name:   "list",
			Usage:  "List all pipelines",
			Action: listPipelines,
		},
		{
			Name:      "get",
			Usage:     "Get details of a specific pipeline",
			ArgsUsage: "<pipeline-id>",
			Action:    getPipeline,
		},
		{
			Name:      "create",
			Usage:     "Create a new pipeline",
			ArgsUsage: "<stages-json>",
			Action:    createPipeline,
		},
		{
			Name:      "trigger",
			Usage:     "Trigger a pipeline run",
			ArgsUsage: "<pipeline-id> <git-ref>",
			Action:    triggerPipeline,
		},
		{
			Name:   "list-runs",
			Usage:  "List all pipeline runs",
			Action: listRuns,
		},
		{
			Name:      "get-run",
			Usage:     "Get details of a specific run",
			ArgsUsage: "<run-id>",
			Action:    getRun,
		},
	},
}

func listPipelines(c *cli.Context) error {
	client := myhttp.NewClient(c.String("url"), myhttp.WithToken(c.String("token")))
	pipelines, err := client.ListPipelines(context.Background())
	if err != nil {
		return fmt.Errorf("error listing pipelines: %w", err)
	}

	for _, p := range pipelines {
		fmt.Printf("ID: %s, Name: %s, Repository: %s\n", p.ID, p.Name, p.Repository)
	}
	return nil
}

func getPipeline(c *cli.Context) error {
	if c.NArg() < 1 {
		return fmt.Errorf("pipeline ID required")
	}

	client := myhttp.NewClient(c.String("url"), myhttp.WithToken(c.String("token")))
	pipeline, err := client.GetPipeline(context.Background(), c.Args().Get(0))
	if err != nil {
		return fmt.Errorf("error getting pipeline: %w", err)
	}

	fmt.Printf("Pipeline: %+v\n", pipeline)
	return nil
}

func createPipeline(c *cli.Context) error {
	if c.NArg() < 1 {
		return fmt.Errorf("pipeline JSON definition required")
	}

	pipelineJSON := c.Args().Get(0)

	// json unmarshal
	var pipeline myhttp.PipelineRequest
	if err := json.Unmarshal([]byte(pipelineJSON), &pipeline); err != nil {
		return fmt.Errorf("error unmarshalling pipeline: %w", err)
	}

	client := myhttp.NewClient(c.String("url"), myhttp.WithToken(c.String("token")))

	resp, err := client.CreatePipeline(context.Background(), pipeline)
	if err != nil {
		return fmt.Errorf("error creating pipeline: %w", err)
	}

	fmt.Printf("Pipeline created. ID: %s\n", resp.ID)
	return nil
}

func triggerPipeline(c *cli.Context) error {
	if c.NArg() < 2 {
		return fmt.Errorf("pipeline ID and Git ref required")
	}

	client := myhttp.NewClient(c.String("url"), myhttp.WithToken(c.String("token")))
	resp, err := client.TriggerPipeline(context.Background(), c.Args().Get(0), c.Args().Get(1))
	if err != nil {
		return fmt.Errorf("error triggering pipeline: %w", err)
	}

	fmt.Printf("Pipeline triggered. Run ID: %s\n", resp.ID)
	return nil
}

func listRuns(c *cli.Context) error {
	client := myhttp.NewClient(c.String("url"), myhttp.WithToken(c.String("token")))
	runs, err := client.ListRuns(context.Background())
	if err != nil {
		return fmt.Errorf("error listing runs: %w", err)
	}

	for _, r := range runs {
		fmt.Printf("Run ID: %s, Status: %s\n", r.ID, r.Status)
	}
	return nil
}

func getRun(c *cli.Context) error {
	if c.NArg() < 1 {
		return fmt.Errorf("run ID required")
	}

	client := myhttp.NewClient(c.String("url"), myhttp.WithToken(c.String("token")))
	run, err := client.GetRun(context.Background(), c.Args().Get(0))
	if err != nil {
		return fmt.Errorf("error getting run: %w", err)
	}

	fmt.Printf("Run: %+v\n", run)
	return nil
}
