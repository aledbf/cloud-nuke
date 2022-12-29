package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/gruntwork-io/cloud-nuke/logging"
	"github.com/gruntwork-io/cloud-nuke/report"
	"github.com/gruntwork-io/go-commons/errors"
)

// Deletes all Target Groups
func nukeAllTargetGroups(session *session.Session, arns []*string) error {
	svc := elbv2.New(session)

	if len(arns) == 0 {
		logging.Logger.Debugf("No Target Groups to nuke in region %s", *session.Config.Region)
		return nil
	}

	logging.Logger.Debugf("Deleting all Target Groups in region %s", *session.Config.Region)
	var deletedArns []*string

	for _, arn := range arns {
		params := &elbv2.DeleteTargetGroupInput{
			TargetGroupArn: arn,
		}

		_, err := svc.DeleteTargetGroup(params)

		// Record status of this resource
		e := report.Entry{
			Identifier:   aws.StringValue(arn),
			ResourceType: "Target Group",
			Error:        err,
		}
		report.Record(e)

		if err != nil {
			logging.Logger.Debugf("[Failed] %s", err)
		} else {
			deletedArns = append(deletedArns, arn)
			logging.Logger.Debugf("Deleted Target Group: %s", *arn)
		}
	}

	logging.Logger.Debugf("[OK] %d V2 Elastic Load Balancer(s) deleted in %s", len(deletedArns), *session.Config.Region)
	return nil
}

func getTargetGroupArns(session *session.Session, elbv2Arns []*string) ([]*string, error) {
	svc := elbv2.New(session)

	if len(elbv2Arns) == 0 {
		logging.Logger.Debugf("No Target Groups to nuke in region %s", *session.Config.Region)
		return nil, nil
	}

	var arns []*string
	for _, elbv2Arn := range elbv2Arns {
		params := &elbv2.DescribeTargetGroupsInput{
			LoadBalancerArn: elbv2Arn,
		}

		result, err := svc.DescribeTargetGroups(params)
		if err != nil {
			return nil, errors.WithStackTrace(err)
		}

		for _, tg := range result.TargetGroups {
			arns = append(arns, tg.TargetGroupArn)
		}
	}

	return arns, nil
}
