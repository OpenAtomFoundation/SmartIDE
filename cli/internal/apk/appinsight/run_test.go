/*
 * @Date: 2022-11-03 17:37:47
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-11-03 20:30:38
 * @FilePath: /cli/internal/apk/appinsight/run_test.go
 */
package appinsight

import (
	_ "embed"
	"fmt"
	"testing"
	"time"
)

func TestSetTrack(t *testing.T) {

	now := time.Now()
	fmt.Println(now.Format("2006-01-02 15:04:05"))

	type args struct {
		cmd       string
		version   string
		args      string
		workModel string
		imageName string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "1",
			args: args{cmd: "remove", version: "v0", args: "--lb", workModel: "model-lb", imageName: "image-1"},
		},
		/* {
			name: "2",
			args: args{cmd: "start", version: "v0", args: "--jn", workModel: "model-jn", imageName: "image-2"},
		}, */
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetTrack(tt.args.cmd, tt.args.version, tt.args.args, tt.args.workModel, tt.args.imageName)
		})
	}

	time.Sleep(time.Minute * 10)
}
