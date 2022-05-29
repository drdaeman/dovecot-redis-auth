package main

import "sort"

type KeyValuePair struct {
	Key   string
	Value string
}

type By func(p1, p2 *KeyValuePair) bool

func ByKey(p1, p2 *KeyValuePair) bool {
	return p1.Key < p2.Key
}

func ByValue(p1, p2 *KeyValuePair) bool {
	if p1.Value == p2.Value {
		return p1.Key < p2.Key
	}
	return p1.Value < p2.Value
}

func (by By) Sort(data []*KeyValuePair) {
	sorter := &kvSorter{
		data: data,
		by:   by,
	}
	sort.Sort(sorter)
}

type kvSorter struct {
	data []*KeyValuePair
	by   func(p1, p2 *KeyValuePair) bool
}

func (s *kvSorter) Len() int {
	return len(s.data)
}

func (s *kvSorter) Swap(i, j int) {
	s.data[i], s.data[j] = s.data[j], s.data[i]
}

func (s *kvSorter) Less(i, j int) bool {
	return s.by(s.data[i], s.data[j])
}
