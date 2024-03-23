package main

import (
	"encoding/json"
	"math"
	"net"
	"os"
	"os/signal"
	"strings"
	"swaystatus/swayipc"
	"syscall"
	"time"
)

func get_focused_name(root map[string]interface{}) string {
	root_focus, _ := root["focus"].([]interface{})
	root_nodes, _ := root["nodes"].([]interface{})

	if len(root_focus) == 0 {
		return root["name"].(string)
	}

	for _, node := range root_nodes {
		if node.(map[string]interface{})["id"].(float64) == root_focus[0] {
			return get_focused_name(node.(map[string]interface{}))
		}
	}
	return ""
}

func get_time() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func get_title() string {
	addr := swayipc.Getaddr()
	conn := swayipc.Getsock(addr)

	msg := swayipc.Pack(4, []byte(""))
	conn.Write(msg)

	var result map[string]interface{}
	resp := swayipc.Unpack(conn)
	json.Unmarshal(resp, &result)
	conn.Close()

	return get_focused_name(result)
}

func update_status_bar(title string, time string) {
	var OFFSET int = 100
	middle := int(math.RoundToEven(float64(len(title)) / 2))
	space := strings.Repeat(" ", OFFSET-middle)

	status := title + space + time + "\n"
	os.Stdout.Write([]byte(status))
}

func poll_changes(title chan string) {
	var addr string = swayipc.Getaddr()
	var conn net.Conn = swayipc.Getsock(addr)
	var events []string = []string{"window"}

	swayipc.Subscribe(conn, events) // subscribe to window change events

	defer conn.Close()

	var result map[string]interface{}
	for {
		response := swayipc.Unpack(conn)
		json.Unmarshal(response, &result)

		if result["change"] == "focus" || result["change"] == "title" {
			window, _ := result["container"].(map[string]interface{})
      
      if window["name"] != nil{
			title <- window["name"].(string) 
      }else{
        title <- ""
      }
		}
	}
}

func main() {
	title_queue := make(chan string)
	interrupt := make(chan os.Signal)

	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	go poll_changes(title_queue)

	var title string = get_title()

mainloop:
	for {
		select {
		case <-tick.C:
			current_time := get_time()
			update_status_bar(title, current_time)

		case title = <-title_queue:
			current_time := get_time()
			update_status_bar(title, current_time)
			tick.Reset(time.Second)

		case <-interrupt:
			os.Stdout.Write([]byte("Closing..."))
			break mainloop
		}
	}
}
