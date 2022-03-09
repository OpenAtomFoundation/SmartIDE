package server

type FeedbackCommandEnum string

const (
	FeedbackCommandEnum_Start  FeedbackCommandEnum = "start"
	FeedbackCommandEnum_Stop   FeedbackCommandEnum = "stop"
	FeedbackCommandEnum_Remove FeedbackCommandEnum = "remove"
)
