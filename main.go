package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/services/ec2"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
}

func main() {
	sess, err := session.NewSession()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatalf("Failed to create session")
	}

	as := autoscaling.New(sess)
	params := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{
			aws.String(config.ASGName),
		},
		MaxRecords: aws.Int64(1),
		NextToken:  aws.String("XmlString"),
	}
}
