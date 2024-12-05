package discord

import "time"

// Message represents a Discord message
type Message struct {
	Content string  `json:"content,omitempty"`
	Embeds  []Embed `json:"embeds,omitempty"`
}

// Embed represents a Discord embed
type Embed struct {
	Title       string       `json:"title,omitempty"`
	Description string       `json:"description,omitempty"`
	Color       int          `json:"color,omitempty"`
	Fields      []EmbedField `json:"fields,omitempty"`
	Footer      *EmbedFooter `json:"footer,omitempty"`
	Timestamp   string       `json:"timestamp,omitempty"`
}

type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type EmbedFooter struct {
	Text string `json:"text"`
}

// Color constants
const (
	ColorGreen = 0x00FF00
	ColorRed   = 0xFF0000
	ColorBlue  = 0x0000FF
)

// NewEmbed creates a new Discord embed
func NewEmbed() *Embed {
	return &Embed{}
}

// SetTitle sets the embed title
func (e *Embed) SetTitle(title string) *Embed {
	e.Title = title
	return e
}

// SetDescription sets the embed description
func (e *Embed) SetDescription(desc string) *Embed {
	e.Description = desc
	return e
}

// SetColor sets the embed color
func (e *Embed) SetColor(color int) *Embed {
	e.Color = color
	return e
}

// AddField adds a field to the embed
func (e *Embed) AddField(name, value string, inline bool) *Embed {
	e.Fields = append(e.Fields, EmbedField{
		Name:   name,
		Value:  value,
		Inline: inline,
	})
	return e
}

// SetFooter sets the embed footer
func (e *Embed) SetFooter(text string) *Embed {
	e.Footer = &EmbedFooter{Text: text}
	return e
}

// SetTimestamp sets the embed timestamp
func (e *Embed) SetTimestamp(t time.Time) *Embed {
	e.Timestamp = t.Format(time.RFC3339)
	return e
}
