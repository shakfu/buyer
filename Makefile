
HOMEBREW := $(shell brew --prefix)
LIB := $(HOMEBREW)/lib
INCLUDE := $(HOMEBREW)/include


.PHONY: all build release format clean fixtures fixtures-sqlite fixtures-json

all: build

build:
	-@swift build -Xlinker -L$(LIB) -Xcc -I$(INCLUDE)


release:
	-@swift build -c release -Xlinker -L$(LIB) -Xcc -I$(INCLUDE)


format:
	-@swift-format --configuration .swiftformatrc --recursive --in-place ./src


clean:
	-@rm -rf .build

fixtures: fixtures-sqlite fixtures-json

fixtures-sqlite:
	@swift run buyer fixtures --store sqlite --output fixtures/procurement.sqlite3 --overwrite

fixtures-json:
	@swift run buyer fixtures --store json --output fixtures/procurement.json --overwrite
