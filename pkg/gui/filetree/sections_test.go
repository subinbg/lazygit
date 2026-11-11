package filetree

import (
	"testing"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestBuildSectionedTreeFromFiles(t *testing.T) {
	files := []*models.File{
		{Path: "dir/a", HasStagedChanges: true, HasUnstagedChanges: true},
		{Path: "dir/b", HasStagedChanges: true},
		{Path: "dir/sub/c", HasStagedChanges: true},
		{Path: "top", HasUnstagedChanges: true},
		{Path: "new"},
	}

	root := BuildSectionedTreeFromFiles(files, NodeSortComparator[models.File]("mixed", false))
	collapsedPaths := NewCollapsedPaths()
	nodes := root.Flatten(collapsedPaths)[1:]

	assert.Equal(t, []string{
		UnstagedSectionPath,
		UnstagedSectionPath + "/dir",
		UnstagedSectionPath + "/dir/a",
		UnstagedSectionPath + "/new",
		UnstagedSectionPath + "/top",
		StagedSectionPath,
		StagedSectionPath + "/dir",
		StagedSectionPath + "/dir/b",
		StagedSectionPath + "/dir/sub",
		StagedSectionPath + "/dir/sub/c",
	}, lo.Map(nodes, func(node *Node[models.File], _ int) string { return node.GetInternalPath() }))

	// the logical paths are the usual ones, and a section has none
	assert.Equal(t, []string{"", "dir", "dir/a", "new", "top", "", "dir", "dir/b", "dir/sub", "dir/sub/c"},
		lo.Map(nodes, func(node *Node[models.File], _ int) string { return node.GetPath() }))
	assert.Equal(t, []bool{true, false, false, false, false, true, false, false, false, false},
		lo.Map(nodes, func(node *Node[models.File], _ int) bool { return node.IsSection() }))

	// a section is never compressed into its only directory,
	// while that directory compresses with its own only subdirectory;
	// and a section without files is still there, to show that it is empty
	onlyStaged := BuildSectionedTreeFromFiles(
		[]*models.File{{Path: "a/b/c", HasStagedChanges: true}},
		NodeSortComparator[models.File]("mixed", false))
	assert.Len(t, onlyStaged.Children, 2)
	assert.True(t, onlyStaged.Children[0].IsEmptySection())
	section := onlyStaged.Children[1]
	assert.True(t, section.IsSection())
	assert.False(t, section.IsEmptySection())
	assert.Equal(t, StagedSectionPath+"/a/b", section.Children[0].GetInternalPath())
	assert.Equal(t, []string{"a/b", "a"}, section.Children[0].GetPaths())

	// no files, no sections
	assert.Empty(t, BuildSectionedTreeFromFiles(nil, NodeSortComparator[models.File]("mixed", false)).Children)

	// collapsing a directory in one section leaves the other section's alone
	collapsedPaths.Collapse(StagedSectionPath + "/dir")
	assert.Len(t, root.Flatten(collapsedPaths)[1:], 7)
}
