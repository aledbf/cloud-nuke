package aws

import (
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	multierror "github.com/hashicorp/go-multierror"

	"github.com/gruntwork-io/cloud-nuke/config"
	"github.com/gruntwork-io/cloud-nuke/logging"
	"github.com/gruntwork-io/cloud-nuke/report"
	"github.com/gruntwork-io/go-commons/errors"
)

func getAllSSMParameters(session *session.Session, excludeAfter time.Time, configObj config.Config) ([]*string, error) {
	svc := ssm.New(session)

	allParameters := []*string{}
	results, err := svc.DescribeParameters(&ssm.DescribeParametersInput{})
	for _, param := range results.Parameters {
		if shouldIncludeSSMParameter(param, excludeAfter, configObj) {
			allParameters = append(allParameters, param.Name)
		}
	}

	return allParameters, errors.WithStackTrace(err)
}

func shouldIncludeSSMParameter(param *ssm.ParameterMetadata, excludeAfter time.Time, configObj config.Config) bool {
	if param == nil {
		return false
	}

	return config.ShouldInclude(
		aws.StringValue(param.Name),
		configObj.SSMParameter.IncludeRule.NamesRegExp,
		configObj.SSMParameter.ExcludeRule.NamesRegExp,
	)
}

func nukeAllSSMParameters(session *session.Session, identifiers []*string) error {
	region := aws.StringValue(session.Config.Region)

	svc := ssm.New(session)

	if len(identifiers) == 0 {
		logging.Logger.Debugf("No Systems Manager Parameters to nuke in region %s", region)
		return nil
	}

	// There is no bulk delete secrets API, so we delete the batch of secrets concurrently using go routines.
	logging.Logger.Debugf("Deleting Systems Manager Parameters in region %s", region)
	wg := new(sync.WaitGroup)
	wg.Add(len(identifiers))
	errChans := make([]chan error, len(identifiers))
	for i, parameterName := range identifiers {
		errChans[i] = make(chan error, 1)
		go deleteParameterAsync(wg, errChans[i], svc, parameterName)
	}
	wg.Wait()

	// Collect all the errors from the async delete calls into a single error struct.
	var allErrs *multierror.Error
	for _, errChan := range errChans {
		if err := <-errChan; err != nil {
			allErrs = multierror.Append(allErrs, err)
			logging.Logger.Errorf("[Failed] %s", err)
		}
	}
	return errors.WithStackTrace(allErrs.ErrorOrNil())
}

// deleteSecretAsync deletes the provided secrets manager secret. Intended to be run in a goroutine, using wait groups
// and a return channel for errors.
func deleteParameterAsync(wg *sync.WaitGroup, errChan chan error, svc *ssm.SSM, name *string) {
	defer wg.Done()

	input := &ssm.DeleteParameterInput{
		Name: name,
	}
	_, err := svc.DeleteParameter(input)

	// Record status of this resource
	e := report.Entry{
		Identifier:   aws.StringValue(name),
		ResourceType: "Systems Manager Parameter",
		Error:        err,
	}
	report.Record(e)

	errChan <- err
}
