package savitzkygolay

import (
	"image/color"
	"math"
	"math/rand"
	"strings"
	"testing"

	assert "github.com/stretchr/testify/assert"
	plot "gonum.org/v1/plot"
	plotter "gonum.org/v1/plot/plotter"
)

const testSize = 500

type xyPairs struct {
	xs []float64
	ys []float64
}

func (p xyPairs) Len() int {
	v := p.xs
	return len(v)
}

func (p xyPairs) XY(i int) (x, y float64) {
	return p.xs[i], p.ys[i]
}

func makexyPairs(size int) xyPairs {
	var r xyPairs
	r.xs = make([]float64, testSize, testSize)
	r.ys = make([]float64, testSize, testSize)
	return r
}

func Test_SavitzkyGolay_Args(t *testing.T) {
	_, err := NewFilterWindow(0)
	assert.Error(t, err, "Window size too small")
	_, err = NewFilterWindow(3)
	assert.Error(t, err, "Window size too small")
	_, err = NewFilterWindow(6)
	assert.Error(t, err, "Window size even")

	_, err = NewFilter(7, -1, 3)
	assert.Error(t, err, "Derivitive must be non-negative")

	_, err = NewFilter(7, 0, -1)
	assert.Error(t, err, "Polynomial must be non-negative")

	f, err := NewFilter(7, 0, 3)
	assert.NoError(t, err, "Filter should be allowed")
	xs := []float64{1, 2, 3, 4, 5}
	_, err = f.Process(xs, xs)
	assert.Error(t, err, "Window larger than data")

}

func Test_SavitzkyGolay_Line(t *testing.T) {
	pairs := makexyPairs(testSize)
	for i := range pairs.xs {
		pairs.ys[i] = math.Pi
		pairs.xs[i] = float64(i)
	}

	filter, err := NewFilterWindow(5)
	assert.NoError(t, err, "No filter initialization error expected")
	sgy, err := filter.Process(pairs.ys, pairs.xs)
	assert.NoError(t, err, "No error expected")
	copy := pairs
	copy.ys = sgy

	max, avg := pairs.difference(&copy)
	assert.NoError(t, err, "No error expected")
	assert.Less(t, avg, float64(0.1), "Small average differences")
	assert.Less(t, max, float64(0.5), "Small average differences")

	visualTest(&pairs, nil, &copy, "Smoothed Line")
}

func Test_SavitzkyGolay_Sin(t *testing.T) {
	pairs := makexyPairs(testSize)
	for i := range pairs.xs {
		pairs.ys[i] = 20 * math.Sin(float64(i)/math.Pi/6)
		pairs.xs[i] = float64(i)
	}
	copy := pairs

	filter, err := NewFilterWindow(41)
	assert.NoError(t, err, "No filter initialization error expected")
	copy.ys, err = filter.Process(pairs.ys, pairs.xs)
	assert.NoError(t, err, "No error expected")

	max, avg := pairs.difference(&copy)
	assert.NoError(t, err, "No error expected")
	assert.Less(t, avg, float64(0.1), "Small average differences")
	assert.Less(t, max, float64(0.5), "Small average differences")

	visualTest(&pairs, nil, &copy, "Smoothed Sin")
}

func noise(size float64) float64 {
	return (rand.Float64() * size) - size/2
}

func Test_SavitzkyGolay_SinNoise(t *testing.T) {
	pairs := makexyPairs(testSize)
	for i := range pairs.xs {
		pairs.ys[i] = 20 * math.Sin(float64(i)/math.Pi/6)
		pairs.xs[i] = float64(i)
	}

	noisy := pairs.addNoise(5.0)

	filter, err := NewFilterWindow(21)
	assert.NoError(t, err, "No filter initialization error expected")
	sgy, err := filter.Process(noisy.ys, noisy.xs)
	assert.NoError(t, err, "No error expected")

	copy := pairs
	copy.ys = sgy
	visualTest(&pairs, &noisy, &copy, "Smoothed Sin with Noise")
}

func (p *xyPairs) addNoise(n float64) xyPairs {
	noisy := *p
	noisy.ys = make([]float64, len(p.ys))
	for i, y := range p.ys {
		noisy.ys[i] = y + noise(n)
	}
	return noisy
}

func Test_SavitzkyGolay_SinNoise_1(t *testing.T) {
	pairs := makexyPairs(testSize)
	for i := range pairs.xs {
		pairs.ys[i] = 20 + 20*math.Sin(float64(i)/math.Pi/4) +
			20*math.Sin(float64(i+10)/math.Pi/2)
		pairs.xs[i] = float64(i)
	}
	noisy := pairs.addNoise(10.0)

	filter, err := NewFilter(7, 0, 1)
	assert.NoError(t, err, "No filter initialization error expected")
	sgy, err := filter.Process(pairs.ys, pairs.xs)
	assert.NoError(t, err, "No error expected")

	copy := pairs
	copy.ys = sgy
	visualTest(&pairs, &noisy, &copy, "Smoothed Sin with Noise Using Single Order Polynomial")
}

func visualTest(original, noisy, filtered *xyPairs, title string) {
	p, _ := plot.New()
	p.Title.Text = title
	if original != nil {
		line, _ := plotter.NewLine(original)
		line.Color = color.Gray{Y: 128}
		p.Add(line)
		p.Legend.Add("Original", line)
	}
	if noisy != nil {
		scatter, _ := plotter.NewScatter(noisy)
		p.Add(scatter)
		p.Legend.Add("Noisy Added", scatter)

		avgXY := noisy
		avgXY.ys = movingAverage(15, noisy.ys)
		line, _ := plotter.NewLine(avgXY)
		line.Color = color.RGBA{B: 255, G: 255}
		p.Add(line)
		p.Legend.Add("15-point Moving average", line)

		avgXY = noisy
		avgXY.ys = movingAverage(30, noisy.ys)
		line, _ = plotter.NewLine(avgXY)
		line.Color = color.RGBA{B: 128, G: 128}
		p.Add(line)
		p.Legend.Add("30-point Moving average", line)

	}

	if filtered != nil {
		green := color.RGBA{R: 255, B: 255}
		line, _ := plotter.NewLine(filtered)
		line.Color = green
		p.Add(line)
		p.Legend.Add("SG Filtered", line)
	}

	path := strings.ToLower(strings.ReplaceAll(title, " ", "_")) + ".png"
	_ = p.Save(512*2, 512, path)
}

func (p *xyPairs) difference(o *xyPairs) (max, average float64) {
	for i, v := range p.ys {
		oy := o.ys[i]
		d := math.Abs(oy - v)
		max = math.Max(max, d)
		average += d
	}
	average = average / float64(len(p.ys))
	return max, average
}

func movingAverage(windowSize int, values []float64) []float64 {
	var r []float64
	var w []float64
	for _, v := range values {
		w = last(append(w, v), windowSize)
		r = append(r, avg(w))
	}
	return r
}

func last(v []float64, l int) []float64 {
	if len(v) > l {
		v = v[len(v)-l : len(v)-1]
	}
	return v
}
func avg(v []float64) float64 {
	r := 0.0
	for _, a := range v {
		r += a
	}
	return r / float64(len(v))

}
