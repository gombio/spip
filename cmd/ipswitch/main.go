package main

import (
	"fmt"
	"ipswitch/internal/ip"
	"log"
	"net"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

const sleepTime = (5 * 1000 * time.Millisecond)

func updateRecord(host, ip string) error {
	log.Println("Updating DNS entry in Route53: ", host, ip)

	_, err := route53Svc.ChangeResourceRecordSets(&route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(host),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(ip),
							},
						},
						TTL:  aws.Int64(120),
						Type: aws.String("A"),
					},
				},
			},
		},
		HostedZoneId: aws.String(hostedZoneID),
	})
	if err != nil {
		return err
	}

	return nil
}

var hostedZoneID string
var route53Svc *route53.Route53

func main() {
	fmt.Println("IPSwitch")

	args := os.Args[1:]
	if len(args) != 1 {
		log.Fatalln("Missing hostname argument")
	}
	host := args[0]

	log.Println("Verifying credentials")
	if os.Getenv("AWS_HOSTED_ZONE_ID") == "" ||
		os.Getenv("AWS_DEFAULT_REGION") == "" ||
		os.Getenv("AWS_ACCESS_KEY_ID") == "" ||
		os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		log.Fatalln("[Error] Please provide AWS Credentials: AWS_HOSTED_ZONE_ID, AWS_DEFAULT_REGION, AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY")
	}
	hostedZoneID = os.Getenv("AWS_HOSTED_ZONE_ID")

	log.Println("Initiating detector")
	d := ip.Detector(ip.Ipify{})

	log.Println("Setting up Route53 service")
	route53Svc = route53.New(session.New())

	log.Println("Looking up IP for Host:", host)
	addrs, err := net.LookupHost(host)
	if err != nil {
		log.Fatal("[Error]", err)
	}
	log.Println("Detected IPs:", addrs)
	if len(addrs) != 1 {
		log.Fatal("[Error] Expected single DNS entry, provided:", len(addrs))
	}

	currentIP := addrs[0]
	for {
		detectedIP, err := d.GetIP()
		if err != nil {
			log.Println("[Error]", err)
			time.Sleep(sleepTime)
			continue
		}

		if currentIP == "" {
			log.Println("Setting up initial IP:", detectedIP)
			currentIP = detectedIP
			time.Sleep(sleepTime)
			continue
		}

		if currentIP != detectedIP {
			log.Println("New IP detected:", detectedIP)
			if err := updateRecord(host, detectedIP); err != nil {
				log.Println("[Error]", err)
				time.Sleep(sleepTime)
				continue
			}

			currentIP = detectedIP
			time.Sleep(sleepTime)
			continue
		}

		log.Println("Nothing happened here")
		time.Sleep(sleepTime)
	}
}
