package email

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-ee/utils/encrypt"
	"github.com/matcornic/hermes/v2"
	"github.com/sirupsen/logrus"
)

const TIME_FORMAT = "2006_01_02__15_04_05_000"

const DEFAULT_FOLDER_PERM os.FileMode = 0777
const DEFAULT_FILE_PERM os.FileMode = 0644

type EmailData struct {
	To       []string
	Name     string
	Subject  string
	Url      string
	Markdown string
	Theme    string
}

func (o *EmailData) ToAsString() string {
	return strings.Join(o.To, ",")
}

type Engine struct {
	hermes.Hermes
	hermes.Body
	Sender
	*encrypt.Encryptor
	hermesMu sync.RWMutex

	emailsFolder string
	storeEmails  bool

	timeFormat string
	folderPerm os.FileMode
	filePerm   os.FileMode

	themes *Themes
}

func NewEngine(config *EngineConfig) (ret *Engine, err error) {

	ret = &Engine{
		Hermes:    config.Hermes.ToHermes(),
		Body:      config.Hermes.Body.ToHermesBody(),
		Sender:    config.Sender,
		hermesMu:  sync.RWMutex{},

		themes:       LoadThemes(config.Hermes.ThemesFolder),
		emailsFolder: config.EmailsFolder,
		storeEmails:  config.StoreEmails,

		timeFormat: TIME_FORMAT,
		folderPerm: DEFAULT_FOLDER_PERM,
		filePerm:   DEFAULT_FILE_PERM,
	}

	if err = ret.checkAndCreateStorage(); err != nil {
		return
	}

	return
}

func (o *Engine) Send(emailData *EmailData) (err error) {
	logrus.Debugf("Send, %v, %v", emailData.To, emailData.Subject)
	o.hermesMu.RLock()
	defer o.hermesMu.RUnlock()

	o.Hermes.Theme = o.themes.LoadTheme(emailData.Theme)

	var message *Message
	if message, err = o.BuildEmail(
		emailData.ToAsString(), emailData.Subject, o.BuildBody(emailData.Markdown)); err != nil {
		return
	}

	if o.storeEmails {
		o.storeEmail("", message, emailData)
	}

	err = o.Sender.Send(message)
	return
}

func (o *Engine) BuildBody(markdown string) (ret hermes.Body) {

	ret = o.Body
	ret.FreeMarkdown = hermes.Markdown(markdown)
	return ret
}

func (o *Engine) BuildEmail(to, subject string, body hermes.Body) (ret *Message, err error) {

	hEmail := hermes.Email{
		Body: body,
	}

	ret = &Message{To: to, Subject: subject}
	if ret.PlainText, err = o.GeneratePlainText(hEmail); err == nil {
		ret.HTML, err = o.GenerateHTML(hEmail)
	}
	return
}

func (o *Engine) storeEmail(label string, htmlMessage *Message, emailData *EmailData) (err error) {
	fileData := []byte(fmt.Sprintf("request:\n%v\n\nmessage:\n%v\n", label, htmlMessage.PlainText))
	filePath := filepath.Clean(fmt.Sprintf("%v/%v/%v_%v.txt",
		o.emailsFolder, strings.Join(emailData.To, "_"), emailData.Subject, time.Now().Format(o.timeFormat)))

	if err = os.MkdirAll(filepath.Dir(filePath), o.folderPerm); err != nil {
		return
	}

	if err = ioutil.WriteFile(filePath, fileData, o.filePerm); err != nil {
		logrus.Warnf("can't write '%v', %v", filePath, err)
	} else {
		logrus.Debugf("written '%v', bytes=%v", filePath, len(fileData))
	}
	return
}

func (o *Engine) checkAndCreateStorage() (err error) {
	if o.storeEmails && o.emailsFolder != "" {
		if err = os.MkdirAll(o.emailsFolder, 0755); err == nil {
			o.storeEmails = true
			logrus.Infof("use the storage path: %v", o.emailsFolder)
		} else {
			logrus.Infof("can't create the storage path '%v': %v", o.emailsFolder, err)
		}
	}
	return
}
