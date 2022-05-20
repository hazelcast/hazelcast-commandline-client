package prompt

// History stores the texts that are entered.
type History struct {
	Commands         []string
	editableCommands []string
	selected         int
}

// Add to add text in History.
func (h *History) Add(input string) {
	h.Commands = append(h.Commands, input)
	h.Clear()
}

// Clear to clear the History.
func (h *History) Clear() {
	h.editableCommands = make([]string, len(h.Commands))
	for i := range h.Commands {
		h.editableCommands[i] = h.Commands[i]
	}
	h.editableCommands = append(h.editableCommands, "")
	h.selected = len(h.editableCommands) - 1
}

// Older saves a buffer of current line and get a buffer of previous line by up-arrow.
// The changes of line buffers are stored until new History is created.
func (h *History) Older(buf *Buffer) (new *Buffer, changed bool) {
	if len(h.editableCommands) == 1 || h.selected == 0 {
		return buf, false
	}
	h.editableCommands[h.selected] = buf.Text()

	h.selected--
	new = NewBuffer()
	new.InsertText(h.editableCommands[h.selected], false, true)
	return new, true
}

// Newer saves a buffer of current line and get a buffer of next line by up-arrow.
// The changes of line buffers are stored until new History is created.
func (h *History) Newer(buf *Buffer) (new *Buffer, changed bool) {
	if h.selected >= len(h.editableCommands)-1 {
		return buf, false
	}
	h.editableCommands[h.selected] = buf.Text()

	h.selected++
	new = NewBuffer()
	new.InsertText(h.editableCommands[h.selected], false, true)
	return new, true
}

// NewHistory returns new History object.
func NewHistory() *History {
	return &History{
		Commands:         []string{},
		editableCommands: []string{""},
		selected:         0,
	}
}
