package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

const apiEndpoint = "https://dixie.instructure.com"

var authHeader string

func main() {
	token := os.Getenv("CANVAS_TOKEN")
	if token == "" {
		log.Fatalf("Must set CANVAS_TOKEN environment variable")
	}
	authHeader = fmt.Sprintf("Bearer %s", token)

	var (
		courseID           int
		assignmentID       int
		assignmentGroupID  int
		includeAssignments bool
		file               string
		dry                bool
	)
	flag.IntVar(&courseID, "course", 0, "Course ID")
	flag.IntVar(&assignmentID, "assignment", 0, "Assignment ID")
	flag.IntVar(&assignmentGroupID, "assignment_group", 0, "Assignment Group ID")
	flag.BoolVar(&includeAssignments, "include_assignments", false, "Fetch assignments in group")
	flag.StringVar(&file, "file", "", "Upload courses and groups from this file")
	flag.BoolVar(&dry, "dry", false, "Dry run")
	flag.Parse()

	switch {
	case courseID > 0 && assignmentID > 0 && file == "":
		reportAssignment(courseID, assignmentID)

	case courseID > 0 && assignmentGroupID > 0 && file == "":
		reportAssignmentGroup(courseID, assignmentGroupID, includeAssignments)

	case courseID > 0 && file == "":
		reportAllAssignmentGroups(courseID, includeAssignments)

	case file != "":
		templates := read(file)
		entries, courseID := applyDefaults(templates, courseID)
		upload(entries, courseID, dry)

	default:
		flag.Usage()
	}
}

func reportAssignment(courseID, assignmentID int) {
	// fetch the assignment
	targetURL := fmt.Sprintf("%s/api/v1/courses/%d/assignments/%d", apiEndpoint, courseID, assignmentID)
	asst := new(Assignment)
	mustFetch(targetURL, asst)

	// output it as JSON
	asst.Cleanup()
	asst.Dump()
}

func reportAssignmentGroup(courseID, assignmentGroupID int, includeAssignments bool) {
	// fetch the assignment group
	include := ""
	if includeAssignments {
		include = "?include=assignments"
	}
	targetURL := fmt.Sprintf("%s/api/v1/courses/%d/assignment_groups/%d%s", apiEndpoint, courseID, assignmentGroupID, include)
	group := new(AssignmentGroup)
	mustFetch(targetURL, group)

	dumpGroups([]*AssignmentGroup{group}, includeAssignments)
}

func reportAllAssignmentGroups(courseID int, includeAssignments bool) {
	// fetch the assignment group
	include := ""
	if includeAssignments {
		include = "?include=assignments"
	}
	targetURL := fmt.Sprintf("%s/api/v1/courses/%d/assignment_groups%s", apiEndpoint, courseID, include)
	var groups []*AssignmentGroup
	mustFetch(targetURL, &groups)

	dumpGroups(groups, includeAssignments)
}

func dumpGroups(groups []*AssignmentGroup, includeAssignments bool) {
	// create a single list
	var lst []AssignmentOrGroup
	for _, group := range groups {
		group.Cleanup()
		assts := group.Assignments
		group.Assignments = nil
		lst = append(lst, AssignmentOrGroup{Group: group})
		for _, elt := range assts {
			lst = append(lst, AssignmentOrGroup{Assignment: elt})
		}
	}
	Dump(lst)
}

func mustFetch(targetURL string, elt interface{}) {
	token := os.Getenv("CANVAS_TOKEN")
	if token == "" {
		log.Fatalf("Must set CANVAS_TOKEN environment variable")
	}

	// fetch the object
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		log.Fatalf("Error creating HTTP request: %v", err)
	}
	req.Header.Add("Authorization", authHeader)

	// report the equivalent curl command
	//log.Printf(`curl -H "Authorization: Bearer $CANVAS_TOKEN" '%s'`, targetURL)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("GET error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Fatalf("GET response %d: %s", resp.StatusCode, resp.Status)
	}

	// decode it
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(elt); err != nil {
		log.Fatalf("Error decoding object: %v", err)
	}
}
