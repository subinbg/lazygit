package presentation

import (
	"fmt"
	"slices"
	"strings"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/common"
	"github.com/jesseduffield/lazygit/pkg/gui/presentation/lanegraph"
	"github.com/jesseduffield/lazygit/pkg/gui/presentation/palette"
	"github.com/jesseduffield/lazygit/pkg/theme"
)

// stashPlacement is where a stash's row goes in the lanes graph.
type stashPlacement struct {
	lanegraph.Stash
	entry *models.StashEntry
}

// placeStashes places the stashes made on one of the commits: each goes above
// the first commit older than the stash, like the commits of a date-ordered
// log, but never below the commit it was made on. A stash made on a commit that
// isn't among them (one not loaded yet, or no longer on any branch) is left
// out. The placements are in the order of their rows.
func placeStashes(commits []*models.Commit, stashes []*models.StashEntry) []stashPlacement {
	if len(stashes) == 0 {
		return nil
	}

	indexOf := make(map[string]int, len(commits))
	for i, commit := range commits {
		indexOf[commit.Hash()] = i
	}

	var placements []stashPlacement
	for _, stash := range stashes {
		baseIdx, ok := indexOf[stash.BaseHash]
		if !ok {
			continue
		}
		before := baseIdx
		for i := range baseIdx {
			if commits[i].UnixTimestamp < stash.UnixTimestamp {
				before = i
				break
			}
		}
		placements = append(placements, stashPlacement{
			Stash: lanegraph.Stash{Hash: stash.Hash, BaseHash: stash.BaseHash, Before: before},
			entry: stash,
		})
	}

	// newest first among the stashes going above the same commit, as listed
	slices.SortStableFunc(placements, func(a, b stashPlacement) int {
		return a.Before - b.Before
	})
	return placements
}

// stashSignature identifies the stash rows of a graph, for its cache key.
func stashSignature(placements []stashPlacement) string {
	var builder strings.Builder
	for _, placement := range placements {
		fmt.Fprintf(&builder, "%s:%d;", placement.Hash, placement.Before)
	}
	return builder.String()
}

// StashRow is the row of a stash in the commits list (see StashRows).
type StashRow struct {
	// index of the commit the row goes above
	Index int
	// the row from the graph column on
	Content string
}

// StashRows returns the rows the commits list shows for opts.Stashes, like
// GitKraken's stash nodes: the stash's node in the lanes graph, with a dotted
// line to the commit it was made on, followed by a label and the stash's
// message. opts are the ones the list is rendered with, so that the rows are
// laid out in the same graph. There are none when the graph is hidden or not in
// the lanes style.
func StashRows(common *common.Common, commits []*models.Commit, showGraph bool, opts CommitListOpts) []StashRow {
	if !showGraph || len(commits) == 0 || len(opts.Stashes) == 0 ||
		common.UserConfig().Git.Log.GraphStyle == "classic" {
		return nil
	}

	mutex.Lock()
	defer mutex.Unlock()

	entry := loadLaneGraph(commits, opts)
	width := entry.graph.Width()
	var rows []StashRow
	for k, placement := range placeStashes(commits, opts.Stashes) {
		rowIdx := entry.graph.RowIndexOfStash(k)
		if rowIdx < 0 {
			continue
		}
		lane, _ := entry.graph.LaneOfStash(placement.Stash)
		rows = append(rows, StashRow{
			Index: placement.Before,
			Content: entry.graph.RenderRow(rowIdx, width) +
				palette.DimLabel(lane).Sprint(" "+placement.entry.RefName()+" ") + " " +
				theme.DefaultTextColor.Sprint(placement.entry.Name),
		})
	}
	return rows
}
