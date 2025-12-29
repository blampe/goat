package svg

// XX  Simplify to closure functions rather than channels?
//      X Callers currently use 'range' -- rewrite would be required.
// XX  Use stdlib package "iter"?
type CanvasIterator func(width int, height int) chan XyIndex

func UpDownMinor(width int, height int) chan XyIndex {
	c := make(chan XyIndex)
	go func() {
		for w := 0; w < width; w++ {
			for h := 0; h < height; h++ {
				c <- XyIndex{w, h}
			}
		}
		close(c)
	}()
	return c
}

func LeftRightMinor(width int, height int) chan XyIndex {
	c := make(chan XyIndex)
	go func() {
		for h := 0; h < height; h++ {
			for w := 0; w < width; w++ {
				c <- XyIndex{w, h}
			}
		}
		close(c)
	}()
	return c
}

func DiagDown(width int, height int) chan XyIndex {
	c := make(chan XyIndex)
	go func() {
		minSum := -height + 1
		maxSum := width

		for sum := minSum; sum <= maxSum; sum++ {
			for w := 0; w < width; w++ {
				for h := 0; h < height; h++ {
					if w-h == sum {
						c <- XyIndex{w, h}
					}
				}
			}
		}
		close(c)
	}()
	return c
}

func DiagUp(width int, height int) chan XyIndex {
	c := make(chan XyIndex)
	go func() {
		maxSum := width + height - 2

		for sum := 0; sum <= maxSum; sum++ {
			for w := 0; w < width; w++ {
				for h := 0; h < height; h++ {
					if h+w == sum {
						c <- XyIndex{w, h}
					}
				}
			}
		}
		close(c)
	}()
	return c
}
