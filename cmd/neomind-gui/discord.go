package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/narkomaniya/neomind/internal/core"
)

// Minimal Discord Rich Presence client. Discord exposes a local IPC socket;
// speaking the framed JSON protocol directly keeps the editor dependency-free.

type discordClient struct {
	mu        sync.Mutex
	clientID  string
	asset     string
	project   string
	language  string
	file      string
	started   int64
	conn      io.ReadWriteCloser
	connected bool
	disabled  bool
	update    chan struct{}
}

var discord = &discordClient{update: make(chan struct{}, 1)}

// Discord accepts a direct image URL as an activity asset, so the NeoMind icon
// works even before a custom asset is uploaded to the Developer Portal.
const discordDefaultAsset = "https://raw.githubusercontent.com/narkomaniya/neomind/main/assets/neomind-discord.png"

type discordConfig struct {
	ClientID string `json:"client_id"`
	Asset    string `json:"asset"`
}

func discordConfigPath() string {
	dir := core.DataDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, "discord.json")
}

func discordLoad() {
	path := discordConfigPath()
	if path == "" {
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cfg discordConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return
	}
	discord.mu.Lock()
	discord.clientID = strings.TrimSpace(cfg.ClientID)
	discord.asset = strings.TrimSpace(cfg.Asset)
	if discord.asset == "" {
		discord.asset = discordDefaultAsset
	}
	discord.mu.Unlock()
	discord.signal()
}

func discordSave(clientID, asset string) error {
	path := discordConfigPath()
	if path == "" {
		return fmt.Errorf("no data folder")
	}
	clientID = strings.TrimSpace(clientID)
	asset = strings.TrimSpace(asset)
	if clientID != "" && len(clientID) < 15 {
		return fmt.Errorf("invalid Discord client id")
	}
	if asset == "" {
		asset = discordDefaultAsset
	}
	b, err := json.Marshal(discordConfig{ClientID: clientID, Asset: asset})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return err
	}
	discord.mu.Lock()
	discord.clientID = clientID
	discord.asset = asset
	if discord.conn != nil {
		_ = discord.conn.Close()
		discord.conn = nil
		discord.connected = false
	}
	discord.mu.Unlock()
	discord.signal()
	return nil
}

func discordStatus() string {
	discord.mu.Lock()
	defer discord.mu.Unlock()
	return fmt.Sprintf(`{"connected":%t,"client_id":%q,"asset":%q,"disabled":%t}`,
		discord.connected, discord.clientID, discord.asset, discord.disabled)
}

func discordSetActivity(project, language, file string, started int64) {
	discord.mu.Lock()
	discord.project = project
	discord.language = language
	discord.file = file
	discord.started = started
	discord.mu.Unlock()
	discord.signal()
}

func discordDisable() {
	discord.mu.Lock()
	discord.disabled = true
	if discord.conn != nil {
		_ = discord.conn.Close()
		discord.conn = nil
	}
	discord.connected = false
	discord.mu.Unlock()
}

func (d *discordClient) signal() {
	select {
	case d.update <- struct{}{}:
	default:
	}
}

func discordDial() (io.ReadWriteCloser, error) {
	if runtime.GOOS == "windows" {
		var last error
		for i := 0; i < 10; i++ {
			path := fmt.Sprintf(`\\.\pipe\discord-ipc-%d`, i)
			f, err := os.OpenFile(path, os.O_RDWR, 0)
			if err == nil {
				return f, nil
			}
			last = err
		}
		if last == nil {
			last = fmt.Errorf("Discord is not running")
		}
		return nil, last
	}
	dirs := []string{os.Getenv("XDG_RUNTIME_DIR"), "/tmp", os.TempDir()}
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		for i := 0; i < 10; i++ {
			path := filepath.Join(dir, fmt.Sprintf("discord-ipc-%d", i))
			c, err := net.DialTimeout("unix", path, time.Second)
			if err == nil {
				return c, nil
			}
		}
	}
	return nil, fmt.Errorf("Discord is not running")
}

func writeDiscordFrame(w io.Writer, op uint32, payload []byte) error {
	// Discord's named pipe is message-oriented on Windows: header and body must
	// go out as one write, otherwise the peer sees two broken frames.
	frame := make([]byte, 8+len(payload))
	binary.LittleEndian.PutUint32(frame[:4], op)
	binary.LittleEndian.PutUint32(frame[4:], uint32(len(payload)))
	copy(frame[8:], payload)
	_, err := w.Write(frame)
	return err
}

func readDiscordFrame(r io.Reader) (uint32, []byte, error) {
	header := make([]byte, 8)
	if _, err := io.ReadFull(r, header); err != nil {
		return 0, nil, err
	}
	op := binary.LittleEndian.Uint32(header[:4])
	size := binary.LittleEndian.Uint32(header[4:])
	if size > 1<<20 {
		return 0, nil, fmt.Errorf("Discord frame too large")
	}
	payload := make([]byte, size)
	if _, err := io.ReadFull(r, payload); err != nil {
		return 0, nil, err
	}
	return op, payload, nil
}

func discordHandshake(conn io.ReadWriteCloser, clientID string) error {
	payload, _ := json.Marshal(map[string]interface{}{"v": 1, "client_id": clientID})
	if err := writeDiscordFrame(conn, 0, payload); err != nil {
		return err
	}
	type result struct {
		err error
	}
	ch := make(chan result, 1)
	go func() {
		op, data, err := readDiscordFrame(conn)
		if err != nil {
			ch <- result{err}
			return
		}
		if op != 1 {
			ch <- result{fmt.Errorf("unexpected Discord handshake opcode %d", op)}
			return
		}
		var reply struct {
			Evt  string `json:"evt"`
			Data struct {
				User map[string]interface{} `json:"user"`
			} `json:"data"`
		}
		if err := json.Unmarshal(data, &reply); err != nil {
			ch <- result{err}
			return
		}
		if reply.Evt != "READY" {
			ch <- result{fmt.Errorf("Discord handshake returned %s", reply.Evt)}
			return
		}
		ch <- result{}
	}()
	select {
	case r := <-ch:
		return r.err
	case <-time.After(5 * time.Second):
		return fmt.Errorf("Discord handshake timed out")
	}
}

func (d *discordClient) push() {
	d.mu.Lock()
	if d.conn == nil || d.clientID == "" {
		d.mu.Unlock()
		return
	}
	conn := d.conn
	details := d.project
	if details == "" {
		details = "NeoMind"
	}
	state := d.language
	if d.file != "" {
		if state == "" {
			state = d.file
		} else {
			state += " · " + d.file
		}
	}
	if len(details) > 120 {
		details = details[:120]
	}
	if len(state) > 120 {
		state = state[:120]
	}
	act := map[string]interface{}{
		"details": details,
		"state":   state,
	}
	if d.asset != "" {
		act["assets"] = map[string]string{"large_image": d.asset, "large_text": "NeoMind"}
	}
	if d.started > 0 {
		act["timestamps"] = map[string]int64{"start": d.started / 1000}
	}
	req := map[string]interface{}{
		"cmd": "SET_ACTIVITY",
		"args": map[string]interface{}{
			"pid":      os.Getpid(),
			"activity": act,
		},
		"nonce": strconv.FormatInt(time.Now().UnixNano(), 10),
	}
	b, err := json.Marshal(req)
	if err != nil {
		d.mu.Unlock()
		return
	}
	d.mu.Unlock()
	// Never hold the state lock across a pipe write: a stalled Discord pipe
	// must not freeze the whole UI or the other bindings.
	if err := writeDiscordFrame(conn, 1, b); err != nil {
		_ = conn.Close()
		d.mu.Lock()
		if d.conn == conn {
			d.conn = nil
			d.connected = false
		}
		d.mu.Unlock()
	}
}

func (d *discordClient) run() {
	for {
		d.mu.Lock()
		disabled, clientID := d.disabled, d.clientID
		d.mu.Unlock()
		if disabled || clientID == "" {
			time.Sleep(3 * time.Second)
			continue
		}
		conn, err := discordDial()
		if err != nil {
			time.Sleep(8 * time.Second)
			continue
		}
		if err := discordHandshake(conn, clientID); err != nil {
			_ = conn.Close()
			time.Sleep(8 * time.Second)
			continue
		}
		d.mu.Lock()
		d.conn = conn
		d.connected = true
		d.mu.Unlock()
		d.push()

		closed := make(chan struct{})
		go func() {
			defer close(closed)
			for {
				if _, _, err := readDiscordFrame(conn); err != nil {
					return
				}
			}
		}()
		tick := time.NewTicker(30 * time.Second)
		alive := true
		for alive {
			select {
			case <-d.update:
				d.push()
			case <-tick.C:
				d.push()
			case <-closed:
				alive = false
			}
		}
		tick.Stop()
		d.mu.Lock()
		d.connected = false
		if d.conn == conn {
			d.conn = nil
		}
		d.mu.Unlock()
		_ = conn.Close()
		time.Sleep(4 * time.Second)
	}
}
