package trace

// 保存链路的上下文信息「traceid，spanid，或者是其他想要传递的内容」
type spanContext struct {
	traceId string
	spanId  string
}

func (sc spanContext) TraceId() string {
	return sc.traceId
}

func (sc spanContext) SpanId() string {
	return sc.spanId
}

func (sc spanContext) Visit(fn func(key, val string) bool) {
	fn(traceIdKey, sc.traceId)
	fn(spanIdKey, sc.spanId)
}
