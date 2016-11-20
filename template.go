package main

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"os"
	"text/template"
)

// TemplateData is just a simple structure for accessing the data once inside the template
type TemplateData struct {
	ASGName   string
	SSLCert   string
	Instances []*ec2.Instance
}

// NewTemplateDataFromEC2Response creates a simpler way to access the ec2
// instance data inside the templates
func NewTemplateDataFromEC2Response(asgName string, sslcert string, instances *ec2.DescribeInstancesOutput) TemplateData {
	td := TemplateData{
		SSLCert:   sslcert,
		ASGName:   asgName,
		Instances: make([]*ec2.Instance, len(instances.Reservations)),
	}

	// Each reservation has one instance
	for n, r := range instances.Reservations {
		td.Instances[n] = r.Instances[0]
	}
	return td
}

// Write writes the template out to a file
func (td *TemplateData) Write(template *template.Template, destinationFile string) error {
	file, err := os.Create(destinationFile)
	if err != nil {
		return err
	}
	err = template.ExecuteTemplate(file, "haproxy", td)
	return err
}

// InstanceIDs returns only the instance IDs from the TemplateData
func (td *TemplateData) InstanceIDs() []string {
	var ids = make([]string, len(td.Instances))
	for n, instance := range td.Instances {
		ids[n] = *instance.InstanceId
	}
	return ids
}
