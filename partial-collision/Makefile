# Define the name of the binary
BINARY := partial-collision

# Define the source directory for the main package
SRCDIR := ./src

# Define the build target
build:
	@cd $(SRCDIR) && go build -o ../$(BINARY)

# Define the run target
run:
	@./$(BINARY)

# Define the clean target
clean:
	rm -rf $(BINARY)

# Define the default target
.PHONY: all