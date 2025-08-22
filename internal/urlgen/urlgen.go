package urlgen

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

type RequestType int

const (
	Stream RequestType = iota
	File
)

var (
	ErrExpiredStreamURL = errors.New("expired stream URL")
	ErrExpiredFileURL   = errors.New("expired file URL")
	ErrInvalidToken     = errors.New("invalid token")
	ErrInvalidData      = errors.New("invalid data")
)

type Generator struct {
	PublicURL string
	aead      cipher.AEAD
}

type Data struct {
	RequestType RequestType
	URL         string
	Playlist    string
	ChannelID   string
}

func NewGenerator(publicURL, secret string) (*Generator, error) {
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}
	return &Generator{PublicURL: publicURL, aead: aead}, nil
}

func (g *Generator) CreateURL(d Data, ttl time.Duration) (*url.URL, error) {
	buf := &bytes.Buffer{}

	buf.WriteByte(byte(d.RequestType))

	g.writeString(buf, d.URL)
	g.writeString(buf, d.Playlist)
	g.writeString(buf, d.ChannelID)

	var expireAt int64
	if ttl != 0 {
		expireAt = time.Now().Add(ttl).Unix()
	}
	binary.Write(buf, binary.LittleEndian, expireAt)

	hash := sha256.Sum256(buf.Bytes())
	nonce := hash[:g.aead.NonceSize()]

	ct := g.aead.Seal(nonce, nonce, buf.Bytes(), nil)
	token := base64.RawURLEncoding.EncodeToString(ct)
	ext := determineExtension(d)

	u, err := url.Parse(fmt.Sprintf("%s/%s/f%s", g.PublicURL, token, ext))
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}
	return u, nil
}

func (g *Generator) writeString(buf *bytes.Buffer, s string) {
	b := []byte(s)
	binary.Write(buf, binary.LittleEndian, uint16(len(b)))
	buf.Write(b)
}

func (g *Generator) readString(buf *bytes.Buffer) (string, error) {
	var length uint16
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return "", err
	}
	if buf.Len() < int(length) {
		return "", ErrInvalidData
	}
	data := make([]byte, length)
	buf.Read(data)
	return string(data), nil
}

func (g *Generator) Decrypt(token string) (*Data, error) {
	ct, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token: %w", err)
	}

	nonceSize := g.aead.NonceSize()
	if len(ct) < nonceSize {
		return nil, ErrInvalidToken
	}

	plain, err := g.aead.Open(nil, ct[:nonceSize], ct[nonceSize:], nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	buf := bytes.NewBuffer(plain)

	rt, err := buf.ReadByte()
	if err != nil {
		return nil, ErrInvalidData
	}

	url, err := g.readString(buf)
	if err != nil {
		return nil, ErrInvalidData
	}

	playlist, err := g.readString(buf)
	if err != nil {
		return nil, ErrInvalidData
	}

	channelID, err := g.readString(buf)
	if err != nil {
		return nil, ErrInvalidData
	}

	var expireAt int64
	if err := binary.Read(buf, binary.LittleEndian, &expireAt); err != nil {
		return nil, ErrInvalidData
	}

	if expireAt > 0 && time.Unix(expireAt, 0).Before(time.Now()) {
		if RequestType(rt) == Stream {
			return nil, ErrExpiredStreamURL
		}
		return nil, ErrExpiredFileURL
	}

	return &Data{
		RequestType: RequestType(rt),
		URL:         url,
		Playlist:    playlist,
		ChannelID:   channelID,
	}, nil
}

func determineExtension(d Data) string {
	if d.RequestType == Stream {
		return ".ts"
	}

	base, _, _ := strings.Cut(d.URL, "?")
	ext := filepath.Ext(base)
	if ext == "" {
		return ".file"
	}
	return ext
}
