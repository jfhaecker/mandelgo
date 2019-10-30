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
	"time"
)

var (
	imageWidth   = 1000
	imageHeight  = 1000
	maxIterStart = 1000
	syncImage    = &SyncImage{
		image: image.NewRGBA(image.Rectangle{
			image.Point{0, 0},
			image.Point{imageWidth, imageHeight},
		})}
	bailoutRadius  float64 = 2000
	maxWorkerCount         = runtime.GOMAXPROCS(0)
	imageCount     int     = 500

	// Interesting Points/Coordinates:
	// http://www.cuug.ab.ca/dewara/mandelbrot/Mandelbrowser.html
	rectangle = &ComplexRectangle{center: complex(-0.235125, 0.827215),
		//	rectangle = &ComplexRectangle{center: borgsHome,
		//	r = &ComplexRectangle{center: complex(-0.74529, 0.113075),
		width:  0.1,
		height: 0.1}
)

type ComplexRectangle struct {
	topLeft     complex128
	bottomRight complex128
	center      complex128
	width       float64
	height      float64
}

func (r *ComplexRectangle) Set(center complex128, width, height float64) {
	r.center = center
	r.width = width
	r.height = height
}

func (r *ComplexRectangle) Scale(factor float64) {
	r.width -= (r.width * factor)
	r.height -= (r.height * factor)
	r.calc()
}

func (r *ComplexRectangle) calc() {
	r.topLeft = complex(real(r.center)-r.width/2,
		imag(r.center)+r.height/2)
	r.bottomRight = complex(real(r.center)+r.width/2,
		imag(r.center)-r.height/2)
}

type ComplexPoint struct {
	z          complex128
	iterations int
}

type SyncImage struct {
	sync.Mutex
	image *image.RGBA
}

func (l *SyncImage) Set(x, y int, color color.Color) {
	l.Lock()
	l.image.Set(x, y, color)
	l.Unlock()
}

func linspace(start, end float64, num int, i int) float64 {
	step := (end - start) / float64(num-1)
	return start + (step * float64(i))
}

func mandelbrot(point *ComplexPoint, maxIter int) *ComplexPoint {
	c, zz := point.z, point.z
	for iter := 1; ; iter++ {
		zz = zz*zz + c
		point.iterations = iter

		// TODO: float mu = iter_count - (log (log (modulus)))/ log (2.0);
		// https://linas.org/art-gallery/escape/escape.html

		if cmplx.Abs(zz) > bailoutRadius {
			return point
		}
		if iter == maxIter {
			return point
		}
	}
}

func getColor(index int) color.RGBA {
	//qu := quake[97 : 97+16]
	qu := quake[0:255]
	//	return qu[index%(len(qu)-1)]
	return qu[(index)%len(qu)]
}

func renderline(jobs <-chan int, wg *sync.WaitGroup, maxIter int) {
	defer wg.Done()
	for line := range jobs {

		for w := 0; w < imageWidth; w++ {

			real := linspace(real(rectangle.topLeft),
				real(rectangle.bottomRight),
				imageWidth, w)
			imag := linspace(imag(rectangle.topLeft),
				imag(rectangle.bottomRight),
				imageHeight, line)

			z := complex(real, imag)
			point := mandelbrot(&ComplexPoint{z: z}, maxIter)
			co := color.RGBA{0, 0, 0, 0}
			if point.iterations == maxIter {
				co = color.RGBA{255, 255, 255, 255}
			} else {
				co = getColor(point.iterations)
			}
			syncImage.Set(w, line, co)
		}
	}
}

func main() {
	fmt.Println("der haex kann das mandeln nicht lassen...")
	fmt.Printf("Using %v workers for %v images\n", maxWorkerCount, imageCount)

	maxIter := maxIterStart
	var wg sync.WaitGroup

	for x := 0; x < imageCount; x++ {
		workerQueue := make(chan int, imageHeight)
		t1 := time.Now()
		rectangle.Scale(0.03)
		fname := fmt.Sprintf("mandel-%03v.png", x)
		/*fmt.Printf("[%v|%v|%v|%v|] -> %v\n", maxIter,
		rectangle.center, rectangle.height,
		rectangle.width, fname)*/

		for i := 0; i < maxWorkerCount; i++ {
			wg.Add(1)
			go renderline(workerQueue, &wg, maxIter)
		}

		for h := 0; h < imageHeight; h++ {
			workerQueue <- h
		}
		close(workerQueue)

		wg.Wait()
		outFile, _ := os.Create(fname)
		png.Encode(outFile, syncImage.image)

		fmt.Printf("%v took %v\n", fname, time.Since(t1))
		/*if x%100 == 0 {
			maxIter *= 10
		}*/
	}

}
