package logs

type LogOption func(*Log)

func WithStatus(status LogStatus) LogOption {
	return func(l *Log) {
		l.Status = status
	}
}

func WithCriticality(criticality int) LogOption {
	return func(l *Log) {
		l.Criticality = criticality
	}
}

func WithRequestID(requestID string) LogOption {
	return func(l *Log) {
		l.RequestID = &requestID
	}
}

func WithUserAgent(userAgent string) LogOption {
	return func(l *Log) {
		l.UserAgent = &userAgent
	}
}

func WithIP(ip string) LogOption {
	return func(l *Log) {
		l.IP = &ip
	}
}

func WithMetadata(key string, value interface{}) LogOption {
	return func(l *Log) {
		if l.Metadata == nil {
			l.Metadata = make(map[string]interface{})
		}
		l.Metadata[key] = value
	}
}

func WithMetadataMap(metadata map[string]interface{}) LogOption {
	return func(l *Log) {
		l.Metadata = metadata
	}
}
