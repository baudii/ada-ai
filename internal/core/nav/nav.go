package nav

type store struct {
	items map[string]FileInfo
}

// FileInfo describes a stored navigation file and its content.
type FileInfo struct {
	Filename string
	Content  []byte
}

// New returns a new navigation store.
func New() *store {
	return &store{
		items: make(map[string]FileInfo),
	}
}

// Add stores a navigation item under key k with filename p and content c.
func (s *store) Add(k, p string, c []byte) {
	s.items[k] = FileInfo{
		Filename: p,
		Content:  c,
	}
}

// Get retrieves a stored navigation file info by filename.
// It returns the item and a boolean indicating whether it was found.
func (s *store) Get(k string) (*FileInfo, bool) {
	f, ok := s.items[k]
	return &f, ok
}

// GetItems returns all stored navigation file infos keyed by their names.
func (s *store) GetItems() map[string]FileInfo {
	return s.items
}
