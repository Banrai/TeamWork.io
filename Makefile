
SHELL = /bin/sh

# define the source/target folders
SRC_ROOT = $(shell pwd)
SERVER   = $(SRC_ROOT)/server

# Server binary
TeamWorkServer: $(SERVER)/main.go
	go build -o $(SERVER)/TeamWorkServer $^

# all components
all: TeamWorkServer

clean:
	rm -f $(addprefix $(SERVER)/, TeamWorkServer)
