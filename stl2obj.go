package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func convertSTLToOBJ(stlPath, objPath string) error {
	// Parse the STL file
	stlModel, err := ParseSTL(stlPath)
	if err != nil {
		return err
	}

	// Create OBJ file
	objFile, err := os.Create(objPath)
	if err != nil {
		return err
	}
	defer objFile.Close()

	// Write OBJ header
	objFile.WriteString("# Generated from STL by MakeMe\n")
	objFile.WriteString("# Object: " + stlModel.Name + "\n\n")

	// Extract unique vertices from triangles
	vertexMap := make(map[string]int)
	vertices := []Vertex{}
	vertexIndex := 1

	for _, tri := range stlModel.Triangles {
		for _, v := range []Vertex{tri.V1, tri.V2, tri.V3} {
			key := fmt.Sprintf("%.6f,%.6f,%.6f", v.X, v.Y, v.Z)
			if _, exists := vertexMap[key]; !exists {
				vertexMap[key] = vertexIndex
				vertices = append(vertices, v)
				vertexIndex++
			}
		}
	}

	// Write vertices
	for _, v := range vertices {
		objFile.WriteString(fmt.Sprintf("v %.6f %.6f %.6f\n", v.X, v.Y, v.Z))
	}

	objFile.WriteString("\n")

	// Write faces
	for _, tri := range stlModel.Triangles {
		v1Key := fmt.Sprintf("%.6f,%.6f,%.6f", tri.V1.X, tri.V1.Y, tri.V1.Z)
		v2Key := fmt.Sprintf("%.6f,%.6f,%.6f", tri.V2.X, tri.V2.Y, tri.V2.Z)
		v3Key := fmt.Sprintf("%.6f,%.6f,%.6f", tri.V3.X, tri.V3.Y, tri.V3.Z)

		i1 := vertexMap[v1Key]
		i2 := vertexMap[v2Key]
		i3 := vertexMap[v3Key]

		objFile.WriteString(fmt.Sprintf("f %d %d %d\n", i1, i2, i3))
	}

	return nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s input.stl output.obj\n", os.Args[0])
		os.Exit(1)
	}

	stlPath := os.Args[1]
	objPath := os.Args[2]

	if !strings.HasSuffix(strings.ToLower(stlPath), ".stl") {
		fmt.Fprintf(os.Stderr, "Input file must be a .stl file\n")
		os.Exit(1)
	}

	if !strings.HasSuffix(strings.ToLower(objPath), ".obj") {
		fmt.Fprintf(os.Stderr, "Output file must be a .obj file\n")
		os.Exit(1)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(objPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	if err := convertSTLToOBJ(stlPath, objPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error converting STL to OBJ: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully converted %s to %s\n", stlPath, objPath)
}