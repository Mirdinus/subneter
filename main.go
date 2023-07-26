package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"

	// Using pflag instead of flag to allow nices boolean flags
	"github.com/spf13/pflag"
)

func main() {
	var subnet string
	var format string
	var verbose bool

	pflag.StringVarP(&format, "format", "f", "comma", "output format (comma, json, count, range)")
	pflag.BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	pflag.Parse()

	subnet = pflag.Arg(0)

	if verbose {
		fmt.Printf("Subnet: %s Format: %s\n", subnet, format)
	}

	if format == "range" {
		start, end := getFirstAndLastIP(subnet)
		fmt.Printf("%s-%s\n", start, end)
		return
	}

	ipAddresses := getAllIPAddresses(subnet)

	switch strings.ToLower(format) {
	case "comma":
		printCommaDelimited(ipAddresses)
	case "json":
		printJSON(ipAddresses)
	case "count":
		fmt.Printf("Available IPs: %d\n", len(ipAddresses))
	default:
		fmt.Println("Invalid output format.")
	}
}

func getFirstAndLastIP(subnet string) (string, string) {
	ip, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		log.Fatalln(err)
		return "", ""
	}

	firstIP := ip.Mask(ipNet.Mask)

	lastIP := make(net.IP, len(firstIP))
	for i := range firstIP {
		lastIP[i] = firstIP[i] | ^ipNet.Mask[i]
	}

	// Add 1 to firstIP to exclude non-usable network address
	firstIP[3]++

	// Subtract 1 from lastIP to exclude non-usable broadcast address
	lastIP[3]--

	return firstIP.String(), lastIP.String()
}

func getAllIPAddresses(subnet string) []string {
	ips := make([]string, 0)

	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		log.Fatalln(err)
		return ips
	}

	ip := ipNet.IP
	for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); inc(ip) {
		if !isExcludedIP(ip) {
			ips = append(ips, ip.String())
		}
	}

	return ips
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func isExcludedIP(ip net.IP) bool {
	// Exclude IP addresses ending with 0 or 255
	return ip[len(ip)-1] == 0 || ip[len(ip)-1] == 255
}

func printCommaDelimited(ipAddresses []string) {
	fmt.Println(strings.Join(ipAddresses, ","))
}

func printJSON(ipAddresses []string) {
	addresses := make(map[string]interface{})
	addresses["ipAddresses"] = ipAddresses

	jsonData, err := json.MarshalIndent(addresses, "", "  ")
	if err != nil {
		log.Fatalln("Error marshaling JSON:", err)
		return
	}

	fmt.Println(string(jsonData))
}

func printRange(ipAddresses []string) {
	if len(ipAddresses) == 0 {
		log.Fatalln("No IP addresses to print.")
		return
	}

	startIP := ipAddresses[0]
	endIP := ipAddresses[len(ipAddresses)-1]

	start := net.ParseIP(startIP).To4()
	end := net.ParseIP(endIP).To4()

	endInt := ipToInt(end)

	fmt.Printf("%s-%s\n", start, intToIP(endInt))
}

func ipToInt(ip net.IP) int {
	b := ip.To4()
	return int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
}

func intToIP(i int) net.IP {
	return net.IPv4(byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
}
