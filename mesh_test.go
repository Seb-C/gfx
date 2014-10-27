// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

import "testing"

func TestMeshType(t *testing.T) {
	a := NewMesh()
	a.Vertices = make([]Vec3, 3)

	b := NewMesh()
	b.Vertices = make([]Vec3, 3)
	b.Colors = make([]Color, 3)

	at := newMeshType(a)
	bt := newMeshType(b)
	if err := at.equals(bt); err == nil {
		t.Fatal("mesh types are incorrectly equal")
	}
}

var meshAppendTests = []struct {
	name                            string
	a, b, want                      []Vec3
	aIndices, bIndices, wantIndices []uint32
}{
	{
		name: "append(Mesh, Mesh)",
		a:    []Vec3{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}},
		b:    []Vec3{{8, 7, 6}, {5, 4, 3}, {2, 1, 0}},
		want: []Vec3{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, {8, 7, 6}, {5, 4, 3}, {2, 1, 0}},
	}, {
		name:        "append(IndexedMesh, IndexedMesh)",
		a:           []Vec3{{0, 0, 0}, {1, 1, 1}, {2, 2, 2}},
		b:           []Vec3{{3, 3, 3}, {4, 4, 4}, {5, 5, 5}},
		want:        []Vec3{{0, 0, 0}, {1, 1, 1}, {2, 2, 2}, {3, 3, 3}, {4, 4, 4}, {5, 5, 5}},
		aIndices:    []uint32{0, 1, 2},
		bIndices:    []uint32{2, 2, 1},
		wantIndices: []uint32{0, 1, 2, 5, 5, 4},
	}, {
		name:     "append(Mesh, IndexedMesh)",
		a:        []Vec3{{0, 0, 0}, {1, 1, 1}, {2, 2, 2}},
		b:        []Vec3{{3, 3, 3}, {4, 4, 4}, {5, 5, 5}},
		want:     []Vec3{{0, 0, 0}, {1, 1, 1}, {2, 2, 2}, {5, 5, 5}, {4, 4, 4}, {3, 3, 3}},
		bIndices: []uint32{2, 1, 0},
	}, {
		name:        "append(IndexedMesh, Mesh)",
		a:           []Vec3{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}},
		b:           []Vec3{{8, 7, 6}, {5, 4, 3}, {2, 1, 0}},
		want:        []Vec3{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, {8, 7, 6}, {5, 4, 3}, {2, 1, 0}},
		aIndices:    []uint32{2, 1, 0},
		wantIndices: []uint32{2, 1, 0, 3, 4, 5},
	},
}

func TestMeshAppend(t *testing.T) {
	for caseNumber, tst := range meshAppendTests {
		// Create the meshes.
		a := NewMesh()
		a.Vertices = tst.a
		a.Indices = tst.aIndices

		b := NewMesh()
		b.Vertices = tst.b
		b.Indices = tst.bIndices

		// Check if we can append the meshes safely:
		if err := a.canAppend(b); err != nil {
			t.Logf("Case %d: %s\n", caseNumber, tst.name)
			t.Fatal(err)
		}

		// Append mesh b to mesh a.
		a.append(b)

		// Validate the vertices slices.
		if len(tst.want) != len(a.Vertices) {
			t.Logf("Case %d: %s\n", caseNumber, tst.name)
			t.Fatal("got", len(tst.want), "vertices, want", len(a.Vertices), "vertices")
		}
		for i, v := range tst.want {
			if a.Vertices[i] != v {
				t.Logf("Case %d: %s\n", caseNumber, tst.name)
				t.Log("got Vertices: ", a.Vertices)
				t.Fatal("want Vertices:", tst.want)
			}
		}

		// Validate the indices slices.
		if len(tst.wantIndices) != len(a.Indices) {
			t.Logf("Case %d: %s\n", caseNumber, tst.name)
			t.Fatal("got", len(a.Indices), "indices, want", len(tst.wantIndices), "indices")
		}
		for i, v := range tst.wantIndices {
			if a.Indices[i] != v {
				t.Logf("Case %d: %s\n", caseNumber, tst.name)
				t.Log("got Indices: ", a.Indices)
				t.Fatal("want Indices:", tst.wantIndices)
			}
		}
	}
}
