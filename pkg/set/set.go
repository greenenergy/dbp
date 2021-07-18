package set

import "fmt"

// --------------------------

type Set struct {
	init     bool
	contents map[string]bool
}

func NewSet() *Set {
	return &Set{
		init:     true,
		contents: make(map[string]bool),
	}
}

func (s *Set) Contains(key string) bool {
	if !s.init {
		s.contents = make(map[string]bool)
		s.init = true
	}
	_, ok := s.contents[key]
	return ok
}

func (s *Set) Add(keys ...string) error {
	if !s.init {
		s.contents = make(map[string]bool)
		s.init = true
	}
	for _, key := range keys {
		if _, ok := s.contents[key]; ok {
			return fmt.Errorf("key exists")
		}
		s.contents[key] = true
	}
	return nil
}

func (s *Set) Union(s2 *Set) error {
	for key := range s2.contents {
		s.Add(key)
	}
	return nil
}
