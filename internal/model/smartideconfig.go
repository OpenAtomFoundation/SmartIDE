/*
 * @Author: kenan
 * @Date: 2022-02-15 17:56:00
 * @LastEditors: kenan
 * @LastEditTime: 2022-02-16 16:35:15
 * @FilePath: /smartide-cli/internal/model/smartideconfig.go
 * @Description:
 *
 * Copyright (c) 2022 by kenanlu@leansoftx.com, All Rights Reserved.
 */
/*
 * @Author: kenan
 * @Date: 2022-02-15 17:56:00
 * @LastEditors: kenan
 * @LastEditTime: 2022-02-15 17:56:01
 * @FilePath: /smartide-cli/internal/model/smartideconfig.go
 * @Description:
 *
 * Copyright (c) 2022 by kenanlu@leansoftx.com, All Rights Reserved.
 */
package model

type Config struct {
	Auths []Auth `json:"auths"`
}

type Auth struct {
	UserName   string      `yaml:"username"`
	Token      interface{} `yaml:"token"`
	LoginUrl   string      `yaml:"login_url"`
	CurrentUse bool        `yaml:"current_use"`
}
