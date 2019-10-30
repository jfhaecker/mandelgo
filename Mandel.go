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
	bailoutRadius        float64 = 2000
	maxMandelWorkerCount         = runtime.GOMAXPROCS(0)
	maxImageWorkerCount          = 1
	imageCount           int     = 500

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
	x, y       int
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

func (l *SyncImage) SetUnlocked(x, y int, color color.Color) {
	l.image.Set(x, y, color)
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
	qu := quake[0:255]
	return qu[(index)%len(qu)]
}

func renderImage(points <-chan *ComplexPoint, wg *sync.WaitGroup, fileName string, maxIter int) {
	defer wg.Done()
	for point := range points {
		co := color.RGBA{0, 0, 0, 0}
		if point.iterations == maxIter {
			co = color.RGBA{255, 255, 255, 255}
		} else {
			co = getColor(point.iterations)
		}
		syncImage.SetUnlocked(point.x, point.y, co)
	}
	outFile, _ := os.Create(fileName)
	png.Encode(outFile, syncImage.image)
}

func renderline(jobs <-chan int, result chan<- *ComplexPoint, wg *sync.WaitGroup, maxIter int) {
	//fmt.Printf("renderline %v\n", len(jobs))
	for y := range jobs {
		for x := 0; x < imageWidth; x++ {
			real := linspace(real(rectangle.topLeft),
				real(rectangle.bottomRight),
				imageWidth, x)
			imag := linspace(imag(rectangle.topLeft),
				imag(rectangle.bottomRight),
				imageHeight, y)

			z := complex(real, imag)
			point := mandelbrot(&ComplexPoint{z: z, x: x, y: y}, maxIter)
			result <- point
		}
	}
	defer wg.Done()
}

func main() {
	fmt.Println("der haex kann das mandeln nicht lassen...")
	fmt.Printf("Using %v mandelworkers and %v imageworkers for %v images\n", maxImageWorkerCount, maxMandelWorkerCount, imageCount)

	maxIter := maxIterStart
	var wg1 sync.WaitGroup
	var wg2 sync.WaitGroup

	for x := 0; x < imageCount; x++ {
		mandelWorkerQ := make(chan int, imageHeight)
		imageWorkerQ := make(chan *ComplexPoint, imageHeight*imageWidth)
		t1 := time.Now()
		rectangle.Scale(0.03)
		fname := fmt.Sprintf("mandel-%03v.png", x)
		/*fmt.Printf("[%v|%v|%v|%v|] -> %v\n", maxIter,
		rectangle.center, rectangle.height,
		rectangle.width, fname)*/

		for i := 0; i < maxImageWorkerCount; i++ {
			wg2.Add(1)
			go renderImage(imageWorkerQ, &wg2, fname, maxIter)
		}

		for i := 0; i < maxMandelWorkerCount; i++ {
			wg1.Add(1)
			go renderline(mandelWorkerQ, imageWorkerQ,
				&wg1, maxIter)
		}

		for h := 0; h < imageHeight; h++ {
			mandelWorkerQ <- h
		}
		close(mandelWorkerQ)
		wg1.Wait()
		close(imageWorkerQ)
		wg2.Wait()

		fmt.Printf("%v took %v\n", fname, time.Since(t1))
		/*https://math.stackexchange.com/questions/16970/a-way-to-determine-the-ideal-number-of-maximum-iterations-for-an-arbitrary-zoom
		if x%100 == 0 {
			maxIter *= 10
		}*/
	}

}
