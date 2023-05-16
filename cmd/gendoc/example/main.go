package main

/* This program renders a supersampled Mandelbulb.
 * It takes about 12 seconds to run on my laptop.
 * CC0 licensed.
 */

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"sync"
)

const (
	width, height      = 1920, 1080
	ssWidth, ssHeight  = 3840, 2160
	aspectRatio        = float64(ssWidth) / float64(ssHeight)
	fov                = math.Pi / 8
	maxIter            = 1000
	power              = 9
	escapeRadius       = 1.6
	supersamplingRatio = 2
)

var wg sync.WaitGroup

type vec3 struct{ x, y, z float64 }

func (v vec3) add(w vec3) vec3    { return vec3{v.x + w.x, v.y + w.y, v.z + w.z} }
func (v vec3) mul(s float64) vec3 { return vec3{v.x * s, v.y * s, v.z * s} }
func (v vec3) length() float64    { return math.Sqrt(v.x*v.x + v.y*v.y + v.z*v.z) }
func (v vec3) normalize() vec3    { return v.mul(1 / v.length()) }

func mandelbulb(p vec3) float64 {
	z, r, theta, phi := p, 0.0, 0.0, 0.0
	for i := 0; i < maxIter; i++ {
		r = z.length()
		if r > escapeRadius {
			break
		}
		theta, phi = math.Acos(z.z/r)*power, math.Atan2(z.y, z.x)*power
		r = math.Pow(r, power)
		z.x, z.y, z.z = p.x+r*math.Sin(theta)*math.Cos(phi), p.y+r*math.Sin(theta)*math.Sin(phi), p.z+r*math.Cos(theta)
	}
	return r
}

func renderPixel(x, y int) color.Color {
	px, py := (2*float64(x)/float64(ssWidth)-1)*aspectRatio*math.Tan(fov/2), (1-2*float64(y)/float64(ssHeight))*math.Tan(fov/2)
	direction := vec3{px, py, -1}.normalize()
	t := 0.0
	for i := 0; i < maxIter; i++ {
		p := direction.mul(t).add(vec3{0, 0, 5})
		d := mandelbulb(p) - 1
		if d < 0.001 {
			return color.Gray{uint8(255 - t*255/100)}
		}
		t += d
		if t > 100 {
			break
		}
	}
	return color.Black
}

func main() {
	ssImg := image.NewRGBA(image.Rect(0, 0, ssWidth, ssHeight))
	for y := 0; y < ssHeight; y++ {
		wg.Add(1)
		go func(y int) {
			defer wg.Done()
			for x := 0; x < ssWidth; x++ {
				ssImg.Set(x, y, renderPixel(x, y))
			}
		}(y)
	}
	wg.Wait()

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := 0, 0, 0, 0
			for dy := 0; dy < supersamplingRatio; dy++ {
				for dx := 0; dx < supersamplingRatio; dx++ {
					c := ssImg.RGBAAt(x*supersamplingRatio+dx, y*supersamplingRatio+dy)
					r += int(c.R)
					g += int(c.G)
					b += int(c.B)
					a += int(c.A)
				}
			}
			r /= supersamplingRatio * supersamplingRatio
			g /= supersamplingRatio * supersamplingRatio
			b /= supersamplingRatio * supersamplingRatio
			a /= supersamplingRatio * supersamplingRatio
			img.Set(x, y, color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)})
		}
	}

	file, err := os.Create("mandelbulb.png")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()
	err = png.Encode(file, img)
	if err != nil {
		fmt.Println("Error encoding image:", err)
	}
}
