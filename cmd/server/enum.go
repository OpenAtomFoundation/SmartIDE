/*
 * @Date: 2022-03-08 17:27:09
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-03-16 15:04:49
 * @FilePath: /smartide-cli/cmd/server/enum.go
 */
package server

type FeedbackCommandEnum string

const (
	FeedbackCommandEnum_Start           FeedbackCommandEnum = "start"
	FeedbackCommandEnum_Stop            FeedbackCommandEnum = "stop"
	FeedbackCommandEnum_Remove          FeedbackCommandEnum = "remove"
	FeedbackCommandEnum_RemoveContainer FeedbackCommandEnum = "remove_container"
)
