package main

import "strings"
import "github.com/aws/aws-sdk-go/aws"

func SliceEqual(a, b []string) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func StringToAWSStringSlice(s string) []*string {
	a := strings.Split(s, ",")
	awsStrings := make([]*string, len(a), len(a))
	for i, n := range a {
		awsStrings[i] = aws.String(n)
	}
	return awsStrings
}
