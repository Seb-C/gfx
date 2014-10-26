// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

import (
	"fmt"
	"image"
)

// Batch merges all of the given objects into a single one (representing the batch). It
// panics if there are no arguments or if the objects do not share the same exact:
//
//  State
//  *Shader
//  []*Texture
//
// Objects whose meshes make use of independent data slices may not be batched,
// or else a panic will occur. I.e. a mesh that has vertex colors cannot be
// batched with a mesh that does not have vertex colors.
//
func Batch(objs ...*Object) *Object {
	// If there are no objects to batch, panic.
	if len(objs) == 0 {
		panic("Batch: no arguments")
	}

	// Lock each object, and each of their meshes, for reading.
	for _, obj := range objs {
		obj.RLock()
		defer obj.RUnlock()
		for _, mesh := range obj.Meshes {
			mesh.RLock()
			defer mesh.RUnlock()
		}
	}

	// Create a new batch object, with the same state, shader, and textures.
	batchMesh := NewMesh()
	batch := NewObject()
	batch.State = objs[0].State
	batch.Shader = objs[0].Shader
	batch.Meshes = []*Mesh{batchMesh}

	// Copy the textures over.
	batch.Textures = make([]*Texture, len(objs[0].Textures))
	copy(batch.Textures, objs[0].Textures)

	// Merge each object into the batch object.
	for _, obj := range objs {
		// Append every mesh of the object to the batch mesh.
		for _, mesh := range obj.Meshes {
			if err := batchMesh.append(mesh); err != nil {
				panic(fmt.Sprintf("Batch: %v", err))
			}
		}
	}
	return batch
}

// A batch represents a single batch of a single type, and all of the objects
// within that batch. It is only used by the Batcher type.
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

	// The mesh type of this batch. Only graphics objects whose meshes have
	// exactly this mesh type can be added to this batch.
	meshType *meshType

	// The graphics objects residing in this batch.
	objects []*Object
}

// matches tells if the type of this batch matches the given object's type. If
// so, the object can be safely added to this batch.
func (b *batch) matches(obj *Object) bool {
	obj.RLock()
	defer obj.RUnlock()
	if obj.Shader != b.shaderType {
		// Object does not share this batch's shader type.
		return false
	}
	if len(obj.Textures) != len(b.textureType) {
		// Object does not share this batch's texture type.
		return false
	}
	for i, tex := range b.textureType {
		if obj.Textures[i] != tex {
			// Object does not share this batch's texture type.
			return false
		}
	}
	if obj.State != b.stateType {
		// Object does not share the this batch's state type.
		return false
	}
	for _, m := range obj.Meshes {
		m.RLock()
		err := newMeshType(m).equals(*b.meshType)
		m.RUnlock()
		if err != nil {
			// Object does not share this batch's mesh type.
			return false
		}
	}
	// Object is a perfect match for this batch.
	return true
}

// Batcher builds batches out of objects automatically.
type Batcher struct {
	// The slice of all the batches the batcher currently has.
	batches []*batch

	// A map of batches by object pointer. This allows us to identify if this
	// batcher already contains a given object (without searching every batch).
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
			continue
		}

		// The batcher does not contain the object already.
		bt = b.findBatch(obj)
		if bt == nil {
			// No batch exists for the object, create a new one.
			b.newBatch(obj)
			continue
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
		b.removeFromBatch(obj, bt)
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
			// The batcher does not contain this object, create a new batch for
			// the object then.
			b.newBatch(obj)
			continue
		}

		// Would we still have the object placed into that batch?
		if !bt.matches(obj) {
			// The batch we would place the object into is not the one it is
			// currently residing in. Remove the object from the old batch, add
			// it to the new one.
			b.removeFromBatch(obj, bt)

			// Find an existing batch for the object to go into.
			wantBatch := b.findBatch(obj)
			if wantBatch != nil {
				// Add the object to the existing batch.
				b.addToBatch(obj, wantBatch)
				continue
			}

			// No existing batch for the object, create a new one for it.
			b.newBatch(obj)
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
	for _, bt := range b.batches {
		// Special case: an object with a nil mesh type must have all of it's
		// object's drawn independently (i.e. not batched).
		if bt.meshType == nil {
			for _, obj := range bt.objects {
				c.Draw(r, obj, cam)
			}
			continue
		}

		// If the batch's object is nil, then all of the objects in the batch
		// need to be merged together to form the object (that will then be
		// drawn).
		if bt.Object == nil {
			bt.Object = Batch(bt.objects...)
		}

		// Draw the batch.
		c.Draw(r, bt.Object, cam)
	}
}

// newBatch creates a new batch for the given type of object. The returned
// batch will have the given object appended to it already, and the internal
// map of batches-by-object will be updated.
//
// This function properly read-lock's the object as needed.
func (b *Batcher) newBatch(obj *Object) {
	// Create a new batch with the object's type.
	bt := &batch{
		stateType:  obj.State,
		shaderType: obj.Shader,
		objects:    []*Object{obj},
	}

	obj.RLock()
	defer obj.RUnlock()

	// Store the mesh type of the object.
	if len(obj.Meshes) > 0 {
		// Grab the first mesh's mesh type.
		first := obj.Meshes[0]
		first.RLock()
		meshType := newMeshType(first)
		bt.meshType = &meshType
		first.RUnlock()

		// We must handle an unfortunate case: what if there exist multiple
		// meshes in an object, each of which has a different mesh type?
		//
		// If this happens we give the batch a nil meshType, which signifies
		// this unfortunate circumstance. If a batch has a nil mesh type, it
		// has each of it's object's drawn independently.
		for _, mesh := range obj.Meshes {
			mesh.RLock()
			mt := newMeshType(mesh)
			mesh.RUnlock()
			if err := mt.equals(meshType); err != nil {
				// The object has mesh's that are not of the same mesh type.
				bt.meshType = nil
				break
			}
		}
	}

	// We explicitly copy the textures slice so that changes to obj by the user
	// do not affect which type of objects the batch can hold.
	bt.textureType = make([]*Texture, len(obj.Textures))
	copy(bt.textureType, obj.Textures)

	// Add the batch to the list of batches in use by the batcher.
	b.batches = append(b.batches, bt)

	// Update the internal map.
	b.batchByObj[obj] = bt
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

// removeFromBatch removes the given object from the given batch's slice of
// objects. It also clears the batch such that a merge of the batch's objects
// is required again (to account for the removed object).
//
// If the batch only contains the given object (to be removed) then the batch
// itself is removed as well.
func (b *Batcher) removeFromBatch(obj *Object, bt *batch) {
	// If the batch literally only has one object, the one to be removed, then
	// we just remove the batch itself.
	if len(bt.objects) == 1 && bt.objects[0] == obj {
		for i, batch := range b.batches {
			if bt != batch {
				// It's not this batch.
				continue
			}
			b.batches = append(b.batches[:i], b.batches[i+1:]...)
		}
		return
	}

	// Find the object and remove it from the batch.
	for i, batchObj := range bt.objects {
		if obj != batchObj {
			// It's not this object.
			continue
		}
		bt.objects = append(bt.objects[:i], bt.objects[i+1:]...)
	}

	// Clear the batch, so that it will be recreated (to account for the
	// removed object) at the next draw.
	bt.Object = nil
}

// findBatch finds the appropriate batch to place the given object into. If no
// such batch currently exists, nil is returned.
func (b *Batcher) findBatch(obj *Object) *batch {
	// We expect that most objects within a single batcher will be similar --
	// making a linear search for the correct batch here not too slow.
	for _, batch := range b.batches {
		if batch.matches(obj) {
			// The batch is an appropriate match for this type of object.
			return batch
		}
	}
	return nil
}

// NewBatcher returns a new and initialized batcher with the given objects
// added to it.
func NewBatcher(objs ...*Object) *Batcher {
	b := &Batcher{
		batchByObj: make(map[*Object]*batch, len(objs)),
	}
	b.Add(objs...)
	return b
}
