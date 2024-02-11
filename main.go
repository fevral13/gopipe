package main

import (
	"fmt"
	"github.com/rivo/tview"
	"github.com/xanzy/go-gitlab"
	"time"
)

/*

Display the state of all main branches: development, release-23.x

... And the state of all open MR's pipelines

Iterate over main branches and for each get last pipeline info -> pipelines
Iterate over open MRs, for each get its source_branch name -> last pipeline info
*/

func streamPipelines(config *AppConfig, client *gitlab.Client, out chan<- PipelinesSnapshot) {
	ticker := time.NewTicker(time.Duration(config.delay) * time.Second)
	defer ticker.Stop()

	for ; true; <-ticker.C {
		pipelines := PipelinesSnapshot{}

		for _, branchName := range config.mainBranches {
			options := gitlab.GetLatestPipelineOptions{Ref: &branchName}
			p, _, err := client.Pipelines.GetLatestPipeline(config.projectId, &options)

			if err != nil {
				continue
			}

			pipelines = append(pipelines, p)
		}

		stateOpened := "opened"
		scope := "all"
		options := gitlab.ListProjectMergeRequestsOptions{State: &stateOpened, Scope: &scope}
		mergeRequests, _, err := client.MergeRequests.ListProjectMergeRequests(config.projectId, &options)

		if err != nil {
			fmt.Printf("Error %v", err)
			return
		} else {
			for _, mergeRequest := range mergeRequests {
				options := gitlab.GetLatestPipelineOptions{Ref: &mergeRequest.SourceBranch}
				p, _, err := client.Pipelines.GetLatestPipeline(config.projectId, &options)

				if err != nil {
					continue
				}

				pipelines = append(pipelines, p)
			}
		}

		out <- pipelines
	}
}

func getGitlabClient(config *AppConfig) *gitlab.Client {
	glab, err := gitlab.NewClient(
		config.apiKey,
		gitlab.WithBaseURL(config.apiUrl),
	)

	if err != nil {
		panic("Error creating GL client")
	}
	return glab
}

func updateUI(pipelinesStream <-chan PipelinesSnapshot, application *tview.Application, view *tview.Table) {
	for pipelines := range pipelinesStream {
		p := Pipelines{pipelinesSnapshot: pipelines}

		application.QueueUpdateDraw(func() {
			view.SetContent(&p)
		})
	}
}

func main() {
	config := getConfig()

	app := tview.NewApplication()
	view := tview.NewTable()

	gl := getGitlabClient(&config)

	out := make(chan PipelinesSnapshot)
	go streamPipelines(&config, gl, out)
	go updateUI(out, app, view)

	if err := app.SetRoot(view, true).Run(); err != nil {
		panic(err)
	}
}
