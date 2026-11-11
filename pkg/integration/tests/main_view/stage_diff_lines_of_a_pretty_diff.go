package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var StageDiffLinesOfAPrettyDiff = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Stage a line and a hunk from the focused main view while it shows the pretty diff, whose rows state the diff lines they show",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.PrettyDiff = true
		cfg.GetUserConfig().Gui.UseHunkModeInDiffView = false
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateFileAndAdd("file1", "one\ntwo\nthree\nfour\nfive\nsix\nseven\neight\nnine\nten\n")
		shell.Commit("one")

		// Two change blocks, far enough apart to stay separate hunks.
		shell.UpdateFile("file1", "one\ntwo\nADD1\nADD2\nthree\nfour\nfive\nsix\nseven\neight\nNINE\nten\n")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			Lines(
				Contains("file1").IsSelected(),
			).
			Press(keys.Universal.FocusMainView)

		// The pretty layout stays on screen when the main view is focused:
		// the rows have line numbers in a gutter in front of them.
		t.Views().Main().
			IsFocused().
			ContainsLines(
				Contains("file1  +3 −1"),
			).
			SelectedLines(
				Contains("3 │+ADD1"),
			).
			PressPrimaryAction()

		t.Views().Files().Lines(
			Contains("MM file1"),
		)
		t.Views().Secondary().
			ContainsLines(
				Contains("│+ADD1"),
			).
			Content(DoesNotContain("+ADD2"))
		t.Views().Main().
			Content(DoesNotContain("+ADD1")).
			ContainsLines(
				Contains("│+ADD2"),
			)

		// A whole change block goes in one press.
		t.Views().Main().
			IsFocused().
			NavigateToLine(Contains("│-nine")).
			Press(keys.Main.ToggleSelectHunk).
			SelectedLines(
				Contains("│-nine"),
				Contains("│+NINE"),
			).
			PressPrimaryAction()

		t.Views().Secondary().ContainsLines(
			Contains("│-nine"),
			Contains("│+NINE"),
		)
		t.Views().Main().Content(DoesNotContain("NINE"))
	},
})
