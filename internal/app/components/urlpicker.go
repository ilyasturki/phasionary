package components

type URLPickerState struct {
	URLs     []string
	Selected int
}

func NewURLPickerState(urls []string) URLPickerState {
	return URLPickerState{URLs: urls, Selected: 0}
}

func (u *URLPickerState) MoveUp() {
	if u.Selected > 0 {
		u.Selected--
	}
}

func (u *URLPickerState) MoveDown() {
	if u.Selected < len(u.URLs)-1 {
		u.Selected++
	}
}

func (u *URLPickerState) SelectedURL() string {
	if u.Selected < 0 || u.Selected >= len(u.URLs) {
		return ""
	}
	return u.URLs[u.Selected]
}
