package main

import "testing"

const TESTDIR = "./test_asm/bin/"

// TestStor tests the STOR instruction
func TestStor(t *testing.T) {
	chip := new(Chip8)
	chip.Init()
	chip.LoadProgram("./test_asm/bin/test_stor.bin")
	for i := 0; i < 3; i++ {
		chip.Execute()
	}
	if chip.memory[0xA] != 0xAB {
		t.Errorf("Got %#x, expected 0xAB", chip.memory[0xA])
	}
}

func TestRead(t *testing.T) {
	chip := NewChip(TESTDIR + "test_read.bin")
	for i := 0; i < 4; i++ {
		chip.Execute()
	}
	if chip.v[2] != 0xAB {
		t.Errorf("Got %#x, expected 0xAB", chip.v[2])
	}
}
