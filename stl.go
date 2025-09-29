package main

import (
	"bufio"
	"math"
	"os"
	"strconv"
	"strings"
)

type Vertex struct {
	X, Y, Z float64
}

type Triangle struct {
	V1, V2, V3 Vertex
	Normal     Vertex
}

type STLModel struct {
	Name      string
	Triangles []Triangle
	MinBounds Vertex
	MaxBounds Vertex
}

func ParseSTL(filename string) (*STLModel, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	model := &STLModel{
		Triangles: make([]Triangle, 0),
		MinBounds: Vertex{X: math.MaxFloat64, Y: math.MaxFloat64, Z: math.MaxFloat64},
		MaxBounds: Vertex{X: -math.MaxFloat64, Y: -math.MaxFloat64, Z: -math.MaxFloat64},
	}

	scanner := bufio.NewScanner(file)
	var currentTriangle Triangle
	var vertexCount int

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.Fields(line)

		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "solid":
			if len(parts) > 1 {
				model.Name = strings.Join(parts[1:], " ")
				model.Name = strings.Trim(model.Name, "\"")
			}
		case "facet":
			if len(parts) >= 5 && parts[1] == "normal" {
				currentTriangle.Normal = Vertex{
					X: parseFloat(parts[2]),
					Y: parseFloat(parts[3]),
					Z: parseFloat(parts[4]),
				}
				vertexCount = 0
			}
		case "vertex":
			if len(parts) >= 4 {
				v := Vertex{
					X: parseFloat(parts[1]),
					Y: parseFloat(parts[2]),
					Z: parseFloat(parts[3]),
				}

				model.updateBounds(v)

				switch vertexCount {
				case 0:
					currentTriangle.V1 = v
				case 1:
					currentTriangle.V2 = v
				case 2:
					currentTriangle.V3 = v
				}
				vertexCount++
			}
		case "endfacet":
			if vertexCount == 3 {
				model.Triangles = append(model.Triangles, currentTriangle)
			}
			currentTriangle = Triangle{}
			vertexCount = 0
		}
	}

	return model, scanner.Err()
}

func (m *STLModel) updateBounds(v Vertex) {
	m.MinBounds.X = math.Min(m.MinBounds.X, v.X)
	m.MinBounds.Y = math.Min(m.MinBounds.Y, v.Y)
	m.MinBounds.Z = math.Min(m.MinBounds.Z, v.Z)
	m.MaxBounds.X = math.Max(m.MaxBounds.X, v.X)
	m.MaxBounds.Y = math.Max(m.MaxBounds.Y, v.Y)
	m.MaxBounds.Z = math.Max(m.MaxBounds.Z, v.Z)
}

func parseFloat(s string) float64 {
	val, _ := strconv.ParseFloat(s, 64)
	return val
}

type Matrix3x3 struct {
	m [3][3]float64
}

func RotationMatrixX(angle float64) Matrix3x3 {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return Matrix3x3{
		m: [3][3]float64{
			{1, 0, 0},
			{0, cos, -sin},
			{0, sin, cos},
		},
	}
}

func RotationMatrixY(angle float64) Matrix3x3 {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return Matrix3x3{
		m: [3][3]float64{
			{cos, 0, sin},
			{0, 1, 0},
			{-sin, 0, cos},
		},
	}
}

func RotationMatrixZ(angle float64) Matrix3x3 {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return Matrix3x3{
		m: [3][3]float64{
			{cos, -sin, 0},
			{sin, cos, 0},
			{0, 0, 1},
		},
	}
}

func (m Matrix3x3) MultiplyVector(v Vertex) Vertex {
	return Vertex{
		X: m.m[0][0]*v.X + m.m[0][1]*v.Y + m.m[0][2]*v.Z,
		Y: m.m[1][0]*v.X + m.m[1][1]*v.Y + m.m[1][2]*v.Z,
		Z: m.m[2][0]*v.X + m.m[2][1]*v.Y + m.m[2][2]*v.Z,
	}
}

func (m Matrix3x3) Multiply(other Matrix3x3) Matrix3x3 {
	result := Matrix3x3{}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				result.m[i][j] += m.m[i][k] * other.m[k][j]
			}
		}
	}
	return result
}

type Renderer struct {
	Width  int
	Height int
	Buffer [][]rune
	ZBuffer [][]float64
}

func NewRenderer(width, height int) *Renderer {
	r := &Renderer{
		Width:  width,
		Height: height,
	}
	r.Clear()
	return r
}

func (r *Renderer) Clear() {
	r.Buffer = make([][]rune, r.Height)
	r.ZBuffer = make([][]float64, r.Height)
	for y := 0; y < r.Height; y++ {
		r.Buffer[y] = make([]rune, r.Width)
		r.ZBuffer[y] = make([]float64, r.Width)
		for x := 0; x < r.Width; x++ {
			r.Buffer[y][x] = ' '
			r.ZBuffer[y][x] = -math.MaxFloat64
		}
	}
}

func (r *Renderer) RenderModel(model *STLModel, rotX, rotY, rotZ float64, style string) string {
	r.Clear()

	center := Vertex{
		X: (model.MinBounds.X + model.MaxBounds.X) / 2,
		Y: (model.MinBounds.Y + model.MaxBounds.Y) / 2,
		Z: (model.MinBounds.Z + model.MaxBounds.Z) / 2,
	}

	maxDim := math.Max(
		model.MaxBounds.X-model.MinBounds.X,
		math.Max(
			model.MaxBounds.Y-model.MinBounds.Y,
			model.MaxBounds.Z-model.MinBounds.Z,
		),
	)

	scale := math.Min(float64(r.Width), float64(r.Height)) * 0.9 / maxDim

	rotMatrix := RotationMatrixX(rotX).
		Multiply(RotationMatrixY(rotY)).
		Multiply(RotationMatrixZ(rotZ))

	chars := []rune{'█', '▓', '▒', '░', '▐', '▌', '·', ' '}
	if style == "wireframe" {
		chars = []rune{'#', '+', '*', 'o', '.', '·', ' '}
	}

	for _, tri := range model.Triangles {
		v1 := r.projectVertex(tri.V1, center, rotMatrix, scale)
		v2 := r.projectVertex(tri.V2, center, rotMatrix, scale)
		v3 := r.projectVertex(tri.V3, center, rotMatrix, scale)

		if style == "wireframe" {
			r.drawLine(v1, v2, chars[0])
			r.drawLine(v2, v3, chars[0])
			r.drawLine(v3, v1, chars[0])
		} else {
			r.fillTriangle(v1, v2, v3, chars)
		}
	}

	var result strings.Builder
	for y := 0; y < r.Height; y++ {
		for x := 0; x < r.Width; x++ {
			result.WriteRune(r.Buffer[y][x])
		}
		if y < r.Height-1 {
			result.WriteString("\n")
		}
	}
	return result.String()
}

func (r *Renderer) projectVertex(v, center Vertex, rotation Matrix3x3, scale float64) Vertex {
	translated := Vertex{
		X: v.X - center.X,
		Y: v.Y - center.Y,
		Z: v.Z - center.Z,
	}

	rotated := rotation.MultiplyVector(translated)

	projected := Vertex{
		X: rotated.X*scale + float64(r.Width)/2.0,
		Y: -rotated.Y*scale + float64(r.Height)/2.0,
		Z: rotated.Z,
	}

	return projected
}

func (r *Renderer) drawLine(v1, v2 Vertex, char rune) {
	x0, y0 := int(v1.X), int(v1.Y)
	x1, y1 := int(v2.X), int(v2.Y)
	
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := 1
	sy := 1
	
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	
	err := dx - dy
	
	for {
		if x0 >= 0 && x0 < r.Width && y0 >= 0 && y0 < r.Height {
			z := (v1.Z + v2.Z) / 2
			if z > r.ZBuffer[y0][x0] {
				r.Buffer[y0][x0] = char
				r.ZBuffer[y0][x0] = z
			}
		}
		
		if x0 == x1 && y0 == y1 {
			break
		}
		
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func (r *Renderer) fillTriangle(v1, v2, v3 Vertex, chars []rune) {
	minX := int(math.Max(0, math.Min(v1.X, math.Min(v2.X, v3.X))))
	maxX := int(math.Min(float64(r.Width-1), math.Max(v1.X, math.Max(v2.X, v3.X))))
	minY := int(math.Max(0, math.Min(v1.Y, math.Min(v2.Y, v3.Y))))
	maxY := int(math.Min(float64(r.Height-1), math.Max(v1.Y, math.Max(v2.Y, v3.Y))))

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if r.pointInTriangle(float64(x), float64(y), v1, v2, v3) {
				z := (v1.Z + v2.Z + v3.Z) / 3
				if z > r.ZBuffer[y][x] {
					depth := (z + 100) / 200
					charIdx := int(depth * float64(len(chars)-1))
					if charIdx < 0 {
						charIdx = 0
					}
					if charIdx >= len(chars) {
						charIdx = len(chars) - 1
					}
					r.Buffer[y][x] = chars[charIdx]
					r.ZBuffer[y][x] = z
				}
			}
		}
	}
}

func (r *Renderer) pointInTriangle(px, py float64, v1, v2, v3 Vertex) bool {
	sign := func(p1, p2, p3 Vertex) float64 {
		return (p1.X-p3.X)*(p2.Y-p3.Y) - (p2.X-p3.X)*(p1.Y-p3.Y)
	}

	p := Vertex{X: px, Y: py, Z: 0}
	d1 := sign(p, v1, v2)
	d2 := sign(p, v2, v3)
	d3 := sign(p, v3, v1)

	hasNeg := (d1 < 0) || (d2 < 0) || (d3 < 0)
	hasPos := (d1 > 0) || (d2 > 0) || (d3 > 0)

	return !(hasNeg && hasPos)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func LoadAndRenderSTL(filename string, width, height int, rotX, rotY, rotZ float64, style string) (string, error) {
	model, err := ParseSTL(filename)
	if err != nil {
		return "", err
	}

	renderer := NewRenderer(width, height)
	output := renderer.RenderModel(model, rotX, rotY, rotZ, style)
	
	return output, nil
}