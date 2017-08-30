package main

type SizeTreeBySize []*SizeTree

func (a SizeTreeBySize) Len() int {
	return len(a)
}

func (a SizeTreeBySize) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a SizeTreeBySize) Less(i, j int) bool {
	return a[i].size > a[j].size
}

type SizeTreeByCount []*SizeTree

func (a SizeTreeByCount) Len() int {
	return len(a)
}

func (a SizeTreeByCount) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a SizeTreeByCount) Less(i, j int) bool {
	return a[i].count > a[j].count
}