package urlgen

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type RequestType int

const (
	Stream RequestType = iota
	File
	ttl = time.Hour * 24 * 30
)

var ErrExpiredURL = errors.New("expired URL")

type Generator struct {
	PublicURL string
	aead      cipher.AEAD
}

type Data struct {
	RequestType RequestType
	URL         string
	Playlist    string
	ChannelID   string
	Created     time.Time
}

func NewGenerator(publicURL, secret string) (*Generator, error) {
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Generator{PublicURL: publicURL, aead: aead}, nil
}

func (g *Generator) CreateURL(d Data) (*url.URL, error) {
	if d.Created.IsZero() {
		d.Created = time.Now()
	}

	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)

	err := w.Write([]string{
		strconv.Itoa(int(d.RequestType)),
		d.URL,
		d.Playlist,
		d.ChannelID,
		strconv.FormatInt(d.Created.Unix(), 10),
	})
	if err != nil {
		return nil, err
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	nonce := make([]byte, g.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ct := g.aead.Seal(nonce, nonce, buf.Bytes(), nil)
	token := base64.RawURLEncoding.EncodeToString(ct)
	ext := determineExtension(d)
	return url.Parse(fmt.Sprintf("%s/%s/f%s", g.PublicURL, token, ext))
}

func (g *Generator) Decrypt(token string) (*Data, error) {
	ct, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}

	nonceSize := g.aead.NonceSize()
	if len(ct) < nonceSize {
		return nil, errors.New("invalid token")
	}

	plain, err := g.aead.Open(nil, ct[:nonceSize], ct[nonceSize:], nil)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(bytes.NewReader(plain))
	rec, err := r.Read()
	if err != nil {
		return nil, err
	}

	if len(rec) < 5 {
		return nil, errors.New("invalid data")
	}

	rt, err := strconv.Atoi(rec[0])
	if err != nil {
		return nil, err
	}

	ts, err := strconv.ParseInt(rec[4], 10, 64)
	if err != nil {
		return nil, err
	}

	created := time.Unix(ts, 0)
	if created.Add(ttl).Before(time.Now()) {
		return nil, ErrExpiredURL
	}

	return &Data{
		RequestType: RequestType(rt),
		URL:         rec[1],
		Playlist:    rec[2],
		ChannelID:   rec[3],
		Created:     created,
	}, nil
}

func determineExtension(d Data) string {
	if d.RequestType == Stream {
		return ".ts"
	}
	base := strings.SplitN(d.URL, "?", 2)[0]
	ext := filepath.Ext(base)
	if ext == "" {
		return ".file"
	}
	return ext
}
