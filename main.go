package main

import (
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	//"github.com/aws/aws-sdk-go/service/sns"
)

type config struct {
	ASGName   string
	Region    string
	LogLevel  string
	LogFormat string
}

var conf config

func init() {
	flag.StringVar(&conf.ASGName, "asg-name", "", "The autoscaling group's name")
	flag.StringVar(&conf.Region, "region", "us-west-2", "The region name")
	flag.StringVar(&conf.LogLevel, "log-level", "info", "Levels: panic|fatal|error|warn|info|debug")
	flag.StringVar(&conf.LogFormat, "log-format", "plain", "Supports json or plain")
	flag.Parse()

	switch conf.LogLevel {
	case "panic":
		log.SetLevel(log.PanicLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	}

	if conf.LogFormat == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	}
}

func main() {

	sess, err := session.NewSession()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Failed to create session")
	}

	// Create all the services
	as := autoscaling.New(sess)
	ec2client := ec2.New(sess)

	instanceIDs := GetASGInstanceIDs(conf.ASGName, as)
	params := &ec2.DescribeInstancesInput{
		InstanceIds: instanceIDs,
	}

	instances, err := ec2client.DescribeInstances(params)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Failed to describe instances")
	}
	fmt.Println(instances)
}

// GetASGInstanceIDs queries the AutoScalingGroup for the instances
// returning only the Instance IDs
func GetASGInstanceIDs(asgName string, as *autoscaling.AutoScaling) (ids []*string) {
	params := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{
			aws.String(conf.ASGName),
		},
		MaxRecords: aws.Int64(100),
	}
	resp, err := as.DescribeAutoScalingGroups(params)

	if err != nil {
		fmt.Println(err.Error())
		log.WithFields(log.Fields{
			"asg-name": conf.ASGName,
			"error":    err,
		}).Fatal("Failed to describe auto scaling group")
	}

	log.WithFields(log.Fields{
		"asg-name": conf.ASGName,
		"value":    fmt.Sprintf("%+v", resp),
	}).Debug("Described ASG")

	if len(resp.AutoScalingGroups) == 0 {
		log.WithFields(log.Fields{
			"asg-name": conf.ASGName,
			"error":    "Unknown autoscaling group",
		}).Fatal("Failed to describe auto scaling group")
	}

	instances := resp.AutoScalingGroups[0].Instances

	ids = make([]*string, len(instances))
	for n, instance := range resp.AutoScalingGroups[0].Instances {
		log.WithFields(log.Fields{
			"asg-name":    conf.ASGName,
			"instance-id": *instance.InstanceId,
		}).Debug("Found instance id")
		ids[n] = instance.InstanceId
	}

	return
}
