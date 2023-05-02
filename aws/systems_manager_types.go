package aws

import (
	awsgo "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gruntwork-io/go-commons/errors"
)

// SystemsManagerParameters - represents all AWS systems manager parameters that should be deleted.
type SystemsManagerParameters struct {
	Names []string
}

// ResourceName - the simple name of the aws resource
func (secret SystemsManagerParameters) ResourceName() string {
	return "systems-manager-parameter"
}

// ResourceIdentifiers - The instance ids of the ec2 instances
func (secret SystemsManagerParameters) ResourceIdentifiers() []string {
	return secret.Names
}

func (secret SystemsManagerParameters) MaxBatchSize() int {
	// Tentative batch size to ensure AWS doesn't throttle. Note that secrets manager does not support bulk delete, so
	// we will be deleting this many in parallel using go routines. We conservatively pick 10 here, both to limit
	// overloading the runtime and to avoid AWS throttling with many API calls.
	return 20
}

// Nuke - nuke 'em all!!!
func (secret SystemsManagerParameters) Nuke(session *session.Session, identifiers []string) error {
	if err := nukeAllSSMParameters(session, awsgo.StringSlice(identifiers)); err != nil {
		return errors.WithStackTrace(err)
	}

	return nil
}
