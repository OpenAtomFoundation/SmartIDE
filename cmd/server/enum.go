/*
 * @Date: 2022-03-08 17:27:09
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-04-21 15:05:46
 * @FilePath: /smartide-cli/cmd/server/enum.go
 */
package server

type FeedbackCommandEnum string

const (
	FeedbackCommandEnum_Start           FeedbackCommandEnum = "start"
	FeedbackCommandEnum_Stop            FeedbackCommandEnum = "stop"
	FeedbackCommandEnum_Remove          FeedbackCommandEnum = "remove"
	FeedbackCommandEnum_RemoveContainer FeedbackCommandEnum = "remove_container"
	FeedbackCommandEnum_New             FeedbackCommandEnum = "new"
)
