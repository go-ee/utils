package email

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/matcornic/hermes/v2"
	"gopkg.in/yaml.v2"
)

type Product struct {
	Name        string `yaml:"name", envconfig:"PRODUCT_NAME"`
	Link        string `yaml:"link", envconfig:"PRODUCT_LINK"`
	Logo        string `yaml:"logo", envconfig:"PRODUCT_LOGO"`
	Copyright   string `yaml:"copyright", envconfig:"PRODUCT_COPYRIGHT"`
	TroubleText string `yaml:"troubleText", envconfig:"PRODUCT_TROUBLE_TEXT"`
}

func (o *Product) ToHermesProduct() hermes.Product {
	return hermes.Product{
		Name:        o.Name,
		Link:        o.Link,
		Logo:        o.Logo,
		Copyright:   o.Copyright,
		TroubleText: o.TroubleText,
	}
}

type Body struct {
	Name       string   `yaml:"name"`       // The name of the contacted person
	Intros     []string `yaml:"intros"`     // Intro sentences, first displayed in the email
	Dictionary []Entry  `yaml:"dictionary"` // A list of key+value (useful for displaying parameters/settings/personal info)
	Table      Table    `yaml:"table"`      // Table is an table where you can put data (pricing grid, a bill, and so on)
	Actions    []Action `yaml:"actions"`    // Actions are a list of actions that the user will be able to execute via a button click
	Outros     []string `yaml:"outros"`     // Outro sentences, last displayed in the email
	Greeting   string   `yaml:"greeting"`   // Greeting for the contacted person (default to 'Hi')
	Signature  string   `yaml:"signature"`  // Signature for the contacted person (default to 'Yours truly')
	Title      string   `yaml:"title"`      // Title replaces the greeting+name when set
}

func (o *Body) ToHermesBody() hermes.Body {
	return hermes.Body{
		Name:      o.Name,
		Intros:    o.Intros,
		Outros:    o.Outros,
		Greeting:  o.Greeting,
		Signature: o.Signature,
		Title:     o.Title,
	}
}

// Entry is a simple entry of a map
// Allows using a slice of entries instead of a map
// Because Golang maps are not ordered
type Entry struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

func (o *Entry) ToHermesEntry() hermes.Entry {
	return hermes.Entry{
		Key:   o.Key,
		Value: o.Value,
	}
}

// Table is an table where you can put data (pricing grid, a bill, and so on)
type Table struct {
	Data    [][]Entry `yaml:"data"`    // Contains data
	Columns Columns   `yaml:"columns"` // Contains meta-data for display purpose (width, alignement)
}

func (o *Table) ToHermesTable() *hermes.Table {
	var data [][]hermes.Entry
	if o.Data != nil {
		data = make([][]hermes.Entry, len(o.Data))
		for i, item := range o.Data {
			data[i] = make([]hermes.Entry, len(item))
			for j, entry := range item {
				data[i][j] = entry.ToHermesEntry()
			}
		}
	}

	return &hermes.Table{
		Data:    data,
		Columns: o.Columns.ToHermesColumns(),
	}
}

// Columns contains meta-data for the different columns
type Columns struct {
	CustomWidth     map[string]string
	CustomAlignment map[string]string
}

func (o *Columns) ToHermesColumns() hermes.Columns {
	return hermes.Columns{
		CustomWidth:     o.CustomWidth,
		CustomAlignment: o.CustomAlignment,
	}
}

// Action is anything the user can act on (i.e., click on a button, view an invite code)
type Action struct {
	Instructions string `yaml:"instructions"`
	Button       Button `yaml:"button"`
	InviteCode   string `yaml:"inviteCode"`
}

func (o *Action) ToHermesAction() *hermes.Action {
	return &hermes.Action{
		Instructions: o.Instructions,
		Button:       o.Button.ToHermesButton(),
		InviteCode:   o.InviteCode,
	}
}

// Button defines an action to launch
type Button struct {
	Color     string `yaml:"color"`
	TextColor string `yaml:"textColor"`
	Text      string `yaml:"text"`
	Link      string `yaml:"link"`
}

func (o *Button) ToHermesButton() hermes.Button {
	return hermes.Button{
		Color:     o.Color,
		TextColor: o.TextColor,
		Text:      o.Text,
		Link:      o.Link,
	}
}

type Hermes struct {
	Product            Product `yaml:"product"`
	DisableCSSInlining bool    `yaml:"disableCSSInlining"`
	Body               Body    `yaml:"body"`
}

func (o *Hermes) ToHermes() hermes.Hermes {
	return hermes.Hermes{
		Product:            o.Product.ToHermesProduct(),
		DisableCSSInlining: o.DisableCSSInlining,
	}
}

type EngineConfig struct {
	Root              string `yaml:"root", envconfig:"PATH_ROOT"`
	PathStorage       string `yaml:"pathStorage", envconfig:"PATH_STORAGE"`
	StoreEmails       bool   `yaml:"storeEmails", envconfig:"STORE_EMAILS"`
	EncryptPassphrase string `yaml:"encryptPassphrase", envconfig:"ENCRYPT_PASSPHRASE"`
	Hermes            Hermes `yaml:"hermes"`
	Sender            Sender `yaml:"sender"`
}

func (o *EngineConfig) Setup() (err error) {
	if o.EncryptPassphrase == "" {
		o.EncryptPassphrase = o.Sender.Email + o.Sender.SMTP.Password
	}

	err = o.Sender.Setup()

	return
}

func EngineConfigLoad(configFileYaml string, cfg *EngineConfig) (err error) {
	var file *os.File
	if file, err = os.Open(configFileYaml); err != nil {
		return
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err = decoder.Decode(cfg); err != nil {
		return
	}
	err = envconfig.Process("", cfg)

	cfg.Setup()
	return
}

func (o *EngineConfig) WriteFileYaml(configFileYaml string) (err error) {
	var file *os.File
	if file, err = os.Create(configFileYaml); err != nil {
		return
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	if err = encoder.Encode(o); err != nil {
		return
	}
	err = encoder.Close()
	return
}

func EngineCinfigDefault() (ret *EngineConfig) {
	ret = &EngineConfig{
		Root:        "templates",
		PathStorage: "storage",
		Sender: Sender{
			Email:    "info@example.com",
			Identity: "Info",
			SMTP: SMTP{
				Server:   "mail.example.com",
				Port:     465,
				User:     "info@example.com",
				Password: "changeMe",
			},
		},
		Hermes: Hermes{
			Product: Product{
				Name:      "ExampleProduct",
				Link:      "www.example.com",
				Logo:      "www.example.com/logo.svg",
				Copyright: "@ Example",
			},
			Body: Body{
				Name:      "",
				Intros:    []string{"Intro 1", "Intro 2"},
				Outros:    []string{"Outro 1", "Outro 2"},
				Greeting:  "Sei gegrüßt, ",
				Signature: "Wir freuen uns",
				Title:     "Herzlichen Glückwunsch",
			},
		},
	}
	return
}


/*

Hinweis: Du bekommst am Tag des jeweiligen Kurses auf deine E-Mail einen Link zum Online-Seminar zu welchemdu dich angemeldet hast. Bitte denke daran die Zoom-Software zu installieren. Hier sind nochmal die Links dazu: Adroid: https://play.google.com/store/apps/details?id=us.zoom.videomeetings&hl=de&gl=USApple: https://apps.apple.com/de/app/zoom-cloud-meetings/id546505307PC: https://zoom.us/support/downloadIch freue mich auf deine Teilnahme,Dr. Leo Frank-- Diese Anmeldung wurde von der Webseite "Seminar Berufung aus christlicher Sicht" (http://vongottberufen.de)gesendet.



 */