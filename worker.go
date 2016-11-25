package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"text/template"
)

func Work(conf *config, as *autoscaling.AutoScaling, ec2client *ec2.EC2, tmpl *template.Template, oldInstances []string) ([]string, error) {
	instanceIDs, err := GetASGInstanceIDs(conf.ASGName, as)

	if err != nil {
		return []string{}, fmt.Errorf("Error getting AutoScalingGroup instances: '%s'", err)
	}

	if len(instanceIDs) == 0 {
		return []string{}, fmt.Errorf("Empty list of instances from ASG! Is '%s' the correct group name?", conf.ASGName)
	}

	instances, err := GetInstances(ec2client, instanceIDs)
	if err != nil {
		return []string{}, fmt.Errorf("Error getting Instances: %s", err)
	}

	// Create template data structure
	td := NewTemplateDataFromEC2Response(conf.ASGName, conf.SSLCert, instances)

	// Write out template
	if SliceEqual(td.InstanceIDs(), oldInstances) {
		log.Info("No changes found in instances, doing nothing")
	} else {
		err = td.Write(tmpl, conf.OutputFile)
		if err != nil {
			return []string{}, fmt.Errorf("Unable to write output file: %s", err)
		}

		// Restart Haproxy
		log.Info("Restarting HAProxy")
		err = RestartHAProxy(conf.Systemd)
		if err != nil {
			return []string{}, fmt.Errorf("Unable to restart haproxy: %s", err)
		}
	}
	return td.InstanceIDs(), nil
}
