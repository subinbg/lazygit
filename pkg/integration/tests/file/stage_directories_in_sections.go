package file

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var StageDirectoriesInSections = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "The tree view split into an unstaged and a staged section: a directory stages or unstages the files listed under it, a section all of its files",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(config *config.AppConfig) {
		config.GetUserConfig().Gui.ShowFileTree = true
		config.GetUserConfig().Gui.ShowStagingSectionsInFilesView = true
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateDir("dir")
		shell.CreateFileAndAdd("dir/a", "1\n")
		shell.CreateFileAndAdd("dir/b", "1\n")
		shell.CreateFileAndAdd("top", "1\n")
		shell.Commit("one")

		// a has staged and unstaged changes, b only staged ones,
		// top only unstaged ones
		shell.UpdateFileAndAdd("dir/a", "2\n")
		shell.UpdateFile("dir/a", "3\n")
		shell.UpdateFileAndAdd("dir/b", "2\n")
		shell.UpdateFile("top", "2\n")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			Lines(
				Equals("▼ Unstaged files (2)").IsSelected(),
				Equals("  ▼ dir"),
				Equals("    MM a"),
				Equals("   M top"),
				Equals("▼ Staged files (1)"),
				Equals("  ▼ dir"),
				Equals("    M  b"),
			).
			// The staged section's dir has only b in it:
			// unstaging it leaves the staged changes of a,
			// which is listed in the other section, alone.
			NavigateToLine(Equals("    M  b")).
			Press(keys.Universal.PrevItem).
			PressPrimaryAction().
			Lines(
				Equals("▼ Unstaged files (3)"),
				Equals("  ▼ dir").IsSelected(),
				Equals("    MM a"),
				Equals("     M b"),
				Equals("   M top"),
				Equals("▼ Staged files (0)"),
			).
			// The unstaged section's dir stages everything in it.
			PressPrimaryAction().
			Lines(
				Equals("▼ Unstaged files (1)"),
				Equals("   M top"),
				Equals("▼ Staged files (2)"),
				Equals("  ▼ dir").IsSelected(),
				Equals("    M  a"),
				Equals("    M  b"),
			).
			// A section's node acts on all the files of the section.
			NavigateToLine(Equals("▼ Staged files (2)")).
			PressPrimaryAction().
			Lines(
				Equals("▼ Unstaged files (3)"),
				Equals("  ▼ dir"),
				Equals("     M a"),
				Equals("     M b"),
				Equals("   M top"),
				Equals("▼ Staged files (0)").IsSelected(),
			).
			// An empty section is listed to show that it is empty,
			// and has nothing to stage or unstage.
			PressPrimaryAction().
			Tap(func() {
				t.ExpectToast(Equals("Disabled: There are no files in this section."))
			}).
			// Sections collapse like directories.
			NavigateToLine(Equals("▼ Unstaged files (3)")).
			PressEnter().
			Lines(
				Equals("▶ Unstaged files (3)").IsSelected(),
				Equals("▼ Staged files (0)"),
			).
			// A section has no path to open or ignore.
			Press(keys.Files.IgnoreFile)

		t.ExpectToast(Contains("Not available on a section"))
	},
})
