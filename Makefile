VERSION=$(shell git rev-parse --short HEAD)

all: build run video play

build: 
	go build -ldflags "-X main.version=${VERSION}"

run:
	./mandelgo

video:
	ffmpeg -y -i mandel-%03d.png -framerate 5 mandel.mp4

play:
	mpv -loop=inf mandel.mp4

clean:
	rm -f mandel-*.png
	rm mandelgo

.PHONY: video play all clean
