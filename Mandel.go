package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/cmplx"
	"os"
	"runtime"
	"sync"
)

var (
	top_left     = complex(-2, 1)
	bottom_right = complex(2, -1)
	image_width  = 1000
	image_height = 1000
	max_iter     = 1000
	sync_image   = &SyncImage{lock: &sync.Mutex{},
		image: image.NewRGBA(image.Rectangle{
			image.Point{0, 0},
			image.Point{image_width, image_height},
		})}
	wg sync.WaitGroup
	q  = make(chan int, 2000)
)

type ComplexPoint struct {
	z          complex128
	iterations int
}

type SyncImage struct {
	lock  *sync.Mutex
	image *image.RGBA
}

func (l SyncImage) Set(x, y int, color color.Color) {
	l.lock.Lock()
	l.image.Set(x, y, color)
	l.lock.Unlock()
}

/*func ___linspace(start, end float64, num int) []float64 {
	r := make([]float64, num, num)
	step := (end - start) / float64(num-1)
	for i := 0; i < num; i++ {
		r[i] = start + (step * float64(i))
	}
	return r
}*/

func linspace(start, end float64, num int, i int) float64 {
	step := (end - start) / float64(num-1)
	return start + (step * float64(i))
}

func mandelbrot(point *ComplexPoint) *ComplexPoint {
	c, zz := point.z, point.z
	for iter := 1; ; iter++ {
		zz = zz*zz + c
		point.iterations = iter
		if cmplx.Abs(zz) > 20 {
			return point
		}
		if iter == max_iter {
			return point
		}
	}
}

func renderline() {

	h := 0
	for {
		if len(q) == 0 {
			break
		}
		h = <-q

		for w := 0; w < image_width; w++ {

			_real := linspace(real(top_left), real(bottom_right),
				image_width, w)
			_imag := linspace(imag(top_left), imag(bottom_right),
				image_height, h)

			z := complex(_real, _imag)
			point := mandelbrot(&ComplexPoint{z: z, iterations: 0})
			co := color.RGBA{0, 0, 0, 0}
			if point.iterations == max_iter {
				co = color.RGBA{255, 255, 255, 255}
			} else {
				co = color.RGBA{0, 0, 0, 255}
			}
			sync_image.Set(w, h, co)
		}

	}
	wg.Done()
}

func main() {
	fmt.Printf("runtime.GOMAXPROCS:%v\n", runtime.GOMAXPROCS(0))
	fmt.Println("der haex kann das mandeln nicht lassen...")

	for h := 0; h < image_height; h++ {
		q <- h
	}

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go renderline()
	}
	wg.Wait()
	out_file, _ := os.Create("mandel.png")
	png.Encode(out_file, sync_image.image)
}
