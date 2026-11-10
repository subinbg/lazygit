package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var DragTextOrLinesOfAPrettyDiff = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Over a pretty diff, dragging from the line numbers selects diff lines, and dragging over the text selects the text",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.PrettyDiff = true
		cfg.GetUserConfig().Gui.MouseTextSelection = true
		cfg.GetUserConfig().Gui.UseHunkModeInDiffView = false
		// Emulate the clipboard by writing to a file.
		cfg.GetUserConfig().OS.CopyToClipboardCmd = "printf '%s' {{text}} > clipboard"
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateFileAndAdd("file1", "one\ntwo\nthree\n")
		shell.Commit("one")

		shell.UpdateFile("file1", "one\ntwo\nADD1\nADD2\nthree\n")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			Press(keys.Universal.FocusMainView)

		// The rows: the file's header, the hunk's, " one", " two", "+ADD1", "+ADD2",
		// " three". Each line of the file starts with a gutter of line numbers that
		// ends in a "│" at column 8.
		t.Views().Main().
			IsFocused().
			ContainsLines(
				Contains("3 │+ADD1"),
				Contains("4 │+ADD2"),
			).
			ClickAndHold(5, 4).
			MouseMove(5, 5).
			MouseRelease().
			SelectedLines(
				Contains("│+ADD1"),
				Contains("│+ADD2"),
			)

		// Dragging over the text copies the text as it reads on screen, and leaves
		// the selected diff line where the drag started.
		t.Views().Main().
			IsFocused().
			ClickAndHold(10, 4).
			MouseMove(13, 4).
			MouseRelease()

		t.ExpectToast(Equals("Selection copied to clipboard"))
		t.FileSystem().FileContent("clipboard", Equals("ADD1"))
		t.Shell().DeleteFile("clipboard")

		t.Views().Main().
			SelectedLines(
				Contains("│+ADD1"),
			)
	},
})
