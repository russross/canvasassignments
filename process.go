package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	"github.com/imdario/mergo"
)

func read(filename string) []AssignmentOrGroup {
	// read the file
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", filename, err)
	}

	// parse the JSON
	var results []AssignmentOrGroup
	if err = json.Unmarshal(contents, &results); err != nil {
		log.Fatalf("Error parsing %s: %v", filename, err)
	}

	return results
}

func applyDefaults(entries []AssignmentOrGroup) (all []AssignmentOrGroup, courseID int) {
	var defaultAsst *Assignment
	var out []AssignmentOrGroup

	for _, aorg := range entries {
		if aorg.Group != nil {
			defaultAsst = nil
			out = append(out, aorg)
		} else if aorg.Assignment != nil {
			asst := aorg.Assignment

			// make sure there is a course ID
			if asst.CourseID == 0 {
				asst.CourseID = courseID
			}
			if asst.CourseID == 0 {
				log.Fatalf("Unable to determine CourseID for assignment")
			}
			if courseID == 0 {
				courseID = asst.CourseID
			}
			if courseID != asst.CourseID {
				log.Fatalf("CourseID mismatch from assignment: found %d but expected %d", asst.CourseID, courseID)
			}

			// is this a new default
			if asst.Default {
				defaultAsst = asst

				// clear a few values that don't belong in a default
				defaultAsst.Default = false
				defaultAsst.ID = 0
				defaultAsst.Name = ""
				defaultAsst.HTMLURL = ""
				defaultAsst.Position = 0

				continue
			}

			// merge with defaults?
			if defaultAsst != nil {
				// merge everything
				mergo.Merge(asst, defaultAsst)

				// apply default timestamps
				asst.DueAt = mergeDates(defaultAsst.DueAt, asst.DueAt)
				asst.LockAt = mergeDates(defaultAsst.LockAt, asst.LockAt)
				asst.UnlockAt = mergeDates(defaultAsst.UnlockAt, asst.UnlockAt)
				asst.PeerReviewsAssignAt = mergeDates(defaultAsst.PeerReviewsAssignAt, asst.PeerReviewsAssignAt)

				// apply relative timestamps
				asst.LockAt = applyAfter(asst.LockAt, asst.DueAt, asst.LockAfter)
				if asst.UnlockBefore != nil {
					asst.UnlockBefore.Duration = -asst.UnlockBefore.Duration
				}
				asst.UnlockAt = applyAfter(asst.UnlockAt, asst.DueAt, asst.UnlockBefore)
				asst.PeerReviewsAssignAt = applyAfter(asst.PeerReviewsAssignAt, asst.DueAt, asst.PeerReviewsAssignAfter)

				asst.LockAfter = nil
				asst.UnlockBefore = nil
				asst.PeerReviewsAssignAfter = nil

				/*
					// copy other defaults
					asst.Name = mergeString(defaultAsst.Name, asst.Name)
					asst.Description = mergeString(defaultAsst.Description, asst.Description)
					asst.CourseID = mergeInt(defaultAsst.CourseID, asst.CourseID)
					asst.AssignmentGroupID = mergeInt(defaultAsst.AssignmentGroupID, asst.AssignmentGroupID)
					asst.AllowedExtensions = mergeStringSlice(defaultAsst.AllowedExtensions, asst.AllowedExtensions)
					asst.TurnitinEnabled = defaultAsst.TurnitinEnabled || asst.TurnitinEnabled
					if asst.TurnitinEnabled {
						asst.TurnitinSettings = mergeTS(defaultAsst.TurnitinSettings, asst.TurnitinSettings)
					} else {
						asst.TurnitinSettings = nil
					}
					asst.GradeGroupStudentsIndividually = defaultAsst.GradeGroupStudentsIndividually || asst.GradeGroupStudentsIndividually
					asst.ExternalToolTagAttributes = mergeETTA(defaultAsst.ExternalToolTagAttributes, asst.ExternalToolTagAttributes)
				*/

				out = append(out, AssignmentOrGroup{Assignment: asst})
			} else {
				out = append(out, AssignmentOrGroup{Assignment: asst})
			}
		} else {
			log.Fatalf("AssignmentOrGroup entry that is neither assignment nor group")
		}
	}

	return out, courseID
}

func mergeDates(def, actual *jsonTime) *jsonTime {
	if def == nil && actual == nil {
		return nil
	}
	if def == nil {
		return actual
	}
	if actual == nil {
		return def
	}

	ts := def.Local()
	out := actual.Local()
	if out.Hour() == 0 && out.Minute() == 0 && out.Second() == 0 {
		// apply time (but not date) from default if time is zero for actual
		year, month, day := out.Date()
		hour, minute, second := ts.Hour(), ts.Minute(), ts.Second()
		out = time.Date(year, month, day, hour, minute, second, 0, time.Local)
	}

	return &jsonTime{out}
}

func mergeAfter(def, actual *jsonDuration) *jsonDuration {
	if actual != nil {
		return actual
	}
	return def
}

func applyAfter(actual, at *jsonTime, after *jsonDuration) *jsonTime {
	if actual != nil || after == nil || at == nil {
		return actual
	}
	return &jsonTime{at.Add(after.Duration)}
}

func mergeString(def, actual string) string {
	if actual != "" {
		return actual
	}
	return def
}

func mergeInt(def, actual int) int {
	if actual != 0 {
		return actual
	}
	return def
}

func mergeStringSlice(def, actual []string) []string {
	merged := make(map[string]bool)
	for _, elt := range def {
		merged[elt] = true
	}
	for _, elt := range actual {
		merged[elt] = true
	}
	if len(merged) == 0 {
		return nil
	}
	var lst []string
	for elt := range merged {
		lst = append(lst, elt)
	}
	return lst
}
