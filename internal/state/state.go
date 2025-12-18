package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileMapping struct {
	NoteID string `json:"note_id"`
}

type Scope string

const (
	ScopeFolder Scope = "folder"
	ScopeNote   Scope = "note"
)

type Context struct {
	Name       string `json:"name"`
	Collection string `json:"collection"`
	Folder     string `json:"folder"`
}

type State struct {
	Contexts         map[string]Context                `json:"contexts"`
	CurrentContext   string                            `json:"current_context"`
	CurrentScope     Scope                             `json:"current_scope,omitempty"`
	CurrentNoteID    string                            `json:"current_note,omitempty"`
	CurrentNoteTitle string                            `json:"current_note_title,omitempty"`
	Files            map[string]map[string]FileMapping `json:"files"`
	path             string
}

func Load(root string) (*State, error) {
	path := filepath.Join(root, ".codestash", "state.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &State{
				Contexts:     make(map[string]Context),
				Files:        make(map[string]map[string]FileMapping),
				CurrentScope: ScopeFolder,
				path:         path,
			}, nil
		}
		return nil, fmt.Errorf("read state: %w", err)
	}

	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, fmt.Errorf("decode state: %w", err)
	}
	if st.Contexts == nil {
		st.Contexts = make(map[string]Context)
	}
	if st.Files == nil {
		st.Files = make(map[string]map[string]FileMapping)
	}
	if st.CurrentScope == "" {
		st.CurrentScope = ScopeFolder
	}
	if st.CurrentScope != ScopeNote {
		st.CurrentNoteID = ""
		st.CurrentNoteTitle = ""
	}
	st.path = path
	return &st, nil
}

func (s *State) Save() error {
	if s.path == "" {
		return errors.New("state path is not set")
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("write state: %w", err)
	}
	return nil
}

func (s *State) SetContext(name, collectionID, folderID string) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "default"
	}
	s.Contexts[name] = Context{
		Name:       name,
		Collection: collectionID,
		Folder:     folderID,
	}
	if s.CurrentContext == "" {
		s.CurrentContext = name
	}
	if s.CurrentScope == "" {
		s.CurrentScope = ScopeFolder
	}
}

func (s *State) SwitchContext(name string) error {
	if _, ok := s.Contexts[name]; !ok {
		return fmt.Errorf("context %q not found", name)
	}
	s.CurrentContext = name
	s.EnterFolderScope()
	return nil
}

func (s *State) Current() (Context, error) {
	ctx, ok := s.Contexts[s.CurrentContext]
	if !ok {
		return Context{}, errors.New("no active context; run `codestash init --folder ...` first")
	}
	return ctx, nil
}

func (s *State) SetFileMapping(ctxName, relativePath, noteID string) {
	if s.Files == nil {
		s.Files = make(map[string]map[string]FileMapping)
	}
	if s.Files[ctxName] == nil {
		s.Files[ctxName] = make(map[string]FileMapping)
	}
	rel := strings.ReplaceAll(relativePath, "\\", "/")
	s.Files[ctxName][rel] = FileMapping{NoteID: noteID}
}

func (s *State) GetFileMapping(ctxName, relativePath string) (FileMapping, bool) {
	if s.Files == nil {
		return FileMapping{}, false
	}
	files := s.Files[ctxName]
	if files == nil {
		return FileMapping{}, false
	}
	rel := strings.ReplaceAll(relativePath, "\\", "/")
	m, ok := files[rel]
	return m, ok
}

func (s *State) EnterFolderScope() {
	s.CurrentScope = ScopeFolder
	s.CurrentNoteID = ""
	s.CurrentNoteTitle = ""
}

func (s *State) EnterNoteScope(noteID, title string) error {
	noteID = strings.TrimSpace(noteID)
	if noteID == "" {
		return errors.New("note id is required")
	}
	s.CurrentScope = ScopeNote
	s.CurrentNoteID = noteID
	s.CurrentNoteTitle = strings.TrimSpace(title)
	return nil
}

func (s *State) Scope() Scope {
	if s.CurrentScope == "" {
		return ScopeFolder
	}
	return s.CurrentScope
}

func (s *State) CurrentNote() (string, string, error) {
	if s.Scope() != ScopeNote {
		return "", "", errors.New("not in note scope")
	}
	if strings.TrimSpace(s.CurrentNoteID) == "" {
		return "", "", errors.New("no active note selected")
	}
	return s.CurrentNoteID, s.CurrentNoteTitle, nil
}
