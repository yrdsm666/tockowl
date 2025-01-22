package main

import (
	"strings"
)

func JoinPath(path ...string) string {
	// 使用斜杠拼接路径部分
	// 使用 filepath.Clean 函数处理多个连续斜杠的情况
	res := CleanPath(strings.Join(path, "/"))
	return res
}

func CleanPath(path string) string {

	// 移除路径中多余的斜杠，保留斜杠作为分隔符
	parts := strings.Split(path, "/")
	cleanedParts := make([]string, 0, len(parts))

	for _, part := range parts {
		if part != "" {
			cleanedParts = append(cleanedParts, part)
		}
	}

	cleanedPath := strings.Join(cleanedParts, "/")
	// 如果路径非空且第一个部分包含斜杠，则保留开头斜杠
	if cleanedPath != "" && strings.HasPrefix(path, "/") {
		cleanedPath = "/" + cleanedPath
	}
	return cleanedPath
}

func GetDir(path string) string {
	// 使用斜杠分隔符获取父目录
	lastIndex := strings.LastIndex(path, "/")
	if lastIndex < 0 {
		return "."
	}
	return path[:lastIndex]
}
