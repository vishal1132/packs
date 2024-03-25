PROJECT_NAME := "packs"

alias b:=build
alias r:=run
alias i:=image

build:
    @echo "Building..."
    @go build -o packs .

run: build
    @echo "Running..."
    @ENV=local ./packs

image:
    @echo "building image"
    @docker build -t packs -f Dockerfile .