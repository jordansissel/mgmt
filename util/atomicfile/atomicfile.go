// Mgmt
// Copyright (C) James Shubin and the project contributors
// Written by James Shubin <james@shubin.ca> and the project contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//
// Additional permission under GNU GPL version 3 section 7
//
// If you modify this program, or any covered work, by linking or combining it
// with embedded mcl code and modules (and that the embedded mcl code and
// modules which link with this program, contain a copy of their source code in
// the authoritative form) containing parts covered by the terms of any other
// license, the licensors of this program grant you additional permission to
// convey the resulting work. Furthermore, the licensors of this program grant
// the original author, James Shubin, additional permission to update this
// additional permission if he deems it necessary to achieve the goals of this
// additional permission.

// Package atomicfile contains a mechanism to write ore replace a given filename
// with a single, atomic operation.
package atomicfile

import (
	"os"
)

// AtomicFile is an os.File-like struct which allows you to write data and
// modify metadata before committing it to a specific filename.
//
// The intended use is to ensure complete files are only ever available at the
// given path.
//
// If, for any reason, a Commit() never occurs, this module attempts to ensure
// that no temporary files are left around.
//
// Thus:
// * If the program crashes before a Commit(), the temporary file shall be
//   lost (on purpose).
// * If the program calls Close() without a Commit(), the temporary file and
//   its writes shall be lost (on purpose).
// * If the computer fails before a Commit(), the temporary file shall be
//   lost (on purpose).
//
// For any of the above scenarios, no disk space should be occupied by these
// lost files.
//
// This module focuses on delivering "whole files" and is not concerned with
// durability or serializability of operations.
type AtomicFile struct {
	os.File
	path string
}

// New opens an unnamed temporary file.
//
// If you wish to commit writes to the given fspath, you must call Commit().
func New(fspath string) (*AtomicFile, error) {
	f, err := open(fspath)

	if err != nil {
		return nil, err
	}

	return &AtomicFile{File: *f, path: fspath}, nil
}

// Commit acts upon the file to atomically replace it on disk.
// This closes the file and further changes should not be attempted.
func (obj *AtomicFile) Commit() error {
	err := obj.commit()

	// Close after attempting to commit, even if the commit fails.
	// Nothing can be done so we should still close the file
	_ = obj.Close()

	return err
}
