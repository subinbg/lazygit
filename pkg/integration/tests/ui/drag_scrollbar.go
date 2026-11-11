package ui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var DragScrollbar = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Dragging a view's scrollbar scrolls that view, and leaves its selection and the focus alone",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.ShowFileTree = false
	},
	SetupRepo: func(shell *Shell) {
		lines := make([]string, 300)
		for i := range lines {
			lines[i] = fmt.Sprintf("line %d", i)
		}
		shell.CreateNCommits(60)
		shell.CreateFileAndAdd("long", strings.Join(lines, "\n")+"\n")
		shell.Commit("long commit")

		for i := range 60 {
			shell.CreateFile(fmt.Sprintf("file%02d", i), "x\n")
		}
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		_, height := t.Views().Files().InnerSize()
		bottomOrigin := 60 - height

		t.Views().Files().
			IsFocused().
			OriginY(0).
			// grabbing the thumb at its top and dragging it to the bottom
			DragScrollbar(0, height-1).
			OriginY(bottomOrigin).
			IsFocused().
			SelectedLines(Contains("file00")).
			// and back up, grabbing it at the bottom
			DragScrollbar(height-1, 0).
			OriginY(0).
			// pressing the track moves the thumb's middle there
			DragScrollbar(height/2, height/2).
			OriginYAtLeast(1).
			IsFocused().
			SelectedLines(Contains("file00"))

		t.Views().Commits().
			Focus().
			SelectedLines(Contains("long commit"))

		_, mainHeight := t.Views().Main().InnerSize()
		t.Views().Main().
			ContainsLines(Contains("+line 0")).
			OriginY(0).
			DragScrollbar(0, mainHeight-1).
			OriginYAtLeast(mainHeight)

		// The commits view renders only the commits in view,
		// so the ones the scrollbar brings into view are rendered too.
		_, commitsHeight := t.Views().Commits().InnerSize()
		t.Views().Commits().
			IsFocused().
			DragScrollbar(0, commitsHeight-1).
			OriginY(61-commitsHeight).
			TopVisibleLine(Contains(fmt.Sprintf("commit-%02d", commitsHeight))).
			DragScrollbar(commitsHeight-1, 0).
			OriginY(0).
			SelectedLines(Contains("long commit"))
	},
})
