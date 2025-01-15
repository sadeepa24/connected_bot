# Define the binary name and source files
BINARY_NAME=yourapp
SRC_FILES=.  # Adjust this if your Go files are in subdirectories

# Set the default target
all: build

# Build the application
build: 
	@echo "Building the application..."
	go build -o $(BINARY_NAME) $(SRC_FILES)

# Clean up the generated binaries
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)

# Run the application
run: build
	@echo "Running the application..."
	./$(BINARY_NAME)

.PHONY: all build clean run