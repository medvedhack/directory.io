package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"encoding/json"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
)

const ResultsPerPage = 30

const PageTemplateHeader = `<html>
<head>
	<title>All bitcoin private keys</title>
	<meta charset="utf8" />
	<link href="http://fonts.googleapis.com/css?family=Open+Sans" rel="stylesheet" type="text/css">
	<style>
		body{font-size: 9pt; font-family: 'Open Sans', sans-serif;}
		a{text-decoration: none}
		a:hover {text-decoration: underline}
		.keys > span:hover { background: #f0f0f0; }
		span:target { background: #ccffcc; }
	</style>
</head>
<body>
<h1>Bitcoin private key database</h1>
<h2>Page %s out of %s</h2>
<a href="/%s">previous</a> | <a href="/%s">next</a>
<pre class="keys">
<strong>Private Key</strong>                                            <strong>Address</strong>                            <strong>b</strong>  <strong>Compressed Address</strong>            <strong>b</strong>
`

const PageTemplateFooter = `</pre>
<pre style="margin-top: 1em; font-size: 8pt">
It took a lot of computing power to generate this database. 
</pre>
Donations welcome: 1LdHnyauFSW1dkvD4xf7FSQN3m6fK3ShiS   <a href="  /%30s   ">                             previous</a> |                             <a href="/%20s"><strong>next</strong> </a>
</body>
</html>`

const KeyTemplate = `<span id="%s"><a href="/warning:understand-how-this-works!/%s">+</a> <span title="%s">%s </span> <a href="https://blockchain.info/address/%s">%34s</a> <strong>%d</strong>  <a href="https://blockchain.info/address/%s">%34s</a> %d </span>
`

var (
	// Total bitcoins
	total = new(big.Int).SetBytes([]byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE,
		0xBA, 0xAE, 0xDC, 0xE6, 0xAF, 0x48, 0xA0, 0x3B, 0xBF, 0xD2, 0x5E, 0x8C, 0xD0, 0x36, 0x41, 0x40,
	})

	// One
	one = big.NewInt(1)

	// Total pages
	_pages = new(big.Int).Div(total, big.NewInt(ResultsPerPage))
	pages  = _pages.Add(_pages, one)
)

type Key struct {
	private      string
	number       string
	compressed   string
	cb           int
	uncompressed string
	ucb          int
}

func getBalance(body []byte) (b int) {

    var f interface{}
    err := json.Unmarshal(body, &f)
    if err != nil {
        fmt.Println(err)
    }
    m := f.(map[string]interface{})
    for k,v := range m{
    if k == "balance" {
    return int(v.(float64))
}
}
    return 0
}


func check_balance3(address string, ch chan int) {
        query_comp := "https://api.blocktrail.com/v1/btc/address/" + address + "?api_key=1"
        resp, err := http.Get(query_comp)
        if err != nil {
                log.Fatalf("Checking balance (uncomp): %s\n", err)
        }
        defer resp.Body.Close()
        body, _ := ioutil.ReadAll(resp.Body)
        balance := getBalance(body)
        ch <- balance
}


func check_balance1(address string, ch chan int) {
	query_comp := "https://blockchain.info/q/addressbalance/" + address
	resp, err := http.Get(query_comp)
	if err != nil {
		log.Fatalf("Checking balance (uncomp): %s\n", err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	bodystring := string(body)
	balance, _ := strconv.Atoi(bodystring)
	//return strconv.Atoi(bodystring)
	ch <- balance
}

func check_balance2(address string, ch chan int) {
	query_comp := "https://blockexplorer.com/api/addr/" + address + "/balance"
	resp, err := http.Get(query_comp)
	if err != nil {
		log.Fatalf("Checking balance (uncomp): %s\n", err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	bodystring := string(body)
	balance, _ := strconv.Atoi(bodystring)
	//return strconv.Atoi(bodystring)
	ch <- balance
}

func compute(count *big.Int) (keys [ResultsPerPage]Key, length int) {
	f, err := os.OpenFile("/home/waterstraw/directory.io/balance.log", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	var padded [32]byte
	//cha := make(chan int)
	chb := make(chan int)
	var i int
	for i = 0; i < ResultsPerPage; i++ {
		// Increment our counter
		count.Add(count, one)
		if count.Cmp(total) > 0 {
			break
		}
		// Copy count value's bytes to padded slice
		copy(padded[32-len(count.Bytes()):], count.Bytes())

		// Get private and public keys
		privKey, public := btcec.PrivKeyFromBytes(btcec.S256(), padded[:])

		// Get compressed and uncompressed addresses for public key
		caddr, _ := btcutil.NewAddressPubKey(public.SerializeCompressed(), &chaincfg.MainNetParams)
		uaddr, _ := btcutil.NewAddressPubKey(public.SerializeUncompressed(), &chaincfg.MainNetParams)

		// Encode addresses
		wif, _ := btcutil.NewWIF(privKey, &chaincfg.MainNetParams, false)
		keys[i].private = wif.String()
		keys[i].number = count.String()
		keys[i].compressed = caddr.EncodeAddress()
		keys[i].uncompressed = uaddr.EncodeAddress()

		var com_balance, uncom_balance int
		//go check_balance1(caddr.EncodeAddress(),cha)
		//com_balance = <-cha
		//time.Sleep(time.Duration(1) * time.Second)
		go check_balance2(uaddr.EncodeAddress(), chb)
		uncom_balance = <-chb

		keys[i].cb = com_balance
		keys[i].ucb = uncom_balance
		fmt.Println(wif.String() + " " + strconv.Itoa(uncom_balance))
		if uncom_balance > 0 || com_balance > 0 {
			var balance int
			if uncom_balance > com_balance {
				balance = uncom_balance
			} else {
				balance = com_balance
			}
			ret_str := "Private Key (Base64); " + wif.String() + "; Balance;" + strconv.Itoa(balance) + ";"
			fmt.Println(ret_str)
			f.WriteString(ret_str)
		}

	}
	return keys, i
}

func PageRequest(w http.ResponseWriter, r *http.Request) {
	// Default page is page 1
	if len(r.URL.Path) <= 1 {
		r.URL.Path = "/1"
	}

	// Convert page number to bignum
	page, success := new(big.Int).SetString(r.URL.Path[1:], 0)
	if !success {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Make sure page number cannot be negative or 0
	page.Abs(page)
	if page.Cmp(one) == -1 {
		page.SetInt64(1)
	}

	// Make sure we're not above page count
	if page.Cmp(pages) > 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Get next and previous page numbers
	previous := new(big.Int).Sub(page, one)
	next := new(big.Int).Add(page, one)

	// Calculate our starting key from page number
	start := new(big.Int).Mul(previous, big.NewInt(ResultsPerPage))

	// Send page header
	fmt.Fprintf(w, PageTemplateHeader, page, pages, previous, next)

	// Send keys
	keys, length := compute(start)
	for i := 0; i < length; i++ {
		key := keys[i]
		fmt.Fprintf(w, KeyTemplate, key.private, key.private, key.number, key.private, key.uncompressed, key.uncompressed, key.ucb, key.compressed, key.compressed, key.cb)
	}

	// Send page footer
	fmt.Fprintf(w, PageTemplateFooter, previous, next)
}

func RedirectRequest(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[36:]

	wif, err := btcutil.DecodeWIF(key)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	page, _ := new(big.Int).DivMod(new(big.Int).SetBytes(wif.PrivKey.D.Bytes()), big.NewInt(ResultsPerPage), big.NewInt(ResultsPerPage))
	page.Add(page, one)

	fragment, _ := btcutil.NewWIF(wif.PrivKey, &chaincfg.MainNetParams, false)

	http.Redirect(w, r, "/"+page.String()+"#"+fragment.String(), http.StatusTemporaryRedirect)
}

func main() {
	http.HandleFunc("/", PageRequest)
	http.HandleFunc("/warning:understand-how-this-works!/", RedirectRequest)

	log.Println("Listening")
	log.Fatal(http.ListenAndServe(":8085", nil))
}
