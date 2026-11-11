package prettydiff

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// The pretty layout is no unified diff,
// so lazygit can't read back which line of which file a row shows.
// Each row that shows a line of the diff therefore starts
// with an OSC 1717 record saying so,
// which the view keeps off the screen
// (see DiffLineHelper in pkg/gui/controllers/helpers,
// and parseDiffLineMetadata there for the format).
// Rows without a record (separators, file facts, the "no newline" marker)
// belong to no line of the diff.

// The kinds of row a record names.
const (
	recordFileHeader = 'f'
	recordHunkHeader = 'h'
	recordContext    = 'c'
	recordAdded      = 'a'
	recordDeleted    = 'd'
)

// record returns the OSC 1717 record for a row.
// newLine is the row's line in the new version of the file
// (for a deletion, the line it sits at;
// for a hunk header, the hunk's first line),
// oldLine its line in the old version, which only a deletion states;
// a file header states neither.
func record(kind byte, newLine int, oldLine int, path string) string {
	newField, oldField := "", ""
	if kind != recordFileHeader {
		newField = strconv.Itoa(newLine)
	}
	if kind == recordDeleted {
		oldField = strconv.Itoa(oldLine)
	}
	return fmt.Sprintf("\x1b]1717;1;%c;%s;%s;%s\x07", kind, newField, oldField, path)
}

// recordPath returns the path a file's rows are recorded under:
// the file's path relative to the repository, as the diff header spells it,
// its new name if it has one.
// It reads the header the way lazygit reads a plain diff,
// so that a row is placed in the same file either way.
// lines are the plain lines of the file's diff,
// starting with its "diff --git" line.
func recordPath(lines []string) string {
	var oldPath, newPath, copiedOrRenamedPath string
	for _, line := range lines[1:] {
		if strings.HasPrefix(line, "@@") {
			break // past the header
		}
		switch {
		case strings.HasPrefix(line, "+++ "):
			newPath = pathFromHeaderField(strings.TrimPrefix(line, "+++ "))
		case strings.HasPrefix(line, "--- "):
			oldPath = pathFromHeaderField(strings.TrimPrefix(line, "--- "))
		case strings.HasPrefix(line, "rename to "):
			copiedOrRenamedPath, _ = unquotePath(strings.TrimPrefix(line, "rename to "))
		case strings.HasPrefix(line, "copy to "):
			copiedOrRenamedPath, _ = unquotePath(strings.TrimPrefix(line, "copy to "))
		}
	}

	switch {
	case newPath != "" && newPath != "/dev/null":
		return newPath
	case oldPath != "" && oldPath != "/dev/null":
		return oldPath
	case copiedOrRenamedPath != "":
		return copiedOrRenamedPath
	}
	return pathFromDiffGitLine(lines[0])
}

// pathFromHeaderField returns the path in a "---" or "+++" line,
// without the prefix git puts in front of it
// (a/, b/, or a mnemonic one such as i/).
func pathFromHeaderField(field string) string {
	path, ok := unquotePath(field)
	if !ok {
		return ""
	}
	if path == "/dev/null" {
		return path
	}
	_, rest, found := strings.Cut(path, "/")
	if !found {
		return ""
	}
	return rest
}

// unquotePath undoes git's quoting of a path with unusual characters in it,
// and drops the tab git ends a path containing a space with.
func unquotePath(field string) (string, bool) {
	field = strings.TrimSuffix(field, "\t")
	if strings.HasPrefix(field, `"`) {
		unquoted, err := strconv.Unquote(field)
		if err != nil {
			return "", false
		}
		return unquoted, true
	}
	return field, true
}

// pathFromDiffGitLine returns the path in a "diff --git" line of a file
// whose path is the same on both sides,
// splitting the line at the space where the two halves name the same path.
func pathFromDiffGitLine(line string) string {
	rest := strings.TrimPrefix(line, "diff --git ")
	for i, c := range rest {
		if c != ' ' {
			continue
		}
		path := pathFromHeaderField(rest[i+1:])
		if path != "" && path == pathFromHeaderField(rest[:i]) {
			return path
		}
	}
	return ""
}

// submoduleLineRegexp matches the line a submodule's part of a diff consists of,
// capturing the submodule's path
// (see submoduleSectionPattern in pkg/gui/controllers/helpers).
var submoduleLineRegexp = regexp.MustCompile(
	`^Submodule (.+) (?:contains (?:untracked|modified) content|[0-9a-f]+\.{2,3}[0-9a-f]+(?: \(.*\))?:?)$`)
