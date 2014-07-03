package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func upload(all []AssignmentOrGroup, courseID int, dry bool) {
	standardJSON = true

	groupID := 0
	for _, aorg := range all {
		if aorg.Group != nil {
			elt := aorg.Group
			oldID := elt.ID
			log.Printf("uploading group %d (%s)", elt.ID, elt.Name)
			groupID = uploadGroup(elt, courseID, dry)
			if oldID == 0 {
				log.Printf("new group ID %d", groupID)
			}
		} else if aorg.Assignment != nil {
			elt := aorg.Assignment
			if elt.Default {
				log.Fatalf("upload found a default assignment")
			}
			if elt.AssignmentGroupID == 0 {
				if groupID == 0 {
					log.Fatalf("unable to determine group ID for assignment")
				}
				elt.AssignmentGroupID = groupID
			} else if elt.AssignmentGroupID != groupID && groupID != 0 {
				log.Fatalf("group ID mismatch for assignment: expected %d but found %d", groupID, elt.AssignmentGroupID)
			}
			if elt.CourseID != courseID {
				log.Fatalf("course ID mismatch for assignment: expected %d but found %d", courseID, elt.CourseID)
			}
			oldID := elt.ID
			log.Printf("uploading assignment %d (%s)", elt.ID, elt.Name)
			newID := uploadAssignment(elt, courseID, dry)
			if oldID == 0 {
				log.Printf("new assignment ID %d", newID)
			}
		} else {
			log.Fatalf("upload did not find a group or an assignment")
		}
	}
}

var fakeGroupID = 1000

func uploadGroup(elt *AssignmentGroup, courseID int, dry bool) int {
	Dump(elt)
	if dry {
		if elt.ID == 0 {
			fakeGroupID++
			return fakeGroupID - 1
		}
		return elt.ID
	}

	raw, err := json.Marshal(elt)
	if err != nil {
		log.Fatalf("Error JSON encoding group: %v", err)
	}
	kind := "POST"
	targetURL := fmt.Sprintf("%s/api/v1/courses/%d/assignment_groups", apiEndpoint, courseID)
	if elt.ID != 0 {
		kind = "PUT"
		targetURL = fmt.Sprintf("%s/api/v1/courses/%d/assignment_groups/%d", apiEndpoint, courseID, elt.ID)
	}
	req, err := http.NewRequest(kind, targetURL, bytes.NewReader(raw))
	if err != nil {
		log.Fatalf("error creating http request in uploadGroup: %v", err)
	}
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("POST error in uploadGroup: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Fatalf("POST response in uploadGroup %d: %s", resp.StatusCode, resp.Status)
	}

	// decode response
	decoder := json.NewDecoder(resp.Body)
	elt = new(AssignmentGroup)
	if err = decoder.Decode(elt); err != nil {
		log.Fatalf("Error decoding object: %v", err)
	}
	return elt.ID
}

var fakeAsstID = 2000

func uploadAssignment(elt *Assignment, courseID int, dry bool) int {
	Dump(elt)
	if dry {
		if elt.ID == 0 {
			fakeAsstID++
			return fakeAsstID - 1
		}
		return elt.ID
	}

	raw, err := json.Marshal(&AssignmentOrGroup{Assignment: elt})
	if err != nil {
		log.Fatalf("Error JSON encoding assignment: %v", err)
	}
	kind := "POST"
	targetURL := fmt.Sprintf("%s/api/v1/courses/%d/assignments", apiEndpoint, courseID)
	if elt.ID != 0 {
		kind = "PUT"
		targetURL = fmt.Sprintf("%s/api/v1/courses/%d/assignments/%d", apiEndpoint, courseID, elt.ID)
	}
	req, err := http.NewRequest(kind, targetURL, bytes.NewReader(raw))
	if err != nil {
		log.Fatalf("error creating http request in uploadAssignment: %v", err)
	}
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("POST error in uploadAssignment: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Fatalf("POST response in uploadAssignment %d: %s", resp.StatusCode, resp.Status)
	}

	// decode response
	decoder := json.NewDecoder(resp.Body)
	elt = new(Assignment)
	if err = decoder.Decode(elt); err != nil {
		log.Fatalf("Error decoding object: %v", err)
	}
	return elt.ID
}
