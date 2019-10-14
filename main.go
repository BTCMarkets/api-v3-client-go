package main

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const apiKey = "your api key"
const privateKey = "your private key in base64"
const baseUrl = "https://server-host-url"

const apiPath = "/v3/orders"

func main() {

	id := extractOrderId(addOrder())
	getOpenOrders()
	cancelOrder(id)
}

func extractOrderId(order string) string {
	re := regexp.MustCompile(`"orderId":\s*"(\w*)"`)
	matches := re.FindStringSubmatch(order)
	return matches[1]
}

func getOpenOrders() {
	log.Println("list orders")
	responseBody := makeHttpCall("GET", apiPath, "status=open", "")
	log.Println(responseBody)
}

func addOrder() string {
	orderData := map[string]string{
		"marketId": "XRP-AUD",
		"price":    "0.1",
		"amount":   "0.1",
		"side":     "Bid",
		"type":     "Limit",
	}
	orderJson, _ := json.Marshal(orderData)
	log.Println("adding order " + string(orderJson))

	responseBody := makeHttpCall("POST", apiPath, "", string(orderJson))
	log.Println(responseBody)
	return responseBody
}

func cancelOrder(orderId string) {
	log.Println("cancelling order " + orderId)
	log.Println(
		makeHttpCall("DELETE", apiPath+"/"+orderId, "", ""))
}

func makeHttpCall(method string, path string, query string, body string) string {

	headers := buildAuthHeaders(method, path, body)

	url := baseUrl + path
	if query != "" {
		url += "?" + query
	}

	var request *http.Request

	if body != "" {
		request, _ = http.NewRequest(method, url, strings.NewReader(body))
	} else {
		request, _ = http.NewRequest(method, url, nil)
	}

	request.Header = headers
	response, err := (&http.Client{}).Do(request)
	log.Println(response)
	data, err := ioutil.ReadAll(response.Body)

	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	responseBody := string(data)
	return responseBody
}

func buildAuthHeaders(method string, path string, body string) http.Header {
	//getting now() in milliseconds
	nowMs := strconv.FormatInt(time.Now().UTC().UnixNano()/1000000, 10)

	stringToSign := method + path + nowMs

	if body != "" {
		stringToSign += body
	}

	return http.Header{
		"Content-Type":      []string{"application/json"},
		"Accept":            []string{"application/json"},
		"Accept-Charset":    []string{"UTF-8"},
		"BM-AUTH-APIKEY":    []string{apiKey},
		"BM-AUTH-TIMESTAMP": []string{nowMs},
		"BM-AUTH-SIGNATURE": []string{signMessage(privateKey, stringToSign)},
	}
}

func signMessage(key string, message string) string {
	encodedKey, _ := base64.StdEncoding.DecodeString(key)
	mac := hmac.New(sha512.New, encodedKey)
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
