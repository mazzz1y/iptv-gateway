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
	streamTTL time.Duration
	fileTTL   time.Duration
}

type Data struct {
	RequestType RequestType
	URL         string
	Playlist    string
	ChannelID   string
	Hidden      bool
}

func NewGenerator(publicURL, secret string, streamTTL, fileTTL time.Duration) (*Generator, error) {
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}
	return &Generator{PublicURL: publicURL, aead: aead, streamTTL: streamTTL, fileTTL: fileTTL}, nil
}

func (g *Generator) CreateURL(d Data) (*url.URL, error) {
	buf := &bytes.Buffer{}

	buf.WriteByte(byte(d.RequestType))

	g.writeString(buf, d.URL)
	g.writeString(buf, d.Playlist)
	g.writeString(buf, d.ChannelID)

	if d.Hidden {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}

	binary.Write(buf, binary.LittleEndian, g.expireInt64(d.RequestType))

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

	u, err := g.readString(buf)
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

	hiddenByte, err := buf.ReadByte()
	if err != nil {
		return nil, ErrInvalidData
	}
	hidden := hiddenByte == 1

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
		URL:         u,
		Playlist:    playlist,
		ChannelID:   channelID,
		Hidden:      hidden,
	}, nil
}

func (g *Generator) expireInt64(r RequestType) int64 {
	switch r {
	case Stream:
		if g.streamTTL == 0 {
			return 0
		}
		return time.Now().Add(g.streamTTL).Unix()
	case File:
		if g.fileTTL == 0 {
			return 0
		}
		return time.Now().Add(g.fileTTL).Unix()
	default:
		return 0
	}
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
