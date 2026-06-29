package store

// emptyMetadata 返回非 nil 的 map，避免写入 jsonb NOT NULL DEFAULT 列时产生 SQL NULL。
func emptyMetadata(m map[string]any) map[string]any {
	if m == nil {
		return map[string]any{}
	}
	return m
}

// emptyStrings 返回非 nil 的字符串切片，避免写入 jsonb 数组列时产生 SQL NULL。
func emptyStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
