// Package aws provides access to Amazon Web Services (AWS).
// AWS credentials need to be passed via environment variables.
package aws

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/simpledb"
)

// AutoScalingGroup describes an AWS auto scaling group.
type AutoScalingGroup struct {
	Name               string
	InstancesInService int
	DesiredCapacity    int
	MinSize            int
	MaxSize            int
}

// AutoScalingGroups returns a list of all auto scaling groups.
func AutoScalingGroups(region string) ([]AutoScalingGroup, error) {
	var groups []AutoScalingGroup
	svc := autoscaling.New(newSession(region))
	err := svc.DescribeAutoScalingGroupsPages(nil, func(out *autoscaling.DescribeAutoScalingGroupsOutput, last bool) bool {
		for _, g := range out.AutoScalingGroups {
			inService := 0
			for _, i := range g.Instances {
				if aws.StringValue(i.LifecycleState) == autoscaling.LifecycleStateInService {
					inService++
				}
			}
			groups = append(groups, AutoScalingGroup{
				Name:               aws.StringValue(g.AutoScalingGroupName),
				InstancesInService: inService,
				DesiredCapacity:    int(aws.Int64Value(g.DesiredCapacity)),
				MinSize:            int(aws.Int64Value(g.MinSize)),
				MaxSize:            int(aws.Int64Value(g.MaxSize)),
			})
		}
		return !last
	})
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// DeleteSimpleDBDomain deletes an existing SimpleDB domain.
func DeleteSimpleDBDomain(domainName, region string) error {
	var domainExists bool
	svc := simpledb.New(newSession(region))
	err := svc.ListDomainsPages(nil, func(out *simpledb.ListDomainsOutput, last bool) bool {
		for _, n := range out.DomainNames {
			if aws.StringValue(n) == domainName {
				domainExists = true
			}
		}
		return !last
	})
	if err != nil {
		return err
	}
	if !domainExists {
		return fmt.Errorf("SimpleDB domain %q does not exist", domainName)
	}
	_, err1 := svc.DeleteDomain(&simpledb.DeleteDomainInput{
		DomainName: aws.String(domainName),
	})
	return err1
}

func newSession(region string) *session.Session {
	return session.Must(session.NewSession(&aws.Config{
		Region:     aws.String(region),
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}))
}
