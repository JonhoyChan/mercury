package config

type Topic map[string]string

func (r *Topic) Get(name string) (string, bool) {
	if r == nil {
		return "", false
	}
	addr, ok := (*r)[name]
	return addr, ok
}

func DefaultTopic() Topic {
	return Topic{
		"push_message":      "mercury-push-message",
		"broadcast_message": "mercury-broadcast-message",
	}
}
