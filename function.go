package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

type Notification struct {
	Incident Incident `json:"incident"`
	Version  string   `json:"version"`
}

type Incident struct {
	IncidentID    string `json:"incident_id"`
	ResourceID    string `json:"resource_id"`
	ResourceName  string `json:"resource_name"`
	State         string `json:"state"`
	StartedAt     int64  `json:"started_at"`
	EndedAt       int64  `json:"ended_at,omitempty"`
	PolicyName    string `json:"policy_name"`
	ConditionName string `json:"condition_name"`
	URL           string `json:"url"`
	Summary       string `json:"summary"`
}

type MessageCard struct {
	Type             string            `json:"@type"`
	Context          string            `json:"@context"`
	Summary          string            `json:"summary,omitempty"`
	Title            string            `json:"title,omitempty"`
	Text             string            `json:"text,omitempty"`
	ThemeColor       string            `json:"themeColor,omitempty"`
	Sections         []Section         `json:"sections,omitempty"`
	PotentialActions []PotentialAction `json:"potentialAction,omitempty"`
}

type Section struct {
	Facts []Fact `json:"facts,omitempty"`
}

type Fact struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PotentialAction struct {
	Type    string              `json:"@type"`
	Name    string              `json:"name"`
	Targets []map[string]string `json:"targets,omitempty"`
}

func toTeams(notification Notification) MessageCard {
	var startedDt time.Time
	var endedDt time.Time

	if st := notification.Incident.StartedAt; st > 0 {
		startedDt = time.Unix(st, 0)
	}

	if et := notification.Incident.EndedAt; et > 0 {
		endedDt = time.Unix(et, 0)
	}

	policyName := notification.Incident.PolicyName
	if policyName == "" {
		policyName = "-"
	}

	conditionName := notification.Incident.ConditionName
	if conditionName == "" {
		conditionName = "-"
	}

	facts := []Fact{
		{
			Name:  "Incident ID",
			Value: notification.Incident.IncidentID,
		},
		{
			Name:  "Condition",
			Value: conditionName,
		},
	}

	if !startedDt.IsZero() {
		facts = append(facts, Fact{
			Name:  "Started at",
			Value: startedDt.String(),
		})
		if !endedDt.IsZero() {
			duration := strings.TrimSpace(humanize.RelTime(startedDt, endedDt, "", ""))
			facts = append(facts, Fact{
				Name:  "Ended at",
				Value: fmt.Sprintf("%s (%s)", endedDt.String(), duration),
			})
		}
	}

	// Blue
	colour := "#1890FF"
	title := fmt.Sprintf(`Incident closed for "%s".`, policyName)
	if notification.Incident.State == "open" {
		// Red
		title = fmt.Sprintf(`Incident opened for "%s".`, policyName)
		colour = "#F5222D"
	}

	summary := "No summary available."
	if notification.Incident.Summary != "" {
		summary = notification.Incident.Summary
	}

	return MessageCard{
		Type:       "MessageCard",
		Context:    "https://schema.org/extensions",
		ThemeColor: colour,
		Title:      title,
		Text:       summary,
		Summary:    summary,
		Sections: []Section{
			{
				Facts: facts,
			},
		},
		PotentialActions: []PotentialAction{
			{
				Type: "OpenUri",
				Name: "View Incident",
				Targets: []map[string]string{
					{
						"os":  "default",
						"uri": notification.Incident.URL,
					},
				},
			},
		},
	}
}

// PubSubMessage is the payload of a Pub/Sub event. Please refer to the docs for
// additional information regarding Pub/Sub events.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

func ReportAlertToTeams(ctx context.Context, m PubSubMessage) error{

	log.Println(string(m.Data))

	teamsWebhookURL := os.Getenv("TEAMS_WEBHOOK_URL")
	if teamsWebhookURL == "" {
		log.Fatalln("`TEAMS_WEBHOOK_URL` is not set in the environments")
	}


	if _, err := url.Parse(teamsWebhookURL); err != nil {
		log.Fatalln(err)
	}

	var notification Notification
	if err := json.NewDecoder(bytes.NewReader(m.Data)).Decode(&notification); err != nil {
		log.Fatalln(err)
	}

	teamsWebhook := toTeams(notification)

	payload, err := json.Marshal(teamsWebhook)
	if err != nil {
		log.Fatalln(err)
	}

	res, err := http.Post(teamsWebhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		log.Println("payload", string(payload))
		log.Fatalln("unexpected status code", res.StatusCode)
	}
	return nil
	
}
