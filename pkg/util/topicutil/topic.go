package topicutil

import "fmt"

const (
	TopicPrefix           = "evhub_"
	DelayTopicPrefix      = "eh_delay_"
	DelaySubNameKeyPrefix = "delay_%v_%v"

	TopicKeyPrefix   = "evhub_%v_%v"
	SubNameKeyPrefix = "group_%v"
)

// Topic returns the topic key associated with the given appID, topicID
func Topic(appID, topicID string) string {
	return fmt.Sprintf(TopicKeyPrefix, appID, topicID)
}

// SubName returns the sub topic key associated with the given dispatcherID
func SubName(dispatcherID string) string {
	return fmt.Sprintf(SubNameKeyPrefix, dispatcherID)
}

// DelayTopic returns the delay topic key associated with the given appID, topicID
func DelayTopic(appID, topicID string) string {
	return fmt.Sprintf(DelayTopicPrefix+"%v_%v", appID, topicID)
}

// DelaySubName returns the delay sub topic key associated with the given appID, topicID
func DelaySubName(appID, topicID string) string {
	return fmt.Sprintf(DelaySubNameKeyPrefix, appID, topicID)
}
