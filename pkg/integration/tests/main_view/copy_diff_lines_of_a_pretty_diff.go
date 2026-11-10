package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var CopyDiffLinesOfAPrettyDiff = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Copy the selected lines of a pretty diff: the clipboard gets the lines as the diff reads, without the line-number gutter",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.PrettyDiff = true
		cfg.GetUserConfig().Gui.UseHunkModeInDiffView = true
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

		t.Views().Main().
			IsFocused().
			SelectedLines(
				Contains("3 │+ADD1"),
				Contains("4 │+ADD2"),
			).
			Press(keys.Universal.CopyToClipboard)

		expectClipboard(t, Equals("ADD1\nADD2\n"))
	},
})
