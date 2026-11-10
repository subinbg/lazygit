package filetree

import (
	"strings"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
)

// With gui.showStagingSectionsInFilesView, the tree view splits the files into
// an unstaged and a staged section, each a directory tree of its own under a
// node of its own (see BuildSectionedTreeFromFiles). A directory with files in
// both sections appears in both, so the nodes of a section have internal paths
// under the section's (e.g. "\x00staged/pkg/a.go"), which keeps their collapsed
// state and their place in the list apart; their logical paths (GetPath) are
// the usual ones.
const (
	UnstagedSectionPath = sectionMarker + "unstaged"
	StagedSectionPath   = sectionMarker + "staged"

	// A NUL can't be part of a path in git, so nothing else starts with it.
	sectionMarker = "\x00"
)

// InUnstagedSection reports whether a file belongs in the unstaged section:
// it has unstaged changes (a file with both kinds of changes is listed there),
// or it is untracked. The staged section has the files whose changes are all
// staged.
func InUnstagedSection(file *models.File) bool {
	return file.HasUnstagedChanges || !file.HasStagedChanges
}

// IsSection reports whether the node is the node of a section, as opposed to a
// file or a directory.
func (self *Node[T]) IsSection() bool {
	return isSectionPath(self.path)
}

func isSectionPath(internalPath string) bool {
	return strings.HasPrefix(internalPath, sectionMarker) && !strings.Contains(internalPath, "/")
}

// logicalPath turns an internal path into the path from the user's point of
// view: without the "./" of the root item, or the section a node is in. A
// section's own logical path is empty.
func logicalPath(internalPath string) string {
	if strings.HasPrefix(internalPath, sectionMarker) {
		_, rest, found := strings.Cut(internalPath, "/")
		if !found {
			return ""
		}
		internalPath = rest
	}
	return strings.TrimPrefix(internalPath, "./")
}

// BuildSectionedTreeFromFiles builds a tree whose top-level nodes are the
// unstaged and the staged section (an empty section is left out), each holding
// a directory tree of its files.
func BuildSectionedTreeFromFiles(
	files []*models.File,
	cmp func(a, b *Node[models.File]) int,
) *Node[models.File] {
	var unstaged, staged []*models.File
	for _, file := range files {
		if InUnstagedSection(file) {
			unstaged = append(unstaged, file)
		} else {
			staged = append(staged, file)
		}
	}

	root := &Node[models.File]{}
	for _, section := range []struct {
		path  string
		files []*models.File
	}{
		{UnstagedSectionPath, unstaged},
		{StagedSectionPath, staged},
	} {
		if len(section.files) > 0 {
			root.Children = append(root.Children, buildSectionTree(section.path, section.files, cmp))
		}
	}
	return root
}

func buildSectionTree(
	sectionPath string,
	files []*models.File,
	cmp func(a, b *Node[models.File]) int,
) *Node[models.File] {
	section := &Node[models.File]{path: sectionPath}
	childrenByNode := map[*Node[models.File]]map[string]*Node[models.File]{}

	for _, file := range files {
		splitPath := split(sectionPath + "/" + file.Path)
		curr := section
		for i := 1; i < len(splitPath); i++ {
			path := join(splitPath[:i+1])
			children, ok := childrenByNode[curr]
			if !ok {
				children = map[string]*Node[models.File]{}
				childrenByNode[curr] = children
			}
			child, ok := children[path]
			if !ok {
				child = &Node[models.File]{path: path}
				if i == len(splitPath)-1 {
					child.File = file
				}
				curr.Children = append(curr.Children, child)
				children[path] = child
			}
			curr = child
		}
	}

	section.Sort(cmp)
	// compresses the directories below the section, never the section itself
	section.Compress()
	return section
}

// sectionedPaths returns the internal paths a path that doesn't name its
// section stands for in a sectioned tree: the path in either section.
func sectionedPaths(path string) []string {
	path = strings.TrimPrefix(path, "./")
	return []string{UnstagedSectionPath + "/" + path, StagedSectionPath + "/" + path}
}
