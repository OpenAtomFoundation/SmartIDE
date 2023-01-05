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

package compose

import (
	"fmt"
	"testing"
)

func TestSecretSimple(t *testing.T) {
	tests := []struct {
		item       string
		wantSource string
		wantErr    bool
	}{
		{item: "my_secret", wantSource: "my_secret", wantErr: false},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			// MarshalYaml
			if !tt.wantErr {
				item := SecretSimple{Source: tt.wantSource}
				result, _ := item.MarshalYAML()
				content := fmt.Sprintf("%s", result)
				if content != tt.item {
					t.Logf("%d %d", len(content), len(tt.item))
					t.Errorf("SecretSimple.MarshalYAML() content = %v, wantContent %v", content, tt.item)
					return
				}
			}
			// UnmarshalYaml
			var item SecretSimple
			err := UnmarshalYaml(tt.item, &item)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecretSimple.UnarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if item.Source != tt.wantSource {
				t.Errorf("SecretSimple.UnarshalYAML() source = %v, wantSouce %v", item.Source, tt.wantSource)
				return
			}
		})
	}
}
