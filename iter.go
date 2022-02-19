package goat

type canvasIterator func(width int, height int) chan Index

func upDown(width int, height int) chan Index {
	c := make(chan Index, width*height)

	go func() {
		for w := 0; w < width; w++ {
			for h := 0; h < height; h++ {
				c <- Index{w, h}
			}
		}
		close(c)
	}()

	return c
}

func leftRight(width int, height int) chan Index {
	c := make(chan Index, width*height)

	// Transpose an upDown order.
	go func() {
		for i := range upDown(height, width) {
			c <- Index{i.y, i.x}
		}

		close(c)
	}()

	return c
}

func diagDown(width int, height int) chan Index {
	c := make(chan Index, width*height)

	go func() {
		minSum := -height + 1
		maxSum := width

		for sum := minSum; sum <= maxSum; sum++ {
			for w := 0; w < width; w++ {
				for h := 0; h < height; h++ {
					if w-h == sum {
						c <- Index{w, h}
					}
				}
			}
		}
		close(c)
	}()

	return c
}

func diagUp(width int, height int) chan Index {
	c := make(chan Index, width*height)

	go func() {
		maxSum := width + height - 2

		for sum := 0; sum <= maxSum; sum++ {
			for w := 0; w < width; w++ {
				for h := 0; h < height; h++ {
					if h+w == sum {
						c <- Index{w, h}
					}
				}
			}
		}
		close(c)
	}()

	return c
}
