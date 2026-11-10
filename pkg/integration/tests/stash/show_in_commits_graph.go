package stash

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var ShowInCommitsGraph = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "The commits graph shows a stash as a node of its own above the commit it was made on; the row can't be selected, and goes when the stash is dropped",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(config *config.AppConfig) {
		config.GetUserConfig().Git.Log.GraphStyle = "lanes"
		config.GetUserConfig().Gui.ShowStashesInCommitsView = true
	},
	SetupRepo: func(shell *Shell) {
		shell.EmptyCommit("one")
		shell.EmptyCommit("two")
		shell.CreateFile("file", "content")
		shell.GitAddAll()
		shell.Stash("my stash")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Commits().
			Focus().
			Lines(
				Contains("▣").Contains("stash@{0}").Contains("On master: my stash"),
				Contains("●─╯").Contains("two").IsSelected(),
				Contains("●").Contains("one"),
			).
			// the stash's row is skipped like a header
			Press(keys.Universal.PrevItem).
			Lines(
				Contains("stash@{0}"),
				Contains("two").IsSelected(),
				Contains("one"),
			)

		t.Views().Stash().
			Focus().
			Lines(
				Contains("my stash").IsSelected(),
			).
			Press(keys.Universal.Remove).
			Tap(func() {
				t.ExpectPopup().Confirmation().
					Title(Equals("Stash drop")).
					Content(Contains("Are you sure you want to drop the selected stash entry(ies)?")).
					Confirm()
			}).
			IsEmpty()

		t.Views().Commits().
			Lines(
				Contains("two"),
				Contains("one"),
			)
	},
})
