package command

import (
	"github.com/urfave/cli/v2"
	"github.com/zengqiang96/tdraw/internal/imgutil"
	"gocv.io/x/gocv"
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
		src := gocv.IMRead(c.String("input"), gocv.IMReadGrayScale)
		imgutil.GrayInvert(&src)
		gocv.IMWrite(c.String("output"), src)
		return nil

	}
	return app
}
