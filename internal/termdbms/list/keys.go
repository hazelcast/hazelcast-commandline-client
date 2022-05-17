package list

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the menu menu.
type KeyMap struct {
	// Keybindings used when browsing the list.
	CursorUp        key.Binding
	CursorDown      key.Binding
	NextPage        key.Binding
	PrevPage        key.Binding
	GoToStart       key.Binding
	GoToEnd         key.Binding
	Filter          key.Binding
	ClearFilter     key.Binding
	DeleteSelection key.Binding

	// Keybindings used when setting a filter.
	CancelWhileFiltering key.Binding
	AcceptWhileFiltering key.Binding

	// Help toggle keybindings.
	ShowFullHelp  key.Binding
	CloseFullHelp key.Binding

	// The quit keybinding. This won't be caught when filtering.
	Quit key.Binding

	// The quit-no-matter-what keybinding. This will be caught when filtering.
	ForceQuit key.Binding
}

// DefaultKeyMap returns a default set of keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		DeleteSelection: key.NewBinding(),
		// Browsing.
		CursorUp: key.NewBinding(
			key.WithKeys("up", "k", "w"),
			key.WithHelp("↑/k", "up"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down", "j", "s"),
			key.WithHelp("↓/j", "down"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys("left", "h", "pgup", "a"),
			key.WithHelp("←/h/pgup", "prev page"),
		),
		NextPage: key.NewBinding(
			key.WithKeys("right", "l", "pgdown", "d"),
			key.WithHelp("→/l/pgdn", "next page"),
		),
		GoToStart: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "go to start"),
		),
		GoToEnd: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "go to end"),
		),
		Filter:      key.NewBinding(),
		ClearFilter: key.NewBinding(),

		// Filtering.
		CancelWhileFiltering: key.NewBinding(),
		AcceptWhileFiltering: key.NewBinding(),

		// Toggle help.
		ShowFullHelp:  key.NewBinding(),
		CloseFullHelp: key.NewBinding(),

		// Quitting.
		Quit:      key.NewBinding(),
		ForceQuit: key.NewBinding(),
	}
}
