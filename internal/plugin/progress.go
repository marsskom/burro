package plugin

import (
	"fmt"
	"sync"

	"gitlab.com/marsskom/burro/internal/cli"
)

type Progress struct {
	Name    string
	Current int
	Total   int
}

type GlobalProgress struct {
	Current int
	Total   int
	Plugins map[string]Progress
}

type ProgressRenderer struct {
	cliIO cli.IO

	mu sync.Mutex
}

func (r *ProgressRenderer) Render(g GlobalProgress) {
	r.mu.Lock()
	defer r.mu.Unlock()

	cli.Clear(r.cliIO)
	cli.ProgressBar(r.cliIO, g.Current, g.Total, 35)

	fmt.Fprintf(r.cliIO.Out, " loading plugins (%d/%d)\n",
		g.Current,
		g.Total,
	)

	for _, p := range g.Plugins {
		fmt.Fprintf(r.cliIO.Out, "  %-15s (%d/%d)", p.Name, p.Current, p.Total)
		cli.ProgressBar(r.cliIO, p.Current, p.Total, 30)
		fmt.Fprint(r.cliIO.Out, "\n")
	}
}
