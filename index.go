package goat

// Index represents a position within an ASCII diagram.
type Index struct {
	// units of cells
	X, Y int
}

// pixel represents the CSS-pixel coordinates for an Index.
type pixel Index  // XX different units -- create separate base type?

func (i *Index) asPixel() pixel {
	// TODO  define constants rather than hard-wire width and height of cell
	return pixel{
		X: i.X * 8,
		Y: i.Y * 16}
}

func (i *Index) asPixelXY() (int, int) {
	p := i.asPixel()
	return p.X, p.Y
}

func (i *Index) east() Index {
	return Index{i.X + 1, i.Y}
}

func (i *Index) west() Index {
	return Index{i.X - 1, i.Y}
}

func (i *Index) north() Index {
	return Index{i.X, i.Y - 1}
}

func (i *Index) south() Index {
	return Index{i.X, i.Y + 1}
}

func (i *Index) nWest() Index {
	return Index{i.X - 1, i.Y - 1}
}

func (i *Index) nEast() Index {
	return Index{i.X + 1, i.Y - 1}
}

func (i *Index) sWest() Index {
	return Index{i.X - 1, i.Y + 1}
}

func (i *Index) sEast() Index {
	return Index{i.X + 1, i.Y + 1}
}
