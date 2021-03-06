// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

import (
	"image/color"
	"testing"
)

func TestNilRenderer(t *testing.T) {
	// Create a nil renderer.
	r := Nil()

	// Create a slice of 50 objects.
	var objects []*Object
	for i := 0; i < 50; i++ {
		objects = append(objects, NewObject())
	}

	// Create a camera
	cam := NewCamera()

	// Convert color.White to our floating-point color model.
	white := ColorModel.Convert(color.White).(Color)

	// Render 30 frames.
	for frame := 0; frame < 30; frame++ {
		rect := r.Bounds()

		// Clear a rectangle on the drawable.
		r.Clear(rect, white)

		// Draw each object on the rectangle of the drawable.
		for _, obj := range objects {
			r.Draw(rect, obj, cam)
		}

		// Render the frame.
		r.Render()
	}
}
