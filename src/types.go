package goaat

type Index struct {
	x int
	y int
}

type Pixel Index

func (i *Index) asPixel() Pixel {
	return Pixel{x: i.x * 8, y: i.y * 16}
}
