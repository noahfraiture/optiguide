package auth

type UserProgress struct {
	UserID string
	Steps  map[int]bool
}

var Progress map[string]UserProgress = make(map[string]UserProgress)
