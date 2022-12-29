package aws

import (
	awsgo "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gruntwork-io/go-commons/errors"
)

// TargetGroup - represents all target groups
type TargetGroup struct {
	Arns []string
}

// ResourceName - the simple name of the aws resource
func (targetGroup TargetGroup) ResourceName() string {
	return "target-group"
}

func (targetGroup TargetGroup) MaxBatchSize() int {
	// Tentative batch size to ensure AWS doesn't throttle
	return 49
}

// ResourceIdentifiers - The arns of the load balancers
func (targetGroup TargetGroup) ResourceIdentifiers() []string {
	return targetGroup.Arns
}

// Nuke - nuke 'em all!!!
func (targetGroup TargetGroup) Nuke(session *session.Session, identifiers []string) error {
	if err := nukeAllTargetGroups(session, awsgo.StringSlice(identifiers)); err != nil {
		return errors.WithStackTrace(err)
	}

	return nil
}
