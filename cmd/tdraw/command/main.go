package command

import (
	"image"
	"image/color"
	"math"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/urfave/cli/v2"
	"gonum.org/v1/gonum/floats"
)

func App() *cli.App {
	app := cli.NewApp()
	app.Name = "tdraw"
	app.Description = "使用thread绘画"

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:     "input",
			Aliases:  []string{"i"},
			Usage:    "输入图片路径",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "output",
			Aliases:  []string{"o"},
			Usage:    "输出图片路径",
			Required: true,
		},
	}

	app.Action = func(c *cli.Context) error {
		src, err := imaging.Open(c.String("input"))
		if err != nil {
			return err
		}

		radius := 50

		filled := imaging.Fill(src, 2*radius, 2*radius, imaging.Center, imaging.Lanczos)
		greyscale := imaging.Grayscale(filled)
		inverted := imaging.Invert(greyscale)

		dc := gg.NewContext(2*radius, 2*radius)
		dc.DrawCircle(float64(radius), float64(radius), float64(radius))
		dc.Clip()
		dc.DrawImage(inverted, 0, 0)
		masked := dc.Image().(*image.RGBA)

		dc.Clear()
		dc.DrawRectangle(0, 0, float64(2*radius), float64(2*radius))
		dc.SetColor(color.White)
		dc.Fill()

		var lines []Line
		oldPin := 0
		coords := pinCoords(radius, 200)
		previousPins := make([]int, 0, 3)

		for range coords {
			oldCoord := coords[oldPin]
			var bestLine int64
			var bestPin int

			for index := 1; index <= 200; index++ {
				pin := (oldPin + index) % 200
				coord := coords[pin]

				lineCoords := linePixels(oldCoord, coord)

				var lineSum int64
				for _, lineCoord := range lineCoords {
					r, g, b, _ := masked.At(lineCoord.x, lineCoord.y).RGBA()
					lineSum += int64(r)
					lineSum += int64(g)
					lineSum += int64(b)
				}
				if lineSum > bestLine && !in(previousPins, pin) {
					bestLine = lineSum
					bestPin = pin
				}
			}

			if len(previousPins) >= 3 {
				previousPins = previousPins[1:]
			}
			previousPins = append(previousPins, bestPin)

			intPoints := linePixels(coords[oldPin], coords[bestPin])
			for _, intPoint := range intPoints {
				c := masked.At(intPoint.x, intPoint.y).(color.RGBA)
				c.R -= 15
				c.G -= 15
				c.B -= 15
				masked.SetRGBA(intPoint.x, intPoint.y, c)
			}

			lines = append(lines, Line{
				s: oldPin,
				e: bestPin,
			})

			dc.DrawLine(oldCoord.x, oldCoord.y, coords[bestPin].x, coords[bestPin].y)
			dc.SetColor(color.Black)
			dc.Fill()
			dc.SavePNG(c.String("output"))

			if bestPin == oldPin {
				break
			}

			oldPin = bestPin
		}

		// err = imaging.Save(masked, "../assert/masked.jpg")
		// err = imaging.Save(greyscale, c.String("output"))
		if err != nil {
			return err
		}

		return nil
	}
	return app
}

func in(l []int, n int) bool {
	for _, c := range l {
		if c == n {
			return true
		}
	}
	return false
}

func pinCoords(radius, numPins int) []Point {
	a := make([]float64, numPins+1)
	floats.Span(a, 0, 2*math.Pi)
	points := make([]Point, 0, numPins)
	for i := 0; i < numPins; i++ {
		points = append(points, Point{
			x: float64(radius) + float64(radius)*math.Cos(a[i]),
			y: float64(radius) + float64(radius)*math.Sin(a[i]),
		})
	}
	return points
}

func linePixels(pin0, pin1 Point) []IntPoint {
	length := int(math.Hypot(pin1.x-pin0.x, pin1.y-pin0.y))
	if length < 2 {
		return nil
	}

	x := make([]float64, length)
	y := make([]float64, length)

	floats.Span(x, pin0.x, pin1.x)
	floats.Span(y, pin0.y, pin1.y)

	p := make([]IntPoint, 0, length)
	for i := 0; i < length; i++ {
		p = append(p, IntPoint{
			x: int(x[i]),
			y: int(y[i]),
		})
	}

	return p
}

type Point struct {
	x float64
	y float64
}

type IntPoint struct {
	x int
	y int
}

type Line struct {
	s int
	e int
}
