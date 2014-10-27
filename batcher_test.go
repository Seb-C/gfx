// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

import (
	"image"
	"math/rand"
	"testing"

	"azul3d.org/lmath.v1"
)

// randObject returns a *Object with a few random properties.
func randObject() *Object {
	rbool := func() bool {
		return rand.Float64() > .5
	}

	similar := rbool()
	obj := NewObject()
	obj.Meshes = []*Mesh{NewMesh()}
	if !similar {
		obj.State.Dithering = rbool()
		obj.Transform.SetPos(lmath.Vec3{rand.Float64(), rand.Float64(), rand.Float64()})
		obj.Textures = make([]*Texture, 2)
	}
	return obj
}

// nRandObjects returns a slice of n random objects, retrieved from randObject.
func nRandObjects(n int) []*Object {
	objs := make([]*Object, n)
	for i := range objs {
		objs[i] = randObject()
	}
	return objs
}

var rand1K = nRandObjects(1000)

// This benchmark creates a batch of 1k random objects and removes, then adds a
// random one every b.N iteration. Remove/Add operates identically to Update,
// but Update should be faster.
func BenchmarkBatchRmAdd1k(b *testing.B) {
	batcher := NewBatcher(rand1K...)
	nilRenderer := Nil()
	for i := 0; i < b.N; i++ {
		batcher.DrawTo(nilRenderer, image.Rect(0, 0, 0, 0), nil)
		obj := rand1K[i%len(rand1K)]
		batcher.Remove(obj)
		batcher.Add(obj)
	}
}

// This benchmark creates a batch of 1k random objects and updates a random one
// every b.N iteration. Update is identical to, but faster than, Remove/Add.
func BenchmarkBatchUpdate1k(b *testing.B) {
	batcher := NewBatcher(rand1K...)
	nilRenderer := Nil()
	for i := 0; i < b.N; i++ {
		batcher.DrawTo(nilRenderer, image.Rect(0, 0, 0, 0), nil)
		obj := rand1K[i%len(rand1K)]
		batcher.Update(obj)
	}
}
