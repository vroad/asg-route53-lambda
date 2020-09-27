package asgroute53

import (
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
)

type mockedRoute53 struct {
	route53iface.Route53API
	listResourceRecordSetsOutput   *route53.ListResourceRecordSetsOutput
	listResourceRecordSetsError    error
	getHostedZoneOutput            *route53.GetHostedZoneOutput
	getHostedZoneError             error
	changeResourceRecordSetsOutput *route53.ChangeResourceRecordSetsOutput
	changeResourceRecordSetError   error
}

func (m *mockedRoute53) ListResourceRecordSets(input *route53.ListResourceRecordSetsInput) (*route53.ListResourceRecordSetsOutput, error) {
	if m.listResourceRecordSetsError != nil {
		return nil, m.listResourceRecordSetsError
	}

	return m.listResourceRecordSetsOutput, nil
}

func (m *mockedRoute53) GetHostedZone(input *route53.GetHostedZoneInput) (*route53.GetHostedZoneOutput, error) {
	if m.getHostedZoneError != nil {
		return nil, m.getHostedZoneError
	}

	return m.getHostedZoneOutput, nil
}

func (m *mockedRoute53) ChangeResourceRecordSets(input *route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error) {
	if m.changeResourceRecordSetError != nil {
		return nil, m.changeResourceRecordSetError
	}

	return m.changeResourceRecordSetsOutput, nil
}
