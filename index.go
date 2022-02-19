package goat

// Index represents a position within an ASCII diagram.
type Index struct {
	x int
	y int
}

// Pixel represents the on-screen coordinates for an Index.
type Pixel Index

func (i *Index) asPixel() Pixel {
	return Pixel{x: i.x * 8, y: i.y * 16}
}

func (i *Index) asPixelXY() (int, int) {
	p := i.asPixel()
	return p.x, p.y
}

func (i *Index) east() Index {
	return Index{i.x + 1, i.y}
}

func (i *Index) west() Index {
	return Index{i.x - 1, i.y}
}

func (i *Index) north() Index {
	return Index{i.x, i.y - 1}
}

func (i *Index) south() Index {
	return Index{i.x, i.y + 1}
}

func (i *Index) nWest() Index {
	return Index{i.x - 1, i.y - 1}
}

func (i *Index) nEast() Index {
	return Index{i.x + 1, i.y - 1}
}

func (i *Index) sWest() Index {
	return Index{i.x - 1, i.y + 1}
}

func (i *Index) sEast() Index {
	return Index{i.x + 1, i.y + 1}
}
