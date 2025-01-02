package domain

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

// Executor is dispatching PipelineRuns to worker go routines for execution.
type Executor struct {
	Store               Store
	workers             int
	queue               *queue
	runChan             chan *PipelineRun
	runStageExecutor    func(ctx context.Context, pipelineRun *PipelineRun, runStage *RunStage) error
	buildStageExecutor  func(ctx context.Context, pipelineRun *PipelineRun, buildStage *BuildStage) error
	deployStageExecutor func(ctx context.Context, pipelineRun *PipelineRun, deployStage *DeployStage) error
}

func NewExecutor(store Store, workers, queueSize int, maxQueuedPerPipeline int, failureRate float64, delay time.Duration) *Executor {
	return &Executor{
		Store:               store,
		workers:             workers,
		queue:               newQueue(queueSize, maxQueuedPerPipeline),
		runChan:             make(chan *PipelineRun, 1),
		runStageExecutor:    runExecFuncConstructor(failureRate, delay),
		buildStageExecutor:  buildExecFuncConstructor(failureRate, delay),
		deployStageExecutor: deployExecFuncConstructor(failureRate, delay),
	}
}

// TriggerPipeline is creating a new pipeline run and enqueuing it for execution
func (e *Executor) TriggerPipeline(ctx context.Context, pipeline *Pipeline, gitRef string) (*PipelineRun, error) {

	pipelineRun := NewPipelineRun(pipeline.ID, gitRef)

	if err := e.Store.CreatePipelineRun(ctx, pipelineRun); err != nil {
		return nil, err
	}

	if err := e.queue.Enqueue(pipelineRun); err != nil {
		pipelineRun.Status = StatusFailed
		e.Store.UpdatePipelineRun(ctx, pipelineRun)
		return nil, err
	}

	return pipelineRun, nil
}

// Start is starting the executor workersand will block until the context is cancelled
func (e *Executor) Start(ctx context.Context) {

	wg := sync.WaitGroup{}

	// start workers
	for i := 0; i < e.workers; i++ {
		wg.Add(1)
		go e.worker(ctx, &wg)
	}

	// event loop of the executor
	for {
		// get the next pipeline run from the queue
		pipelineRun, err := e.queue.Dequeue()
		if err != nil {
			if err == ErrQueueEmpty {
				// look for new pipeline runs every second
				time.Sleep(1 * time.Second)
				continue
			}
			// TODO: handle error
			panic(err)
		}

		select {
		case <-ctx.Done():
			// wait for all workers to finish
			wg.Wait()
			return
		// send the pipeline run to the workers
		case e.runChan <- pipelineRun:
			continue
		}
	}
}

// worker is reading pipeline runs from the run channel and executing them.
func (e *Executor) worker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case pipelineRun := <-e.runChan:
			log.Println("worker: picked up next pipeline run", pipelineRun.ID)
			e.execute(ctx, pipelineRun)
		}
	}
}

// logTmpl is the template for logging pipeline run events.
var logTmpl = "Pipeline: %s, Run: %s, Stage: %s, Status: %s - %s\n"

// addLog is adding a log message to the pipeline run logs and also printing it to the console
func addLog(pipelineRun *PipelineRun, stage string, status string, content string) {
	msg := fmt.Sprintf(logTmpl, pipelineRun.PipelineID, pipelineRun.ID, stage, status, content)
	pipelineRun.Logs[stage] += msg
	log.Println(msg)
}

// execute is the main logic for executing a pipeline run through all stages and is called by workers
func (e *Executor) execute(ctx context.Context, pipelineRun *PipelineRun) {

	// get the pipeline definition from the store
	pipeline, err := e.Store.GetPipeline(ctx, pipelineRun.PipelineID)
	if err != nil {
		addLog(pipelineRun, StageRun, StatusFailed, fmt.Sprintf("error getting pipeline from store: %v", err))
		pipelineRun.Status = StatusFailed
		if err := e.Store.UpdatePipelineRun(ctx, pipelineRun); err != nil {
			log.Println(err)
		}
		return
	}

	// if we find another run for this pipeline which is not finished, we need wait for it to finish
	repeat := true
	for repeat {
		otherPipelineRuns, err := e.Store.ListPipelineRuns(ctx)
		if err != nil {
			if err != ErrNotFound {
				addLog(pipelineRun, StageRun, StatusFailed, fmt.Sprintf("error getting other pipeline runs from store: %v", err))
				pipelineRun.Status = StatusFailed
				if err := e.Store.UpdatePipelineRun(ctx, pipelineRun); err != nil {
					log.Println(err)
				}
				return
			}
		}

		repeat = false
		for _, run := range otherPipelineRuns {
			if run.ID == pipelineRun.ID {
				// skip the current run
				continue
			}

			if run.PipelineID == pipeline.ID {
				// if the other run for this pipeline is not finished, we need to wait for it to finish
				if run.Status == StatusRunning {
					addLog(pipelineRun, StageRun, StatusPending, fmt.Sprintf("waiting for previous run %s to finish", run.ID))
					time.Sleep(1 * time.Second)
					repeat = true
					break
				}
			}
		}
	}

	pipelineRun.Status = StatusRunning

	// execute the run stage
	if err := e.runStageExecutor(ctx, pipelineRun, pipeline.Stages[StageRun].(*RunStage)); err != nil {
		pipelineRun.UpdatedAt = time.Now()
		if !pipeline.Stages[StageRun].ContinueOnError() {
			pipelineRun.Status = StatusFailed
			if err := e.Store.UpdatePipelineRun(ctx, pipelineRun); err != nil {
				log.Println(err)
			}
			return
		}
	}

	pipelineRun.UpdatedAt = time.Now()
	if err := e.Store.UpdatePipelineRun(ctx, pipelineRun); err != nil {
		log.Println(err)
	}

	// execute the build stage
	if err := e.buildStageExecutor(ctx, pipelineRun, pipeline.Stages[StageBuild].(*BuildStage)); err != nil {
		if !pipeline.Stages[StageBuild].ContinueOnError() {
			pipelineRun.Status = StatusFailed
			pipelineRun.UpdatedAt = time.Now()
			if err := e.Store.UpdatePipelineRun(ctx, pipelineRun); err != nil {
				log.Println(err)
			}
			return
		}
	}

	pipelineRun.UpdatedAt = time.Now()
	if err := e.Store.UpdatePipelineRun(ctx, pipelineRun); err != nil {
		log.Println(err)
	}

	// execute the deploy stage
	if err := e.deployStageExecutor(ctx, pipelineRun, pipeline.Stages[StageDeploy].(*DeployStage)); err != nil {
		pipelineRun.Status = StatusFailed
		pipelineRun.UpdatedAt = time.Now()
		if err := e.Store.UpdatePipelineRun(ctx, pipelineRun); err != nil {
			log.Println(err)
		}
		return
	}

	pipelineRun.Status = StatusSuccess
	pipelineRun.UpdatedAt = time.Now()
	if err := e.Store.UpdatePipelineRun(ctx, pipelineRun); err != nil {
		log.Println(err)
	}
}

// runExecFuncConstructor is a factory function that returns a run stage executor function
// with a given failure rate and delay for testing purposes
func runExecFuncConstructor(failureRate float64, delay time.Duration) func(ctx context.Context, pipelineRun *PipelineRun, runStage *RunStage) error {
	return func(ctx context.Context, pipelineRun *PipelineRun, runStage *RunStage) error {

		if err := runStage.Validate(); err != nil {
			pipelineRun.Logs[StageRun] = err.Error()
			pipelineRun.RunStatus = StatusFailed
			return err
		}

		addLog(pipelineRun, StageRun, StatusRunning, "starting...")
		addLog(pipelineRun, StageRun, StatusRunning, fmt.Sprintf("command: %s", runStage.Command))

		// simulate a failure
		if rand.Float64() < failureRate {
			addLog(pipelineRun, StageRun, StatusFailed, "failed")
			pipelineRun.RunStatus = StatusFailed
			return errors.New("failed")
		}

		// simulate a long running command
		time.Sleep(delay)

		addLog(pipelineRun, StageRun, StatusSuccess, "finished")
		pipelineRun.RunStatus = StatusSuccess

		return nil
	}
}

// buildExecFuncConstructor is a factory function that returns a build stage executor function
// with a given failure rate and delay
func buildExecFuncConstructor(failureRate float64, delay time.Duration) func(ctx context.Context, pipelineRun *PipelineRun, buildStage *BuildStage) error {
	return func(ctx context.Context, pipelineRun *PipelineRun, buildStage *BuildStage) error {

		if err := buildStage.Validate(); err != nil {
			pipelineRun.Logs[StageBuild] = err.Error()
			pipelineRun.BuildStatus = StatusFailed
			return err
		}

		addLog(pipelineRun, StageBuild, StatusRunning, "starting...")
		addLog(pipelineRun, StageBuild, StatusRunning, fmt.Sprintf("dockerfile path: %s", buildStage.DockerfilePath))

		// simulate a failure
		if rand.Float64() < failureRate {
			addLog(pipelineRun, StageBuild, StatusFailed, "failed")
			pipelineRun.BuildStatus = StatusFailed
			return errors.New("failed")
		}

		// simulate a long running command
		time.Sleep(delay)

		addLog(pipelineRun, StageBuild, StatusSuccess, "finished")
		pipelineRun.BuildStatus = StatusSuccess

		return nil
	}
}

// deployExecFuncConstructor is a factory function that returns a deploy stage executor function
// with a given failure rate and delay
func deployExecFuncConstructor(failureRate float64, delay time.Duration) func(ctx context.Context, pipelineRun *PipelineRun, deployStage *DeployStage) error {
	return func(ctx context.Context, pipelineRun *PipelineRun, deployStage *DeployStage) error {

		if err := deployStage.Validate(); err != nil {
			pipelineRun.Logs[StageDeploy] = err.Error()
			pipelineRun.DeployStatus = StatusFailed
			return err
		}

		addLog(pipelineRun, StageDeploy, StatusRunning, "starting...")
		addLog(pipelineRun, StageDeploy, StatusRunning, fmt.Sprintf("deploying to cluster name: %s", deployStage.ClusterName))

		// simulate a failure
		if rand.Float64() < failureRate {
			addLog(pipelineRun, StageDeploy, StatusFailed, "failed")
			pipelineRun.DeployStatus = StatusFailed
			return errors.New("failed")
		}

		// simulate a long running command
		time.Sleep(delay)

		addLog(pipelineRun, StageDeploy, StatusSuccess, "finished")
		pipelineRun.DeployStatus = StatusSuccess

		return nil
	}
}
