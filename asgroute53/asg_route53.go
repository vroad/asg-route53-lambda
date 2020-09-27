package asgroute53

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
)

const ttl = 10

// ASGRoute53 handles updating and deleting DNS record for EC2 instances in an ASG
type ASGRoute53 struct {
	route53Client route53iface.Route53API
}

// New creates new instance of asgRoute53
func New(route53Client route53iface.Route53API) *ASGRoute53 {
	return &ASGRoute53{
		route53Client: route53Client,
	}
}

// DeleteRecordSets deletes record set from hosted zone
func (r *ASGRoute53) DeleteRecordSets(config *Route53ZoneConfig, ec2Instance *ec2.Instance) error {
	var changes []*route53.Change
	for _, record := range config.DNSRecords {
		recordSet, err := r.getARecordSet(config.HostedZoneID, record, config.SetIdentifier)
		if err != nil {
			return err
		}

		newChanges := r.getChanges("DELETE", config, record, ttl, *ec2Instance.InstanceId, recordSet.ResourceRecords)
		changes = append(changes, newChanges...)
	}

	_, err := r.route53Client.ChangeResourceRecordSets(&route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: changes,
		},
		HostedZoneId: &config.HostedZoneID,
	})

	return err
}

// UpsertRecordSets creates DNS record for an EC2 instance
func (r *ASGRoute53) UpsertRecordSets(config *Route53ZoneConfig, ec2Instance *ec2.Instance) error {
	ipAddress := ec2Instance.PrivateIpAddress
	if config.IsPublic {
		ipAddress = ec2Instance.PublicIpAddress
	}

	var changes []*route53.Change
	for _, record := range config.DNSRecords {
		newChanges := r.getChanges("UPSERT", config, record, ttl, *ec2Instance.InstanceId, []*route53.ResourceRecord{
			{
				Value: ipAddress,
			},
		})
		changes = append(changes, newChanges...)
	}

	_, err := r.route53Client.ChangeResourceRecordSets(&route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: changes,
		},
		HostedZoneId: aws.String(config.HostedZoneID),
	})

	return err
}

func (r *ASGRoute53) getChanges(action string,
	config *Route53ZoneConfig,
	name string,
	ttl int64,
	instanceID string,
	aResourceRecords []*route53.ResourceRecord) []*route53.Change {
	return []*route53.Change{
		{
			Action: aws.String(action),
			ResourceRecordSet: &route53.ResourceRecordSet{
				Name: aws.String(name),
				Type: aws.String("TXT"),
				ResourceRecords: []*route53.ResourceRecord{
					{
						Value: aws.String(fmt.Sprintf("\"%s\"", instanceID)),
					},
				},
				TTL:              aws.Int64(ttl),
				SetIdentifier:    config.SetIdentifier,
				MultiValueAnswer: config.MultiValueAnswer(),
			},
		},
		{
			Action: aws.String(action),
			ResourceRecordSet: &route53.ResourceRecordSet{
				Name:             aws.String(name),
				Type:             aws.String("A"),
				ResourceRecords:  aResourceRecords,
				TTL:              aws.Int64(ttl),
				SetIdentifier:    config.SetIdentifier,
				MultiValueAnswer: config.MultiValueAnswer(),
			},
		},
	}
}

func (r *ASGRoute53) getARecordSet(hostedZoneID string, name string, setIdentifier *string) (*route53.ResourceRecordSet, error) {
	recordOutput, err := r.route53Client.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{
		HostedZoneId:    aws.String(hostedZoneID),
		StartRecordType: aws.String("A"),
		StartRecordName: aws.String(name),
	})

	if err != nil {
		return nil, err
	}

	for _, recordSet := range recordOutput.ResourceRecordSets {
		if setIdentifier == nil && recordSet.SetIdentifier == nil {
			return recordSet, nil
		} else if setIdentifier != nil && recordSet.SetIdentifier != nil && *setIdentifier == *recordSet.SetIdentifier {
			return recordSet, nil
		}
	}

	return nil, fmt.Errorf("Could not find A record or SetIdentifier did not match: %s", name)
}
