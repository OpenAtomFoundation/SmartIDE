package user

import (
	"os/user"
	"runtime"
)

type ResUser struct {
	Uid string
	Gid string
}

func GetUserInfo() *ResUser {
	resuser := &ResUser{}
	resuser.Uid = "1000"
	resuser.Gid = "1000"

	u, err := user.Current()
	if err != nil {
		return resuser
	} else {
		switch runtime.GOOS {
		case "linux", "darwin":
			return &ResUser{
				Uid: u.Uid,
				Gid: u.Gid}
		case "windows":
			return resuser
		default:
			return resuser
		}
	}

}
