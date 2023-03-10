/*
SmartIDE - CLI
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
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func publicKeyAuthFunc(pemBytes, keyPassword []byte) ssh.AuthMethod {
	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKeyWithPassphrase(pemBytes, keyPassword)
	if err != nil {
		logrus.WithError(err).Error("parse ssh key from bytes failed")
		return nil
	}
	return ssh.PublicKeys(signer)
}

func SshRemoteRunCommand(sshClient *ssh.Client, command string) (string, error) {
	session, err := sshClient.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	var buf bytes.Buffer
	session.Stdout = &buf
	err = session.Run(command)
	logString := buf.String()
	//logrus.WithField("CMD:", command).Info(logString)
	if err != nil {
		return logString, fmt.Errorf("CMD: %s  OUT: %s  ERROR: %s", command, logString, err)
	}
	return logString, err
}
func NewSshClientConfig(sshUser, sshPassword, sshType, sshKey, sshKeyPassword string) (config *ssh.ClientConfig, err error) {
	if sshUser == "" {
		return nil, errors.New("ssh_user can not be empty")
	}
	config = &ssh.ClientConfig{
		Timeout: time.Second * 3,
		User:    sshUser,
		//HostKeyCallback: ssh.InsecureIgnoreHostKey(), //这个可以， 但是不够安全
		//HostKeyCallback: hostKeyCallBackFunc(h.Host),
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// use OpenSSH's known_hosts file if you care about host validation
			return nil
		},
	}
	switch sshType {
	case "password":
		config.Auth = []ssh.AuthMethod{ssh.Password(sshPassword)}
	case "key":
		config.Auth = []ssh.AuthMethod{publicKeyAuthFunc([]byte(sshKey), []byte(sshKeyPassword))}
	default:
		return nil, fmt.Errorf("unknow ssh auth type: %s", sshType)
	}
	return
}
func CreateSimpleSshClient(sshUser, sshPassword, sshAddr string) (*ssh.Client, error) {
	if !strings.Contains(sshAddr, ":") {
		sshAddr = fmt.Sprintf("%s:22", sshAddr)
	}

	targetConfig, err := NewSshClientConfig(sshUser, sshPassword, "password", "", "")
	if err != nil {
		return nil, fmt.Errorf("cluster jumper proxy ssh config failed:%s", err)
	}
	return ssh.Dial("tcp", sshAddr, targetConfig)
}
