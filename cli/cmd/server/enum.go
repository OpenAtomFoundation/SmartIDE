/*
 * @Date: 2022-03-08 17:27:09
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-09-05 22:40:03
 * @FilePath: /cli/cmd/server/enum.go
 */
package server

type FeedbackCommandEnum string

const (
	FeedbackCommandEnum_Start           FeedbackCommandEnum = "start"
	FeedbackCommandEnum_Stop            FeedbackCommandEnum = "stop"
	FeedbackCommandEnum_Remove          FeedbackCommandEnum = "remove"
	FeedbackCommandEnum_RemoveContainer FeedbackCommandEnum = "remove_container"
	FeedbackCommandEnum_New             FeedbackCommandEnum = "new"
	FeedbackCommandEnum_K8S             FeedbackCommandEnum = "k8s"
	FeedbackCommandEnum_ApplySSH        FeedbackCommandEnum = "applyssh"
	FeedbackCommandEnum_Ingress         FeedbackCommandEnum = "ingress"
)
