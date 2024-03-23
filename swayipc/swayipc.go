package swayipc

import (
	"encoding/binary"
	"encoding/json"
	"net"
	"os"
)

const (
	MAGIC = "i3-ipc"
)

func encodeuint(x uint32) []byte {
	buf := make([]byte, 4)
	binary.NativeEndian.PutUint32(buf, x)

	return buf
}

func Getaddr() string {
	SOCK_ADDR := os.Getenv("I3SOCK")

	if SOCK_ADDR == "" {
		SOCK_ADDR = os.Getenv("SWAYSOCK")
	}

	if SOCK_ADDR == "" {
		panic("Socket file not in environment variables.")
	}
	return SOCK_ADDR
}

func Getsock(addr string) net.Conn {
	conn, err := net.Dial("unix", addr)
	if err != nil {
		panic(err)
	}
	return conn
}

func Pack(event_type uint32, encoded_payload []byte) []byte {
	magic_encoded := []byte(MAGIC)

	encoded_payload_length := encodeuint(uint32(len(encoded_payload)))
	encoded_event_type := encodeuint(event_type)

	header := append(magic_encoded, encoded_payload_length...)
	header = append(header, encoded_event_type...)

	msg := append(header, encoded_payload...)

	return msg
}

func Unpack(conn net.Conn) []byte {
	var header_struct []byte = make([]byte, 14)
	conn.Read(header_struct)

	var response_length uint32 = binary.NativeEndian.Uint32(header_struct[6:10])
	var response []byte = make([]byte, response_length)
	conn.Read(response)

	return response
}

func Subscribe(conn net.Conn, events []string) bool {
	payload, err := json.Marshal(events)
	if err != nil {
		panic(err)
	}

	msg := Pack(2, payload)
	conn.Write(msg)

	var result map[string]bool
	resp := Unpack(conn)
	json.Unmarshal(resp, &result)

	return result["success"]
}

func get_version(conn net.Conn) string {
	payload := []byte("")
	msg := Pack(7, payload)
	conn.Write(msg)
	var result map[string]any
	var resp []byte = Unpack(conn)
	json.Unmarshal(resp, &result)

	return result["human_readable"].(string)
}
