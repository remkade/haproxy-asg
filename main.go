package main

import (
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"text/template"
	"time"
)

type config struct {
	ASGName      string
	Region       string
	LogLevel     string
	LogFormat    string
	OutputFile   string
	Template     string
	SSLCert      string
	PollInterval int
	Systemd      bool
}

var conf config

func init() {
	flag.StringVar(&conf.ASGName, "asg-name", "", "The autoscaling group's name")
	flag.StringVar(&conf.Region, "region", "us-west-2", "The region name")
	flag.StringVar(&conf.LogLevel, "log-level", "info", "Levels: panic|fatal|error|warn|info|debug")
	flag.StringVar(&conf.LogFormat, "log-format", "plain", "Supports json or plain")
	flag.StringVar(&conf.OutputFile, "output", "/etc/haproxy/haproxy.cfg", "The HAProxy config file to write")
	flag.StringVar(&conf.Template, "template", "haproxy.cfg.tmpl", "The HAProxy template file")
	flag.StringVar(&conf.SSLCert, "ssl-cert", "/etc/letsencrypt/live/example.com.crt", "The SSL Cert + Key file for haproxy")
	flag.IntVar(&conf.PollInterval, "poll-interval", 150, "Query AWS API after this many seconds")
	flag.BoolVar(&conf.Systemd, "systemd", false, "Restart using systemd")
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
	log.Info("haproxy-asg starting up")
	sess, err := session.NewSession(&aws.Config{Region: aws.String(conf.Region)})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Failed to create session")
	}

	// Create all the services
	as := autoscaling.New(sess)
	ec2client := ec2.New(sess)

	// Initialize the Template
	templateContents, err := ioutil.ReadFile(conf.Template)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Unable to read template file")
	}

	tmpl, err := template.New("haproxy").Parse(string(templateContents))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Error parsing template!")
	}

	// Create the channels
	LoopTimer := time.Tick(time.Duration(conf.PollInterval) * time.Second)
	ControlChannel := make(chan os.Signal, 2)
	signal.Notify(ControlChannel, os.Interrupt, syscall.SIGTERM)
	oldInstances := make([]string, 0)

	log.WithFields(log.Fields{
		"asg_name": conf.ASGName,
	}).Debug("Starting initial pull")

	oldInstances, err = Work(&conf, as, ec2client, tmpl, oldInstances)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Error on initial sync!")
	}

	log.WithFields(log.Fields{
		"polling_interval": conf.PollInterval,
	}).Debug("Starting main loop")
	// Start the loop
	for {
		select {
		// Exit if we receive something on teh control channel
		case <-ControlChannel:
			return
		case <-LoopTimer:
			oldInstances, err = Work(&conf, as, ec2client, tmpl, oldInstances)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("Error!")
			}
		}
	}
}

// GetASGInstanceIDs queries the AutoScalingGroup for the instances
// returning only the Instance IDs
func GetASGInstanceIDs(asgName string, as *autoscaling.AutoScaling) ([]*string, error) {
	log.WithFields(log.Fields{
		"asg-name": asgName,
	}).Debug("Getting ASG Instance IDs")

	params := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{
			aws.String(asgName),
		},
		MaxRecords: aws.Int64(100),
	}

	resp, err := as.DescribeAutoScalingGroups(params)

	if err != nil {
		return []*string{}, err
	}

	instances := resp.AutoScalingGroups[0].Instances

	log.WithFields(log.Fields{
		"asg-name":  asgName,
		"instances": fmt.Sprintf("%v", instances),
	}).Debug("Described ASG")

	ids := make([]*string, len(instances))
	for n, instance := range resp.AutoScalingGroups[0].Instances {
		log.WithFields(log.Fields{
			"asg-name":    asgName,
			"instance-id": *instance.InstanceId,
		}).Debug("Found instance id")
		ids[n] = instance.InstanceId
	}
	return ids, nil
}

// GetInstances takes the instance IDs and returns the instnace data
func GetInstances(ec2client *ec2.EC2, instanceIDs []*string) (*ec2.DescribeInstancesOutput, error) {
	params := &ec2.DescribeInstancesInput{
		InstanceIds: instanceIDs,
	}

	instances, err := ec2client.DescribeInstances(params)
	return instances, err
}

// RestartHAProxy does the HAproxy restart
func RestartHAProxy(systemd bool) (err error) {
	if systemd {
		cmd := exec.Command("systemctl", "haproxy", "reload")
		err = cmd.Run()
	} else {
		cmd := exec.Command("service", "haproxy", "reload")
		err = cmd.Run()
	}
	return
}
