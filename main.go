package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/anthonynsimon/bild/imgio"
)

type Node struct {
	Parent *Node
	X, Y   int
}

const (
	Width     = 854
	Height    = 480
	SeedCount = 100
)

var (
	nodes     [Width][Height]*Node
	heightMap [Width][Height]int
	total     int
)

func (n *Node) Visit() {
	heightMap[n.X][n.Y]++
	if n.Parent != nil {
		n.Parent.Visit()
	}
}

func DLA() *image.Gray {

	now := time.Now()

	rnd := rand.New(rand.NewPCG(1, 2))

	// Initialize seeds
	for range SeedCount {
		x := rnd.IntN(Width)
		y := rnd.IntN(Height)
		n := &Node{nil, x, y}
		nodes[x][y] = n
		n.Visit()
		total++
	}

	// DLA growth with parallelization
	var mu sync.Mutex
	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		localRnd := rand.New(rand.NewPCG(1, 2))
		for {
			mu.Lock()
			if total >= Width*Height {
				mu.Unlock()
				return
			}
			mu.Unlock()

			x := localRnd.IntN(Width)
			y := localRnd.IntN(Height)
			if nodes[x][y] != nil {
				continue
			}

			lx, ly := x, y
			hit := false
			for !hit {
				lx, ly = x, y
				dir := localRnd.IntN(4)
				switch dir {
				case 0:
					x++
				case 1:
					x--
				case 2:
					y++
				case 3:
					y--
				}
				x = (x + Width) % Width
				y = (y + Height) % Height
				if nodes[x][y] != nil {
					hit = true
					mu.Lock()
					total++
					n := &Node{nodes[x][y], lx, ly}
					nodes[lx][ly] = n
					n.Visit()
					mu.Unlock()
				}
			}
		}
	}

	for range 4 { // 4 workers
		wg.Add(1)
		go worker()
	}

	wg.Wait()
	fmt.Println(time.Since(now).Seconds())
	// Create image
	img := image.NewGray(image.Rect(0, 0, Width, Height))
	for x := range Width {
		for y := range Height {
			val := float64(heightMap[x][y])
			c := uint8(math.Min(math.Log(val+1)*20, 255))
			img.SetGray(x, y, color.Gray{Y: c})
		}
	}
	return img
}

func main() {
	im := DLA()
	imgio.Save("out/o2.png", im, imgio.PNGEncoder())
}
