// Copyright (C) xooooooox

package main

import (
	"github.com/xooooooox/sea"
)

// Insert upload files
func Insert(insert ...interface{}) error {
	_, err := sea.Add(insert...)
	return err
}
