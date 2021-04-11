package imgutil

import "gocv.io/x/gocv"

func GrayInvert(src *gocv.Mat) {
	for row := 0; row < src.Rows(); row++ {
		for col := 0; col < src.Cols(); col++ {
			src.SetUCharAt(row, col, 255-src.GetUCharAt(row, col))
		}
	}
}
