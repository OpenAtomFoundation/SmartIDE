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

package model

type Config struct {
	Auths []Auth `json:"auths"`
}

type Auth struct {
	UserName   string      `yaml:"username" json:"username"`
	Token      interface{} `yaml:"token" json:"token"`
	LoginUrl   string      `yaml:"login_url" json:"login_url"`
	CurrentUse bool        `yaml:"current_use" json:"current_use"`
}

func (auth Auth) IsNil() bool {
	return auth.Token == "" || auth.LoginUrl == ""
}

func (auth Auth) IsNotNil() bool {
	return !auth.IsNil()
}
