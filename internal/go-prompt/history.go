package prompt

// History stores the texts that are entered.
type History struct {
	Histories []string
	tmp       []string
	selected  int
}

// Add to add text in History.
func (h *History) Add(input string) {
	h.Histories = append(h.Histories, input)
	h.Clear()
}

// Clear to clear the History.
func (h *History) Clear() {
	h.tmp = make([]string, len(h.Histories))
	for i := range h.Histories {
		h.tmp[i] = h.Histories[i]
	}
	h.tmp = append(h.tmp, "")
	h.selected = len(h.tmp) - 1
}

// Older saves a buffer of current line and get a buffer of previous line by up-arrow.
// The changes of line buffers are stored until new History is created.
func (h *History) Older(buf *Buffer) (new *Buffer, changed bool) {
	if len(h.tmp) == 1 || h.selected == 0 {
		return buf, false
	}
	h.tmp[h.selected] = buf.Text()

	h.selected--
	new = NewBuffer()
	new.InsertText(h.tmp[h.selected], false, true)
	return new, true
}

// Newer saves a buffer of current line and get a buffer of next line by up-arrow.
// The changes of line buffers are stored until new History is created.
func (h *History) Newer(buf *Buffer) (new *Buffer, changed bool) {
	if h.selected >= len(h.tmp)-1 {
		return buf, false
	}
	h.tmp[h.selected] = buf.Text()

	h.selected++
	new = NewBuffer()
	new.InsertText(h.tmp[h.selected], false, true)
	return new, true
}

// NewHistory returns new History object.
func NewHistory() *History {
	return &History{
		Histories: []string{},
		tmp:       []string{""},
		selected:  0,
	}
}
