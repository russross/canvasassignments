package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var standardJSON = false

type Assignment struct {
	Default                        bool                       `json:"default,omitempty" yaml:"default,omitempty"`
	ID                             int                        `json:"id,omitempty" yaml:"id,omitempty"`
	Name                           string                     `json:"name,omitempty" yaml:"name,omitempty"`
	Description                    string                     `json:"description,omitempty" yaml:"description,omitempty"`
	DueAt                          *jsonTime                  `json:"due_at,omitempty" yaml:"due_at,omitempty"`
	LockAt                         *jsonTime                  `json:"lock_at,omitempty" yaml:"lock_at,omitempty"`
	LockAfter                      *jsonDuration              `json:"lock_after,omitempty" yaml:"lock_after,omitempty"`
	UnlockAt                       *jsonTime                  `json:"unlock_at,omitempty" yaml:"unlock_at,omitempty"`
	UnlockBefore                   *jsonDuration              `json:"unlock_before,omitempty" yaml:"unlock_before,omitempty"`
	CourseID                       int                        `json:"course_id,omitempty" yaml:"course_id,omitempty"`
	HTMLURL                        string                     `json:"html_url,omitempty" yaml:"html_url,omitempty"`
	AssignmentGroupID              int                        `json:"assignment_group_id,omitempty" yaml:"assignment_group_id,omitempty"`
	AllowedExtensions              []string                   `json:"allowed_extensions,omitempty" yaml:"allowed_extensions,omitempty,flow"`
	TurnitinEnabled                bool                       `json:"turnitin_enabled,omitempty" yaml:"turnitin_enabled,omitempty"`
	TurnitinSettings               *TurnitinSettings          `json:"turnitin_settings,omitempty" yaml:"turnitin_settings,omitempty"`
	GradeGroupStudentsIndividually bool                       `json:"grade_group_students_individually,omitempty" yaml:"grade_group_students_individually,omitempty"`
	ExternalToolTagAttributes      *ExternalToolTagAttributes `json:"external_tool_tag_attributes,omitempty" yaml:"external_tool_tag_attributes,omitempty"`
	PeerReviews                    bool                       `json:"peer_reviews,omitempty" yaml:"peer_reviews,omitempty"`
	AutomaticPeerReviews           bool                       `json:"automatic_peer_reviews,omitempty" yaml:"automatic_peer_reviews,omitempty"`
	PeerReviewCount                int                        `json:"peer_review_count,omitempty" yaml:"peer_review_count,omitempty"`
	PeerReviewsAssignAt            *jsonTime                  `json:"peer_reviews_assign_at,omitempty" yaml:"peer_reviews_assign_at,omitempty"`
	PeerReviewsAssignAfter         *jsonDuration              `json:"peer_reviews_assign_after,omitempty" yaml:"peer_reviews_assign_after,omitempty"`
	GroupCategoryID                int                        `json:"group_category_id,omitempty" yaml:"group_category_id,omitempty"`
	NeedsGradingCount              int                        `json:"needs_grading_count,omitempty" yaml:"needs_grading_count,omitempty"`
	Position                       int                        `json:"position,omitempty" yaml:"position,omitempty"`
	PostToSIS                      bool                       `json:"post_to_sis,omitempty" yaml:"post_to_sis,omitempty"`
	Muted                          bool                       `json:"muted,omitempty" yaml:"muted,omitempty"`
	PointsPossible                 float64                    `json:"points_possible,omitempty" yaml:"points_possible,omitempty"`
	SubmissionTypes                []string                   `json:"submission_types,omitempty" yaml:"submission_types,omitempty,flow"`
	GradingType                    string                     `json:"grading_type,omitempty" yaml:"grading_type,omitempty"`
	GradingStandardID              int                        `json:"grading_standard_id,omitempty" yaml:"grading_standard_id,omitempty"`
	Published                      bool                       `json:"published,omitempty" yaml:"published,omitempty"`
	Unpublishable                  bool                       `json:"unpublishable,omitempty" yaml:"unpublishable,omitempty"`
	OnlyVisibleToOverrides         bool                       `json:"only_visible_to_overrides,omitempty" yaml:"only_visible_to_overrides,omitempty"`
	LockedForUser                  bool                       `json:"locked_for_user,omitempty" yaml:"locked_for_user,omitempty"`
	LockInfo                       string                     `json:"lock_info,omitempty" yaml:"lock_info,omitempty"`
	LockExplanation                string                     `json:"lock_explanation,omitempty" yaml:"lock_explanation,omitempty"`
	QuizID                         int                        `json:"quiz_id,omitempty" yaml:"quiz_id,omitempty"`
	AnonymousSubmissions           bool                       `json:"anonymous_submissions,omitempty" yaml:"anonymous_submissions,omitempty"`
	DiscussionTopic                string                     `json:"discussion_topic,omitempty" yaml:"discussion_topic,omitempty"`
	FreezeOnCopy                   bool                       `json:"freeze_on_copy,omitempty" yaml:"freeze_on_copy,omitempty"`
	Frozen                         bool                       `json:"frozen,omitempty" yaml:"frozen,omitempty"`
	FrozenAttributes               []string                   `json:"frozen_attributes,omitempty" yaml:"frozen_attributes,omitempty"`
	Submission                     *Submission                `json:"submission,omitempty" yaml:"submission,omitempty"`
	UseRubricForGrading            bool                       `json:"use_rubric_for_grading,omitempty" yaml:"use_rubric_for_grading,omitempty"`
	RubricSettings                 *RubricSettings            `json:"rubricsettings,omitempty" yaml:"rubricsettings,omitempty"`
	Rubric                         []*RubricCriteria          `json:"rubric,omitempty" yaml:"rubric,omitempty"`
}

func (elt *Assignment) Cleanup() {
	if !elt.TurnitinEnabled {
		elt.TurnitinSettings = nil
	}
	if !elt.PeerReviews {
		elt.AutomaticPeerReviews = false
		elt.PeerReviewCount = 0
		elt.PeerReviewsAssignAt = nil
	}
	elt.HTMLURL = ""
	elt.Unpublishable = false
	/*
		if elt.DueAt != nil && elt.LockAt != nil {
			gap := jsonDuration{elt.LockAt.Sub(*elt.DueAt)}
			elt.LockAfter = &gap
		}
		if elt.DueAt != nil && elt.UnlockAt != nil {
			gap := jsonDuration{elt.DueAt.Sub(*elt.UnlockAt)}
			elt.UnlockBefore = &gap
		}
		if elt.DueAt != nil && elt.PeerReviewsAssignAt != nil {
			gap := jsonDuration{elt.PeerReviewsAssignAt.Sub(*elt.DueAt)}
			elt.PeerReviewsAssignAfter = &gap
		}
	*/
}

func (elt *Assignment) Clone() *Assignment {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(elt); err != nil {
		log.Fatalf("Error gob encoding assignment: %v", err)
	}
	decoder := gob.NewDecoder(&buf)
	clone := new(Assignment)
	if err := decoder.Decode(clone); err != nil {
		log.Fatalf("Error gob decoding assignment: %v", err)
	}
	return clone
}

func (elt *Assignment) Dump() {
	raw, err := json.MarshalIndent([]AssignmentOrGroup{AssignmentOrGroup{Assignment: elt}}, "", "    ")
	if err != nil {
		log.Fatalf("JSON error encoding assignment: %v", err)
	}
	os.Stdout.Write(raw)
	fmt.Println()
}

type ExternalToolTagAttributes struct {
	URL            string `json:"url,omitempty" yaml:"url,omitempty"`
	NewTab         bool   `json:"new_tab" yaml:"new_tab"`
	ResourceLinkID string `json:"resource_link_id,omitempty" yaml:"resource_link_id,omitempty"`
}

type RubricSettings struct {
	ID                        int     `json:"id,omitempty" yaml:"id,omitempty"`
	Title                     string  `json:"title,omitempty" yaml:"title,omitempty"`
	PointsPossible            float64 `json:"points_possible,omitempty" yaml:"points_possible,omitempty"`
	FreeFormCriterionComments bool    `json:"free_form_criterion_comments,omitempty" yaml:"free_form_criterion_comments,omitempty"`
}

type RubricRating struct {
	Points      float64 `json:"points,omitempty" yaml:"points,omitempty"`
	ID          string  `json:"id,omitempty" yaml:"id,omitempty"`
	Description string  `json:"description,omitempty" yaml:"description,omitempty"`
}

type RubricCriteria struct {
	Points      float64         `json:"points,omitempty" yaml:"points,omitempty"`
	ID          string          `json:"id,omitempty" yaml:"id,omitempty"`
	Description string          `json:"description,omitempty" yaml:"description,omitempty"`
	Ratings     []*RubricRating `json:"ratings,omitempty" yaml:"ratings,omitempty"`
}

type TurnitinSettings struct {
	OriginalityReportVisibility string  `json:"originality_report_visibility,omitempty" yaml:"originality_report_visibility,omitempty"`
	SPaperCheck                 bool    `json:"s_paper_check,omitempty" yaml:"s_paper_check,omitempty"`
	InternetCheck               bool    `json:"internet_check,omitempty" yaml:"internet_check,omitempty"`
	JournalCheck                bool    `json:"journal_check,omitempty" yaml:"journal_check,omitempty"`
	ExcludeBiblio               bool    `json:"exclude_biblio,omitempty" yaml:"exclude_biblio,omitempty"`
	ExcludeQuoted               bool    `json:"exclude_quoted,omitempty" yaml:"exclude_quoted,omitempty"`
	ExcludeSmallMatchesType     string  `json:"exclude_small_matches_type,omitempty" yaml:"exclude_small_matches_type,omitempty"`
	ExcludeSmallMatchesValue    float64 `json:"exclude_small_matches_value,omitempty" yaml:"exclude_small_matches_value,omitempty"`
}

type Submission struct {
}

type AssignmentGroup struct {
	Default     bool          `json:"-" yaml:"default,omitempty"`
	ID          int           `json:"id,omitempty" yaml:"id,omitempty"`
	Name        string        `json:"name,omitempty" yaml:"name,omitempty"`
	Position    int           `json:"position,omitempty" yaml:"position,omitempty"`
	GroupWeight float64       `json:"group_weight,omitempty" yaml:"group_weight,omitempty"`
	Assignments []*Assignment `json:"assignments,omitempty" yaml:"assignments,omitempty"`
	Rules       *GradingRules `json:"rules,omitempty" yaml:"rules,omitempty"`
}

func (elt *AssignmentGroup) Cleanup() {
	for _, asst := range elt.Assignments {
		asst.Cleanup()
	}
	if elt.Rules != nil && elt.Rules.DropLowest == 0 && elt.Rules.DropHighest == 0 && len(elt.Rules.NeverDrop) == 0 {
		elt.Rules = nil
	}
}

func (elt *AssignmentGroup) Dump() {
	raw, err := json.MarshalIndent([]AssignmentOrGroup{AssignmentOrGroup{Group: elt}}, "", "    ")
	if err != nil {
		log.Fatalf("JSON error encoding group: %v", err)
	}
	os.Stdout.Write(raw)
	fmt.Println()
}

type GradingRules struct {
	DropLowest  int   `json:"drop_lowest,omitempty" yaml:"drop_lowest,omitempty"`
	DropHighest int   `json:"drop_highest,omitempty" yaml:"drop_highest,omitempty"`
	NeverDrop   []int `json:"never_drop,omitempty" yaml:"never_drop,omitempty"`
}

type AssignmentOrGroup struct {
	Assignment *Assignment      `json:"assignment,omitempty" yaml:"assignment,omitempty"`
	Group      *AssignmentGroup `json:"assignment_group,omitempty" yaml:"assignment_group,omitempty"`
}

func (elt *AssignmentOrGroup) Dump() {
	if elt.Group != nil {
		elt.Group.Dump()
	} else if elt.Assignment != nil {
		elt.Assignment.Dump()
	} else {
		log.Fatalf("AssignmentOrGroup with no assignment or group")
	}
}

func Dump(elt interface{}) {
	raw, err := json.MarshalIndent(elt, "", "    ")
	if err != nil {
		log.Fatalf("JSON error encoding element: %v", err)
	}
	os.Stdout.Write(raw)
	fmt.Println()
}

type jsonTime struct {
	time.Time
}

func (elt jsonTime) MarshalJSON() ([]byte, error) {
	if standardJSON {
		return []byte(elt.UTC().Format(`"` + time.RFC3339Nano + `"`)), nil
	}

	t := elt.Local()
	year, month, day := t.Date()
	if year == 0 && month == 0 && day == 0 {
		return []byte(t.Format(`"15:04:05"`)), nil
	}
	hour, minute, second, ns := t.Hour(), t.Minute(), t.Second(), t.Nanosecond()
	if hour == 0 && minute == 0 && second == 0 && ns == 0 {
		return []byte(t.Format("2006-01-02")), nil
	}
	return []byte(t.Format(`"2006-01-02 15:04:05"`)), nil
}

func (elt *jsonTime) UnmarshalJSON(b []byte) error {
	s := string(b)
	t, err := time.ParseInLocation(`"2006-01-02 15:04:05"`, s, time.Local)
	if err == nil {
		*elt = jsonTime{t}
		return nil
	}
	t, err = time.ParseInLocation(`"2006-01-02"`, s, time.Local)
	if err == nil {
		*elt = jsonTime{t}
		return nil
	}
	t, err = time.ParseInLocation(`"15:04:05"`, s, time.Local)
	if err == nil {
		*elt = jsonTime{t}
		return nil
	}
	t, err = time.Parse(`"`+time.RFC3339+`"`, s)
	*elt = jsonTime{t}
	return err
}

type jsonDuration struct {
	time.Duration
}

func (d jsonDuration) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

func (d *jsonDuration) UnmarshalJSON(b []byte) error {
	s := string(b)
	if strings.HasPrefix(s, `"`) {
		s = s[1:]
	}
	if strings.HasSuffix(s, `"`) {
		s = s[:len(s)-1]
	}
	if elt, err := time.ParseDuration(s); err != nil {
		return err
	} else {
		d.Duration = elt
	}
	return nil
}
