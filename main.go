package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bitly/go-nsq"
	"github.com/crosbymichael/octokat"
)

const (
	VERSION = "v0.1.0"
)

var (
	lookupd string
	topic   string
	channel string
	ghtoken string
	debug   bool
	version bool
)

var labelmap map[string]string = map[string]string{
	"#help-wanted":       "status/help-wanted",
	"#helpwanted":        "status/help-wanted",
	"#helpneeded":        "status/help-wanted",
	"#help-needed":       "status/help-wanted",
	"#needhelp":          "status/help-wanted",
	"#help":              "status/help-wanted",
	"#dibs":              "status/claimed",
	"#claimed":           "status/claimed",
	"#mine":              "status/claimed",
	"#docs-help":         "status/docs-help",
	"#docs-needed":       "status/docs-help",
	"#docshelp":          "status/docs-help",
	"#docsreview":        "status/docs-help",
	"#docsneeded":        "status/docs-help",
	"#moreinfo":          "status/more-info-needed",
	"#more-info":         "status/more-info-needed",
	"#need-info":         "status/more-info-needed",
	"#needinfo":          "status/more-info-needed",
	"+exp/novice":        "exp/novice",
	"+exp/beginner":      "exp/beginner",
	"+exp/proficient":    "exp/proficient",
	"+exp/expert":        "exp/expert",
	"+exp/master":        "exp/master",
	"+exp/mastery":       "exp/master",
	"+ exp/novice":       "exp/novice",
	"+ exp/beginner":     "exp/beginner",
	"+ exp/proficient":   "exp/proficient",
	"+ exp/expert":       "exp/expert",
	"+ exp/master":       "exp/master",
	"+ exp/mastery":      "exp/master",
	"+novice":            "exp/novice",
	"+beginner":          "exp/beginner",
	"+proficient":        "exp/proficient",
	"+expert":            "exp/expert",
	"+master":            "exp/master",
	"+mastery":           "exp/master",
	"+ novice":           "exp/novice",
	"+ beginner":         "exp/beginner",
	"+ proficient":       "exp/proficient",
	"+ expert":           "exp/expert",
	"+ master":           "exp/master",
	"+ mastery":          "exp/master",
	"+kind/proposal":     "kind/proposal",
	"+kind/enhancement":  "kind/enhancement",
	"+kind/bug":          "kind/bug",
	"+kind/cleanup":      "kind/cleanup",
	"+kind/graphics":     "kind/graphics",
	"+kind/writing":      "kind/writing",
	"+kind/docs":         "kind/writing",
	"+kind/security":     "kind/security",
	"+kind/question":     "kind/question",
	"+kind/regression":   "kind/regression",
	"+kind/feature":      "kind/feature",
	"+kind/video":        "kind/video",
	"+ kind/proposal":    "kind/proposal",
	"+ kind/enhancement": "kind/enhancement",
	"+ kind/bug":         "kind/bug",
	"+ kind/cleanup":     "kind/cleanup",
	"+ kind/graphics":    "kind/graphics",
	"+ kind/writing":     "kind/writing",
	"+ kind/docs":        "kind/writing",
	"+ kind/security":    "kind/security",
	"+ kind/question":    "kind/question",
	"+ kind/regression":  "kind/regression",
	"+ kind/feature":     "kind/feature",
	"+ kind/video":       "kind/video",
	"+proposal":          "kind/proposal",
	"+enhancement":       "kind/enhancement",
	"+bug":               "kind/bug",
	"+cleanup":           "kind/cleanup",
	"+graphics":          "kind/graphics",
	"+writing":           "kind/writing",
	"+docs":              "kind/writing",
	"+security":          "kind/security",
	"+question":          "kind/question",
	"+regression":        "kind/regression",
	"+feature":           "kind/feature",
	"+video":             "kind/video",
	"+ proposal":         "kind/proposal",
	"+ enhancement":      "kind/enhancement",
	"+ bug":              "kind/bug",
	"+ cleanup":          "kind/cleanup",
	"+ graphics":         "kind/graphics",
	"+ writing":          "kind/writing",
	"+ docs":             "kind/writing",
	"+ security":         "kind/security",
	"+ question":         "kind/question",
	"+ regression":       "kind/regression",
	"+ feature":          "kind/feature",
	"+ video":            "kind/video",
}

func init() {
	// parse flags
	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.BoolVar(&version, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "d", false, "run in debug mode")
	flag.StringVar(&lookupd, "lookupd-addr", "nsqlookupd:4161", "nsq lookupd address")
	flag.StringVar(&topic, "topic", "hooks-docker", "nsq topic")
	flag.StringVar(&channel, "channel", "patch-parser", "nsq channel")
	flag.StringVar(&ghtoken, "gh-token", "", "github access token")
	flag.Parse()
}

type Handler struct {
	GHToken string
}

func (h *Handler) HandleMessage(m *nsq.Message) error {
	prHook, err := octokat.ParsePullRequestHook(m.Body)
	if err == nil {
		return h.handlePullRequest(prHook)
	}

	// there was an error
	// so it wasn't a pull request hook
	// lets see if its an issue hook
	issueHook, err := octokat.ParseIssueHook(m.Body)
	if err == nil {
		return h.handleIssue(issueHook)
	}

	// if there was an error it means
	// it wasnt an Issue or Pull Request Hook
	// so we don't care about it
	return nil
}

func (h *Handler) handleIssue(issueHook *octokat.IssueHook) error {
	if !issueHook.IsComment() {
		// we only want comments
		return nil
	}

	for token, label := range labelmap {
		if strings.Contains(issueHook.Comment.Body, token) {
			if err := addLabel(h.getGH(), getRepo(issueHook.Repo), issueHook.Issue.Number, label); err != nil {
				return err
			}
		}
	}

	return nil
}

func getRepo(repo *octokat.Repository) octokat.Repo {
	return octokat.Repo{
		Name:     repo.Name,
		UserName: repo.Owner.Login,
	}
}

func (h *Handler) getGH() *octokat.Client {
	gh := octokat.NewClient()
	gh = gh.WithToken(h.GHToken)
	return gh
}

func (h *Handler) handlePullRequest(prHook *octokat.PullRequestHook) error {
	// we only want the prs that are opened
	if !prHook.IsOpened() {
		return nil
	}

	// get the PR
	pr := prHook.PullRequest

	// get the patch set
	patchSet, err := getPatchSet(pr.DiffURL)
	if err != nil {
		return err
	}

	// initialize github client
	gh := h.getGH()
	repo := getRepo(prHook.Repo)

	// we only want apply labels
	// to opened pull requests
	var labels []string

	// check if it's a proposal
	isProposal := strings.Contains(strings.ToLower(prHook.PullRequest.Title), "proposal")
	switch {
	case isProposal:
		labels = []string{"status/needs-design-review"}
	case isDocsOnly(patchSet):
		labels = []string{"status/needs-docs-review"}
	default:
		labels = []string{"status/needs-triage"}
	}

	// sleep before we apply the labels to try and stop waffle from removing them
	// this is gross i know
	time.Sleep(30 * time.Second)

	// add labels if there are any
	if len(labels) > 0 {
		log.Debugf("Adding labels %#v to pr %d", labels, prHook.Number)

		if err := addLabel(gh, repo, prHook.Number, labels...); err != nil {
			return err
		}

		log.Infof("Added labels %#v to pr %d", labels, prHook.Number)
	}

	// check if all the commits are signed
	if !commitsAreSigned(gh, repo, pr) {
		// add comment about having to sign commits
		comment := `Can you please sign your commits following these rules:

https://github.com/docker/docker/blob/master/CONTRIBUTING.md#sign-your-work

The easiest way to do this is to amend the last commit:

~~~console
`
		comment += fmt.Sprintf("$ git clone -b %q %s %s\n", pr.Head.Ref, pr.Head.Repo.SSHURL, "somewhere")
		comment += fmt.Sprintf("$ cd %s\n", "somewhere")
		if pr.Commits > 1 {
			comment += fmt.Sprintf("$ git rebase -i HEAD~%d\n", pr.Commits)
			comment += "editor opens\nchange each 'pick' to 'edit'\nsave the file and quit\n"
		}
		comment += "$ git commit --amend -s --no-edit\n"
		if pr.Commits > 1 {
			comment += "$ git rebase --continue # and repeat the amend for each commit\n"
		}
		comment += "$ git push -f\n"
		comment += `~~~
This will update the existing PR, so you do not need to open a new one.
`

		if err := addComment(gh, repo, strconv.Itoa(prHook.Number), comment, "sign your commits"); err != nil {
			return err
		}

		if err := addLabel(gh, repo, prHook.Number, "dco/no"); err != nil {
			return err
		}

	} else {
		if err := addLabel(gh, repo, prHook.Number, "dco/yes"); err != nil {
			return err
		}
	}

	// checkout the repository in a temp dir
	temp, err := ioutil.TempDir("", fmt.Sprintf("pr-%d", prHook.Number))
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)

	if err := checkout(temp, pr.Base.Repo.HTMLURL, prHook.Number); err != nil {
		// if it is a merge error, comment on the PR
		if err != MergeError {
			return err
		}

		comment := "Looks like we would not be able to merge this PR because of merge conflicts. Please fix them and force push to your branch."

		if err := addComment(gh, repo, strconv.Itoa(prHook.Number), comment, "conflicts"); err != nil {
			return err
		}
		return nil
	}

	// check if the files are gofmt'd
	isGoFmtd, files := checkGofmt(temp, patchSet)
	if !isGoFmtd {
		comment := fmt.Sprintf("These files are not properly gofmt'd:\n%s\n", strings.Join(files, "\n"))
		comment += "Please reformat the above files using `gofmt -s -w` and amend to the commit the result."

		if err := addComment(gh, repo, strconv.Itoa(prHook.Number), comment, "gofmt"); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	// set log level
	if debug {
		log.SetLevel(log.DebugLevel)
	}

	if version {
		fmt.Println(VERSION)
		return
	}

	bb := &Handler{GHToken: ghtoken}
	if err := ProcessQueue(bb, QueueOptsFromContext(topic, channel, lookupd)); err != nil {
		log.Fatal(err)
	}
}
