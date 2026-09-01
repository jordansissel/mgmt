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

//go:build linux

package atomicfile

import (
	"os"
	"path"

	"golang.org/x/sys/unix"
)

func open(fspath string) (*os.File, error) {
	// Note: this only works on certain filesystems on Linux. See the `O_TMPFILE`
	// documentation on `open(2)` for more details.
	//
	// Relevant notes from Linux's open(2) are:
	//
	// > Create an unnamed temporary regular file.  The path
	// > argument specifies a directory; an unnamed inode will be
	// > created in that directory's filesystem.
	return os.OpenFile(path.Dir(fspath), unix.O_RDWR|unix.O_TMPFILE, 0600)
}

func (obj *AtomicFile) commit() error {
	temp, err := os.CreateTemp(path.Dir(obj.path), path.Base(obj.path)+".new*")
	if err != nil {
		return err
	}
	defer temp.Close()

	err = os.Remove(temp.Name())
	if err != nil {
		return err
	}
	// Note: There is a race condition here between the Remove and Linkat calls.
	// Above, CreateTemp is used to create a temporary, named file. It is removed
	// in order to link our unnamed file to that pathname.
	//
	// If, between the Remove and Linkat, something else creates a file with the
	// same name, the Linkat call will fail due to the filename already existing..
	//
	// A future improvement would be to try (possibly multiple times) different
	// names to give to our temporary file.

	// Previously, our open() func used O_TMPFILE to create an unnamed file, and
	// in order to rename(2) to deliver our committed file, we must give our
	// unnamed file a name!
	// See linkat(2) and open(2)'s O_TMPFILE documentation on Linux for more information.
	err = unix.Linkat(int(obj.Fd()), "", unix.AT_FDCWD, temp.Name(), unix.AT_EMPTY_PATH)
	if err != nil {
		return err
	}
	defer obj.Close()

	return os.Rename(temp.Name(), obj.path)
}
