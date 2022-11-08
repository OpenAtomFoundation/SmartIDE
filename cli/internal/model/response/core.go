/*
 * @Date: 2022-11-03 14:31:26
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-11-03 14:31:26
 * @FilePath: /cli/internal/model/response/core.go
 */

package response

import "time"

type GVA_MODEL struct {
	ID        uint      // 主键ID
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 更新时间
}

type DefaultResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}
