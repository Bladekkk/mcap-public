package service

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

type RCON struct {
	host     string
	port     int
	password string
	timeout  time.Duration
	conn     net.Conn
}

func NewRCON(host string, port int, password string, timeout time.Duration) *RCON {
	return &RCON{
		host:     host,
		port:     port,
		password: password,
		timeout:  timeout,
	}
}

func (r *RCON) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", r.host, r.port), r.timeout)
	if err != nil {
		return err
	}

	r.conn = conn
	return r.authenticate()
}

func (r *RCON) authenticate() error {
	resp, err := r.sendCommandInternal(3, r.password)
	if err != nil {
		return err
	}
	// Ответ должен содержать session ID
	if len(resp) < 4 {
		return fmt.Errorf("invalid auth response")
	}
	return nil
}

func (r *RCON) sendCommand(cmd string) (string, error) {
	return r.sendCommandInternal(4, cmd)
}

func (r *RCON) sendCommandInternal(requestID int32, command string) (string, error) {
	if r.conn == nil {
		return "", fmt.Errorf("not connected")
	}

	// RCON packet structure:
	// 4 bytes: size (excluding header)
	// 4 bytes: request id
	// 4 bytes: packet type (0=command, 3=auth, 4=command)
	// N bytes: command string + \x00\x00
	// Padding to multiple of 4

	data := []byte(command)
	packet := make([]byte, 0)

	// размер = 4 (requestID) + 4 (type) + len(data) + 2 (double null) - 4 (header size)
	size := int32(4 + 4 + len(data) + 2)
	packet = append(packet, encodeInt32(size)...)
	packet = append(packet, encodeInt32(requestID)...)
	packet = append(packet, encodeInt32(4)...) // command type
	packet = append(packet, data...)
	packet = append(packet, 0, 0)

	// Padding to 4-byte boundary
	for len(packet)%4 != 0 {
		packet = append(packet, 0)
	}

	if _, err := r.conn.Write(packet); err != nil {
		return "", err
	}

	// Читаем ответ
	return r.readResponse(requestID)
}

func (r *RCON) readResponse(expectedID int32) (string, error) {
	header := make([]byte, 8)
	if _, err := r.conn.Read(header); err != nil {
		return "", err
	}

	size := int32(decodeInt32(header[:4]))
	requestID := int32(decodeInt32(header[4:8]))

	if requestID != expectedID {
		return "", fmt.Errorf("request ID mismatch: got %d, expected %d", requestID, expectedID)
	}

	// Читаем тип и тело
	typeBytes := make([]byte, 4)
	if _, err := r.conn.Read(typeBytes); err != nil {
		return "", err
	}

	bodyLen := size - 14 // size - (requestID + type + terminator)
	if bodyLen < 0 {
		bodyLen = 0
	}

	body := make([]byte, bodyLen)
	if _, err := r.conn.Read(body); err != nil {
		return "", err
	}

	// Убираем double null terminator
	result := strings.TrimRight(string(body), "\x00")
	return result, nil
}

func (r *RCON) Close() error {
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

func encodeInt32(n int32) []byte {
	return []byte{byte(n & 0xFF), byte((n >> 8) & 0xFF), byte((n >> 16) & 0xFF), byte((n >> 24) & 0xFF)}
}

func decodeInt32(b []byte) int32 {
	return int32(b[0]) | int32(b[1])<<8 | int32(b[2])<<16 | int32(b[3])<<24
}
