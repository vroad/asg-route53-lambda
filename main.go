package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/vroad/asg-route53/asgroute53"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
)

type (
	asgLifecycleEventDetail struct {
		LifecycleActionToken string
		AutoScalingGroupName string
		LifecycleHookName    string
		EC2InstanceID        string
		LifecycleTransition  string
	}
)

func completeLifecycleAction(asgClient autoscalingiface.AutoScalingAPI, event *asgLifecycleEventDetail, result string) error {
	if _, err := asgClient.CompleteLifecycleAction(&autoscaling.CompleteLifecycleActionInput{
		InstanceId:            &event.EC2InstanceID,
		LifecycleHookName:     &event.LifecycleHookName,
		LifecycleActionToken:  &event.LifecycleActionToken,
		AutoScalingGroupName:  &event.AutoScalingGroupName,
		LifecycleActionResult: aws.String(result),
	}); err != nil {
		fmt.Println("Failed completing lifecycle action: ", result)
		return err
	}

	fmt.Println("Completed lifecycle action: ", result)

	return nil
}

func appendZoneConfig(zoneConfigLoader *asgroute53.Route53ZoneConfigLoader,
	zoneConfigs []*asgroute53.Route53ZoneConfig,
	tags *[]*ec2.Tag,
	isPublic bool) ([]*asgroute53.Route53ZoneConfig, error) {
	zoneConfig, err := zoneConfigLoader.Load(tags, false)
	if err != nil {
		return nil, err
	}

	if zoneConfig == nil {
		return zoneConfigs, nil
	}

	return append(zoneConfigs, zoneConfig), nil
}

func lifecycleEventHandler(session *session.Session, event *asgLifecycleEventDetail) error {
	route53Client := route53.New(session)
	ec2Client := ec2.New(session)
	asgRoute53 := asgroute53.New(route53Client)
	zoneConfigLoader := asgroute53.NewZoneConfigLoader(route53Client)

	describeInstancesResp, err := ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(event.EC2InstanceID),
		},
	})
	if err != nil {
		return err
	}

	if len(describeInstancesResp.Reservations) < 0 || len(describeInstancesResp.Reservations[0].Instances) < 0 {
		return errors.New("failed to find an EC2 instance")
	}

	instance := describeInstancesResp.Reservations[0].Instances[0]
	zoneConfigs := []*asgroute53.Route53ZoneConfig{}

	zoneConfigs, err = appendZoneConfig(zoneConfigLoader, zoneConfigs, &instance.Tags, false)
	if err != nil {
		return err
	}
	zoneConfigs, err = appendZoneConfig(zoneConfigLoader, zoneConfigs, &instance.Tags, true)
	if err != nil {
		return err
	}
	zoneConfigsJSON, _ := json.Marshal(zoneConfigs)
	fmt.Println("zoneConfigs", string(zoneConfigsJSON))

	switch event.LifecycleTransition {
	case "autoscaling:EC2_INSTANCE_LAUNCHING":
		fmt.Println("Running upsert")
		for _, zoneConfig := range zoneConfigs {
			err := asgRoute53.UpsertRecordSets(zoneConfig, instance)
			if err != nil {
				return err
			}
		}
	case "autoscaling:EC2_INSTANCE_TERMINATING":
		fmt.Println("Running delete")
		for _, zoneConfig := range zoneConfigs {
			err := asgRoute53.DeleteRecordSets(zoneConfig, instance)
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unsupported lifecycle transition: %s", event.LifecycleTransition)
	}

	return nil
}

// Handler for Lambda
func Handler(ctx context.Context, snsEvent *events.SNSEvent) error {
	SNSMessage := snsEvent.Records[0].SNS.Message
	fmt.Println("SNS Message", SNSMessage)

	var event asgLifecycleEventDetail
	err := json.Unmarshal([]byte(SNSMessage), &event)
	if err != nil {
		return err
	}

	if event.LifecycleTransition != "autoscaling:EC2_INSTANCE_LAUNCHING" && event.LifecycleTransition != "autoscaling:EC2_INSTANCE_TERMINATING" {
		fmt.Println("The event does not contain supported LifecycleTransition, exiting.")
		return nil
	}

	session := session.Must(session.NewSession())
	asgClient := autoscaling.New(session)
	err = lifecycleEventHandler(session, &event)
	if err != nil && event.LifecycleTransition == "autoscaling:EC2_INSTANCE_LAUNCHING" {
		completeLifecycleAction(asgClient, &event, "ABANDON")
		return err
	}

	err = completeLifecycleAction(asgClient, &event, "CONTINUE")
	if err != nil {
		return err
	}

	return nil
}

func main() {
	lambda.Start(Handler)
}
