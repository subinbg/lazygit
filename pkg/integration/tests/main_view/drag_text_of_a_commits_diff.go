package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var DragTextOfACommitsDiff = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Dragging over a commit's header in the main view selects its text and copies it, without focusing the main view first",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.MouseTextSelection = true
		// Emulate the clipboard by writing to a file.
		cfg.GetUserConfig().OS.CopyToClipboardCmd = "printf '%s' {{text}} > clipboard"
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateFileAndAdd("file1", "one\n")
		shell.Commit("first commit")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Commits().
			Focus().
			Lines(
				Contains("first commit").IsSelected(),
			)

		t.Views().Main().
			ContainsLines(
				Contains("commit "),
			).
			ClickAndHold(0, 0).
			MouseMove(5, 0).
			MouseRelease()

		t.ExpectToast(Equals("Selection copied to clipboard"))
		t.FileSystem().FileContent("clipboard", Equals("commit"))
		t.Shell().DeleteFile("clipboard")
	},
})
