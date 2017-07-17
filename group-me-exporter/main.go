package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/tabwriter"
	"time"
)

const (
	baseURI = "https://api.groupme.com/v3"
)

var (
	token = os.Getenv("GROUP_ME_TOKEN")
	gid   string
)

func init() {
	flag.StringVar(&gid, "gid", "", "specific group id to get chat history for")

	flag.Parse()

	if token == "" {
		log.Fatal("Token cannot be empty, set GROUP_ME_TOKEN env variable.")
	}
}

func main() {
	if gid == "" {
		// list groups
		groups, err := getGroups(token)
		if err != nil {
			log.Fatal(err)
		}

		// create the tabwriter
		w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)

		// print header
		fmt.Fprintln(w, "ID\tNAME\tTYPE\tDESCRIPTION")

		// print the groups
		for _, g := range groups {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", g.ID, g.Name, g.Type, g.Description)
		}

		w.Flush()
		return
	}

	// a gid was specified lets get the chat logs
	messages, err := getMessages(token, gid, "")
	if err != nil {
		log.Fatal(err)
	}

	// print the messages
	var lastDay time.Time
	for _, m := range messages {
		if m.Date.IsZero() {
			// parse the timestamp
			m.Date = time.Unix(m.CreatedAt, 0)
		}
		m.Date = m.Date.Local()
		if m.Date.Day() != lastDay.Day() ||
			m.Date.Month() != lastDay.Month() ||
			m.Date.Year() != lastDay.Year() {
			if !lastDay.IsZero() {
				os.Stdout.WriteString("\n")
			}
			os.Stdout.WriteString(fmt.Sprintf("============================= %s ============================\n", m.Date.Format("Monday, January 6 2006")))
		}
		lastDay = m.Date

		os.Stdout.WriteString(fmt.Sprintf("[%s] <%s> %s\n", m.Date.Format("15:04:05"), m.Name, m.Text))
	}
}

type chatResponse struct {
	Result messageResponse `json:"response"`
}
type messageResponse struct {
	Count    int       `json:"count"`
	Messages []message `json:"messages"`
}

// from: https://dev.groupme.com/docs/v3#messages_index
type message struct {
	ID        string    `json:"id"`
	Date      time.Time `json:"date"`
	CreatedAt int64     `json:"created_at"`
	Name      string    `json:"name"`
	UserID    string    `json:"user_id"`
	Text      string    `json:"text"`
}

func getMessages(token, gid, beforeID string) ([]message, error) {
	url := "/groups/" + gid + "/messages?limit=50"
	if beforeID != "" {
		url += "&before_id=" + beforeID
	}
	resp, err := request("GET", url, nil, token)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// no messages found will return code 304
	if resp.StatusCode == 304 {
		return nil, nil
	}

	var r chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("decoding json failed: %v", err)
	}
	result := r.Result

	if len(result.Messages) > 0 {
		// paginate results
		beforeID = result.Messages[len(result.Messages)-1].ID
		for beforeID != "" {
			messages, err := getMessages(token, gid, beforeID)
			if err != nil {
				return nil, err
			}
			return append(result.Messages, messages...), nil
		}
	}

	return result.Messages, nil
}

type groupResponse struct {
	Groups []group `json:"response"`
}

// from: https://dev.groupme.com/docs/v3#groups_index
type group struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

func getGroups(token string) ([]group, error) {
	resp, err := request("GET", "/groups?per_page=100", nil, token)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result groupResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding json failed: %v", err)
	}

	return result.Groups, nil
}

func request(method, urlStr string, body io.Reader, token string) (*http.Response, error) {
	// create the client
	client := &http.Client{}

	// create the request
	url := baseURI + urlStr + "&token=" + token
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "application/json")

	return client.Do(req)
}
