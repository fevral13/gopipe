package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/xanzy/go-gitlab"
	"math"
	"time"
)

var titles = []string{
	"ID",
	"Branch",
	"Status",
	"Started At",
	"Duration",
	"Author",
}

type PipelinesSnapshot = []*gitlab.Pipeline

type Pipelines struct {
	tview.TableContentReadOnly
	pipelinesSnapshot PipelinesSnapshot
}

func (ps *Pipelines) GetCell(row, column int) *tview.TableCell {
	if row == 0 {
		return getHeader(column)
	}

	pipeline := ps.pipelinesSnapshot[row-1]
	content := ""
	switch column {
	case 0:
		content = fmt.Sprintf("%v", pipeline.ID)
	case 1:
		content = pipeline.Ref
	case 2:
		content = pipeline.Status
	case 3:
		if pipeline.StartedAt != nil {
			content = pipeline.StartedAt.Local().Format(time.DateTime)
		} else {
			content = "..."
		}
	case 4:
		if pipeline.Status == "running" {
			pipelineDuration := time.Now().Sub(*pipeline.StartedAt).Seconds()
			minutes := math.Round(pipelineDuration / 60)
			seconds := math.Round(math.Mod(pipelineDuration, 60))

			content = fmt.Sprintf("%v:%02.f", minutes, seconds)
		} else {
			content = ""
		}
	case 5:
		content = pipeline.User.Name
	default:
		content = ""
	}

	cell := tview.NewTableCell(content)
	if pipeline.Status == "failed" {
		cell.SetTextColor(tcell.ColorRed)
	} else if pipeline.Status == "success" {
		cell.SetTextColor(tcell.ColorGreen)
	}

	return cell
}

func (ps *Pipelines) GetRowCount() int {
	return len(ps.pipelinesSnapshot) + 1
}

func (ps *Pipelines) GetColumnCount() int {
	return len(titles)
}

func getHeader(column int) *tview.TableCell {
	content := titles[column]

	cell := tview.NewTableCell(content)
	cell.SetTextColor(tcell.ColorGray)
	return cell
}
