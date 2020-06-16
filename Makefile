
# I've found that a good make file can be super helpful for go projects

GOCMD=go 
GOBUILD=$(GOCMD) build 
GOTEST=$(GOCMD) test 


# build entry points

build: api task

# regression testing
test:
	$(GOTEST) ./...

# sub-sections

api:
	@echo "building api..."
	@$(GOBUILD) -o $(GOPATH)/bin/api ./cmd/api/

task:
	@echo "building task..."
	@$(GOBUILD) -o $(GOPATH)/bin/task ./cmd/task/
