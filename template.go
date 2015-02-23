package template

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/leeola/muta"
)

// Return an Options struct with Default Options
func NewOptions() Options {
	return Options{
		IgnoreTemplateErrors: false,
		IncludeFrontMatter:   true,
		CtxTemplateKeyword:   "template",
	}
}

type Options struct {
	// The keyword used on the Ctx map to find the template name.
	CtxTemplateKeyword string

	// Ignore any encountered template errors, and simply return
	// the raw data.
	IgnoreTemplateErrors bool

	// If found, include the front matter into the template's data
	IncludeFrontMatter bool
}

// Load all of the given paths as Templates, with names relative
// to the base name. Ie, `foo.tmpl`, `../bar.tmpl` and `partial/baz.tmpl`
func loadRelTemplates(ps []string) (*template.Template, error) {
	if len(ps) == 0 {
		return nil, errors.New("At least one path is required")
	}

	// TODO: Intelligently assign base to the most base path
	base := filepath.Dir(ps[0])

	var t *template.Template
	var lt *template.Template
	for _, tmpl := range ps {
		tmplName, err := filepath.Rel(base, tmpl)
		if err != nil {
			return nil, err
		}

		if t == nil {
			t = template.New(tmplName)
			lt = t
		} else {
			lt = t.New(tmplName)
		}

		b, err := ioutil.ReadFile(tmpl)
		if err != nil {
			return nil, err
		}

		_, err = lt.Parse(string(b))
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}

func newStreamRenderer(s string, opts Options) (
	*streamRenderer, error) {
	tmpls := []string{}
	filepath.Walk(s, func(p string, info os.FileInfo, err error) error {
		if filepath.Ext(p) == ".tmpl" {
			tmpls = append(tmpls, p)
		}
		return err
	})

	t, err := loadRelTemplates(tmpls)
	if err != nil {
		return &streamRenderer{}, err
	}

	return &streamRenderer{
		Opts:     opts,
		Template: t,
	}, nil
}

type streamRenderer struct {
	Opts Options

	Template *template.Template
}

func (sr *streamRenderer) Render(fi *muta.FileInfo,
	chunk []byte) (*muta.FileInfo, []byte, error) {

	if fi.Ctx[sr.Opts.CtxTemplateKeyword] == nil {
		// If there is no template specified, we can't do anything
		return fi, chunk, nil
	}

	t, ok := fi.Ctx[sr.Opts.CtxTemplateKeyword].(string)
	if ok == false {
		if sr.Opts.IgnoreTemplateErrors {
			return fi, chunk, nil
		} else {
			return nil, nil, errors.New(fmt.Sprintf(
				`Ctx["%s"] was not a string`, sr.Opts.CtxTemplateKeyword))
		}
	}

	tData := map[string]interface{}{
		"Content": string(chunk),
	}

	if sr.Opts.IncludeFrontMatter && fi.Ctx["frontmatter"] != nil {
		tData["frontmatter"] = fi.Ctx["frontmatter"]
	}

	var b bytes.Buffer
	err := sr.Template.ExecuteTemplate(&b, t, tData)
	if err != nil {
		return fi, chunk, err
	}

	chunk, _ = ioutil.ReadAll(&b)
	return fi, chunk, nil

}

func TemplateOpts(templatesPath string, opts Options) muta.Streamer {
	sr, err := newStreamRenderer(templatesPath, opts)
	var b bytes.Buffer
	return muta.NewEasyStreamer("template.Template", func(fi *muta.FileInfo,
		chunk []byte) (*muta.FileInfo, []byte, error) {

		switch {
		case err != nil:
			return fi, chunk, err

		case fi == nil:
			return nil, nil, nil

		case chunk == nil:
			chunk, _ := ioutil.ReadAll(&b)
			if len(chunk) == 0 {
				return fi, nil, nil
			}
			return sr.Render(fi, chunk)

		default:
			// Buffer all incoming data
			b.Write(chunk)
			return nil, nil, nil
		}
	})
}

// An alias for TemplateOpts, with default options and user data
// Not Implemented
func TemplateData(p string, data interface{}) {
}

// An alias for TemplateOpts, with default Options.
func Template(p string) muta.Streamer {
	return TemplateOpts(p, NewOptions())
}
