/*
savitzkygolay provides a filter on a set of data which provides
an effective way of smoothing data that generally follows
curves found in polynomials. It is particularly good alternative
to a moving average since it does not introduce a
delay proportial to about half the window length.

Example:

	noise := 15.0
	xs := make([]float64, testSize, testSize)
	ys = make([]float64, testSize, testSize)
	for i := range xs {
		ys[i] = 20 * math.Sin(float64(i)/math.Pi/6) +
					(rand.Float64() * noise) - noise/2.0)
		xs[i] = float64(i)
	}

	filter, err := savitzkygolay.NewFilterWindow(11)
	sgy, err := filter.Process(ys, xs)

Filter interface may be retained to avoid the overhead of pre-computing
the polynomials however the size is proportial to the square of the
window size.

The filter run on O(number of elements * size of window)

Project unit tests generate outputs which illustrate the filter.
*/
package savitzkygolay
