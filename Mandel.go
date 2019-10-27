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
	//top_left     = complex(-2, 1)
	//bottom_right = complex(1, -1)
	image_width  = 1000
	image_height = 1000
	max_iter     = 1000
	sync_image   = &SyncImage{lock: &sync.Mutex{},
		image: image.NewRGBA(image.Rectangle{
			image.Point{0, 0},
			image.Point{image_width, image_height},
		})}
	wg               sync.WaitGroup
	bailout_radius   float64 = 200
	max_worker_count         = runtime.GOMAXPROCS(0)
	q                        = make(chan int, image_height)
	colorIndex       int
	image_count      int = 100
	r                    = &ComplexRectangle{center: complex(-0.235125, 0.827215),
		width:  2,
		height: 2}
)

type ComplexRectangle struct {
	top_left     complex128
	bottom_right complex128
	center       complex128
	width        float64
	height       float64
}

func (r ComplexRectangle) Set(center complex128, width, height float64) {
	r.center = center
	r.width = width
	r.height = height
	//	r.calc()
}

func (r *ComplexRectangle) Scale(factor float64) {
	r.width -= (r.width * factor)
	r.height -= (r.height * factor)
	r.Calc()
}

func (r *ComplexRectangle) Calc() {
	r.top_left = complex(real(r.center)-r.width/2, imag(r.center)+r.height/2)
	r.bottom_right = complex(real(r.center)+r.width/2, imag(r.center)-r.height/2)
}

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
		if cmplx.Abs(zz) > bailout_radius {
			return point
		}
		if iter == max_iter {
			return point
		}
	}
}

func getColor(index int) color.RGBA {
	//qu := quake[97 : 97+16]
	qu := quake[0:128]
	//	return qu[index%(len(qu)-1)]
	return qu[(index+colorIndex)%len(qu)]
}

func renderline() {
	h := 0
	for {
		if len(q) == 0 {
			break
		}
		h = <-q

		for w := 0; w < image_width; w++ {

			real := linspace(real(r.top_left), real(r.bottom_right),
				image_width, w)
			imag := linspace(imag(r.top_left), imag(r.bottom_right),
				image_height, h)

			z := complex(real, imag)
			point := mandelbrot(&ComplexPoint{z: z, iterations: 0})
			co := color.RGBA{0, 0, 0, 0}
			if point.iterations == max_iter {
				co = color.RGBA{255, 255, 255, 255}
			} else {
				co = getColor(point.iterations)
				//co = color.RGBA{139, 139, 130, 255}
			}
			sync_image.Set(w, h, co)
		}
	}
	wg.Done()
}

func main() {
	fmt.Println("der haex kann das mandeln nicht lassen...")
	fmt.Printf("%#v\n", r)
	r.Calc()
	fmt.Printf("%#v\n", r)

	for x := 0; x < image_count; x++ {
		fmt.Printf("%#v\n", r)
		r.Scale(0.1)
		r.Calc()
		fmt.Printf("%#v\n", r)
		fname := fmt.Sprintf("mandel-%03v.png", x)
		fmt.Println("Generating image:" + fname)
		for h := 0; h < image_height; h++ {
			q <- h
		}

		fmt.Printf("Starten %v workers for %v\n", max_worker_count, fname)
		for i := 0; i < max_worker_count; i++ {
			wg.Add(1)
			go renderline()
		}
		wg.Wait()

		out_file, _ := os.Create(fname)
		png.Encode(out_file, sync_image.image)
		colorIndex++
	}
}
