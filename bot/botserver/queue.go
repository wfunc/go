package main

import (
    "bufio"
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "sync"
)

type BacklogMessage struct {
    Type      string `json:"type"`   // "text" or "html"
    Content   string `json:"content"`
    Token     string `json:"token,omitempty"`
    ChatID    int64  `json:"chat_id,omitempty"`
    CreatedAt int64  `json:"created_at"`
}

type Backlog struct {
    path string
    mu   sync.Mutex
}

func NewBacklog(path string) (*Backlog, error) {
    if path == "" {
        return nil, errors.New("backlog path empty")
    }
    if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
        return nil, err
    }
    // touch
    f, err := os.OpenFile(path, os.O_CREATE, 0o644)
    if err == nil {
        f.Close()
    }
    return &Backlog{path: path}, nil
}

func (b *Backlog) Append(m BacklogMessage) error {
    b.mu.Lock()
    defer b.mu.Unlock()
    f, err := os.OpenFile(b.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
    if err != nil {
        return err
    }
    defer f.Close()
    enc, _ := json.Marshal(m)
    if _, err = f.Write(append(enc, '\n')); err != nil {
        return err
    }
    return nil
}

func (b *Backlog) LoadAll() ([]BacklogMessage, error) {
    b.mu.Lock()
    defer b.mu.Unlock()
    f, err := os.OpenFile(b.path, os.O_RDONLY|os.O_CREATE, 0o644)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    var out []BacklogMessage
    sc := bufio.NewScanner(f)
    for sc.Scan() {
        line := sc.Bytes()
        if len(line) == 0 {
            continue
        }
        var m BacklogMessage
        if err := json.Unmarshal(line, &m); err == nil {
            out = append(out, m)
        }
    }
    return out, nil
}

func (b *Backlog) Rewrite(msgs []BacklogMessage) error {
    b.mu.Lock()
    defer b.mu.Unlock()
    tmp := b.path + ".tmp"
    f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
    if err != nil {
        return err
    }
    enc := json.NewEncoder(f)
    for _, m := range msgs {
        if err := enc.Encode(&m); err != nil {
            f.Close()
            return err
        }
    }
    f.Close()
    return os.Rename(tmp, b.path)
}

func (b *Backlog) Count() (int, error) {
    msgs, err := b.LoadAll()
    if err != nil {
        return 0, err
    }
    return len(msgs), nil
}

func (b *Backlog) Replay(sendTextFn func(token string, chatID int64, msg string) error, sendHTMLFn func(token string, chatID int64, msg string) error) (int, int, error) {
    msgs, err := b.LoadAll()
    if err != nil {
        return 0, 0, err
    }
    if len(msgs) == 0 {
        return 0, 0, nil
    }
    var remain []BacklogMessage
    var ok, fail int
    for _, m := range msgs {
        var e error
        switch m.Type {
        case "text":
            e = sendTextFn(m.Token, m.ChatID, m.Content)
        case "html":
            e = sendHTMLFn(m.Token, m.ChatID, m.Content)
        default:
            e = fmt.Errorf("unknown type: %s", m.Type)
        }
        if e != nil {
            remain = append(remain, m)
            fail++
        } else {
            ok++
        }
    }
    if err := b.Rewrite(remain); err != nil {
        return ok, fail, err
    }
    return ok, fail, nil
}

