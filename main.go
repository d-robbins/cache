package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

/****************************************************************
 	* Virtual address:   3 page bits 		---- 13 offset bits
	* Physical address:  11 page frame bits ---- 13 offset bits
*****************************************************************/

const PAGE_SIZE int = 8

type PageEntry struct {
	present_ bool
	ref_     bool
	mod_     bool
	right_   uint8
	frame_   uint16
}

type Reference struct {
	address_ uint16
	op_      string
}

type References struct {
	refs_ []Reference
}

type PageTable struct {
	entries_ []PageEntry
}

type Process struct {
	table_       PageTable
	free_frames_ []uint16
	used_frames_ []uint16
	refs_        References
	reads_       uint16
	writes_      uint16
}

func (p *Process) ProcessReferences() {
	for _, reference := range p.refs_.refs_ {
		page, offset := reference.DecompVirtualAddress()
		if page >= uint8(PAGE_SIZE) {
			fmt.Printf("Reference %.4X accessing invalid page: %d", reference.address_, page)
		} else {
			present_ready := p.table_.entries_[page].present_
			for !present_ready {
				// LRU
				if len(p.free_frames_) == 0 {
					// Pick a frame to evict
					for _, curr_page := range p.table_.entries_ {
						if curr_page.present_ {
							// Sorry bud youre the one
							if curr_page.mod_ {
								fmt.Println("WRITE BACK")
							}
							p.free_frames_ = append(p.free_frames_, curr_page.frame_)

							to_remove := 0
							for i, find_frame := range p.used_frames_ {
								if find_frame == curr_page.frame_ {
									to_remove = i
								}
							}

							if to_remove == 0 {
								p.used_frames_ = p.used_frames_[1:]
							} else if to_remove == len(p.used_frames_)-1 {
								p.used_frames_ = p.used_frames_[0 : len(p.used_frames_)-1]
							} else {
								lhs := p.used_frames_[0:(to_remove - 1)]
								rhs := p.used_frames_[(to_remove + 1):]
								for _, x := range rhs {
									lhs = append(lhs, x)
								}
								p.used_frames_ = lhs
							}

							fmt.Println("HELLO")

							break
						}
					}

				}

				if len(p.free_frames_) != 0 {
					p.table_.entries_[page].present_ = true
					p.table_.entries_[page].ref_ = false
					p.table_.entries_[page].mod_ = false

					p.table_.entries_[page].frame_ = p.free_frames_[0]
					p.used_frames_ = append(p.used_frames_, p.free_frames_[0])

					to_remove := 0
					for i, find_frame := range p.free_frames_ {
						if find_frame == p.table_.entries_[page].frame_ {
							to_remove = i
						}
					}

					if to_remove == 0 {
						p.free_frames_ = p.free_frames_[1:]
					} else if to_remove == len(p.free_frames_)-1 {
						p.free_frames_ = p.free_frames_[0 : len(p.free_frames_)-1]
					} else {
						lhs := p.free_frames_[0:(to_remove - 1)]
						rhs := p.free_frames_[(to_remove + 1):]
						for _, x := range rhs {
							lhs = append(lhs, x)
						}
						p.free_frames_ = lhs
					}
				} else {
					fmt.Println("ERROR IN LRU")
				}

				present_ready = p.table_.entries_[page].present_
			}

			if reference.op_ == "R" {
				p.table_.entries_[page].ref_ = true
			} else if reference.op_ == "W" {
				p.table_.entries_[page].mod_ = true
				p.table_.entries_[page].ref_ = true
			}

			physical := (uint32(p.table_.entries_[page].frame_) << 13) | uint32(offset)
			fmt.Printf("%.4X %s %.1X    %.6X\n", reference.address_, reference.op_, page, physical)
		}
	}
}

// Create a process from a file
// file begins with max page frame allocation
//      from there on the file will list page
//      number and access rights
func CreateProcess(file string) *Process {
	data, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	defer data.Close()

	new_proc := new(Process)

	// allocate pages for this process
	for range [PAGE_SIZE]int{} {
		new_proc.table_.CreatePageEntry(false, false, false, 0, 0x000)
	}

	scanner := bufio.NewScanner(data)
	i := 0
	for scanner.Scan() {
		words := strings.Split(scanner.Text(), " ")
		if i == 0 {
			// Tell process what pages it's allowed to use
			to_allocate, _ := strconv.Atoi(words[0])
			for index := 0; index < to_allocate; index++ {
				new_proc.free_frames_ = append(new_proc.free_frames_, 0x400+uint16(index))
			}
			i++
		} else if len(words) != 2 && i != 0 {
			fmt.Println("Invalid process page data")
		} else {
			page, _ := strconv.Atoi(words[0])
			rights, _ := strconv.Atoi(words[1])

			if page >= PAGE_SIZE {
				fmt.Printf("Invalid page size: %d, max page value: %d\n", page, PAGE_SIZE)
			}

			new_proc.table_.entries_[page].right_ = uint8(rights)
		}
	}

	return new_proc
}

// Decompose references page and offset bits
func (r *Reference) DecompVirtualAddress() (page uint8, off uint16) {
	page = uint8((r.address_ & 0xE000) >> 13)
	off = r.address_ & 0x1FFF
	return page, off
}

// Print the references
func (r *References) PrintReferences() {
	for _, ref := range r.refs_ {
		page, off := ref.DecompVirtualAddress()
		fmt.Printf("%.4X %s %.1X %.4x\n", ref.address_, ref.op_, page, off)
	}
}

// Load memory references and operations
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
	process := CreateProcess("page")
	process.refs_.LoadReferences("refs", 10)

	process.ProcessReferences()

	process.table_.PrintTable()
	process.refs_.PrintReferences()
}
