package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/gonum/matrix/mat64"
)

var (
	degree int     = 4
	t      float64 = 10
	r      float64 = 40
	digits int     = 2
)

// Encode a string into coeeficients. The highest dimension is the last element
// in the array.
func getCoefficients(word string) []float64 {
	word = strings.ToUpper(word)
	n := len(word) / degree
	if n < 1 {
		n = 1
	}
	var substring []string

	for i := 0; i < len(word); i += n {

		top := int(math.Min(float64(len(word)), float64(i+n)))
		substring = append(substring, word[i:top])
	}

	var coeffs []float64

	for _, sub := range substring {
		num := 0.0
		for i, char := range sub {
			num += float64(int(char)) * math.Pow10(digits*i)
		}
		coeffs = append(coeffs, num)
	}
	return coeffs
}

// Evaluate a polynomial at a point
func evalAt(x float64, coeffs []float64) float64 {
	ret := 0.0

	for i, coefficient := range coeffs {
		ret += math.Pow(x, float64(i)) * coefficient
	}
	return ret
}

func approxEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

// Return a regressed polynomial from a set of points
func polyfit(X, Y []float64) []float64 {
	ret := make([]float64, degree+1)

	a := vandermonde(X, degree)
	b := mat64.NewDense(len(Y), 1, Y)
	c := mat64.NewDense(degree+1, 1, nil)

	qr := new(mat64.QR)
	qr.Factorize(a)

	err := c.SolveQR(qr, false, b)

	if err != nil {
		fmt.Println(err)
		return nil
	} else {
		mat64.Col(ret, 0, c)

		for i := 0; i < len(ret); i++ {
			ret[i] = round(ret[i], 0.01)
		}

		return ret
	}
}

func vandermonde(a []float64, degree int) *mat64.Dense {
	x := mat64.NewDense(len(a), degree+1, nil)
	for i := range a {
		for j, p := 0, 1.; j <= degree; j, p = j+1, p*a[i] {
			x.Set(i, j, p)
		}
	}
	return x
}

func round(x, unit float64) float64 {
	return math.Round(x/unit) * unit
}

func maxFloat64Slice(s []float64) float64 {
	m := math.Inf(-1)
	for _, e := range s {
		if e > m {
			m = e
		}
	}
	return m
}

func generateRandomTemplate(n int) []float64 {
	var ret []float64
	for i := 0; i < n; i++ {
		y := rand.Float64() * math.Pow10(digits)
		ret = append(ret, y)
	}
	return ret

}

// Lock a string in a fuzzy fault given a fingerprint template
func Lock(secret string, template []float64) [][]float64 {
	var vault [][]float64
	coeffs := getCoefficients(secret)
	fmt.Printf("Coded Coeffs: %v\n", coeffs)

	maxY := math.Inf(-1)

	for _, point := range template {
		y := evalAt(point, coeffs)
		if y > maxY {
			maxY = y
		}
		vault = append(vault, []float64{point, y})
	}

	maxX := maxFloat64Slice(template)

	// Add random floats to disguise points
	for i := t; i < r; i++ {
		// Increase the max to reveal less information about the vault
		xI := rand.Float64() * maxX * 1.1
		yI := rand.Float64() * maxY * 1.1
		vault = append(vault, []float64{xI, yI})
	}

	// Shuffle the order of the vault
	rand.Shuffle(len(vault), func(i, j int) {
		vault[i], vault[j] = vault[j], vault[i]
	})

	return vault
}

func Unlock(template []float64, vault [][]float64) []float64 {
	project := func(x float64) (float64, float64) {
		for _, point := range vault {
			if approxEqual(x, point[0], 0.001) {
				return x, point[1]
			}
		}
		return -1, -1
	}

	var tempX []float64
	var tempY []float64

	for _, val := range template {
		if x, y := project(val); x != -1 {
			tempX = append(tempX, x)
			tempY = append(tempY, y)
		}
	}

	return polyfit(tempX, tempY)
}

func Decode(coeffs []float64) string {
	s := ""
	for _, c := range coeffs {

		if c == 0 {
			continue
		}

		log100 := (int(math.Log10(c)) + 1) / 2

		for i := 0; i < log100; i++ {
			char := int(c/math.Pow10(digits*i)) % int(math.Pow10(digits))
			s += string(rune(char))
		}
	}
	return s
}

func main() {
	rand.Seed(42)
	fmt.Printf("Locking Phrase in Fuzzy Vault\n")

	word := "WHAT'S GOOD!"
	template := generateRandomTemplate(30)
	vault := Lock(word, template)
	fmt.Println("vault:", vault)
	coeffs := Unlock(template, vault)
	fmt.Println("Decoded Coeffs:", coeffs)
	ret := Decode(coeffs)
	fmt.Println("Decoded String:", ret)
}
