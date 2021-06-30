package main

import "testing"

func TestPageTableCreation(t *testing.T) {
	page_table_ := PageTable{}

	for range [8]int{} {
		page_table_.CreatePageEntry(false, false, false, 0, 0x000)
	}

	// Make sure were allocating 8 entries and all values false
	t.Run("Ensure correct allocation size", func(t *testing.T) {
		if len(page_table_.entries_) != 8 {
			t.Errorf("Wanted table size: 8 Got table size: %d\n", len(page_table_.entries_))
		}
	})

	t.Run("Ensure correct allocation allocation", func(t *testing.T) {
		for _, e := range page_table_.entries_ {
			if e.mod_ && e.present_ && e.ref_ != false && e.right_ != 0 {
				t.Errorf("Table not being intialized to false")
			}
		}
	})
}

func TestFileLoading(t *testing.T) {

	t.Run("Make sure correct amount loaded", func(t *testing.T) {
		refs := References{}

		var load_amn uint32 = 5

		refs.LoadReferences("testfileloading.txt", load_amn)
		if uint32(len(refs.refs_)) != load_amn {
			t.Errorf("Refs size should be %v, instead its %v\n", load_amn, len(refs.refs_))
		}

		load_amn = 8

		refs.LoadReferences("testfileloading.txt", load_amn)
		if uint32(len(refs.refs_)) != load_amn {
			t.Errorf("Refs size should be %v, instead its %v\n", load_amn, len(refs.refs_))
		}

		load_amn = 0

		refs.LoadReferences("testfileloading.txt", load_amn)
		if uint32(len(refs.refs_)) != load_amn {
			t.Errorf("Refs size should be %v, instead its %v\n", load_amn, len(refs.refs_))
		}
	})
}
