package main

import (
	"flag"
	"fmt"
	"math/bits"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

const progStart = 0x200
const memSize = 4096
const FONTSET_SIZE = 80
const FONT_OFFSET = 0x50

var fontSet = [FONTSET_SIZE]uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

// Chip8 is our emulated processor state
type Chip8 struct {
	inst       uint16
	memory     []uint8
	v          [16]uint8 // register block
	index      uint16    // index reg
	pc         uint16    // program counter
	gfx        []uint8   // pixel array for graphics
	delayTimer uint8
	soundTimer uint8
	stack      [16]uint16
	sp         uint16
}

/*
0x000-0x1FF - Chip 8 interpreter (contains font set in emu)
0x050-0x0A0 - Used for the built in 4x5 pixel font set (0-F)
0x200-0xFFF - Program ROM and work RAM
*/

// LoadProgram loads the program from a file into the Chip8's memory.
func (c *Chip8) LoadProgram(prog string) {
	progFile, err := os.Open(prog)
	if err != nil {
		panic(err)
	}
	_, err = progFile.Read(c.memory[progStart:])
	if err != nil {
		panic(err)
	}
}

// Init initializes the chip8 instance.
func (c *Chip8) Init() {
	c.inst = 0
	c.index = 0
	c.pc = progStart
	c.sp = 0
	c.delayTimer = 0
	c.soundTimer = 0
	c.memory = make([]uint8, memSize)
	c.gfx = make([]uint8, 64*32)
	for i, d := range fontSet {
		c.memory[FONT_OFFSET+i] = d
	}
}

// NewChip creates a new Chip8 instance loaded with the binary passed in
func NewChip(bin string) *Chip8 {
	c := new(Chip8)
	c.Init()
	c.LoadProgram(bin)
	return c
}

// Decode decodes a single instruction.
func (c *Chip8) Decode() {
	topByte := bits.RotateLeft16(uint16(c.memory[c.pc]), 8) // shift the top byte up 8
	bottomByte := uint16(c.memory[c.pc+1])
	c.inst = topByte | bottomByte
}

// ToString prints out the chip's state: index, pc, sp, and reg block
func (c *Chip8) ToString() string {
	return fmt.Sprintf("Chip State:\n\tinst: %#x\n\tindex: %#x\n\tpc: %#x\n\tsp: %d\n\tregs: %+v\n", c.inst, c.index, c.pc, c.sp, c.v)
}

func topNibble(i uint16) uint16 {
	return (i & 0xF000) >> 12
}
func bottomNibble(i uint16) uint16 {
	return (i & 0x000F)
}

func bottomByte(i uint16) uint16 {
	return i & 0x00FF
}

func targetAddr(i uint16) uint16 {
	return (i & 0x0FFF)
}

// SetIndex sets the index register if current inst is ANNN
func (c *Chip8) SetIndex() {
	c.index = c.inst & 0x0FFF
}

// SetPC sets the PC register to the given address
func (c *Chip8) SetPC(newaddr uint16) {
	c.pc = newaddr
}

// IncPC increments the PC (adds 2 since the word is a short)
func (c *Chip8) IncPC() {
	c.pc += 2
}

// GetImm pulls out the immediate value from the current instruction.
// numDigs is the number of hex digits to extract from the instruction.
func (c *Chip8) GetImm(numDigs int) uint8 {
	switch numDigs {
	case 1:
		return uint8(c.inst & 0x000F)
	case 2:
		return uint8(c.inst & 0x00FF)
	case 3:
		return uint8(c.inst & 0x0FFF)
	default:
		panic("bad arg")
	}
}

func (c *Chip8) GetXReg() uint16 {
	return c.inst & 0x0F00 >> 8
}

func (c *Chip8) GetYReg() uint16 {
	return c.inst & 0x00F0 >> 4
}

// Math8 executes the correct math instruction based on the bottom nibble of an inst starting with 0x8.
func (c *Chip8) Math8() {
	x := c.GetXReg()
	y := c.GetYReg()
	xVal := c.v[x]
	yVal := c.v[y]
	switch bottomNibble(c.inst) {
	case 0x0:
		c.v[x] = yVal
	case 0x1:
		c.v[x] = xVal | yVal
	case 0x2:
		c.v[x] = xVal & yVal
	case 0x3:
		c.v[x] = xVal ^ yVal
	case 0x4:
		add := xVal + yVal
		if add > 255 {
			c.v[0xF] = 1
		}
		c.v[x] = uint8(add & 0xFF)
	case 0x5:
		c.v[x] = xVal - yVal
	case 0x6:
		c.v[x] = xVal >> 1
	case 0x7:
		c.v[x] = yVal - xVal
	case 0xE:
		c.v[x] = xVal << 1
	}
}

// Execute executes a single instruction.
func (c *Chip8) Execute() {
	c.Decode()
	if c.inst == 0x0 {
		return
	}
	fmt.Println(c.ToString())
	top := topNibble(c.inst)
	x := c.GetXReg()
	y := c.GetYReg()
	switch top {
	case 0x0:
		switch bottomNibble(c.inst) {
		// CLR
		case 0x0:
			fmt.Println("clear screen")
			clear(c.gfx)
		// RET
		case 0xE:
			fmt.Println("ret")
		}
		c.IncPC()
	// JUMP
	case 0x1:
		c.SetPC(targetAddr(c.inst))
	// CALL
	case 0x2:
		c.SetPC(targetAddr(c.inst))
	// SKE
	case 0x3:
		imm := c.GetImm(2)
		c.IncPC()
		if imm == c.v[x] {
			c.IncPC() // skip inst
		}
	// SKNE
	case 0x4:
		imm := c.GetImm(2)
		c.IncPC()
		if imm != c.v[x] {
			c.IncPC()
		}
	// SKRE
	case 0x5:
		c.IncPC()
		if c.v[x] == c.v[y] {
			c.IncPC()
		}
	// LOAD
	case 0x6:
		imm := c.GetImm(2)
		c.v[x] = imm
		c.IncPC()
	// ADD
	case 0x7:
		imm := c.GetImm(2)
		c.v[x] += imm
		c.IncPC()
	// OR | AND | XOR | ADDR | SUB | SHR | SHL
	case 0x8:
		c.Math8()
		c.IncPC()
	// SKNRE
	case 0x9:
		c.IncPC()
		if c.v[x] != c.v[y] {
			c.IncPC()
		}
	// LOADI
	case 0xA:
		c.SetIndex()
		c.IncPC()
	case 0xD:
		// Get address in I
		// memory[I:I+n] -> gfx[x+y]
		x := c.v[c.GetXReg()]
		y := c.v[c.GetYReg()]
		n := c.GetImm(1)
		spriteAddr := c.index
		spriteLength := n
		var j uint8
		for i := spriteAddr; i < uint16(spriteLength+uint8(spriteAddr)); i++ {
			c.gfx[64*x+y+j] = c.memory[i]
			j++
		}
		c.IncPC()
	case 0xF:
		bottom := bottomByte(c.inst)
		switch bottom {
		// STOR
		case 0x55:
			c.memory[c.index] = c.v[x]
			c.IncPC()
		// READ
		case 0x65:
			c.v[x] = c.memory[c.index]
			c.IncPC()
		}
	}

	if c.delayTimer > 0 {
		c.delayTimer--
	}

	if c.soundTimer > 0 {
		c.soundTimer--
	}

}

func main() {
	chip := new(Chip8)
	chip.Init()
	var file = flag.String("file", "", "file to run")
	flag.Parse()
	chip.LoadProgram(*file)

	// for {
	// 	chip.Execute()
	// }

	// Set up window and canvas
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		1000, 1000, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}
	surface.FillRect(nil, 0)

	// offset := 0x55
	// chip.drawLetter(surface, window, offset, 40, 50)

	running := true
	// for i := 0; i < FONT_OFFSET+FONTSET_SIZE; i++ {
	// 	chip.gfx[i] = chip.memory[FONT_OFFSET+i]
	// }
	for running {
		chip.Execute()
		chip.drawMemory(surface, window)
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
			}
		}
	}
}

func (c *Chip8) drawMemory(surface *sdl.Surface, window *sdl.Window) {
	for i := 0; i < len(c.gfx); i++ {
		data := bits.Reverse8(c.gfx[i])
		for j := 0; j < 8; j++ {
			pixelVal := (data & (1 << j)) >> j
			rect := sdl.Rect{X: int32((j * 10)), Y: int32((i * 10)), W: 10, H: 10}
			color := sdl.Color{}
			if pixelVal == 1 {
				color = sdl.Color{R: 255, G: 255, B: 255, A: 255}
			} else {
				color = sdl.Color{R: 0, G: 0, B: 0, A: 0}
			}
			pixel := sdl.MapRGBA(surface.Format, color.R, color.G, color.B, color.A)

			surface.FillRect(&rect, pixel)
		}
	}
	window.UpdateSurface()
}

func (c *Chip8) drawLetter(surface *sdl.Surface, window *sdl.Window, offset, x, y int) {
	for i := 0; i < 5; i++ {
		data := bits.Reverse8(c.memory[offset+i])
		for j := 0; j < 8; j++ {
			pixelVal := (data & (1 << j)) >> j
			rect := sdl.Rect{X: int32((j * 10) + x), Y: int32((i * 10) + y), W: 10, H: 10}
			color := sdl.Color{}
			if pixelVal == 1 {
				color = sdl.Color{R: 255, G: 255, B: 255, A: 255}
			} else {
				color = sdl.Color{R: 0, G: 0, B: 0, A: 0}
			}

			pixel := sdl.MapRGBA(surface.Format, color.R, color.G, color.B, color.A)

			surface.FillRect(&rect, pixel)
		}
	}
	window.UpdateSurface()
}
