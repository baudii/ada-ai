package nav

type store struct {
	items map[string]FileInfo
}

type FileInfo struct {
	Filename string
	Content  []byte
}

func New() *store {
	return &store{
		items: make(map[string]FileInfo),
	}
}

func (s *store) Add(k, p string, c []byte) {
	s.items[k] = FileInfo{
		Filename: p,
		Content:  c,
	}
}

func (s *store) Get(k string) (*FileInfo, bool) {
	f, ok := s.items[k]
	return &f, ok
}

func (s *store) GetItems() map[string]FileInfo {
	return s.items
}
