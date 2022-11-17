package indi

import "fmt"

type Slice []string

func New(a ...string) Slice {
	return Slice(a)
}

func (s *Slice) Push(v string) {
	*s = append(*s, v)
}

func (s *Slice) Update(v string) {
	*s = append(*s, v)
}

func (s *Slice) Pop(i int64) (v string) {
	v = (*s)[i]
	*s = append((*s)[:i], (*s)[i+1:]...)
	return v
}

func (s Slice) Tail(size int) Slice {
	length := len(s)
	if length <= size {
		win := make(Slice, length)
		copy(win, s)
		return win
	}

	win := make(Slice, size)
	copy(win, s[length-size:])
	return win
}
func (s *Slice) Last() string {
	length := len(*s)
	if length > 0 {
		return (*s)[length-1]
	}
	return ""
}

func (s *Slice) Index(i int) string {
	length := len(*s)
	if length-i <= 0 || i < 0 {
		return ""
	}
	return (*s)[length-i-1]
}

func (s *Slice) Length() int {
	return len(*s)
}
func (s Slice) Change() bool {
	fmt.Println(s.Last(), s.Index(1))
	return s.Last() != s.Index(1)
}
