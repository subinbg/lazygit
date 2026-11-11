package controllers

import (
	"github.com/jesseduffield/lazygit/pkg/gui/context"
)

// clickUnlessSelectingText runs a click on a main pane now or,
// when its press started a text selection (see gui.mouseTextSelection),
// when the mouse button is released,
// unless the press turned into a drag that selected text.
// So selecting text to copy it leaves the focus and the diff selection alone.
func clickUnlessSelectingText(pane *context.MainContext, click func() error) error {
	if pane.GetView().IsSelectingText() {
		pane.SetPendingClick(click)
		return nil
	}
	return click()
}
