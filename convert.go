package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/russross/blackfriday"
)

const (
	ital = 1 << iota
	bold
	code
)

var (
	fontName = []string{
		"R",
		"I",
		"B",
		"[BI]",
		"C",
		"[CI]",
		"[CB]",
		"[CBI]",
	}
)

type md2PDF struct {
	doc     *docData
	data    []byte
	buf     *bytes.Buffer
	listNum int
	raw     bool
	font    int
}

type docData struct {
	Title     string
	Subtitle  string
	Name      string
	URI       string
	Draft     bool
	Date      time.Time
	PtPerInch float64
}

func (m *md2PDF) convert(x *blackfriday.Node) error {
	// Before
	switch x.Type {
	default:
		log.Fatalf("%s: node %s not handled", m.doc.Name, x.Type)
	case blackfriday.Document, blackfriday.Hardbreak:
		// nothing
	case blackfriday.Link:
		fmt.Fprintf(m.buf, "\\W'%s'", x.Destination)
		defer fmt.Fprintf(m.buf, "\\W")
	case blackfriday.Paragraph:
		nl(m.buf)
		if x.FirstChild != nil && x.FirstChild.Type == blackfriday.Text && bytes.HasPrefix(x.FirstChild.Literal, []byte(".")) {
			m.raw = true
			defer func() {
				m.raw = false
			}()
			break
		}
		fmt.Fprintf(m.buf, ".PP\n")
	case blackfriday.Text:
		if m.raw {
			m.buf.Write(x.Literal)
			break
		}
		literal(m.buf, x.Literal)
	case blackfriday.BlockQuote:
		nl(m.buf)
		fmt.Fprintf(m.buf, ".QS\n")
		defer func() {
			nl(m.buf)
			fmt.Fprintf(m.buf, ".QE\n")
		}()
	case blackfriday.CodeBlock:
		nl(m.buf)
		lit, wid := maxLineWidth(x.Literal)
		indent := ""
		if wid > 70 {
			indent = fmt.Sprintf(" -%.2fi", (float64(wid)-57)/2*(4.5/65))
		}
		fmt.Fprintf(m.buf, ".P1%s\n", indent)
		literal(m.buf, lit)
		nl(m.buf)
		fmt.Fprintf(m.buf, ".P2\n")

		return nil
	case blackfriday.Emph:
		m.addFont(ital)
		defer m.subFont(ital)
	case blackfriday.Code:
		if m.font&bold != 0 {
			fmt.Fprintf(m.buf, "\\f(CB")
		} else {
			fmt.Fprintf(m.buf, "\\fC")
		}
		literal(m.buf, x.Literal)
		fmt.Fprintf(m.buf, "\\fP")

		return nil
	case blackfriday.Strong:
		m.addFont(bold)
		defer m.subFont(bold)
	case blackfriday.Heading:
		nl(m.buf)
		f := m.font
		m.font |= bold
		defer func() {
			m.font = f
		}()
		fmt.Fprintf(m.buf, ".H%d\n", x.Level)
	case blackfriday.List:
		nl(m.buf)
		oldListNum := m.listNum
		if x.ListFlags&blackfriday.ListTypeOrdered != 0 {
			fmt.Fprintf(m.buf, ".OL\n")
			m.listNum = 1
		} else {
			fmt.Fprintf(m.buf, ".UL\n")
			m.listNum = 0
		}
		defer func() {
			nl(m.buf)
			m.listNum = oldListNum
			fmt.Fprintf(m.buf, ".LE\n")
		}()
	case blackfriday.Item:
		nl(m.buf)
		if m.listNum > 0 {
			fmt.Fprintf(m.buf, ".LI %d.\n", m.listNum)
			m.listNum++
		} else {
			fmt.Fprintf(m.buf, ".LI –\n")
		}
	case blackfriday.HTMLSpan, blackfriday.HTMLBlock:
		fmt.Fprintf(m.buf, "<HTML>")
	case blackfriday.Image:
		dest := x.LinkData.Destination

		var (
			data []byte
			err  error
		)

		if strings.HasPrefix(string(dest), "http") {
			// Download the image file.
			resp, err := http.Get(string(dest))
			if err != nil {
				return err
			}
			data, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
		} else {
			data, err = ioutil.ReadFile(string(dest) + ".eps")
			if err != nil {
				return err
			}
		}

		i := bytes.Index(data, []byte("%%BoundingBox:"))
		if i < 0 {
			return fmt.Errorf("%s: no bounding box", dest)
		}
		data = data[i+len("%%BoundingBox:"):]
		j := bytes.IndexByte(data, '\n')
		if j < 0 {
			j = len(data)
		}
		f := strings.Fields(string(data[:j]))
		if len(f) != 4 {
			return fmt.Errorf("%s: bad bounding box %s", dest, data[:j])
		}
		x1, err1 := strconv.Atoi(f[0])
		y1, err2 := strconv.Atoi(f[1])
		x2, err3 := strconv.Atoi(f[2])
		y2, err4 := strconv.Atoi(f[3])
		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			return fmt.Errorf("%s: bad bounding box %s", dest, data[:j])
		}
		nl(m.buf)
		ptPerInch := m.doc.PtPerInch
		if ptPerInch == 0 {
			ptPerInch = 144
		}
		dx := float64(x2-x1) / ptPerInch
		dy := float64(y2-y1) / ptPerInch
		fmt.Fprintf(m.buf, ".sp -0.5\n")
		fmt.Fprintf(m.buf, ".ne %.2fi\n", dy)
		fmt.Fprintf(m.buf, ".PI %s.eps %.2fi,%.2fi,0i,%.2fi\n", dest, dy, dx, (4.5-dx)/2)
		fmt.Fprintf(m.buf, ".sp %.2fi\n", dy)

		return nil
	}

	for c := x.FirstChild; c != nil; c = c.Next {
		m.convert(c)
	}

	return nil
}

func (m *md2PDF) addFont(f int) {
	m.font |= f
	fmt.Fprintf(m.buf, "\\f%s", fontName[m.font])
}

func (m *md2PDF) subFont(f int) {
	m.font &^= f
	fmt.Fprintf(m.buf, "\\f%s", fontName[m.font])
}

func (m *md2PDF) Convert() (string, error) {
	// Initialize blackfriday.
	x := blackfriday.New(blackfriday.WithExtensions(blackfriday.CommonExtensions)).Parse(m.data)

	// Create a new buffer.
	m.buf = new(bytes.Buffer)

	// Write document data to the buffer.
	fmt.Fprintf(m.buf, ".mso thesis.me\n")
	fmt.Fprintf(m.buf, ".ds title %s\n", m.doc.Title)
	fmt.Fprintf(m.buf, ".ds subtitle %s\n", m.doc.Subtitle)

	// Write draft data to the buffer.
	draft := 0
	if m.doc.Draft {
		draft = 1
	}
	fmt.Fprintf(m.buf, ".nr draft %d\n", draft)

	// Write URL data to the buffer.
	fmt.Fprintf(m.buf, ".ds url %s/%s\n", m.doc.URI, m.doc.Name)

	// Write the document date to the buffer.
	if m.doc.Date.IsZero() {
		m.doc.Date = time.Now()
	}
	fmt.Fprintf(m.buf, ".ds date %s\n", m.doc.Date.Format("January 2, 2006"))

	// Start the document.
	fmt.Fprintf(m.buf, ".start\n")

	// Create a temporary working directory.
	tmpd, err := ioutil.TempDir("", "md2pdf")
	if err != nil {
		return "", err
	}
	// Cleanup the working directory when we are finished.
	defer os.RemoveAll(tmpd)

	// Create the name for the output file.
	draftSuffix := ""
	if m.doc.Draft {
		draftSuffix = "-DRAFT"
	}
	output := m.doc.Name + draftSuffix + ".pdf"

	// Convert the document.
	if err := m.convert(x); err != nil {
		return "", fmt.Errorf("convert: %v", err)
	}
	nl(m.buf)

	// Write the buffer to the .tr file.
	trOut := filepath.Join(tmpd, "_"+m.doc.Name+".tr")
	if err := ioutil.WriteFile(trOut, m.buf.Bytes(), 0666); err != nil {
		return "", err
	}
	defer os.Remove(trOut)

	// Create the file for troff Stdout.
	ditOut := filepath.Join(tmpd, "_"+m.doc.Name+".dit")
	f, err := os.Create(ditOut)
	if err != nil {
		return "", err
	}
	defer os.Remove(ditOut)

	// Run troff.
	cmd := exec.Command("troff", "-mpictures", trOut)
	cmd.Stdout = f
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("troff: %v", err)
	}
	f.Close()

	// Create the file for dpost Stdout.
	psOut := filepath.Join(tmpd, "_"+m.doc.Name+".ps")
	f, err = os.Create(psOut)
	if err != nil {
		return "", err
	}
	defer os.Remove(psOut)

	// Run dpost.
	cmd = exec.Command("dpost", ditOut)
	cmd.Stdout = f
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("dpost: %v", err)
	}

	// Run ps2pdf.
	cmd = exec.Command("ps2pdf", psOut, output)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ps2pdf: %v", err)
	}

	return output, nil
}

func maxLineWidth(literal []byte) ([]byte, int) {
	max := 0
	var buf bytes.Buffer
	lines := strings.Split(string(literal), "\n")
	for _, line := range lines[:len(lines)-1] {
		w := 0
		for _, r := range line {
			if r == '\t' {
				buf.WriteRune(' ')
				w++
				for w%4 != 0 {
					buf.WriteRune(' ')
					w++
				}
			} else {
				buf.WriteRune(r)
				w++
			}
		}
		buf.WriteRune('\n')
		if max < w {
			max = w
		}
	}
	return buf.Bytes(), max
}

func nl(buf *bytes.Buffer) {
	if buf.Len() > 0 && buf.Bytes()[buf.Len()-1] != '\n' {
		buf.WriteByte('\n')
	}
}

func literal(buf *bytes.Buffer, text []byte) {
	for len(text) > 0 {
		r, size := utf8.DecodeRune(text)
		text = text[size:]
		switch r {
		case '\\':
			buf.WriteString(`\`)
		case '.':
			if buf.Len() == 0 || buf.Bytes()[buf.Len()-1] == '\n' {
				buf.WriteString(`\&`)
			}
		case '\n':
			buf.WriteString(`\&`)
		}
		buf.WriteRune(r)
	}
}
