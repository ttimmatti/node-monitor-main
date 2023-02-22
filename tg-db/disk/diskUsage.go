package disk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var CHARACTERS_NEED_ESCAPING []string = []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}

var DISK_PORT = ":6598"

func GetDiskSpaceForSs(s_ips []string) string {
	var chans []chan struct {
		Ip string
		Df string
	}

	for i, s_ip := range s_ips {
		s_ip = strings.Join(strings.Split(s_ip, "\\"), "")
		chans = append(chans, make(chan struct {
			Ip string
			Df string
		}))
		go GetDiskSpace(s_ip, chans[i])
	}

	var text string
	for i, c := range chans {
		s := <-c
		if i != 0 {
			text += escapeMarkdown("=======================================\n")
		}
		text += fmt.Sprintf("%d\\. ", i+1) + "*" + escapeMarkdown(s.Ip) + "*\n"
		if s.Df != "" {
			text += escapeMarkdown(s.Df)
		} else {
			text += escapeMarkdown("No answer. Check if software is installed.\n")
		}
	}

	return text
}

func GetDiskSpace(server_ip string, c chan struct {
	Ip string
	Df string
}) {
	uri := "http://" + server_ip + DISK_PORT + "/df"

	ctx, _ := context.WithTimeout(context.Background(),
		4000*time.Millisecond)
	r, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
		uri, nil)
	if err != nil {
		c <- struct {
			Ip string
			Df string
		}{server_ip, ""}
		return
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		c <- struct {
			Ip string
			Df string
		}{server_ip, ""}
		return
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	var res map[string]string
	if err := json.Unmarshal(b, &res); err != nil {
		c <- struct {
			Ip string
			Df string
		}{server_ip, ""}
		return
	}

	if res["result"] != "ok" {
		c <- struct {
			Ip string
			Df string
		}{server_ip, ""}
		return
	}

	c <- struct {
		Ip string
		Df string
	}{server_ip, res["response"]}
}

func escapeMarkdown(str string) string {
	for _, esc_char := range CHARACTERS_NEED_ESCAPING {
		str = strings.Join(strings.Split(str, esc_char), "\\"+esc_char)
	}
	return str
}
