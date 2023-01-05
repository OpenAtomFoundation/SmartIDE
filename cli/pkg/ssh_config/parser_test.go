/*
SmartIDE - Dev Containers
Copyright (C) 2023 leansoftX.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package ssh_config

import (
	"errors"
	"testing"
)

type errReader struct {
}

func (b *errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error occurred")
}

func TestIOError(t *testing.T) {
	buf := &errReader{}
	_, err := Decode(buf)
	if err == nil {
		t.Fatal("expected non-nil err, got nil")
	}
	if err.Error() != "read error occurred" {
		t.Errorf("expected read error msg, got %v", err)
	}
}
