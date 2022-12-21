/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-12-21 15:47:10
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
