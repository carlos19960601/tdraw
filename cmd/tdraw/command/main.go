package command

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"math"
	"os"

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
			Usage:    "输入图片名称",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "output",
			Aliases:  []string{"o"},
			Usage:    "输出图片目录",
			Required: true,
		},
		&cli.IntFlag{
			Name:    "radius",
			Aliases: []string{"r"},
			Usage:   "输出图片圆的半径",
			Value:   500,
		},
		&cli.IntFlag{
			Name:    "pins",
			Aliases: []string{"p"},
			Usage:   "钉的数量",
			Value:   200,
		},
		&cli.IntFlag{
			Name:    "lines",
			Aliases: []string{"l"},
			Usage:   "线的数量",
			Value:   1000,
		},
		&cli.BoolFlag{
			Name:    "inverted",
			Aliases: []string{"iv"},
			Usage:   "是否打印inverted",
			Value:   false,
		},
		&cli.BoolFlag{
			Name:    "masked",
			Aliases: []string{"m"},
			Usage:   "是否打印masked",
			Value:   false,
		},
		&cli.Float64Flag{
			Name:    "lwidth",
			Aliases: []string{"w"},
			Usage:   "线宽(0-1)",
			Value:   0.5,
		},
		&cli.UintFlag{
			Name:    "ldeepth",
			Aliases: []string{"d"},
			Usage:   "线颜色深度(0-255)",
			Value:   0,
		},
		&cli.BoolFlag{
			Name:    "ppins",
			Aliases: []string{"pp"},
			Usage:   "打印pin的顺序",
			Value:   false,
		},
		&cli.BoolFlag{
			Name:    "gif",
			Aliases: []string{"g"},
			Usage:   "是否生成gif",
			Value:   false,
		},
	}

	app.Action = func(c *cli.Context) error {
		src, err := imaging.Open(c.String("input"))
		if err != nil {
			return err
		}

		output := c.String("output")
		radius := c.Int("radius")
		numPins := c.Int("pins")
		numLines := c.Int("lines")
		filled := imaging.Fill(src, 2*radius, 2*radius, imaging.Center, imaging.Lanczos)
		greyscale := imaging.Grayscale(filled)
		inverted := imaging.Invert(greyscale)
		if c.Bool("inverted") {
			if err = imaging.Save(inverted, fmt.Sprintf("%s/inverted.png", output)); err != nil {
				return err
			}
		}

		dc := gg.NewContext(2*radius, 2*radius)
		dc.DrawCircle(float64(radius), float64(radius), float64(radius))
		dc.Clip()
		dc.DrawImage(inverted, 0, 0)
		masked := dc.Image().(*image.RGBA)
		if c.Bool("masked") {
			if err = imaging.Save(masked, fmt.Sprintf("%s/masked.png", output)); err != nil {
				return err
			}
		}

		dc = gg.NewContext(2*radius, 2*radius)
		dc.DrawRectangle(0, 0, float64(2*radius), float64(2*radius))
		dc.SetColor(color.Transparent)
		dc.Fill()
		dc.SetColor(color.Gray{Y: uint8(c.Uint("ldeepth"))})

		outGif := &gif.GIF{}
		myPalette := color.Palette{
			color.Transparent,
			color.Black,
			color.White,
		}
		lines := make([]int, 0, numLines)
		oldPin := 0
		lines = append(lines, oldPin)
		coords := pinCoords(radius, numPins)
		previousPins := make([]int, 0, 3)

		for index := 0; index < numLines; index++ {
			oldCoord := coords[oldPin]
			var bestLine int64
			var bestPin int

			for index := 1; index <= numPins; index++ {
				pin := (oldPin + index) % numPins
				coord := coords[pin]

				lineCoords := linePixels(oldCoord, coord)

				var lineSum int64
				for _, lineCoord := range lineCoords {
					c := color.GrayModel.Convert(masked.At(lineCoord.x, lineCoord.y)).(color.Gray)
					lineSum += int64(c.Y)
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
				c := color.GrayModel.Convert(masked.At(intPoint.x, intPoint.y)).(color.Gray)
				c.Y = 0
				masked.SetRGBA(intPoint.x, intPoint.y, color.RGBAModel.Convert(c).(color.RGBA))
			}

			lines = append(lines, bestPin)

			dc.DrawLine(oldCoord.x, oldCoord.y, coords[bestPin].x, coords[bestPin].y)
			dc.SetLineWidth(c.Float64("lwidth"))
			dc.Stroke()

			if c.Bool("gif") {
				img := dc.Image()
				bounds := img.Bounds()
				palettedImage := image.NewPaletted(bounds, myPalette)
				draw.Draw(palettedImage, palettedImage.Rect, img, bounds.Min, draw.Src)
				outGif.Image = append(outGif.Image, palettedImage)
				outGif.Disposal = append(outGif.Disposal, gif.DisposalBackground) //透明图片需要设置
				outGif.Delay = append(outGif.Delay, 0)
			}

			if bestPin == oldPin {
				break
			}
			oldPin = bestPin
		}

		if err = dc.SavePNG(fmt.Sprintf("%s/result.out.png", output)); err != nil {
			return err
		}

		if c.Bool("gif") {
			opfile, err := os.Create(fmt.Sprintf("%s/result.out.gif", output))
			if err != nil {
				return errors.New("Failed ceating .gif file on disk: " + err.Error())
			}

			err = gif.EncodeAll(opfile, outGif)
			if err != nil {
				return errors.New("Failed gif encoding: " + err.Error())
			}
		}

		if c.Bool("ppins") {
			fmt.Println("======================= pins ========================")
			for index, line := range lines {
				fmt.Printf("%3d  ", line)
				if index%10 == 0 {
					fmt.Println()
				}
			}
			fmt.Println()
			fmt.Println("======================= pins ========================")
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
