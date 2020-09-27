package asgroute53

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
)

func TestASGRoute53_DeleteRecordSets(t *testing.T) {
	type args struct {
		config      *Route53ZoneConfig
		ec2Instance *ec2.Instance
	}
	tests := []struct {
		name    string
		r       *ASGRoute53
		args    args
		wantErr bool
	}{
		{
			name: "not-found",
			r: New(&mockedRoute53{
				listResourceRecordSetsOutput: &route53.ListResourceRecordSetsOutput{
					ResourceRecordSets: []*route53.ResourceRecordSet{},
				},
			}),
			args: args{
				config: &Route53ZoneConfig{
					HostedZoneID:  "ID",
					DNSRecords:    []string{"foo.example.com"},
					SetIdentifier: aws.String("identifier"),
				},
			},
			wantErr: true,
		},
		{
			name: "found",
			r: New(&mockedRoute53{
				listResourceRecordSetsOutput: &route53.ListResourceRecordSetsOutput{
					ResourceRecordSets: []*route53.ResourceRecordSet{
						{
							ResourceRecords: []*route53.ResourceRecord{
								{
									Value: aws.String("foo.example.com"),
								},
							},
							SetIdentifier: aws.String("identifier"),
						},
					},
				},
				changeResourceRecordSetsOutput: &route53.ChangeResourceRecordSetsOutput{},
			}),
			args: args{
				config: &Route53ZoneConfig{
					HostedZoneID:  "ID",
					DNSRecords:    []string{"foo.example.com"},
					SetIdentifier: aws.String("identifier"),
				},
				ec2Instance: &ec2.Instance{
					InstanceId: aws.String("i-123456789abcdef"),
				},
			},
			wantErr: false,
		},
		{
			name: "found-no-set-identifier",
			r: New(&mockedRoute53{
				listResourceRecordSetsOutput: &route53.ListResourceRecordSetsOutput{
					ResourceRecordSets: []*route53.ResourceRecordSet{
						{
							ResourceRecords: []*route53.ResourceRecord{
								{
									Value: aws.String("foo.example.com"),
								},
							},
						},
					},
				},
				changeResourceRecordSetsOutput: &route53.ChangeResourceRecordSetsOutput{},
			}),
			args: args{
				config: &Route53ZoneConfig{
					HostedZoneID: "ID",
					DNSRecords:   []string{"foo.example.com"},
				},
				ec2Instance: &ec2.Instance{
					InstanceId: aws.String("i-123456789abcdef"),
				},
			},
			wantErr: false,
		},
		{
			name: "list-error",
			r: New(&mockedRoute53{
				listResourceRecordSetsError: errors.New("listError"),
			}),
			args: args{
				config: &Route53ZoneConfig{
					HostedZoneID:  "ID",
					DNSRecords:    []string{"foo.example.com"},
					SetIdentifier: aws.String("identifier"),
				},
				ec2Instance: &ec2.Instance{
					InstanceId: aws.String("i-123456789abcdef"),
				},
			},
			wantErr: true,
		},
		{
			name: "change-error",
			r: New(&mockedRoute53{
				listResourceRecordSetsOutput: &route53.ListResourceRecordSetsOutput{
					ResourceRecordSets: []*route53.ResourceRecordSet{
						{
							ResourceRecords: []*route53.ResourceRecord{
								{
									Value: aws.String("foo.example.com"),
								},
							},
							SetIdentifier: aws.String("identifier"),
						},
					},
				},
				changeResourceRecordSetError: errors.New("changeError"),
			}),
			args: args{
				config: &Route53ZoneConfig{
					HostedZoneID:  "ID",
					DNSRecords:    []string{"foo.example.com"},
					SetIdentifier: aws.String("identifier"),
				},
				ec2Instance: &ec2.Instance{
					InstanceId: aws.String("i-123456789abcdef"),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.DeleteRecordSets(tt.args.config, tt.args.ec2Instance); (err != nil) != tt.wantErr {
				t.Errorf("ASGRoute53.DeleteRecordSets() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestASGRoute53_UpsertRecordSets(t *testing.T) {
	type args struct {
		config      *Route53ZoneConfig
		ec2Instance *ec2.Instance
	}
	tests := []struct {
		name    string
		r       *ASGRoute53
		args    args
		wantErr bool
	}{
		{
			name: "private",
			r: New(&mockedRoute53{
				changeResourceRecordSetsOutput: &route53.ChangeResourceRecordSetsOutput{},
			}),
			args: args{
				config: &Route53ZoneConfig{
					HostedZoneID:  "ID",
					DNSRecords:    []string{"foo.example.com"},
					SetIdentifier: aws.String("identifier"),
					IsPublic:      false,
				},
				ec2Instance: &ec2.Instance{
					InstanceId:       aws.String("i-123456789abcdef"),
					PrivateIpAddress: aws.String("0.0.0.0"),
				},
			},
			wantErr: false,
		},
		{
			name: "public",
			r: New(&mockedRoute53{
				changeResourceRecordSetsOutput: &route53.ChangeResourceRecordSetsOutput{},
			}),
			args: args{
				config: &Route53ZoneConfig{
					HostedZoneID:  "ID",
					DNSRecords:    []string{"foo.example.com"},
					SetIdentifier: aws.String("identifier"),
					IsPublic:      true,
				},
				ec2Instance: &ec2.Instance{
					InstanceId:      aws.String("i-123456789abcdef"),
					PublicIpAddress: aws.String("0.0.0.0"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.UpsertRecordSets(tt.args.config, tt.args.ec2Instance); (err != nil) != tt.wantErr {
				t.Errorf("ASGRoute53.UpsertRecordSets() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
