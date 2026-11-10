package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var DragTextOfACommitsDiff = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Dragging over a commit's header in the main view selects its text and copies it, leaving the focus in the commits panel; a click without a drag still focuses the main view",
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
		shell.CreateFileAndAdd("file2", "two\n")
		shell.Commit("second commit")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Commits().
			Focus().
			Lines(
				Contains("second commit").IsSelected(),
				Contains("first commit"),
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

		// The focus stayed in the commits panel, so the keys still move through it.
		t.Views().Commits().
			IsFocused().
			Press(keys.Universal.NextItem).
			Lines(
				Contains("second commit"),
				Contains("first commit").IsSelected(),
			)

		// A click without a drag focuses the main view, as without text selection.
		t.Views().Main().
			Click(0, 0).
			IsFocused()
	},
})
