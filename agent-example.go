package main

import (
	"P4Mid/lib/Application"
	"P4Mid/lib/Manager"
	"strconv"
	"strings"
	"time"

	"github.com/hpcloud/tail"
	log "github.com/sirupsen/logrus"
)

/*
Start a controller listening for switches
*/
func main() {
	bmv2JSON := "/path/to/bmv2json"
	bmv2p4info := "/path/to/bmv2p4info"
	log.SetLevel(log.TraceLevel)
	// Create a new 'manager' instance, loaded with the appropriate P4 primitives for BMV2
	manager, err := Manager.NewLocalManager(bmv2JSON, bmv2p4info)
	if err != nil {
		log.Fatalln(err)
	}
	// Start the manager process
	defer manager.Stop()
	err = manager.Start()
	if err != nil {
		log.Fatalln(err)
	}

	// When the manager has started, add a single switch
	log.Print("Manager started")
	err = manager.AddSwitch("localhost", 50051, 1)
	if err != nil {
		log.Fatalln(err)
	}

	log.Info("Connected to switches")

	tabman := manager.TableManager()

	for { // This loop is where we can wait for other threads, or, parse responses
		time.Sleep(1000 * time.Microsecond)
	}
}

// This is a function to demonstrate splitting a space delineated string to read the 5-tuple flow fields 
func generateFlowRuleFromLogEntry(line string) (flowupdate *Application.FlowRule) {
	split := strings.Split(line, " ") // Split based on space
	priorityidx := getIndex(split, "[Priority:")
	log.Trace("Found element at ", priorityidx)
	//protoString := split[priorityidx]
	log.Println(split)
	srcString := split[priorityidx+3]
	log.Println(srcString)
	dstString := split[priorityidx+5]
	log.Println(dstString)
	srcSlice := strings.Split(srcString, ":")
	dstSlice := strings.Split(dstString, ":")
	srcPort, _ := strconv.Atoi(srcSlice[1])
	dstPort, _ := strconv.Atoi(dstSlice[1])

	// Create a new FlowRule for device 1
	flowupdate = &Application.FlowRule{DeviceId: 1}

	// Set the action name of this FlowRule, this is used to find the appropriate table, in conjunction with the flow fields
	flowupdate.ActionName = "ingress.snuffleForward"
	flowupdate.FlowFields = append(flowupdate.FlowFields, &Application.FlowField{Name: "hdr.ipv4.srcAddr", Value: srcSlice[0], Mask: []byte{255, 255, 255, 255}})
	flowupdate.FlowFields = append(flowupdate.FlowFields, &Application.FlowField{Name: "hdr.ipv4.dstAddr", Value: dstSlice[0], Mask: []byte{255, 255, 255, 255}})
	flowupdate.FlowFields = append(flowupdate.FlowFields, &Application.FlowField{Name: "meta.l4_meta.srcPort", Value: uint16(srcPort), Mask: []byte{255, 255}})
	flowupdate.FlowFields = append(flowupdate.FlowFields, &Application.FlowField{Name: "meta.l4_meta.dstPort", Value: uint16(dstPort), Mask: []byte{255, 255}})
	flowupdate.ActionParameters = append(flowupdate.ActionParameters, &Application.ActionParam{"port", uint16(64)})

	// Set the priority, this is used for conflict resolution
	flowupdate.Priority = 1
	return flowupdate
}

// Helper function to get the index of a string 
func getIndex(slice []string, searchString string) int {
	for idx, val := range slice {
		if strings.Contains(val, searchString) {
			return idx
		}
	}
	return -1
}
