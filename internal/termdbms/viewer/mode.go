package viewer

func MoveCursorWithinBounds(m *TuiModel) {
	defer func() {
		if recover() != nil {
			// todo this should not happen, and if does we should log it
		}
	}()
	offset := GetOffsetForLineNumber(m.Format.CursorY)
	l := len(*m.Format.EditSlices[m.Format.CursorY])
	end := l - 1 - offset
	if m.Format.CursorX > end {
		m.Format.CursorX = end
	}
}
