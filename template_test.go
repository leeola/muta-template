package template

import (
	"path/filepath"
	"testing"

	"github.com/leeola/muta"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLoadRelTemplates(t *testing.T) {
	tmplDir := filepath.Join("_test", "fixtures", "templates")

	Convey("Should load templates in the given path", t, func() {
		t, err := loadRelTemplates([]string{
			filepath.Join(tmplDir, "plain.tmpl"),
			filepath.Join(tmplDir, "partials", "partial.tmpl"),
		})
		So(err, ShouldBeNil)
		So(t, ShouldNotBeNil)
		So(t.Lookup("plain.tmpl"), ShouldNotBeNil)
		So(t.Lookup("partials/partial.tmpl"), ShouldNotBeNil)
	})

	// TODO: sr.Templates.ExecuteTemplate(&x, "Name") only works with the
	// basename of the loaded templates. With that said, at this time
	// i'm not sure how to
	Convey("Should store template names relative to the most "+
		"base path", t, nil)
}

func TestNewStreamRenderer(t *testing.T) {
	tmplDir := filepath.Join("_test", "fixtures", "templates")

	Convey("Should load templates in the given path", t, func() {
		sr, err := newStreamRenderer(tmplDir, NewOptions())
		So(err, ShouldBeNil)
		So(sr.Template, ShouldNotBeNil)
		So(sr.Template.Lookup("plain.tmpl"), ShouldNotBeNil)
		So(sr.Template.Lookup("partials/partial.tmpl"), ShouldNotBeNil)
		So(sr.Template.Lookup("import.tmpl"), ShouldNotBeNil)
	})

}

func TestRender(t *testing.T) {
	tmplDir := filepath.Join("_test", "fixtures", "templates")

	Convey("Should ignore FileInfo's without template Ctx", t, func() {
		sr, _ := newStreamRenderer(tmplDir, NewOptions())
		fi := &muta.FileInfo{Ctx: map[string]interface{}{}}
		_, chunk, err := sr.Render(fi, []byte("foo"))
		So(err, ShouldBeNil)
		So(string(chunk), ShouldResemble, "foo")
	})

	Convey(`Should render the Ctx["template"]`, t, func() {
		sr, _ := newStreamRenderer(tmplDir, NewOptions())
		fi := &muta.FileInfo{Ctx: map[string]interface{}{
			"template": "plain.tmpl",
		}}
		_, chunk, err := sr.Render(fi, []byte("foo"))
		So(err, ShouldBeNil)
		So(string(chunk), ShouldResemble, "<div>Plain</div>\n")
	})

	Convey("Should add chunk with the Content keyword", t, func() {
		sr, _ := newStreamRenderer(tmplDir, NewOptions())
		fi := &muta.FileInfo{Ctx: map[string]interface{}{
			"template": "content.tmpl",
		}}
		_, chunk, err := sr.Render(fi, []byte("foo"))
		So(err, ShouldBeNil)
		So(string(chunk), ShouldResemble, "<div>Content foo</div>\n")
	})

	Convey("Should add frontmatter with the FrontMatter keyword", t, func() {
		sr, _ := newStreamRenderer(tmplDir, NewOptions())
		fi := &muta.FileInfo{Ctx: map[string]interface{}{
			"template":    "frontmatter.tmpl",
			"frontmatter": struct{ Bar string }{Bar: "bar"},
		}}
		_, chunk, err := sr.Render(fi, []byte("foo"))
		So(err, ShouldBeNil)
		So(string(chunk), ShouldResemble,
			"<div>Content foo, FrontMatter bar</div>\n")
	})

	Convey("Work with imports", t, func() {
		sr, _ := newStreamRenderer(tmplDir, NewOptions())
		fi := &muta.FileInfo{Ctx: map[string]interface{}{
			"template": "import.tmpl",
		}}
		_, chunk, err := sr.Render(fi, []byte("foo"))
		So(err, ShouldBeNil)
		So(string(chunk), ShouldResemble,
			"<div>Import <div>Partial</div>\n</div>\n")
	})
}

func TestTemplate(t *testing.T) {
	tmplDir := filepath.Join("_test", "fixtures", "templates")

	Convey("Should render templates with nil content", t, func() {
		s := Template(tmplDir).Stream
		oFi := muta.NewFileInfo("foo")
		oFi.Ctx["template"] = "plain.tmpl"
		fi, chunk, err := s(oFi, nil)
		So(err, ShouldBeNil)
		So(fi, ShouldEqual, oFi)
		So(string(chunk), ShouldEqual, "<div>Plain</div>\n")

		fi, chunk, err = s(oFi, nil)
		So(err, ShouldBeNil)
		So(fi, ShouldEqual, oFi)
		So(chunk, ShouldBeNil)
	})
}
