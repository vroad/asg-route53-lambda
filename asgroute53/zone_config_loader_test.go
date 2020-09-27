package asgroute53

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
)

func Test_Load(t *testing.T) {
	publicSetIdentifier := aws.String("public-set-identifier")
	privateSetIdentifier := aws.String("private-set-identifier")
	validTags := &[]*ec2.Tag{
		{
			Key:   aws.String(publicHostedZoneIDKey),
			Value: aws.String("PUBLIC-ZONE-ID"),
		},
		{
			Key:   aws.String(publicDNSRecordsKey),
			Value: aws.String("public0.example.com,public1.example.com"),
		},
		{
			Key:   aws.String(publicSetIdentifierKey),
			Value: publicSetIdentifier,
		},
		{
			Key:   aws.String(privateHostedZoneIDKey),
			Value: aws.String("PRIVATE-ZONE-ID"),
		},
		{
			Key:   aws.String(privateDNSRecordsKey),
			Value: aws.String("private0.example.com,private1.example.com"),
		},
		{
			Key:   aws.String(privateSetIdentifierKey),
			Value: privateSetIdentifier,
		},
	}
	zoneIDMissingTags := &[]*ec2.Tag{
		{
			Key:   aws.String(privateDNSRecordsKey),
			Value: aws.String("private.example.com"),
		},
		{
			Key:   aws.String(publicDNSRecordsKey),
			Value: aws.String("public.example.com"),
		},
	}
	dnsRecordMissingTags := &[]*ec2.Tag{
		{
			Key:   aws.String(publicHostedZoneIDKey),
			Value: aws.String("PUBLIC-ZONE-ID"),
		},
		{
			Key:   aws.String(privateHostedZoneIDKey),
			Value: aws.String("PRIVATE-ZONE-ID"),
		},
	}
	type args struct {
		tags     *[]*ec2.Tag
		isPublic bool
	}
	tests := []struct {
		name    string
		l       *Route53ZoneConfigLoader
		args    args
		want    *Route53ZoneConfig
		wantErr bool
	}{
		{
			name: "private",
			l: NewZoneConfigLoader(&mockedRoute53{
				getHostedZoneOutput: &route53.GetHostedZoneOutput{
					HostedZone: &route53.HostedZone{},
				},
			}),
			args: args{
				tags:     validTags,
				isPublic: false,
			},
			want: &Route53ZoneConfig{
				HostedZoneID:  "PRIVATE-ZONE-ID",
				DNSRecords:    []string{"private0.example.com", "private1.example.com"},
				SetIdentifier: privateSetIdentifier,
				IsPublic:      false,
			},
			wantErr: false,
		},
		{
			name: "public",
			l: NewZoneConfigLoader(&mockedRoute53{
				getHostedZoneOutput: &route53.GetHostedZoneOutput{
					HostedZone: &route53.HostedZone{},
				},
			}),
			args: args{
				tags:     validTags,
				isPublic: true,
			},
			want: &Route53ZoneConfig{
				HostedZoneID:  "PUBLIC-ZONE-ID",
				DNSRecords:    []string{"public0.example.com", "public1.example.com"},
				SetIdentifier: publicSetIdentifier,
				IsPublic:      true,
			},
			wantErr: false,
		},
		{
			name: "private-zone-id-missing",
			l: NewZoneConfigLoader(&mockedRoute53{
				getHostedZoneOutput: &route53.GetHostedZoneOutput{
					HostedZone: &route53.HostedZone{},
				},
			}),
			args: args{
				tags:     zoneIDMissingTags,
				isPublic: false,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "public-zone-id-missing",
			l: NewZoneConfigLoader(&mockedRoute53{
				getHostedZoneOutput: &route53.GetHostedZoneOutput{
					HostedZone: &route53.HostedZone{},
				},
			}),
			args: args{
				tags:     zoneIDMissingTags,
				isPublic: true,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "private-dns-records-missing",
			l: NewZoneConfigLoader(&mockedRoute53{
				getHostedZoneOutput: &route53.GetHostedZoneOutput{
					HostedZone: &route53.HostedZone{},
				},
			}),
			args: args{
				tags:     dnsRecordMissingTags,
				isPublic: false,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "public-dns-records-missing",
			l: NewZoneConfigLoader(&mockedRoute53{
				getHostedZoneOutput: &route53.GetHostedZoneOutput{
					HostedZone: &route53.HostedZone{},
				},
			}),
			args: args{
				tags:     dnsRecordMissingTags,
				isPublic: true,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "hosted-zone-request-error",
			l: NewZoneConfigLoader(&mockedRoute53{
				getHostedZoneError: errors.New("someError"),
			}),
			args: args{
				tags:     validTags,
				isPublic: true,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "private-empty",
			l:    NewZoneConfigLoader(&mockedRoute53{}),
			args: args{
				tags:     &[]*ec2.Tag{},
				isPublic: false,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "public-empty",
			l:    NewZoneConfigLoader(&mockedRoute53{}),
			args: args{
				tags:     &[]*ec2.Tag{},
				isPublic: false,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.l.Load(tt.args.tags, tt.args.isPublic)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Load() = %v, want %v", got, tt.want)
			}
		})
	}
}
