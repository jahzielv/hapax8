package main

import "testing"

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
