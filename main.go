package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"megafin_farmer/core"
	"megafin_farmer/global"
	"megafin_farmer/utils"
	"os"
	"strconv"
	"strings"
	"sync"
)

func inputUser() string {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	return strings.TrimSpace(scanner.Text())
}

func handlePanic() {
	if r := recover(); r != nil {
		log.Printf("Unexpected Error: %v", r)
		fmt.Println("Press Enter to Exit..")
		_, err := fmt.Scanln()
		if err != nil {
			os.Exit(1)
		}
		os.Exit(1)
	}
}

func startTasks(userAction int,
	accountsList []string,
	proxyList []string) {
	var wg sync.WaitGroup

	if userAction == 1 {
		fmt.Println()
		fmt.Printf("Threads: ")
		threads, err := strconv.Atoi(inputUser())

		if err != nil {
			log.Panicf("Wrong Threads Number: %s", err)
		}

		global.Semaphore = make(chan struct{}, threads)
		fmt.Println()

		for i := 0; i < len(accountsList); i++ {
			account := accountsList[i]
			proxy := proxyList[i]

			wg.Add(1)

			go func(acc, prox string) {
				defer wg.Done()

				core.StartFarmAccount(acc, prox)
			}(account, proxy)
		}

		wg.Wait()

	} else if userAction == 2 {
		var generatedAccountsList []string
		fmt.Printf("\nHow Many Accounts Do You Want To Generate: ")
		accountsCountToGenerate, err := strconv.Atoi(inputUser())
		fmt.Println()

		if err != nil {
			log.Panicf("Error converting account count to int: %v", err)
		}

		for i := 0; i < accountsCountToGenerate; i++ {
			privateKey, err := crypto.GenerateKey()
			if err != nil {
				log.Printf("Error generating private key: %v", err)
				continue
			}

			privateKeyHex := hex.EncodeToString(crypto.FromECDSA(privateKey))

			log.Printf("Successfully Generated Account %s | [%d/%d]", privateKeyHex, i+1, accountsCountToGenerate)

			generatedAccountsList = append(generatedAccountsList, privateKeyHex)
		}

		utils.AppendFile("./data/accounts.txt", strings.Join(generatedAccountsList, "\n")+"\n")

	} else if userAction == 3 {
		var totalMgfBalance, totalUsdcBalance float64
		var mu sync.Mutex

		fmt.Println()
		fmt.Printf("Threads: ")
		threads, err := strconv.Atoi(inputUser())

		if err != nil {
			log.Panicf("Wrong Threads Number: %s", err)
		}

		global.Semaphore = make(chan struct{}, threads)
		fmt.Println()

		for i := 0; i < len(accountsList); i++ {
			account := accountsList[i]
			proxy := proxyList[i]

			wg.Add(1)

			go func(acc, prox string) {
				defer wg.Done()

				mgfBalance, usdcBalance := core.ParseAccountBalance(acc, prox)

				mu.Lock()
				totalMgfBalance += mgfBalance
				totalUsdcBalance += usdcBalance
				mu.Unlock()
			}(account, proxy)
		}

		wg.Wait()

		fmt.Println()
		fmt.Printf("Total MGF Balance: %f\n", totalMgfBalance)
		fmt.Printf("Total USDC Balance: %f\n", totalUsdcBalance)

	}
}

func main() {
	var proxyListSorted []string
	var accountsListSorted []string
	defer handlePanic()

	err := utils.ReadJson("./data/config.json", &global.ConfigFile)

	if err != nil {
		log.Panicf("Error reading config file: %s", err)
	}

	accountsList, err := utils.ReadFileByRows("./data/accounts.txt")

	if err != nil {
		log.Panicf("Error While Reading Accounts File: %s", err)
	}

	proxyList, err := utils.ReadFileByRows("./data/proxies.txt")

	if err != nil {
		log.Panicf("Error While Reading Proxy File: %s", err)
	}

	for _, proxy := range proxyList {
		parsedProxy, err := utils.ParseProxy(proxy)

		if err != nil {
			log.Printf("%s wrong proxy format", proxy)
			continue
		}

		proxyListSorted = append(proxyListSorted, parsedProxy)
	}

	for _, account := range accountsList {
		parsedAccountKey := utils.RemoveHexPrefix(account)
		accountsListSorted = append(accountsListSorted, parsedAccountKey)
	}

	fmt.Printf("Successfully Loaded %d Accounts // %d Proxies", len(accountsList), len(proxyListSorted))

	limit := len(accountsListSorted)

	if len(proxyListSorted) < limit {
		limit = len(proxyListSorted)
	}

	accountsListSorted = accountsListSorted[:limit]
	proxyListSorted = proxyListSorted[:limit]

	fmt.Printf("\n1. Start Farming" +
		"\n2. Generate Accounts" +
		"\n3. Parse Accounts Balance" +
		"\nEnter Your Action: ")

	userAction, err := strconv.Atoi(inputUser())

	if err != nil {
		log.Panicf("Error When Enter Your Action: %s", err)
	}

	startTasks(userAction, accountsListSorted, proxyListSorted)

	fmt.Printf("The Work Has Beeen Successfully Finished")
	fmt.Printf("\n\nPress Enter to Exit..")
	inputUser()
}
