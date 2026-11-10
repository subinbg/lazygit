package git_commands

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/utils"
	"github.com/samber/lo"
)

type RemoteLoader struct {
	*GitCommon
}

func NewRemoteLoader(gitCommon *GitCommon) *RemoteLoader {
	return &RemoteLoader{GitCommon: gitCommon}
}

// GetRemotes returns the repo's remotes, without their branches; those are
// loaded separately with GetRemoteBranchesByRemoteName, which takes a lot longer
// in a repo with many remote branches.
func (self *RemoteLoader) GetRemotes() []*models.Remote {
	remotes := self.getRemotesFromConfig()

	// now lets sort our remotes by name alphabetically
	slices.SortFunc(remotes, func(a, b *models.Remote) int {
		// we want origin at the top because we'll be most likely to want it
		if a.Name == "origin" {
			return -1
		}
		if b.Name == "origin" {
			return 1
		}
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})

	return remotes
}

func (self *RemoteLoader) getRemotesFromConfig() []*models.Remote {
	cmdArgs := NewGitCmd("config").
		Arg("--local", "--get-regexp", `^remote\.[^.]+\.(url|pushurl)$`).ToArgv()
	output, err := self.cmd.New(cmdArgs).DontLog().RunWithOutput()
	if err != nil {
		// exit code 1 means no matching keys (no remotes configured)
		return nil
	}

	remotesByName := make(map[string]*models.Remote)

	for _, line := range strings.Split(output, "\n") {
		key, url, found := strings.Cut(strings.TrimSpace(line), " ")
		if !found {
			continue
		}
		// key is "remote.<name>.url" or "remote.<name>.pushurl";
		// strip prefix and suffix to get the name
		rest, ok := strings.CutPrefix(key, "remote.")
		if !ok {
			continue
		}
		var remoteName string
		var isPushUrl bool
		if name, ok := strings.CutSuffix(rest, ".pushurl"); ok {
			remoteName, isPushUrl = name, true
		} else if name, ok := strings.CutSuffix(rest, ".url"); ok {
			remoteName, isPushUrl = name, false
		} else {
			continue
		}
		if _, ok := remotesByName[remoteName]; !ok {
			remotesByName[remoteName] = &models.Remote{Name: remoteName}
		}
		if isPushUrl {
			remotesByName[remoteName].PushUrls = append(remotesByName[remoteName].PushUrls, url)
		} else {
			remotesByName[remoteName].Urls = append(remotesByName[remoteName].Urls, url)
		}
	}

	return slices.Collect(maps.Values(remotesByName))
}

// GetRemoteBranchesByRemoteName returns all remote branches, keyed by the name
// of the remote they belong to.
func (self *RemoteLoader) GetRemoteBranchesByRemoteName() (map[string][]*models.RemoteBranch, error) {
	remoteBranches, err := self.getRemoteBranches()
	if err != nil {
		return nil, err
	}

	return lo.GroupBy(remoteBranches, func(branch *models.RemoteBranch) string {
		return branch.RemoteName
	}), nil
}

// Returns all remote branches, sorted the way the config asks for
func (self *RemoteLoader) getRemoteBranches() ([]*models.RemoteBranch, error) {
	sortByDate := strings.ToLower(self.UserConfig().Git.RemoteBranchSortOrder) == "date"
	sortOrder := "refname"
	if sortByDate {
		sortOrder = "-committerdate"
	}

	cmdArgs := NewGitCmd("for-each-ref").
		Arg(fmt.Sprintf("--sort=%s", sortOrder)).
		Arg("--format=%(refname)%00%(objectname)%00%(committerdate:unix)").
		Arg("refs/remotes").
		ToArgv()

	remoteBranches := []*models.RemoteBranch{}
	tips := map[string]refTip{}
	err := self.cmd.New(cmdArgs).DontLog().RunAndProcessLines(func(line string) (bool, error) {
		fields := strings.Split(strings.TrimSpace(line), "\x00")
		refName := fields[0]

		if len(fields) != 3 {
			return false, nil
		}
		commitHash := fields[1]
		recency := ""
		if timestamp, err := strconv.ParseInt(fields[2], 10, 64); err == nil {
			recency = utils.UnixToTimeAgo(timestamp)
		}

		split := strings.SplitN(refName, "/", 4)
		if len(split) != 4 {
			return false, nil
		}
		remoteName := split[2]
		name := split[3]

		if name == "HEAD" {
			return false, nil
		}

		remoteBranches = append(remoteBranches,
			&models.RemoteBranch{
				Name:       name,
				RemoteName: remoteName,
				CommitHash: commitHash,
				Recency:    recency,
			})
		if sortByDate {
			tips[refName] = refTip{hash: commitHash, committerDate: fields[2]}
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	if sortByDate {
		if err := sortRefsWithEqualDatesByAncestry(
			self.cmd, self.version, remoteBranches, (*models.RemoteBranch).FullRefName, tips,
		); err != nil {
			self.Log.Errorf("Failed to sort remote branches by ancestry: %v", err)
		}
	}

	return remoteBranches, nil
}
