package savitzkygolay

import (
	"fmt"
	"math"
)

// Filter provides an interface for filtering data based on a filter configuration
// data is an array of data being filtered
// x is the corresponding x-position of the coordinate
// errors occur only if the filter window is larger than the data
type Filter interface {
	Process(data []float64, x []float64) ([]float64, error)
}

// filterConfiguration provides configurations for the filter
type filterConfiguration struct {
	//  windowSize is the number of points in the window
	windowSize int
	// derivative
	derivative int
	// polynomial is the order of the polynomial used in the fit
	polynomial int

	weights [][]float64
}

// NewFilter creates new savitzky golay based on the provided attributes
func NewFilter(windowSize int, derivitive int, polynomial int) (Filter, error) {
	options := filterConfiguration{windowSize: windowSize, derivative: derivitive, polynomial: polynomial}
	if options.windowSize%2 == 0 || options.windowSize < 5 {
		return nil, fmt.Errorf("options.WindowSize [%d] must be odd and equal to or greater than 5", options.windowSize)
	}
	if options.derivative < 0 {
		return nil, fmt.Errorf("options.Derivative [%d] must be euqal or greater than 0", options.derivative)
	}
	if options.polynomial < 0 {
		return nil, fmt.Errorf("options.Polynomial [%d] must be equal or greater than 0", options.polynomial)
	}
	options.weights = options.computeWeights()
	return options, nil
}

// NewFilterWindow creates a new savitzky golay filter with default settings of a 3rd order polynomial
// windowSize must be odd and greater than 4
func NewFilterWindow(windowSize int) (Filter, error) {
	return NewFilter(windowSize, 0, 3)
}

// Process executes a on the input set
func (options filterConfiguration) Process(data []float64, h []float64) ([]float64, error) {
	if options.windowSize > len(data) {
		return nil, fmt.Errorf("data length [%d] must be larger than options.WindowSize[%d]", len(data), options.windowSize)
	}

	halfWindow := int(math.Floor(float64(options.windowSize) / 2.0))
	numPoints := len(data)
	results := make([]float64, numPoints)
	weights := options.weights
	hs := 0.0

	//For the borders
	for i := 0; i < halfWindow; i++ {
		wg1 := weights[halfWindow-i-1]
		wg2 := weights[halfWindow+i+1]
		d1 := 0.0
		d2 := 0.0
		for l := 0; l < options.windowSize; l++ {
			d1 += wg1[l] * data[l]
			d2 += wg2[l] * data[numPoints-options.windowSize+l]
		}
		hs = getHs(h, halfWindow-i-1, halfWindow, options.derivative)
		results[halfWindow-i-1] = d1 / hs
		hs = getHs(h, numPoints-halfWindow+i, halfWindow, options.derivative)
		results[numPoints-halfWindow+i] = d2 / hs
	}

	//For the internal points
	wg := weights[halfWindow]
	for i := options.windowSize; i <= numPoints; i++ {
		d := 0.0
		for l := 0; l < options.windowSize; l++ {
			d += wg[l] * data[l+i-options.windowSize]
		}
		hs = getHs(h, i-halfWindow-1, halfWindow, options.derivative)
		results[i-halfWindow-1] = d / hs
	}
	return results, nil
}

func getHs(h []float64, center int, half int, derivative int) float64 {
	hs := 0.0
	count := 0
	for i := center - half; i < center+half; i++ {
		if i >= 0 && i < len(h)-1 {
			hs += h[i+1] - h[i]
			count++
		}
	}
	return math.Pow(hs/float64(count), float64(derivative))
}

func gramPolynomial(i int, m int, k int, s int) float64 {
	result := 0.0
	if k > 0 {
		result =
			float64(float64(4*k-2)/float64(k*(2*m-k+1)))*
				(float64(i)*gramPolynomial(i, m, k-1, s)+float64(s)*gramPolynomial(i, m, k-1, s-1)) -
				(float64((k-1)*(2*m+k))/float64(k*(2*m-k+1)))*
					gramPolynomial(i, m, k-2, s)
	} else {
		if k == 0 && s == 0 {
			result = 1
		} else {
			result = 0
		}
	}
	return result
}

func productOfRange(a, b int) int {
	gf := 1
	if a >= b {
		for j := a - b + 1; j <= a; j++ {
			gf *= j
		}
	}
	return gf
}

func polyWeight(i, t, windowMiddle, polynomial, derivitive int) float64 {
	sum := 0.0
	for k := 0; k <= polynomial; k++ {
		sum +=
			float64(2*k+1) *
				(float64(productOfRange(2*windowMiddle, k)) / float64(productOfRange(2*windowMiddle+k+1, k+1))) *
				gramPolynomial(i, windowMiddle, k, 0) * gramPolynomial(t, windowMiddle, k, derivitive)
	}
	return sum
}

func (options *filterConfiguration) computeWeights() [][]float64 {
	weights := make([][]float64, options.windowSize)
	windowMiddle := int(math.Floor(float64(options.windowSize) / 2.0))
	for row := -windowMiddle; row <= windowMiddle; row++ {
		weights[row+windowMiddle] = make([]float64, options.windowSize)
		for col := -windowMiddle; col <= windowMiddle; col++ {
			weights[row+windowMiddle][col+windowMiddle] = polyWeight(col, row, windowMiddle, options.polynomial, options.derivative)
		}
	}
	return weights
}
