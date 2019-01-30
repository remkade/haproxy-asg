package main

import "testing"
import "github.com/aws/aws-sdk-go/aws"

func TestStringToAWSStringSlice(t *testing.T) {
	correct := []*string{aws.String("a"), aws.String("b")}
	s := StringToAWSStringSlice("a,b")
	if *s[0] != *correct[0] || *s[1] != *correct[1] {
		t.Errorf("StringToAWS('a,b') failed. Got [%+v, %+v]", *s[0], *s[1])
	}
}
