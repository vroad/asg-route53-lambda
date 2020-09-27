package asgroute53

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
)

type (
	// Route53ZoneConfigLoader loads record set configurations from instance tags
	Route53ZoneConfigLoader struct {
		route53Client route53iface.Route53API
	}
	// Route53ZoneConfig holds record set configuration
	Route53ZoneConfig struct {
		HostedZoneID  string
		DNSRecords    []string
		SetIdentifier *string
		IsPublic      bool
	}
)

const privateHostedZoneIDKey = "asg-route53-lambda:private-hosted-zone-id"
const privateDNSRecordsKey = "asg-route53-lambda:private-dns-records"
const privateSetIdentifierKey = "asg-route53-lambda:private-set-identifier"
const publicHostedZoneIDKey = "asg-route53-lambda:public-hosted-zone-id"
const publicDNSRecordsKey = "asg-route53-lambda:public-dns-records"
const publicSetIdentifierKey = "asg-route53-lambda:public-set-identifier"

// NewZoneConfigLoader creates new instance of Route53ZoneConfigLoader
func NewZoneConfigLoader(route53Client route53iface.Route53API) *Route53ZoneConfigLoader {
	return &Route53ZoneConfigLoader{
		route53Client: route53Client,
	}
}

func (l Route53ZoneConfigLoader) findValueFromEC2Tags(tags *[]*ec2.Tag, key string) *string {
	for _, tag := range *tags {
		if *tag.Key == key {
			return tag.Value
		}
	}

	return nil
}

// Load loads record set config from EC2 tags
func (l Route53ZoneConfigLoader) Load(tags *[]*ec2.Tag, isPublic bool) (*Route53ZoneConfig, error) {
	zoneIDKey := privateHostedZoneIDKey
	recordsKey := privateDNSRecordsKey
	setIdentifierKey := privateSetIdentifierKey

	if isPublic {
		zoneIDKey = publicHostedZoneIDKey
		recordsKey = publicDNSRecordsKey
		setIdentifierKey = publicSetIdentifierKey
	}

	zoneID := l.findValueFromEC2Tags(tags, zoneIDKey)
	inDNSRecords := l.findValueFromEC2Tags(tags, recordsKey)
	setIdentifier := l.findValueFromEC2Tags(tags, setIdentifierKey)

	if (zoneID != nil && inDNSRecords == nil) ||
		(zoneID == nil && inDNSRecords != nil) {
		return nil, fmt.Errorf("both %s and %s should be specified", zoneIDKey, recordsKey)
	}

	if zoneID != nil && inDNSRecords != nil {
		_, err := l.route53Client.GetHostedZone(&route53.GetHostedZoneInput{
			Id: zoneID,
		})

		if err != nil {
			return nil, err
		}

		return &Route53ZoneConfig{
			HostedZoneID:  *zoneID,
			DNSRecords:    strings.Split(*inDNSRecords, ","),
			SetIdentifier: setIdentifier,
			IsPublic:      isPublic,
		}, nil
	}

	return nil, nil
}

// MultiValueAnswer returns true if the record needs to be inserted with multi value answer option
func (c *Route53ZoneConfig) MultiValueAnswer() *bool {
	if c.SetIdentifier != nil {
		return aws.Bool(true)
	}

	return nil
}
