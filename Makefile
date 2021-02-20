TESTDIR=./test_asm
build:
	go build -o chip8 main.go
test: $(TESTDIR)/*.asm
	$(foreach file, $(wildcard $(TESTDIR)/*.asm), @c8asm -i $(file) -o $(TESTDIR)/bin/$(basename $(notdir $(file))).bin > /dev/null;)
	go test