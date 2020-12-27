package email

import (
	"github.com/matcornic/hermes/v2"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
)

const THEME_HTML_FILE = "html.tmpl"
const THEME_TEXT_FILE = "text.tmpl"

type Theme struct {
	name string
	html string
	text string
}

func (o *Theme) Name() string {
	return o.name
}

func (o *Theme) HTMLTemplate() string {
	return o.html
}

func (o *Theme) PlainTextTemplate() string {
	return o.text
}

type Themes struct {
	defaultTheme hermes.Theme
	themes       map[string]hermes.Theme
}

func (o *Themes) LoadTheme(theme string) (ret hermes.Theme) {
	ret = o.themes[theme]
	if ret == nil {
		logrus.Warnf("can't load the the '%v', use the theme '%v'", theme, o.defaultTheme.Name())
		ret = o.defaultTheme
	}
	return ret
}

func LoadThemes(themesFolder string) (ret *Themes) {
	defaultTheme := &hermes.Default{}
	ret = &Themes{
		defaultTheme: defaultTheme,
		themes: map[string]hermes.Theme{
			"default": defaultTheme, "flat": &hermes.Flat{},
		},
	}

	var files []os.FileInfo
	var logErr error
	if files, logErr = ioutil.ReadDir(themesFolder); logErr != nil {
		logrus.Infof("can't load external themes from '%v' because '%v'", themesFolder, logErr)
		return
	}

	for _, fileInfo := range files {
		var html, text []byte
		var loadErr error
		if fileInfo.IsDir() {
			themeFolder := filepath.Join(themesFolder, fileInfo.Name())
			if html, loadErr = ioutil.ReadFile(filepath.Join(themeFolder, THEME_HTML_FILE)); loadErr != nil {
				logErrorLoad(fileInfo, loadErr)
				continue
			}

			if text, loadErr = ioutil.ReadFile(filepath.Join(themeFolder, THEME_TEXT_FILE)); loadErr != nil {
				logErrorLoad(fileInfo, loadErr)
				continue
			}

			theme := &Theme{
				name: fileInfo.Name(),
				html: string(html),
				text: string(text),
			}
			logrus.Infof("the hermes theme loaded '%v'", theme.Name())
			ret.themes[theme.Name()] = theme
		}
	}
	return
}

func logErrorLoad(fileInfo os.FileInfo, loadErr error) {
	logrus.Infof("skip theme '%v' because '%v'", fileInfo.Name(), loadErr)
}
