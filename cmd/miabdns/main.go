package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/luv2code/gomiabdns"
	"golang.org/x/exp/slices"
)

var email string
var password string
var url string
var totp string
var command string
var recordType string
var recordName string
var recordValue string
var zone string

var commands = []string{"list", "add", "update", "delete", "zones", "zonefile"}

func init() {
	flag.StringVar(&command, "command", "list", "the command to perform: "+strings.Join(commands, ","))
	flag.StringVar(&email, "email", "", "The email address of the admin user")
	flag.StringVar(&url, "url", "", "The url of the endpoint for the admin API on your Mail-In-A-Box instance. Ex: https://box.mydomain.net/admin")
	flag.StringVar(&totp, "totp", "", "The secret key to generate a TOTP token with. Only needed when multi factor authentication is enabled")
	flag.StringVar(&password, "password", "", "The password of the admin user")
	flag.StringVar(&recordType, "rtype", "", "The record type to act on (optional) defaults to 'A' ")
	flag.StringVar(&recordName, "rname", "", "The record name to act on")
	flag.StringVar(&recordValue, "rvalue", "", "The record value to act on")
	flag.StringVar(&zone, "zone", "", "The zone for which to retrieve the zone file")
	flag.Parse()
}
func main() {
	if command == "" {
		command = "list"
	}
	if !slices.Contains(commands, command) {
		fmt.Println("The command argument must be a valid command: " + strings.Join(commands, ","))
		return
	}
	c := gomiabdns.New(url, email, password, totp)
	switch command {
	case "list":
		records, err := getRecords(c)
		if err != nil {
			panic(err)
		}
		printRecords(records)
	case "add":
		if err := addRecord(c); err != nil {
			panic(err)
		}
		fmt.Println("record added")
	case "update":
		if err := updateRecord(c); err != nil {
			panic(err)
		}
		fmt.Println("record updated")
	case "delete":
		if err := deleteRecord(c); err != nil {
			panic(err)
		}
		fmt.Println("record deleted")
	case "zones":
		zones, err := getZones(c)
		if err != nil {
			panic(err)
		}
		printZones(zones)
	case "zonefile":
		if zone == "" {
			panic(fmt.Errorf("Zone for which to retrieve the zonefile must be provided"))
		}
		zonefile, err := getZonefile(c, zone)
		if err != nil {
			panic(err)
		}
		printZonefile(zonefile, zone)
	}
}

func getRecords(c *gomiabdns.Client) ([]gomiabdns.DNSRecord, error) {
	records, err := c.GetHosts(context.TODO(), recordName, gomiabdns.RecordType(recordType))
	if err != nil {
		return nil, err
	}
	return records, nil
}

func addRecord(c *gomiabdns.Client) error {
	if recordName == "" || recordType == "" || recordValue == "" {
		return fmt.Errorf("Missing parameters to add command. all are required. rname: %s, rtype: %s, rvalue: %s ", recordName, recordType, recordValue)
	}
	if err := c.AddHost(context.TODO(), recordName, gomiabdns.RecordType(recordType), recordValue); err != nil {
		return err
	}
	return nil
}

func updateRecord(c *gomiabdns.Client) error {
	if recordName == "" || recordType == "" || recordValue == "" {
		return fmt.Errorf("Missing parameters to update command. all are required. rname: %s, rtype: %s, rvalue: %s ", recordName, recordType, recordValue)
	}
	if err := c.UpdateHost(context.TODO(), recordName, gomiabdns.RecordType(recordType), recordValue); err != nil {
		return err
	}
	return nil
}

func deleteRecord(c *gomiabdns.Client) error {
	if recordName == "" || recordType == "" {
		return fmt.Errorf("Missing parameters to delete command. rname and rtype are required.")
	}
	if err := c.DeleteHost(context.TODO(), recordName, gomiabdns.RecordType(recordType), recordValue); err != nil {
		return err
	}
	return nil
}

func printRecords(records []gomiabdns.DNSRecord) {
	writer := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', tabwriter.Debug)
	fmt.Fprintln(writer, "Name\t Type\t Value")

	for _, dr := range records {
		fmt.Fprintf(writer, "%s\t %s\t %s\n", dr.QualifiedName, dr.RecordType, dr.Value)
	}

	if err := writer.Flush(); err != nil {
		fmt.Printf("error flushing tab writer %s\n", err)
	}
}

func getZones(c *gomiabdns.Client) ([]gomiabdns.DNSZone, error) {
	records, err := c.GetZones(context.TODO())
        if err != nil {
                return nil, err
        }
        return records, nil
}

func printZones(zones []gomiabdns.DNSZone) {
	for _, dz := range zones {
		fmt.Println("Zone: " + dz)
	}
}

func getZonefile(c *gomiabdns.Client, zone string) (string, error) {
	zonefile, err := c.GetZonefile(context.TODO(), zone)
	if err != nil {
                return "", err
        }
	return zonefile, nil
}

func printZonefile(zonefile string, zone string) {
	fmt.Println("Zonefile for zone: " + zone)
	fmt.Printf(zonefile)
}
