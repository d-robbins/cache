package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

/********
 * Virtual address:   13 offset bits --- 3 page bits
 * Physical address:  13 offset bits --- 11 page frame bits
*********/

type PageEntry struct {
	present_ bool
	ref_     bool
	mod_     bool
	right_   uint8
	frame_   uint8
}

type Reference struct {
	address_ uint16
	op_      string
}

type References struct {
	refs_ []Reference
}

func (r *References) LoadReferences(file string, n uint32) {
	// Clear old references
	r.refs_ = []Reference{}
	data, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	defer data.Close()

	scanner := bufio.NewScanner(data)

	for scanner.Scan() {
		if n > 0 {
			words := strings.Split(scanner.Text(), " ")
			if len(words) == 2 {
				i, _ := strconv.ParseInt(words[0], 16, 64)
				r.refs_ = append(r.refs_, Reference{address_: uint16(i), op_: words[1]})
			}

			n--
		} else {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

}

type PageTable struct {
	entries_ []PageEntry
}

/// Print a page table
func (pt *PageTable) PrintTable() {
	fmt.Printf("     A  P  R  M FRM\n")
	for j, i := range pt.entries_ {
		fmt.Printf("[%v]: %v  %v  %v  %v %.3X\n", j, i.right_, b2i(i.present_), b2i(i.ref_), b2i(i.mod_), i.frame_)
	}
}

/// Create a pagetable entry and add it to the table
/// Needs present bit, reference bit, modified bit, and 3 access right bits
func (pt *PageTable) CreatePageEntry(p bool, ref bool, mod bool, rights uint8, frame uint8) {
	pt.entries_ = append(pt.entries_, PageEntry{
		present_: p,
		ref_:     ref,
		mod_:     mod,
		right_:   rights,
	})
}

/// Bool to integer
func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func main() {
	p := PageTable{}
	for _, entry := range p.entries_ {
		fmt.Println(entry)
	}

	for range [8]int{} {
		p.CreatePageEntry(false, false, false, 0, 0x000)
	}

	p.PrintTable()

	refs := References{}
	refs.LoadReferences("testfileloading.txt", 10)

	for _, y := range refs.refs_ {
		fmt.Printf("%.4X %s\n", y.address_, y.op_)
	}
}
