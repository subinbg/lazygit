package helpers

import (
	"regexp"

	"github.com/jesseduffield/lazygit/pkg/gocui"
	"github.com/jesseduffield/lazygit/pkg/utils"
)

// the line-number gutter of a pretty diff row (see the prettydiff package), up
// to and including its separator
var prettyDiffGutterRegexp = regexp.MustCompile(`^ *\d* +\d* *│`)

// InSelectionGutter reports whether a press at the given content position of a
// main view lands on the gutter of a row that shows a line of the diff: the
// line numbers of a pretty diff, or the +/- column of git's own. Dragging from
// there selects diff lines; dragging from anywhere else selects text (see
// gui.mouseTextSelection).
func (self *DiffLineHelper) InSelectionGutter(view *gocui.View, x int, y int) bool {
	info, ok := self.GetDiffLineInfo(view, y)
	if !ok || !info.IsContent() {
		return false
	}
	text, ok := view.ViewLineText(y)
	if !ok {
		return false
	}
	return x <= gutterEnd(text)
}

// gutterEnd returns the column the gutter of a diff row ends at: the separator
// of a pretty diff's line numbers, or the +/- column of git's own diff.
func gutterEnd(text string) int {
	if loc := prettyDiffGutterRegexp.FindStringIndex(text); loc != nil {
		return utils.StringWidth(text[:loc[1]]) - 1
	}
	return 0
}
