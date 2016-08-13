package server

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/smtp"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/jfrazelle/hulk/api/grpc/types"
)

const (
	emailTemplate = `From: {{.From}}
To: {{.Job.EmailRecipient}}
Subject: [HULK SMASH]: {{.Job.Name}} {{.Job.Status}}

Job -->
Id: {{.Job.Id}}
Name: {{.Job.Name}}
Command: {{.Job.Args}}
Status: {{.Job.Status}}

{{if .Stderr}}
----- Stderr -----
{{.Stderr}}
{{end}}


{{if .Stdout}}
----- Stdout -----
{{.Stdout}}
{{end}}
`
)

type jobEmail struct {
	From   string
	Stdout string
	Stderr string
	Job    types.Job
}

func (s *apiServer) sendEmail(job types.Job) error {
	// Connect to the remote SMTP server.
	c, err := smtp.Dial(s.SMTPInfo.Server)
	if err != nil {
		return fmt.Errorf("Connecting to smtp server at %s failed: %v", s.SMTPInfo.Server, err)
	}

	// Set the auth.
	if err := c.Auth(s.SMTPInfo.Auth); err != nil {
		return fmt.Errorf("Setting authentication for smtp server failed: %v", err)
	}

	// Set the email sender.
	if err := c.Mail(s.SMTPInfo.Sender); err != nil {
		return fmt.Errorf("Setting email sender as %s failed: %v", s.SMTPInfo.Sender, err)
	}

	// Set the email recipient.
	if err := c.Rcpt(job.EmailRecipient); err != nil {
		return fmt.Errorf("Setting email recipient as %s failed: %v", job.EmailRecipient, err)
	}

	// Get the email writer.
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("Getting email writer failed: %v", err)
	}
	defer w.Close()

	email := jobEmail{
		From: s.SMTPInfo.Sender,
		Job:  job,
	}

	f := filepath.Join(s.StateDir, string(jobIDByte(job.Id)), "stdout")
	b, err := ioutil.ReadFile(f)
	if err != nil {
		logrus.Warnf("Could not read stdout file %s for email: %v", f, err)
	}
	email.Stdout = string(b)

	f = filepath.Join(s.StateDir, string(jobIDByte(job.Id)), "stderr")
	b, err = ioutil.ReadFile(f)
	if err != nil {
		logrus.Warnf("Could not read stderr file %s for email: %v", f, err)
	}
	email.Stderr = string(b)

	// create the template
	compiled, err := template.New("email").Parse(emailTemplate)
	if err != nil {
		return fmt.Errorf("Parsing template failed: %v", err)
	}
	if err := compiled.Execute(w, email); err != nil {
		return fmt.Errorf("Executing template failed: %v", err)
	}

	if err = w.Close(); err != nil {
		return fmt.Errorf("Closing email writer failed: %v", err)
	}

	// Send the QUIT command and close the connection.
	if err = c.Quit(); err != nil {
		return fmt.Errorf("Sending quit command to close connection failed: %v", err)
	}

	return nil
}
