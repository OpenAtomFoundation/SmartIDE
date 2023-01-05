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

package common

import "testing"

func Test_gitOperation_GetRepositoryUrl(t *testing.T) {
	SmartIDELog.InitLogger("debug")

	type args struct {
		actualGitRepoUrl string
	}
	tests := []struct {
		name string
		g    gitOperation
		args args
		want string
	}{
		{"ssh", gitOperation{}, args{"git@github.com:idcf-boat-house/boathouse-calculator.git"}, ""},
		{"https", gitOperation{}, args{"https://github.com/idcf-boat-house/boathouse-calculator.git"}, "https://github.com/idcf-boat-house/boathouse-calculator"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gitOperation{}
			if got := g.GetRepositoryUrl(tt.args.actualGitRepoUrl); got != tt.want {
				t.Errorf("gitOperation.GetRepositoryUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}
