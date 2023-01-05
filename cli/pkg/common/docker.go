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

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/system"
	"github.com/docker/go-connections/nat"
)

type Docker struct {
	client *client.Client
}

func NewDocker(c *client.Client) *Docker {
	return &Docker{
		client: c,
	}
}

func (d Docker) Pull(ctx context.Context, image, user, password string) (string, error) {
	authConfig := types.AuthConfig{
		Username: user,
		Password: password,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}

	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	reader, err := d.client.ImagePull(ctx, image, types.ImagePullOptions{RegistryAuth: authStr})
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(reader)
	return buf.String(), nil
}

func (d Docker) Import(ctx context.Context, file string, image string) (string, error) {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return "", err
	}

	reader, err := d.client.ImageImport(ctx, types.ImageImportSource{Source: f, SourceName: "-"}, image, types.ImageImportOptions{})
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(reader)
	return buf.String(), nil
}

func (d Docker) Run(ctx context.Context, name string, image string, cmd []string, volumes map[string]struct{}, ports nat.PortSet) error {
	resp, err := d.client.ContainerCreate(ctx, &container.Config{Image: image, Volumes: volumes, ExposedPorts: ports, Cmd: cmd}, nil, nil, nil, name)
	if err != nil {
		return err
	}
	if err = d.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}
	return nil
}

func (d Docker) Copy(ctx context.Context, file string, dest string, container string) error {
	srcPath := file
	dstPath := dest
	// Prepare destination copy info by stat-ing the container path.
	dstInfo := archive.CopyInfo{Path: dstPath}
	dstStat, err := d.client.ContainerStatPath(ctx, container, dstPath)

	// If the destination is a symbolic link, we should evaluate it.
	if err == nil && dstStat.Mode&os.ModeSymlink != 0 {
		linkTarget := dstStat.LinkTarget
		if !system.IsAbs(linkTarget) {
			// Join with the parent directory.
			dstParent, _ := archive.SplitPathDirEntry(dstPath)
			linkTarget = filepath.Join(dstParent, linkTarget)
		}

		dstInfo.Path = linkTarget
		dstStat, err = d.client.ContainerStatPath(ctx, container, linkTarget)
	}

	if err == nil {
		dstInfo.Exists, dstInfo.IsDir = true, dstStat.Mode.IsDir()
	}

	var (
		content         io.Reader
		resolvedDstPath string
	)

	// Prepare source copy info.
	srcInfo, err := archive.CopyInfoSourcePath(srcPath, true)
	if err != nil {
		return err
	}

	srcArchive, err := archive.TarResource(srcInfo)
	if err != nil {
		return err
	}
	defer srcArchive.Close()

	dstDir, preparedArchive, err := archive.PrepareArchiveCopy(srcArchive, srcInfo, dstInfo)
	if err != nil {
		return err
	}
	defer preparedArchive.Close()

	resolvedDstPath = dstDir
	content = preparedArchive

	options := types.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
	}
	return d.client.CopyToContainer(ctx, container, resolvedDstPath, content, options)
}

func (d Docker) Start(ctx context.Context, container string) error {
	err := d.client.ContainerStart(ctx, container, types.ContainerStartOptions{})
	return err
}

func (d Docker) Stop(ctx context.Context, container string) error {
	timeout := time.Second * 5
	err := d.client.ContainerStop(ctx, container, &timeout)
	return err
}

func (d Docker) Rm(ctx context.Context, container string, force bool) error {
	err := d.client.ContainerRemove(ctx, container, types.ContainerRemoveOptions{Force: force})
	return err
}

func (d Docker) Push(ctx context.Context, image string, user string, password string) (string, error) {
	authConfig := types.AuthConfig{
		Username: user,
		Password: password,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}

	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	reader, err := d.client.ImagePush(ctx, image, types.ImagePushOptions{RegistryAuth: authStr})
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	return buf.String(), err
}

func (d Docker) Exec(ctx context.Context, container string, chdir string, cmd []string, env []string) (string, error) {
	id, err := d.client.ContainerExecCreate(ctx, container, types.ExecConfig{Tty: true, WorkingDir: chdir, Cmd: cmd, Env: env, AttachStderr: true, AttachStdout: true})
	if err != nil {
		return "", err
	}
	resp, err := d.client.ContainerExecAttach(ctx, id.ID, types.ExecStartCheck{})
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Reader)
	return buf.String(), err
}

func (d Docker) Restart(ctx context.Context, container string) error {
	_ = d.Stop(ctx, container)
	return d.Start(ctx, container)
}

func (d Docker) IsRun(ctx context.Context, container string) bool {
	stat, err := d.client.ContainerInspect(ctx, container)
	if err != nil {
		return false
	}
	if !stat.State.Running {
		return false
	}
	return true
}
