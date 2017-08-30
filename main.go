package main

import (
	"flag"
	"os"
	"io/ioutil"
	"fmt"
	"path"
	"github.com/jroimartin/gocui"
	"sort"
	"github.com/olekukonko/tablewriter"
	"strconv"
	"github.com/dustin/go-humanize"
)

var (
	startPath = flag.String("path", ".", "path")
	st        *SizeTree
	cst       *SizeTree
	selected  int
)

func init() {
	selected = 0
}

type SizeTree struct {
	path, fullpath string
	entries        []*SizeTree
	size           uint64
	count          uint64
}

func (st *SizeTree) AddSubEntry(name string, entry *SizeTree) {
	if entry == nil {
		return
	}

	if st.entries == nil {
		st.entries = make([]*SizeTree, 0)
	}

	st.entries = append(st.entries, entry)
	st.size += entry.size
	st.count += entry.count
}

func (st *SizeTree) Length() int {
	return len(st.entries)
}

func (st *SizeTree) EntriesBySize() []*SizeTree {
	entries := make([]*SizeTree, st.Length())
	i := 0
	for _, e := range st.entries {
		entries[i] = e
		i++
	}

	sort.Sort(SizeTreeBySize(entries))
	return entries
}

func (st *SizeTree) String() string {
	return fmt.Sprintf("%s: %d files, %d bytes", st.fullpath, st.count, st.size)
}

func (st *SizeTree) CalculateSubentries() {
	stat, err := os.Stat(st.fullpath)
	if os.IsNotExist(err) {
		panic(err)
	}

	if stat.IsDir() {
		st.entries = make([]*SizeTree, 0)

		entries, err := ioutil.ReadDir(st.fullpath)
		if err != nil {
			panic(err)
		}

		for _, entry := range entries {
			entryPath := path.Join(st.fullpath, entry.Name())
			es := NewSizeTree(entryPath)
			es.CalculateSubentries()
			st.AddSubEntry(entry.Name(), es)
		}
	} else {
		st.size = uint64(stat.Size())
		st.count = 1
	}
}

func NewSizeTree(startPath string) *SizeTree {
	return &SizeTree{
		path:     path.Base(startPath),
		fullpath: startPath,
	}
}

func main() {
	flag.Parse()

	st = NewSizeTree(*startPath)
	st.CalculateSubentries()
	cst = st

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		panic(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, selectNext); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, selectPrevious); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, selectInner); err != nil {
		panic(err)
	}
	if err := g.SetKeybinding("", gocui.KeyBackspace2, gocui.ModNone, revert); err != nil {
		panic(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		panic(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("default", 0, 0, maxX-1, maxY-1); err != nil && err != gocui.ErrUnknownView {
		return err
	} else {
		v.Clear()
		fmt.Fprintf(v, "\t\t%s\n\n", cst.fullpath)

		i := 0
		table := tablewriter.NewWriter(v)
		table.SetHeader([]string{"", "Name", "Total size", "Files count", ""})
		table.SetBorder(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetCenterSeparator(" ")
		table.SetColumnSeparator(" ")
		table.SetRowSeparator(" ")

		for _, entry := range cst.EntriesBySize() {
			prefix := ""
			postfix := ""

			if i == selected {
				prefix = "\x1b[47;1m"
				postfix = "\x1b[0m"
			}

			path := entry.path
			size := humanize.Bytes(entry.size)
			count := strconv.FormatUint(entry.count, 10)

			table.Append([]string{prefix, path, size, count, postfix})
			i++
		}

		table.Render()
	}
	return nil
}

func selectNext(g *gocui.Gui, v *gocui.View) error {
	if selected >= cst.Length()-1 {
		selected = cst.Length() - 1
	} else {
		selected++
	}

	return nil
}

func selectPrevious(g *gocui.Gui, v *gocui.View) error {
	if selected <= 0 {
		selected = 0
	} else {
		selected--
	}

	return nil
}

func selectInner(g *gocui.Gui, v *gocui.View) error {
	cst = cst.EntriesBySize()[selected]
	selected = 0
	return nil
}

func revert(g *gocui.Gui, v *gocui.View) error {
	cst = st
	selected = 0
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
