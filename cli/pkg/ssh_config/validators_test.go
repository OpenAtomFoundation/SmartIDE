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
	"testing"
)

var validateTests = []struct {
	key string
	val string
	err string
}{
	{"IdentitiesOnly", "yes", ""},
	{"IdentitiesOnly", "Yes", `ssh_config: value for key "IdentitiesOnly" must be 'yes' or 'no', got "Yes"`},
	{"Port", "22", ``},
	{"Port", "yes", `ssh_config: strconv.ParseUint: parsing "yes": invalid syntax`},
}

func TestValidate(t *testing.T) {
	for _, tt := range validateTests {
		err := validate(tt.key, tt.val)
		if tt.err == "" && err != nil {
			t.Errorf("validate(%q, %q): got %v, want nil", tt.key, tt.val, err)
		}
		if tt.err != "" {
			if err == nil {
				t.Errorf("validate(%q, %q): got nil error, want %v", tt.key, tt.val, tt.err)
			} else if err.Error() != tt.err {
				t.Errorf("validate(%q, %q): got err %v, want %v", tt.key, tt.val, err, tt.err)
			}
		}
	}
}

func TestDefault(t *testing.T) {
	if v := Default("VisualHostKey"); v != "no" {
		t.Errorf("Default(%q): got %v, want 'no'", "VisualHostKey", v)
	}
	if v := Default("visualhostkey"); v != "no" {
		t.Errorf("Default(%q): got %v, want 'no'", "visualhostkey", v)
	}
	if v := Default("notfound"); v != "" {
		t.Errorf("Default(%q): got %v, want ''", "notfound", v)
	}
}
