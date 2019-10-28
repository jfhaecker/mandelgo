set -e
go build && ./mandelgo
ffmpeg -y -i mandel-%03d.png -framerate 5 mandel.mp4
mpv -loop=inf mandel.mp4
