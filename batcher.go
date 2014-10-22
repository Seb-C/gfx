// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

import "image"

// Batch merges all of the given objects into a single one (representing the batch). It
// panics if TODO (the objects do not share the same exact shader, textures, etc).
func Batch(objs ...*Object) *Object {
	return nil
}

type batch struct {
	// The merged objects, or nil if the objects need to be merged.
	*Object

	// The graphics state type of this batch. Only graphics objects with
	// exactly this graphics state can be added to this batch.
	stateType State

	// The shader type of this batch. Only graphics objects with exactly this
	// shader can be added to this batch.
	shaderType *Shader

	// The texture type of this batch. Only graphics objects with exactly this
	// texture set can be added to this batch.
	textureType []*Texture

	// The graphics objects residing in this batch.
	objects []*Object
}

// remove removes the given object from this batch's slice of objects. It also
// clears the batch such that a merge of the batch's objects is required again
// (to account for the removed object).
func (b *batch) remove(obj *Object) {
	for i, batchObj := range b.objects {
		if obj != batchObj {
			// It's not this object.
			continue
		}
		b.objects = append(b.objects[:i], b.objects[i+1:]...)
	}

	// Clear the batch, so that it will be recreated (to account for the
	// removed object) at the next draw.
	b.Object = nil
}

// Batcher builds batches out of objects automatically.
type Batcher struct {
	batches    []*batch
	batchByObj map[*Object]*batch
}

// Add adds the given objects to the batcher.
func (b *Batcher) Add(objs ...*Object) {
	for _, obj := range objs {
		bt, ok := b.batchByObj[obj]
		if ok {
			// The batcher already contains the object. We don't need to add it
			// again, so instead just clear the batch and continue.
			bt.Object = nil
		}

		// The batcher does not contain the object already.
		bt = b.findBatch(obj)
		if bt == nil {
			// No batch exists for the object, create a new one.
			bt = &batch{
				stateType:   obj.State,
				shaderType:  obj.Shader,
				textureType: make([]*Texture, len(obj.Textures)),
			}
			copy(bt.textureType, obj.Textures)
			b.batches = append(b.batches, bt)
		}

		// Add the object to the batch.
		b.addToBatch(obj, bt)
	}
}

// Remove removes the given objects from the batcher.
func (b *Batcher) Remove(objs ...*Object) {
	for _, obj := range objs {
		// Find the batch associate with the object.
		bt, ok := b.batchByObj[obj]
		if !ok {
			// The batcher does not contain this object, do nothing.
			continue
		}

		// Remove the object from the batch.
		bt.remove(obj)
	}
}

// Update marks all of the given objects as updated. Batches containing these
// objects will be rebuilt upon the next time this batcher is drawn to a
// canvas. This is equivalent to (but faster than) writing:
//
//  b.Remove(objs...)
//  b.Add(objs...)
//
func (b *Batcher) Update(objs ...*Object) {
	for _, obj := range objs {
		// Find the batch associate with the object.
		bt, ok := b.batchByObj[obj]
		if !ok {
			// The batcher does not contain this object, add it then.
			// TODO(slimsag): consider locking problems here when making
			// Batcher safe for concurrent use.
			b.Add(obj)
			continue
		}

		// Would we still have the object placed into that batch?
		wantBatch := b.findBatch(obj)
		if bt != wantBatch {
			// The batch we would place the object into is not the one it is
			// currently residing in. Remove the object from the old batch, add
			// it to the new one.
			bt.remove(obj)
			b.addToBatch(obj, wantBatch)
			continue
		}

		// If we're here then we know the object would still be placed in the
		// same exact batch. All we need to do then is clear the batch so that
		// it will be recreated at the next draw.
		bt.Object = nil
	}
}

// DrawTo draws all of the objects in this batcher to the given rectangle of
// the canvas, as seen by the given camera.
//
// If any objects in the batcher have been updated since the last call to this
// method, then the batches will be rebuilt and then drawn to the canvas.
func (b *Batcher) DrawTo(c Canvas, r image.Rectangle, cam *Camera) {
	// TODO(slimsag): draw the batches to the canvas.
}

// addToBatch adds the given object to the given batch. It appends the object
// to the batch, updates the internal map of batches-by-object, and clears the
// batch so it will be rebuilt to account for the newly added object upon the
// next draw.
func (b *Batcher) addToBatch(obj *Object, bt *batch) {
	// Append the object.
	bt.objects = append(bt.objects, obj)

	// Update the internal map.
	b.batchByObj[obj] = bt

	// Clear the batch, so that it will be merged once again at the next
	// draw.
	bt.Object = nil
}

// findBatch finds the appropriate batch to place the given object into. If no
// such batch currently exists, nil is returned.
func (b *Batcher) findBatch(obj *Object) *batch {
	// We expect that most objects within a single batcher will be similar --
	// making a linear search for the correct batch here not too slow.
	for _, batch := range b.batches {
		if obj.Shader != batch.shaderType {
			// Object does not share this batch's shader type.
			continue
		}
		if len(obj.Textures) != len(batch.textureType) {
			// Object does not share this batch's texture type.
			continue
		}
		for i, tex := range batch.textureType {
			if obj.Textures[i] != tex {
				// Object does not share this batch's texture type.
				continue
			}
		}
		if obj.State != batch.stateType {
			// Object does not share the this batch's state type.
			continue
		}
	}
	return nil
}