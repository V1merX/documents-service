package domain

import (
	"time"
	"unicode/utf8"
	"uuid"
)

type Document struct {
	id      uuid.UUID
	owner   Login
	name    string
	mime    string
	file    bool
	public  bool
	json    []byte
	content []byte
	created time.Time
	grant   []Login
}

func NewDocument(
	owner Login,
	name, mime string,
	file, public bool,
	json, content []byte,
	grant []Login,
) (*Document, error) {
	if utf8.RuneCountInString(name) == 0 {
		return nil, ErrEmptyFileName
	}

	if mime == "" && file {
		return nil, ErrRequiredMime
	}

	return &Document{
		id:      uuid.New(),
		owner:   owner,
		name:    name,
		mime:    mime,
		file:    file,
		public:  public,
		json:    json,
		content: content,
		created: time.Now().UTC(),
		grant:   grant,
	}, nil
}

func (d *Document) ID() uuid.UUID      { return d.id }
func (d *Document) Owner() Login       { return d.owner }
func (d *Document) Name() string       { return d.name }
func (d *Document) Mime() string       { return d.mime }
func (d *Document) File() bool         { return d.file }
func (d *Document) Public() bool       { return d.public }
func (d *Document) Created() time.Time { return d.created }
func (d *Document) JSON() []byte       { return d.json }
func (d *Document) Content() []byte    { return d.content }
func (d *Document) Grant() []Login {
	grant := make([]Login, len(d.grant))
	copy(grant, d.grant)
	return grant
}

func (d *Document) HasAccess(login Login) bool {
	if d.Public() {
		return true
	}

	if d.owner == login {
		return true
	}

	for _, l := range d.grant {
		if l.Value() == login.Value() {
			return true
		}
	}

	return false
}

func RestoreDocument(
	id uuid.UUID,
	owner string,
	name, mime string,
	file, public bool,
	json, content []byte,
	created time.Time,
	grant []string,
) *Document {
	g := make([]Login, 0, len(grant))
	for _, l := range grant {
		g = append(g, Login{value: l})
	}

	return &Document{
		id:      id,
		owner:   Login{value: owner},
		name:    name,
		mime:    mime,
		file:    file,
		public:  public,
		json:    json,
		content: content,
		created: created,
		grant:   g,
	}
}
