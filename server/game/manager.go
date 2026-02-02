package game

type Manager struct {
	Title  string              // name of the game; key into trivia/*.json
	Code   string              // unique game code, 6 uppercase letters/numbers
	Board  map[string]*Player  // category item -> player who claimed it (nil if unclaimed)
	Colors map[string]struct{} // set of assigned colors
	Time   int                 // seconds remaining (60 until start, then 180)
}

// NewManager creates a Manager with the given title and code. Time is set to 60,
// board and colors are initialized empty.
func NewManager(title, code string) *Manager {
	return &Manager{
		Title:  title,
		Code:   code,
		Board:  make(map[string]*Player),
		Colors: make(map[string]struct{}),
		Time:   60,
	}
}
