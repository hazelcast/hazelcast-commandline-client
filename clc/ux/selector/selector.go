package selector

import (
	"context"

	"github.com/pterm/pterm"
)

func Show(ctx context.Context, text string, options ...string) (result string, canceled bool, err error) {
	widget := pterm.DefaultInteractiveSelect.WithOptions(options)
	widget.TextStyle = &pterm.ThemeDefault.DefaultText
	widget.SelectorStyle = pterm.NewStyle(pterm.Bold)
	widget = widget.WithOnInterruptFunc(func() {
		canceled = true
	})
	option, err := widget.Show(text)
	if err != nil {
		return "", false, err
	}
	return option, canceled, nil
}
