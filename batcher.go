// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

import "image"

// Batch merges all of the given objects into a single one (representing the batch). It
// panics if TODO (the objects do not share the same exact shader, textures, etc).
func Batch(objs ...*gfx.Object) *gfx.Object {
}

// Batcher builds batches out of objects automatically.
type Batcher struct{}

// Add adds the given objects to the batcher.
func (b *Batcher) Add(objs ...*Object) {
	// TODO(slimsag): add the objects.
}

// Remove removes the given objects from the batcher.
func (b *Batcher) Remove(objs ...*Object) {
	// TODO(slimsag): remove the objects.
}

// Update marks all of the given objects as updated. Batches containing these
// objects will be rebuilt upon the next time this batcher is drawn to a
// canvas. This is equivalent to (but faster than) writing:
//
//  b.Remove(objs...)
//  b.Add(objs...)
//
func (b *Batcher) Update(objs ...*Object) {
	// TODO(slimsag): remove the objects.
}

// DrawTo draws all of the objects in this batcher to the given rectangle of
// the canvas, as seen by the given camera.
//
// If any objects in the batcher have been updated since the last call to this
// method, then the batches will be rebuilt and then drawn to the canvas.
func (b *Batcher) DrawTo(c Canvas, r image.Rectangle, cam *Camera) {
	// TODO(slimsag): draw the batches to the canvas.
}
