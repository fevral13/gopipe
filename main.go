package main

import (
	"context"
	"fmt"
	"github.com/rivo/tview"
	"github.com/xanzy/go-gitlab"
	"strings"
	"time"
)

/*
Display the state of all main branches: development, release-23.x

... And the state of all open MR's pipelines

Iterate over main branches and for each get last pipeline info -> pipelines
Iterate over open MRs, for each get its source_branch name -> last pipeline info
*/
type appState struct {
	pipelines PipelinesSnapshot
	events    []string
}

func newAppState() *appState {
	return &appState{
		pipelines: make(PipelinesSnapshot, 0),
		events:    make([]string, 0),
	}
}

type eventsChannel chan string
type pipelinesSnapshotChannel chan PipelinesSnapshot
type appStateChannel chan appState

func streamPipelines(
	config *AppConfig,
	client *gitlab.Client,
	eventsCh eventsChannel,
	pipelinesCh pipelinesSnapshotChannel,
) {
	ticker := time.NewTicker(time.Duration(config.delay) * time.Second)
	defer ticker.Stop()

	for ; true; <-ticker.C {
		pipelines := PipelinesSnapshot{}

		for _, branchName := range config.mainBranches {
			options := gitlab.GetLatestPipelineOptions{Ref: &branchName}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			p, _, err := client.Pipelines.GetLatestPipeline(config.projectId, &options, gitlab.WithContext(ctx))

			if err != nil {
				eventsCh <- fmt.Sprintf("Error fetching branch %s", branchName)
				continue
			}

			eventsCh <- fmt.Sprintf("Fetched branch %s", branchName)
			pipelines = append(pipelines, p)
		}

		stateOpened := "opened"
		scope := "all"
		options := gitlab.ListProjectMergeRequestsOptions{State: &stateOpened, Scope: &scope}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		mergeRequests, _, err := client.MergeRequests.ListProjectMergeRequests(
			config.projectId,
			&options,
			gitlab.WithContext(ctx),
		)

		if err != nil {
			eventsCh <- fmt.Sprintf("Error project merge requests %d", config.projectId)
		} else {
			for _, mergeRequest := range mergeRequests {
				options := gitlab.GetLatestPipelineOptions{Ref: &mergeRequest.SourceBranch}
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				p, _, err := client.Pipelines.GetLatestPipeline(
					config.projectId,
					&options,
					gitlab.WithContext(ctx),
				)

				if err != nil {
					eventsCh <- fmt.Sprintf("Error fetching branch %s", mergeRequest.SourceBranch)
					continue
				}

				eventsCh <- fmt.Sprintf("Fetched branch %s", mergeRequest.SourceBranch)
				pipelines = append(pipelines, p)
			}
		}

		pipelinesCh <- pipelines
	}
}

func mustGetGitlabClient(apiKey, baseUrl string) *gitlab.Client {
	glab, err := gitlab.NewClient(apiKey, gitlab.WithBaseURL(baseUrl))

	if err != nil {
		panic("Error creating GL client")
	}
	return glab
}

func updateUI(
	appStateCh appStateChannel,
	application *tview.Application,
	view *tview.Table,
	logBox *tview.TextView,
) {
	for pipelines := range appStateCh {
		p := Pipelines{
			pipelinesSnapshot: pipelines.pipelines,
			events:            pipelines.events,
		}
		logBox.SetText(strings.Join(p.events, "\n"))

		application.QueueUpdateDraw(func() {
			view.SetContent(&p)
		})
	}
}

func aggregator(evensCh eventsChannel, pipelinesCh pipelinesSnapshotChannel, appStateCh appStateChannel) {
	appState := newAppState()

	for {
		appStateCh <- *appState

		select {
		case newEvent := <-evensCh:
			if len(appState.events) > 10 {
				appState.events = appState.events[1:]
			}
			appState.events = append(appState.events, newEvent)
		case pipelines := <-pipelinesCh:
			appState.pipelines = pipelines
		}
	}
}

func main() {
	config := getConfig()

	gl := mustGetGitlabClient(config.apiKey, config.apiUrl)

	eventsCh := make(eventsChannel)
	pipelinesCh := make(pipelinesSnapshotChannel)
	appStateCh := make(appStateChannel)

	go aggregator(eventsCh, pipelinesCh, appStateCh)
	go streamPipelines(config, gl, eventsCh, pipelinesCh)

	app := tview.NewApplication()
	table := tview.NewTable()
	logBox := tview.NewTextView().SetChangedFunc(func() {
		app.Draw()
	})
	logBox.SetMaxLines(10)

	grid := tview.NewGrid()
	grid.SetBorder(true)
	grid.SetTitle("Pipelines")
	grid.SetColumns(-1)
	grid.SetRows(-3, -1)
	grid.AddItem(table, 0, 0, 1, 1, 10, 10, true)
	grid.AddItem(logBox, 1, 0, 1, 1, 10, 10, false)

	go updateUI(appStateCh, app, table, logBox)

	if err := app.SetRoot(grid, true).Run(); err != nil {
		panic(err)
	}
}
