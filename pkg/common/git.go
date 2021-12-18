/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors:
 * @LastEditTime:
 */
package common

import (
	"regexp"
)

func CheckGitRemoteUrl(url string) bool {
	pattern := `((git|ssh|http(s)?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)(/)?`

	match, _ := regexp.MatchString(pattern, url)
	return match

	/* rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{url},
	})

	// We can then use every Remote functions to retrieve wanted information
	_, err := rem.List(&git.ListOptions{})

	return err == nil */
}
