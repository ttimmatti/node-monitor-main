package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var ASSET_ID string
var GRAFFITI string

func main() {
	SetLogger()

	log.Println("Reading Graffiti")
	graffiti, err := GetGraffiti()
	if err != nil {
		log.Fatalln("main:", err)
	}
	GRAFFITI = graffiti

	log.Printf("Starting server")

	http.HandleFunc("/get", Get)
	http.HandleFunc("/ping", Pong)

	http.HandleFunc("/mint", MintRequest)
	http.HandleFunc("/faucet", FaucetRequest)
	http.HandleFunc("/burn", BurnRequest)
	http.HandleFunc("/send", SendRequest)

	log.Println("Listening on port 6596")
	http.ListenAndServe("0.0.0.0:6596", nil)

	//TODO: it needs to write both to the log
	// and to the logfile
	// and implement readlogs
	// make it systemctl able and enabled for restart
	//make a ping func to know if it's working or not,
	// so i can check before executing further cmds
}

func AutoFaucet() {
	//requests faucet every 24 hours
	for {
		result, err := IronFaucet()
		if err != nil {
			log.Printf("AutoFaucet: %s", err)
		}
		if result {
			log.Println("AutoFaucet: Faucet request succeeded!")
		} else {
			log.Println("AutoFaucet: Faucet request did not succeed !!!")
		}
		time.Sleep(24 * time.Hour)
	}
}

func SendRequest(w http.ResponseWriter, r *http.Request) {
	result, err := IronSend()
	if err != nil {
		log.Println(err)
		w.Write(ErrInJson(fmt.Errorf("MintRequest: %w", err)))
		return
	}

	w.Write([]byte(result))
}

func IronSend() (string, error) {
	if len(ASSET_ID) < 10 {
		return "", fmt.Errorf("IronSend: wrong ASSET_ID")
	}

	//ironfish wallet:send -t dfc2679369551e64e3950e06a88e68466e813c63b100283520045925adbe59ca -i $ASSET -a 0.01 -o 0.00000001 --confirm

	amount := fmt.Sprintf("%f", float32(rand.Intn(1000000)+1)/float32(10000000))
	fee := "0.00000001"
	wallet_burn := "wallet:send"
	a := "-a=" + amount
	o := "-o=" + fee
	c := "--confirm"
	i := "-i=" + ASSET_ID
	t := "-t=" + "dfc2679369551e64e3950e06a88e68466e813c63b100283520045925adbe59ca"
	cmd := exec.Command("node", "/usr/bin/ironfish",
		wallet_burn,
		a, o, c, i, t)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("IronSend: %s", stderr.String())
	}

	hash, err := GetHash(out.String())
	if err != nil {
		return "", fmt.Errorf("IronSend: %w", err)
	}

	return `{"result":"ok","hash":"` + hash + `"`, nil
}

func BurnRequest(w http.ResponseWriter, r *http.Request) {
	result, err := IronBurn()
	if err != nil {
		log.Println(err)
		w.Write(ErrInJson(fmt.Errorf("BurnRequest: %w", err)))
		return
	}

	w.Write([]byte(result))
}

func IronBurn() (string, error) {
	if len(ASSET_ID) < 10 {
		return "", fmt.Errorf("IronBurn: wrong ASSET_ID")
	}

	//ironfish wallet:burn -i $ASSET -a 0.01 -o 0.00000001 --confirm

	log.Println("CHANGE FILE PATH FOR CONFIG")

	amount := fmt.Sprintf("%f", float32(rand.Intn(1000000)+1)/float32(10000000))
	fee := "0.00000001"
	wallet_burn := "wallet:burn"
	a := "-a=" + amount
	o := "-o=" + fee
	c := "--confirm"
	i := "-i=" + ASSET_ID
	cmd := exec.Command("node", "/usr/bin/ironfish",
		wallet_burn,
		i, a, o, c)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("IronBurn: %s", stderr.String())
	}

	hash, err := GetHash(out.String())
	if err != nil {
		return "", fmt.Errorf("IronBurn: %w", err)
	}

	return `{"result":"ok","hash":"` + hash + `"`, nil
}

func MintRequest(w http.ResponseWriter, r *http.Request) {
	result, err := IronMint()
	if err != nil {
		log.Println(err)
		w.Write(ErrInJson(fmt.Errorf("MintRequest: %w", err)))
		return
	}

	w.Write([]byte(result))
}

func IronMint() (string, error) {
	amount := fmt.Sprintf("%d", rand.Int()%2000+5)
	fee := "0.00000001"
	wallet_mint := "wallet:mint"
	a := "-a=" + amount
	o := "-o=" + fee
	c := "--confirm"
	n := "-n=" + GRAFFITI
	m := "-m=" + GRAFFITI
	cmd := exec.Command("node", "/usr/bin/ironfish",
		wallet_mint,
		a, o, c, n, m)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("IronMint: %s", stderr.String())
	}

	//save asset_id
	id, err := GetAssetId(out.String())
	if err != nil {
		return "", fmt.Errorf("IronMint: %w", err)
	}

	hash, err := GetHash(out.String())
	if err != nil {
		return "", fmt.Errorf("IronMint: %w", err)
	}

	ASSET_ID = id

	return `{"result":"ok","asset_id":"` + ASSET_ID + `","hash":"` + hash + `"`, nil
}

// func IronMintExample() (string, error) {
// 	mint_out := IR_MINT_EX
// 	asset_id, err := GetAssetId(mint_out)
// 	if err != nil {
// 		return "", fmt.Errorf("IronMint: %w", err)
// 	}
// 	hash, err := GetHash(mint_out)
// 	if err != nil {
// 		return "", fmt.Errorf("IronMint: %w", err)
// 	}

// 	ASSET_ID = asset_id

// 	return `{"id":"` + asset_id + `","hash":"` + hash + `"}`, nil
// }

func GetAssetId(mint_out string) (string, error) {
	log.Println(mint_out)
	id_N := strings.Index(mint_out, "Identifier")
	if id_N == -1 {
		return "", fmt.Errorf("GetAssetId: No 'Identifier':" + mint_out)
	}
	line_N := strings.Index(mint_out[id_N:], "\n")

	id := mint_out[id_N+12 : id_N+line_N]
	if len(id) < 10 {
		return "", fmt.Errorf("GetAssetId: len(id)<10; ID: " + id)
	}

	return id, nil
}

func GetHash(tx_out string) (string, error) {
	log.Println(tx_out)
	id_N := strings.Index(tx_out, "Hash")
	if id_N == -1 {
		return "", fmt.Errorf("GetHash: No 'Hash':" + tx_out)
	}
	line_N := strings.Index(tx_out[id_N:], "\n")

	hash := tx_out[id_N+6 : id_N+line_N]
	if len(hash) < 10 {
		return "", fmt.Errorf("GetHash: len(id)<10; ID: " + hash)
	}

	return hash, nil
}

func FaucetRequest(w http.ResponseWriter, r *http.Request) {
	result, err := IronFaucet()
	if err != nil {
		log.Println(err)
		w.Write(ErrInJson(fmt.Errorf("MintRequest: %w", err)))
		return
	}

	w.Write([]byte(`{"result":"ok","faucet":"` + fmt.Sprintf("%v", result) + `"}`))
}

func IronFaucet() (bool, error) {
	cmd := exec.Command("node", "/usr/bin/ironfish",
		"faucet")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return false, fmt.Errorf("IronMint: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return false, fmt.Errorf("IronMint: %w: %s", err, stderr.String())
	}
	stdin.Write([]byte("\n"))

	if err := cmd.Wait(); err != nil {
		return false, fmt.Errorf("IronMint: %w: %s", err, stderr.String())
	}

	output := out.String()
	if !strings.Contains(output, "Congratulations!") {
		return false, fmt.Errorf("faucet no good; outp: %s", output)
	}

	return true, nil
}

func Get(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	method := query.Get("ret")

	response, err := Output(method)
	if err != nil {
		log.Println(err)
		w.Write(ErrInJson(err))
		return
	}

	log.Print(" + ")

	w.Write(response)
}

func Pong(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("{\"result\":\"ok\"}"))
}

func Output(method string) ([]byte, error) {
	status, last_block, version, err := GetAll()
	if err != nil {
		return []byte{}, fmt.Errorf("Output: %w", err)
	}

	var mp map[string]string

	switch method {
	case "all":
		mp = map[string]string{
			"lastBlock": fmt.Sprintf("%d", last_block),
			"version":   version,
		}
	case "status":
		mp = map[string]string{
			"status": status,
		}
	default:
		return []byte{}, fmt.Errorf("request method not satisfied")
	}

	mpJson, _ := json.Marshal(mp)

	return mpJson, nil
}

// status, last_block, version, nil
func GetAll() (string, int, string, error) {
	status_response, err := GetStatus()
	if err != nil {
		return "", 0, "", fmt.Errorf("GetAll: %w", err)
	}

	last_block, err := GetLastBlock(status_response)
	if err != nil {
		return "", 0, "", fmt.Errorf("GetLastBlock: %w", err)
	}

	version, err := GetVersion(status_response)
	if err != nil {
		return "", 0, "", fmt.Errorf("GetVersion: %w", err)
	}

	return status_response, last_block, version, nil
}

func GetLastBlock(status string) (int, error) {
	blockchain := strings.SplitAfter(status, "Blockchain")[1]
	//TODO: if empty or don't have a value return error

	indBlockOpens := strings.Index(blockchain, "(")
	indBlockCloses := strings.Index(blockchain, ")")
	block := blockchain[indBlockOpens+1 : indBlockCloses]

	blockInt, err := strconv.Atoi(block)
	if err != nil {
		return 0, fmt.Errorf("could not convert from string to int. Str: %s. Err: %w", block, err)
	}

	return blockInt, nil
}

func GetVersion(status string) (string, error) {
	indVersionOpens := strings.Index(status, "0")
	indVersionCloses := strings.Index(status, "@")
	//TODO: if -1 then error

	version := status[indVersionOpens : indVersionCloses-1]

	return version, nil
}

func GetStatus() (string, error) {
	cmd := exec.Command("node", "/usr/bin/ironfish", "status")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("GetStatus: %w: %s", err, stderr.String())
	}

	return out.String(), nil
}

// func GetStatusExample() (string, error) {
// 	st, _ := os.ReadFile("response_example.txt")
// 	return string(st), nil
// }

func ErrInJson(err error) []byte {
	mp := map[string]string{
		"error": err.Error(),
	}
	jsErr, _ := json.Marshal(mp)

	return jsErr
}

func SetLogger() {
	log_file, err := os.OpenFile("/root/iron_tg/logs.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println(err)
	}

	log.Default().SetFlags(log.Lshortfile)

	log.Default().SetOutput(log_file)
}

// func whichIronfish() string {
// 	which, _ := filepath.Abs("/usr/bin/ironfish")

// 	return string(which)
// }

// func home() string {
// 	home, _ := os.LookupEnv("HOME")

// 	return string(home)
// }

func GetGraffiti() (string, error) {
	configB, err := os.ReadFile("/root/iron_tg/config.json")
	if err != nil {
		return "", fmt.Errorf("GetGraffiti: %w", err)
	}
	config := map[string]string{}
	if err := json.Unmarshal(configB, &config); err != nil {
		return "", fmt.Errorf("GetGraffiti: %w", err)
	}
	graffiti := config["graffiti"]
	if graffiti == "" {
		return "", fmt.Errorf("GetGraffiti: graffiti empty")
	}
	return graffiti, nil
}

// const IR_MINT_EX = `Creating the transaction: [░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░] 0% | ETA: 0s
// Creating the transaction: [░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░] 1% | ETA: 100s
// Creating the transaction: [█░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░] 3% | ETA: 98s
// Creating the transaction: [████████████████████████████████████████] 100% | ETA: 0s

// Minted asset 123 from 123
// Asset Identifier: eb49786430eea33e340e52161f013833b0ce37aad900427e044be5d80fb7f125
// Value: 1552.00000000

// Transaction Hash: 5a9c2ff741b72875b38c153112bf6eefd6c0ecee6225cba34c49c14eb5d34594

// Find the transaction on https://explorer.ironfish.network/transaction/5a9c2ff741b72875b38c153112bf6eefd6c0ecee6225cba34c49c14eb5d34594 (it can take a few minutes before the transaction appears in the Explorer)
// `
